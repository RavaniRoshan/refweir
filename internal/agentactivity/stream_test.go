package agentactivity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHub_SubscribeAndBroadcast(t *testing.T) {
	hub := NewHub(10)
	ch, cancel := hub.Subscribe(5)
	defer cancel()

	act := Activity{AgentName: "agent-live", Action: "repo_clone", Status: "success"}
	hub.Record(act)

	select {
	case received := <-ch:
		if received.AgentName != "agent-live" {
			t.Errorf("expected agent-live, got %s", received.AgentName)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for activity on subscriber channel")
	}

	cancel()

	// Subsequent records should not panic after cancel
	hub.Record(Activity{AgentName: "agent-after", Action: "file_write"})
}

func TestHub_SubscribeSlowSubscriberNonBlocking(t *testing.T) {
	hub := NewHub(10)
	// Buffer size 1 to easily overflow
	ch, cancel := hub.Subscribe(1)
	defer cancel()

	// Fill buffer
	hub.Record(Activity{AgentName: "msg-1"})
	// Overflow buffer - Record must not block or deadlock
	done := make(chan struct{})
	go func() {
		hub.Record(Activity{AgentName: "msg-2"})
		hub.Record(Activity{AgentName: "msg-3"})
		close(done)
	}()

	select {
	case <-done:
		// success, didn't block
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Record blocked on slow subscriber")
	}

	// Drain first message
	item := <-ch
	if item.AgentName != "msg-1" {
		t.Errorf("expected msg-1, got %s", item.AgentName)
	}
}

func TestStreamHandler_SnapshotAndLiveActivity(t *testing.T) {
	hub := NewHub(10)
	hub.Record(Activity{AgentName: "seed-agent", Action: "init"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agent/activities/stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		StreamHandler(hub)(rec, req)
		close(done)
	}()

	// Wait for headers and initial snapshot
	time.Sleep(50 * time.Millisecond)

	hub.Record(Activity{AgentName: "live-streamed", Action: "git_push", Status: "success"})

	// Give time to flush live activity then cancel request
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done

	body := rec.Body.String()
	if !strings.Contains(body, "event: snapshot") {
		t.Errorf("expected event: snapshot in SSE output, got:\n%s", body)
	}
	if !strings.Contains(body, "seed-agent") {
		t.Errorf("expected seed-agent in SSE snapshot, got:\n%s", body)
	}
	if !strings.Contains(body, "event: activity") {
		t.Errorf("expected event: activity in SSE output, got:\n%s", body)
	}
	if !strings.Contains(body, "live-streamed") {
		t.Errorf("expected live-streamed in SSE output, got:\n%s", body)
	}
}
