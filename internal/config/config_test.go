package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSecret = "test-only-jwt-secret-that-is-32-bytes!"

func TestConfigDefaults(t *testing.T) {
	t.Setenv("REFWEIR_JWT_SECRET", testSecret)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if strings.Contains(cfg.PublicURL, "cursor.com") {
		t.Errorf("self-hosted public URL must not default to Cursor: %s", cfg.PublicURL)
	}
	if strings.Contains(cfg.GitURL, "cursor.com") {
		t.Errorf("self-hosted git URL must not default to Cursor: %s", cfg.GitURL)
	}
	if cfg.PublicURL != "http://127.0.0.1:8080" {
		t.Errorf("publicURL = %s, want http://127.0.0.1:8080", cfg.PublicURL)
	}
	if cfg.GitURL != "127.0.0.1:8080" {
		t.Errorf("gitURL = %s, want 127.0.0.1:8080", cfg.GitURL)
	}
	if cfg.ListenAddr != "127.0.0.1:8080" {
		t.Errorf("listenAddr = %s, want 127.0.0.1:8080", cfg.ListenAddr)
	}
	if cfg.DataDir != "data" {
		t.Errorf("dataDir = %s, want data", cfg.DataDir)
	}
	if cfg.Endpoint != cfg.PublicURL {
		t.Errorf("endpoint alias mismatch: got %s, want %s", cfg.Endpoint, cfg.PublicURL)
	}
	if cfg.GitHost != cfg.GitURL {
		t.Errorf("gitHost alias mismatch: got %s, want %s", cfg.GitHost, cfg.GitURL)
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	t.Setenv("REFWEIR_PUBLIC_URL", "https://forge.example.com")
	t.Setenv("REFWEIR_GIT_URL", "git.example.com")
	t.Setenv("REFWEIR_LISTEN_ADDR", "0.0.0.0:9090")
	t.Setenv("REFWEIR_DATA_DIR", t.TempDir())
	t.Setenv("REFWEIR_JWT_SECRET", testSecret)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.PublicURL != "https://forge.example.com" {
		t.Errorf("publicURL = %s", cfg.PublicURL)
	}
	if cfg.GitURL != "git.example.com" {
		t.Errorf("gitURL = %s", cfg.GitURL)
	}
	if cfg.ListenAddr != "0.0.0.0:9090" {
		t.Errorf("listenAddr = %s", cfg.ListenAddr)
	}
}

func TestConfigLegacyEnvOverrides(t *testing.T) {
	t.Setenv("REFWEIR_ENDPOINT", "https://legacy.example.com")
	t.Setenv("REFWEIR_GIT_HOST", "legacy-git.example.com")
	t.Setenv("REFWEIR_JWT_SECRET", testSecret)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.PublicURL != "https://legacy.example.com" {
		t.Errorf("publicURL from legacy REFWEIR_ENDPOINT = %s", cfg.PublicURL)
	}
	if cfg.GitURL != "legacy-git.example.com" {
		t.Errorf("gitURL from legacy REFWEIR_GIT_HOST = %s", cfg.GitURL)
	}
}

func TestJWTSecret_NoLiteralDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("REFWEIR_DATA_DIR", dir)
	if err := os.Unsetenv("REFWEIR_JWT_SECRET"); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.JWTSecret == "" {
		t.Fatal("JWT secret is empty")
	}
	removedDefault := strings.Join([]string{"refweir-ephemeral", "agent-lease", "signing-key"}, "-")
	if cfg.JWTSecret == removedDefault {
		t.Fatal("JWT secret equals the removed literal default")
	}
	if len(cfg.JWTSecret) < minJWTSecretBytes {
		t.Fatalf("JWT secret is %d bytes, want >= %d", len(cfg.JWTSecret), minJWTSecretBytes)
	}

	info, err := os.Stat(filepath.Join(dir, "jwt_secret"))
	if err != nil {
		t.Fatalf("stat jwt_secret: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("jwt_secret mode = %o, want 600", info.Mode().Perm())
	}
}

func TestJWTSecret_Persists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("REFWEIR_DATA_DIR", dir)
	if err := os.Unsetenv("REFWEIR_JWT_SECRET"); err != nil {
		t.Fatal(err)
	}

	first, err := Load()
	if err != nil {
		t.Fatalf("first Load: %v", err)
	}
	second, err := Load()
	if err != nil {
		t.Fatalf("second Load: %v", err)
	}
	if first.JWTSecret != second.JWTSecret {
		t.Fatalf("secret changed across loads: %q != %q", first.JWTSecret, second.JWTSecret)
	}

	body, err := os.ReadFile(filepath.Join(dir, "jwt_secret"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(body)) != first.JWTSecret {
		t.Fatal("persisted secret does not match the loaded secret")
	}
}

func TestJWTSecret_RejectsShortSecret(t *testing.T) {
	t.Setenv("REFWEIR_JWT_SECRET", "too-short")
	if _, err := Load(); err == nil {
		t.Fatal("Load accepted a secret shorter than 32 bytes")
	}
}

func TestConfigTOMLFile(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "refweir.toml")
	tomlContent := `
# Refweir policy and server configuration
[transport]
http = ":8088"
ssh = ":2224"
public_url = "https://forge.acme.corp"
git_url = "forge.acme.corp"
`
	if err := os.WriteFile(tomlPath, []byte(tomlContent), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("REFWEIR_CONFIG", tomlPath)
	t.Setenv("REFWEIR_DATA_DIR", dir)
	t.Setenv("REFWEIR_JWT_SECRET", testSecret)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.ListenAddr != "127.0.0.1:8088" {
		t.Errorf("listenAddr = %s, want 127.0.0.1:8088", cfg.ListenAddr)
	}
	if cfg.SSHAddr != "127.0.0.1:2224" {
		t.Errorf("sshAddr = %s, want 127.0.0.1:2224", cfg.SSHAddr)
	}
	if cfg.PublicURL != "https://forge.acme.corp" {
		t.Errorf("publicURL = %s, want https://forge.acme.corp", cfg.PublicURL)
	}
	if cfg.GitURL != "forge.acme.corp" {
		t.Errorf("gitURL = %s, want forge.acme.corp", cfg.GitURL)
	}
}

func TestConfigTOML_FullFields(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "refweir.toml")
	tomlContent := `
[transport]
http = ":9090"
ssh = ":2225"
public_url = "https://forge.full.test"
git_url = "git.full.test"

[agentlease]
secret = "configured-via-toml-secret-that-is-at-least-32-bytes!"

[smtp]
host = "smtp.mailgun.org"
port = "465"
username = "postmaster@mailgun.org"
password = "supersecretpassword"
from = "alerts@refweir.dev"
tls = "true"

[telemetry]
enabled = "true"
endpoint = "https://telemetry.refweir.dev/v1/events"

[cluster]
role = "replica"
replica_dir = "/var/lib/refweir/replica"
node_id = "node-alpha"
`
	if err := os.WriteFile(tomlPath, []byte(tomlContent), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("REFWEIR_CONFIG", tomlPath)
	t.Setenv("REFWEIR_DATA_DIR", dir)
	if err := os.Unsetenv("REFWEIR_JWT_SECRET"); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.ListenAddr != "127.0.0.1:9090" {
		t.Errorf("listenAddr = %s, want 127.0.0.1:9090", cfg.ListenAddr)
	}
	if cfg.JWTSecret != "configured-via-toml-secret-that-is-at-least-32-bytes!" {
		t.Errorf("jwtSecret = %s, want configured-via-toml-secret-that-is-at-least-32-bytes!", cfg.JWTSecret)
	}
	if cfg.SMTPHost != "smtp.mailgun.org" {
		t.Errorf("smtpHost = %s", cfg.SMTPHost)
	}
	if cfg.SMTPPort != 465 {
		t.Errorf("smtpPort = %d, want 465", cfg.SMTPPort)
	}
	if cfg.SMTPUsername != "postmaster@mailgun.org" {
		t.Errorf("smtpUsername = %s", cfg.SMTPUsername)
	}
	if cfg.SMTPPassword != "supersecretpassword" {
		t.Errorf("smtpPassword = %s", cfg.SMTPPassword)
	}
	if cfg.SMTPFrom != "alerts@refweir.dev" {
		t.Errorf("smtpFrom = %s", cfg.SMTPFrom)
	}
	if !cfg.SMTPTLS {
		t.Errorf("smtpTLS = %v, want true", cfg.SMTPTLS)
	}
	if !cfg.TelemetryEnabled {
		t.Errorf("telemetryEnabled = %v, want true", cfg.TelemetryEnabled)
	}
	if cfg.TelemetryEndpoint != "https://telemetry.refweir.dev/v1/events" {
		t.Errorf("telemetryEndpoint = %s", cfg.TelemetryEndpoint)
	}
	if cfg.ClusterRole != "replica" {
		t.Errorf("clusterRole = %s, want replica", cfg.ClusterRole)
	}
	if cfg.ReplicaDir != "/var/lib/refweir/replica" {
		t.Errorf("replicaDir = %s", cfg.ReplicaDir)
	}
	if cfg.NodeID != "node-alpha" {
		t.Errorf("nodeID = %s", cfg.NodeID)
	}
}
