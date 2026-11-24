#!/usr/bin/env bash
# Seeds a running instance with an open agent pull request for gallery capture.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORK="${1:-/tmp/refweir-capture}"
PORT="${2:-18080}"

rm -rf "$WORK"
mkdir -p "$WORK/data" "$WORK/bin"

export REFWEIR_TELEMETRY_OPT_IN=0

VERSION="$(tr -d '[:space:]' < "$ROOT_DIR/VERSION")"
(cd "$ROOT_DIR" && CGO_ENABLED=0 go build -ldflags="-X main.version=$VERSION" -o "$WORK/bin/refweird" "$ROOT_DIR/cmd/refweird")
"$WORK/bin/refweird" bootstrap --email admin@corp.refweir.dev --namespace acme --data-dir "$WORK/data" < /dev/null > "$WORK/bootstrap.out" 2>&1
API_KEY="$(grep 'API Key:' "$WORK/bootstrap.out" | awk '{print $3}')"
if [ -z "$API_KEY" ]; then
    echo "failed to parse the bootstrap API key" >&2
    cat "$WORK/bootstrap.out" >&2
    exit 1
fi
printf '%s' "$API_KEY" > "$WORK/key"
printf '%s' "$PORT" > "$WORK/port"

REFWEIR_DATA_DIR="$WORK/data" \
REFWEIR_LISTEN_ADDR="127.0.0.1:$PORT" \
REFWEIR_PUBLIC_URL="http://127.0.0.1:$PORT" \
REFWEIR_GIT_URL="http://127.0.0.1:$PORT" \
REFWEIR_SSH_ADDR="none" \
setsid "$WORK/bin/refweird" > "$WORK/daemon.log" 2>&1 &
echo $! > "$WORK/daemon.pid"

for _ in $(seq 1 60); do
    if curl -sf "http://127.0.0.1:$PORT/healthz" >/dev/null; then break; fi
    sleep 0.25
done

# 1. Issue first active capability lease for agent
LEASE_RESPONSE="$(curl -sS -X POST "http://127.0.0.1:$PORT/mcp" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lease_request","arguments":{"agentId":"forge-capture","repoId":"*","allowedRefPrefixes":["refs/heads/agent/"],"ttlSeconds":7200}}}')"
LEASE_TOKEN="$(printf '%s' "$LEASE_RESPONSE" | python3 -c 'import json,sys; d=json.load(sys.stdin); print(json.loads(d["result"]["content"][0]["text"])["token"])')"

# 2. Clone using the capability lease token
git clone "http://lease:${LEASE_TOKEN}@127.0.0.1:${PORT}/acme/demo.git" "$WORK/worktree" >/dev/null 2>&1
cd "$WORK/worktree"
git config user.name "Forge Capture"
git config user.email "capture@refweir.dev"
git checkout -b agent/forge-demo >/dev/null 2>&1
printf '\n## Captured with the Forge UI\n\nThis change was pushed by an agent under a capability lease.\n' >> README.md
git commit -am "feat: add Forge capture note" >/dev/null 2>&1
git push origin agent/forge-demo >/dev/null 2>&1

# 3. Create agent pull request #1
curl -sS -X POST "http://127.0.0.1:$PORT/mcp" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"pull_request_create","arguments":{"ownerSlug":"acme","repoName":"demo","title":"Add Forge capture note","headRef":"agent/forge-demo","baseRef":"main","body":"Captured from a live instance."}}}' > "$WORK/pr.json"

# 4. Attach passing agent review and checks to PR #1
curl -sS -X POST "http://127.0.0.1:$PORT/api/v1/repos/acme/demo/pulls/1/reviews/agent" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" >/dev/null 2>&1 || true

# 5. Add a verified thread comment on PR #1
curl -sS -X POST "http://127.0.0.1:$PORT/mcp" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"thread_comment","arguments":{"ownerSlug":"acme","repoName":"demo","pullNumber":1,"body":"Automated verification completed: all policy checks passing."}}}' >/dev/null 2>&1 || true

# 6. Issue a second active lease for a concurrent agent
curl -sS -X POST "http://127.0.0.1:$PORT/mcp" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"lease_request","arguments":{"agentId":"claude-code-worker","repoId":"acme/demo","allowedRefPrefixes":["refs/heads/agent/"],"ttlSeconds":3600}}}' >/dev/null 2>&1 || true

# 7. Create branch via MCP for additional audit event
curl -sS -X POST "http://127.0.0.1:$PORT/mcp" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"branch_create","arguments":{"ownerSlug":"acme","repoName":"demo","branchName":"agent/perf-tune","fromRef":"main"}}}' >/dev/null 2>&1 || true

echo "seeded instance on http://127.0.0.1:${PORT} (data: $WORK/data)"
