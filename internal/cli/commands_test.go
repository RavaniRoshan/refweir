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

	// 4. lease issue with asymmetric Ed25519 (fresh generated key)
	stdout.Reset()
	requestedPath = ""
	if err := c.Run([]string{"lease", "issue", "--agent", "agent:bot-ed", "-R", "acme/rocket", "--algorithm", "ed25519", "--ttl", "15m"}); err != nil {
		t.Fatalf("lease issue --algorithm ed25519: %v", err)
	}
	if requestedPath != "" {
		t.Errorf("expected local issuance without HTTP request, but got request to %s", requestedPath)
	}
	outEd := stdout.String()
	if !strings.Contains(outEd, "Algorithm:  ed25519") || !strings.Contains(outEd, "Token:      ed25519.") {
		t.Errorf("expected ed25519 token output, got: %s", outEd)
	}
	var pubKeyStr, tokenStr string
	for _, line := range strings.Split(outEd, "\n") {
		if strings.HasPrefix(line, "Public Key:") {
			pubKeyStr = strings.TrimSpace(strings.TrimPrefix(line, "Public Key:"))
		}
		if strings.HasPrefix(line, "Token:") {
			tokenStr = strings.TrimSpace(strings.TrimPrefix(line, "Token:"))
		}
	}
	if pubKeyStr == "" || tokenStr == "" {
		t.Fatalf("failed to extract public key or token from output: %s", outEd)
	}

	// 5. lease verify Ed25519 token with extracted public key
	stdout.Reset()
	if err := c.Run([]string{"lease", "verify", tokenStr, "--pubkey", pubKeyStr}); err != nil {
		t.Fatalf("lease verify ed25519 token: %v", err)
	}
	if !strings.Contains(stdout.String(), "VALID") || !strings.Contains(stdout.String(), "ed25519") {
		t.Errorf("expected VALID ed25519 verification, got: %s", stdout.String())
	}

	// 6. lease issue with local HMAC secret via --privkey and verify
	stdout.Reset()
	hmacSecret := "super-secure-cluster-hmac-secret-12345"
	if err := c.Run([]string{"lease", "issue", "--agent", "agent:bot-hmac", "-R", "acme/rocket", "--privkey", hmacSecret, "--json"}); err != nil {
		t.Fatalf("lease issue --privkey hmac: %v", err)
	}
	var hmacResp struct {
		Token     string `json:"token"`
		ID        string `json:"id"`
		Algorithm string `json:"algorithm"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &hmacResp); err != nil {
		t.Fatalf("failed to unmarshal hmac json: %v, raw: %s", err, stdout.String())
	}
	if hmacResp.Algorithm != "hmac-sha256" || strings.HasPrefix(hmacResp.Token, "ed25519.") {
		t.Errorf("unexpected hmac response: %+v", hmacResp)
	}

	stdout.Reset()
	if err := c.Run([]string{"lease", "verify", hmacResp.Token, "--secret", hmacSecret, "--json"}); err != nil {
		t.Fatalf("lease verify hmac token: %v", err)
	}
	if !strings.Contains(stdout.String(), `"valid": true`) {
		t.Errorf("expected valid hmac verification, got: %s", stdout.String())
	}

	// 7. lease verify with invalid signature fails
	stdout.Reset()
	if err := c.Run([]string{"lease", "verify", hmacResp.Token, "--secret", "wrong-secret"}); err == nil {
		t.Errorf("expected verification failure with wrong secret, but got nil error")
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

func TestCLI_Doctor(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/healthz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		case "/api/v1/admin/health":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"serverStatus": "healthy",
				"version": "refweird 1.0.0",
				"uptime": "120s",
				"goroutines": 15,
				"sqliteWalSizeBytes": 65536,
				"allocatedMemoryMb": "8.5"
			}`))
		case "/api/v1/auth/session":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"user": {
					"id": "usr_001",
					"login": "alice",
					"role": "admin"
				}
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "test_tok_admin",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	// 1. Text report
	if err := c.Run([]string{"doctor"}); err != nil {
		t.Fatalf("doctor: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "Refweir Diagnostic Report") {
		t.Errorf("expected title in report, got: %s", out)
	}
	if !strings.Contains(out, "Daemon Connectivity") || !strings.Contains(out, "alice") {
		t.Errorf("expected connectivity and user alice in report: %s", out)
	}

	// 2. JSON report
	stdout.Reset()
	if err := c.Run([]string{"doctor", "--json"}); err != nil {
		t.Fatalf("doctor --json: %v", err)
	}
	var docReport map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &docReport); err != nil {
		t.Fatalf("failed to parse doctor JSON output: %v, raw: %s", err, stdout.String())
	}
	if docReport["healthy"] != true {
		t.Errorf("expected healthy=true, got %+v", docReport)
	}
}

func TestCLI_Completion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	c := &CLI{
		Stdout: &stdout,
		Stderr: &stderr,
	}

	// 1. bash
	if err := c.Run([]string{"completion", "bash"}); err != nil {
		t.Fatalf("completion bash: %v", err)
	}
	if !strings.Contains(stdout.String(), "complete -F _refweir_completion refweir") {
		t.Errorf("expected bash completion script, got: %s", stdout.String())
	}

	// 2. zsh
	stdout.Reset()
	if err := c.Run([]string{"completion", "zsh"}); err != nil {
		t.Fatalf("completion zsh: %v", err)
	}
	if !strings.Contains(stdout.String(), "#compdef refweir") {
		t.Errorf("expected zsh completion script, got: %s", stdout.String())
	}

	// 3. fish
	stdout.Reset()
	if err := c.Run([]string{"completion", "fish"}); err != nil {
		t.Fatalf("completion fish: %v", err)
	}
	if !strings.Contains(stdout.String(), "complete -c refweir") {
		t.Errorf("expected fish completion script, got: %s", stdout.String())
	}

	// 4. unsupported shell
	if err := c.Run([]string{"completion", "powershell"}); err == nil {
		t.Fatalf("expected error for unsupported shell, got nil")
	}
}

func TestCLI_Audit(t *testing.T) {
	var requestedPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"events": [
				{
					"id": "aud_01",
					"actor": "alice",
					"action": "agent.lease_issued",
					"resource": "acme/rocket",
					"severity": "info",
					"timestamp": "2026-11-24T06:00:00Z",
					"details": "Capability lease issued"
				}
			]
		}`))
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "test_tok_admin",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	// 1. audit list formatted
	if err := c.Run([]string{"audit", "list", "--actor", "alice", "--action", "agent.lease_issued", "--limit", "10"}); err != nil {
		t.Fatalf("audit list: %v", err)
	}
	if !strings.Contains(requestedPath, "actor=alice") || !strings.Contains(requestedPath, "action=agent.lease_issued") {
		t.Errorf("expected query params in requested path, got: %s", requestedPath)
	}
	if !strings.Contains(stdout.String(), "aud_01") || !strings.Contains(stdout.String(), "agent.lease_issued") {
		t.Errorf("expected aud_01 in stdout: %s", stdout.String())
	}

	// 2. audit list json
	stdout.Reset()
	if err := c.Run([]string{"audit", "list", "--json"}); err != nil {
		t.Fatalf("audit list --json: %v", err)
	}
	var auditResp map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &auditResp); err != nil {
		t.Fatalf("failed to parse audit json: %v", err)
	}
	events, ok := auditResp["events"].([]any)
	if !ok || len(events) != 1 {
		t.Errorf("expected 1 audit event, got: %+v", auditResp)
	}
}

func TestCLI_ClusterCommands(t *testing.T) {
	var requestedPath, requestedMethod string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/cluster/replication":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"nodeId": "node-1", "role": "primary", "health": "healthy", "seq": 42}`))
		case "/api/v1/cluster/checkpoint":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok": true, "status": "checkpointed"}`))
		case "/api/v1/cluster/snapshot":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"snapshotId": "snap_01", "sizeBytes": 65536, "checksumSha256": "abc123"}`))
		case "/api/v1/cluster/failover/promote":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "promoted", "role": "primary"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	var stdout, stderr bytes.Buffer
	c := &CLI{
		Endpoint:   ts.URL,
		Token:      "test_tok_admin",
		HTTPClient: ts.Client(),
		Stdout:     &stdout,
		Stderr:     &stderr,
	}

	// 1. cluster status
	if err := c.Run([]string{"cluster", "status"}); err != nil {
		t.Fatalf("cluster status: %v", err)
	}
	if requestedPath != "/api/v1/cluster/replication" || requestedMethod != http.MethodGet {
		t.Errorf("expected GET /api/v1/cluster/replication, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "primary") || !strings.Contains(stdout.String(), "healthy") {
		t.Errorf("expected primary and healthy in stdout: %s", stdout.String())
	}

	// 2. cluster checkpoint
	stdout.Reset()
	if err := c.Run([]string{"cluster", "checkpoint"}); err != nil {
		t.Fatalf("cluster checkpoint: %v", err)
	}
	if requestedPath != "/api/v1/cluster/checkpoint" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/cluster/checkpoint, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "checkpointed") {
		t.Errorf("expected checkpointed in stdout: %s", stdout.String())
	}

	// 3. cluster snapshot
	stdout.Reset()
	if err := c.Run([]string{"cluster", "snapshot"}); err != nil {
		t.Fatalf("cluster snapshot: %v", err)
	}
	if requestedPath != "/api/v1/cluster/snapshot" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/cluster/snapshot, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "snap_01") {
		t.Errorf("expected snap_01 in stdout: %s", stdout.String())
	}

	// 4. cluster promote
	stdout.Reset()
	if err := c.Run([]string{"cluster", "promote", "--yes"}); err != nil {
		t.Fatalf("cluster promote: %v", err)
	}
	if requestedPath != "/api/v1/cluster/failover/promote" || requestedMethod != http.MethodPost {
		t.Errorf("expected POST /api/v1/cluster/failover/promote, got %s %s", requestedMethod, requestedPath)
	}
	if !strings.Contains(stdout.String(), "promoted") {
		t.Errorf("expected promoted in stdout: %s", stdout.String())
	}
}
