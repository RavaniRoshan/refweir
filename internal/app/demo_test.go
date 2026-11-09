package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"
	"refweir.dev/internal/audit"
	"refweir.dev/internal/config"
	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
)

func newDemoTestApp(t *testing.T, demoMode bool) *App {
	t.Helper()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:8080",
		DataDir:    t.TempDir(),
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
		JWTSecret:  "demo-test-secret-at-least-32-bytes-long!",
		DemoMode:   demoMode,
		DemoRepo:   "demo/agent-lab",
	}
	application, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := application.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})

	ctx := context.Background()
	nsID := idgen.NewNS()
	if err := application.Store.CreateNamespace(ctx, store.Namespace{
		ID:   nsID,
		Slug: "demo",
		Type: "team",
		Name: "Demo",
	}); err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	repoID := idgen.NewRepo()
	if err := application.Store.CreateRepo(ctx, store.Repo{
		ID:            repoID,
		NamespaceID:   nsID,
		Name:          "agent-lab",
		FullName:      "demo/agent-lab",
		DefaultBranch: "main",
		Visibility:    "public",
	}); err != nil {
		t.Fatalf("create repo: %v", err)
	}
	if _, err := application.Git.Init("demo", "agent-lab", "main"); err != nil {
		t.Fatalf("init repo: %v", err)
	}
	sig := object.Signature{Name: "Demo", Email: "demo@example.com", When: time.Now().UTC()}
	if _, err := application.Git.WriteFiles("demo", "agent-lab", "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Demo Agent Lab\n"},
	}, sig, "Initial commit"); err != nil {
		t.Fatalf("seed commit: %v", err)
	}
	return application
}

func TestDemoAgentRunCreatesRealPRAndAudit(t *testing.T) {
	var auditBuf bytes.Buffer
	prev := audit.DefaultLogger
	audit.SetDefaultLogger(audit.NewLogger(&auditBuf))
	t.Cleanup(func() { audit.SetDefaultLogger(prev) })

	application := newDemoTestApp(t, true)
	srv := httptest.NewServer(application.Handler())
	defer srv.Close()
	// The runner reaches its own Git transport through this URL.
	application.Config.DemoSelfURL = srv.URL

	req := httptest.NewRequest(http.MethodPost, "/api/v1/demo/agent-run", nil)
	req.RemoteAddr = "203.0.113.7:12345"
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var res demoAgentRunResult
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !res.OK || res.PRNumber <= 0 || res.Branch == "" || res.LeaseID == "" {
		t.Fatalf("unexpected demo result: %+v", res)
	}
	if !strings.HasPrefix(res.Branch, "agent/demo-") {
		t.Fatalf("branch %q is not an agent/* branch", res.Branch)
	}

	if _, err := application.Git.ResolveRef("demo", "agent-lab", res.Branch); err != nil {
		t.Fatalf("agent branch does not exist: %v", err)
	}
	repo, err := application.Store.GetRepo(context.Background(), "demo", "agent-lab")
	if err != nil {
		t.Fatalf("get repo: %v", err)
	}
	pr, err := application.Store.GetPullRequest(context.Background(), repo.ID, res.PRNumber)
	if err != nil {
		t.Fatalf("get pull request: %v", err)
	}
	if pr.HeadRef != res.Branch || pr.BaseRef != "main" {
		t.Fatalf("unexpected PR refs: %+v", pr)
	}
	if pr.HeadSHA == pr.BaseSHA {
		t.Fatalf("expected a new head commit, got base SHA for both")
	}

	logged := auditBuf.String()
	for _, want := range []string{"agent.lease_issued", "demo.agent_run", res.LeaseID} {
		if !strings.Contains(logged, want) {
			t.Fatalf("expected audit stream to contain %q, got: %s", want, logged)
		}
	}

	// A second immediate run is rate-limited.
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/demo/agent-run", nil)
	req2.RemoteAddr = "203.0.113.7:12345"
	rec2 := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second run status = %d, want 429", rec2.Code)
	}
}

func TestDemoRoutesAbsentWhenDemoModeDisabled(t *testing.T) {
	application := newDemoTestApp(t, false)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/demo/agent-run", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	// No demo handler is registered. The SPA fallback may serve index.html, but
	// the endpoint must never return a demo run result or create an agent branch.
	if strings.Contains(rec.Body.String(), "leaseId") || strings.Contains(rec.Body.String(), "prNumber") {
		t.Fatalf("demo endpoint responded while demo mode was off: %s", rec.Body.String())
	}
	branches, err := application.Git.ListBranches("demo", "agent-lab")
	if err != nil {
		t.Fatalf("list branches: %v", err)
	}
	for _, branch := range branches {
		if strings.HasPrefix(branch.Name, "agent/") {
			t.Fatalf("unexpected agent branch %q in a non-demo instance", branch.Name)
		}
	}
}

func TestDemoRateLimiter(t *testing.T) {
	l := newDemoRateLimiter()
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	if ok, _ := l.allow("a", now); !ok {
		t.Fatal("first call should be allowed")
	}
	if ok, wait := l.allow("a", now.Add(5*time.Second)); ok || wait <= 0 {
		t.Fatalf("call inside the 20s window should be denied with a wait, got ok=%v wait=%v", ok, wait)
	}
	if ok, _ := l.allow("b", now.Add(5*time.Second)); !ok {
		t.Fatal("a different client must not be limited by another client's hit")
	}
	if ok, _ := l.allow("a", now.Add(21*time.Second)); !ok {
		t.Fatal("call after the interval should be allowed")
	}
	// Exhaust the hourly budget, then confirm the sixth call is denied.
	base := now.Add(time.Minute)
	for i := 0; i < demoRunsPerHour-2; i++ {
		if ok, _ := l.allow("a", base.Add(time.Duration(i)*time.Minute)); !ok {
			t.Fatalf("call %d should be allowed", i)
		}
	}
	if ok, _ := l.allow("a", base.Add(10*time.Minute)); ok {
		t.Fatal("expected the hourly cap to deny the next call")
	}
}

func TestDemoCloneURLUsesLeaseCredentialAndLoopback(t *testing.T) {
	application := &App{Config: config.Config{
		ListenAddr: "0.0.0.0:9090",
		DemoRepo:   "demo/agent-lab",
	}}
	got, err := application.demoCloneURL("demo/agent-lab", "token-value")
	if err != nil {
		t.Fatalf("demoCloneURL: %v", err)
	}
	if !strings.HasPrefix(got, "http://agent:token-value@127.0.0.1:9090/demo/agent-lab.git") {
		t.Fatalf("unexpected clone URL: %s", got)
	}

	application.Config.DemoSelfURL = filepath.ToSlash("https://demo.refweir.dev/")
	got, err = application.demoCloneURL("demo/agent-lab", "tok")
	if err != nil {
		t.Fatalf("demoCloneURL with override: %v", err)
	}
	if !strings.HasPrefix(got, "https://agent:tok@demo.refweir.dev/demo/agent-lab.git") {
		t.Fatalf("unexpected clone URL with override: %s", got)
	}
}
