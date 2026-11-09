#!/usr/bin/env bash
# Seed the hosted Refweir demo with a realistic repository, a human history, an
# agent branch pushed under a real capability lease, a pull request, issues, a
# release, and an audit trail.
#
# Usage: deploy/demo/seed.sh [BASE_URL] [API_KEY]
#   BASE_URL  default http://127.0.0.1:8088
#   API_KEY   admin API key printed by `refweird bootstrap`
#
# Requires: git, curl, python3.

set -euo pipefail

BASE_URL="${1:-http://127.0.0.1:8088}"
API_KEY="${2:-}"
if [ -z "$API_KEY" ]; then
    echo "usage: $0 [BASE_URL] [API_KEY]" >&2
    echo "       API_KEY is printed by 'refweird bootstrap'" >&2
    exit 2
fi
BASE_URL="${BASE_URL%/}"
REPO_FULL="demo/agent-lab"
OWNER="demo"
NAME="agent-lab"

for cmd in git curl python3; do
    command -v "$cmd" >/dev/null 2>&1 || {
        echo "error: '$cmd' is required" >&2
        exit 1
    }
done

WORK="$(mktemp -d "${TMPDIR:-/tmp}/refweir-demo-seed.XXXXXX")"
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

json_get() {
    python3 -c 'import json,sys
d = json.load(sys.stdin)
for key in sys.argv[1].split("."):
    d = d[int(key)] if key.isdigit() else d[key]
print(d)' "$1"
}

api() {
    curl -fsS -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" "$@"
}

repo_url() {
    python3 - "$BASE_URL" "$1" "$2" "$REPO_FULL" <<'PY'
import sys, urllib.parse
base, user, secret, full = sys.argv[1:5]
u = urllib.parse.urlsplit(base)
netloc = user + ":" + urllib.parse.quote(secret, safe="") + "@" + u.netloc
path = u.path.rstrip("/") + "/" + full + ".git"
print(urllib.parse.urlunsplit((u.scheme, netloc, path, "", "")))
PY
}

echo "==> waiting for $BASE_URL/healthz"
ready=0
for _ in $(seq 1 120); do
    if curl -fsS "$BASE_URL/healthz" >/dev/null 2>&1; then
        ready=1
        break
    fi
    sleep 0.5
done
[ "$ready" = "1" ] || { echo "error: demo server did not become healthy" >&2; exit 1; }

echo "==> removing any previous demo repositories"
delete_repo() {
    local full="$1"
    curl -fsS -X DELETE \
        -H "Authorization: Bearer $API_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"confirm_name\":\"$full\",\"sudo_secret\":\"$API_KEY\"}" \
        "$BASE_URL/api/v1/repos/$full" >/dev/null 2>&1 || true
}
delete_repo "demo/demo"
delete_repo "$REPO_FULL"

echo "==> creating $REPO_FULL"
REPO_JSON="$(api -X POST "$BASE_URL/api/v1/repos" -d "{
  \"owner\": \"$OWNER\",
  \"name\": \"$NAME\",
  \"description\": \"Demo service with a human history and an agent opening a scoped change.\",
  \"default_branch\": \"main\",
  \"visibility\": \"public\",
  \"init_readme\": true
}")"
REPO_ID="$(json_get repo.ID <<<"$REPO_JSON")"
[ -n "$REPO_ID" ] || { echo "error: could not create the demo repository" >&2; exit 1; }

CLONE_URL="$(repo_url admin "$API_KEY")"
git clone -q "$CLONE_URL" "$WORK/repo"
cd "$WORK/repo"

echo "==> writing the human history"
mkdir -p service docs .github/workflows

cat >service/README.md <<'EOF'
# Checkout service

Small demo service used to exercise agent containment. The retry budget is the
interesting part: an agent should be able to propose a change to it on an
`agent/*` branch, but never push straight to `main`.
EOF

cat >service/app.py <<'EOF'
"""Checkout service (demo)."""

from service.retry import RetryBudget


def handle(order_id: str) -> dict:
    budget = RetryBudget(max_attempts=3)
    return {"order_id": order_id, "attempts": budget.max_attempts}
EOF

cat >service/retry.py <<'EOF'
"""Retry budget for the checkout path."""

from dataclasses import dataclass


@dataclass
class RetryBudget:
    max_attempts: int = 3

    def remaining(self, used: int) -> int:
        return max(0, self.max_attempts - used)
EOF

cat >service/test_retry.py <<'EOF'
from service.retry import RetryBudget


def test_remaining_never_goes_negative():
    assert RetryBudget(max_attempts=3).remaining(10) == 0
EOF

cat >docs/architecture.md <<'EOF'
# Architecture

The service calls a payment provider and retries transient failures. Retries are
bounded by `RetryBudget` so a provider outage cannot turn into a retry storm.

## Agent rule

Agents propose changes on `refs/heads/agent/*` and open a pull request. The
pre-receive gate rejects a lease that targets `main`.
EOF

cat >.github/workflows/ci.yml <<'EOF'
name: ci
on:
  pull_request:
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: python3 -m pytest -q
EOF

cat >.gitignore <<'EOF'
__pycache__/
*.pyc
EOF

git add .
git -c user.name="Maya Chen" -c user.email="maya@example.com" \
    commit -q -m "feat: add checkout retry budget and service docs"

cat >service/metrics.py <<'EOF'
"""Minimal counters for the demo service."""

COUNTERS: dict[str, int] = {}


def increment(name: str) -> None:
    COUNTERS[name] = COUNTERS.get(name, 0) + 1
EOF
git add service/metrics.py
git -c user.name="Jonas Weber" -c user.email="jonas@example.com" \
    commit -q -m "feat: count retry attempts"

cat >docs/runbook.md <<'EOF'
# Runbook

1. Check the checkout error rate.
2. Inspect the retry counter.
3. If the provider is degraded, cap the retry budget.
EOF
git add docs/runbook.md
git -c user.name="Priya Nair" -c user.email="priya@example.com" \
    commit -q -m "docs: add on-call runbook"

git push -q origin main

echo "==> issuing a real capability lease for the demo agent"
LEASE_JSON="$(api -X POST "$BASE_URL/api/v1/agent/leases" -d "{
  \"agentId\": \"claude-code\",
  \"repoId\": \"$REPO_ID\",
  \"allowedRefPrefixes\": [\"refs/heads/agent/\"],
  \"disallowedRefPrefixes\": [\"refs/heads/main\", \"refs/heads/master\"],
  \"ttlSeconds\": 3600
}")"
LEASE_TOKEN="$(json_get token <<<"$LEASE_JSON")"
LEASE_ID="$(json_get id <<<"$LEASE_JSON")"
LEASE_EXPIRES="$(json_get expiresAt <<<"$LEASE_JSON")"
[ -n "$LEASE_TOKEN" ] || { echo "error: could not issue a demo lease" >&2; exit 1; }

echo "==> pushing the agent branch under the lease"
git checkout -q -b agent/retry-budget
cat >service/retry.py <<'EOF'
"""Retry budget for the checkout path."""

from dataclasses import dataclass

# Keep one attempt in reserve for the final settle call.
RESERVED_ATTEMPTS = 1


@dataclass
class RetryBudget:
    max_attempts: int = 3

    def remaining(self, used: int) -> int:
        return max(0, self.max_attempts - used - RESERVED_ATTEMPTS)
EOF
cat >service/test_retry.py <<'EOF'
from service.retry import RESERVED_ATTEMPTS, RetryBudget


def test_remaining_never_goes_negative():
    assert RetryBudget(max_attempts=3).remaining(10) == 0


def test_reserves_one_attempt():
    assert RetryBudget(max_attempts=3).remaining(0) == 3 - RESERVED_ATTEMPTS
EOF
git add service/retry.py service/test_retry.py
git -c user.name="Claude Agent" -c user.email="claude@agent.refweir.dev" \
    commit -q -m "fix: reserve one retry attempt for settle"

LEASE_URL="$(repo_url agent "$LEASE_TOKEN")"
git push -q "$LEASE_URL" HEAD:refs/heads/agent/retry-budget
git checkout -q main

echo "==> opening the pull request, issues, and release"
PR_JSON="$(api -X POST "$BASE_URL/api/v1/repos/$OWNER/$NAME/pulls" -d '{
  "title": "fix: reserve one retry attempt for settle",
  "head": "agent/retry-budget",
  "base": "main",
  "body": "Opened by `claude-code` after a lease-scoped push to `refs/heads/agent/retry-budget`.\n\nThe lease allows `refs/heads/agent/*` and denies `main`/`master`. The audit log records the lease, the accepted push, and this pull request."
}')"
PR_NUMBER="$(json_get pull.number <<<"$PR_JSON")"

api -X POST "$BASE_URL/api/v1/repos/$OWNER/$NAME/issues" -d '{
  "title": "Agent retry storms exhaust the review queue",
  "body": "When a provider is degraded, several agents retry at once. We need a bounded budget and a visible signal before the queue backs up.",
  "labels": ["agent", "reliability"]
}' >/dev/null

api -X POST "$BASE_URL/api/v1/repos/$OWNER/$NAME/issues" -d '{
  "title": "Export the audit log to our SIEM",
  "body": "The JSONL file is useful. A pushed export with rotation would make retention review easier.",
  "labels": ["audit", "enhancement"]
}' >/dev/null

api -X POST "$BASE_URL/api/v1/repos/$OWNER/$NAME/releases" -d '{
  "tag_name": "v0.3.0",
  "target_commitish": "main",
  "name": "v0.3.0 — bounded retries",
  "body": "Adds a retry budget and the first agent-authored pull request.",
  "draft": false,
  "prerelease": false
}' >/dev/null

cat <<EOF

================================================================
 Demo seeded
================================================================
 repository:   $REPO_FULL
 pull request: #$PR_NUMBER (agent/retry-budget)
 lease:        $LEASE_ID
 lease agent:  claude-code
 lease scope:  refs/heads/agent/ only
 lease expiry: $LEASE_EXPIRES
 audit log:    \$REFWEIR_DATA_DIR/audit.log

 Open: $BASE_URL/?tab=prs
 Try:  curl -s -X POST $BASE_URL/api/v1/demo/agent-run | python3 -m json.tool
================================================================
EOF
