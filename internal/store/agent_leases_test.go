package store_test

import (
	"context"
	"path/filepath"
	"testing"

	"refweir.dev/internal/store"
)

func TestAgentLeases_Revocation(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	leaseID := "lease_test_123"

	// 1. Initial lease is not revoked
	revoked, err := db.IsAgentLeaseRevoked(ctx, leaseID)
	if err != nil {
		t.Fatalf("IsAgentLeaseRevoked: %v", err)
	}
	if revoked {
		t.Fatal("expected lease to not be revoked")
	}

	// 2. Revoke lease
	err = db.RevokeAgentLease(ctx, leaseID, "agent-claude", "repo-1", "work completed")
	if err != nil {
		t.Fatalf("RevokeAgentLease: %v", err)
	}

	// 3. Now it should be revoked
	revoked, err = db.IsAgentLeaseRevoked(ctx, leaseID)
	if err != nil {
		t.Fatalf("IsAgentLeaseRevoked: %v", err)
	}
	if !revoked {
		t.Fatal("expected lease to be marked as revoked")
	}
}

func TestAgentLeases_RepoRevocation(t *testing.T) {
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repoID := "repo_test_arch"

	// 1. Initial lease on repo is not revoked
	revoked, err := db.IsLeaseRevoked(ctx, "lease_active_1", repoID)
	if err != nil {
		t.Fatalf("IsLeaseRevoked: %v", err)
	}
	if revoked {
		t.Fatal("expected lease to not be revoked initially")
	}

	// 2. Revoke whole repo leases
	if err := db.RevokeRepoAgentLeases(ctx, repoID, "repository archived"); err != nil {
		t.Fatalf("RevokeRepoAgentLeases: %v", err)
	}

	// 3. Now any lease for that repo is marked revoked
	revoked, err = db.IsLeaseRevoked(ctx, "lease_active_1", repoID)
	if err != nil {
		t.Fatalf("IsLeaseRevoked: %v", err)
	}
	if !revoked {
		t.Fatal("expected lease for archived repo to be marked revoked")
	}

	// 4. A lease on a different repo is still active
	revokedOther, err := db.IsLeaseRevoked(ctx, "lease_other", "repo_other")
	if err != nil {
		t.Fatalf("IsLeaseRevoked other: %v", err)
	}
	if revokedOther {
		t.Fatal("expected lease on other repo to remain active")
	}
}
