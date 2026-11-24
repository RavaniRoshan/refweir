# Deep Research Report: Enterprise Launch Readiness & Git Forge Parity

**Date:** 2026-11-24  
**Topic:** Architectural specifications, security baselines, developer ergonomics, and self-hosted reliability for Refweir  
**Status:** Synthesized & Adversarially Verified

---

## 1. Executive Summary

To prepare Refweir for enterprise self-hosting and general availability, this deep research study investigated six critical dimensions: Git forge parity, enterprise identity/RBAC, auditability/SIEM streaming, agent observability and safeguards, high availability/disaster recovery, and developer ergonomics. 

Following multi-agent exploration and adversarial verification, key technical corrections were identified:
1. **Three-Dot Diffing**: PR diffs must calculate diffs relative to the merge base (`git diff $(git merge-base base head) head`). Refweir's current `internal/gitstore/merge.go` performs a naive 2-dot diff, which incorrectly shows concurrent commits merged into `base` as deletions by the PR author.
2. **Commit Signature Verification**: Refweir supports OpenSSH SIGSPEC and GPG/PGP signatures; claims of S/MIME support were refuted by inspection of `internal/gitstore/signature.go`.
3. **Pre-Receive Hook Scanning**: Hook revision range scanning must account for new branch pushes (`$oldrev = 0000000000000000000000000000000000000000`) and use `git rev-list $newrev --not --branches --tags` rather than naive diff-tree ranges.
4. **Developer Ergonomics & Anti-Slop**: Top-tier developer tools (Linear, Raycast, GitHub) are characterized by keyboard shortcut discoverability (`?` modal with focus trap), WCAG 2.1.4 single-key toggles, flattened virtualized code trees, synchronized split-diff scrolling, and Roving Tabindex tables with calibrated functional tokens.

---

## 2. Key Findings by Domain

### 2.1 Git Forge Parity (Branch Rulesets & Diffing)
- **Three-Dot Diffing vs Two-Dot Diffing**: Pull request file views must display changes introduced by the feature branch since it diverged from the base branch. In `refweir/internal/gitstore/merge.go`, `baseTree.Diff(headTree)` must be upgraded to find the common ancestor commit tree `mergeBaseTree.Diff(headTree)`.
- **Mergeability Calculation**: `git merge-tree --write-tree` performs in-memory 3-way merges and outputs a Tree OID or conflict details. Forges evaluate this asynchronously and cache `mergeable_state` (`clean`, `dirty`, `blocked`, `behind`).
- **Linear History & Strict Ancestry**: Enforcing linear history prevents multi-parent merge commits on trunk branches, preserving `git bisect` bisectability. When strict ancestry is required, feature branches must integrate target `HEAD` before merging.
- **Merge Queues**: Speculatively tests ordered batches of PRs against cumulative merge heads to prevent "rebase races" in high-concurrency teams and multi-agent swarms.

### 2.2 Enterprise Identity & RBAC
- **OIDC Discovery & PKCE**: Clients must fetch `/.well-known/openid-configuration` (OpenID Connect Discovery 1.0) and enforce Authorization Code Flow with PKCE (`S256`), CSRF `state`, and ID token `nonce` validation.
- **JWKS Key Caching**: Signature verification (RS256/ES256) requires local public key caching with dynamic cache busting on unseen `kid` identifiers.
- **SCIM 2.0 Directory Syncing**: Microsoft Entra ID and Okta mandate `/Users` and `/Groups` endpoints with `PATCH` support and soft deletion (`"active": false`). Searches with zero results must return HTTP 200 with `totalResults: 0`.
- **Role Tier Realities**: While GitHub defines 5 tiers (Admin, Maintain, Write, Triage, Read) and GitLab defines 5 different tiers, lightweight self-hosted Go engines (and Refweir's current schema) operate on a clean 3-tier model (`admin`, `member`, `service`), sufficient for secure repository collaboration.

### 2.3 Auditability & SIEM Integration
- **Compliance Mandates**: SOC 2 (CC6.1, CC6.8, CC7.2) and ISO 27001:2022 (A.8.15) mandate control outcomes (monitoring, non-repudiation, retention), not named algorithms.
- **Schema Standards (OCSF / ECS)**: Audit records must capture ISO 8601 UTC timestamps, Actor identity, Action verb (e.g., `repo.create`, `branch_rule.override`), Target entity, Network context (IP, User-Agent, Trace ID), and Outcome.
- **Tamper-Evident Hash Chaining**: Monotonic sequence numbers and `prevHash` chaining (`hash = SHA-256(payload || prevHash || seq)`). To defend against local appliance compromises, head hashes must be anchored via signed checkpoints or external WORM / S3 Object Lock storage.
- **Real-Time Streaming**: Modern SIEMs ingest via OpenTelemetry Logs (OTLP) and signed HTTPS Webhook streams (with `X-Hub-Signature-256` HMAC signatures and retry DLQs). Legacy CEF and RFC 5424 Syslog remain supported through collectors.

### 2.4 Agentic Observability & Safeguards
- **OpenTelemetry Semantic Conventions**: Agent tool calls should be instrumented with `gen_ai.operation.name = "execute_tool"`, `gen_ai.tool.name`, and `gen_ai.client.token.usage` metrics. W3C `traceparent` headers link agent reasoning to tool execution.
- **Git Lease Lifecycle**: Concurrency monitoring requires tracking `lock_acquisition_failures_total`, `resource_lock_hold_duration_seconds`, and lease expiration/revocation rates.
- **Pre-Receive Secret Scanning**: Hooks must inspect commit payloads within 5-second execution budgets. Must skip binary files and large blobs (>1MB), verify commit ancestry safely, and defer active network secret verification to async pipelines.

### 2.5 Storage, Backup & Disaster Recovery
- **Dual-Storage Challenge**: Git forges maintain relational metadata in SQLite and Git objects in bare repositories on disk. Backups must coordinate both layers to avoid orphaned refs.
- **SQLite WAL & Online Backups**: `VACUUM INTO 'target.db'` executes as a non-blocking read transaction in WAL mode, creating a clean defragmented snapshot without write lock contention.
- **Litestream WAL Streaming**: Continuous incremental WAL frame streaming to S3 provides RPO < 2s for single-writer setups. Bare Git repositories must be snapshotted or backed up concurrently, accompanied by post-restore reconciliation tools.
- **Git Repository Integrity**: `transfer.fsckObjects` on ingest, fast `git verify-pack` for storage checks, and scheduled background `git fsck --connectivity-only` ensure long-term packfile health.

### 2.6 Developer Ergonomics & Anti-Slop Interface Patterns
- **Keyboard Shortcut Discoverability (`?`)**: A dedicated modal overlay displaying all available shortcuts, implemented with `aria-modal="true"`, focus trapping, `Esc` restoration, and WCAG 2.1.4 single-key disable toggles.
- **Vim Navigation**: `j`/`k` list cursor movement, `x` multi-select, and `Enter` selection with keyboard boundary protection that prevents triggering inside text inputs.
- **Virtualized Code & Blame Navigation**: Monorepo scale demands flattened 1D virtualized arrays (`@tanstack/react-virtual`) rather than deeply nested recursive DOM elements.
- **High-Density Table Design**: Compact 28–36px row heights, Roving Tabindex (single tab stop with 2D arrow key navigation), and semantic contrast tokens adhering strictly to dark/light contrast baselines.

---

## 3. Adversarial Resolutions & Technical Constraints

| Item | Researcher Finding | Adversarial Resolution / Correction |
|:---|:---|:---|
| **Diffing Mechanism** | Cited `casrai.org` for 3-dot diffs | Rejected citation (hallucination). Identified that Refweir's `merge.go` currently uses 2-dot diffing (`baseTree.Diff(headTree)`) and must be upgraded to 3-dot merge-base diffing. |
| **Commit Signatures** | Claimed Refweir supports S/MIME | Refuted. `signature.go` strictly supports OpenSSH and GPG; S/MIME returns unsupported format. |
| **Pre-Receive Ranges** | Proposed `git diff-tree $oldrev..$newrev` | Refuted. Crashes on new branches (`$oldrev = 0000...`). Correct approach: check for zero hash and use `git rev-list $newrev --not --branches --tags`. |
| **Litestream Format** | Claimed Litestream streams LTX files | Corrected. LTX belongs to LiteFS. Litestream streams generational LZ4-compressed WAL frames. |
| **Audit Standards** | Promoted 2013 CADF and legacy CEF | Flagged as outdated. Modern enterprise deployments standardize on OCSF, OTLP Logs, and signed JSON Webhook streams. |
| **RBAC Roles** | Asserted 5-tier role model as required | Nuanced. GitHub uses 5 tiers, GitLab uses 5 different tiers, and Refweir successfully implements a 3-tier model (`admin`, `member`, `service`). |

---

## 4. Full Source List
- OpenID Connect Core 1.0 & Discovery 1.0: https://openid.net/specs/
- IETF RFC 7636 (PKCE), RFC 7643 & 7644 (SCIM 2.0), RFC 5424 (Syslog): https://datatracker.ietf.org/
- OpenTelemetry GenAI Semantic Conventions: https://github.com/open-telemetry/semantic-conventions-genai
- SQLite WAL Mode & VACUUM INTO: https://www.sqlite.org/wal.html, https://www.sqlite.org/lang_vacuum.html
- Git Documentation (`git merge-tree`, `git fsck`, `git verify-pack`): https://git-scm.com/docs
- GitHub Enterprise Rulesets & Merge Queue: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/
- GitHub Primer & WCAG 2.1.4 Accessibility: https://primer.style/, https://github.blog/open-source/accessibility/
- Litestream Architecture & Guides: https://litestream.io/how-it-works/
