package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"refweir.dev/internal/app"
	"refweir.dev/internal/audit"
	"refweir.dev/internal/config"
	"refweir.dev/internal/store"
	"refweir.dev/pkg/agentlease"
	"refweir.dev/pkg/idgen"
)

func TestConcurrentLeaseIssuanceAndStress(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()

	jwtSecret := "super-secure-32-byte-secret-key-12345!"
	cfg := config.Config{
		ListenAddr: "127.0.0.1:0",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
		JWTSecret:  jwtSecret,
	}

	application, err := app.Open(cfg)
	if err != nil {
		t.Fatalf("app.Open: %v", err)
	}
	defer application.Close()

	srv := httptest.NewServer(application.Handler())
	defer srv.Close()

	nsID := idgen.NewNS()
	if err := application.Store.CreateNamespace(ctx, store.Namespace{
		ID:   nsID,
		Slug: "stress-corp",
		Type: "team",
		Name: "Stress Corp",
	}); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	repoID := idgen.NewRepo()
	if err := application.Store.CreateRepo(ctx, store.Repo{
		ID:            repoID,
		NamespaceID:   nsID,
		Name:          "stress-repo",
		FullName:      "stress-corp/stress-repo",
		DefaultBranch: "main",
	}); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// Concurrently issue 10 leases from 10 distinct agents
	const numWorkers = 10
	var wg sync.WaitGroup
	errCh := make(chan error, numWorkers)
	issuedTokens := make([][]byte, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			agentName := fmt.Sprintf("agent-worker-%02d", workerID)

			reqBody, _ := json.Marshal(map[string]any{
				"agentId":            agentName,
				"repoId":             repoID,
				"allowedRefPrefixes": []string{"refs/heads/agent/"},
				"ttlSeconds":         3600,
			})

			resp, err := http.Post(srv.URL+"/api/v1/agent/leases", "application/json", bytes.NewReader(reqBody))
			if err != nil {
				errCh <- fmt.Errorf("worker %d POST lease error: %w", workerID, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("worker %d expected 200, got %d", workerID, resp.StatusCode)
				return
			}

			var payload struct {
				Token string `json:"token"`
				ID    string `json:"id"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				errCh <- fmt.Errorf("worker %d decode json error: %w", workerID, err)
				return
			}

			if payload.Token == "" {
				errCh <- fmt.Errorf("worker %d received empty token", workerID)
				return
			}

			// Validate lease token cryptographically
			parsedLease, err := agentlease.Verify([]byte(jwtSecret), payload.Token)
			if err != nil {
				errCh <- fmt.Errorf("worker %d verify token error: %w", workerID, err)
				return
			}

			if parsedLease.AgentID != agentName {
				errCh <- fmt.Errorf("worker %d agentId mismatch: got %s, want %s", workerID, parsedLease.AgentID, agentName)
				return
			}

			issuedTokens[workerID] = []byte(payload.Token)

			// Record an audit log event concurrently
			_ = audit.Log(audit.Event{
				Actor:     agentName,
				ActorType: "agent",
				Action:    "agent.concurrent_ping",
				Resource:  "stress-corp/stress-repo",
				Severity:  audit.SeverityInfo,
				Details:   fmt.Sprintf("worker %d concurrent stress test", workerID),
			})
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify all 10 tokens were issued and distinct
	seenTokens := make(map[string]bool)
	for i, tok := range issuedTokens {
		str := string(tok)
		if seenTokens[str] {
			t.Fatalf("duplicate token issued at worker %d", i)
		}
		seenTokens[str] = true
	}

	// Verify audit logs captured the events without race or loss
	recent := audit.Filter("", "agent.concurrent_ping", 50)
	if len(recent) < numWorkers {
		t.Fatalf("expected at least %d audit events, got %d", numWorkers, len(recent))
	}
}
