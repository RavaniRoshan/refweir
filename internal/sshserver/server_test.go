package sshserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"path/filepath"
	"testing"
	"time"

	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/store"
	"golang.org/x/crypto/ssh"
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
