package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCLI_VersionAndHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	c := &CLI{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	if err := c.Run([]string{"--version"}); err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(stdout.String(), "refweir version") {
		t.Errorf("expected version string, got: %s", stdout.String())
	}

	stdout.Reset()
	if err := c.Run([]string{"--help"}); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(stdout.String(), "refweir <command>") {
		t.Errorf("expected help string, got: %s", stdout.String())
	}
}

func TestCLI_RepoCommands(t *testing.T) {
	var requestedPath, requestedMethod string
	var requestedBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		requestedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "test_tok_123",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	// 1. repo list
	if err := c.Run([]string{"repo", "list", "--namespace", "acme"}); err != nil {
		t.Fatalf("repo list: %v", err)
	}
	if requestedPath != "/repos/acme" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /repos/acme, got %s %s", requestedMethod, requestedPath)
	}

	// 2. repo view and repo get alias
	if err := c.Run([]string{"repo", "view", "acme/rocket"}); err != nil {
		t.Fatalf("repo view: %v", err)
	}
	if requestedPath != "/repos/acme/rocket" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /repos/acme/rocket, got %s %s", requestedMethod, requestedPath)
	}

	if err := c.Run([]string{"repo", "get", "acme/rocket"}); err != nil {
		t.Fatalf("repo get: %v", err)
	}
	if requestedPath != "/repos/acme/rocket" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /repos/acme/rocket for get alias, got %s %s", requestedMethod, requestedPath)
	}

	// 3. repo create
	if err := c.Run([]string{"repo", "create", "acme/rocket", "--default-branch", "develop"}); err != nil {
		t.Fatalf("repo create: %v", err)
	}
	if requestedPath != "/repos/acme" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /repos/acme, got %s %s", requestedMethod, requestedPath)
	}
	var createReq map[string]any
	_ = json.Unmarshal(requestedBody, &createReq)
	repoMap, ok := createReq["repo"].(map[string]any)
	if !ok || repoMap["name"] != "rocket" || repoMap["defaultBranch"] != "develop" {
		t.Errorf("unexpected body: %+v", createReq)
	}

	// 4. repo mirror sync
	if err := c.Run([]string{"repo", "mirror", "sync", "acme/rocket", "--ref", "refs/heads/main"}); err != nil {
		t.Fatalf("repo mirror sync: %v", err)
	}
	if requestedPath != "/repos/acme/rocket:syncMirror" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /repos/acme/rocket:syncMirror, got %s %s", requestedMethod, requestedPath)
	}
}

func TestCLI_PRCommands(t *testing.T) {
	var requestedPath, requestedMethod string
	var requestedBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		requestedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:    ts.URL,
		Token:       "test_tok_123",
		DefaultRepo: "acme/rocket",
		HTTPClient:  ts.Client(),
		Stdout:      &stdout,
		Stderr:      &stderr,
	}

	// 1. pr list
	if err := c.Run([]string{"pr", "list"}); err != nil {
		t.Fatalf("pr list: %v", err)
	}
	if requestedPath != "/repos/acme/rocket/pulls" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /repos/acme/rocket/pulls, got %s %s", requestedMethod, requestedPath)
	}

	// 2. pr merge posts to .../merge
	if err := c.Run([]string{"pr", "merge", "42", "--method", "squash"}); err != nil {
		t.Fatalf("pr merge: %v", err)
	}
	if requestedPath != "/repos/acme/rocket/pulls/42/merge" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /repos/acme/rocket/pulls/42/merge, got %s %s", requestedMethod, requestedPath)
	}
	var mergeReq map[string]any
	_ = json.Unmarshal(requestedBody, &mergeReq)
	if mergeReq["mergeMethod"] != "squash" {
		t.Errorf("expected mergeMethod squash, got %+v", mergeReq)
	}

	// 3. pr close sends PATCH .../pulls/42 with {"state":"closed"}
	if err := c.Run([]string{"pr", "close", "42"}); err != nil {
		t.Fatalf("pr close: %v", err)
	}
	if requestedPath != "/repos/acme/rocket/pulls/42" || requestedMethod != http.MethodPatch {
		t.Errorf("expected PATCH /repos/acme/rocket/pulls/42, got %s %s", requestedMethod, requestedPath)
	}
	var closeReq map[string]any
	_ = json.Unmarshal(requestedBody, &closeReq)
	if closeReq["state"] != "closed" {
		t.Errorf("expected state closed, got %+v", closeReq)
	}

	// 4. pr list with --state
	var requestedRawQuery string
	tsQuery := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer tsQuery.Close()
	c.Endpoint = tsQuery.URL
	c.HTTPClient = tsQuery.Client()

	if err := c.Run([]string{"pr", "list", "--state", "closed"}); err != nil {
		t.Fatalf("pr list --state: %v", err)
	}
	if requestedRawQuery != "state=closed" {
		t.Errorf("expected query state=closed, got %s", requestedRawQuery)
	}
}

func TestCLI_APICommand(t *testing.T) {
	var requestedPath, requestedMethod string
	var requestedHeader string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		requestedHeader = r.Header.Get("Authorization")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"rateLimit": {"limit": 5000}}`))
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "my_secret_token",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	if err := c.Run([]string{"api", "/rate_limit"}); err != nil {
		t.Fatalf("api: %v", err)
	}
	if requestedPath != "/rate_limit" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /rate_limit, got %s %s", requestedMethod, requestedPath)
	}
	if requestedHeader != "Bearer my_secret_token" {
		t.Errorf("expected bearer token header, got: %s", requestedHeader)
	}
}

func TestCLI_LeaseCommands(t *testing.T) {
	var requestedPath, requestedMethod string
	var requestedBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		requestedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/revoke") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok": true}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "lease_cli_123",
			"token": "signed_token_abc",
			"agentId": "agent:test-worker",
			"repoId": "acme/rocket",
			"issuedAt": "2026-11-24T00:00:00Z",
			"expiresAt": "2026-11-24T00:30:00Z"
		}`))
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "tok_123",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	// 1. lease issue formatted
	if err := c.Run([]string{"lease", "issue", "--agent", "agent:test-worker", "-R", "acme/rocket", "--ttl", "30m"}); err != nil {
		t.Fatalf("lease issue: %v", err)
	}
	if requestedPath != "/api/v1/agent/leases" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/agent/leases, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "Capability Lease Issued: lease_cli_123") {
		t.Errorf("expected lease ID in stdout, got: %s", stdout.String())
	}

	// 2. lease issue json
	stdout.Reset()
	if err := c.Run([]string{"lease", "issue", "--agent", "agent:test-worker", "-R", "acme/rocket", "--json"}); err != nil {
		t.Fatalf("lease issue --json: %v", err)
	}
	if !strings.Contains(stdout.String(), `"lease_cli_123"`) {
		t.Errorf("expected raw json output, got: %s", stdout.String())
	}

	// 3. lease revoke
	stdout.Reset()
	if err := c.Run([]string{"lease", "revoke", "lease_cli_123", "--reason", "security rotation"}); err != nil {
		t.Fatalf("lease revoke: %v", err)
	}
	if requestedPath != "/api/v1/agent/leases/lease_cli_123/revoke" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/agent/leases/lease_cli_123/revoke, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "Revoked capability lease lease_cli_123") {
		t.Errorf("expected revoked message in stdout, got: %s", stdout.String())
	}
	var revokeBody map[string]any
	_ = json.Unmarshal(requestedBody, &revokeBody)
	if revokeBody["reason"] != "security rotation" {
		t.Errorf("expected reason 'security rotation', got: %v", revokeBody["reason"])
	}
}

func TestCLI_IssueCommands(t *testing.T) {
	var requestedPath, requestedMethod string
	var requestedBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		requestedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/issues") {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"ok": true, "issue": {"number": 1, "title": "Bug: Buffer overflow"}}`))
			return
		}
		if r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/v1/repos/acme/rocket/issues/") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok": true, "state": "closed"}`))
			return
		}
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/comments") {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"ok": true, "comment": {"id": "c1", "body": "Fixed in commit abc"}}`))
			return
		}
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/issues/1") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"number": 1, "title": "Bug: Buffer overflow", "body": "Details here", "state": "open", "author": "alice"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"issues": [{"number": 1, "title": "Bug: Buffer overflow", "state": "open", "author": "alice", "comments": 2}]}`))
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "tok_issue",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	// 1. issue list
	if err := c.Run([]string{"issue", "list", "-R", "acme/rocket", "--state", "open"}); err != nil {
		t.Fatalf("issue list: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/issues" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /api/v1/repos/acme/rocket/issues, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "#1") {
		t.Errorf("expected issue #1 in stdout: %s", stdout.String())
	}

	// 2. issue create
	stdout.Reset()
	if err := c.Run([]string{"issue", "create", "-R", "acme/rocket", "--title", "Bug: Buffer overflow", "--body", "Details here"}); err != nil {
		t.Fatalf("issue create: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/issues" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/repos/acme/rocket/issues, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "Created issue #1") {
		t.Errorf("expected created message in stdout: %s", stdout.String())
	}
	var createBody map[string]any
	_ = json.Unmarshal(requestedBody, &createBody)
	if createBody["title"] != "Bug: Buffer overflow" {
		t.Errorf("expected title 'Bug: Buffer overflow', got: %v", createBody["title"])
	}

	// 3. issue view
	stdout.Reset()
	if err := c.Run([]string{"issue", "view", "1", "-R", "acme/rocket"}); err != nil {
		t.Fatalf("issue view: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/issues/1" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /api/v1/repos/acme/rocket/issues/1, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "#1 Bug: Buffer overflow") {
		t.Errorf("expected issue title in stdout: %s", stdout.String())
	}

	// 4. issue comment
	stdout.Reset()
	if err := c.Run([]string{"issue", "comment", "1", "-R", "acme/rocket", "--body", "Fixed in commit abc"}); err != nil {
		t.Fatalf("issue comment: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/issues/1/comments" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/repos/acme/rocket/issues/1/comments, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "Added comment to issue #1") {
		t.Errorf("expected comment confirmation in stdout: %s", stdout.String())
	}

	// 5. issue close
	stdout.Reset()
	if err := c.Run([]string{"issue", "close", "1", "-R", "acme/rocket"}); err != nil {
		t.Fatalf("issue close: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/issues/1" || requestedMethod != http.MethodPatch {
		t.Errorf("expected PATCH /api/v1/repos/acme/rocket/issues/1, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "Closed issue #1") {
		t.Errorf("expected closed message in stdout: %s", stdout.String())
	}
}

func TestCLI_WebhookCommands(t *testing.T) {
	var requestedPath, requestedMethod string
	var requestedBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		requestedBody, _ = io.ReadAll(r.Body)

		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/hooks") && r.Method == http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":"whk_01","url":"https://ci.example.com","events":"[\"push\"]","active":true,"insecure_ssl":false}]`))
		case strings.HasSuffix(r.URL.Path, "/hooks") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"whk_02","url":"https://ci.example.com/deploy"}`))
		case strings.HasSuffix(r.URL.Path, "/tests") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true,"delivery_id":"del_01","status":"delivered"}`))
		case strings.HasSuffix(r.URL.Path, "/deliveries") && r.Method == http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":"del_01","event":"push","delivered":true,"response_status_code":200,"duration_ms":42,"created_at":"2026-11-24T00:00:00Z"}]`))
		case strings.Contains(r.URL.Path, "/redeliver") && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true,"redelivery_id":"del_02"}`))
		case strings.Contains(r.URL.Path, "/hooks/") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true,"deleted":"whk_01"}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "test_tok_whk",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	// 1. webhook list
	if err := c.Run([]string{"webhook", "list", "-R", "acme/rocket"}); err != nil {
		t.Fatalf("webhook list: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/hooks" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /api/v1/repos/acme/rocket/hooks, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "whk_01") {
		t.Errorf("expected whk_01 in stdout: %s", stdout.String())
	}

	// 2. webhook create
	stdout.Reset()
	if err := c.Run([]string{"webhook", "create", "-R", "acme/rocket", "--url", "https://ci.example.com/deploy", "--secret", "supersecret", "--events", "push,pull_request"}); err != nil {
		t.Fatalf("webhook create: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/hooks" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/repos/acme/rocket/hooks, got %s %s", requestedMethod, requestedPath)
	}
	var createdReq map[string]any
	_ = json.Unmarshal(requestedBody, &createdReq)
	if createdReq["url"] != "https://ci.example.com/deploy" {
		t.Errorf("unexpected url in body: %v", createdReq["url"])
	}
	if !strings.Contains(stdout.String(), "Created webhook whk_02") {
		t.Errorf("expected creation confirmation in stdout: %s", stdout.String())
	}

	// 3. webhook ping
	stdout.Reset()
	if err := c.Run([]string{"webhook", "ping", "whk_01", "-R", "acme/rocket"}); err != nil {
		t.Fatalf("webhook ping: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/hooks/whk_01/tests" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/repos/acme/rocket/hooks/whk_01/tests, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "Ping dispatched") {
		t.Errorf("expected ping confirmation in stdout: %s", stdout.String())
	}

	// 4. webhook deliveries
	stdout.Reset()
	if err := c.Run([]string{"webhook", "deliveries", "whk_01", "-R", "acme/rocket"}); err != nil {
		t.Fatalf("webhook deliveries: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/hooks/whk_01/deliveries" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /api/v1/repos/acme/rocket/hooks/whk_01/deliveries, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "del_01") {
		t.Errorf("expected del_01 in stdout: %s", stdout.String())
	}

	// 5. webhook redeliver
	stdout.Reset()
	if err := c.Run([]string{"webhook", "redeliver", "whk_01", "del_01", "-R", "acme/rocket"}); err != nil {
		t.Fatalf("webhook redeliver: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/hooks/whk_01/deliveries/del_01/redeliver" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/repos/acme/rocket/hooks/whk_01/deliveries/del_01/redeliver, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "Redelivery triggered successfully") {
		t.Errorf("expected redelivery confirmation in stdout: %s", stdout.String())
	}

	// 6. webhook delete
	stdout.Reset()
	if err := c.Run([]string{"webhook", "delete", "whk_01", "-R", "acme/rocket"}); err != nil {
		t.Fatalf("webhook delete: %v", err)
	}
	if requestedPath != "/api/v1/repos/acme/rocket/hooks/whk_01" || requestedMethod != http.MethodDelete {
		t.Errorf("expected DELETE /api/v1/repos/acme/rocket/hooks/whk_01, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "Deleted webhook whk_01") {
		t.Errorf("expected delete confirmation in stdout: %s", stdout.String())
	}
}
