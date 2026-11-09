package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"refweir.dev/internal/config"
)

func TestOpenAPIForgeSpecEndpoints(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:8080",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}

	application, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() {
		_ = application.Close()
	}()

	endpoints := []string{"/api/openapi.yaml", "/api/v1/openapi.yaml"}
	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		rec := httptest.NewRecorder()
		application.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200; body = %s", ep, rec.Code, rec.Body.String())
		}

		ct := rec.Header().Get("Content-Type")
		if !strings.Contains(ct, "yaml") {
			t.Errorf("GET %s Content-Type = %s, want yaml", ep, ct)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "openapi: 3.1.0") {
			t.Errorf("GET %s body missing openapi 3.1.0", ep)
		}
		if !strings.Contains(body, "Refweir Forge API") {
			t.Errorf("GET %s body missing Refweir Forge API title", ep)
		}
		if !strings.Contains(body, "/api/v1/repos/{owner}/{name}/pulls") {
			t.Errorf("GET %s body missing pulls path", ep)
		}
		if !strings.Contains(body, "/api/v1/agent/activities/stream") {
			t.Errorf("GET %s body missing SSE stream path", ep)
		}
	}
}
