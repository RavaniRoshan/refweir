package agentlease

import (
	"testing"
	"time"
)

func TestAgentLease_Lifecycle(t *testing.T) {
	secret := []byte("super-secret-cluster-signing-key-123")
	agentID := "agent_worker_claude_1"
	repoID := "repo_test_123"
	allowedRefs := []string{"refs/heads/agent/", "refs/heads/feature/"}
	disallowedRefs := []string{"refs/heads/main", "refs/heads/master", "refs/tags/"}

	// 1. Issue Lease
	token, lease, err := Issue(secret, agentID, repoID, allowedRefs, disallowedRefs, 10*time.Minute)
	if err != nil {
		t.Fatalf("issue lease failed: %v", err)
	}
	if token == "" || lease == nil {
		t.Fatalf("expected non-empty token and lease")
	}

	// 2. Verify Valid Lease
	verified, err := Verify(secret, token)
	if err != nil {
		t.Fatalf("verify valid lease failed: %v", err)
	}
	if verified.AgentID != agentID {
		t.Errorf("got agentID %q, want %q", verified.AgentID, agentID)
	}
	if verified.RepoID != repoID {
		t.Errorf("got repoID %q, want %q", verified.RepoID, repoID)
	}

	// 3. Test Ref Permissions
	if !verified.CanPushRef("refs/heads/agent/fix-bug-42") {
		t.Errorf("expected permission for refs/heads/agent/fix-bug-42, got false")
	}
	if !verified.CanPushRef("refs/heads/feature/auth") {
		t.Errorf("expected permission for refs/heads/feature/auth, got false")
	}
	if verified.CanPushRef("refs/heads/main") {
		t.Errorf("expected rejection for refs/heads/main, got true")
	}
	if verified.CanPushRef("refs/heads/master") {
		t.Errorf("expected rejection for refs/heads/master, got true")
	}
	if verified.CanPushRef("refs/heads/random/branch") {
		t.Errorf("expected rejection for random branch outside allowed prefix, got true")
	}

	// 4. Test Tampered Token
	tampered := token + "corrupt"
	if _, err := Verify(secret, tampered); err == nil {
		t.Fatalf("expected error for tampered token, got nil")
	}

	// 5. Test Wrong Secret
	wrongSecret := []byte("wrong-signing-secret")
	if _, err := Verify(wrongSecret, token); err == nil {
		t.Fatalf("expected error for wrong secret, got nil")
	}

	// 6. Test Expired Lease
	expiredToken, _, err := Issue(secret, agentID, repoID, allowedRefs, disallowedRefs, -1*time.Minute)
	if err != nil {
		t.Fatalf("issue expired lease failed: %v", err)
	}
	if _, err := Verify(secret, expiredToken); err != ErrExpiredLease {
		t.Fatalf("expected ErrExpiredLease, got %v", err)
	}
}

func TestAgentLease_RefTraversalAndIsolation(t *testing.T) {
	secret := []byte("cluster-signing-key-456")
	allowedRefs := []string{"refs/heads/agent/"}
	disallowedRefs := []string{"refs/heads/main"}

	_, lease, err := Issue(secret, "agent_1", "repo_1", allowedRefs, disallowedRefs, 10*time.Minute)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	// Traversal attacks must be rejected
	traversals := []string{
		"refs/heads/agent/../../heads/main",
		"refs/heads/agent/../main",
		"refs/heads/agent/../heads/main",
		"refs/heads/agent//main",
		"refs/heads/agent/..",
		"refs/heads/agent\\main",
	}
	for _, target := range traversals {
		if lease.CanPushRef(target) {
			t.Errorf("expected CanPushRef(%q) = false, got true (path traversal bypass!)", target)
		}
	}

	// Sibling branch without trailing slash must not match
	if lease.CanPushRef("refs/heads/agent-production") {
		t.Errorf("expected CanPushRef(refs/heads/agent-production) = false, got true (sibling branch bypass!)")
	}

	// Legitimate child branch must succeed
	if !lease.CanPushRef("refs/heads/agent/task-123") {
		t.Errorf("expected CanPushRef(refs/heads/agent/task-123) = true, got false")
	}
}
