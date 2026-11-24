package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"
	"refweir.dev/internal/config"
	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
)

func TestAppLifecycle(t *testing.T) {
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
		if err := application.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	// 1. Health check endpoint works
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want 200", rec.Code)
	}
	var health struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("GET /healthz body is not JSON: %v", err)
	}
	if health.Status != "ok" || health.Version == "" {
		t.Fatalf("GET /healthz = %+v, want status ok and a version", health)
	}

	// 2. Origin route is mounted (GET /v1/origin/namespaces returns 401 unauthenticated when no token provided)
	reqRepo := httptest.NewRequest(http.MethodGet, "/v1/origin/namespaces", nil)
	recRepo := httptest.NewRecorder()
	application.Handler().ServeHTTP(recRepo, reqRepo)

	if recRepo.Code != http.StatusUnauthorized {
		t.Fatalf("GET /v1/origin/namespaces without auth = %d, want 401", recRepo.Code)
	}

	// 3. Prometheus metrics endpoint is mounted
	reqMetrics := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recMetrics := httptest.NewRecorder()
	application.Handler().ServeHTTP(recMetrics, reqMetrics)
	if recMetrics.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want 200", recMetrics.Code)
	}

	// 4. Admin health endpoint returns live cluster metrics
	reqHealth := httptest.NewRequest(http.MethodGet, "/api/v1/admin/health", nil)
	recHealth := httptest.NewRecorder()
	application.Handler().ServeHTTP(recHealth, reqHealth)
	if recHealth.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/health status = %d, want 200", recHealth.Code)
	}
	var healthData map[string]any
	if err := json.Unmarshal(recHealth.Body.Bytes(), &healthData); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}
	if _, ok := healthData["history"]; !ok {
		t.Fatalf("expected health response to contain 'history' field")
	}

	// 4b. Admin metrics history endpoint returns retained samples
	reqHistory := httptest.NewRequest(http.MethodGet, "/api/v1/admin/metrics/history", nil)
	recHistory := httptest.NewRecorder()
	application.Handler().ServeHTTP(recHistory, reqHistory)
	if recHistory.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/metrics/history status = %d, want 200", recHistory.Code)
	}

	// 5. Admin audit logs endpoint returns json log records (no users yet, so
	// the fresh-install bootstrap path is allowed)
	reqAudit := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs", nil)
	recAudit := httptest.NewRecorder()
	application.Handler().ServeHTTP(recAudit, reqAudit)
	if recAudit.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/audit-logs status = %d, want 200", recAudit.Code)
	}

	// 5b. Cluster replication status endpoint returns valid JSON
	reqRepl := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/replication", nil)
	recRepl := httptest.NewRecorder()
	application.Handler().ServeHTTP(recRepl, reqRepl)
	if recRepl.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/cluster/replication status = %d, want 200", recRepl.Code)
	}

	// 5c. Failover promotion endpoint succeeds
	reqPromote := httptest.NewRequest(http.MethodPost, "/api/v1/cluster/failover/promote", nil)
	recPromote := httptest.NewRecorder()
	application.Handler().ServeHTTP(recPromote, reqPromote)
	if recPromote.Code != http.StatusOK {
		t.Fatalf("POST /api/v1/cluster/failover/promote status = %d, want 200", recPromote.Code)
	}

	// 5d. Raft consensus status and proposal endpoints
	reqRaft := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/raft/status", nil)
	recRaft := httptest.NewRecorder()
	application.Handler().ServeHTTP(recRaft, reqRaft)
	if recRaft.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/cluster/raft/status status = %d, want 200", recRaft.Code)
	}

	// 5e. Distributed object shards telemetry endpoint
	reqShards := httptest.NewRequest(http.MethodGet, "/api/v1/cluster/shards", nil)
	recShards := httptest.NewRecorder()
	application.Handler().ServeHTTP(recShards, reqShards)
	if recShards.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/cluster/shards status = %d, want 200", recShards.Code)
	}

	// 6. Embedded Web UI is served on root / and SPA routes
	reqUI := httptest.NewRequest(http.MethodGet, "/", nil)
	recUI := httptest.NewRecorder()
	application.Handler().ServeHTTP(recUI, reqUI)
	if recUI.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", recUI.Code)
	}
	if !strings.Contains(recUI.Body.String(), "<!doctype html>") && !strings.Contains(recUI.Body.String(), "<html") {
		t.Fatalf("expected embedded HTML on /, got %s", recUI.Body.String())
	}

	reqSPA := httptest.NewRequest(http.MethodGet, "/acme/frontend-app", nil)
	recSPA := httptest.NewRecorder()
	application.Handler().ServeHTTP(recSPA, reqSPA)
	if recSPA.Code != http.StatusOK {
		t.Fatalf("GET /acme/frontend-app status = %d, want 200", recSPA.Code)
	}
	if !strings.Contains(recSPA.Body.String(), "<!doctype html>") && !strings.Contains(recSPA.Body.String(), "<html") {
		t.Fatalf("expected embedded HTML SPA fallback, got %s", recSPA.Body.String())
	}

	// 7. Git requests are delegated to git transport (not HTML)
	reqGit := httptest.NewRequest(http.MethodGet, "/acme/repo.git/info/refs?service=git-upload-pack", nil)
	recGit := httptest.NewRecorder()
	application.Handler().ServeHTTP(recGit, reqGit)
	if strings.Contains(recGit.Body.String(), "<!doctype html>") {
		t.Fatalf("expected Git transport response, got HTML")
	}
}

func TestAppOpenFailure(t *testing.T) {
	f := t.TempDir() + "/blocked"
	if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		DataDir: f + "/cannot-mkdir",
	}

	_, err := Open(cfg)
	if err == nil {
		t.Fatal("expected Open to fail for blocked directory path, but got nil")
	}
}

func TestAuditLogsRequireAuthenticationOnceInitialized(t *testing.T) {
	cfg := config.Config{
		ListenAddr: "127.0.0.1:8080",
		DataDir:    t.TempDir(),
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}
	application, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	if _, err := application.Store.UpsertUserByEmail(context.Background(), "admin@example.com", func() string {
		return "usr_test_admin"
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/v1/admin/audit-logs without auth = %d, want 401", rec.Code)
	}
}

func TestRESTCommit_PushProtection(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:8080",
		DataDir:    t.TempDir(),
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}
	app, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer app.Close()

	ns := store.Namespace{
		ID:   idgen.NewNS(),
		Slug: "protectteam",
		Type: "team",
		Name: "Protect Team",
	}
	if err := app.Store.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "gatekeeper",
		FullName:      "protectteam/gatekeeper",
		DefaultBranch: "main",
	}
	if err := app.Store.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	if _, err := app.Git.Init("protectteam", "gatekeeper", "main"); err != nil {
		t.Fatalf("Git.Init: %v", err)
	}
	sig := object.Signature{Name: "Admin", Email: "admin@refweir.dev", When: time.Now()}
	if _, err := app.Git.WriteFiles("protectteam", "gatekeeper", "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Gatekeeper\n"},
	}, sig, "init"); err != nil {
		t.Fatalf("WriteFiles: %v", err)
	}

	// 1. Commit containing AWS secret key -> rejected with 422 Unprocessable Entity
	secretPayload := `{"branch":"main","path":"credentials.txt","content":"AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\n","message":"leak"}`
	reqSecret := httptest.NewRequest(http.MethodPost, "/api/v1/repos/protectteam/gatekeeper/contents", strings.NewReader(secretPayload))
	recSecret := httptest.NewRecorder()
	app.Handler().ServeHTTP(recSecret, reqSecret)
	if recSecret.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for secret commit, got status %d: %s", recSecret.Code, recSecret.Body.String())
	}

	// 2. Clean commit to feature branch -> 201 Created
	cleanPayload := `{"branch":"feat-clean","path":"hello.txt","content":"clean code\n","message":"add hello"}`
	reqClean := httptest.NewRequest(http.MethodPost, "/api/v1/repos/protectteam/gatekeeper/contents", strings.NewReader(cleanPayload))
	recClean := httptest.NewRecorder()
	app.Handler().ServeHTTP(recClean, reqClean)
	if recClean.Code != http.StatusCreated {
		t.Fatalf("expected 201 for clean commit, got status %d: %s", recClean.Code, recClean.Body.String())
	}

	// 3. Add ruleset blocking direct pushes to main (requiring PR)
	rs := store.Ruleset{
		ID:               idgen.NewRuleset(),
		RepoID:           repo.ID,
		Name:             "protect-main",
		Enforcement:      "active",
		Kind:             "push_branch",
		IncludedRefNames: []string{"refs/heads/main"},
		Rules: []store.RulesetRule{
			{
				RuleType: "pull_request",
			},
		},
	}
	if _, err := app.Store.CreateRuleset(ctx, rs); err != nil {
		t.Fatalf("CreateRuleset: %v", err)
	}

	// 4. Direct commit to main branch is now blocked with 403 Forbidden
	directPayload := `{"branch":"main","path":"direct.txt","content":"direct edit\n","message":"bypass attempt"}`
	reqDirect := httptest.NewRequest(http.MethodPost, "/api/v1/repos/protectteam/gatekeeper/contents", strings.NewReader(directPayload))
	recDirect := httptest.NewRecorder()
	app.Handler().ServeHTTP(recDirect, reqDirect)
	if recDirect.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for direct commit to protected branch, got status %d: %s", recDirect.Code, recDirect.Body.String())
	}
}

