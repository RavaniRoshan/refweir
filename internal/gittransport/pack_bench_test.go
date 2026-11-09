package gittransport

import (
	"fmt"
	"testing"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/revlist"
)

func setupMultiRefRepo(b *testing.B, commitCount, branchCount int) (*git.Repository, []plumbing.Hash) {
	b.Helper()
	dir := b.TempDir()
	r, err := git.PlainInit(dir, false)
	if err != nil {
		b.Fatalf("PlainInit: %v", err)
	}
	wt, err := r.Worktree()
	if err != nil {
		b.Fatalf("Worktree: %v", err)
	}

	hashes := make([]plumbing.Hash, 0, commitCount)
	for i := 0; i < commitCount; i++ {
		h, err := wt.Commit(fmt.Sprintf("commit %d", i), &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Benchmark",
				Email: "bench@refweir.dev",
				When:  time.Now(),
			},
			AllowEmptyCommits: true,
		})
		if err != nil {
			b.Fatalf("Commit: %v", err)
		}
		hashes = append(hashes, h)
	}

	// Create branches distributed across the commits
	refHashes := make([]plumbing.Hash, 0, branchCount)
	for i := 0; i < branchCount; i++ {
		targetHash := hashes[i%commitCount]
		refName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/branch-%d", i))
		ref := plumbing.NewHashReference(refName, targetHash)
		if err := r.Storer.SetReference(ref); err != nil {
			b.Fatalf("SetReference: %v", err)
		}
		refHashes = append(refHashes, targetHash)
	}

	return r, refHashes
}

func BenchmarkRevlist_1WantedRef(b *testing.B) {
	r, refHashes := setupMultiRefRepo(b, 300, 256)
	wants := refHashes[:1]

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		objs, err := revlist.Objects(r.Storer, wants, nil)
		if err != nil {
			b.Fatalf("revlist.Objects: %v", err)
		}
		if len(objs) == 0 {
			b.Fatal("expected reachable objects")
		}
	}
}

func BenchmarkRevlist_256WantedRefs(b *testing.B) {
	r, refHashes := setupMultiRefRepo(b, 300, 256)
	wants := refHashes

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		objs, err := revlist.Objects(r.Storer, wants, nil)
		if err != nil {
			b.Fatalf("revlist.Objects: %v", err)
		}
		if len(objs) == 0 {
			b.Fatal("expected reachable objects")
		}
	}
}
