# Refweir Architectural Deep Research & Comparative Audit
## Reliability, Scalability, and Security Under High-Frequency Autonomous Agent Swarms

<p align="center">
  <strong>Comprehensive System Audit, Comparative Analysis against GitHub Enterprise, and Prioritized Engineering Remediation Plan.</strong>
  <br>
  <em>Conducted under Superpowers Antigravity BCP 14 standards. Verified by Adversarial Codebase Inspection.</em>
</p>

---

## Executive Summary & High-Level Comparison Matrix

Traditional Git hosting platforms (such as GitHub Enterprise) were engineered over the past 15 years to serve human interaction patterns: infrequent commits, synchronous browser-based code reviews, human-cadence CI triggering, and OAuth/PAT-based user authentication.

**Refweir** is engineered from first principles as a dual-tenant Git forge designed for both human engineering teams and high-frequency autonomous AI coding agents (Claude Code, Cursor, Cline, Codex). However, under heavy autonomous agent swarms (e.g. 50+ concurrent agents branching, committing, pushing, requesting capability leases, and merging PRs simultaneously), the differences in transport architecture, hook execution, webhook dispatch, and database persistence determine whether the system remains resilient or experiences cascading failure.

| Architectural Domain | GitHub Enterprise / Cloud Standard | Refweir Current State (`v1.0.0-rc.1`) | Swarm Failure Mode / Risk Level | Verified Status & Severity |
| :--- | :--- | :--- | :--- | :--- |
| **Git Transport & Concurrency** | Git Protocol v2 with stateless RPC; offloads packfile generation to background workers; static CDN object storage (Project Rubicon); cluster-wide 3PC (Spokes/DGit). | CGI `git-http-backend` via `exec.CommandContext` and SSH daemon (`internal/sshserver`), throttled by a hardcoded 64-worker channel semaphore (`sem: make(chan struct{}, 64)`). | Beyond 64 concurrent pushes/clones, operations fail immediately with `503 Service Unavailable` or SSH exit status 1. | **VERIFIED (High)** |
| **Hook Execution & Push Protection** | Containerized/isolated pre-receive sandbox with fast C reachability bitmaps; commits quarantined in temporary object directories before promotion. | Spawns `refweird hook pre-receive` as an uncontained host OS subprocess executing pure-Go regex scanning across pushed commits. | P95 push latency escalates from 3.32s (1 agent) to 54.67s (50 agents), consuming 90–97% of loop time. Non-quarantined push leaves dangling objects on hook abort. | **VERIFIED (Critical)** |
| **Webhook Delivery & Backpressure** | Distributed message queue (Kafka) with Redis/Resque workers; strict 10s receiver timeout; **zero automatic retry** for standard repository webhooks to prevent retry storms; manual/API redeliveries reuse original `X-GitHub-Delivery` UUID for subscriber idempotency. | Asynchronous `go func(h *store.RepoWebhook)` spawning unbounded goroutines; retries once after 500ms; **generates a NEW GUID on retry**, breaking subscriber deduplication; synchronous 4KB response write to SQLite. | Swarm events cause goroutine spikes; downstream services receiving retries process duplicate events due to altered GUID; SQLite WAL write contention. | **VERIFIED (High)** |
| **Authentication & Capability Leases** | Asymmetric Ed25519/RSA token verification; granular fine-grained PAT scopes; instant token revocation propagation; trusted reverse-proxy CIDR parsing for client IP. | Ephemeral HMAC-SHA256 capability leases (`pkg/agentlease`); `POST /api/v1/agent/leases/{id}/revoke` lacks authentication; `POST /api/v1/agent/leases` bootstrap check allows unauthenticated lease creation when `CountUsers == 0`. | Anonymous attackers can unilaterally revoke any agent lease (DoS); attackers can forge leases during fresh installation; client IP spoofing via leftmost `X-Forwarded-For`. | **VERIFIED (Critical)** |
| **Ruleset & Merge Engine** | GitHub Rulesets v2 with Speculative Merge Queues; evaluates branch protection, signed commits, code owners, and linear history at merge time; optimistic concurrency with queue rebase. | CAS reference checking (`expectedHeadSha` and `CheckAndSetReference`); `engine.EvaluateMerge` checks only approvals and status checks; AutoMerge locks per-PR rather than per-target branch. | Bypasses linear history and commit signature enforcement at merge time; concurrent PR merges race against base branch CAS, causing the loser to fail without queuing. | **VERIFIED (Medium)** |
| **Database & Disaster Recovery** | Multi-region active-passive replication with Raft/consensus; point-in-time recovery via WAL streaming and online database backups. | Single-node SQLite in WAL mode; 30s `PRAGMA wal_checkpoint(TRUNCATE)` blocks writers; `CreateSnapshot` performs raw `os.Open`/`io.Copy` on live `meta.db`. | Checkpoint lock convoys during burst agent writes; raw file copy during active writes risks torn pages and corrupted backup snapshots. | **VERIFIED (High)** |
| **Agent Containment & MCP Surface** | Strict outbound SSRF egress proxies blocking RFC 1918/RFC 3927; least-privilege token-bound scopes; bounded TTLs. | `security.ValidateWebhookURL` only string-checks `169.254.169.254`; does NOT block RFC 1918 subnets (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`) or loopback; `/mcp/readonly` is route-only; secret scanner misses OpenAI/Anthropic keys. | Agents can trigger webhooks to probe internal VPC infrastructure (SSRF); PATs used for `/mcp/readonly` can be retargeted to `/mcp` for mutating operations. | **VERIFIED (Critical)** |

---

## 1. Git Transport, Hook Lifecycle & Data Integrity under Concurrent Scale

### 1.1 Concurrency Throttle & Capacity Cliff
In Refweir, Git Smart-HTTP (`internal/gittransport/server.go`) and embedded SSH (`internal/sshserver/server.go`) gate concurrent transfer operations using an in-memory buffered channel semaphore:
```go
// internal/gittransport/server.go lines 54-56
sem: make(chan struct{}, 64)
```
When more than 64 concurrent pushes or fetches arrive from an agent swarm, `select` with `default` triggers immediately:
```go
// internal/gittransport/server.go lines 382-388
select {
case s.sem <- struct{}{}:
    defer func() { <-s.sem }()
default:
    http.Error(w, "server busy: git concurrency limit reached", http.StatusServiceUnavailable)
    return
}
```
**Impact:** A swarm of 70 agents attempting coordinated branch sync experiences an immediate HTTP 503 error storm without queueing or graceful degradation.

### 1.2 Subprocess Pre-Receive Hook Latency Escalation
Git Push operations (`git-receive-pack`) execute `internal/gitstore/gitstore.go:EnsureHooks`, which writes an executable shell script that spawns `refweird hook pre-receive` as a new operating system subprocess. 

According to verified performance baseline metrics (`plans/2026-11-09-refweir-performance-baseline.md` lines 120–135):
- At **1 agent**, Git push p50 is **3.32s** (total loop 3.42s).
- At **20 agents**, Git push p50 is **21.53s** (total loop 21.91s).
- At **50 agents**, Git push p50 is **54.67s** (total loop 56.09s).

Git push overhead accounts for **90%–97%** of the entire agent execution time. Spawning 50 concurrent OS processes to execute pure-Go regex pattern scanning on unpacked commit objects overwhelms host CPU context switching.

### 1.3 TOCTOU Race Condition in REST API Commits
In `internal/origin/handlers_commits.go` (and `internal/app`), `CreateCommitFromFiles` checks the current branch head against `req.ExpectedHeadSHA`:
```go
if headSHA != req.ExpectedHeadSHA {
    http.Error(w, "head sha mismatch", http.StatusConflict)
    return
}
```
However, after this check, the handler executes secret scanning and then calls `gitstore.WriteFiles`, which internally resolves the reference again and commits via `repo.Storer.SetReference` **without acquiring `s.RepoLock(namespace, repo)`**.
Two concurrent agents issuing REST file commits simultaneously can pass the `headSHA` check concurrently and then overwrite each other's commits in an uncoordinated race.

---

## 2. Webhook Reliability, Delivery Semantics & Backpressure

### 2.1 The Idempotency & GUID Regeneration Flaw
In `internal/app/webhook_dispatch.go`:
```go
// DispatchRepoWebhook generates a new delivery GUID:
guid := idgen.New("guid")
...
reqHeaders := http.Header{
    "X-GitHub-Delivery":  []string{guid},
    "X-Refweir-Delivery": []string{guid},
}
```
When a webhook delivery fails or returns a `5xx` error, Refweir executes an automatic retry:
```go
// internal/app/webhook_dispatch.go lines 150-155
time.Sleep(500 * time.Millisecond)
_, _ = DispatchRepoWebhook(dispatchCtx, db, h, event, payload)
```
Because `DispatchRepoWebhook` is called again from scratch, **it generates a completely new GUID**.
In contrast, GitHub Enterprise and standard enterprise webhook systems (e.g. Stripe, Hookdeck) strictly guarantee that delivery retries retain the *identical delivery GUID* so that downstream subscribers can safely deduplicate deliveries in their database idempotency tables. In Refweir, the downstream agent or webhook receiver sees two different GUIDs and processes the payload twice.

### 2.2 Unbounded Goroutine Dispatch
`DispatchRepoEvent` iterates over all matching webhooks and immediately spawns a bare goroutine:
```go
for _, hook := range hooks {
    ...
    go func(h *store.RepoWebhook) { ... }(hook)
}
```
Under an active swarm emitting 100 push events across 20 webhooks, 2,000 unbounded goroutines are spawned concurrently. Each goroutine attempts an HTTP connection and executes a synchronous SQL write to `repo_webhook_deliveries` (`db.RecordRepoWebhookDelivery`), saturating SQLite's WAL single-writer queue.

---

## 3. Authentication, Capability Leases & Credential Boundaries

### 3.1 Unauthenticated Capability Lease Revocation
In `internal/app/app.go`:
```go
mux.HandleFunc("POST /api/v1/agent/leases/{id}/revoke", func(w http.ResponseWriter, r *http.Request) {
    ...
    leaseID := r.PathValue("id")
    actor := "anonymous"
    actorType := "system"
    if u, err := auth.ResolveUser(r.Context(), db, r); err == nil && u != nil {
        actor = u.ID
        actorType = "user"
    }

    if err := db.RevokeAgentLease(r.Context(), leaseID, "", "", reason); err != nil {
        http.Error(w, "Failed to revoke lease: "+err.Error(), http.StatusInternalServerError)
        return
    }
})
```
**Vulnerability:** The endpoint explicitly permits `actor := "anonymous"`. If a request is received without an `Authorization` header, `auth.ResolveUser` fails silently, but the function proceeds to revoke the lease! Any unauthenticated network client can iterate through lease IDs and unilaterally revoke all active agent capability leases, causing an immediate denial of service across the agent swarm.

### 3.2 Bootstrap Lease Creation Vulnerability
In `internal/app/app.go`:
```go
mux.HandleFunc("POST /api/v1/agent/leases", func(w http.ResponseWriter, r *http.Request) {
    count, _ := db.CountUsers(r.Context())
    if count > 0 {
        if _, err := auth.ResolveUser(r.Context(), db, r); err != nil {
            http.Error(w, "Unauthorized: Authentication required to issue agent capability leases", http.StatusUnauthorized)
            return
        }
    }
    baseLeaseHandler(w, r)
})
```
During initial deployment bootstrap before the first admin user completes registration (`count == 0`), any unauthenticated caller can request capability leases with arbitrary ref permissions (`allowed_ref_prefixes: ["refs/heads/main"]`).

### 3.3 Symmetric HMAC-SHA256 vs. Asymmetric Signing
Refweir capability leases (`pkg/agentlease/lease.go`) use symmetric HMAC-SHA256:
```go
sig := sign(secret, payloadB64)
token := fmt.Sprintf("%s.%s", payloadB64, sig)
```
Because verification requires the secret key, any microservice, hook, or verification worker that validates a lease must possess the identical HMAC secret. If that secret is compromised or distributed across worker nodes, any compromised worker can forge arbitrary capability leases for any repository. GitHub and modern distributed forges use asymmetric public-key cryptography (Ed25519) where only the core authority holds the private signing key, while edge workers verify using a public key.

---

## 4. Ruleset Evaluation, Atomic CAS References & Auto-Merge Concurrency

### 4.1 Missing Invariants in Ruleset Merge Evaluation
In `internal/rulesets/engine.go`, `engine.EvaluateMerge` only evaluates two rule types:
1. `pull_request` (approvals, dismissal)
2. `require_status_checks` (required CI checks)

It does **not** evaluate:
- `require_linear_history`
- `require_signed_commits`
- `require_code_owner_review`
- `restricted_push_actors`

Even if these rules are configured in the repository ruleset, a PR merge executed via REST API bypasses them entirely, in contrast to GitHub Rulesets v2 which rigorously evaluates all branch invariants at merge evaluation time.

### 4.2 Auto-Merge Base Ref CAS Contention
In `internal/rulesets/automerge.go`, the auto-merge engine acquires an in-memory lock keyed by `pr.ID` (`prLocks`).
However, when multiple PRs targeting the same base branch (`refs/heads/main`) are queued for auto-merge, they execute concurrent merges against the Git repository. The first PR updates `refs/heads/main` using CAS (`repo.Storer.CheckAndSetReference`). The second PR's CAS fails because the base commit has moved.
Unlike GitHub's Merge Queue (which automatically rebases or speculatively groups pending PRs), Refweir marks the second PR as failed and abandons auto-merge without retry.

---

## 5. Storage Scalability, SQLite WAL Contention & Backup Integrity

### 5.1 Checkpoint Convoys & Single-Writer Contention
Refweir configures SQLite in WAL mode with a 5-second busy timeout (`_busy_timeout=5000`).
Every 30 seconds, `replication/controller.go` triggers:
```go
_, err := c.db.SQL().ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
```
Under SQLite WAL architecture, `TRUNCATE` checkpoints block all new incoming write transactions while waiting for active readers to finish. During a concurrent swarm burst where 50 agents are writing audit logs, lease records, check runs, and webhook deliveries, this periodic truncate causes lock convoys that exhaust the 5-second busy timeout, surfacing as `SQLITE_BUSY` errors.

### 5.2 Dangerous Raw File Copy During Live Snapshot
In `internal/replication/controller.go`:
```go
dbPath := filepath.Join(c.cfg.DataDir, "meta.db")
srcFile, err := os.Open(dbPath)
...
_, err = io.Copy(destFile, srcFile)
```
Performing an `os.Open` and `io.Copy` on a live, actively written SQLite database file is a known anti-pattern. While WAL mode keeps main changes in `-wal`, dirty pages in the main database file or concurrent checkpoint activity can cause `io.Copy` to read torn 4KB pages, producing corrupted snapshot backup archives.
SQLite provides the `VACUUM INTO 'filepath'` command (or the online backup API) which safely captures a consistent point-in-time snapshot without risking torn pages.

---

## 6. Agent Containment, Attack Surface & MCP Security

### 6.1 Server-Side Request Forgery (SSRF) in Webhooks
In `internal/security/url_validation.go`:
```go
// Only blocks literal 169.254.169.254 and link-local IP
lowerHost := strings.ToLower(hostname)
if lowerHost == "169.254.169.254" || lowerHost == "metadata.google.internal" || lowerHost == "instance-data" {
    return ErrBlockedMetadataHost
}
if ip := net.ParseIP(hostname); ip != nil {
    if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
        return ErrBlockedMetadataHost
    }
}
```
**Vulnerability:** `ValidateWebhookURL` does NOT resolve hostnames to IP addresses via DNS at dial-time, and does NOT block:
- RFC 1918 Private Networks: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`
- Loopback addresses: `127.0.0.0/8`, `::1`
- Carrier-grade NAT: `100.64.0.0/10`
- DNS rebinding attacks: A domain resolving to an external IP during validation, then resolving to `127.0.0.1` during HTTP request dispatch.

An autonomous agent with repository admin permissions can register a webhook pointing to `http://127.0.0.1:8080/api/...` or internal Kubernetes cluster IPs (`http://10.96.0.1`) and use Refweir's webhook dispatcher as an SSRF proxy to exfiltrate internal cluster state.

### 6.2 Secret Scanner Coverage Gaps
In `internal/security/scanner.go`, `NewScanner()` includes patterns for AWS, Slack, Private Keys, GitHub tokens, and generic high-entropy strings, but is missing critical modern AI and cloud credentials:
- OpenAI Project & User Keys (`sk-proj-[A-Za-z0-9_-]{48,}`, `sk-[A-Za-z0-9]{48}`)
- Anthropic API Keys (`sk-ant-[A-Za-z0-9_-]{80,}`)
- GitLab Personal Access Tokens (`glpat-[0-9a-zA-Z_-]{20}`)
- Stripe Secret Keys (`sk_live_[0-9a-zA-Z]{24}`)

### 6.3 MCP Readonly Sandbox Scope Enforcement
The Model Context Protocol (MCP) server routes at `/mcp` and `/mcp/readonly` both authenticate using `auth.ResolveUser`. 
The read-only restriction is bound solely to the HTTP route URL. If an agent is issued a PAT intended for read-only MCP access, the token itself has no route restriction; the agent can simply change its MCP client URL from `/mcp/readonly` to `/mcp` and gain full write tool access (creating branches, pushing commits, deleting repos).

---

## 7. Prioritized Engineering Remediation Roadmap

All remediations will be executed strictly under the **Superpowers Antigravity 5-Stage TDD Workflow**:
1. Write failing unit/integration tests reproducing the vulnerability or bottleneck.
2. Implement minimal, clean, pure-Go / TypeScript changes.
3. Verify 100% green tests with concrete terminal evidence.
4. Adversarially verify before committing.

### Priority 0 (Immediate Security & Data Integrity - Iterations 47 & 48)
1. **SSRF Webhook Hardening (`internal/security/url_validation.go`):**
   - Resolve DNS at dial time.
   - Block RFC 1918 (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), loopback (`127.0.0.0/8`, `::1`), link-local (`169.254.0.0/16`), and IPv6 private/multicast ranges.
   - Configure custom `http.Transport` dialer with IP validation to prevent DNS rebinding.
   - Disallow HTTP redirects or validate redirect targets against SSRF rules.
2. **Capability Lease Revocation Authentication (`internal/app/app.go`):**
   - Require valid user authentication (`auth.ResolveUser`) or active lease token before permitting lease revocation.
   - Verify that the calling user has admin/write permissions on the lease's target repository.
3. **Webhook Delivery GUID Idempotency (`internal/app/webhook_dispatch.go`):**
   - Preserve the original `delivery.GUID` across retries so subscribers can deduplicate events.
   - Add exponential backoff with full jitter (e.g. 500ms * 2^attempt + jitter) instead of a fixed 500ms sleep.
4. **SQLite Snapshot Safety (`internal/replication/controller.go`):**
   - Replace raw `os.Open`/`io.Copy` with SQLite `VACUUM INTO 'path'` to guarantee atomic, consistent, torn-page-free backups.
5. **Modern Secret Scanner Patterns (`internal/security/scanner.go`):**
   - Add regex patterns for OpenAI (`sk-proj-`, `sk-`), Anthropic (`sk-ant-`), GitLab (`glpat-`), and Stripe (`sk_live_`).

### Priority 1 (Reliability & Scalability - Iterations 49 & 50)
1. **Bounded Webhook Worker Pool:** Replace unbounded `go func` with a bounded worker pool (e.g. 16 workers, channel buffer of 1024) with backpressure metrics.
2. **REST Commit RepoLock CAS:** Acquire `s.RepoLock(namespace, repo).Lock()` in `CreateCommitFromFiles` to guarantee single-writer atomicity across REST API file commits.
3. **Ruleset Linear History & Signature Enforcement:** Wire `engine.EvaluateMerge` to check linear history and signature validity before approving PR merges.
4. **Auto-Merge CAS Retry Loop:** Add optimistic concurrency retry loop (up to 3 attempts with rebase/refresh) in `AutoMergeEngine` when base branch CAS fails due to concurrent merges.

### Priority 2 (Long-Term Architecture - Iterations 51+)
1. **Asymmetric Capability Leases:** Transition `pkg/agentlease` to Ed25519 signing pairs so verification nodes only need public keys.
2. **In-Process Git Hook Execution:** Embed secret scanning and branch policy directly into the transport handler to eliminate process-spawn overhead during Git pushes.
3. **SQLite Incremental Vacuum:** Enable `PRAGMA auto_vacuum = INCREMENTAL` and schedule batch `incremental_vacuum` to mitigate database bloat from high lease churn.
