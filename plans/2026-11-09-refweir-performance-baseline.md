# Refweir Performance Baseline & Inventory

**Date:** 2026-11-09  
**Host Environment:** Linux RoshansdaamPC 6.6.87.2-microsoft-standard-WSL2 (AMD Ryzen 7 4800H, 16 vCPUs, 7.5 GiB RAM)  
**Status:** Task 1 Inventory Complete, Task 2 Measurements in Progress  

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

To ensure repeatable measurements across Git transports and store engines:

1. **Fixture A: Small Repository (`fixture-small`)**
   - 1 branch (`main`), 10 commits, 5 files (< 50 KB total tree size).
   - Simulates initial repo creation, quick agent clones, simple commits.
2. **Fixture B: Large History (`fixture-large-history`)**
   - 1 branch, 1,000 commits with sequential linear history, 100 KB tree size.
   - Simulates established projects with deep commit logs; evaluates commit walk and rev-list performance.
3. **Fixture C: Many Refs (`fixture-many-refs`)**
   - 256 branches/tags sharing commit objects.
   - Tests `upload-pack` advertise-refs negotiation, matching `go-git` upstream PR #2351 benchmark topology.
4. **Fixture D: Large Diff / Large Objects (`fixture-large-diff`)**
   - Single commit modifying 100 files with 5,000+ line diffs and 2 MB binary object.
   - Evaluates diff generation latency, line rendering allocations, and memory pressure during pack building.

---

## 5. Deployment Shapes and Expected Concurrency

- **Topology Default:**
  - Single-binary deployment (`refweird`), hosting HTTP `:8080` (Smart-HTTP + REST API + MCP) and SSH `:2222`.
  - Storage: Local filesystem directory for bare Git repos (`$REFWEIR_DATA_DIR/git/`) and SQLite database in WAL mode (`$REFWEIR_DATA_DIR/refweir.db`).
  - Zero external database (no PostgreSQL, no Redis required).
- **Enterprise Helm Shape (`deploy/helm/refweir-enterprise`):**
  - Kubernetes StatefulSet or single-replica Deployment (`replicaCount: 1`).
  - PersistentVolumeClaim for `/var/lib/refweir` (single-node lock constraint for SQLite WAL).
  - Multi-replica is currently gated by single-writer SQLite WAL constraints; scale-out read replicas or PostgreSQL clustering are break-glass options.
- **Expected Concurrency Profiles:**
  - Evaluated in `bench/main.go`:
    - Tier 1: 1 concurrent agent (baseline single-user latency)
    - Tier 2: 5 concurrent agents (small team / multi-agent pair programming)
    - Tier 3: 20 concurrent agents (active team peak burst)
    - Tier 4: 50 concurrent agents (stress limit for single-host WAL writer)

---

## 6. Pre-Change Correctness Gates Status

Executed on 2026-11-09:

1. **Go Backend Validation (`go test ./...`):**
   - Result: **PASS** (33 packages tested, 0 failures, total wall time ~17s).
2. **Site Documentation & Marketing Validation (`cd site && npm run verify`):**
   - Result: **PASS** (`typecheck` clean, Next.js 16.4.4 static build 28 pages generated, 45 pages / 649 links verified, 9 claims / 19 forbidden phrases verified).
3. **Frontend Vitest Suites (`cd web && npm test`):**
   - Result: **FAIL (1 suite failed out of 45, 44 passed, 166 tests passed)**
   - Root Cause: `src/app/__tests__/AppShell.test.tsx` fails with `ReferenceError: EOF is not defined` due to a syntax defect introduced in commit `ce90841` where `EOF` strings were accidentally appended to `src/lib/store.ts` (line 20), `src/lib/events.ts` (line 49), `src/lib/queryClient.ts` (line 12), and `src/app/AppShell.tsx` (line 83).
4. **Frontend Production Build (`cd web && npm run build`):**
   - Result: **FAIL** (TypeScript error `TS2304: Cannot find name 'EOF'` in 4 files mentioned above).
