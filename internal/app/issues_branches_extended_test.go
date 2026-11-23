package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"refweir.dev/internal/config"
	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/store"
)

func TestIssuesAndBranchesExtendedEndpoints(t *testing.T) {
	dataDir := t.TempDir()
	cfg := config.Config{
		ListenAddr: "127.0.0.1:0",
		DataDir:    dataDir,
		PublicURL:  "http://127.0.0.1:8080",
		GitURL:     "127.0.0.1:8080",
	}

	app, err := Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer app.Close()

	ctx := t.Context()
	handler := app.Handler()

	// Setup namespace & repo
	ns := store.Namespace{ID: "ns_ext_test", Slug: "extorg", Type: "team", Name: "ExtOrg"}
	if err := app.Store.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	repo := store.Repo{
		ID:            "repo_ext_test",
		NamespaceID:   ns.ID,
		Name:          "widget-service",
		DefaultBranch: "main",
	}
	if err := app.Store.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}
	gitRepo, err := app.Git.Init("extorg", "widget-service", "main")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	// 1. Create an issue
	iss := &store.Issue{
		RepoID:       repo.ID,
		AuthorUserID: "",
		Title:        "Security advisory",
		Body:         "Found vulnerability in parsing",
		State:        "open",
	}
	if err := app.Store.CreateIssue(ctx, iss); err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}

	// Test GET /api/v1/repos/extorg/widget-service/issues/{number}
	req := httptest.NewRequest("GET", "/api/v1/repos/extorg/widget-service/issues/1", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET issue returned status %d, want 200. Body: %s", rec.Code, rec.Body.String())
	}
	var issueResp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &issueResp); err != nil {
		t.Fatalf("Unmarshal issue response: %v", err)
	}
	if issueResp["id"] != iss.ID || issueResp["title"] != "Security advisory" {
		t.Fatalf("unexpected issue payload: %+v", issueResp)
	}

	// Non-existent issue -> 404
	req404 := httptest.NewRequest("GET", "/api/v1/repos/extorg/widget-service/issues/999", nil)
	rec404 := httptest.NewRecorder()
	handler.ServeHTTP(rec404, req404)
	if rec404.Code != http.StatusNotFound {
		t.Fatalf("GET non-existent issue returned status %d, want 404", rec404.Code)
	}

	// 2. Test branch endpoints
	// Create an initial commit on main
	blob := gitRepo.Storer.NewEncodedObject()
	blob.SetType(plumbing.BlobObject)
	w, _ := blob.Writer()
	_, _ = w.Write([]byte("initial commit content"))
	_ = w.Close()
	blobHash, _ := gitRepo.Storer.SetEncodedObject(blob)

	tree := &object.Tree{Entries: []object.TreeEntry{{
		Name: "README.md", Mode: 0o100644, Hash: blobHash,
	}}}
	treeObj := gitRepo.Storer.NewEncodedObject()
	_ = tree.Encode(treeObj)
	treeHash, _ := gitRepo.Storer.SetEncodedObject(treeObj)

	commitHash, err := gitstore.CommitObject(gitRepo, "initial commit", object.Signature{Name: "Tester", Email: "test@example.com"}, object.Signature{Name: "Tester", Email: "test@example.com"}, treeHash, nil)
	if err != nil {
		t.Fatalf("CommitObject: %v", err)
	}
	_ = gitRepo.Storer.SetReference(plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/main"), commitHash))

	// Create feature-1 branch ref
	_, err = app.Git.CreateBranchRef("extorg", "widget-service", "refs/heads/feature-1", commitHash.String())
	if err != nil {
		t.Fatalf("CreateBranchRef: %v", err)
	}

	// Delete default branch must be rejected with 400
	delDefaultReq := httptest.NewRequest("DELETE", "/api/v1/repos/extorg/widget-service/branches/main", nil)
	delDefaultRec := httptest.NewRecorder()
	handler.ServeHTTP(delDefaultRec, delDefaultReq)
	if delDefaultRec.Code != http.StatusBadRequest {
		t.Fatalf("DELETE default branch status = %d, want 400. Body: %s", delDefaultRec.Code, delDefaultRec.Body.String())
	}

	// Delete feature-1 branch should succeed with 200
	delFeatReq := httptest.NewRequest("DELETE", "/api/v1/repos/extorg/widget-service/branches/feature-1", nil)
	delFeatRec := httptest.NewRecorder()
	handler.ServeHTTP(delFeatRec, delFeatReq)
	if delFeatRec.Code != http.StatusOK {
		t.Fatalf("DELETE feature-1 branch status = %d, want 200. Body: %s", delFeatRec.Code, delFeatRec.Body.String())
	}

	// Deleting non-existent branch returns 404
	delMissingReq := httptest.NewRequest("DELETE", "/api/v1/repos/extorg/widget-service/branches/does-not-exist", nil)
	delMissingRec := httptest.NewRecorder()
	handler.ServeHTTP(delMissingRec, delMissingReq)
	if delMissingRec.Code != http.StatusNotFound {
		t.Fatalf("DELETE non-existent branch status = %d, want 404. Body: %s", delMissingRec.Code, delMissingRec.Body.String())
	}
}
