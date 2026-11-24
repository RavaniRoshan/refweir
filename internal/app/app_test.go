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
	"refweir.dev/internal/auth"
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

	// 5a-1. Admin audit logs export endpoint (NDJSON format)
	reqExportNDJSON := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export?format=ndjson", nil)
	recExportNDJSON := httptest.NewRecorder()
	application.Handler().ServeHTTP(recExportNDJSON, reqExportNDJSON)
	if recExportNDJSON.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/audit-logs/export ndjson status = %d, want 200", recExportNDJSON.Code)
	}
	if ct := recExportNDJSON.Header().Get("Content-Type"); !strings.Contains(ct, "ndjson") {
		t.Errorf("expected Content-Type containing ndjson, got: %s", ct)
	}

	// 5a-2. Admin audit logs export endpoint (CSV format)
	reqExportCSV := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export?format=csv", nil)
	recExportCSV := httptest.NewRecorder()
	application.Handler().ServeHTTP(recExportCSV, reqExportCSV)
	if recExportCSV.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/audit-logs/export csv status = %d, want 200", recExportCSV.Code)
	}
	if ct := recExportCSV.Header().Get("Content-Type"); !strings.Contains(ct, "csv") {
		t.Errorf("expected Content-Type containing csv, got: %s", ct)
	}
	if !strings.Contains(recExportCSV.Body.String(), "id,timestamp,actor") {
		t.Errorf("expected CSV header row in audit export")
	}

	// 5a-3. Admin audit logs export endpoint (JSON format)
	reqExportJSON := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/export?format=json", nil)
	recExportJSON := httptest.NewRecorder()
	application.Handler().ServeHTTP(recExportJSON, reqExportJSON)
	if recExportJSON.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/audit-logs/export json status = %d, want 200", recExportJSON.Code)
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

func TestCapabilityLease_AuthorizationGating(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:8080",
		DataDir:    t.TempDir(),
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
		JWTSecret:  "test-lease-jwt-secret-key-12345",
	}
	app, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer app.Close()

	// 1. Create admin and member users
	adminUser := store.User{
		ID:        idgen.NewUser(),
		Login:     "admincorp",
		Email:     "admin@corp.dev",
		Role:      "admin",
		Status:    "active",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	_ = app.Store.CreateUserWithPassword(ctx, adminUser)
	rawAdminKey, adminPfx, adminHash, _ := auth.GenerateAPIKey()
	_ = app.Store.CreateAPIKey(ctx, store.APIKey{
		ID:        idgen.New("key"),
		UserID:    adminUser.ID,
		KeyHash:   adminHash,
		KeyPrefix: adminPfx,
		Name:      "admin-key",
		Scopes:    "*",
	})

	memberUser := store.User{
		ID:        idgen.NewUser(),
		Login:     "memberbob",
		Email:     "bob@corp.dev",
		Role:      "member",
		Status:    "active",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	_ = app.Store.CreateUserWithPassword(ctx, memberUser)
	rawMemberKey, memberPfx, memberHash, _ := auth.GenerateAPIKey()
	_ = app.Store.CreateAPIKey(ctx, store.APIKey{
		ID:        idgen.New("key"),
		UserID:    memberUser.ID,
		KeyHash:   memberHash,
		KeyPrefix: memberPfx,
		Name:      "member-key",
		Scopes:    "repo,write",
	})

	// 2. Create namespace and private repo owned by admin
	ns := store.Namespace{
		ID:          idgen.NewNS(),
		Slug:        "privcorp",
		Type:        "team",
		Name:        "Private Corp",
		OwnerUserID: adminUser.ID,
	}
	_ = app.Store.CreateNamespace(ctx, ns)
	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "vault",
		FullName:      "privcorp/vault",
		DefaultBranch: "main",
		Visibility:    "private",
	}
	_ = app.Store.CreateRepo(ctx, repo)

	// A. Unauthenticated request when users exist -> 401 Unauthorized
	reqUnauth := httptest.NewRequest(http.MethodPost, "/api/v1/agent/leases", strings.NewReader(`{"agentId":"bot","repoId":"privcorp/vault"}`))
	recUnauth := httptest.NewRecorder()
	app.Handler().ServeHTTP(recUnauth, reqUnauth)
	if recUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated lease request, got %d", recUnauth.Code)
	}

	// B. Member user attempts to grant capability lease on private repo without access -> 403 Forbidden
	reqMemberNoAccess := httptest.NewRequest(http.MethodPost, "/api/v1/agent/leases", strings.NewReader(`{"agentId":"bot","repoId":"privcorp/vault","allowedRefPrefixes":["refs/heads/agent/"]}`))
	reqMemberNoAccess.Header.Set("Authorization", "Bearer "+rawMemberKey)
	recMemberNoAccess := httptest.NewRecorder()
	app.Handler().ServeHTTP(recMemberNoAccess, reqMemberNoAccess)
	if recMemberNoAccess.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for member without repo access, got %d", recMemberNoAccess.Code)
	}

	// C. Grant member write access to repo, but member tries to grant direct refs/heads/main access -> 403 Forbidden
	_, _ = app.Store.UpsertGrant(ctx, store.Grant{
		ID:            idgen.New("grant"),
		Scope:         "repository",
		RepoID:        repo.ID,
		NamespaceID:   ns.ID,
		PrincipalKind: "user",
		UserID:        memberUser.ID,
		Permission:    "write",
	})
	reqMemberEscalate := httptest.NewRequest(http.MethodPost, "/api/v1/agent/leases", strings.NewReader(`{"agentId":"bot","repoId":"privcorp/vault","allowedRefPrefixes":["refs/heads/main"]}`))
	reqMemberEscalate.Header.Set("Authorization", "Bearer "+rawMemberKey)
	recMemberEscalate := httptest.NewRecorder()
	app.Handler().ServeHTTP(recMemberEscalate, reqMemberEscalate)
	if recMemberEscalate.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for member attempting to grant main access directly, got %d", recMemberEscalate.Code)
	}

	// D. Member issues scoped lease to refs/heads/agent/ -> 200 OK
	reqMemberOk := httptest.NewRequest(http.MethodPost, "/api/v1/agent/leases", strings.NewReader(`{"agentId":"bot","repoId":"privcorp/vault","allowedRefPrefixes":["refs/heads/agent/"]}`))
	reqMemberOk.Header.Set("Authorization", "Bearer "+rawMemberKey)
	recMemberOk := httptest.NewRecorder()
	app.Handler().ServeHTTP(recMemberOk, reqMemberOk)
	if recMemberOk.Code != http.StatusOK {
		t.Fatalf("expected 200 for member scoped lease, got %d: %s", recMemberOk.Code, recMemberOk.Body.String())
	}

	// E. Admin issues lease granting refs/heads/main -> 200 OK
	reqAdminMain := httptest.NewRequest(http.MethodPost, "/api/v1/agent/leases", strings.NewReader(`{"agentId":"admin-bot","repoId":"privcorp/vault","allowedRefPrefixes":["refs/heads/main"]}`))
	reqAdminMain.Header.Set("Authorization", "Bearer "+rawAdminKey)
	recAdminMain := httptest.NewRecorder()
	app.Handler().ServeHTTP(recAdminMain, reqAdminMain)
	if recAdminMain.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin lease to main, got %d: %s", recAdminMain.Code, recAdminMain.Body.String())
	}
}

func TestApp_ForkEndpoints(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:8080",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}

	app, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer app.Close()
	ctx := context.Background()

	ns := store.Namespace{ID: idgen.NewNS(), Slug: "upstream-org", Type: "team", Name: "upstream-org"}
	if err := app.Store.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "engine",
		FullName:      "upstream-org/engine",
		DefaultBranch: "main",
		Visibility:    "public",
		Description:   "core engine",
	}
	if err := app.Store.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	if _, err := app.Git.Init("upstream-org", "engine", "main"); err != nil {
		t.Fatalf("Git.Init: %v", err)
	}
	sig := object.Signature{Name: "Dev", Email: "dev@upstream.org", When: time.Now()}
	if _, err := app.Git.WriteFiles("upstream-org", "engine", "main", []gitstore.FileChange{
		{Path: "main.go", Content: "package main\n"},
	}, sig, "init"); err != nil {
		t.Fatalf("WriteFiles: %v", err)
	}

	// 1. POST /api/v1/repos/upstream-org/engine/forks
	forkBody := `{"organization":"fork-team","name":"custom-engine"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/repos/upstream-org/engine/forks", strings.NewReader(forkBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d: %s", rec.Code, rec.Body.String())
	}

	var forkedRepo store.Repo
	if err := json.Unmarshal(rec.Body.Bytes(), &forkedRepo); err != nil {
		t.Fatalf("decode fork response: %v", err)
	}
	if forkedRepo.FullName != "fork-team/custom-engine" || !forkedRepo.IsFork || forkedRepo.ParentRepoID != repo.ID {
		t.Errorf("unexpected forked repo metadata: %+v", forkedRepo)
	}

	// 2. GET /api/v1/repos/upstream-org/engine/forks
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/repos/upstream-org/engine/forks", nil)
	recGet := httptest.NewRecorder()
	app.Handler().ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", recGet.Code, recGet.Body.String())
	}
	var forksList []store.Repo
	if err := json.Unmarshal(recGet.Body.Bytes(), &forksList); err != nil {
		t.Fatalf("decode forks list: %v", err)
	}
	if len(forksList) != 1 || forksList[0].ID != forkedRepo.ID {
		t.Errorf("expected 1 fork with ID %s, got %v", forkedRepo.ID, forksList)
	}

	// 3. Prevent forking into itself
	selfForkBody := `{"organization":"upstream-org","name":"engine"}`
	reqSelf := httptest.NewRequest(http.MethodPost, "/api/v1/repos/upstream-org/engine/forks", strings.NewReader(selfForkBody))
	recSelf := httptest.NewRecorder()
	app.Handler().ServeHTTP(recSelf, reqSelf)
	if recSelf.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for self-fork, got %d", recSelf.Code)
	}
}

func TestApp_SyncForkEndpoint(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:8080",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}

	app, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer app.Close()
	ctx := context.Background()

	nsUpstream := store.Namespace{ID: idgen.NewNS(), Slug: "upstream-dev", Type: "team", Name: "upstream-dev"}
	_ = app.Store.CreateNamespace(ctx, nsUpstream)
	upstreamRepo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   nsUpstream.ID,
		Name:          "core",
		FullName:      "upstream-dev/core",
		DefaultBranch: "main",
		Visibility:    "public",
	}
	_ = app.Store.CreateRepo(ctx, upstreamRepo)
	_, _ = app.Git.Init("upstream-dev", "core", "main")
	sig := object.Signature{Name: "Dev", Email: "dev@upstream.org", When: time.Now()}
	_, _ = app.Git.WriteFiles("upstream-dev", "core", "main", []gitstore.FileChange{
		{Path: "readme.txt", Content: "v1\n"},
	}, sig, "v1")

	// Fork into contributor-dev/core
	nsFork := store.Namespace{ID: idgen.NewNS(), Slug: "contributor-dev", Type: "user", Name: "contributor-dev"}
	_ = app.Store.CreateNamespace(ctx, nsFork)
	forkRepo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   nsFork.ID,
		Name:          "core",
		FullName:      "contributor-dev/core",
		DefaultBranch: "main",
		Visibility:    "public",
		ParentRepoID:  upstreamRepo.ID,
		IsFork:        true,
	}
	_ = app.Store.CreateRepo(ctx, forkRepo)
	_, _ = app.Git.Fork("upstream-dev", "core", "contributor-dev", "core")

	// 1. Non-fork sync should return 400
	reqNonFork := httptest.NewRequest(http.MethodPost, "/api/v1/repos/upstream-dev/core/forks/sync", nil)
	recNonFork := httptest.NewRecorder()
	app.Handler().ServeHTTP(recNonFork, reqNonFork)
	if recNonFork.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-fork repo sync, got %d", recNonFork.Code)
	}

	// 2. Advance upstream
	_, _ = app.Git.WriteFiles("upstream-dev", "core", "main", []gitstore.FileChange{
		{Path: "feature.txt", Content: "v2 feature\n"},
	}, sig, "v2")

	// 3. Sync fork via REST
	syncBody := `{"branch":"main"}`
	reqSync := httptest.NewRequest(http.MethodPost, "/api/v1/repos/contributor-dev/core/forks/sync", strings.NewReader(syncBody))
	reqSync.Header.Set("Content-Type", "application/json")
	recSync := httptest.NewRecorder()
	app.Handler().ServeHTTP(recSync, reqSync)

	if recSync.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", recSync.Code, recSync.Body.String())
	}

	var res gitstore.SyncForkResult
	if err := json.Unmarshal(recSync.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode sync response: %v", err)
	}
	if res.MergeType != "fast-forward" {
		t.Errorf("expected merge_type 'fast-forward', got %q", res.MergeType)
	}
}



