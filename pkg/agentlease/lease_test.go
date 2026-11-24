package agentlease

import (
	"crypto/ed25519"
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

func TestAgentLease_Ed25519(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	agentID := "agent_asymmetric_1"
	repoID := "repo_asymmetric_123"
	allowedRefs := []string{"refs/heads/agent/"}
	disallowedRefs := []string{"refs/heads/main"}

	// 1. Issue with private key
	token, lease, err := IssueEd25519(priv, agentID, repoID, allowedRefs, disallowedRefs, 15*time.Minute)
	if err != nil {
		t.Fatalf("IssueEd25519 failed: %v", err)
	}

	// 2. Verify with public key
	verified, err := VerifyEd25519(pub, token)
	if err != nil {
		t.Fatalf("VerifyEd25519 failed: %v", err)
	}
	if verified.AgentID != agentID || verified.RepoID != repoID {
		t.Errorf("verified lease mismatch: %+v", verified)
	}

	// 3. Verify via polymorphic Verify(pub, token)
	verifiedPoly, err := Verify(pub, token)
	if err != nil {
		t.Fatalf("Verify with public key failed: %v", err)
	}
	if verifiedPoly.ID != lease.ID {
		t.Errorf("Verify polymorphic ID mismatch: %s != %s", verifiedPoly.ID, lease.ID)
	}

	// 4. Verify tampering is rejected
	tampered := token + "bad"
	if _, err := VerifyEd25519(pub, tampered); err == nil {
		t.Errorf("expected tampered token to fail verification")
	}

	// 5. Verify wrong public key is rejected
	wrongPub, _, _ := ed25519.GenerateKey(nil)
	if _, err := VerifyEd25519(wrongPub, token); err == nil {
		t.Errorf("expected wrong public key to fail verification")
	}
}
