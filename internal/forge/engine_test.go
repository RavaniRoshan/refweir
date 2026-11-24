package forge

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"
	"refweir.dev/internal/events"
	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/store"
)

func setupTestEngine(t *testing.T) (*Engine, *store.DB, *gitstore.Store, func()) {
	t.Helper()
	t.Setenv("REFWEIR_ALLOW_LOOPBACK_WEBHOOKS", "true")
	tmpDir := t.TempDir()

	db, err := store.Open(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}

	gitStore, err := gitstore.New(filepath.Join(tmpDir, "repos"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	bus := events.NewBus(db, nil)
	eng := New(db, gitStore, bus, Config{PublicURL: "https://refweir.test"})

	cleanup := func() {
		_ = db.Close()
	}
	return eng, db, gitStore, cleanup
}

func TestEngine_CreateRepository(t *testing.T) {
	eng, _, _, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	repo, err := eng.CreateRepository(ctx, CreateRepoRequest{
		Owner:         "alice",
		Name:          "web-framework",
		Description:   "A clean modern framework",
		DefaultBranch: "main",
		Actor:         "alice",
		ActorType:     "user",
	})
	if err != nil {
		t.Fatalf("CreateRepository failed: %v", err)
	}

	if repo.FullName != "alice/web-framework" {
		t.Errorf("expected FullName 'alice/web-framework', got %q", repo.FullName)
	}
	if repo.CloneURL != "https://refweir.test/alice/web-framework.git" {
		t.Errorf("expected CloneURL 'https://refweir.test/alice/web-framework.git', got %q", repo.CloneURL)
	}

	// Conflict test
	_, err = eng.CreateRepository(ctx, CreateRepoRequest{
		Owner: "alice",
		Name:  "web-framework",
	})
	if err == nil {
		t.Fatalf("expected conflict error, got nil")
	}
}

func TestEngine_ForkAndSyncLifecycle(t *testing.T) {
	eng, _, gitStore, cleanup := setupTestEngine(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Create upstream repo
	upstream, err := eng.CreateRepository(ctx, CreateRepoRequest{
		Owner:         "upstream-org",
		Name:          "core-lib",
		DefaultBranch: "main",
		Actor:         "maintainer",
	})
	if err != nil {
		t.Fatalf("Create upstream repo: %v", err)
	}

	// Commit initial file to upstream
	sig := object.Signature{Name: "Upstream", Email: "upstream@refweir.test", When: time.Now()}
	_, err = gitStore.WriteFiles("upstream-org", "core-lib", "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Upstream\nInitial commit."},
	}, sig, "Initial commit")
	if err != nil {
		t.Fatalf("WriteFiles: %v", err)
	}

	// 2. Self-fork rejection
	_, err = eng.ForkRepository(ctx, ForkRepoRequest{
		SourceOwner:     "upstream-org",
		SourceName:      "core-lib",
		TargetNamespace: "upstream-org",
		TargetName:      "core-lib",
	})
	if err == nil {
		t.Fatalf("expected ErrCannotSelfFork, got nil")
	}

	// 3. Valid Fork
	forkRepo, err := eng.ForkRepository(ctx, ForkRepoRequest{
		SourceOwner:     "upstream-org",
		SourceName:      "core-lib",
		TargetNamespace: "bob",
		TargetName:      "core-lib-fork",
		Actor:           "bob",
		ActorType:       "user",
	})
	if err != nil {
		t.Fatalf("ForkRepository: %v", err)
	}

	if !forkRepo.IsFork {
		t.Errorf("expected forkRepo.IsFork = true")
	}
	if forkRepo.ParentRepoID != upstream.ID {
		t.Errorf("expected ParentRepoID %q, got %q", upstream.ID, forkRepo.ParentRepoID)
	}

	// 4. List Forks
	forks, err := eng.ListForks(ctx, "upstream-org", "core-lib", 10)
	if err != nil {
		t.Fatalf("ListForks: %v", err)
	}
	if len(forks) != 1 || forks[0].FullName != "bob/core-lib-fork" {
		t.Fatalf("expected 1 fork 'bob/core-lib-fork', got %+v", forks)
	}

	// 5. Advance upstream
	_, err = gitStore.WriteFiles("upstream-org", "core-lib", "main", []gitstore.FileChange{
		{Path: "FEATURE.txt", Content: "New feature v2\n"},
	}, sig, "Add feature v2")
	if err != nil {
		t.Fatalf("Advance upstream commit: %v", err)
	}

	// 6. Sync Fork
	syncRes, err := eng.SyncFork(ctx, SyncForkRequest{
		Owner:  "bob",
		Name:   "core-lib-fork",
		Branch: "main",
		Actor:  "bob",
	})
	if err != nil {
		t.Fatalf("SyncFork: %v", err)
	}

	if syncRes.MergeType != "fast-forward" {
		t.Errorf("expected 'fast-forward' merge type, got %q", syncRes.MergeType)
	}

	// Verify fork has FEATURE.txt
	data, _, _, isDir, err := gitStore.ReadFileAtPath("bob", "core-lib-fork", "main", "FEATURE.txt")
	if err != nil || isDir {
		t.Fatalf("ReadFileAtPath on synced fork failed: %v", err)
	}
	if string(data) != "New feature v2\n" {
		t.Errorf("unexpected content: %q", string(data))
	}
}
