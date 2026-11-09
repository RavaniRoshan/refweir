# Containment Threat Model

**Status:** Draft for security review
**Owner:** Security/Infra + Backend
**Gate:** G1 — proof assets
**Last updated:** 2026-11-09

This document describes what Refweir is trusted to protect, who it defends
against, which controls are actually enforced in code, and what remains the
operator's responsibility. Every control below points to a test or a demo step.
If a claim cannot be traced to code and a test, it is listed under
[residual risks](#7-residual-risks) or [non-goals](#8-non-goals).

For the marketing position and the claims we refuse to make, see
[`docs/launch/POSITIONING.md`](../launch/POSITIONING.md) and [`CLAIMS.md`](../../CLAIMS.md).

---

## 1. Scope and deployment assumptions

Refweir is a self-hosted Git forge. The threat model covers the server daemon
(`refweird`), the Git transports (Smart-HTTP on `:8080`, pure-Go SSH on `:2222`),
the MCP endpoints (`/mcp`, `/mcp/readonly`), capability leases, the pre-receive
gate, and the audit writer.

**Assumed deployment:**

- The operator owns the host, the data directory, and the process.
- TLS terminates at a reverse proxy the operator controls; the daemon itself is
  not a hardened internet edge.
- The SQLite database (`meta.db`), bare repositories, and `audit.log` live on the
  same operator-controlled filesystem, or on storage the operator trusts.
- The operator manages backups, log shipping, patching, and monitoring.
- The lease signing secret is derived from `REFWEIR_JWT_SECRET` (or a random
  process-local secret when unset), and is not shared with agents.

**Out of scope:** the agent's own model, prompt, tool implementation, host
sandbox, or network egress. Refweir constrains what an agent can do **to Git**,
not what the agent does inside its own runtime.

---

## 2. Assets

| Asset | Where it lives | Why it matters |
|---|---|---|
| Repository contents and history | Bare repos under `$REFWEIR_DATA_DIR/git/` | Source code, possibly regulated or air-gapped |
| Protected refs (`refs/heads/main`, `refs/heads/master`) | Bare repos | The line between "agent proposal" and "shipped code" |
| Capability lease token | Client-side, presented as Basic/Bearer credential | Grants temporary write scope to one agent |
| Lease signing secret | Server process (`REFWEIR_JWT_SECRET` or random) | Forging a lease would bypass ref scope |
| Long-lived credentials (PATs, API keys, deploy keys, session cookies) | SQLite, hashed | Broader blast radius than a lease |
| Audit events | `audit.log` JSONL + in-memory ring buffer + admin API | Evidence of who did what |
| Pushed secrets | Commit contents probed by the pre-receive hook | Credential leakage into history |
| Availability | Server process, SQLite writer, Git subprocesses | Agents can create push storms |

---

## 3. Trust boundaries

```text
 agent runtime (untrusted)
        │  lease token / PAT / SSH key
        ▼
 ┌──────────────────────────── refweird ────────────────────────────┐
 │ HTTP :8080  Smart-HTTP + REST + MCP + web UI                     │
 │ SSH  :2222  pure-Go Git transport                                │
 │                                                                  │
 │  authenticate ──► lease verify ──► repo access ──► exec git      │
 │                                        │                         │
 │                                        ▼                         │
 │                          pre-receive hook (secret scan,          │
 │                          ref scope, protected branches,          │
 │                          revocation/expiry env)                  │
 │                                        │                         │
 │                                        ▼                         │
 │                          bare repo + SQLite + audit.log          │
 └──────────────────────────────────────────────────────────────────┘
        ▲
        │ filesystem access (trusted operator)
        ▼
 operator backups, log shipping, monitoring
```

The important boundary is **inside** the daemon: MCP, REST, and Git transport
credentials are different surfaces, but they converge on the same store and the
same pre-receive gate. A control that exists only in the web UI is not a control.

---

## 4. Adversaries

| Adversary | Capability | Primary goal |
|---|---|---|
| **Over-permissioned agent** | Valid lease or PAT; can run Git and MCP calls | Push somewhere it should not, usually `main` |
| **Prompt-injected agent** | Same as above, but manipulated by repository content or tool output | Exfiltrate secrets, rewrite history, delete refs |
| **Compromised agent / stolen lease** | A valid, unexpired lease token | Use the lease until it expires or is revoked |
| **Malicious dependency or build script** | Code execution in CI or on a developer host | Harvest credentials, push to protected refs |
| **Insider with repository write** | Legitimate account | Bypass review, conceal changes |
| **Network attacker** | Position between agent and server | Steal credentials, tamper with transport |
| **Operator error** | Misconfiguration | Expose the daemon, disable TLS, over-broad lease |

Refweir assumes the agent runtime is **not** trustworthy. It does not assume the
operator is malicious; an operator with filesystem access can read or rewrite
anything, including the audit log.

---

## 5. Control matrix

Status legend: **Enforced** = code path plus automated test; **Partial** = code
path exists with a documented gap; **Operator** = not a product control.

| # | Control | Where enforced | Verification | Limits |
|---|---|---|---|---|
| C1 | Lease is HMAC-SHA256 signed | `pkg/agentlease/lease.go` (`Issue`, `Verify`) | `pkg/agentlease/lease_test.go` — tampered token and wrong secret rejected | Secret strength is operator-controlled |
| C2 | Lease binds agent, repo, allowed refs, disallowed refs, expiry | `pkg/agentlease/lease.go` | `pkg/agentlease/lease_test.go` (`CanPushRef`) | Scope is repo + refs, **not verbs** |
| C3 | Transport rejects a lease for another repo | `internal/gittransport/server.go` (`validateLease`) | `internal/gittransport/server_test.go` | `RepoID == "*"` is a wildcard by design |
| C4 | Transport rejects a revoked lease | `internal/gittransport/server.go`; `internal/store/agent_leases.go` | `internal/store/agent_leases_test.go`; `internal/app/agent_loop_e2e_test.go` | Revocation is checked at auth time; a push already in flight completes |
| C5 | Pre-receive rejects expired leases | `internal/security/hook.go` | `internal/security/hook_test.go` (expired case) | Hook reads expiry from env set by the transport |
| C6 | Pre-receive rejects disallowed refs (`main`, `master`) | `internal/security/hook.go` | `internal/security/hook_test.go` (lease to `main`) | Operators can configure a broader disallowed list |
| C7 | Pre-receive rejects refs outside allowed prefixes | `internal/security/hook.go` | `internal/security/hook_test.go` (`outside allowed prefixes`) | Default allow-list is `refs/heads/agent/` |
| C8 | Pre-receive rejects ref deletion under lease | `internal/security/hook.go` | `internal/security/hook_test.go` (deletion case) | Non-lease users can delete non-protected branches |
| C9 | Protected default branch deletion is rejected (`GH006`) | `internal/security/hook.go` | `internal/security/hook_test.go` (`TestPreReceiveHook_ProtectedBranchProtection`) | Branch protection is `main`/`master` by default |
| C10 | Non-fast-forward force-push to protected branch is rejected | `internal/security/hook.go` | `internal/security/hook_test.go`; trusted in `git merge-base` | Depends on `git` being present in the hook environment |
| C11 | Secret/private-key scanning blocks the push | `internal/security/scanner.go`, `hook.go` | `internal/security/scanner_test.go`, `hook_test.go` | Pattern-based; no entropy model, no binary inspection |
| C12 | Read-only MCP endpoint rejects mutating tools | `internal/mcp` (`ReadOnly` tool flag) | `internal/mcp/server_test.go` (`TestMCP_ToolsList_ReadWrite_And_ReadOnly`, `TestMCP_ToolCall_RepositoryAndReadOnlyGuard`) | `/mcp/readonly` is a separate mount; both share the store |
| C13 | Archived repositories reject writes | `internal/app`, `internal/gittransport`, `internal/sshserver` | `internal/app/repos_lifecycle_test.go` (issue create returns 403); transport checks `repo.Archived` | Archive is a soft freeze, not a backup |
| C14 | Deploy keys are scoped to one repository | `internal/sshserver`, `internal/store` | `internal/sshserver/server_test.go` (`TestSSHServerDeployKeyAuthentication`); `internal/app/deploy_keys_and_webhooks_test.go` | Read-only deploy keys abort receive-pack |
| C15 | Audit records actor, actor type, action, resource, client IP, severity, details | `internal/audit/audit.go` | `internal/audit/audit_test.go` | JSONL file + 1,000-event memory buffer |
| C16 | Lease issuance/revocation is audited | `internal/mcp/tools.go`, `internal/app/app.go`, `pkg/agentlease/handler.go` | `internal/mcp/server_test.go` (`TestMCP_LeaseTools`); `pkg/agentlease/handler_test.go` | HTTP issuance logs the authenticated user; MCP revocation logs `mcp-tool` |
| C17 | Push rejection is audited | `internal/gittransport/server.go`, `internal/security/hook.go` | `internal/security/hook_test.go` (`TestPreReceiveHook_AuditsLeaseAndProtectedBranchRejections`) | Hook audit goes to stderr when run standalone; daemon uses the file logger |
| C18 | Audit log survives restart | `cmd/refweird/main.go` wires `audit.NewFileLogger($REFWEIR_DATA_DIR/audit.log)` | `internal/audit/audit_test.go` (`TestAuditFileLoggerPersistsAndCloses`) | File is append-only by convention; no sealing or remote sink |
| C19 | Agent loop works end to end with a real MCP client | `internal/app/agent_loop_e2e_test.go` | `TestAgentLoop_EndToEnd_RealServer` | Single-node test topology |

### 5.1 What "enforced" means here

- A control is enforced only if it runs **at or below the transport** that an
  agent can reach. The pre-receive hook runs in the Git receive path, so it sees
  pushes that bypass the web UI and MCP.
- A UI-only restriction is not counted. For example, the web UI hiding a delete
  button does not protect a ref.
- The hook trusts environment variables set by the daemon. An operator or an
  attacker who can set `REFWEIR_LEASE_*` in the hook's environment can change
  the outcome. That is an operator-compromise scenario, not an agent capability.

---

## 6. Threat scenarios and expected outcomes

Each scenario maps to a demo step in
[`docs/launch/CONTAINMENT_DEMO.md`](../launch/CONTAINMENT_DEMO.md).

| Scenario | Expected outcome | Control |
|---|---|---|
| Agent with a valid lease pushes `refs/heads/agent/feature-123` | Push succeeds; audit records `git.receive_pack` with actor type `agent` and lease metadata | C2, C15, C16 |
| Agent with a valid lease pushes `refs/heads/main` | Push rejected with `GH007`; audit records `agent.lease_rejected` | C6, C17 |
| Agent with a valid lease pushes a ref outside its allowed prefixes | Push rejected with `GH007`; audit records `agent.lease_rejected` | C7, C17 |
| Agent with an expired lease pushes | Authentication or hook rejects; no ref moves | C5 |
| Agent with a revoked lease pushes | Authentication fails with 401; audit records `git.push_rejected` | C4, C17 |
| Agent pushes a commit containing an AWS key | Push rejected; audit records `security.secret_push_blocked` at `critical` | C11, C17 |
| Agent deletes `refs/heads/main` | Push rejected with `GH006`; audit records the block | C9, C17 |
| Agent force-pushes `refs/heads/main` | Push rejected with `GH006` | C10 |
| Agent calls a mutating MCP tool on `/mcp/readonly` | Tool call returns a read-only error; no mutation | C12 |
| Agent pushes to an archived repository | `403 Forbidden` before receive-pack | C13 |
| Deploy key for repo A is used against repo B | Rejected; key is bound to repo A | C14 |
| Agent calls `lease_request`, then `lease_revoke` | Both actions are audited; subsequent push fails | C16, C4 |
| Process restarts after a push | Previous audit events remain in `audit.log` | C18 |

---

## 7. Residual risks

These are real and must be stated in launch material.

1. **Operator compromise is total.** Anyone who can read `REFWEIR_JWT_SECRET`
   can forge leases. Anyone with filesystem access can edit `audit.log` and the
   SQLite database. There is no cryptographic sealing, write-once storage, or
   remote attestation.
2. **Audit durability is local.** `audit.log` is an append-only JSONL file on the
   operator's disk. It is not shipped to an immutable sink, not signed, and not
   rotated by the daemon. Use filesystem permissions, backups, and log shipping.
3. **Lease scope is coarse.** A lease scopes repository and refs, not verbs. An
   agent with a lease for `refs/heads/agent/*` can create, update, and (if the
   hook is bypassed) delete any ref under that prefix. Action scope comes from
   the MCP surface and credential type.
4. **No model, prompt hash, or approver metadata.** Leases and audit events do
   not record which model produced a change, a prompt hash, or a human approver.
   Adding those fields without a trustworthy source would be security theater.
5. **Revocation is checked at authentication.** A push that has already passed
   authentication can finish even if the lease is revoked mid-flight. Revocation
   is not a distributed lock.
6. **Secret scanning is pattern-based.** It catches common key formats and
   private-key headers. It does not catch novel formats, encoded payloads, or
   secrets split across files.
7. **Branch protection is fixed to `main`/`master`.** Per-repo configurable
   rulesets are not part of the lease gate yet. Operators with other default
   branches need a follow-up change.
8. **Concurrency is bounded, not infinite.** The transport uses a semaphore and
   SQLite has a serialized writer. At 50 concurrent pushes, measured p95 push
   latency was 40.4 s (`docs/benchmarks.md`). Size the deployment for small to
   medium teams, not a global forge.
9. **No tenant isolation in the OSS binary.** One daemon serves one trust domain.
   Multi-tenant hosted deployments need isolation above Refweir.
10. **TLS and network hardening are the operator's job.** The daemon binds to
    `127.0.0.1` by default. Exposing it publicly without TLS, a reverse proxy,
    and rate limits is a misconfiguration.

---

## 8. Non-goals

- Protecting against a compromised host, kernel, hypervisor, or operator.
- Sandboxing the agent's own process, filesystem, or network access.
- Scanning model prompts or outputs for policy violations.
- Acting as a secrets manager or a runtime credential broker.
- Guaranteeing that a merged commit is safe, correct, or reviewed by a human.
- Replacing GitHub/GitLab feature parity for every workflow.
- Immutable, tamper-evident audit storage in the OSS core.
- Per-lease verb-level scopes (read/write/merge) in the current model.

---

## 9. Verification

Run the controls locally:

```bash
# Backend security and transport tests
CGO_ENABLED=0 go test ./internal/security ./internal/gittransport ./internal/sshserver ./internal/audit ./internal/mcp ./internal/store ./internal/app

# End-to-end agent loop against a real server
CGO_ENABLED=0 go test -run TestAgentLoop_EndToEnd_RealServer -v ./internal/app

# Reproducible containment demo
examples/containment/run.sh
```

Inspect the durable audit trail after a run:

```bash
tail -f "${REFWEIR_DATA_DIR:-./data}/audit.log"
curl -s http://127.0.0.1:8080/api/v1/admin/audit-logs | jq .
```

---

## 10. Review checklist

- [ ] Security review of `pkg/agentlease`, `internal/gittransport`, and
      `internal/security/hook.go`.
- [ ] Confirm every row in the control matrix still points to a live test.
- [ ] Decide whether `audit.log` ships with rotation and a remote sink for the
      hosted tier.
- [ ] Revisit residual risk 4 (model/prompt/approver metadata) before any
      enterprise claim.
- [ ] Re-run the demo on a fresh machine and capture the output.
