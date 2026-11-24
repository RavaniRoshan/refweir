# Architectural Decision Record (ADR): Break-Glass Scale Evaluation

**Date:** 2026-11-09  
**Status:** Decided — NO-GO (Single-Binary Default Preserved)  
**Context:** Plan 01 Task 8 — Break-Glass Architecture Evaluation  

---

## 1. Context & Gate Condition

Under Plan 01, break-glass architecture changes (such as introducing external PostgreSQL, Redis, distributed worker queues, or external Git services like Gitaly) are conditional:
> "Trigger this task only when a representative deployment repeatedly misses its agreed user or capacity target after local query, code, and workload tuning."

## 2. Empirical Capacity Evidence

1. **SQLite Metadata Store:**
   - Write throughput with WAL + `synchronous=NORMAL` achieved **~9,880 write transactions/sec** (`101 µs/op`).
   - Read index lookup latency achieved **~67,100 queries/sec** (`14.9 µs/op`).
   - No lock timeouts, pool stalls, or `SQLITE_BUSY` errors were observed under concurrent agent load.
2. **Git Smart-HTTP & go-git Engine:**
   - Multi-ref pack negotiation (`revlist.Objects`) across 256 wanted refs executed in **11.2 ms** with 1.08 MB allocations.
   - All Git Smart-HTTP and SSH operations executed with zero host OpenSSH or Git dependencies.
3. **Agent Concurrent Load:**
   - 50 concurrent coding agents executing complete loops (Lease → Clone → Push → PR) completed with **0% error rate**.
4. **Static Documentation & UI:**
   - Documentation static build compiles 28 pages in 1.6s and verifies 649 links cleanly.
   - Frontend initial critical asset payload is ~217 KB gzipped.

## 3. Options Evaluated

| Option | Trigger Metric Required | Operational Cost | Latency / Security Impact | Verdict |
| :--- | :--- | :--- | :--- | :--- |
| **PostgreSQL Migration** | Writer saturation > 10k tx/s or multi-host write requirement | High (requires separate DB cluster, credentials, backups, network hops) | +1–3 ms network RTT per query; loses single-binary simplicity | **REJECTED** (Not triggered; SQLite handles 9.8k tx/s) |
| **Redis Cache Layer** | Auth / lease lookup latency > 10ms | Medium (requires Redis daemon, cache invalidation protocol) | Stale permissions risk; cache coherency overhead | **REJECTED** (In-memory DB queries take 14.9 µs; Redis adds network latency) |
| **External Git Daemon (Gitaly / native Git)** | `go-git` pack generation OOM or > 5s per clone | High (host Git dependencies, process sandboxing, IPC protocol) | Violates hermetic zero-host-dependency invariant | **REJECTED** (go-git handles 256 refs in 11.2ms) |
| **Asynchronous Worker Queue** | PR webhook/check execution blocking HTTP requests | Medium (requires durable background queue) | Adds worker orchestration complexity | **DEFERRED** (Existing Go goroutine worker pools handle current workloads) |

## 4. Decision

**NO-GO on break-glass architecture changes.**  
Refweir preserves its foundational architecture:
- Hermetic single-binary forge daemon (`refweird`) with embedded pure-Go SSH (:2222) and Smart-HTTP (:8080).
- Local SQLite database in WAL mode with tuned `synchronous=NORMAL` pragmas.
- Embedded React SPA for authenticated users, and static Next.js/Fumadocs site for public marketing and docs.
