# Containment Demo

**Status:** Reproducible locally
**Owner:** Backend + DevRel
**Gate:** G1 — proof assets
**Script:** [`examples/containment/run.sh`](../../examples/containment/run.sh)

This demo answers one question on a fresh machine:

> Can an AI coding agent ship a change without holding a credential that can
> write to `main`, and can I prove what it did afterwards?

It boots a throwaway Refweir instance, issues two scoped capability leases, and
shows four outcomes: an allowed push, a blocked push to `main`, a blocked expired
lease, and a blocked revoked lease. The durable audit log records every decision.

---

## 1. What it proves

| Step | Outcome | Control |
|---|---|---|
| Agent A pushes `refs/heads/agent/feature-123` under a lease | Accepted | Lease signature, repo binding, allowed ref prefix |
| Agent B pushes `refs/heads/main` under a lease | Rejected with `GH007` | Disallowed ref prefix enforced in the pre-receive hook |
| Agent uses a lease with a 1-second TTL after it expires | Rejected at authentication | Lease expiry checked in `agentlease.Verify` |
| Agent uses a lease after `lease_revoke` | Rejected at authentication | Revocation table checked on every push |
| Operator inspects `audit.log` | All five decisions are present | Durable JSONL audit appended by the daemon and hook |

This maps to controls C2, C3, C4, C5, C6, C7, C15, C16, C17, and C18 in
[`docs/security/THREAT_MODEL.md`](../security/THREAT_MODEL.md).

---

## 2. Prerequisites

- Go 1.24+ (the module targets the toolchain in `go.mod`)
- `git`
- `curl`
- `python3`

No Docker, database, SSH daemon, or external service is required. The script
builds `refweird`, creates a temp data directory, and runs on `127.0.0.1` with a
random free port.

---

## 3. Run it

```bash
cd /path/to/refweir
examples/containment/run.sh
```

Optional:

```bash
# Fixed port
REFWEIR_DEMO_PORT=18080 examples/containment/run.sh

# Keep the temp workspace and print its path
KEEP=1 examples/containment/run.sh
```

The script exits non-zero if any containment step fails. It cleans up the
server and temp directory on exit unless `KEEP=1`.

---

## 4. Expected output

```text
================================================================
 Refweir containment demo
 workspace: /tmp/refweir-containment.XXXXXX
 server:    http://127.0.0.1:XXXXX
================================================================

[build] compiling refweird with CGO disabled...
[bootstrap] creating admin, namespace, and demo repository...
[serve] starting refweird...
[clone] acme/demo

== 1. Valid lease: Agent A pushes refs/heads/agent/feature-123 ==
   lease:   lease_...
   agent:   agent-a
   scope:   refs/heads/agent/ (main/master explicitly denied)
   expires: 2026-11-09T05:43:09Z
   result:  ACCEPTED

== 2. Over-reach: Agent B pushes refs/heads/main ==
   lease:   lease_...
   attempt: HEAD -> refs/heads/main
   result:  REJECTED (GH007) — protected ref untouched

== 3. Expired lease: issued with a 1 second TTL ==
   lease:   lease_...
   waiting: 2 seconds (lease is now expired)
   result:  REJECTED — expired token cannot authenticate

== 4. Revoked lease: revoke Agent A via MCP, then push ==
   revoked: lease_...
   result:  REJECTED — revoked token cannot authenticate

== 5. Durable audit trail ==
   file: /tmp/refweir-containment.XXXXXX/data/audit.log

   actorType actor                  action                           resource
   ------- ---------------------- -------------------------------- ----------------------------
   user    user_...               agent.lease_issued               repo_...
   agent   agent:agent-a          git.receive_pack                 acme/demo
   user    user_...               agent.lease_issued               repo_...
   agent   agent:agent-b          git.receive_pack                 acme/demo
   agent   agent-b                agent.lease_rejected             refs/heads/main
   user    user_...               agent.lease_issued               repo_...
   agent   agent:agent-expired    git.push_rejected                acme/demo
   system  mcp-tool               agent.lease_revoked              *
   agent   agent-mcp              mcp.tool_execute:lease_revoke
   agent   agent:agent-a          git.push_rejected                acme/demo

================================================================
 PASS — containment holds and the audit trail is durable
================================================================
```

Inspect the raw events while the workspace is kept:

```bash
tail -n 20 /tmp/refweir-containment.XXXXXX/data/audit.log | jq .
```

Each line is a JSON object with `id`, `timestamp`, `actor`, `actorType`,
`action`, `resource`, `severity`, `details`, and optional `metadata` (lease ID,
agent ID, ref prefixes, expiry).

---

## 5. What the audit trail records

- **Who:** the authenticated principal (`user_...` or `agent:<id>`).
- **What kind of principal:** `user`, `agent`, or `system`.
- **What happened:** `agent.lease_issued`, `git.receive_pack`,
  `agent.lease_rejected`, `git.push_rejected`, `agent.lease_revoked`.
- **Where:** repository (`acme/demo`) or ref (`refs/heads/main`).
- **Lease context:** lease ID, agent ID, allowed/disallowed ref prefixes, and
  expiry.
- **Why it was rejected:** the hook or authentication error string.

### What it does **not** record

- The model name or version.
- A prompt or prompt hash.
- A named human approver.

Those fields are not trustworthy without an external attestation source, so
Refweir does not fabricate them. They are listed as a non-goal in
[`docs/security/THREAT_MODEL.md`](../security/THREAT_MODEL.md#8-non-goals).

Also note: a lease scopes a **repository and ref prefixes**, not individual
verbs. The read-only MCP surface is a separate control.

---

## 6. Recording shot list (60–90 seconds)

Record a terminal at 1080p with a large font (16–18 px), no desktop chrome, and
captions burned in. Keep the run under 90 seconds.

| Time | Shot | Caption |
|---|---|---|
| 0–6s | Title card | “An AI agent should not hold the keys to `main`.” |
| 6–14s | `examples/containment/run.sh` starts; bootstrap log scrolls | “One binary. One temp directory. No external services.” |
| 14–24s | Step 1 output: lease ID, scope, expiry, `ACCEPTED` | “Each agent gets a signed lease: this repo, these refs, this expiry.” |
| 24–36s | Step 2: `GH007`, `REJECTED` | “The same credential cannot touch `main`. The hook stops it, not the UI.” |
| 36–46s | Step 3: expired lease, `REJECTED` | “Expiry is enforced at authentication.” |
| 46–56s | Step 4: MCP revoke + `REJECTED` | “Revocation is one tool call. The next push fails.” |
| 56–76s | Audit table scrolls; highlight `agent.lease_issued`, `agent.lease_rejected`, `git.push_rejected` | “Every decision is in a durable audit log: actor, lease, ref, reason.” |
| 76–90s | End card | “Containment you can audit. Self-host Refweir.” + `refweir.dev` |

Do not add a voiceover that claims model attribution, prompt hashing, or human
approval. The demo does not show those, and the product does not claim them.

---

## 7. Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `go build failed` | Toolchain too old or module cache unavailable | Run `go version`; `go mod download` |
| `server did not become ready` | Port conflict or data dir error | Read the printed daemon log; set `REFWEIR_DEMO_PORT` |
| `expected GH007` but the push failed differently | `git` missing from PATH in the hook environment | Ensure `git` is on PATH for the daemon’s user |
| Audit file missing events | `REFWEIR_AUDIT_LOG=off` in the environment | Unset it or point it at a writable path |
| Push hangs | Git is prompting for credentials | The script sets `GIT_TERMINAL_PROMPT=0`; check the URL credentials |

---

## 8. Cleanup

The script removes its temp workspace on exit. With `KEEP=1`:

```bash
rm -rf /tmp/refweir-containment.XXXXXX
```

If a daemon is still running from an interrupted run:

```bash
pkill -f 'refweird.*refweir-containment' || true
```
