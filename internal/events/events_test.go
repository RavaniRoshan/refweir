package events

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
)

func TestEventsBus_PublishAndDispatch(t *testing.T) {
	t.Setenv("REFWEIR_ALLOW_LOOPBACK_WEBHOOKS", "true")

	tmpDir := t.TempDir()
	db, err := store.Open(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Setup webhook receiver
	receivedEvent := make(chan string, 1)
	receivedPayload := make(chan map[string]any, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedEvent <- r.Header.Get("X-Refweir-Event")
		var p map[string]any
		_ = json.NewDecoder(r.Body).Decode(&p)
		receivedPayload <- p
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Setup namespace & repo
	ns := store.Namespace{ID: idgen.NewNS(), Slug: "testorg", Name: "Test Org", Type: "user"}
	_ = db.CreateNamespace(ctx, ns)
	repo := store.Repo{
		ID:          idgen.NewRepo(),
		NamespaceID: ns.ID,
		Name:        "testrepo",
		FullName:    "testorg/testrepo",
	}
	_ = db.CreateRepo(ctx, repo)

	// Create subscribed webhook
	eventsJSON, _ := json.Marshal([]string{"fork_sync", "push"})
	hook, err := db.CreateRepoWebhook(ctx, repo.ID, srv.URL, "", "json", string(eventsJSON), true, true)
	if err != nil {
		t.Fatalf("CreateRepoWebhook: %v", err)
	}
	if hook == nil {
		t.Fatalf("expected non-nil hook")
	}

	bus := NewBus(db, nil)

	bus.Publish(ctx, DomainEvent{
		RepoID:    repo.ID,
		Event:     "fork_sync",
		Payload:   map[string]any{"action": "synchronized", "merge_type": "fast-forward"},
		Actor:     "testuser",
		ActorType: "user",
		Action:    "repo.fork_sync",
		Resource:  repo.FullName,
		Details:   "Synchronized fork",
	})

	select {
	case evt := <-receivedEvent:
		if evt != "fork_sync" {
			t.Fatalf("expected receivedEvent 'fork_sync', got %q", evt)
		}
		p := <-receivedPayload
		if p["merge_type"] != "fast-forward" {
			t.Fatalf("expected merge_type fast-forward, got %v", p["merge_type"])
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for webhook delivery")
	}
}
