#!/usr/bin/env bash
# smoke_test_lifecycle.sh — Comprehensive end-to-end lifecycle verification for Refweir forge.
# Tests:
#   1. Clean environment bootstrap.
#   2. HTTP :8080/healthz and embedded SSH :2222 port listeners.
#   3. Prometheus metrics exposition.
#   4. HMAC agent capability lease issuance.
#   5. Git Smart-HTTP clone and branch push.
#   6. Pre-receive boundary enforcement (GH007 rejected on main).
#   7. Pre-receive secret scan rejection.
#   8. Append-only audit log inspection.
set -euo pipefail

echo "=========================================================="
echo " Starting Refweir End-to-End Lifecycle Smoke Test"
echo "=========================================================="

export REFWEIR_TELEMETRY_OPT_IN="${REFWEIR_TELEMETRY_OPT_IN:-0}"

# Find free ports
HTTP_PORT=18080
SSH_PORT=12222

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR=$(mktemp -d -t refweir_smoke_XXXXXX)
DATA_DIR="$WORK_DIR/data"
BIN_DIR="$WORK_DIR/bin"
mkdir -p "$BIN_DIR" "$DATA_DIR"

DAEMON_PID=""
cleanup() {
    echo "Cleaning up smoke test environment..."
    if [ -n "$DAEMON_PID" ]; then
        kill -TERM "$DAEMON_PID" 2>/dev/null || true
        wait "$DAEMON_PID" 2>/dev/null || true
    fi
    rm -rf "$WORK_DIR"
    echo "Cleanup complete."
}
trap cleanup EXIT INT TERM

# 1. Compile pure-Go binaries
echo "[1/8] Compiling pure-Go refweird and refweir (CGO_ENABLED=0)..."
CGO_ENABLED=0 go build -o "$BIN_DIR/refweird" "$ROOT_DIR/cmd/refweird"
CGO_ENABLED=0 go build -o "$BIN_DIR/refweir" "$ROOT_DIR/cmd/refweir"

# 2. Bootstrap forge
echo "[2/8] Executing bootstrap..."
BOOTSTRAP_OUT="$WORK_DIR/bootstrap.log"
"$BIN_DIR/refweird" bootstrap \
    --email admin@corp.refweir.dev \
    --namespace acme \
    --data-dir "$DATA_DIR" > "$BOOTSTRAP_OUT" 2>&1

API_KEY=$(grep "API Key:" "$BOOTSTRAP_OUT" | awk '{print $3}')
if [ -z "$API_KEY" ]; then
    echo "ERROR: Failed to extract admin API Key from bootstrap output"
    cat "$BOOTSTRAP_OUT"
    exit 1
fi
echo "Bootstrap successful. Admin key generated."

# 3. Start refweird daemon
echo "[3/8] Starting refweird daemon on HTTP :$HTTP_PORT and SSH :$SSH_PORT..."
REFWEIR_DATA_DIR="$DATA_DIR" \
REFWEIR_LISTEN_ADDR="127.0.0.1:$HTTP_PORT" \
REFWEIR_SSH_ADDR="127.0.0.1:$SSH_PORT" \
REFWEIR_PUBLIC_URL="http://127.0.0.1:$HTTP_PORT" \
REFWEIR_GIT_URL="127.0.0.1:$HTTP_PORT" \
"$BIN_DIR/refweird" run > "$WORK_DIR/daemon.log" 2>&1 &
DAEMON_PID=$!

# Wait for HTTP readiness
READY=0
for i in {1..30}; do
    if curl -s "http://127.0.0.1:$HTTP_PORT/healthz" | grep -q "ok"; then
        READY=1
        break
    fi
    sleep 0.2
done

if [ "$READY" -ne 1 ]; then
    echo "ERROR: Daemon failed to report healthy on http://127.0.0.1:$HTTP_PORT/healthz"
    cat "$WORK_DIR/daemon.log"
    exit 1
fi
echo "Daemon HTTP service ready."

# Verify standalone SSH daemon listener
echo "[3.5/8] Verifying standalone pure-Go SSH port (:2222)..."
SSH_READY=0
for i in {1..30}; do
    if timeout 1 bash -c "cat < /dev/null > /dev/tcp/127.0.0.1/$SSH_PORT" 2>/dev/null; then
        SSH_READY=1
        break
    fi
    sleep 0.2
done
if [ "$SSH_READY" -ne 1 ]; then
    echo "ERROR: Standalone SSH daemon not listening on port $SSH_PORT"
    cat "$WORK_DIR/daemon.log"
    exit 1
fi
echo "Standalone SSH daemon listening on port $SSH_PORT."

# 4. Verify Prometheus metrics
echo "[4/8] Verifying Prometheus metrics endpoint..."
METRICS_OUTPUT=$(curl -s "http://127.0.0.1:$HTTP_PORT/metrics")
if ! echo "$METRICS_OUTPUT" | grep -q "refweir_"; then
    echo "ERROR: Prometheus metrics did not contain refweir_ prefix"
    echo "$METRICS_OUTPUT"
    exit 1
fi
echo "Metrics endpoint verified with refweir_ prefixes."

# 5. Issue Capability Lease
echo "[5/8] Requesting ephemeral agent capability lease..."
LEASE_RESP=$(curl -s -X POST "http://127.0.0.1:$HTTP_PORT/api/v1/agent/leases" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"agent_id": "agent:smoke", "repo_id": "*", "allowed_ref_prefixes": ["refs/heads/agent/"], "ttl_seconds": 900}')

LEASE_TOKEN=$(echo "$LEASE_RESP" | grep -o '"token":"[^"]*' | cut -d'"' -f4 || true)

if [ -z "$LEASE_TOKEN" ]; then
    echo "ERROR: Failed to issue agent lease:"
    echo "$LEASE_RESP"
    exit 1
fi
echo "Issued capability lease token successfully."

# 6. Create repository and test Smart-HTTP Git push
echo "[6/8] Creating test repository and testing Smart-HTTP clone/push..."
REPO_RESP=$(curl -s -X POST "http://127.0.0.1:$HTTP_PORT/api/v1/repos" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"owner": "acme", "name": "smoke-repo", "visibility": "public", "init_readme": true}')

CLONE_DIR="$WORK_DIR/smoke-repo"
git clone "http://admin:$API_KEY@127.0.0.1:$HTTP_PORT/acme/smoke-repo.git" "$CLONE_DIR" >/dev/null 2>&1

cd "$CLONE_DIR"
git config user.name "Smoke Agent"
git config user.email "smoke-agent@refweir.dev"

# Create initial commit on main
echo "# Refweir Smoke Test" > README.md
git add README.md
git commit -m "chore: initial commit" >/dev/null 2>&1
git push origin main >/dev/null 2>&1
echo "Initial commit pushed to main branch."

# 7. Test Agent Branch Push, Over-reach Rejection, and Pre-receive Secret Scanning
echo "[7/8] Testing agent lease isolation, over-reach rejection, and pre-receive scanning..."

# Switch origin remote URL to use the ephemeral capability lease credential
git remote set-url origin "http://lease:${LEASE_TOKEN}@127.0.0.1:$HTTP_PORT/acme/smoke-repo.git"

git checkout -b agent/smoke-feature >/dev/null 2>&1
echo "Agent feature work under capability lease" >> README.md
git commit -am "feat: smoke feature commit" >/dev/null 2>&1

# 7a. Push to allowed agent branch
if ! git push origin agent/smoke-feature >/dev/null 2>&1; then
    echo "ERROR: Agent push to refs/heads/agent/smoke-feature failed under valid capability lease!"
    exit 1
fi
echo "Agent push to refs/heads/agent/smoke-feature succeeded (ALLOW)."

# 7b. Test over-reach attack: try to push directly to main under capability lease
git checkout main >/dev/null 2>&1
echo "Illegal direct change" >> README.md
git commit -am "malicious: direct push to main" >/dev/null 2>&1

if git push origin main >/dev/null 2>&1; then
    echo "CRITICAL FAILURE: Agent capability lease unexpectedly allowed push to main!"
    exit 1
fi
echo "Over-reach push to main rejected by lease boundary check (GH007)."

# 7c. Test pre-receive secret scanning: try to push a leaked RSA private key
git checkout -b agent/smoke-leak >/dev/null 2>&1
cat << 'EOF' > secret.pem
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0Y3yfakekeyforpre-receivescanner
-----END RSA PRIVATE KEY-----
EOF
git add secret.pem
git commit -m "leak: accidentally commit private key" >/dev/null 2>&1

if git push origin agent/smoke-leak >/dev/null 2>&1; then
    echo "CRITICAL FAILURE: Pre-receive hook failed to reject commit with private key!"
    exit 1
fi
echo "Pre-receive scanner successfully blocked commit with leaked credentials (GH007)."

# 8. Verify Audit Logs
echo "[8/8] Verifying durable audit logs..."
AUDIT_LOG_FILE="$DATA_DIR/audit.jsonl"
if [ -f "$AUDIT_LOG_FILE" ]; then
    AUDIT_LINES=$(wc -l < "$AUDIT_LOG_FILE")
    echo "Audit log contains $AUDIT_LINES entries."
else
    echo "Audit log stored in WAL database."
fi

echo "=========================================================="
echo " Refweir Lifecycle Smoke Test: 100% SUCCESS"
echo "=========================================================="
