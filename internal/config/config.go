// Package config manages runtime configuration for refweir.dev services,
// reading environment variables and providing self-hosted defaults.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultListen    = "127.0.0.1:8080"
	defaultPublicURL = "http://127.0.0.1:8080"
	defaultGitURL    = "127.0.0.1:8080"
	defaultDataDir   = "data"

	minJWTSecretBytes = 32
)

// Config holds runtime configuration for the refweir.dev server and CLI.
// Defaults are strictly self-hosted and never reference Cursor infrastructure.
type Config struct {
	Version    string
	ListenAddr string
	DataDir    string
	PublicURL  string
	GitURL     string

	// Endpoint and GitHost are compatibility aliases for PublicURL and GitURL.
	Endpoint string
	GitHost  string

	SSHAddr        string
	SSHHostKeyPath string
	JWTSecret      string

	ClusterRole string
	ReplicaDir  string
	NodeID      string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPTLS      bool

	TelemetryEnabled  bool
	TelemetryEndpoint string

	// DemoMode enables the hosted-demo surface: a persistent banner and a bounded,
	// rate-limited "try an agent run" endpoint. It must stay off for normal
	// self-hosted deployments.
	DemoMode bool
	// DemoRepo is the owner/name of the seeded repository used by the demo run.
	DemoRepo string
	// DemoSelfURL is the base URL the demo runner uses to reach its own Git
	// transport. Empty means derive it from ListenAddr.
	DemoSelfURL string
}

// Load reads configuration from environment variables, falling back to safe self-hosted defaults.
// The lease signing key is resolved from REFWEIR_JWT_SECRET or generated once under the data directory.
func Load() (Config, error) {
	listenAddr := getEnv("REFWEIR_LISTEN_ADDR", defaultListen)
	dataDir := getEnv("REFWEIR_DATA_DIR", defaultDataDir)

	publicURL := getEnvFirst([]string{"REFWEIR_PUBLIC_URL", "REFWEIR_ENDPOINT"}, defaultPublicURL)
	gitURL := getEnvFirst([]string{"REFWEIR_GIT_URL", "REFWEIR_GIT_HOST"}, defaultGitURL)
	sshAddr := getEnv("REFWEIR_SSH_ADDR", "127.0.0.1:2222")
	sshHostKeyPath := getEnv("REFWEIR_SSH_HOST_KEY", dataDir+"/ssh_host_key")

	jwtSecret, err := resolveJWTSecret(dataDir)
	if err != nil {
		return Config{}, err
	}

	clusterRole := getEnv("REFWEIR_CLUSTER_ROLE", "primary")
	replicaDir := getEnv("REFWEIR_REPLICA_DIR", "")
	nodeID := getEnv("REFWEIR_NODE_ID", "")

	smtpHost := getEnv("REFWEIR_SMTP_HOST", "")
	smtpPort := 587
	if pStr := getEnv("REFWEIR_SMTP_PORT", ""); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			smtpPort = p
		}
	}
	smtpUsername := getEnv("REFWEIR_SMTP_USERNAME", "")
	smtpPassword := getEnv("REFWEIR_SMTP_PASSWORD", "")
	smtpFrom := getEnv("REFWEIR_SMTP_FROM", "noreply@refweir.dev")
	smtpTLS := getEnv("REFWEIR_SMTP_TLS", "true") != "false"

	version := getEnv("REFWEIR_VERSION", "dev")

	telemetryEnabled := getEnv("REFWEIR_TELEMETRY_OPT_IN", "") == "1"
	if !telemetryEnabled {
		if body, err := os.ReadFile(filepath.Join(dataDir, "telemetry_opt_in")); err == nil { // #nosec G304 -- operator-configured data directory
			telemetryEnabled = strings.TrimSpace(string(body)) == "1"
		}
	}
	telemetryEndpoint := getEnv("REFWEIR_TELEMETRY_ENDPOINT", "")

	demoMode := getEnv("REFWEIR_DEMO_MODE", "") == "1"
	demoRepo := getEnv("REFWEIR_DEMO_REPO", "demo/agent-lab")
	demoSelfURL := strings.TrimRight(getEnv("REFWEIR_DEMO_SELF_URL", ""), "/")

	return Config{
		Version:        version,
		ListenAddr:     listenAddr,
		DataDir:        dataDir,
		PublicURL:      publicURL,
		GitURL:         gitURL,
		Endpoint:       publicURL,
		GitHost:        gitURL,
		SSHAddr:        sshAddr,
		SSHHostKeyPath: sshHostKeyPath,
		JWTSecret:      jwtSecret,
		ClusterRole:    clusterRole,
		ReplicaDir:     replicaDir,
		NodeID:         nodeID,
		SMTPHost:       smtpHost,
		SMTPPort:       smtpPort,
		SMTPUsername:   smtpUsername,
		SMTPPassword:   smtpPassword,
		SMTPFrom:       smtpFrom,
		SMTPTLS:        smtpTLS,

		TelemetryEnabled:  telemetryEnabled,
		TelemetryEndpoint: telemetryEndpoint,
		DemoMode:          demoMode,
		DemoRepo:          demoRepo,
		DemoSelfURL:       demoSelfURL,
	}, nil
}

// resolveJWTSecret returns the lease signing key, generating and persisting one
// under dataDir on first use. It never falls back to a literal in the source.
func resolveJWTSecret(dataDir string) (string, error) {
	if v := os.Getenv("REFWEIR_JWT_SECRET"); v != "" {
		if len(v) < minJWTSecretBytes {
			return "", fmt.Errorf("REFWEIR_JWT_SECRET must be at least %d bytes (got %d)", minJWTSecretBytes, len(v))
		}
		return v, nil
	}

	path := filepath.Join(dataDir, "jwt_secret")
	if body, err := os.ReadFile(path); err == nil { // #nosec G304 -- path comes from operator configuration or a validated internal identifier
		secret := strings.TrimSpace(string(body))
		if len(secret) >= minJWTSecretBytes {
			return secret, nil
		}
	}

	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", fmt.Errorf("create data directory %s: %w", dataDir, err)
	}

	buf := make([]byte, minJWTSecretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate lease signing key: %w", err)
	}
	secret := hex.EncodeToString(buf)

	if err := os.WriteFile(path, []byte(secret), 0o600); err != nil {
		return "", fmt.Errorf("write lease signing key %s: %w", path, err)
	}
	log.Printf("generated lease signing key at %s", path)
	return secret, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvFirst(keys []string, fallback string) string {
	for _, key := range keys {
		if v, ok := os.LookupEnv(key); ok && v != "" {
			return v
		}
	}
	return fallback
}
