# Refweir Performance Program Scorecard & Final Decision Record

**Date:** 2026-11-09  
**Execution Plan:** `plans/2026-11-09-plan-01-refweir-performance-execution.md`  
**Host Environment:** Linux RoshansdaamPC 6.6.87.2-microsoft-standard-WSL2 (AMD Ryzen 7 4800H, 16 vCPUs, 7.5 GiB RAM)  
**Status:** All Tasks (1–9) Complete — 100% Green Gates  

---

## 1. Before vs After Program Scorecard

| Area / Workload | Pre-Change Baseline | Post-Optimization State | Measured Impact / Delta | Status |
| :--- | :--- | :--- | :--- | :--- |
| **SQLite Disk Write Transactions** | 2,725,942 ns (~2.73 ms) | 101,205 ns (~101.2 µs) | **26.9× faster throughput** (~366 tx/s → ~9,880 tx/s) | **ACCEPTED** (Task 7) |
| **SQLite Index Read Lookups** | 14,912 ns (~14.9 µs) | 14,909 ns (~14.9 µs) | **~67,100 queries/s** sustained | **ACCEPTED** (Task 7) |
| **`InlineDiffViewer` Line Render** | $O(N \times M)$ filter loop on every diff line | $O(1)$ indexed lookup via `useMemo` map | Eliminated redundant comment array allocations | **ACCEPTED** (Task 3) |
| **SPA Route Preloading (Preparatory)** | Default eager/on-click | TanStack Router configured with `intent` preloading + 30s stale time | Preparatory routing foundation for app shell migration; unmounted in current `<App />` tab view | **PREPARATORY / NO LIVE RUNTIME IMPACT** (Task 3) |
| **Frontend Unit Test Suite (`web`)** | 44 passed, 1 failed (stray `EOF` defect) | **45/45 passed, 169/169 tests passed** | Zero test failures, 100% green | **ACCEPTED** (Task 2 & 3) |
| **Frontend Production Build (`web`)** | Failed (TypeScript syntax errors) | **Clean build (1.26s Vite transform)** | Production bundles emitted to `internal/webui/dist` | **ACCEPTED** (Task 2 & 3) |
| **Static Docs Site Verify (`site`)** | 14.00s wall-clock, 649 links verified | **14.00s wall-clock, 649 links verified** | 28 static pages, 0 broken links | **ACCEPTED** (Task 4) |
| **Package Manager (pnpm vs npm)** | npm ci: 19.92s warm reinstall | pnpm: 1.81s warm, but **broke 100% of Vitest tests** | Retained npm; rejected pnpm due to test failure | **REJECTED** (Task 5) |
| **Git Engine (Native Git vs go-git)** | Pure-Go `go-git/v6` | Upstream PR #2351 revlist walk confirmed (11.2ms for 256 refs) | Retained pure-Go; rejected host Git dependency | **REJECTED** (Task 6) |
| **Break-Glass Database / Queues** | SQLite WAL | No capacity gate breached; 9.8k tx/s capacity | Retained SQLite WAL; rejected Postgres/Redis | **NO-GO** (Task 8) |

---

## 2. Accepted Changes

1. **SQLite WAL Pragma Optimization (`internal/store/store.go`):**
   - Configured `synchronous(NORMAL)` and `cache_size(-16000)` on on-disk SQLite WAL databases.
   - Reduced write transaction latency from ~2.73 ms to 101 µs (26.9× speedup, ~9,880 transactions/sec capacity).
   - Enforced by regression test `internal/store/store_test.go:TestDiskPragmas`.
2. **`InlineDiffViewer` Comment Indexing (`web/src/components/views/InlineDiffViewer.tsx`):**
   - Pre-indexes PR review comments into a composite `path:line:side` map using `React.useMemo`.
   - Replaces $O(N \times M)$ linear filtering on every rendered diff hunk line with $O(1)$ constant-time lookup.
   - Enforced by test in `web/src/components/views/__tests__/InlineDiffViewer.test.tsx`.
3. **TanStack Router Intent Preloading (`web/src/router.ts`):**
   - Configured `defaultPreload: 'intent'` and `defaultPreloadStaleTime: 30_000` to establish the intent-based preloading baseline for the ongoing app shell migration.
   - Note: The running product currently renders `<App />` with internal tab state; this configuration becomes active on client routes once `<RouterProvider>` is mounted.
4. **Correction of Build & Test Syntax Artifacts:**
   - Removed stray `EOF` strings from `web/src/lib/events.ts`, `web/src/lib/queryClient.ts`, `web/src/lib/store.ts`, and `web/src/app/AppShell.tsx`.
   - Restored Vitest passing count from 44 to 45 suites (169/169 tests) and restored `npm run build` from failing to 100% green.
5. **Comprehensive Repeatable Performance Benchmarks Added:**
   - `internal/store/store_bench_test.go`: Measures database open, read queries, and write transactions.
   - `internal/gitstore/gitstore_bench_test.go`: Measures commit log walks and bare repository initialization.
   - `internal/gittransport/pack_bench_test.go`: Measures revlist multi-ref pack object traversals (1 vs 256 refs).

---

## 3. Rejected Changes (With Reasons)

1. **pnpm Migration (Rejected in Task 5):**
   - *Reason:* While warm installs were 10× faster, pnpm's strict symlinked layout broke 100% of the project's Vitest test suites (`ReferenceError: document is not defined` across all 45 test files). Fixing it would require altering the dependency graph with undeclared packaging dependencies or disabling isolated linking. `npm` with `npm ci` was retained.
2. **Native Git Subprocess / Host Git Executable (Rejected in Task 6):**
   - *Reason:* Benchmark evidence confirmed `go-git/v6` already incorporates upstream PR #2351 and traverses 256 wanted refs across 300 commits in only 11.2 ms with 1.08 MB allocations. Adding native Git would violate the project's core invariant of hermetic pure-Go execution (`CGO_ENABLED=0`, zero host OpenSSH/Git dependencies).
3. **Break-Glass External Services (Rejected in Task 8):**
   - *Reason:* No capacity gates failed. SQLite WAL with tuned pragmas sustains ~9,880 write transactions/sec and ~67,100 index lookups/sec. Adding external PostgreSQL or Redis would introduce operational burden, network round-trips, and cluster failure modes without measurable necessity.

---

## 4. Deferred Opportunities

1. **Vitest JSDOM Worker Pooling:**
   - Vitest logs identify that JSDOM environment creation accounts for ~35% of total test runtime (~110s across 45 workers). Migrating to `pool: 'vmThreads'` or worker reuse (`isolate: false`) can accelerate developer feedback once upstream React 19 testing library isolation is verified.
2. **Git Smart-HTTP Push Pipeline Pre-Receive Scanning:**
   - In `bench/main.go`, Git push operations dominate 90–97% of agent loop time due to the external subprocess execution of `refweird hook pre-receive` and secret pattern scanning. In-process hook execution or pre-compiled regex evaluation can be explored in a future security/transport optimization cycle.

---

## 5. Final Correctness Gate Verification Evidence

All commands run fresh on the final candidate:

- **Go Backend Test Suite:**  
  `go test ./...` → **PASS** (33 packages, 0 failures).
- **Go Zero-CGO Build:**  
  `CGO_ENABLED=0 go build -v ./cmd/refweir ./cmd/refweird` → **PASS** (exit code 0).
- **Web Frontend Test Suite:**  
  `cd web && npm test` → **PASS** (45/45 test files, 169/169 tests passed).
- **Web Frontend Production Build:**  
  `cd web && npm run build` → **PASS** (Vite production bundle built cleanly).
- **Site Static Documentation & Marketing:**  
  `cd site && npm run verify` → **PASS** (TypeScript clean, Next.js build 28 pages, 45 HTML pages / 649 links verified, 9 claims verified).
