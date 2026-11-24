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

// Load reads configuration from environment variables and optional refweir.toml files,
// falling back to safe self-hosted defaults.
// The lease signing key is resolved from REFWEIR_JWT_SECRET or generated once under the data directory.
func Load() (Config, error) {
	dataDir := getEnv("REFWEIR_DATA_DIR", defaultDataDir)

	var tomlMap map[string]map[string]string
	cfgPath := getEnv("REFWEIR_CONFIG", "refweir.toml")
	if data, err := os.ReadFile(cfgPath); err == nil {
		tomlMap = parseSimpleTOML(string(data))
	} else if data, err := os.ReadFile(filepath.Join(dataDir, "refweir.toml")); err == nil {
		tomlMap = parseSimpleTOML(string(data))
	}

	getVal := func(envKey, tomlSec, tomlKey, def string) string {
		if v := os.Getenv(envKey); v != "" {
			return v
		}
		if tomlMap != nil && tomlMap[tomlSec] != nil {
			if v, ok := tomlMap[tomlSec][tomlKey]; ok && v != "" {
				return v
			}
		}
		return def
	}

	normalizeAddr := func(raw, def string) string {
		if raw == "" {
			return def
		}
		if raw == "none" || raw == "disabled" {
			return "none"
		}
		if strings.HasPrefix(raw, ":") {
			return "127.0.0.1" + raw
		}
		return raw
	}

	listenAddr := normalizeAddr(getVal("REFWEIR_LISTEN_ADDR", "transport", "http", defaultListen), defaultListen)
	publicURL := getVal("REFWEIR_PUBLIC_URL", "transport", "public_url", getEnvFirst([]string{"REFWEIR_ENDPOINT"}, defaultPublicURL))
	gitURL := getVal("REFWEIR_GIT_URL", "transport", "git_url", getEnvFirst([]string{"REFWEIR_GIT_HOST"}, defaultGitURL))
	sshAddr := normalizeAddr(getVal("REFWEIR_SSH_ADDR", "transport", "ssh", "127.0.0.1:2222"), "127.0.0.1:2222")
	sshHostKeyPath := getVal("REFWEIR_SSH_HOST_KEY", "transport", "ssh_host_key", filepath.Join(dataDir, "ssh_host_key"))

	jwtSecretVal := getVal("REFWEIR_JWT_SECRET", "agentlease", "secret", "")
	jwtSecret, err := resolveJWTSecret(dataDir, jwtSecretVal)
	if err != nil {
		return Config{}, err
	}

	clusterRole := getVal("REFWEIR_CLUSTER_ROLE", "cluster", "role", "primary")
	replicaDir := getVal("REFWEIR_REPLICA_DIR", "cluster", "replica_dir", "")
	nodeID := getVal("REFWEIR_NODE_ID", "cluster", "node_id", "")

	smtpHost := getVal("REFWEIR_SMTP_HOST", "smtp", "host", "")
	smtpPort := 587
	if pStr := getVal("REFWEIR_SMTP_PORT", "smtp", "port", ""); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			smtpPort = p
		}
	}
	smtpUsername := getVal("REFWEIR_SMTP_USERNAME", "smtp", "username", "")
	smtpPassword := getVal("REFWEIR_SMTP_PASSWORD", "smtp", "password", "")
	smtpFrom := getVal("REFWEIR_SMTP_FROM", "smtp", "from", "noreply@refweir.dev")
	smtpTLS := getVal("REFWEIR_SMTP_TLS", "smtp", "tls", "true") != "false"

	version := getEnv("REFWEIR_VERSION", "dev")

	telemetryOpt := getVal("REFWEIR_TELEMETRY_OPT_IN", "telemetry", "enabled", "")
	telemetryEnabled := telemetryOpt == "1" || telemetryOpt == "true"
	if !telemetryEnabled {
		if body, err := os.ReadFile(filepath.Join(dataDir, "telemetry_opt_in")); err == nil { // #nosec G304 -- operator-configured data directory
			trimmed := strings.TrimSpace(string(body))
			telemetryEnabled = trimmed == "1" || trimmed == "true"
		}
	}
	telemetryEndpoint := getVal("REFWEIR_TELEMETRY_ENDPOINT", "telemetry", "endpoint", "")

	demoMode := (getVal("REFWEIR_DEMO_MODE", "demo", "enabled", "") == "1" || getVal("REFWEIR_DEMO_MODE", "demo", "enabled", "") == "true")
	demoRepo := getVal("REFWEIR_DEMO_REPO", "demo", "repo", "demo/agent-lab")
	demoSelfURL := strings.TrimRight(getVal("REFWEIR_DEMO_SELF_URL", "demo", "self_url", ""), "/")

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
func resolveJWTSecret(dataDir string, configuredSecret string) (string, error) {
	if configuredSecret != "" {
		if len(configuredSecret) < minJWTSecretBytes {
			return "", fmt.Errorf("REFWEIR_JWT_SECRET must be at least %d bytes (got %d)", minJWTSecretBytes, len(configuredSecret))
		}
		return configuredSecret, nil
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

// parseSimpleTOML parses basic [section] and key = "value" pairs from a TOML string.
func parseSimpleTOML(content string) map[string]map[string]string {
	result := make(map[string]map[string]string)
	currentSection := "global"

	lines := strings.Split(content, "\n")
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip inline comment (# not inside quotes)
		inQuotes := false
		splitIdx := -1
		for i, r := range line {
			if r == '"' {
				inQuotes = !inQuotes
			} else if r == '#' && !inQuotes {
				splitIdx = i
				break
			}
		}
		if splitIdx != -1 {
			line = strings.TrimSpace(line[:splitIdx])
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			if _, exists := result[currentSection]; !exists {
				result[currentSection] = make(map[string]string)
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			val := strings.TrimSpace(parts[1])
			if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1 : len(val)-1]
			}
			if _, exists := result[currentSection]; !exists {
				result[currentSection] = make(map[string]string)
			}
			result[currentSection][key] = val
		}
	}
	return result
}
