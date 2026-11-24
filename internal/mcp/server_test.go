package mcp

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
	"refweir.dev/internal/auth"
	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
)

func setupTestMCP(t *testing.T) (*Server, *store.DB, *gitstore.Store) {
	t.Helper()
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

	srv := New(db, git)
	return srv, db, git
}

func callRPC(handler http.Handler, reqBody map[string]any) *httptest.ResponseRecorder {
	return callRPCWithHeader(handler, reqBody, nil)
}

func callRPCWithHeader(handler http.Handler, reqBody map[string]any, headers map[string]string) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(raw))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestMCP_Initialize(t *testing.T) {
	srv, _, _ := setupTestMCP(t)
	h := srv.Handler(false)

	rec := callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]any{},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("initialize status: %d", rec.Code)
	}

	var resp rpcResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	res, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("expected map result: %+v", resp)
	}
	if res["protocolVersion"] != "2024-11-05" {
		t.Fatalf("expected protocolVersion 2024-11-05, got: %v", res["protocolVersion"])
	}
}

func TestMCP_ToolsList_ReadWrite_And_ReadOnly(t *testing.T) {
	srv, _, _ := setupTestMCP(t)

	// 1. Read-write handler (/mcp)
	rwHandler := srv.Handler(false)
	rec := callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("tools/list status: %d", rec.Code)
	}

	var resp rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res := resp.Result.(map[string]any)
	tools := res["tools"].([]any)
	if len(tools) < 33 {
		t.Fatalf("expected at least 33 tools, got %d", len(tools))
	}

	// 2. Read-only handler (/mcp/readonly)
	roHandler := srv.Handler(true)
	recRO := callRPC(roHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/list",
	})
	if recRO.Code != http.StatusOK {
		t.Fatalf("tools/list readonly status: %d", recRO.Code)
	}

	var respRO rpcResponse
	_ = json.Unmarshal(recRO.Body.Bytes(), &respRO)
	resRO := respRO.Result.(map[string]any)
	toolsRO := resRO["tools"].([]any)

	// Ensure no write tools are present in readonly list
	for _, rawT := range toolsRO {
		tDef := rawT.(map[string]any)
		name := tDef["name"].(string)
		if name == "create_repository" || name == "merge_pull_request" || name == "create_pull_request" {
			t.Fatalf("found write tool %q in /mcp/readonly tools/list", name)
		}
	}
}

func TestMCP_ToolCall_RepositoryAndReadOnlyGuard(t *testing.T) {
	srv, db, _ := setupTestMCP(t)
	ctx := context.Background()

	// Seed namespace
	ns := store.Namespace{
		ID:   idgen.NewNS(),
		Slug: "acme",
		Type: "team",
		Name: "Acme Corp",
	}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("create namespace: %v", err)
	}

	// 1. Call create_repository on read-only endpoint -> should return isError: true
	roHandler := srv.Handler(true)
	rec := callRPC(roHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      10,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_repository",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"name":      "rocket",
			},
		},
	})
	var resp rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res := resp.Result.(map[string]any)
	if res["isError"] != true {
		t.Fatalf("expected write tool on readonly to fail with isError=true, got: %+v", res)
	}

	// 2. Call create_repository on read-write endpoint -> should succeed
	rwHandler := srv.Handler(false)
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      11,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_repository",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"name":      "rocket",
			},
		},
	})
	var respCreate rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &respCreate)
	resCreate := respCreate.Result.(map[string]any)
	if resCreate["isError"] == true {
		t.Fatalf("create_repository failed: %+v", resCreate)
	}

	// 3. Call get_repository -> should return repo metadata
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      12,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "get_repository",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"repoName":  "rocket",
			},
		},
	})
	var respGet rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &respGet)
	resGet := respGet.Result.(map[string]any)
	if resGet["isError"] == true {
		t.Fatalf("get_repository failed: %+v", resGet)
	}

	// 4. Call list_repositories
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      13,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "list_repositories",
			"arguments": map[string]any{
				"ownerSlug": "acme",
			},
		},
	})
	var respList rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &respList)
	resList := respList.Result.(map[string]any)
	if resList["isError"] == true {
		t.Fatalf("list_repositories failed: %+v", resList)
	}
}

func TestMCP_LeaseTools(t *testing.T) {
	var auditBuf bytes.Buffer
	prevLogger := audit.DefaultLogger
	audit.SetDefaultLogger(audit.NewLogger(&auditBuf))
	t.Cleanup(func() { audit.SetDefaultLogger(prevLogger) })

	srv, _, _ := setupTestMCP(t)
	srv.SetLeaseSecret([]byte("test-super-secret-hmac-key-12345678"))
	rwHandler := srv.Handler(false)

	// 1. Request lease
	rec := callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      20,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "lease_request",
			"arguments": map[string]any{
				"agentId":    "agent-claude",
				"repoId":     "repo-123",
				"ttlSeconds": 300,
			},
		},
	})
	var resp rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res := resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("lease_request failed: %+v", res)
	}

	content := res["content"].([]any)[0].(map[string]any)["text"].(string)
	var leaseData map[string]any
	if err := json.Unmarshal([]byte(content), &leaseData); err != nil {
		t.Fatalf("unmarshal leaseData: %v", err)
	}
	token := leaseData["token"].(string)
	leaseID := leaseData["id"].(string)
	if token == "" || leaseID == "" {
		t.Fatalf("empty token or leaseID in lease_request response")
	}

	// 2. Inspect lease
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      21,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "lease_inspect",
			"arguments": map[string]any{
				"token": token,
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("lease_inspect failed: %+v", res)
	}
	content = res["content"].([]any)[0].(map[string]any)["text"].(string)
	var inspectData map[string]any
	_ = json.Unmarshal([]byte(content), &inspectData)
	if inspectData["agentId"] != "agent-claude" || inspectData["revoked"] != false {
		t.Fatalf("unexpected inspect data: %+v", inspectData)
	}

	// 3. Revoke lease
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      22,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "lease_revoke",
			"arguments": map[string]any{
				"leaseId": leaseID,
				"reason":  "task completed",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("lease_revoke failed: %+v", res)
	}

	// 4. Re-inspect -> revoked should be true
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      23,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "lease_inspect",
			"arguments": map[string]any{
				"token": token,
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	content = res["content"].([]any)[0].(map[string]any)["text"].(string)
	_ = json.Unmarshal([]byte(content), &inspectData)
	if inspectData["revoked"] != true {
		t.Fatalf("expected revoked=true, got %+v", inspectData)
	}

	logged := auditBuf.String()
	for _, want := range []string{"agent.lease_issued", "agent.lease_revoked", leaseID, "agent-claude"} {
		if !strings.Contains(logged, want) {
			t.Fatalf("expected lease audit stream to contain %q, got: %s", want, logged)
		}
	}
}

func TestMCP_CanonicalAgentLoopTools(t *testing.T) {
	srv, db, git := setupTestMCP(t)
	ctx := context.Background()
	rwHandler := srv.Handler(false)

	// 1. Seed namespace and repo
	nsID := idgen.NewNS()
	if err := db.CreateNamespace(ctx, store.Namespace{
		ID:   nsID,
		Slug: "agentcorp",
		Type: "team",
		Name: "Agent Corp",
	}); err != nil {
		t.Fatal(err)
	}

	repoID := idgen.NewRepo()
	if err := db.CreateRepo(ctx, store.Repo{
		ID:            repoID,
		NamespaceID:   nsID,
		Name:          "webforge",
		FullName:      "agentcorp/webforge",
		DefaultBranch: "main",
		Visibility:    "public",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := git.Init("agentcorp", "webforge", "main"); err != nil {
		t.Fatal(err)
	}

	sig := object.Signature{Name: "Dev", Email: "dev@refweir.dev", When: time.Now().UTC()}
	c1, err := git.WriteFiles("agentcorp", "webforge", "main", []gitstore.FileChange{
		{Path: "main.go", Content: "package main\n\nfunc main() {}\n"},
	}, sig, "Initial commit")
	if err != nil {
		t.Fatal(err)
	}

	// 2. Branch refs/heads/agent/feature
	if _, err := git.CreateBranchRef("agentcorp", "webforge", "refs/heads/agent/feature", c1.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := git.WriteFiles("agentcorp", "webforge", "agent/feature", []gitstore.FileChange{
		{Path: "agent.go", Content: "package main\n\nfunc RunAgent() string {\n\treturn \"autonomous\"\n}\n"},
	}, sig, "Add agent.go"); err != nil {
		t.Fatal(err)
	}

	// 3. Test repo_tree
	rec := callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      30,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "repo_tree",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"ref":       "agent/feature",
			},
		},
	})
	var resp rpcResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res := resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("repo_tree failed: %+v", res)
	}

	// 4. Test file_read
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      31,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "file_read",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"ref":       "agent/feature",
				"path":      "agent.go",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("file_read failed: %+v", res)
	}

	// 5. Test diff_compare
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      32,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "diff_compare",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"base":      "main",
				"head":      "agent/feature",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("diff_compare failed: %+v", res)
	}

	// 6. Test symbol_search
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      33,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "symbol_search",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"query":     "RunAgent",
				"ref":       "agent/feature",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("symbol_search failed: %+v", res)
	}

	// 7. Test pull_request_create
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      34,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "pull_request_create",
			"arguments": map[string]any{
				"ownerSlug": "agentcorp",
				"repoName":  "webforge",
				"title":     "Add autonomous agent",
				"headRef":   "agent/feature",
				"baseRef":   "main",
				"body":      "Implements RunAgent function",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("pull_request_create failed: %+v", res)
	}

	// 8. Test thread_comment
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      35,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "thread_comment",
			"arguments": map[string]any{
				"ownerSlug":  "agentcorp",
				"repoName":   "webforge",
				"pullNumber": 1,
				"body":       "Looks great, verified AST symbols and signatures.",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("thread_comment failed: %+v", res)
	}

	// 9. Add ruleset requiring check 'ci/test'
	headSHA, _ := git.ResolveRef("agentcorp", "webforge", "agent/feature")
	_, _ = db.CreateRuleset(ctx, store.Ruleset{
		ID:               idgen.New("rs"),
		RepoID:           repoID,
		Name:             "require-ci",
		Kind:             "merge_branch",
		Enforcement:      "active",
		IncludedRefNames: []string{"~DEFAULT_BRANCH"},
		Rules: []store.RulesetRule{
			{
				RuleType: "require_status_checks",
				Parameters: map[string]any{
					"requiredChecks": []any{
						map[string]any{"name": "ci/test", "context": "ci/test"},
					},
				},
			},
		},
	})

	// 10. Test pull_request_merge BEFORE status check -> MUST FAIL due to blocker!
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      36,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "pull_request_merge",
			"arguments": map[string]any{
				"ownerSlug":  "agentcorp",
				"repoName":   "webforge",
				"pullNumber": 1,
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] != true {
		t.Fatalf("expected pull_request_merge to be blocked by missing check run, but it succeeded: %+v", res)
	}

	// 11. Test status_check_report -> report passing ci/test check run
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      37,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "status_check_report",
			"arguments": map[string]any{
				"ownerSlug":  "agentcorp",
				"repoName":   "webforge",
				"headSha":    headSHA.String(),
				"name":       "ci/test",
				"status":     "completed",
				"conclusion": "success",
				"title":      "Unit tests",
				"summary":    "All 42 tests passed in 12ms",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("status_check_report failed: %+v", res)
	}

	// 12. Test pull_request_merge AFTER status check passes -> MUST SUCCEED!
	rec = callRPC(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      38,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "pull_request_merge",
			"arguments": map[string]any{
				"ownerSlug":  "agentcorp",
				"repoName":   "webforge",
				"pullNumber": 1,
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	res = resp.Result.(map[string]any)
	if res["isError"] == true {
		t.Fatalf("pull_request_merge after passing check failed: %+v", res)
	}
}

func TestMCP_IssueTools(t *testing.T) {
	srv, db, _ := setupTestMCP(t)
	h := srv.Handler(false)
	ctx := context.Background()

	// 1. Create repo
	_ = db.CreateNamespace(ctx, store.Namespace{ID: "ns-1", Slug: "acme", Name: "Acme"})
	repo := store.Repo{
		ID:            "repo-issues",
		NamespaceID:   "ns-1",
		Name:          "tracker",
		FullName:      "acme/tracker",
		DefaultBranch: "main",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("create repo: %v", err)
	}

	// 2. create_issue via MCP
	rec := callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_issue",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"repoName":  "tracker",
				"title":     "High memory consumption in buffer pool",
				"body":      "Worker allocates without freeing during batch synchronization.",
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("create_issue status: %d", rec.Code)
	}
	var resp struct {
		Result map[string]any `json:"result"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] == true {
		t.Fatalf("create_issue error: %+v", resp.Result)
	}

	// 3. create_issue_comment via MCP
	rec = callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_issue_comment",
			"arguments": map[string]any{
				"ownerSlug":   "acme",
				"repoName":    "tracker",
				"issueNumber": 1,
				"body":        "Reproduced on v1.4.2 under 10k connections. Working on patch.",
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("create_issue_comment status: %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] == true {
		t.Fatalf("create_issue_comment error: %+v", resp.Result)
	}

	// 4. get_issue via MCP
	rec = callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "get_issue",
			"arguments": map[string]any{
				"ownerSlug":   "acme",
				"repoName":    "tracker",
				"issueNumber": 1,
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("get_issue status: %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] == true {
		t.Fatalf("get_issue error: %+v", resp.Result)
	}

	// 5. list_issues via MCP
	rec = callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "list_issues",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"repoName":  "tracker",
				"state":     "open",
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("list_issues status: %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] == true {
		t.Fatalf("list_issues error: %+v", resp.Result)
	}

	// 6. close_issue via MCP
	rec = callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      5,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "close_issue",
			"arguments": map[string]any{
				"ownerSlug":   "acme",
				"repoName":    "tracker",
				"issueNumber": 1,
				"comment":     "Resolved by PR #42 in commit abc1234",
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("close_issue status: %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] == true {
		t.Fatalf("close_issue error: %+v", resp.Result)
	}

	// 7. reopen_issue via MCP
	rec = callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      6,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "reopen_issue",
			"arguments": map[string]any{
				"ownerSlug":   "acme",
				"repoName":    "tracker",
				"issueNumber": 1,
				"comment":     "Reopening due to edge case regression on ARM64",
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("reopen_issue status: %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] == true {
		t.Fatalf("reopen_issue error: %+v", resp.Result)
	}
}

func TestMCP_WebhookTools(t *testing.T) {
	srv, db, _ := setupTestMCP(t)
	h := srv.Handler(false)
	ctx := context.Background()

	nsID := idgen.NewNS()
	_ = db.CreateNamespace(ctx, store.Namespace{
		ID:   nsID,
		Slug: "acme",
		Type: "organization",
		Name: "Acme Corp",
	})
	repoID := idgen.NewRepo()
	repo := store.Repo{
		ID:            repoID,
		NamespaceID:   nsID,
		Name:          "api-gateway",
		FullName:      "acme/api-gateway",
		DefaultBranch: "main",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("create repo: %v", err)
	}

	// 1. create_webhook via MCP
	rec := callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_webhook",
			"arguments": map[string]any{
				"ownerSlug":   "acme",
				"repoName":    "api-gateway",
				"url":         "https://ci.example.com/hooks/refweir",
				"secret":      "supersecretkey123",
				"events":      []any{"push", "pull_request"},
				"insecureSsl": false,
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("create_webhook status: %d", rec.Code)
	}
	var resp struct {
		Result map[string]any `json:"result"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] == true {
		t.Fatalf("create_webhook error: %+v", resp.Result)
	}
	content := resp.Result["content"].([]any)
	c0 := content[0].(map[string]any)
	var createdHook store.RepoWebhook
	if err := json.Unmarshal([]byte(c0["text"].(string)), &createdHook); err != nil {
		t.Fatalf("unmarshal created hook: %v", err)
	}
	hookID := createdHook.ID
	if hookID == "" {
		t.Fatalf("expected non-empty webhook id: %+v", createdHook)
	}

	// 2. list_webhooks via MCP
	rec = callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "list_webhooks",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"repoName":  "api-gateway",
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("list_webhooks status: %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] == true {
		t.Fatalf("list_webhooks error: %+v", resp.Result)
	}
	content = resp.Result["content"].([]any)
	c0 = content[0].(map[string]any)
	var hooksList []*store.RepoWebhook
	if err := json.Unmarshal([]byte(c0["text"].(string)), &hooksList); err != nil {
		t.Fatalf("unmarshal hooks list: %v", err)
	}
	if len(hooksList) != 1 {
		t.Fatalf("expected 1 webhook in list, got %d", len(hooksList))
	}
	if hooksList[0].URL != "https://ci.example.com/hooks/refweir" {
		t.Errorf("expected url https://ci.example.com/hooks/refweir, got %v", hooksList[0].URL)
	}

	// 3. delete_webhook via MCP
	rec = callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "delete_webhook",
			"arguments": map[string]any{
				"webhookId": hookID,
			},
		},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("delete_webhook status: %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] == true {
		t.Fatalf("delete_webhook error: %+v", resp.Result)
	}

	// Verify webhook deleted
	hooks, err := db.ListRepoWebhooks(ctx, repo.ID)
	if err != nil {
		t.Fatalf("ListRepoWebhooks: %v", err)
	}
	if len(hooks) != 0 {
		t.Fatalf("expected 0 webhooks after delete, got %d", len(hooks))
	}

	// 4. create_webhook with SSRF target should fail
	rec = callRPC(h, map[string]any{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_webhook",
			"arguments": map[string]any{
				"ownerSlug": "acme",
				"repoName":  "api-gateway",
				"url":       "http://169.254.169.254/latest/meta-data/",
			},
		},
	})
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Result["isError"] != true {
		t.Fatalf("expected create_webhook with cloud metadata IP to fail with isError=true")
	}
}

func TestMCP_ScopeEnforcement_ReadOnlyTokenBlockedOnWrite(t *testing.T) {
	srv, db, _ := setupTestMCP(t)
	ctx := context.Background()

	// 1. Create User
	userID := idgen.NewUser()
	if err := db.CreateUserWithPassword(ctx, store.User{
		ID:    userID,
		Login: "sandbox-agent",
		Name:  "Sandbox Agent",
		Email: "agent@refweir.dev",
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// 2. Create Read-Only API Key ("read:repo")
	rawROKey, roPrefix, roHash, _ := auth.GenerateAPIKey()
	if err := db.CreateAPIKey(ctx, store.APIKey{
		ID:        idgen.New("key"),
		UserID:    userID,
		KeyHash:   roHash,
		KeyPrefix: roPrefix,
		Name:      "agent-ro-key",
		Scopes:    "read:repo",
	}); err != nil {
		t.Fatalf("CreateAPIKey RO: %v", err)
	}

	// 3. Create Read-Write API Key ("repo")
	rawRWKey, rwPrefix, rwHash, _ := auth.GenerateAPIKey()
	if err := db.CreateAPIKey(ctx, store.APIKey{
		ID:        idgen.New("key"),
		UserID:    userID,
		KeyHash:   rwHash,
		KeyPrefix: rwPrefix,
		Name:      "agent-rw-key",
		Scopes:    "repo",
	}); err != nil {
		t.Fatalf("CreateAPIKey RW: %v", err)
	}

	// Create Namespace
	ns := store.Namespace{ID: idgen.NewNS(), Slug: "agent-corp", Type: "team", Name: "Agent Corp"}
	_ = db.CreateNamespace(ctx, ns)

	// Handler is on /mcp (the read-write endpoint)
	rwHandler := srv.Handler(false)

	// 4. Calling read tool (list_repositories) with Read-Only key on /mcp -> SUCCEEDS
	recList := callRPCWithHeader(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "list_repositories",
			"arguments": map[string]any{},
		},
	}, map[string]string{
		"Authorization": "Bearer " + rawROKey,
	})
	if recList.Code != http.StatusOK {
		t.Fatalf("list_repositories status: %d", recList.Code)
	}
	var respList rpcResponse
	_ = json.Unmarshal(recList.Body.Bytes(), &respList)
	resList, ok := respList.Result.(map[string]any)
	if !ok || resList["isError"] == true {
		t.Fatalf("expected list_repositories to succeed with read-only key, got error: %+v", respList.Result)
	}

	// 5. Calling write tool (create_repository) with Read-Only key on /mcp -> MUST FAIL with isError: true
	recWrite := callRPCWithHeader(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_repository",
			"arguments": map[string]any{
				"ownerSlug": "agent-corp",
				"name":      "blocked-repo",
			},
		},
	}, map[string]string{
		"Authorization": "Bearer " + rawROKey,
	})
	var respWrite rpcResponse
	_ = json.Unmarshal(recWrite.Body.Bytes(), &respWrite)
	resWrite, ok := respWrite.Result.(map[string]any)
	if !ok || resWrite["isError"] != true {
		t.Fatalf("expected write tool with read-only key to fail with isError=true on /mcp")
	}

	// 6. Calling write tool (create_repository) with Read-Write key on /mcp -> MUST SUCCEED
	recWriteOK := callRPCWithHeader(rwHandler, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "create_repository",
			"arguments": map[string]any{
				"ownerSlug": "agent-corp",
				"name":      "allowed-repo",
			},
		},
	}, map[string]string{
		"Authorization": "Bearer " + rawRWKey,
	})
	var respWriteOK rpcResponse
	_ = json.Unmarshal(recWriteOK.Body.Bytes(), &respWriteOK)
	resWriteOK, ok := respWriteOK.Result.(map[string]any)
	if !ok || resWriteOK["isError"] == true {
		t.Fatalf("expected write tool with read-write key to succeed, got error: %+v", respWriteOK.Result)
	}
}

