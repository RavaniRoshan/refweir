#!/usr/bin/env bash
# Reproducible containment demo for Refweir.
#
# Boots a throwaway self-hosted instance, issues scoped capability leases, and
# demonstrates the containment boundary:
#   1. Agent A with a lease for refs/heads/agent/* pushes successfully.
#   2. Agent B with a lease attempts refs/heads/main and is rejected (GH007).
#   3. An expired lease is rejected at authentication.
#   4. A revoked lease is rejected at authentication.
#   5. The durable audit log shows every decision.
#
# Requires: go, git, curl, python3.
# Usage:    examples/containment/run.sh
# Env:      REFWEIR_DEMO_PORT (default: random free port)
#           KEEP=1 to keep the temp workspace and print its path.

set -euo pipefail

export GIT_TERMINAL_PROMPT=0

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

need() {
    command -v "$1" >/dev/null 2>&1 || {
        echo "error: '$1' is required but not installed" >&2
        exit 1
    }
}
for cmd in go git curl python3; do need "$cmd"; done

WORK="$(mktemp -d "${TMPDIR:-/tmp}/refweir-containment.XXXXXX")"
DATA="$WORK/data"
BIN="$WORK/bin"
CLONE="$WORK/repo"
AUDIT_LOG="$DATA/audit.log"
DAEMON_LOG="$WORK/daemon.log"
DAEMON_PID=""

cleanup() {
    if [ -n "$DAEMON_PID" ] && kill -0 "$DAEMON_PID" 2>/dev/null; then
        kill "$DAEMON_PID" 2>/dev/null || true
        wait "$DAEMON_PID" 2>/dev/null || true
    fi
    if [ "${KEEP:-0}" = "1" ]; then
        echo
        echo "workspace kept at: $WORK"
        echo "audit log:         $AUDIT_LOG"
    else
        rm -rf "$WORK"
    fi
}
trap cleanup EXIT

# PACING=1 slows the demo for a screen recording. PACING_MARKERS, when set,
# receives "<section> <epoch-seconds>" lines so the recording can time captions.
# Neither changes what the demo does; the default run stays fast.
PACING="${PACING:-0}"
mark() {
    if [ -n "${PACING_MARKERS:-}" ]; then
        printf '%s %s\n' "$1" "$(date +%s.%N)" >>"$PACING_MARKERS"
    fi
}
beat() {
    if [ "$PACING" = "1" ]; then
        sleep "${1:-1}"
    fi
}

fail() {
    echo
    echo "FAIL: $*" >&2
    if [ -f "$DAEMON_LOG" ]; then
        echo "--- daemon log (tail) ---" >&2
        tail -30 "$DAEMON_LOG" >&2 || true
    fi
    exit 1
}

json_field() {
    python3 -c 'import json,sys; print(json.load(sys.stdin)[sys.argv[1]])' "$1"
}

audit_contains() {
    grep -q "$1" "$AUDIT_LOG" 2>/dev/null || fail "audit log is missing $1"
}

if [ -n "${REFWEIR_DEMO_PORT:-}" ]; then
    PORT="$REFWEIR_DEMO_PORT"
else
    PORT="$(python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
)"
fi
BASE="http://127.0.0.1:${PORT}"

mark run_start
echo "================================================================"
echo " Refweir containment demo"
echo " workspace: $WORK"
echo " server:    $BASE"
echo "================================================================"
beat 2

mkdir -p "$BIN" "$DATA"

echo
echo "[build] compiling refweird with CGO disabled..."
(cd "$ROOT_DIR" && CGO_ENABLED=0 go build -o "$BIN/refweird" ./cmd/refweird) \
    || fail "go build failed"

echo "[bootstrap] creating admin, namespace, and demo repository..."
REFWEIR_TELEMETRY_OPT_IN=0 "$BIN/refweird" bootstrap \
    --email admin@example.com \
    --namespace acme \
    --data-dir "$DATA" >"$WORK/bootstrap.log" 2>&1 \
    || { cat "$WORK/bootstrap.log" >&2; fail "bootstrap failed"; }

API_KEY="$(awk '/^API Key:/ {print $3; exit}' "$WORK/bootstrap.log")"
FULL_NAME="$(awk '/^Repository:/ {print $2; exit}' "$WORK/bootstrap.log")"
[ -n "$API_KEY" ] || fail "could not parse the bootstrap API key"
[ -n "$FULL_NAME" ] || fail "could not parse the bootstrap repository name"

echo "[serve] starting refweird..."
REFWEIR_LISTEN_ADDR="127.0.0.1:${PORT}" \
REFWEIR_PUBLIC_URL="$BASE" \
REFWEIR_GIT_URL="127.0.0.1:${PORT}" \
REFWEIR_SSH_ADDR=none \
REFWEIR_DATA_DIR="$DATA" \
    "$BIN/refweird" >"$DAEMON_LOG" 2>&1 &
DAEMON_PID=$!

ready=0
for _ in $(seq 1 100); do
    if curl -fsS "$BASE/api/v1/system/info" >/dev/null 2>&1; then
        ready=1
        break
    fi
    if ! kill -0 "$DAEMON_PID" 2>/dev/null; then
        break
    fi
    sleep 0.1
done
[ "$ready" = "1" ] || fail "server did not become ready"

REPOS_JSON="$(curl -fsS -H "Authorization: Bearer $API_KEY" "$BASE/api/v1/repos")"
REPO_ID="$(python3 -c 'import json,sys; print(json.load(sys.stdin)["repositories"][0]["id"])' <<<"$REPOS_JSON")"

if [ -z "$REPO_ID" ] || [ "$REPO_ID" = "None" ]; then
    fail "could not resolve the demo repository ID"
fi

issue_lease() {
    local agent="$1" allowed="$2" ttl="$3"
    curl -fsS \
        -H "Authorization: Bearer $API_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"agentId\":\"${agent}\",\"repoId\":\"${REPO_ID}\",\"allowedRefPrefixes\":${allowed},\"disallowedRefPrefixes\":[\"refs/heads/main\",\"refs/heads/master\"],\"ttlSeconds\":${ttl}}" \
        "$BASE/api/v1/agent/leases"
}

clone_url() { printf 'http://admin:%s@127.0.0.1:%s/%s.git' "$API_KEY" "$PORT" "$FULL_NAME"; }
push_url() { printf 'http://%s:%s@127.0.0.1:%s/%s.git' "$1" "$2" "$PORT" "$FULL_NAME"; }

echo "[clone] $FULL_NAME"
git clone -q "$(clone_url)" "$CLONE" || fail "clone failed"
cd "$CLONE"
git config user.email "agent-a@example.com"
git config user.name "Agent A"
git checkout -q -b agent/feature-123
printf '\nContainment demo change made by Agent A.\n' >>README.md
git add README.md
git commit -q -m "agent: add containment demo note"
beat 1

# ---------------------------------------------------------------------------
mark step1
echo
echo "== 1. Valid lease: Agent A pushes refs/heads/agent/feature-123 =="
LEASE_A_JSON="$(issue_lease agent-a '["refs/heads/agent/"]' 900)" || fail "lease A issuance failed"
LEASE_A="$(json_field token <<<"$LEASE_A_JSON")"
LEASE_A_ID="$(json_field id <<<"$LEASE_A_JSON")"
LEASE_A_EXPIRES="$(json_field expiresAt <<<"$LEASE_A_JSON")"
echo "   lease:   $LEASE_A_ID"
echo "   agent:   agent-a"
echo "   scope:   refs/heads/agent/ (main/master explicitly denied)"
echo "   expires: $LEASE_A_EXPIRES"

if ! git push -q "$(push_url agent-a "$LEASE_A")" HEAD:refs/heads/agent/feature-123 2>"$WORK/push-a.log"; then
    cat "$WORK/push-a.log" >&2
    fail "Agent A should have been allowed to push refs/heads/agent/feature-123"
fi
echo "   result:  ACCEPTED"
audit_contains '"action":"agent.lease_issued"'
audit_contains '"action":"git.receive_pack"'
audit_contains '"actorType":"agent"'

# ---------------------------------------------------------------------------
beat 9
mark step2
echo
echo "== 2. Over-reach: Agent B pushes refs/heads/main =="
LEASE_B_JSON="$(issue_lease agent-b '["refs/heads/agent/"]' 900)" || fail "lease B issuance failed"
LEASE_B="$(json_field token <<<"$LEASE_B_JSON")"
echo "   lease:   $(json_field id <<<"$LEASE_B_JSON")"
echo "   attempt: HEAD -> refs/heads/main"

if git push "$(push_url agent-b "$LEASE_B")" HEAD:refs/heads/main >"$WORK/push-b.log" 2>&1; then
    cat "$WORK/push-b.log" >&2
    fail "Agent B was allowed to push refs/heads/main; containment failed"
fi
if ! grep -q 'GH007' "$WORK/push-b.log"; then
    cat "$WORK/push-b.log" >&2
    fail "expected GH007 lease rejection for refs/heads/main"
fi
echo "   result:  REJECTED (GH007) — protected ref untouched"
audit_contains '"action":"agent.lease_rejected"'
audit_contains '"refs/heads/main"'

# ---------------------------------------------------------------------------
beat 9
mark step3
echo
echo "== 3. Expired lease: issued with a 1 second TTL =="
LEASE_X_JSON="$(issue_lease agent-expired '["refs/heads/agent/"]' 1)" || fail "expiring lease issuance failed"
LEASE_X="$(json_field token <<<"$LEASE_X_JSON")"
echo "   lease:   $(json_field id <<<"$LEASE_X_JSON")"
sleep 2
echo "   waiting: 2 seconds (lease is now expired)"

if git push "$(push_url agent-expired "$LEASE_X")" HEAD:refs/heads/agent/expired-demo >"$WORK/push-expired.log" 2>&1; then
    cat "$WORK/push-expired.log" >&2
    fail "an expired lease was accepted"
fi
if ! grep -Eqi 'authentication|401' "$WORK/push-expired.log"; then
    cat "$WORK/push-expired.log" >&2
    fail "expected an authentication rejection for the expired lease"
fi
echo "   result:  REJECTED — expired token cannot authenticate"

# ---------------------------------------------------------------------------
beat 9
mark step4
echo
echo "== 4. Revoked lease: revoke Agent A via MCP, then push =="
REVOKE_JSON="$(curl -fsS \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"lease_revoke\",\"arguments\":{\"leaseId\":\"${LEASE_A_ID}\",\"reason\":\"containment demo finished\"}}}" \
    "$BASE/mcp")" || fail "MCP lease_revoke request failed"
if ! python3 -c '
import json, sys
r = json.load(sys.stdin)
if r.get("error"):
    sys.exit("MCP error: %s" % r["error"])
if (r.get("result") or {}).get("isError"):
    sys.exit("tool error: %s" % r["result"])
' <<<"$REVOKE_JSON"; then
    fail "lease revocation failed"
fi
echo "   revoked: $LEASE_A_ID"
sleep 1

if git push "$(push_url agent-a "$LEASE_A")" HEAD:refs/heads/agent/after-revoke >"$WORK/push-revoked.log" 2>&1; then
    cat "$WORK/push-revoked.log" >&2
    fail "a revoked lease was accepted"
fi
if ! grep -Eqi 'authentication|401' "$WORK/push-revoked.log"; then
    cat "$WORK/push-revoked.log" >&2
    fail "expected an authentication rejection for the revoked lease"
fi
echo "   result:  REJECTED — revoked token cannot authenticate"
audit_contains '"action":"agent.lease_revoked"'
audit_contains '"action":"git.push_rejected"'

# ---------------------------------------------------------------------------
beat 9
mark audit
echo
echo "== 5. Durable audit trail =="
[ -s "$AUDIT_LOG" ] || fail "audit log was not written to $AUDIT_LOG"
audit_contains '"action":"agent.lease_issued"'
audit_contains '"action":"agent.lease_revoked"'
audit_contains '"action":"git.receive_pack"'
audit_contains '"action":"git.push_rejected"'
audit_contains '"action":"agent.lease_rejected"'

echo "   file: $AUDIT_LOG"
echo
python3 - "$AUDIT_LOG" <<'PY'
import json, sys

rows = []
with open(sys.argv[1], encoding="utf-8") as fh:
    for line in fh:
        line = line.strip()
        if line:
            rows.append(json.loads(line))

print(f"   {'actorType':<7} {'actor':<22} {'action':<32} resource")
print(f"   {'-'*7} {'-'*22} {'-'*32} {'-'*28}")
for ev in rows[-14:]:
    print(f"   {ev.get('actorType',''):<7} {ev.get('actor',''):<22} {ev.get('action',''):<32} {ev.get('resource','')}")
PY

beat 12
mark end
echo
echo "================================================================"
echo " PASS — containment holds and the audit trail is durable"
echo "================================================================"
echo
echo "Read the raw events:"
echo "  tail -n 20 \"$AUDIT_LOG\""
echo "Query the live admin API while the server is running (admin auth required):"
echo "  curl -s -H 'Authorization: Bearer <admin-token>' $BASE/api/v1/admin/audit-logs | python3 -m json.tool"
echo
echo "Not recorded (and not claimed): model name, prompt hash, or human approver."
echo "Leases scope a repository and ref prefixes; they do not scope individual verbs."
beat 5
