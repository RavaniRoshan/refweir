# Positioning One-Pager

**Status:** Draft for approval
**Owner:** Founder/PM + DevRel
**Gate:** G1 — positioning approved
**Created:** 2026-11-09

This is the single source of truth for the README hero, site hero, Product Hunt
tagline/description, launch posts, and sales conversations. Every factual claim
below links to code, a test, or a reproduced measurement. If a sentence cannot be
backed, it does not ship.

---

## 1. The position

**For** platform-engineering and security teams running several AI coding agents
against private or regulated code,
**who** need agents to ship changes without handing every agent a long-lived
token that can write to `main`,
**Refweir** is a **self-hosted Git forge built for coding agents**
**that** gives each agent a short-lived, ref-scoped, revocable capability lease and
records every Git write in an append-only audit trail.

**Category:** self-hosted Git forge with agent containment at the Git boundary.

**One-liner:**

> Refweir is the self-hosted Git forge that gives every AI agent a scoped,
> expiring lease and an audit trail — so agent fleets can ship without holding
> the keys to `main`.

**Tagline options:**

- Contain every agent. Prove every merge. *(brand line)*
- Self-hosted Git forge built for coding agents. *(45/60, Product Hunt)*
- Your code. Your agents. Your refs. *(social)*

---

## 2. Unlike the alternatives

| Alternative | Where it fits | The gap Refweir targets |
|---|---|---|
| **GitHub / GitLab with agent seats** | Teams already standardised on a cloud forge | Agent identity is usually a shared app, bot, or PAT; branch protection is repo-wide, not per-agent; audit lives in a vendor cloud; self-hosted/air-gapped options are expensive or limited |
| **Forgejo / Gitea** | Teams that want a lightweight self-hosted forge | Excellent human forge, but no native MCP tool surface and no first-class per-agent lease/ref-isolation model |
| **Cursor Origin and vendor-owned agent backends** | Teams standardised on one agent vendor | Ties the agent workspace and its controls to one vendor’s accounts and infrastructure; not designed for a mixed Claude Code / Codex / Cline / Cursor fleet |
| **“Just give the agent a bot PAT”** | Small teams, low risk | A PAT is long-lived, repo-wide, and hard to attribute; revocation is all-or-nothing and the audit trail is thin |

Refweir does not claim to replace a cloud forge for every team. It targets the
case the alternatives handle worst: **many agents, private code, and a requirement
to prove what each agent did.**

---

## 3. Three pillars

### Pillar 1 — Per-agent identity and scoped capability leases

Each agent gets its own lease: a signed token with a repository scope, an allowed
ref prefix, a disallowed ref prefix, an expiry, and a revocation record.

- `pkg/agentlease/lease.go` — HMAC-SHA256 signed lease with `agent_id`, `repo_id`,
  `allowed_ref_prefixes`, `disallowed_ref_prefixes`, `issued_at`, `expires_at`.
- Transport enforcement in `internal/gittransport/server.go`; pre-receive
  enforcement in `internal/security/hook.go`.
- Revocation is stored in `revoked_agent_leases` and checked on push
  (`internal/store/agent_leases.go`, `internal/gittransport/server.go:261`).
- MCP tools `lease_request`, `lease_inspect`, `lease_revoke` let an operator or
  supervising agent issue, inspect, and revoke leases.

**Honest scope:** leases scope repository and refs, not individual verbs. Action
scope comes from the MCP surface (`/mcp` vs `/mcp/readonly`) and the credential
type, not from a per-lease action list.

### Pillar 2 — Policy and audit at the Git boundary

Controls run where the write happens, not only in a dashboard.

- **Pre-receive gate:** secret/private-key scanning, protected-branch deletion
  rejection (`GH006`), non-fast-forward protection, and lease ref checks
  (`internal/security/hook.go`).
- **Read-only MCP sandbox:** mutating tools are rejected on `/mcp/readonly`
  (`internal/mcp`, tested in `internal/mcp/server_test.go`).
- **Append-only audit:** every event records actor, actor type, action, resource,
  client IP, severity, and details; the API exposes no update/delete path
  (`internal/audit/audit.go`).
- **Archived repositories are frozen:** write operations are rejected before work
  starts.

### Pillar 3 — Vendor-neutral, self-hosted, and small enough to audit

- One pure-Go server binary (`refweird`, `CGO_ENABLED=0`) with the web UI embedded
  via `go:embed`; SQLite WAL via `modernc.org/sqlite`.
- Standard Git over Smart-HTTP and a pure-Go SSH server on `:2222`; no host
  OpenSSH dependency.
- Native MCP JSON-RPC 2.0 at `/mcp` and `/mcp/readonly`; works with Claude Code,
  Cursor, Cline, Codex, and any MCP client.
- Apache-2.0; runs on a laptop, a VPS, or an air-gapped network.

---

## 4. Proof points

Use these exact numbers or link to the source. Do not round them into a stronger
claim.

| Proof | Number | Source |
|---|---|---|
| Lease issuance, 1 agent | 4 ms p50 / 4 ms p95 | `docs/benchmarks.md` |
| Pull request open via MCP, 1 agent | 8 ms p50 / 8 ms p95 | `docs/benchmarks.md` |
| Full loop (lease → clone → push → PR), 1 agent | 2.94 s p50 / 3.10 s p95 | `docs/benchmarks.md` |
| Full loop, 5 concurrent agents | 4.07 s p50 / 4.32 s p95 | `docs/benchmarks.md` |
| Ref isolation + lease enforcement | Transport and pre-receive tests | `internal/gittransport/server_test.go`, `internal/security/hook_test.go` |
| Revocation | Store + MCP tests; push rejected after revoke | `internal/store/agent_leases_test.go`, `internal/app/agent_loop_e2e_test.go` |
| Read-only sandbox | Mutating tools rejected | `internal/mcp/server_test.go` |
| Durable audit | Lease issuance/revocation, accepted pushes, and rejected pushes are appended to `$REFWEIR_DATA_DIR/audit.log` | `internal/audit/audit_test.go`, `internal/security/hook_test.go`, `internal/mcp/server_test.go`; demo step 5 |

**Benchmark caveat:** the published numbers are a single local run on WSL2, 16 CPU
cores, Go 1.27.2, no competing workload. They are reproducible with
`go run ./bench/main.go`, not a guarantee for every host. At 50 concurrent pushes,
p95 push latency was 40.4 s — say so when discussing scale.

---

## 5. Anti-claims

We will not say these, in any asset:

1. **Not “the first Git forge for agents.”** MCP servers, bot accounts, and agent
   integrations already exist across forges. The claim is unverifiable and invites
   a correction in the launch thread.
2. **Not “GitHub is broken.”** GitHub is a good forge for many teams. The specific
   gap is per-agent containment, self-hosted audit, and vendor neutrality — not a
   general failure of GitHub.
3. **Not “MCP is the moat.”** MCP is an open protocol any competitor can implement.
   The moat is enforcement at the Git boundary plus a verifiable audit trail.
4. **Not “enterprise-ready” without qualification.** SSO/SCIM, Helm, and the
   Terraform provider exist, but the Helm chart and some subsystems are marked
   experimental. Say “enterprise integration points,” not “enterprise-certified.”
5. **Not “zero lock contention” or “infinite scale.”** SQLite uses a serialized
   writer. The honest limit is small-to-medium teams (roughly 1–25 concurrent
   agents/engineers); at 50 concurrent pushes, p95 push latency was 40 s.
6. **Not “secure by default” without the deployment caveat.** The daemon binds to
   `127.0.0.1` by default. Public exposure requires TLS, a reverse proxy, and
   operator hardening.

---

## 6. Objection handling

### “Why not just use GitHub / GitLab?”

Use them if a shared cloud forge and repo-wide branch protection are enough. Choose
Refweir when you need (a) a distinct, revocable credential per agent, (b) ref-level
isolation enforced at the Git transport and hook, (c) the audit log to stay on your
infrastructure, or (d) to run air-gapped. Refweir is not a drop-in replacement for
every GitHub feature; it is a containment layer for agent writes.

### “MCP already works with my forge. What is different?”

MCP gives an agent a tool surface. It does not, by itself, constrain which refs the
agent can write or prove who did what after the fact. Refweir pairs MCP with signed
leases, ref-prefix enforcement in the pre-receive hook, revocation, and an
append-only audit trail.

### “Self-hosting is a liability.”

It can be. Refweir reduces the operational surface: one server binary, embedded UI,
SQLite WAL, no external database, and no host OpenSSH dependency. It does not remove
the need for backups, TLS, patching, and monitoring. The Helm chart and Terraform
provider are integration points, not a managed service.

### “Another forge? I already have one.”

Refweir is not asking you to migrate every repository on day one. Run it beside an
existing forge for the agent-heavy repositories, or use it as the containment
boundary for agent branches. The value is the lease + audit model, not the file
tree.

### “Is this production-ready?”

The core agent loop, Git transports, leases, revocation, and audit are covered by
tests and a reproducible benchmark. Some enterprise surfaces (Helm, Terraform,
replication) are marked experimental. Check `CLAIMS.md` for the current status of
each claim instead of trusting a badge.

---

## 7. Message hierarchy

Use the same order everywhere:

1. **Hook:** agents are shipping code; they should not hold long-lived keys to
   `main`.
2. **Category:** a self-hosted Git forge built for coding agents.
3. **Mechanism:** scoped, expiring, revocable leases + pre-receive policy +
   append-only audit.
4. **Proof:** one server binary; MCP and standard Git; reproducible loop latency;
   tests for isolation and revocation.
5. **CTA:** run the containment demo, then self-host it.

### Asset mapping

| Asset | Lead sentence |
|---|---|
| README hero | “Self-hosted Git forge built for coding agents.” |
| Site hero | “The self-hosted forge for coding agents.” + lease/audit subline |
| Product Hunt tagline | “Self-hosted Git forge built for coding agents” (45/60) |
| PH description | Lead with containment + audit; one proof number; Apache-2.0 |
| Launch post | Start with the failure mode (agent + PAT + `main`), then the lease model |
| Sales deck | ICP, the four alternatives, the three pillars, the proof table |

---

## 8. Approval

| Field | Value |
|---|---|
| Drafted | 2026-11-09 |
| Positioning approved | _pending_ |
| Reviewed against `CLAIMS.md` | _pending_ |
| Reused in README | _pending_ |
| Reused in site hero | _pending_ |
| Reused in PH assets | _pending_ |
