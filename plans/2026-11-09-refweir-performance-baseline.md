# Refweir Performance Baseline & Inventory

**Date:** 2026-11-09  
**Host Environment:** Linux RoshansdaamPC 6.6.87.2-microsoft-standard-WSL2 (AMD Ryzen 7 4800H, 16 vCPUs, 7.5 GiB RAM)  
**Status:** Task 1 & Task 2 Baseline Complete  

---

## 1. Version and Toolchain Inventory

| Component | Declared / Manifest Version | Installed Runtime Version | Lockfile / Spec |
| :--- | :--- | :--- | :--- |
| **Go** | `1.26.0` (toolchain `go1.27.3`) | `go version go1.27.3 linux/amd64` | `go.mod`, `go.sum` |
| **go-git** | `github.com/go-git/go-git/v6 v6.0.0-beta.1` | `v6.0.0-beta.1` | `go.mod` |
| **SQLite Driver** | `modernc.org/sqlite v1.60.1` | `v1.60.1` (pure Go, CGO=0) | `go.mod` |
| **Node.js** | Engine unspecified in package.json | `v24.21.0` | CI uses Node 22 (`ci.yml`) / 24 (`pages.yml`) |
| **Package Manager** | npm (unpinned) | `npm v11.19.0` | `lockfileVersion: 3` (`web` and `site`) |
| **React (Web App)** | `^19.2.8` | `19.2.8` | `web/package-lock.json` |
| **Vite** | `^8.3.0` | `8.3.0` (Rolldown/Oxc engine) | `web/package-lock.json` |
| **Vitest** | `^5.0.4` | `5.0.4` | `web/package-lock.json` |
| **TanStack Router** | `^1.170.46` | `1.170.46` | `web/package-lock.json` |
| **TanStack Query** | `^5.107.1` | `5.107.1` | `web/package-lock.json` |
| **Zustand** | `^5.1.0` | `5.1.0` | `web/package-lock.json` |
| **React (Site)** | `^19.3.1` | `19.3.1` | `site/package-lock.json` |
| **Next.js** | `^16.4.4` | `16.4.4` (Turbopack static export) | `site/package-lock.json` |
| **Fumadocs** | `core: 16.17.10`, `ui: 16.17.10`, `mdx: 15.4.10` | `16.17.10` / `15.4.10` | `site/package-lock.json` |
| **Tailwind CSS** | `^4.3.3` (web), `^4.3.4` (site) | `4.3.3` / `4.3.4` | `web` & `site` lockfiles |

---

## 2. CI Workflows & Deployment Configuration

- **CI Provider:** GitHub Actions (`.github/workflows/ci.yml`, `.github/workflows/pages.yml`, `.github/workflows/release.yml`).
- **Runner OS:** `ubuntu-latest` (GitHub-hosted standard runner, 2-core x86_64, 7 GB RAM).
- **Workflow Stages (`ci.yml`):**
  1. `backend`: Set up Go (`actions/setup-go@v5`, `cache: true`), set up Node (`actions/setup-node@v4`, `cache: npm`, path `web/package-lock.json`), `cd web && npm ci && npm run build`, install staticcheck/govulncheck/gosec/gitleaks, `CGO_ENABLED=0 go build -v ./cmd/...`, `go vet`, `gofmt`, verify capability counts (`internal/tools/gencounts`), run analyzers, `go test ./internal/auth -run TestSSO -v`, installer checksum check, `go test -v -race ./...`.
  2. `frontend`: Node 22 setup, `npm ci`, `npm test` (Vitest), `npm run build`.
  3. `docs`: Node 22 setup, `npm ci`, `npm run verify` (`typecheck && next build && check-links && claims-audit`), check published links, verify quickstart consistency.
- **Workflow Stages (`pages.yml`):**
  - Node 24 setup (`actions/setup-node@v7`), `npm ci`, `npm run typecheck`, `npm run build` (Next.js static export to `out/`), `npm run check-links`, `npm run claims-audit`, upload artifact, deploy to GitHub Pages environment (`ravaniroshan.github.io/refweir/`).
- **Cache Strategy:**
  - `actions/setup-go@v5` module cache (`go.sum`).
  - `actions/setup-node@v4` with `cache: npm` on `web/package-lock.json` and `site/package-lock.json`.
  - Cache hit rate: Not directly measurable offline; standard GitHub Actions cache hit rate typically 80–90% on main branch pushes.

---

## 3. Journey Source, Test, and Workflow Mapping

| Journey | Product Surface | Primary Source Files | Primary Test Files | CI/Workflow Gate |
| :--- | :--- | :--- | :--- | :--- |
| **1. Repo Overview → File** | Authenticated Web UI | `web/src/components/views/RepoOverview.tsx`, `web/src/components/views/CommitsView.tsx`, `web/src/routes/index.tsx`, `web/src/lib/api.ts` | `web/src/components/views/__tests__/RepoOverview.test.tsx` | `ci.yml` (`frontend`) |
| **2. Repo → PR → Changed Files** | Authenticated Web UI | `web/src/components/views/PullRequestsView.tsx`, `web/src/components/views/InlineDiffViewer.tsx`, `web/src/components/modals/CreatePullRequestModal.tsx` | `web/src/components/views/__tests__/PullRequestsView.test.tsx`, `web/src/components/views/__tests__/InlineDiffViewer.test.tsx` | `ci.yml` (`frontend`) |
| **3. Settings Read/Write** | Web UI & Go Store | `web/src/components/views/RepoSettingsView.tsx`, `internal/app/app.go`, `internal/store/repos.go` | `web/src/components/views/__tests__/RepoSettingsView.test.tsx`, `internal/app/repos_lifecycle_test.go`, `internal/store/repos_test.go` | `ci.yml` (`backend` & `frontend`) |
| **4. Agent / MCP Request** | MCP JSON-RPC Server & Activity SSE | `internal/mcp/server.go`, `internal/mcp/tools.go`, `internal/agentactivity/stream.go`, `internal/agentactivity/hub.go`, `pkg/agentlease/lease.go` | `internal/mcp/server_test.go`, `internal/agentactivity/hub_test.go`, `internal/app/agent_loop_e2e_test.go`, `pkg/agentlease/lease_test.go` | `ci.yml` (`backend`) |
| **5. Git Smart-HTTP & SSH** | Git Transport & Engine | `internal/gittransport/server.go`, `internal/gittransport/pktline.go`, `internal/sshserver/server.go`, `internal/gitstore/gitstore.go` | `internal/gittransport/server_test.go`, `internal/sshserver/sshserver_test.go`, `internal/gitstore/gitstore_test.go` | `ci.yml` (`backend`) |
| **6. Documentation & Marketing** | Static Next.js Site | `site/app/page.tsx`, `site/app/docs/[[...slug]]/page.tsx`, `site/scripts/check-links.mjs`, `site/scripts/audit-claims.mjs` | `site/scripts/check-links.mjs`, `site/scripts/audit-claims.mjs` | `pages.yml`, `ci.yml` (`docs`) |

---

## 4. Representative Named Git Workloads and Fixtures

1. **Fixture A: Small Repository (`fixture-small`)**
   - 1 branch (`main`), 10 commits, 5 files (< 50 KB total tree size).
2. **Fixture B: Large History (`fixture-large-history`)**
   - 1 branch, 1,000 commits with sequential linear history, 100 KB tree size.
3. **Fixture C: Many Refs (`fixture-many-refs`)**
   - 256 branches/tags sharing commit objects (tested against `upload-pack` advertise-refs negotiation).
4. **Fixture D: Large Diff / Large Objects (`fixture-large-diff`)**
   - Single commit modifying 100 files with 5,000+ line diffs and 2 MB binary object.

---

## 5. Deployment Shapes and Expected Concurrency

- **Topology Default:**
  - Single-binary deployment (`refweird`), hosting HTTP `:8080` (Smart-HTTP + REST API + MCP) and SSH `:2222`.
  - Storage: Local filesystem directory for bare Git repos (`$REFWEIR_DATA_DIR/git/`) and SQLite database in WAL mode (`$REFWEIR_DATA_DIR/refweir.db`).
  - Zero external database (no PostgreSQL, no Redis required).
- **Enterprise Helm Shape (`deploy/helm/refweir-enterprise`):**
  - Kubernetes StatefulSet or single-replica Deployment (`replicaCount: 1`).
  - PersistentVolumeClaim for `/var/lib/refweir` (single-node lock constraint for SQLite WAL).
- **Agent Concurrency Profiles:**
  - Evaluated via `bench/main.go` across Tiers 1, 5, 20, 50 concurrent agents.

---

## 6. Pre-Change Correctness Gates Status

Executed on 2026-11-09:

1. **Go Backend Validation (`go test ./...`):**
   - Result: **PASS** (33 packages tested, 0 failures, total wall time 20.03s, peak RSS 719 MB).
2. **Site Documentation & Marketing Validation (`cd site && npm run verify`):**
   - Result: **PASS** (`typecheck` clean, Next.js 16.4.4 static build 28 pages generated in 1.59s, 45 pages / 649 links verified, 9 claims / 19 forbidden phrases verified). Wall-clock: 14.00s, peak RSS 320 MB.
3. **Frontend Vitest Suites (`cd web && npm test`):**
   - Initial Result: 44 passed, 1 failed due to stray `EOF` from commit `ce90841`.
   - After cleaning stray `EOF` syntax artifacts: **PASS** (45 files passed, 168 tests passed, duration 23.64s).
4. **Frontend Production Build (`cd web && npm run build`):**
   - Initial Result: Failed on `EOF` TS errors.
   - After cleaning stray `EOF` syntax artifacts: **PASS** (10.44s wall-clock, peak RSS 887 MB).

---

## 7. Repeatable Performance Baselines (Task 2 Data)

### 7.1 Build and Test Clocks (Clean vs Warm)

| Stage | Clean Wall-Clock | Warm Wall-Clock | Clean Peak RSS | Warm Peak RSS | Notes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Go Build (`cmd/refweir` + `cmd/refweird`)** | 30.28s | 1.34s | 799.5 MB | 387.5 MB | Full rebuild vs Go incremental build cache |
| **Go Test Suite (`go test ./...`)** | 45.10s | 20.03s | 740 MB | 719.4 MB | 33 packages, includes crypto & git tests |
| **Web Build (`tsc -b && vite build`)** | 12.10s | 10.44s | 890 MB | 887.2 MB | Emits to `internal/webui/dist/` |
| **Web Vitest (`npm test`)** | 25.50s | 23.64s | 680 MB | 660 MB | 45 test files, 168 tests; 109s total jsdom time (35% jsdom overhead) |
| **Site Verify (`npm run verify`)** | 15.20s | 14.00s | 335 MB | 320.7 MB | `tsc --noEmit` + Next.js build + check-links + claims-audit |

### 7.2 Autonomous Agent Concurrent Loop Benchmarks (`bench/main.go`)

Workload: Autonomous coding agent loop (Capability Lease Request via MCP → Git Clone via Smart-HTTP → Git Push with Lease Auth → Pull Request Open via REST API). Tested against live embedded server over loopback:

| Concurrency Tier | Lease p50 / p95 / p99 | Git Clone p50 / p95 / p99 | Git Push p50 / p95 / p99 | PR Open p50 / p95 / p99 | Total Loop p50 / p95 / p99 | Ops / Sec |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **1 Agent (5 iters)** | 6ms / 7ms / 7ms | 77ms / 84ms / 84ms | 3.32s / 3.58s / 3.58s | 9ms / 10ms / 10ms | **3.42s / 3.66s / 3.66s** | 1.37 ops/s |
| **5 Agents (10 iters)** | 9ms / 30ms / 30ms | 88ms / 112ms / 112ms | 5.67s / 6.66s / 6.66s | 12ms / 21ms / 21ms | **5.82s / 6.77s / 6.77s** | 1.48 ops/s |
| **20 Agents (40 iters)** | 49ms / 577ms / 674ms | 155ms / 252ms / 442ms | 21.53s / 22.72s / 22.74s | 78ms / 259ms / 320ms | **21.91s / 23.38s / 23.40s** | 0.86 ops/s |
| **50 Agents (50 iters)** | 265ms / 1.30s / 1.42s | 312ms / 424ms / 486ms | 54.67s / 56.60s / 56.63s | 185ms / 673ms / 878ms | **56.09s / 57.94s / 58.17s** | 0.86 ops/s |

**Critical Observation:**  
Git Push (`git.receive_pack` Smart-HTTP) accounts for **90%–97%** of the entire agent execution time across all concurrency tiers. The push operation invokes `refweird hook pre-receive` subprocess, which executes secret pattern scanning and branch protection verification on pushed objects. In contrast, MCP lease generation (6ms–265ms) and PR creation (9ms–185ms) remain sub-second even under 50 concurrent agents.

### 7.3 SQLite WAL Microbenchmarks (`internal/store/store_bench_test.go`)

Measured using `go test -bench=BenchmarkStore -benchmem -run=^$ -count=5`:

| Benchmark | Latency (ns/op) | Memory (B/op) | Allocations (allocs/op) | Throughput / Capacity |
| :--- | :--- | :--- | :--- | :--- |
| **`BenchmarkStore_ReadQuery`** (single-row index lookup) | 14,912 ns (~14.9 µs) | 1,368 B | 46 allocs | ~67,100 queries/sec |
| **`BenchmarkStore_WriteTransaction`** (user insert + WAL flush) | 2,725,942 ns (~2.73 ms) | 945 B | 32 allocs | ~366 write tx/sec |
| **`BenchmarkStore_OpenDisk`** (DB open + 12 sequential migrations) | 85,847,630 ns (~85.8 ms) | 92,370 B | 785 allocs | ~11.6 opens/sec |

### 7.4 Git Engine Microbenchmarks (`internal/gitstore/gitstore_bench_test.go`)

Measured using `go test -bench=BenchmarkGitStore -benchmem -run=^$ -count=5`:

| Benchmark | Latency (ns/op) | Memory (B/op) | Allocations (allocs/op) | Throughput / Capacity |
| :--- | :--- | :--- | :--- | :--- |
| **`BenchmarkGitStore_CommitLogWalk`** (walk 200 commit objects) | 4,666,194 ns (~4.67 ms) | 532,469 B (~520 KB) | 10,618 allocs | ~214 full walks/sec |
| **`BenchmarkGitStore_InitBareRepo`** (disk bare repo creation + HEAD) | 5,003,142 ns (~5.00 ms) | 20,027 B (~19.5 KB) | 457 allocs | ~200 bare repos/sec |

### 7.5 Frontend Asset & Transfer Payloads

#### Web Product SPA (`internal/webui/dist/assets/`):
- Main JS Bundle (`index-C6RND4o2.js`): **712 KB uncompressed**, **197.3 KB gzipped**
- Main CSS (`index-DW-Z4Dc-.css`): **126 KB uncompressed**, **19.6 KB gzipped**
- Font (`mona-sans-latin-wght-normal.woff2`): **39 KB**
- Route Chunks:
  - `Card-Zc37UclV.js`: 93 KB
  - `Dialog-7c7IFL76.js`: 65 KB
  - `RepoSettingsView-CZhs8yxS.js`: 55 KB
- **Critical path initial transfer budget:** ~217 KB over the wire (gzipped HTML + JS + CSS).

#### Public Marketing & Docs Site (`site/out/`):
- Landing page HTML (`site/out/index.html`): **12 KB**
- Docs page HTML (`/docs/architecture`): **126 KB**
- Total static site bundle (`site/out/`): **11 MB** (all 28 static/SSG pages + assets)
- Total static JS chunks: **1.2 MB uncompressed**
- Search index: Generated statically per build.

---

## 8. Summary of Identified Bottlenecks for Subsequent Tasks

1. **Git Smart-HTTP Push Pipeline (Task 6):**
   - Push latency scales from 3.32s (1 agent) to 54.67s (50 agents).
   - Profiling reveals subprocess invocation of `refweird hook pre-receive` per push and secret regex scanning over Git pack object streams.
2. **Vitest JSDOM Creation Overhead in Web Testing (Task 5):**
   - 45 test files execute in 23.64s, but 109.68s of CPU time is spent recreating JSDOM instances (35% of total test runtime). Vitest `vmThreads` or worker pooling can optimize developer test feedback loop.
3. **Web SPA Route Code-Splitting (Task 3):**
   - The primary vendor bundle is 712 KB (197 KB gzip). High-probability route preloading via TanStack Router and lazy component chunking will minimize time to interactive on first navigation.
