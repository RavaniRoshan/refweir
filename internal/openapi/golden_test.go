package openapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"refweir.dev/internal/auth"
	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/origin"
	"refweir.dev/internal/store"
)

func TestOpenAPI_Coverage(t *testing.T) {
	// 1. Initialize store and origin server to register all routes
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	git, err := gitstore.New(filepath.Join(dir, "git"))
	if err != nil {
		t.Fatalf("init gitstore: %v", err)
	}

	registry := auth.NewRegistry()
	srv := origin.New(db, git, registry, "http://localhost:8080/v1/origin", "localhost:8080")
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	// 2. Collect all registered operations
	registeredOps := make(map[string]bool)
	for _, op := range registry.Operations() {
		registeredOps[op] = true
	}

	if len(registeredOps) < 119 {
		t.Fatalf("expected at least 119 registered operations, got %d", len(registeredOps))
	}

	// 3. Validate that every operation in the OpenAPI specification has a registered handler
	missing, err := ValidateCoverage(registeredOps)
	if err != nil {
		t.Fatalf("OpenAPI coverage verification failed (%d missing): %v", len(missing), err)
	}
	t.Logf("Validated full 100%% OpenAPI coverage across all %d operations!", len(registeredOps))
}

func TestOpenAPI_Handler(t *testing.T) {
	h := Handler()
	req := httptest.NewRequest(http.MethodGet, "/v1/origin/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "yaml") {
		t.Errorf("expected yaml content-type, got: %s", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "openapi: 3.1.0") {
		t.Errorf("expected openapi 3.1.0 declaration in body")
	}
	if !strings.Contains(body, "Cursor Origin API") {
		t.Errorf("expected title in body")
	}
}

func TestOpenAPI_ForgeHandler(t *testing.T) {
	h := ForgeHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "yaml") {
		t.Errorf("expected yaml content-type, got: %s", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Refweir Forge API") {
		t.Errorf("expected Refweir Forge API title in body")
	}
}

func TestOpenAPI_ForgeSyncWithRootAPI(t *testing.T) {
	rootPath := filepath.Join("..", "..", "api", "openapi.yaml")
	rootBytes, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", rootPath, err)
	}
	if !bytes.Equal(ForgeSpec, rootBytes) {
		t.Fatalf("internal/openapi/forge.openapi.yaml and %s have drifted; they must be identical", rootPath)
	}
}
