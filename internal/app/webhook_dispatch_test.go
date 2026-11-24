package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
)

func newWebhookTestDB(t *testing.T) *store.DB {
	t.Helper()
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestDispatchRepoWebhook_RealHTTP(t *testing.T) {
	db := newWebhookTestDB(t)
	ctx := context.Background()

	secret := "super-secret-key-123"
	receivedEvent := make(chan string, 1)
	receivedSig := make(chan string, 1)
	receivedPayload := make(chan string, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedEvent <- r.Header.Get("X-Refweir-Event")
		receivedSig <- r.Header.Get("X-Hub-Signature-256")
		body, _ := io.ReadAll(r.Body)
		receivedPayload <- string(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received": true}`))
	}))
	defer srv.Close()

	ns := store.Namespace{ID: idgen.NewNS(), Slug: "acme", Type: "team", Name: "Acme Corp"}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "demo-repo",
		FullName:      "acme/demo-repo",
		DefaultBranch: "main",
		Visibility:    "public",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	hook, err := db.CreateRepoWebhook(ctx, repo.ID, srv.URL, secret, "json", `["pull_request", "ping"]`, true, false)
	if err != nil {
		t.Fatalf("CreateRepoWebhook: %v", err)
	}

	payload := map[string]any{"action": "opened", "number": 42}
	delivery, err := DispatchRepoWebhook(ctx, db, hook, "pull_request", payload)
	if err != nil {
		t.Fatalf("DispatchRepoWebhook: %v", err)
	}

	if !delivery.Delivered {
		t.Errorf("expected delivery.Delivered == true, got %v", delivery.Delivered)
	}
	if delivery.ResponseStatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", delivery.ResponseStatusCode)
	}

	select {
	case evt := <-receivedEvent:
		if evt != "pull_request" {
			t.Errorf("expected event 'pull_request', got %s", evt)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for HTTP webhook event")
	}

	select {
	case sig := <-receivedSig:
		expectedPayload, _ := json.Marshal(payload)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(expectedPayload)
		expectedSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if sig != expectedSig {
			t.Errorf("expected signature %s, got %s", expectedSig, sig)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for HTTP signature")
	}

	// Verify DB delivery records
	deliveries, err := db.ListRepoWebhookDeliveries(ctx, hook.ID, 50)
	if err != nil {
		t.Fatalf("ListRepoWebhookDeliveries: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("expected 1 delivery record, got %d", len(deliveries))
	}
	if deliveries[0].ID != delivery.ID {
		t.Errorf("expected ID %s, got %s", delivery.ID, deliveries[0].ID)
	}
}

func TestDispatchRepoEvent_AsyncFiltering(t *testing.T) {
	db := newWebhookTestDB(t)
	ctx := context.Background()

	received := make(chan string, 5)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Get("X-Refweir-Event")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ns := store.Namespace{ID: idgen.NewNS(), Slug: "acme-filt", Type: "team", Name: "Acme Filter"}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "demo-repo-2",
		FullName:      "acme-filt/demo-repo-2",
		DefaultBranch: "main",
		Visibility:    "public",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// Hook 1: subscribed to issues only
	_, err := db.CreateRepoWebhook(ctx, repo.ID, srv.URL, "", "json", `["issues"]`, true, false)
	if err != nil {
		t.Fatalf("CreateRepoWebhook 1: %v", err)
	}

	// Hook 2: subscribed to * (all)
	_, err = db.CreateRepoWebhook(ctx, repo.ID, srv.URL, "", "json", `["*"]`, true, false)
	if err != nil {
		t.Fatalf("CreateRepoWebhook 2: %v", err)
	}

	// Dispatch pull_request event - should only trigger Hook 2
	DispatchRepoEvent(ctx, db, repo.ID, "pull_request", map[string]string{"state": "open"})

	select {
	case evt := <-received:
		if evt != "pull_request" {
			t.Errorf("expected pull_request event, got %s", evt)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for async webhook event")
	}

	// Ensure no extra events received
	select {
	case extra := <-received:
		t.Fatalf("unexpected extra webhook delivery received: %s", extra)
	case <-time.After(200 * time.Millisecond):
		// OK
	}
}

func TestWebhookDispatch_SSRFRejection(t *testing.T) {
	ctx := context.Background()
	db := newWebhookTestDB(t)

	nsID := idgen.New("ns")
	_ = db.CreateNamespace(ctx, store.Namespace{ID: nsID, Slug: "acme-sec", Name: "Acme Sec"})
	repoID := idgen.New("repo")
	_ = db.CreateRepo(ctx, store.Repo{ID: repoID, NamespaceID: nsID, Name: "sec-repo", FullName: "acme-sec/sec-repo", DefaultBranch: "main"})

	// Direct attempt with AWS metadata IP
	hook := &store.RepoWebhook{
		ID:        "hook_ssrf_1",
		RepoID:    repoID,
		URL:       "http://169.254.169.254/latest/meta-data/",
		Events:    `["push"]`,
		Active:    true,
	}

	delivery, err := DispatchRepoWebhook(ctx, db, hook, "push", map[string]string{"ref": "refs/heads/main"})
	if err == nil {
		t.Fatal("expected error dispatching to metadata IP, got nil")
	}
	if delivery.Delivered {
		t.Errorf("expected delivery.Delivered to be false")
	}
}
