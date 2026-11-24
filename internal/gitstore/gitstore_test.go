package gitstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(filepath.Join(t.TempDir(), "git"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestInitCreatesBareRepo(t *testing.T) {
	s := newStore(t)

	repo, err := s.Init("acme", "widgets", "main")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	// A bare repo has no worktree.
	wt, err := repo.Worktree()
	if err == nil && wt != nil {
		t.Error("expected bare repo (no worktree)")
	}

	// The directory exists on disk with a HEAD.
	path := s.repoPath("acme", "widgets")
	if _, err := os.Stat(filepath.Join(path, "HEAD")); err != nil {
		t.Errorf("HEAD not created: %v", err)
	}
	if !s.Exists("acme", "widgets") {
		t.Error("Exists = false after Init")
	}

	// HEAD points at the default branch.
	head, err := repo.Reference(plumbing.HEAD, false)
	if err != nil {
		t.Fatalf("Reference(HEAD): %v", err)
	}
	if head.Target().Short() != "main" {
		t.Errorf("HEAD target = %q, want main", head.Target().Short())
	}
}

func TestInitIdempotentConflict(t *testing.T) {
	s := newStore(t)
	if _, err := s.Init("acme", "widgets", "main"); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	if _, err := s.Init("acme", "widgets", "main"); err == nil {
		t.Error("second Init = nil, want already-exists error")
	}
}

func TestOpenRoundtrip(t *testing.T) {
	s := newStore(t)
	if _, err := s.Init("acme", "widgets", "main"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	repo, err := s.Open("acme", "widgets")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if repo == nil {
		t.Fatal("Open returned nil repo")
	}
	// Opening a nonexistent repo errors.
	if _, err := s.Open("acme", "ghost"); err == nil {
		t.Error("Open(ghost) = nil, want error")
	}
}

// TestCloneFromLocalPath roundtrips commits: build a source repo with a commit,
// clone it via the store, and confirm the commit hash matches.
func TestCloneFromLocalPath(t *testing.T) {
	// Build a source repo with one commit.
	srcDir := filepath.Join(t.TempDir(), "src")
	src, err := git.PlainInit(srcDir, false)
	if err != nil {
		t.Fatalf("PlainInit src: %v", err)
	}
	wt, err := src.Worktree()
	if err != nil {
		t.Fatalf("Worktree: %v", err)
	}
	f := filepath.Join(srcDir, "README.md")
	if err := os.WriteFile(f, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("Add: %v", err)
	}
	sig := &object.Signature{Name: "Test", Email: "t@example.com"}
	srcHash, err := wt.Commit("initial", &git.CommitOptions{Author: sig, Committer: sig})
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// Clone via the store (local path, no auth).
	s := newStore(t)
	cloned, err := s.Clone("acme", "mirror", srcDir, nil)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	head, err := cloned.Head()
	if err != nil {
		t.Fatalf("cloned Head: %v", err)
	}
	if head.Hash() != srcHash {
		t.Errorf("cloned HEAD = %s, want %s", head.Hash(), srcHash)
	}

	// The object is retrievable.
	_, err = cloned.CommitObject(head.Hash())
	if err != nil {
		t.Errorf("CommitObject: %v", err)
	}
}

func TestCommitObject(t *testing.T) {
	s := newStore(t)
	repo, err := s.Init("acme", "widgets", "main")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	// Build a tree with one blob.
	blob := repo.Storer.NewEncodedObject()
	blob.SetType(plumbing.BlobObject)
	w, err := blob.Writer()
	if err != nil {
		t.Fatalf("blob writer: %v", err)
	}
	if _, err := w.Write([]byte("content")); err != nil {
		t.Fatalf("blob write: %v", err)
	}
	w.Close()
	blobHash, err := repo.Storer.SetEncodedObject(blob)
	if err != nil {
		t.Fatalf("SetEncodedObject blob: %v", err)
	}

	tree := &object.Tree{Entries: []object.TreeEntry{{
		Name: "file.txt", Mode: 0o100644, Hash: blobHash,
	}}}
	treeObj := repo.Storer.NewEncodedObject()
	if err := tree.Encode(treeObj); err != nil {
		t.Fatalf("tree encode: %v", err)
	}
	treeHash, err := repo.Storer.SetEncodedObject(treeObj)
	if err != nil {
		t.Fatalf("SetEncodedObject tree: %v", err)
	}

	sig := object.Signature{Name: "T", Email: "t@e.com"}
	h, err := CommitObject(repo, "root commit", sig, sig, treeHash, nil)
	if err != nil {
		t.Fatalf("CommitObject: %v", err)
	}
	if h == plumbing.ZeroHash {
		t.Fatal("CommitObject returned zero hash")
	}

	// The commit is retrievable and decodes to the tree we built.
	c, err := repo.CommitObject(h)
	if err != nil {
		t.Fatalf("CommitObject decode: %v", err)
	}
	if c.TreeHash != treeHash {
		t.Errorf("commit tree = %s, want %s", c.TreeHash, treeHash)
	}
	if c.Message != "root commit" {
		t.Errorf("commit message = %q", c.Message)
	}
}

func TestStore_Delete_Rename_Locks(t *testing.T) {
	s := newStore(t)

	// 1. Init repository
	if _, err := s.Init("acme", "project-a", "main"); err != nil {
		t.Fatalf("Init project-a: %v", err)
	}
	if !s.Exists("acme", "project-a") {
		t.Fatal("expected project-a to exist")
	}

	// 2. Rename project-a to project-b with concurrent lock safety
	if err := s.Rename("acme", "project-a", "project-b"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if s.Exists("acme", "project-a") {
		t.Fatal("expected project-a to be gone after rename")
	}
	if !s.Exists("acme", "project-b") {
		t.Fatal("expected project-b to exist after rename")
	}

	// 3. Delete project-b with lock safety
	if err := s.Delete("acme", "project-b"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if s.Exists("acme", "project-b") {
		t.Fatal("expected project-b to be removed after Delete")
	}
}

func TestEnsureHooks_ScriptAndPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	if err := EnsureHooks(tmpDir); err != nil {
		t.Fatalf("EnsureHooks: %v", err)
	}

	hookPath := filepath.Join(tmpDir, "hooks", "pre-receive")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("Stat hooks/pre-receive: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("expected hook to be executable, got mode: %v", info.Mode())
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("ReadFile hooks/pre-receive: %v", err)
	}
	str := string(content)
	if !strings.Contains(str, "#!/bin/sh") {
		t.Error("expected shebang in hook script")
	}
	if !strings.Contains(str, "REFWEIRD_BIN") {
		t.Error("expected REFWEIRD_BIN in hook script")
	}
	if !strings.Contains(str, "hook pre-receive") {
		t.Error("expected 'hook pre-receive' invocation in hook script")
	}
	if !strings.Contains(str, "command -v refweird") {
		t.Error("expected command -v refweird in hook script")
	}
	if !strings.Contains(str, "command -v refweir") {
		t.Error("expected command -v refweir in hook script")
	}
}

func TestStore_Fork(t *testing.T) {
	s := newStore(t)

	// 1. Init upstream repo and create a root commit
	repo, err := s.Init("upstream", "core", "main")
	if err != nil {
		t.Fatalf("Init upstream/core: %v", err)
	}
	blob := repo.Storer.NewEncodedObject()
	blob.SetType(plumbing.BlobObject)
	w, _ := blob.Writer()
	_, _ = w.Write([]byte("core code"))
	w.Close()
	blobHash, _ := repo.Storer.SetEncodedObject(blob)
	tree := &object.Tree{Entries: []object.TreeEntry{{Name: "main.go", Mode: 0o100644, Hash: blobHash}}}
	treeObj := repo.Storer.NewEncodedObject()
	_ = tree.Encode(treeObj)
	treeHash, _ := repo.Storer.SetEncodedObject(treeObj)
	sig := object.Signature{Name: "CoreDev", Email: "dev@upstream.org"}
	commitHash, err := CommitObject(repo, "initial release", sig, sig, treeHash, nil)
	if err != nil {
		t.Fatalf("CommitObject: %v", err)
	}

	// Update HEAD reference
	ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/main"), commitHash)
	if err := repo.Storer.SetReference(ref); err != nil {
		t.Fatalf("SetReference: %v", err)
	}

	// 2. Fork into contributor/core
	forkedRepo, err := s.Fork("upstream", "core", "contributor", "core")
	if err != nil {
		t.Fatalf("Fork: %v", err)
	}
	if !s.Exists("contributor", "core") {
		t.Fatal("expected contributor/core to exist on disk")
	}

	// 3. Object retrievability in fork
	c, err := forkedRepo.CommitObject(commitHash)
	if err != nil {
		t.Fatalf("forkedRepo CommitObject: %v", err)
	}
	if c.Message != "initial release" {
		t.Errorf("commit message = %q, want 'initial release'", c.Message)
	}

	// 4. Verify pre-receive hook in fork
	hookInfo, err := os.Stat(filepath.Join(s.repoPath("contributor", "core"), "hooks", "pre-receive"))
	if err != nil {
		t.Fatalf("fork hooks/pre-receive not created: %v", err)
	}
	if hookInfo.Mode()&0o111 == 0 {
		t.Error("fork hook not executable")
	}
}


