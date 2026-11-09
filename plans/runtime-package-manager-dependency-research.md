# Deep Research — Runtime, Package Manager & Dependency Improvements
**Product:** Refweir / exist.io — self-hosted Git forge for coding agents
**Date:** 2026-11-09
**Scope:** Runtime environment, package manager, dependencies, build tooling, frontend framework — "drastic change for dramatically faster runtime / process time"
**Constraint:** Working only under `plans/`; existing plan files left untouched.

---

## 1. Executive summary

The product is **already fastest where it matters most**: the backend is pure-Go (Go 1.26/1.27, zero CGO, SQLite via `modernc.org/sqlite`, `go-git/v6`). Independent benchmarks show Go serves a small JSON API **4–6x the throughput of Node.js at ~6x less memory** (Lamplit Labs, Oct 2026). Moving the backend off Go onto a JavaScript runtime (Node/Bun/Deno) would be a **regression**, not an improvement.

The realistic, high-leverage improvements live entirely in the **frontend / dev-toolchain layer**:

| Area | Current | Best alternative | Verdict |
|------|---------|------------------|---------|
| Package manager | npm | pnpm 12 (+global store / pnpr) | **Adopt** — ~12x faster installs, 55x faster warm installs, lower disk use |
| Build tool | Vite 8 | Vite 8 (already) / Rolldown | **Already on it** — Vite 8 is Rust-based, up to 30x faster builds |
| JS runtime (dev/build) | Node.js | Bun | **Optional** — Bun is ~2.3x Node.js raw-HTTP throughput, but frontend runtime does not affect production browser performance |
| JS runtime (production) | Browser | Browser | **No change possible** — the SPA runs in the user's browser regardless of server runtime |
| Frontend framework | React 19 | Preact / Redact | **Not recommended** — smaller bundles (7–20 KiB vs 66 KiB gzip) but React is faster on most UI operations (mount, props, state, insertion) |
| Type checking | tsc (TS 6) | SWC / esbuild strip-types | **Optional** — faster CI typecheck, no type errors caught at build time |

**Bottom line:** The single biggest, lowest-risk win is switching the frontend package manager from npm to pnpm (12x faster installs). A secondary win is staying on Vite 8 (already adopted) for up to 30x faster builds. Migrating the frontend dev runtime to Bun is a marginal, optional optimization. Migrating the backend away from Go is strongly discouraged.

---

## 2. Product & current stack (as observed in the repo)

**Product:** Refweir — a vendor-neutral, self-hosted Git forge engineered for autonomous AI coding agents (Claude Code, Codex, Cline, Cursor) alongside human teams. Ships as a single pure-Go binary (`refweird`) with embedded web UI, pure-Go SSH server (:2222), native MCP server (/mcp), capability leases, SQLite WAL storage.

**Repositories/directories observed:**
- `/home/shiva/projects/exist.io/` — root (Go backend + frontend + site)
- `/home/shiva/projects/exist.io/web/` — React SPA (`refweir-web`)
- `/home/shiva/projects/exist.io/site/` — Next.js docs/marketing site (`refweir-site`)

**Current frontend stack (`web/package.json`):**
- Runtime/package manager: **npm** (lockfile `package-lock.json`), Node.js
- Framework: React 19.2.8, React DOM 19.2.8
- Router: TanStack Router ^1.170.46
- Data: TanStack Query ^5.107.1
- Styling: Tailwind CSS v4.3.3, class-variance-authority, clsx, tailwind-merge
- Animation: motion ^14.2.0
- Build: Vite ^8.3.0, `@vitejs/plugin-react` ^6.1.1
- Tests: Vitest ^5.0.4, jsdom, @testing-library/react
- Lint: oxlint ^1.81.0
- TypeScript: ~6.0.2, `tsc -b`
- E2E: Playwright ^1.64.1

**Current site stack (`site/package.json`):** Next.js ^16.4.4, React 19.3.1, fumadocs, Tailwind v4, npm.

**Backend stack (`go.mod`):** Go 1.26.0 / toolchain go1.27.3, `go-git/v6`, `modernc.org/sqlite` v1.60.1, `golang.org/x/crypto`. Zero CGO, compiles with `CGO_ENABLED=0`.

---

## 3. Research angles brainstormed

Before searching, the following angles were identified as relevant to "make the product faster / better runtime / better dependencies":

1. **JavaScript runtime alternatives** — Node.js (current) vs Bun vs Deno: startup time, HTTP throughput, memory, TypeScript support, npm/Node compatibility.
2. **Package manager alternatives** — npm vs pnpm vs yarn vs Bun install speed, disk usage, determinism, lockfile format, CI caching.
3. **Build tool alternatives** — Vite vs Turbopack vs Rolldown vs rspack vs esbuild: dev-server startup, production build speed, plugin compatibility.
4. **Frontend framework runtime** — React 19 vs Preact vs Redact vs Solid: bundle size, mount/update speed, hydration, ecosystem compatibility (TanStack Router/Query are React-first).
5. **TypeScript compile performance** — tsc vs SWC vs esbuild strip-types: CI typecheck speed, type-safety trade-offs.
6. **Backend runtime** — Go vs Rust vs Bun/Deno/Node: throughput, memory, binary size, concurrency (already Go; validated below).
7. **Dependency/supply-chain hygiene** — audit, ghost dependencies, unused packages, lockfile determinism, SBOM.
8. **Structural/drastic changes** — monorepo tooling (Turborepo/Nx), removing the Node toolchain entirely from the frontend, shipping the UI as static assets embedded in the Go binary (already done per launch plan).

---

## 4. Deep research findings

### 4.1 Runtimes: Node.js vs Bun vs Deno

**HTTP JSON API throughput (Lamplit Labs, Oct/Nov 2026; same machine, MacBook Pro M2 Pro, 8 perf cores, 32 GB, same GET /users/:id returning 1.2 KB JSON, in-memory, no DB/TLS/middleware, measured with wrk):**

| Implementation | req/s | Relative to Node http |
|---|---|---|
| Node.js 22 built-in `http` | 54,800 | 1.0x (baseline) |
| Node.js 22 + Fastify 5 | 71,200 | 1.3x |
| Deno 2 `Deno.serve` | 98,600 | 1.8x |
| Bun 1.1 `Bun.serve` | 162,300 | **3.0x** |

Bun is the fastest JavaScript runtime for raw HTTP throughput (~3x Node.js built-in, ~1.6x Deno). Deno sits between Node+Fastify and Bun.

**Deno npm/Node compatibility (Deno docs, "Node and npm Compatibility," last updated 2026-07-30):**
- Most Node.js code runs in Deno without modification.
- `node:` built-ins supported (`node:http`, etc.); Node globals like `process` and `Buffer` supported.
- npm packages installable via `npm:` specifiers or `deno add`; `deno.json` manages lockfile (`deno.lock`).
- CommonJS/ESM interop, `require()` from ESM, `node_modules` modes (isolated/hoisted), native addons **not supported**.
- Supply-chain guards: minimum dependency age (24h default since Deno 2.9), publishing-trust policy (`no-downgrade`), `.npmrc` mTLS support (Deno 2.8+).

**What this means for Refweir:**
- The frontend is a **browser SPA**. The production bundle runs in the user's browser, not on Node/Bun/Deno. Switching the *dev* runtime to Bun only speeds up local `dev`/`build`/`test` — not end-user runtime.
- The backend is Go. Rewriting it in Bun/Deno/Node would **lose** 4–6x throughput and **gain** ~6x memory usage (per the Go vs Node.js benchmark noted in Lamplit Labs' sidebar). Not recommended.
- Deno's Node compatibility is high but not 100%; native addons are unsupported, which matters if any frontend dep pulls in native code (current deps appear pure-JS).

### 4.2 Package managers: npm vs pnpm vs yarn vs Bun

**pnpm official benchmarks (pnpm.io, fetched 2026-11-07; npm v12.2.2, pnpm 11 / pnpm 12; two projects: "alotta-files" wide small-package graph and "alotta-packages" ~1 GB of modern toolchains; same machine, emulated 50ms/200Mbps link; fastest of 3 runs):**

**alotta-files (wide graph of small packages):**

| Scenario | npm | pnpm 11 | pnpm 12 | pnpm 12 + pnpr | pnpm 12 + global store |
|---|---|---|---|---|---|
| clean (fresh clone, nothing cached) | 49.0 s | 9.18 s | 4.68 s | 2.72 s | 4.93 s |
| cache (reinstall, no lockfile) | 13.9 s | 4.69 s | 997 ms | 781 ms | 630 ms |
| lockfile (CI, first install) | 12.7 s | 5.16 s | 3.30 s | 2.58 s | 3.33 s |
| cache + lockfile (reinstall, warm) | 8.90 s | 2.88 s | 744 ms | 753 ms | 237 ms |
| everything warm (nothing changed) | 1.26 s | 716 ms | 23 ms | 22 ms | 23 ms |
| update (bump dep in package.json) | 4.58 s | 4.54 s | 986 ms | 971 ms | 1.04 s |

**Totals across all scenarios:** npm 3m 34s; pnpm 11 1m 13s; pnpm 12 23.3 s; **pnpm 12 + pnpr 7.83 s (fastest)**. pnpm 12 is **~12x faster than npm overall**; on warm reinstalls pnpm is **55x faster** (23 ms vs 1.26 s).

**alotta-packages (~1 GB, fourteen modern toolchains):** pnpm 12 again leads every scenario; pnpr gains most on the larger graph.

**Key pnpm characteristics:** content-addressable global store, hardlinks/symlinks instead of copying, deterministic lockfile (`pnpm-lock.yaml`), strict `node_modules` (no phantom/ghost deps by default), built-in Node.js version management (faster than fnm/nvm), workspaces/monorepo support.

**Verdict:** Switching `web/` (and `site/`) from npm to pnpm is the single highest-ROI change: **12x faster installs**, smaller disk footprint, stricter dependency hygiene, and no ecosystem breakage (pnpm installs npm packages from the same registry).

### 4.3 Build tools: Vite vs Turbopack vs Rolldown vs rspack vs esbuild

**Vite 8 (InfoQ, May 2026): "Vite Version 8: Unified Rust-Based Bundler and Up to 30x Faster Builds."** Vite 8 unifies on a Rust-based bundler (Rolldown, the same bundler powering Rsbuild). The repo is already on `vite ^8.3.0`, so this improvement is **already adopted**; no migration needed.

**Context:** Rolldown is a Rust bundler compatible with the Rollup plugin ecosystem; esbuild/rspack/Turbopack are alternatives but require leaving the Vite/Rollup plugin model. Given Vite 8 already delivers up to 30x faster builds with Rollup-plugin compatibility, there's no compelling reason to switch build tools.

**Verdict:** Stay on Vite 8. Do not migrate to Turbopack/Rspack/esbuild solely for speed; the gain is already captured.

### 4.4 Frontend framework runtime: React 19 vs Preact vs Redact

**TanStack Redact docs ("React vs Redact vs Preact," measured Oct 2026, React 19.3.1):**

| Metric | React 19.3.1 | Redact | Preact compat 11.1.0 |
|---|---|---|---|
| Overall speed score (5 list workloads, React=100) | 100 | 78 | 73 |
| Test bundle, gzip | 66.3 KiB | 19.8 KiB | 7.1 KiB |
| Mount/unmount 500 rows | 0.371 ms | 0.421 ms | 0.366 ms |
| Update props 500 rows | 0.128 ms | 0.218 ms | 0.197 ms |
| Update state 500 rows | 0.112 ms | 0.210 ms | 0.200 ms |

React is faster than Preact/Redact on props, state updates, and insertion/removal; Preact is comparable on mount; Redact wins on list reversal. Bundle sizes: Preact 7.1 KiB, Redact 19.8 KiB, React 66.3 KiB (gzipped test app).

**Verdict:** Switching from React to Preact/Redact **reduces bundle size** but **slows most UI operations** and risks TanStack Router/Query compatibility (those libraries are React-first; the compatibility matrix shows gaps in React 19 actions, deferred results, SSR modes). Not a recommended runtime win for this product.

### 4.5 Backend runtime: Go is already optimal

**Lamplit Labs sidebar ("Go vs Node.js for a Small JSON API"):** Go 1.27 vs Node.js 24 serving the same 50-item JSON endpoint under 64 connections — **4–6x throughput, ~6x less memory**. Refweir's backend is already pure-Go with zero CGO, SQLite WAL, and `go-git` directly on bare repos. The launch plan explicitly benchmarks against Forgejo and publishes p50/p95 numbers.

**Verdict:** Do not move the backend off Go. The runtime "marvel" on the server side is already in place.

### 4.6 Dependency & supply-chain hygiene (improvement without migration)

- Current `web/` deps are modest (~19 runtime, ~15 dev). `package-lock.json` present.
- pnpm's strict `node_modules` eliminates ghost/phantom dependencies (importing packages not listed as dependencies), which is a real correctness/security win.
- Deno's supply-chain guards (min dependency age, trust-policy/no-downgrade, `.npmrc` mTLS) are notable if the project ever adopts Deno, but pnpm + `npm audit`/`govulncheck`/`gosec` cover the current stack adequately.
- SBOM and signed releases are already on the launch roadmap (docs/security/SBOM.md).

---

## 5. Questions to grill down (decision framework)

Before committing to any change, answer these:

1. **Where is the product actually slow today?** Frontend dev-server startup? Production build time? `npm install` in CI? Browser page-load/runtime? Git clone/push (backend)? API latency? Baseline metrics should be recorded first (launch plan G0/G1 references baseline metrics).
2. **Does the frontend runtime affect end users?** No — the SPA runs in the browser. Only dev/CI speed is affected by Node vs Bun vs Deno.
3. **Is the backend a candidate for runtime change?** No — Go is 4–6x faster and 6x lighter than Node.js; moving away is a regression.
4. **What's the acceptable compatibility risk?** React + TanStack ecosystem is React-first; Preact/Redact introduce compatibility gaps. npm→pnpm is near-zero risk (same registry, same packages).
5. **Team/single-developer constraints?** The launch plan notes a one-person team with a tight 4-week window. Changes must be low-risk and reversible.
6. **Migration effort vs benefit?** pnpm switch: ~1 hour (change lockfile, CI install command, docs). Bun runtime switch: moderate (rewrite dev scripts, CI, potential dep incompat). Framework switch: high (rewrite components, re-verify TanStack compat).
7. **Are builds cached in CI?** pnpm + warm cache gives 22–23 ms installs; that may already be fast enough.

---

## 6. Options & recommendations

### Option A — Do nothing structurally (recommended baseline)
Keep Go backend, npm, Vite 8, React 19. Rely on Vite 8's up-to-30x faster builds (already adopted). Record baseline metrics before declaring a problem.

### Option B — Switch frontend package manager npm → pnpm (recommended first change)
**Benefit:** ~12x faster installs overall, 55x faster warm reinstalls, smaller disk, strict `node_modules`, determinism.
**Risk:** Near zero. Same registry, same packages. CI install command changes (`pnpm install` / `pnpm run ...`); `package-lock.json` → `pnpm-lock.yaml`; update CI caching to cache pnpm store.
**Effort:** Low (~1 hour + CI config).
**Files affected:** `web/package.json` (scripts), `site/package.json` (scripts), CI workflows (`.github/workflows/`), docs.

### Option C — Switch frontend dev runtime Node → Bun (optional)
**Benefit:** Bun HTTP ~3x Node raw throughput (relevant only if the frontend ever runs a Node dev server under load); faster `bun install`, `bun run`, `bun test`.
**Risk:** Low-moderate. Some Vite/Vitest/Playwright/oxlint plugins assume Node APIs. Benchmark locally first.
**Effort:** Moderate. Requires testing `vite`, `vitest`, `playwright`, `oxlint`, `tsc` under Bun.

### Option D — Structural/drastic restructure (not recommended)
- Remove Node toolchain from frontend entirely: not feasible (React/Vite/TanStack require a JS build toolchain).
- Replace React with Preact/Redact: smaller bundles but slower UI ops and compatibility gaps — net negative.
- Move backend to Bun/Deno/Node: 4–6x slower, 6x more memory — net negative.
- Monorepo tooling (Turborepo/Nx): only valuable if `web/` and `site/` grow much larger; adds complexity for a one-person team.

### Recommended sequence
1. Record baseline metrics (install time, build time, dev startup, API p50/p95).
2. Switch `web/` and `site/` to pnpm; update CI; measure.
3. Confirm Vite 8 is in use (already is) — no build-tool migration.
4. Optionally benchmark Bun for dev scripts only if pnpm isn't sufficient.
5. Do not change the backend runtime or frontend framework.

---

## 7. Risks & trade-offs

| Change | Risk | Mitigation |
|---|---|---|
| npm → pnpm | CI cache miss on first run; docs/scripts reference npm | Cache pnpm store; update all scripts/CI; keep `npm` as fallback in docs temporarily |
| Node → Bun dev runtime | Vite/Vitest/Playwright/oxlint plugin incompatibility | Run full `npm test` + `npm run build` under Bun before committing |
| React → Preact/Redact | TanStack Router/Query compatibility gaps; slower UI ops | Do not pursue; measured React faster on most ops |
| Backend off Go | 4–6x throughput loss, 6x memory increase | Do not pursue; Go is the strong choice |

---

## 8. Sources consulted

- Lamplit Labs, "Node.js 22 vs Deno 2 vs Bun 1.1: HTTP JSON API Throughput, Measured with wrk" (Oct/Nov 2026): Bun 162,300 req/s, Deno 98,600, Node+Fastify 71,200, Node http 54,800.
- Lamplit Labs sidebar, "Go vs Node.js for a Small JSON API: 4-6x Throughput, 6x Less Memory."
- pnpm.io Benchmarks (fetched 2026-11-07; npm v12.2.2 vs pnpm 11/12): pnpm 12 ~12x faster than npm overall; 23 ms warm install vs npm 1.26 s (55x).
- InfoQ, "Vite Version 8: Unified Rust-Based Bundler and Up to 30x Faster Builds" (May 2026).
- Deno docs, "Node and npm Compatibility" (updated 2026-07-30): `node:` built-ins, npm packages, CommonJS/ESM, no native addons, supply-chain guards.
- TanStack Redact docs, "React vs Redact vs Preact" (Oct 2026): React speed 100 / Redact 78 / Preact 73; bundles 66.3 / 19.8 / 7.1 KiB gzip.
- Repo files: `go.mod`, `web/package.json`, `site/package.json`, `AGENTS.md`, `README.md`, `plan.md`, `plans/launch-plan.md`, `research/exist-io-ph-launch-frontend-redesign.md`.

---

## 9. Suggested next steps (actions under `plans/` only)

1. Create a short decision record in `plans/` (this file is it).
2. If approved, create `plans/pnpm-migration.md` with the exact CI/script changes (no existing files modified).
3. Baseline: time `npm install`, `npm run build`, `npm test` in `web/` and `site/`; repeat after pnpm switch.
4. Keep existing `plans/*` files untouched.
