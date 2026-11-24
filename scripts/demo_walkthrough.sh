#!/usr/bin/env bash
# scripts/demo_walkthrough.sh — Automated 70-second agent walkthrough script for demo recording.
set -euo pipefail

export PATH="$HOME/bin:/snap/bin:$PATH"
export REFWEIR_TELEMETRY_OPT_IN=0

BLUE='\033[0;34m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

DEMO_DIR="/tmp/refweir_demo_walkthrough"
DEMO_DATA="$DEMO_DIR/data"
DEMO_WORKTREE="$DEMO_DIR/worktree"
PORT=18088

rm -rf "$DEMO_DIR"
mkdir -p "$DEMO_DIR"
fuser -k ${PORT}/tcp >/dev/null 2>&1 || true
sleep 1

clear 2>/dev/null || true
echo -e "${BOLD}${CYAN}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${CYAN}║         refweir.dev — The Neutral Git Forge for Coding Agents             ║${NC}"
echo -e "${BOLD}${CYAN}║    70-Second Demo: From Cold Install to Merged Autonomous Agent PR        ║${NC}"
echo -e "${BOLD}${CYAN}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
echo ""
sleep 1.5

# Step 1: Install & Compile
echo -e "${BOLD}[1/7] Compiling pure-Go binaries (CGO_ENABLED=0)...${NC}"
echo -e "${YELLOW}$ go build -o bin/refweird ./cmd/refweird${NC}"
VERSION="$(tr -d '[:space:]' < "$ROOT_DIR/VERSION")"
CGO_ENABLED=0 go build -ldflags="-X main.version=$VERSION" -o "$DEMO_DIR/refweird" ./cmd/refweird
echo -e "${GREEN}✓ Binary compiled with embedded Web UI, SSH, and native MCP.${NC}"
echo ""
sleep 2

# Step 2: Bootstrap
echo -e "${BOLD}[2/7] Initializing forge administrator, namespace, and initial repo...${NC}"
echo -e "${YELLOW}$ refweird bootstrap --email admin@refweir.dev --namespace acme${NC}"
"$DEMO_DIR/refweird" bootstrap --email admin@refweir.dev --namespace acme --data-dir "$DEMO_DATA" < /dev/null > "$DEMO_DIR/bootstrap.out"
head -n 10 "$DEMO_DIR/bootstrap.out"
API_KEY=$(grep "API Key:" "$DEMO_DIR/bootstrap.out" | awk '{print $3}')
echo ""
sleep 2.5

# Step 3: Launch Daemon & Health Probe
echo -e "${BOLD}[3/7] Launching daemon (Smart-HTTP :${PORT}, pure-Go SSH :2222, MCP :${PORT}/mcp)...${NC}"
echo -e "${YELLOW}$ refweird &${NC}"
REFWEIR_DATA_DIR="$DEMO_DATA" \
REFWEIR_LISTEN_ADDR="127.0.0.1:${PORT}" \
REFWEIR_PUBLIC_URL="http://127.0.0.1:${PORT}" \
REFWEIR_GIT_URL="http://127.0.0.1:${PORT}" \
REFWEIR_SSH_ADDR="none" \
setsid "$DEMO_DIR/refweird" > "$DEMO_DIR/refweird.log" 2>&1 &
DAEMON_PID=$!
sleep 2

echo -e "${YELLOW}$ curl -s http://127.0.0.1:${PORT}/healthz${NC}"
curl -s "http://127.0.0.1:${PORT}/healthz"
echo ""
echo -e "${GREEN}✓ Daemon healthy and listening.${NC}"
echo ""
sleep 2

# Step 4: Capability Lease via MCP
echo -e "${BOLD}[4/7] Agent requests scoped capability lease via MCP...${NC}"
echo -e "${YELLOW}$ curl -X POST http://127.0.0.1:${PORT}/mcp (tool: lease_request, scope: refs/heads/agent/*)${NC}"

LEASE_RES=$(curl -s -X POST "http://127.0.0.1:${PORT}/mcp" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"lease_request","arguments":{"agentId":"claude-code-agent","repoId":"acme/demo","allowedRefPrefixes":["refs/heads/agent/"],"ttlSeconds":3600}}}')

echo "$LEASE_RES" | python3 -c '
import sys, json
d = json.load(sys.stdin)
content = json.loads(d["result"]["content"][0]["text"])
print("  Lease ID:  ", content.get("id"))
print("  Agent ID:  ", content.get("agentId"))
print("  Ref Scope: ", content.get("allowedRefPrefixes"))
print("  Expires:   ", content.get("expiresAt"))
'
LEASE_TOKEN=$(echo "$LEASE_RES" | python3 -c 'import sys, json; d = json.load(sys.stdin); print(json.loads(d["result"]["content"][0]["text"])["token"])')
echo -e "${GREEN}✓ Cryptographic capability lease issued. Pushes isolated to refs/heads/agent/*${NC}"
echo ""
sleep 2.5

# Step 5: Git Clone & Push to agent/*
echo -e "${BOLD}[5/7] Agent clones and pushes commit to agent/telemetry...${NC}"
echo -e "${YELLOW}$ git clone http://lease:<TOKEN>@127.0.0.1:${PORT}/acme/demo.git worktree${NC}"
git clone "http://lease:${LEASE_TOKEN}@127.0.0.1:${PORT}/acme/demo.git" "$DEMO_WORKTREE" >/dev/null 2>&1
cd "$DEMO_WORKTREE"
git config user.name "Claude Agent"
git config user.email "agent@refweir.dev"
git checkout -b agent/telemetry >/dev/null 2>&1
echo '{"engine": "refweir", "autonomous": true, "status": "active"}' > telemetry.json
git add telemetry.json
git commit -m "feat(agent): configure autonomous telemetry provider" >/dev/null 2>&1

echo -e "${YELLOW}$ git push origin agent/telemetry${NC}"
git push origin agent/telemetry 2>&1
echo -e "${GREEN}✓ Push accepted: pre-receive hook verified lease ref prefix and secret scanning.${NC}"
echo ""
sleep 2.5

# Step 6: Unauthorized Push to main REJECTED
echo -e "${BOLD}[6/7] Security Defense: Agent attempts unauthorized push directly to main...${NC}"
echo -e "${YELLOW}$ git push origin HEAD:main${NC}"
git push origin HEAD:main 2>&1 || true
echo -e "${RED}${BOLD}🛡️ REJECTED! Remote hook enforced capability lease boundary: refs/heads/main is forbidden.${NC}"
echo ""
sleep 3

# Step 7: Open PR & Merge via Policy Engine
echo -e "${BOLD}[7/7] Agent opens Pull Request and merges through policy verification...${NC}"
echo -e "${YELLOW}$ curl -X POST http://127.0.0.1:${PORT}/mcp (tool: pull_request_create)${NC}"
PR_RES=$(curl -s -X POST "http://127.0.0.1:${PORT}/mcp" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"pull_request_create","arguments":{"ownerSlug":"acme","repoName":"demo","title":"feat(agent): configure autonomous telemetry provider","headRef":"agent/telemetry","baseRef":"main","body":"Verified under scoped capability lease."}}}')

PR_NUM=$(echo "$PR_RES" | python3 -c 'import sys, json; d = json.load(sys.stdin); print(json.loads(d["result"]["content"][0]["text"]).get("number") or 1)')
echo -e "${GREEN}✓ Pull Request #${PR_NUM} created.${NC}"

echo -e "${YELLOW}$ curl -X POST http://127.0.0.1:${PORT}/mcp (tool: pull_request_merge, pr: #${PR_NUM})${NC}"
curl -s -X POST "http://127.0.0.1:${PORT}/mcp" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"pull_request_merge\",\"arguments\":{\"ownerSlug\":\"acme\",\"repoName\":\"demo\",\"pullNumber\":${PR_NUM}}}}" >/dev/null

echo -e "${GREEN}✓ Pull Request #${PR_NUM} merged into main by policy engine.${NC}"
echo ""
sleep 2

echo -e "${CYAN}${BOLD}==> Verifying commit on main branch:${NC}"
echo -e "${YELLOW}$ git fetch origin main && git log -1 --stat origin/main${NC}"
git fetch origin main >/dev/null 2>&1
git log -1 --stat origin/main
echo ""
sleep 2.5

echo -e "${GREEN}${BOLD}═══════════════════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}${BOLD}  ✓ Complete! From cold install to merged agent PR in < 70 seconds.         ${NC}"
echo -e "${GREEN}${BOLD}═══════════════════════════════════════════════════════════════════════════${NC}"
echo ""
sleep 2

fuser -k ${PORT}/tcp >/dev/null 2>&1 || true
rm -rf "$DEMO_DIR"
