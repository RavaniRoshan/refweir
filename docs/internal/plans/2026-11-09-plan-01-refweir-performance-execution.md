# Plan 01 — Refweir Performance and Frontend Improvement

> **For agentic workers:** Use the `executing-plans` skill to follow this plan in order. Use TDD for every code change. This is a gate-driven program plan: do not implement a candidate optimization unless its measurement gate passes.

**Goal:** Find and ship only measurable improvements to Refweir’s user-facing performance and developer loop while preserving its self-hosted product model and correctness guarantees.

**Architecture:** Keep the pure-Go modular forge and embedded React product UI as the default. Keep marketing and documentation together in the owner-approved static Next.js site, separate from the authenticated product UI. Establish workload-specific baselines, then optimize the product frontend, Git/database hot paths, or CI only where measurements identify a real bottleneck; reserve topology changes for a separate gated decision.

**Tech Stack:** Go, `go-git/v6`, SQLite WAL via `modernc.org/sqlite`, React/Vite product SPA, Next.js/Fumadocs static site, existing npm/lockfiles pending inventory.

**Spec:** [Refweir performance research](2026-11-09-refweir-performance-research.md); owner-approved public-site direction in [site unification plan](2026-11-site-unification-redesign.md).

## Global Constraints

- Go backend builds with `CGO_ENABLED=0` on the supported architectures.
- SQLite uses `modernc.org/sqlite` in WAL mode with `PRAGMA foreign_keys = ON` unless a separately approved break-glass decision changes that requirement.
- Git operations use `github.com/go-git/go-git/v6` directly against bare repository directories unless a separately approved break-glass decision changes that requirement.
- In production (`isDemoModeActive() === false`), never default to or flash mock data.
- Preserve the Primer design system and accessible navigation in the product UI; public marketing/docs keep the Paper × Yellow design already specified in the existing site plan.
- Preserve archive/write restrictions, repository security boundaries, and danger-zone safeguards while changing caches, route data freshness, Git execution, or background work.
- Keep full validation available: `go test ./...`, frontend Vitest suites, `npm run build`, and the site’s `npm run verify` workflow.
- Builds and checks must complete without warnings or errors, consistent with the project handbook.
- Do not weaken CI gates or security checks to improve timing.
- All research and plan artifacts for this initiative stay under `plans/`; existing files in that directory remain untouched.

## Review Focus

1. **Stale repository policy/data:** visibility, archive state, branch protection, and authorization must not be served from stale cache state; test invalidation and server confirmation.
2. **Large or adversarial Git inputs:** many refs, large histories, large objects, and concurrent fetches must not cause unbounded memory growth or bypass hooks; benchmark peak RSS and retain security tests.
3. **Virtualized UI accessibility:** keyboard navigation, focus retention, screen-reader semantics, and scroll behavior must remain correct for trees, lists, and diffs.
4. **Static-site routing:** direct load and refresh for nested docs, `/exist.io` base paths, trailing slash routes, and legacy redirects must work on GitHub Pages.
5. **Cache and task graph correctness:** a hit must never return stale build output or skip required validation; verify cache keys, inputs, outputs, and full protected-branch gates.

---

## Decision rules

- No runtime/framework/database/service migration starts before the baseline task is complete.
- Compare repeated runs on the same hardware, OS, dataset, and configuration. Use medians and a statistical comparison for Go microbenchmarks; record run-to-run variance for builds and browser tests.
- A change is accepted only when the target metric improves beyond observed noise, the project’s agreed user/developer target is met or moved materially closer, and correctness/accessibility/security checks remain green.
- Set the exact pass threshold for each target after recording its baseline; do not copy a vendor benchmark’s percentage as Refweir’s success criterion.
- A package-manager experiment must preserve the dependency graph and run both frontend projects’ full build/test validation. A runtime experiment must keep the production browser bundle and product behavior equivalent.
- Any break-glass proposal requires a separate ADR with trigger metric, security/operations impact, rollback, and an operator migration path before implementation.

## Work packages

### Task 1: Inventory versions, workloads, and current validation

**Files:**

- Create: `plans/2026-11-09-refweir-performance-baseline.md`
- Read only: `go.mod`, `web/package.json`, frontend lockfiles, `site/package.json`, site lockfiles, relevant `.github/workflows/*`, and current benchmark/metrics configuration.

**Deliverable:** A version/workflow inventory and named workload set. No source or workflow changes in this task.

- [ ] Record exact Go/toolchain, `go-git/v6`, SQLite driver, Node, React, Vite, Vitest, Next, Fumadocs, package-manager, and lockfile versions.
- [ ] Record CI provider, runner OS/size, workflow stages, cache keys, cache hit rate, and site deployment path.
- [ ] Map the exact source, test, and workflow files for each measured journey when implementation work begins; record this map before any code edit. This plan does not guess function names because this research was constrained to `plans/`.
- [ ] Name representative product journeys: repo overview → file, repo → pull request → changed files, settings read/write, and agent/MCP request.
- [ ] Name Git fixtures from observed/available repositories: small repo, large history, many refs, and large diff; record sizes/ref counts/commit counts.
- [ ] Record supported self-host deployment shapes and expected concurrency supplied by product requirements; if unknown, document that explicitly rather than inventing a target.
- [ ] Run existing correctness gates once and record exact commands/results as the pre-change state.

**Verification:** The inventory identifies every current runtime and lockfile and the named workloads have concrete sizes/configurations. Existing files under `plans/` remain unchanged.

**Implementation rule:** For every code change in Tasks 3–7, write or update the narrowest relevant test first, run it to confirm the expected failure, implement the smallest change, rerun the test, then run that task’s full suite.

### Task 2: Capture repeatable baselines for each performance clock

**Files:**

- Modify: `plans/2026-11-09-refweir-performance-baseline.md`
- Candidate measurement areas after source inspection: `internal/metrics/`, `internal/app/`, `internal/gittransport/`, `internal/gitstore/`, `internal/store/`, product browser tests, and site audit scripts.

**Deliverable:** A reproducible before-state with p50/p95/p99, throughput, failures, CPU/RSS, and run conditions.

- [ ] Measure `go test ./...`, Go builds, `web` install/build/test, `site` install/build/verify, and container build by stage; record clean and warm runs separately.
- [ ] Measure API/MCP request duration and errors by low-cardinality route/tool name; do not label metrics with repo IDs or raw paths.
- [ ] Measure clone/fetch/push duration and peak RSS with the named repository fixtures at increasing concurrency.
- [ ] Measure SQLite read/write transaction latency, lock waits/errors, checkpoint duration, pool wait time, and WAL size under a representative mixed workload.
- [ ] Measure browser LCP/INP/CLS and route JS transfer for product overview, file/tree, large PR diff, marketing landing, docs home, and docs search on desktop and mobile conditions.
- [ ] Run Go microbenchmarks at least 10 times where feasible and compare with `benchstat`; for expensive installs/builds, record enough repeated clean/warm runs to quantify variance.
- [ ] Store hardware, OS, browser, network throttling, cache state, fixture sizes, and tool versions beside every result.

**Verification:** Another engineer can rerun each workload from the baseline file and produce comparable results. No optimization is selected solely from a vendor benchmark.

### Task 3: Optimize the authenticated product frontend only where traces show cost

**Files:**

- Candidate implementation: measured route/component files under `web/src/` and any existing route/data layer in `web/src/lib/`.
- Tests: the relevant `web/src/**/__tests__/*.test.tsx` suites and existing Playwright route tests.

**Deliverable:** A route-level change with before/after browser evidence, or a documented no-change decision if the routes meet targets.

- [ ] Identify slow navigation waterfalls and large route payloads from Task 2; test TanStack Router intent preloading only on high-probability destinations.
- [ ] For stable data, tune route-specific query freshness; for mutable issues/reviews/branches/commits, add explicit invalidation tests after successful writes.
- [ ] Benchmark tree/list/diff data at representative sizes. Paginate or incrementally fetch oversized datasets before considering virtualization.
- [ ] If virtualization is warranted, add keyboard, focus, screen-reader, expansion, and scrolling tests before enabling it in the route.
- [ ] If diffs are the bottleneck, measure line rendering/state allocation first; keep menus/comments out of per-line state where the existing architecture permits and compare memory/INP.
- [ ] Defer a browser editor/rich diff bundle to routes that use it only if the inventory confirms such a component exists and its load contributes materially to route cost.
- [ ] Keep destructive/security-sensitive mutations server-authoritative; do not let optimistic UI imply visibility, transfer, delete, merge, or archive success before confirmation.

**Verification:** Target route p75 INP is within the agreed budget or materially improved beyond run variance; route bundle size and memory do not regress; all component, accessibility, and end-to-end tests pass.

### Task 4: Keep the public marketing/docs site static and verify its real user paths

**Files:**

- Candidate implementation: `site/app/`, `site/public/`, `site/scripts/`, and the existing Pages workflow, following `plans/2026-11-site-unification-redesign.md`.
- Tests: existing site `typecheck`, `check-links`, `claims-audit`, and build/verify scripts; browser route checks where available.

**Deliverable:** An audited static output with route/performance/accessibility results. Do not change the existing site design or architecture plan as part of this performance plan.

- [ ] Capture field Web Vitals where traffic exists and repeatable lab LCP/INP/CLS for landing, docs home, and search.
- [ ] Inspect initial HTML and compressed route JS/CSS/image/font sizes; remove or defer only payloads shown to affect the measured routes.
- [ ] Measure Fumadocs search-index compressed size and mobile search interaction. Keep local static search while it fits the agreed download/latency budget; evaluate hosted search only if the measured index exceeds that budget.
- [ ] Verify canonical URLs, titles/descriptions, sitemap entries, internal marketing↔docs paths, basePath, trailing slash, direct nested route loads, refresh, 404, and redirect stubs on the deployed-host path.
- [ ] Audit keyboard/focus/target-size behavior, contrast, reduced-motion, and responsive layouts at the approved widths.
- [ ] If considering Astro, build a separate comparison prototype only after the existing static Next output misses a target; compare bundle/vitals, Fumadocs compatibility, migration effort, and operational cost before any architecture change.

**Verification:** `cd site && npm run verify` passes; direct loads and refreshes work for sampled public routes; WCAG review has no critical keyboard/focus/contrast issue; metrics are reported with route and device context.

### Task 5: Test npm versus pnpm and remove measured CI waste

**Files:**

- Candidate only after a passing experiment: `web/package.json`, its lockfile, `site/package.json`, its lockfile, and `.github/workflows/*`.
- Preserve all existing scripts and validation commands.

**Deliverable:** A controlled manager/cache comparison and a written adopt/reject decision.

- [ ] Compare npm `ci`/clean install with a pinned pnpm version on both actual dependency graphs; run cold and warm cache trials on the same runner class.
- [ ] Record elapsed time, cache restore/save time, disk use, lockfile behavior, lifecycle scripts, and dependency-resolution differences.
- [ ] If pnpm’s gain exceeds noise and compatibility checks pass, migrate one frontend at a time, pin the package manager, use frozen/CI installs, and update workflow cache paths.
- [ ] Benchmark built-in `setup-go` and `setup-node` caches before adding Nx/Turborepo. Do not cache `node_modules` unless the chosen manager’s documented layout and cache strategy justify it.
- [ ] Consider Playwright sharding only when browser tests dominate wall time; retain all tests on protected/release workflows.
- [ ] Do not switch Node to Bun/Deno as part of a package-manager trial; keep the runtime and package-manager variables separate.

**Verification:** `cd web && npm test && npm run build` / `cd site && npm run verify` pass when npm remains selected; use the corresponding `pnpm` script commands if pnpm is adopted. Full Go validation remains green; clean and warm CI time improve beyond run variance; dependency graphs and built artifacts are reviewed for unintended changes.

### Task 6: Benchmark Go and Git hot paths before changing the server runtime

**Files:**

- Candidate benchmarks/tests under `internal/gittransport/` and `internal/gitstore/`; any implementation change must have a package-level test in `internal/<pkg>/<pkg>_test.go`.
- Dependency review: `go.mod` and `go.sum` only if an update is supported by the benchmark.

**Deliverable:** A measured Go/Git decision with a safe implementation or a no-change result.

- [ ] Confirm the exact `go-git/v6` revision and check whether the optimization in upstream PR #2351 is already present before reproducing it.
- [ ] Add/reuse a benchmark that covers one wanted ref and many refs on a shared-history fixture; record bytes allocated, elapsed time, and peak RSS with repeated runs.
- [ ] Run clone/fetch/push fixtures with large objects and concurrent requests; ensure pre-receive hooks, archive restrictions, deploy-key boundaries, and branch protections remain active.
- [ ] Profile CPU/heap/block/mutex before considering Go PGO. Keep PGO out if the representative CPU profile does not show a meaningful opportunity.
- [ ] If a native Git comparison is warranted, run it only as a benchmark in an isolated harness. Do not add a host Git runtime dependency without a separate break-glass approval and security review.

**Verification:** `go test ./...` passes, `CGO_ENABLED=0` builds remain supported, benchmarks report conditions and variance, and every Git security invariant is covered by regression tests.

### Task 7: Tune SQLite before adding infrastructure

**Files:**

- Candidate implementation/tests under `internal/store/` with migrations only if schema evidence requires them; do not add a migration for a benchmark-only change.

**Deliverable:** A database tuning result tied to measured lock, query, or checkpoint pressure.

- [ ] Capture query plans and transaction boundaries for the slowest measured reads/writes.
- [ ] Test the existing connection-pool limit, transaction length, indexes, and checkpoint policy under the same workload; measure memory, pool waits, WAL size, and p95/p99.
- [ ] Add safe retry/transaction tests only for observed lock behavior; specifically test read-then-write `SQLITE_BUSY_SNAPSHOT` cases if the code uses that pattern.
- [ ] Consider another pure-Go driver only if SQLite CPU dominates profiles; benchmark the exact Refweir WAL workload.
- [ ] Do not add Redis cache for authorization/archive/visibility/protection unless the cache invalidation/security model is explicit and tested.
- [ ] Open a separate ADR only if sustained writer waits or multi-host requirements make PostgreSQL a real need.

**Verification:** Store tests pass, the mixed workload meets the agreed latency/error target, WAL/checkpoint behavior is stable, and security decisions never depend on unbounded stale data.

### Task 8: Evaluate break-glass scale options only after a failed capacity gate

**Files:**

- Create a separate ADR in `plans/` before changing deployment architecture.
- Candidate areas depend on that ADR: background worker wiring, database configuration, object storage, Git storage routing, Helm/deployment profiles.

**Deliverable:** A go/no-go architecture decision, not an automatic microservice migration.

- [ ] Trigger this task only when a representative deployment repeatedly misses its agreed user or capacity target after local query, code, and workload tuning.
- [ ] Compare optional worker queues, PostgreSQL, object storage for immutable LFS/artifacts, repository sharding/Git service, and external cache as separate options.
- [ ] For each option, document latency/network hops, operations, security boundaries, backup/recovery, self-hosted installation burden, failure mode, migration, and rollback.
- [ ] Preserve the single-binary path as the documented default unless the owner explicitly chooses a new default.

**Verification:** The ADR names a measured trigger, concrete evidence, operational cost, rollback, and compatibility/security plan; no service is added solely for hypothetical scale.

### Task 9: Re-run gates and publish the measured decision record

**Files:**

- Update only new files created for this initiative under `plans/`; leave all pre-existing plan files unchanged.
- Update code/package/workflow files only for changes that passed their task gate.

**Deliverable:** A before/after scorecard and a short accepted/deferred/rejected change list.

- [ ] Re-run the exact baseline workloads from Task 2 against the final candidate.
- [ ] Re-run `go test ./...`, `cd web && npm test && npm run build`, and `cd site && npm run verify`.
- [ ] Compare p50/p95/p99, errors, throughput, CPU/RSS, browser payload/vitals, and CI stage time under the same environment.
- [ ] Record regressions, run variance, and any change that was reverted; retain enough detail to reproduce the scorecard.
- [ ] Mark a candidate accepted only if the target gain exceeds measurement noise and all security, correctness, accessibility, and platform gates pass.

**Completion contract:** The scorecard can be rerun; all accepted changes have green validation and a measured before/after; all deferred and rejected changes have a reason; existing `plans/` files remain untouched.

## Execution order

Run Tasks 1–2 first. Tasks 3–7 are independent after the baseline and can be split into small reviewed changes. Task 8 is conditional and must not run until its trigger is met. Task 9 closes the program. At implementation time, use an isolated worktree and split the frontend, CI/toolchain, Git, and data changes into separate reviewable branches if they do not share code or acceptance gates.

## Deferred until measurement justifies them

- Replacing Go with Node, Bun, Deno, or Rust.
- Replacing React or Vite wholesale.
- Moving the public site off the owner-approved static Next.js/Fumadocs architecture.
- Adding a micro-frontend, Nx, or Turborepo solely because the repository has multiple frontend folders.
- Replacing SQLite or introducing Redis/PostgreSQL/Gitaly as the default self-hosted topology.
- Using a runtime or framework vendor benchmark as a forecast for Refweir.
