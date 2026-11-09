package gitstore

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

func setupBenchRepo(b *testing.B, commitCount int) (*git.Repository, plumbing.Hash) {
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

	var lastHash plumbing.Hash
	for i := 0; i < commitCount; i++ {
		h, err := wt.Commit("benchmark commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Benchmarker",
				Email: "bench@refweir.dev",
				When:  time.Now(),
			},
			AllowEmptyCommits: true,
		})
		if err != nil {
			b.Fatalf("Commit %d: %v", i, err)
		}
		lastHash = h
	}
	return r, lastHash
}

func BenchmarkGitStore_CommitLogWalk(b *testing.B) {
	r, head := setupBenchRepo(b, 200)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cIter, err := r.Log(&git.LogOptions{From: head})
		if err != nil {
			b.Fatalf("Log: %v", err)
		}
		count := 0
		_ = cIter.ForEach(func(c *object.Commit) error {
			count++
			return nil
		})
		cIter.Close()
		if count != 200 {
			b.Fatalf("walked %d commits, want 200", count)
		}
	}
}

func BenchmarkGitStore_InitBareRepo(b *testing.B) {
	dir := b.TempDir()
	s, err := New(filepath.Join(dir, "git"))
	if err != nil {
		b.Fatalf("New store: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		repo, err := s.Init("benchorg", fmt.Sprintf("repo_%d", i), "main")
		if err != nil || repo == nil {
			b.Fatalf("Init: %v", err)
		}
	}
}
