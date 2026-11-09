package agentlease

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerWithObserverAuditsIssuedLease(t *testing.T) {
	secret := []byte("handler-observer-secret-12345678")
	var observed *Lease
	handler := HandlerWithObserver(secret, func(_ *http.Request, lease *Lease) {
		observed = lease
	})

	body, _ := json.Marshal(LeaseRequest{
		AgentID:    "agent-claude",
		RepoID:     "repo-1",
		TTLSeconds: 120,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/leases", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if observed == nil {
		t.Fatal("expected observer to be called")
	}
	if observed.AgentID != "agent-claude" || observed.RepoID != "repo-1" {
		t.Fatalf("unexpected observed lease: %+v", observed)
	}

	var resp LeaseResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != observed.ID || resp.Token == "" {
		t.Fatalf("response does not match observed lease: %+v vs %+v", resp, observed)
	}
}

func TestHandlerWithObserverNilObserver(t *testing.T) {
	handler := Handler([]byte("handler-nil-observer-secret-1234"))
	body, _ := json.Marshal(LeaseRequest{AgentID: "agent"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/leases", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
