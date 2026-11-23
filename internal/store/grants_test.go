package store

import (
	"context"
	"testing"

	"refweir.dev/pkg/idgen"
)

func seedTestUser(t *testing.T, db *DB, id, email string) {
	t.Helper()
	_, err := db.sql.Exec(`INSERT INTO users (id, email, created_at) VALUES (?, ?, ?)`, id, email, now())
	if err != nil {
		t.Fatalf("seedTestUser: %v", err)
	}
}

// TestCheckRepoAccess_Precedence proves each rung of the access control precedence order:
// Rung 1: Owner (namespace owner has full access)
// Rung 2: Repository grant (explicit grant on the repository)
// Rung 3: Namespace grant (inherited grant on the namespace)
// Rung 4: Deny (no matching grant)
func TestCheckRepoAccess_Precedence(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()

	ownerID := idgen.NewUser()
	repoWriteID := idgen.NewUser()
	repoReadID := idgen.NewUser()
	nsAdminID := idgen.NewUser()
	nsReadID := idgen.NewUser()
	strangerID := idgen.NewUser()

	seedTestUser(t, db, ownerID, "owner@example.com")
	seedTestUser(t, db, repoWriteID, "repowrite@example.com")
	seedTestUser(t, db, repoReadID, "reporead@example.com")
	seedTestUser(t, db, nsAdminID, "nsadmin@example.com")
	seedTestUser(t, db, nsReadID, "nsread@example.com")
	seedTestUser(t, db, strangerID, "stranger@example.com")

	ns := Namespace{
		ID:          idgen.NewNS(),
		Slug:        "precedence-org",
		Type:        "team",
		Name:        "Precedence Org",
		OwnerUserID: ownerID,
	}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	repo := Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "precedence-repo",
		FullName:      "precedence-org/precedence-repo",
		DefaultBranch: "main",
		Visibility:    "private",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// Setup Rung 2: Repository grants
	_, err := db.UpsertGrant(ctx, Grant{
		ID:            idgen.NewGrant(),
		Scope:         "repository",
		NamespaceID:   ns.ID,
		RepoID:        repo.ID,
		PrincipalKind: "user",
		UserID:        repoWriteID,
		Permission:    "permission_write",
	})
	if err != nil {
		t.Fatalf("UpsertGrant repo write: %v", err)
	}

	_, err = db.UpsertGrant(ctx, Grant{
		ID:            idgen.NewGrant(),
		Scope:         "repository",
		NamespaceID:   ns.ID,
		RepoID:        repo.ID,
		PrincipalKind: "user",
		UserID:        repoReadID,
		Permission:    "permission_read",
	})
	if err != nil {
		t.Fatalf("UpsertGrant repo read: %v", err)
	}

	// Setup Rung 3: Namespace grants
	_, err = db.UpsertGrant(ctx, Grant{
		ID:            idgen.NewGrant(),
		Scope:         "namespace",
		NamespaceID:   ns.ID,
		PrincipalKind: "user",
		UserID:        nsAdminID,
		Permission:    "permission_admin",
	})
	if err != nil {
		t.Fatalf("UpsertGrant ns admin: %v", err)
	}

	_, err = db.UpsertGrant(ctx, Grant{
		ID:            idgen.NewGrant(),
		Scope:         "namespace",
		NamespaceID:   ns.ID,
		PrincipalKind: "user",
		UserID:        nsReadID,
		Permission:    "permission_read",
	})
	if err != nil {
		t.Fatalf("UpsertGrant ns read: %v", err)
	}

	tests := []struct {
		name       string
		userID     string
		write      bool
		wantAccess bool
	}{
		// Rung 1: Owner
		{"Rung 1: Owner read", ownerID, false, true},
		{"Rung 1: Owner write", ownerID, true, true},

		// Rung 2: Repository grant
		{"Rung 2: Repo grant write allows write", repoWriteID, true, true},
		{"Rung 2: Repo grant write allows read", repoWriteID, false, true},
		{"Rung 2: Repo grant read allows read", repoReadID, false, true},
		{"Rung 2: Repo grant read denies write", repoReadID, true, false},

		// Rung 3: Namespace grant
		{"Rung 3: NS grant admin allows write", nsAdminID, true, true},
		{"Rung 3: NS grant admin allows read", nsAdminID, false, true},
		{"Rung 3: NS grant read allows read", nsReadID, false, true},
		{"Rung 3: NS grant read denies write", nsReadID, true, false},

		// Rung 4: Deny
		{"Rung 4: Stranger read denied", strangerID, false, false},
		{"Rung 4: Stranger write denied", strangerID, true, false},
		{"Empty user denied", "", false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := db.CheckRepoAccess(ctx, repo, tc.userID, tc.write)
			if err != nil {
				t.Fatalf("CheckRepoAccess unexpected error: %v", err)
			}
			if got != tc.wantAccess {
				t.Errorf("CheckRepoAccess(userID=%s, write=%v) = %v; want %v", tc.userID, tc.write, got, tc.wantAccess)
			}
		})
	}

	// Test EnsureNamespaceAdminGrant on create unblocks stranger
	if err := db.EnsureNamespaceAdminGrant(ctx, ns.ID, strangerID); err != nil {
		t.Fatalf("EnsureNamespaceAdminGrant: %v", err)
	}
	allowed, err := db.CheckRepoAccess(ctx, repo, strangerID, true)
	if err != nil || !allowed {
		t.Errorf("CheckRepoAccess after EnsureNamespaceAdminGrant = %v (err=%v), want true", allowed, err)
	}

	// Re-running EnsureNamespaceAdminGrant should be idempotent
	if err := db.EnsureNamespaceAdminGrant(ctx, ns.ID, strangerID); err != nil {
		t.Errorf("EnsureNamespaceAdminGrant idempotent call failed: %v", err)
	}
}
