package sshserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"
	"golang.org/x/crypto/ssh"
	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
)

func TestSSHServerAuthentication(t *testing.T) {
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	git, err := gitstore.New(filepath.Join(dataDir, "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	// 1. Generate client keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("NewPublicKey: %v", err)
	}
	fingerprint := ssh.FingerprintSHA256(sshPub)

	// 2. Enroll user and SSH key in database
	ctx := context.Background()
	user := store.User{
		ID:        "user_ssh_tester",
		Email:     "ssh@refweir.dev",
		Login:     "sshtester",
		Role:      "member",
		Status:    "active",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := db.CreateUserWithPassword(ctx, user); err != nil {
		t.Fatalf("CreateUserWithPassword: %v", err)
	}

	keyRecord := store.SSHKey{
		ID:          "sshkey_test",
		UserID:      user.ID,
		Name:        "Test Key",
		Fingerprint: fingerprint,
		PublicKey:   string(ssh.MarshalAuthorizedKey(sshPub)),
		KeyType:     sshPub.Type(),
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if err := db.AddUserSSHKey(ctx, keyRecord); err != nil {
		t.Fatalf("AddUserSSHKey: %v", err)
	}

	// 3. Start SSH Server on dynamic port
	server, err := New("127.0.0.1:0", db, git, filepath.Join(dataDir, "host_key"))
	if err != nil {
		t.Fatalf("New SSH Server: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Close()

	// 4. Test client connection with enrolled key
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("NewSignerFromKey: %v", err)
	}
	clientConfig := &ssh.ClientConfig{
		User:            "git",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         3 * time.Second,
	}

	client, err := ssh.Dial("tcp", server.Addr().String(), clientConfig)
	if err != nil {
		t.Fatalf("ssh.Dial with enrolled key failed: %v", err)
	}
	defer client.Close()

	// 5. Test connection with an unenrolled key (should fail)
	_, badPriv, _ := ed25519.GenerateKey(rand.Reader)
	badSigner, _ := ssh.NewSignerFromKey(badPriv)
	badConfig := &ssh.ClientConfig{
		User:            "git",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(badSigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         1 * time.Second,
	}

	_, err = ssh.Dial("tcp", server.Addr().String(), badConfig)
	if err == nil {
		t.Fatalf("expected dial with unenrolled key to fail, but succeeded")
	}
}

func TestSSHServerDeployKeyAuthentication(t *testing.T) {
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	git, err := gitstore.New(filepath.Join(dataDir, "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	ctx := context.Background()
	ns := store.Namespace{ID: "ns_deploy", Slug: "deploycorp", Type: "team", Name: "Deploy Corp"}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	repo := store.Repo{
		ID:            "repo_deploy_target",
		NamespaceID:   ns.ID,
		Name:          "api-service",
		FullName:      "deploycorp/api-service",
		DefaultBranch: "main",
		Visibility:    "private",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	// Generate deploy keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("NewPublicKey: %v", err)
	}
	pubKeyStr := string(ssh.MarshalAuthorizedKey(sshPub))

	// Enroll as deploy key
	_, err = db.CreateDeployKey(ctx, repo.ID, "Production CI Deploy Key", pubKeyStr, true)
	if err != nil {
		t.Fatalf("CreateDeployKey: %v", err)
	}

	// Start SSH Server
	server, err := New("127.0.0.1:0", db, git, filepath.Join(dataDir, "host_key"))
	if err != nil {
		t.Fatalf("New SSH Server: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Close()

	// Connect using deploy key
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("NewSignerFromKey: %v", err)
	}
	clientConfig := &ssh.ClientConfig{
		User:            "git",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         3 * time.Second,
	}

	client, err := ssh.Dial("tcp", server.Addr().String(), clientConfig)
	if err != nil {
		t.Fatalf("ssh.Dial with deploy key failed: %v", err)
	}
	defer client.Close()
}

func TestSSHServerCertificateAuthentication(t *testing.T) {
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	git, err := gitstore.New(filepath.Join(dataDir, "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	ctx := context.Background()
	ns := store.Namespace{ID: "ns_cert", Slug: "certcorp", Type: "team", Name: "Cert Corp"}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	// 1. Generate real CA keypair
	caPubRaw, caPrivRaw, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey CA: %v", err)
	}
	caPub, err := ssh.NewPublicKey(caPubRaw)
	if err != nil {
		t.Fatalf("NewPublicKey CA: %v", err)
	}
	caSigner, err := ssh.NewSignerFromKey(caPrivRaw)
	if err != nil {
		t.Fatalf("NewSignerFromKey CA: %v", err)
	}

	// Enroll CA in database
	_, err = db.AddSSHCA(ctx, ns.ID, "Enterprise CA", string(ssh.MarshalAuthorizedKey(caPub)))
	if err != nil {
		t.Fatalf("AddSSHCA: %v", err)
	}

	// 2. Generate user keypair
	userPubRaw, userPrivRaw, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey user: %v", err)
	}
	userPub, err := ssh.NewPublicKey(userPubRaw)
	if err != nil {
		t.Fatalf("NewPublicKey user: %v", err)
	}
	userSigner, err := ssh.NewSignerFromKey(userPrivRaw)
	if err != nil {
		t.Fatalf("NewSignerFromKey user: %v", err)
	}

	// Start SSH Server
	server, err := New("127.0.0.1:0", db, git, filepath.Join(dataDir, "host_key"))
	if err != nil {
		t.Fatalf("New SSH Server: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Close()

	now := time.Now()

	// A. Valid CA-signed certificate
	validCert := &ssh.Certificate{
		Key:             userPub,
		CertType:        ssh.UserCert,
		KeyId:           "alice@certcorp.com",
		ValidPrincipals: []string{"alice"},
		ValidAfter:      uint64(now.Add(-1 * time.Hour).Unix()),
		ValidBefore:     uint64(now.Add(1 * time.Hour).Unix()),
	}
	if err := validCert.SignCert(rand.Reader, caSigner); err != nil {
		t.Fatalf("SignCert valid: %v", err)
	}

	validCertSigner, err := ssh.NewCertSigner(validCert, userSigner)
	if err != nil {
		t.Fatalf("NewCertSigner valid: %v", err)
	}

	validConfig := &ssh.ClientConfig{
		User:            "git",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(validCertSigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         2 * time.Second,
	}

	client, err := ssh.Dial("tcp", server.Addr().String(), validConfig)
	if err != nil {
		t.Fatalf("ssh.Dial with valid CA certificate failed: %v", err)
	}
	client.Close()

	// B. Forged certificate: claims signature key is CA, but signed by attacker rogue key
	_, roguePrivRaw, _ := ed25519.GenerateKey(rand.Reader)
	rogueSigner, _ := ssh.NewSignerFromKey(roguePrivRaw)

	forgedCert := &ssh.Certificate{
		Key:             userPub,
		CertType:        ssh.UserCert,
		KeyId:           "mallory@certcorp.com",
		ValidPrincipals: []string{"mallory"},
		ValidAfter:      uint64(now.Add(-1 * time.Hour).Unix()),
		ValidBefore:     uint64(now.Add(1 * time.Hour).Unix()),
	}
	// Sign with rogue key
	if err := forgedCert.SignCert(rand.Reader, rogueSigner); err != nil {
		t.Fatalf("SignCert rogue: %v", err)
	}
	// Tamper: replace SignatureKey with authentic CA pub key
	forgedCert.SignatureKey = caPub

	forgedCertSigner, err := ssh.NewCertSigner(forgedCert, userSigner)
	if err != nil {
		t.Fatalf("NewCertSigner forged: %v", err)
	}

	forgedConfig := &ssh.ClientConfig{
		User:            "git",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(forgedCertSigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         1 * time.Second,
	}

	_, err = ssh.Dial("tcp", server.Addr().String(), forgedConfig)
	if err == nil {
		t.Fatalf("expected dial with forged certificate to FAIL, but it succeeded!")
	}

	// C. Expired certificate
	expiredCert := &ssh.Certificate{
		Key:             userPub,
		CertType:        ssh.UserCert,
		KeyId:           "alice@certcorp.com",
		ValidPrincipals: []string{"alice"},
		ValidAfter:      uint64(now.Add(-2 * time.Hour).Unix()),
		ValidBefore:     uint64(now.Add(-1 * time.Hour).Unix()),
	}
	if err := expiredCert.SignCert(rand.Reader, caSigner); err != nil {
		t.Fatalf("SignCert expired: %v", err)
	}

	expiredCertSigner, err := ssh.NewCertSigner(expiredCert, userSigner)
	if err != nil {
		t.Fatalf("NewCertSigner expired: %v", err)
	}

	expiredConfig := &ssh.ClientConfig{
		User:            "git",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(expiredCertSigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         1 * time.Second,
	}

	_, err = ssh.Dial("tcp", server.Addr().String(), expiredConfig)
	if err == nil {
		t.Fatalf("expected dial with expired certificate to FAIL, but it succeeded!")
	}
}

func TestSSHServer_PushEventHookTriggered(t *testing.T) {
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	git, err := gitstore.New(filepath.Join(dataDir, "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	keyDir := t.TempDir()
	keyPath := filepath.Join(keyDir, "id_ed25519")
	cmdKeygen := exec.Command("ssh-keygen", "-t", "ed25519", "-N", "", "-f", keyPath)
	if out, err := cmdKeygen.CombinedOutput(); err != nil {
		t.Fatalf("ssh-keygen failed: %v: %s", err, string(out))
	}
	pubBytes, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		t.Fatalf("read pubkey: %v", err)
	}
	sshPub, _, _, _, err := ssh.ParseAuthorizedKey(pubBytes)
	if err != nil {
		t.Fatalf("ParseAuthorizedKey: %v", err)
	}
	fingerprint := ssh.FingerprintSHA256(sshPub)

	ctx := context.Background()
	user := store.User{
		ID:        idgen.NewUser(),
		Email:     "sshpush@refweir.dev",
		Login:     "sshpushuser",
		Role:      "member",
		Status:    "active",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := db.CreateUserWithPassword(ctx, user); err != nil {
		t.Fatalf("CreateUserWithPassword: %v", err)
	}

	ns := store.Namespace{
		ID:          idgen.NewNS(),
		Slug:        "sshhookteam",
		Type:        "team",
		Name:        "SSH Hook Team",
		OwnerUserID: user.ID,
	}
	if err := db.CreateNamespace(ctx, ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	repo := store.Repo{
		ID:            idgen.NewRepo(),
		NamespaceID:   ns.ID,
		Name:          "ssh-repo",
		FullName:      "sshhookteam/ssh-repo",
		DefaultBranch: "main",
		Visibility:    "private",
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	keyRecord := store.SSHKey{
		ID:          idgen.New("sshkey"),
		UserID:      user.ID,
		Name:        "Push Test Key",
		Fingerprint: fingerprint,
		PublicKey:   string(ssh.MarshalAuthorizedKey(sshPub)),
		KeyType:     sshPub.Type(),
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if err := db.AddUserSSHKey(ctx, keyRecord); err != nil {
		t.Fatalf("AddUserSSHKey: %v", err)
	}

	if _, err := git.Init("sshhookteam", "ssh-repo", "main"); err != nil {
		t.Fatalf("git.Init: %v", err)
	}
	sig := object.Signature{Name: "SSH User", Email: "sshpush@refweir.dev", When: time.Now()}
	if _, err := git.WriteFiles("sshhookteam", "ssh-repo", "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# SSH Hook Repo\n"},
	}, sig, "initial commit"); err != nil {
		t.Fatalf("git.WriteFiles: %v", err)
	}

	hookCalled := make(chan string, 1)
	server, err := New("127.0.0.1:0", db, git, filepath.Join(dataDir, "host_key"))
	if err != nil {
		t.Fatalf("New SSH Server: %v", err)
	}
	server.SetPushHook(func(ctx context.Context, gotRepoID, ownerSlug, repoName string) {
		hookCalled <- gotRepoID
	})
	if err := server.Start(); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Close()

	port := server.Addr().(*net.TCPAddr).Port
	sshCmd := fmt.Sprintf("ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o IdentityAgent=none -p %d", keyPath, port)

	localRepoDir := filepath.Join(t.TempDir(), "local-ssh-repo")
	_ = os.MkdirAll(localRepoDir, 0o755)

	cmdInit := exec.Command("git", "init", "-b", "main")
	cmdInit.Dir = localRepoDir
	if out, err := cmdInit.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v: %s", err, string(out))
	}

	cRemote := exec.Command("git", "remote", "add", "origin", fmt.Sprintf("ssh://git@127.0.0.1:%d/sshhookteam/ssh-repo.git", port))
	cRemote.Dir = localRepoDir
	_ = cRemote.Run()

	cFetch := exec.Command("git", "fetch", "origin", "main")
	cFetch.Dir = localRepoDir
	cFetch.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)
	if out, err := cFetch.CombinedOutput(); err != nil {
		t.Fatalf("git fetch failed: %v: %s", err, string(out))
	}

	cReset := exec.Command("git", "reset", "--hard", "origin/main")
	cReset.Dir = localRepoDir
	_ = cReset.Run()

	if err := os.WriteFile(filepath.Join(localRepoDir, "ssh_push_test.txt"), []byte("ssh push hook verified\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cAdd := exec.Command("git", "add", "ssh_push_test.txt")
	cAdd.Dir = localRepoDir
	_ = cAdd.Run()

	cCommit := exec.Command("git", "-c", "user.name=SSH User", "-c", "user.email=sshpush@refweir.dev", "commit", "-m", "ssh push test")
	cCommit.Dir = localRepoDir
	_ = cCommit.Run()

	cPush := exec.Command("git", "push", "origin", "main")
	cPush.Dir = localRepoDir
	cPush.Env = append(os.Environ(), "GIT_SSH_COMMAND="+sshCmd)
	if out, err := cPush.CombinedOutput(); err != nil {
		t.Fatalf("git push failed: %v: %s", err, string(out))
	}

	select {
	case calledID := <-hookCalled:
		if calledID != repo.ID {
			t.Errorf("got repoID %q, want %q", calledID, repo.ID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for OnPushHook to be called after SSH git push")
	}
}
