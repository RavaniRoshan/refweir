# Refweir Product Readiness: Comprehensive Multi-Perspective Internal Audit
**Audit Date:** 2026-11-23  
**Audited Repositories:**  
- `refweir` (Forge daemon `refweird`, Web UI `web/`, CLI `refweir`, docs `site/`, deploy `deploy/demo/`)  
- `refweir-site` (Production marketing landing page & visual story)  
**Methodology:** 6-Angle Deep Technical Codebase Audit & Static Analysis Verification

---

## Executive Summary & Readiness Verdict

**Verdict:** **NOT READY FOR PRODUCTION WITHOUT CRITICAL REMEDIATION (Remediation In Progress)**

While Refweir delivers exceptionally deep, idiomatic architecture across pure-Go Git transport, SQLite-backed WAL concurrency, AST agent code reviewing, and a rich GitHub-inspired React UI, our comprehensive multi-perspective audit uncovered **6 critical security vulnerabilities**, **3 silent data/consistency bugs**, and multiple **production deployment/marketing discrepancies** that would fail in a hardened enterprise or public launch environment.

---

## Audit Angle 1: Git Engine & Core Forge Transport

### Verified Strengths
1. **Pure-Go Git Transport & Packet-Line Framing:** Zero CGO dependencies. `internal/gittransport` implements compliant packet-line encoding with strict 65,520-byte payload limits, handling flush (`0000`), delimiter (`0001`), and response-end (`0002`) frames.
2. **Concurrent Request Protection:** Active semaphore limiting (`s.sem = make(chan struct{}, 64)`) protects the host from Git subprocess fork-bombing.
3. **Repository Path Escapes Blocked:** Clean path validation prevents path traversal (`..` and backslashes) into host directories.

### Critical Vulnerabilities & Defects
1. **Critical SSH Certificate Signature Verification Bypass (`internal/sshserver/server.go:94-115`):**
   - *Defect:* Refweir matched only `ssh.FingerprintSHA256(cert.SignatureKey)` against SQLite without verifying the cryptographic signature (`cert.SignatureKey.Verify(...)` or `ssh.CertChecker.CheckCert(...)`), and completely ignored validity timestamps (`cert.ValidBefore` / `cert.ValidAfter`).
   - *Impact:* An attacker knowing an enrolled CA's public key could forge an unsigned certificate and gain authenticated access as any user.
2. **SSH Connection & Process Leaks (`internal/sshserver/server.go:213, 355, 383-390`):**
   - *Defect:* Disconnected TCP sessions left `git-upload-pack` and `git-receive-pack` subprocesses running indefinitely due to lack of `context.Context` cancellation on `exec.Command`. Leaked subprocesses permanently held slots in `s.sem`.
3. **Missing Git LFS `/locks/verify` Endpoint (`internal/gitlfs/server.go:107-108, 296-382`):**
   - *Defect:* Git LFS specification mandates `POST .../info/lfs/locks/verify` prior to pushing locked files. Refweir returned HTTP 404. Furthermore, `/locks/{id}/unlock` allowed any authenticated caller to delete another user's locks without `force: true`.
4. **Git Smart-HTTP Swallowed Stderr & Premature 200 OK (`internal/gittransport/server.go:396-437`):**
   - *Defect:* If `git-http-backend` crashed before writing CGI headers, Refweir returned HTTP 200 with an empty body, masking errors.
5. **Git Merge Non-Atomic Ref Overwrite (`internal/gitstore/merge.go:225, 281-285`):**
   - *Defect:* `SetReference` unconditionally overwrote branch heads without Compare-And-Swap (CAS) validation against `baseSHA`. Also, SHA-256 tree hashes (64 hex characters) failed the hardcoded `len == 40` check.

---

## Audit Angle 2: Security, Authentication, & Leases

### Critical Vulnerabilities & Defects
1. **Unverified App Token JWTs (`internal/auth/db_resolver.go:58`, `middleware.go:150`):**
   - *Defect:* `ParseAppToken` only unmarshaled JSON claims without verifying the cryptographic signature. `db_resolver.go` admitted any token whose subject started with `app_`.
   - *Exploit:* Unauthenticated callers could forge `{"sub": "app_admin", "scope": "*"}` and execute API calls as installed applications.
2. **Blanket Agent Bypass in Access Grants (`internal/store/grants.go:224-226`):**
   - *Defect:* `CheckRepoAccess` contained:
     ```go
     if strings.HasPrefix(userID, "agent:") {
         return true, nil
     }
     ```
     This granted any agent user unconditional read/write access to **all** repositories regardless of organization membership or repository visibility.
3. **Ref Path Traversal Vulnerability in Leases (`pkg/agentlease/lease.go:101-120`, `internal/security/hook.go:82-110`):**
   - *Defect:* Evaluated ref permissions via naive `strings.HasPrefix`. A ref like `refs/heads/agent/../../heads/main` matched `refs/heads/agent/`, bypassing branch isolation.
4. **Deprovisioned/Suspended User Git Bypass (`internal/gittransport/server.go:250-256`):**
   - *Defect:* While web sessions checked `user.Status == "suspended"`, HTTP Git basic authentication and personal access token verification did not verify account status.
5. **Expired SCIM Token Bypass (`internal/store/federation.go:113-122`):**
   - *Defect:* `ValidateSCIMToken` checked only `token_hash = ?` and ignored `expires_at`.

---

## Audit Angle 3: API & MCP Agent Ergonomics

### Defects & Discrepancies
1. **Silent Failures in MCP Merge Enforcement (`internal/mcp/tools.go:584-589, 1510-1515`):**
   - *Defect:* Errors from `ListRulesets`, `ListReviews`, and `ListCheckRunsForCommit` were silently discarded with `_`. Transient store errors bypassed branch protection rules.
2. **Stubbed `sync_mirror` Tool (`internal/mcp/tools.go:341-361`):**
   - *Defect:* Returned hardcoded `synced: true` without performing background fetch or sync.
3. **Forge REST API Pagination Omission (`internal/app/app.go`):**
   - *Defect:* No Forge REST list endpoint honored pagination tokens (`pageSize`, `pageToken`, `next_page_token`), hardcoding limits to 50 or 100.
4. **Missing Forge Endpoints:**
   - Missing `GET /api/v1/repos/{owner}/{name}/issues/{number}`.
   - Missing `DELETE /api/v1/repos/{owner}/{name}/branches/{branch}`.
   - Fake mock asset injected into releases with SHA-256 of empty string.
5. **Empty Input Validation Flaws:**
   - PR title updates accepted empty strings.
   - Branch creation endpoints omitted git reference traversal validation.

---

## Audit Angle 4: Web UI / UX & Frontend Robustness

### Defects & Polish Needs
1. **Unhandled Errors & Stuck Loading States:**
   - `PullRequestsView.tsx:66-113`: Ignored TanStack Query `isError`, locking in empty state on query failure.
   - `IssuesView.tsx:93-151`: `handleCreateNewIssue` and comments lacked `try/catch/finally`, locking buttons in disabled "Submitting..." state on API failure.
   - `UserProfileView.tsx:172-214`: Unhandled Promise rejections on profile and key fetches.
2. **Deceptive Client-Side Fallback in SSH Key Modal (`UserProfileView.tsx:273-285`):**
   - *Defect:* When adding an SSH key failed on the server, the client injected a synthetic key with a fake SHA-256 fingerprint into the UI and closed the modal, misleading the user.
3. **Contrast Ratio Failures in Dark Mode:**
   - White text on bright brass `--warn` (`ChangeVisibilityModal.tsx`, `TransferRepoModal.tsx`) produced ~1.6:1 contrast (WCAG AA failure).
   - White text on bright cyan `--info` (`MergeBox.tsx`) produced ~1.7:1 contrast.
4. **Theme Synchronization Root Defect:**
   - Dark mode class toggling was implemented solely inside `AppShell.tsx`. In standard view (`/`), system dark mode preferences were not applied to `document.documentElement`.

---

## Audit Angle 5: Production Deployment & Observability

### Defects & Hardening Gaps
1. **Uncalled HTTP Metrics Counter (`internal/metrics/metrics.go:86-88`, `internal/app/app.go`):**
   - *Defect:* `refweir_http_requests_total` was defined and exported on `/metrics`, but `IncHTTPRequests()` was never called in the request pipeline; metrics reported `0` in production.
2. **Static `/healthz` Probe (`internal/app/app.go:182-190`):**
   - *Defect:* `/healthz` hardcoded `200 OK` without pinging SQLite (`db.Ping()`) or checking disk write capability. Deadlocks or corrupted databases were masked from container orchestrators.
3. **Rate Limiting Spoofing via `X-Forwarded-For` (`internal/app/demo.go:470-481`):**
   - *Defect:* `clientKey(r)` took `strings.Split(forwarded, ",")[0]`. Clients could supply spoofed headers to bypass demo rate limits.
4. **Systemd Service Resource Limits Missing (`deploy/demo/systemd/refweir-demo.service`):**
   - Missing `MemoryMax=`, `CPUQuota=`, and `TasksMax=`, leaving the host vulnerable to resource exhaustion.

---

## Audit Angle 6: Marketing Site, SEO, & Conversion Integrity

### Defects & Claims Alignments
1. **Broken Doc Link:**
   - `refweir-site/src/sections/Problem.tsx:256`: Linked to non-existent `/docs/explanation/capability-leases`.
2. **Third-Party CDN Unsplash Images (`refweir-site/src/sections/Benefits.tsx`, `Problem.tsx`):**
   - External dependencies leaked visitor metadata and caused image loading failures in air-gapped environments.
3. **Missing SEO Structured Data (JSON-LD):**
   - Zero Schema.org schemas (`SoftwareApplication`, `FAQPage`) present on `index.html`.
4. **Technical Claim Discrepancies:**
   - `Gallery.tsx:278` displayed building `./cmd/refweir` instead of daemon `./cmd/refweird`.
   - `Policy.tsx:354` presented `refweir.toml`, whereas Refweir is environment-variable configured.
   - `workflow-demos.tsx:183` cited `POST /api/agentlease` instead of `POST /api/v1/agent/leases`.
5. **Mobile Nav Crowding:**
   - Small screens (< 360px) wrapped top nav items due to unhidden desktop links.

---

## Remediation Roadmap & Execution Plan

All 6 angles are mapped into discrete iterations in `LOOP_GOALS.md`:
- **Iteration 13:** Critical SSH Certificate Signature Verification & Process/Goroutine Safety.
- **Iteration 14:** Unverified App Token JWT Exploit Fix & Token Resolver Hardening.
- **Iteration 15:** Blanket Agent Permission Removal & Agent Lease Ref Path Traversal Prevention.
- **Iteration 16:** Git LFS `/locks/verify` Implementation & Lock Ownership Enforcement.
- **Iteration 17:** HTTP Metrics Middleware Wiring (`refweir_http_requests_total`) & Live Database Health Probe (`/healthz`).
- **Iteration 18:** Marketing Site Link Correction, Image Localization & Dual Theme Color / SEO Schema Enhancement.
- **Iteration 19:** Web UI Contrast Token & Error Boundary Hardening.
