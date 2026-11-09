# OpenAPI 3.1 Specification & TanStack Query Integration Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide an authoritative OpenAPI 3.1 specification served by the backend daemon and introduce type-safe TanStack Query hooks and cache management across the React frontend.

**Architecture:** 
1. Backend: Complete the OpenAPI 3.1 forge specification in `api/openapi.yaml`, serve it via `GET /api/openapi.yaml` and `GET /api/v1/openapi.yaml` in `internal/app`, and add an automated Go test ensuring all REST endpoints are documented.
2. Frontend: Implement a structured TanStack Query data layer (`web/src/lib/queries/`) with a type-safe `queryKeys` factory, custom hooks (`useRepoQuery`, `usePullsQuery`, `useIssuesQuery`, `useCommitsQuery`, etc.), mutations with targeted cache invalidation, and global query cache reset on sandbox demo mode switches.
3. Views: Incrementally migrate key views (`RepoOverview`, `PullRequestsView`, `GlobalHomeView`) to use TanStack Query hooks, eliminating manual loading state waterfalls and duplicate network requests.

**Tech Stack:** Go 1.24 (`modernc.org/sqlite`, `go-git/v6`), React 19, TypeScript 5.8, `@tanstack/react-query` v5.107, Vitest.

**Spec:** `plans/2026-11-09-plan-01-refweir-performance-execution.md` (Task 3: TanStack Query freshness & invalidation), `AGENTS.md` (Section 2 & 4).

## Global Constraints

- **Zero Mock Data in Production (`isDemoModeActive() === false`):** Mock data is strictly restricted to sandbox demo mode.
- **Pure-Go Backend (Zero CGO):** Backend compiles with `CGO_ENABLED=0` across all architectures.
- **Server Authoritative Mutations:** High-impact operations (repository transfer, deletion, visibility change, merges) must wait for server confirmation before mutating UI state.
- **Zero-Error Compilation:** `go test ./...`, `npm test` (all Vitest suites), and `npm run build` must pass with 100% success.

## Review Focus

1. **Query Cache Invalidation on Writes:** Creating a PR or merging a PR must invalidate the repository's pull request queries so UI stays fresh without page reloads.
2. **Demo Mode Isolation:** Toggling `refweir_sandbox_demo_mode` must clear the query cache (`queryClient.clear()` / `queryClient.resetQueries()`) to prevent mock data leaking into live sessions or vice versa.
3. **OpenAPI Schema Accuracy:** Endpoints documented in `api/openapi.yaml` must match the actual REST handler route signatures, query parameters, and response structures.
4. **Cancellation on Navigation:** Query hooks must pass `signal` from TanStack Query context to `fetch` calls so rapid tab switching aborts in-flight HTTP requests.
5. **No Regressions in Existing Views:** Migrating views to TanStack Query must retain existing Primer UI designs, keyboard shortcuts, and test assertions.

---

### Task 1: Complete and Serve OpenAPI 3.1 Forge Specification

**Files:**
- Modify: `api/openapi.yaml`
- Modify: `internal/app/app.go`
- Test: `internal/app/openapi_forge_test.go`

**Deliverable:** An OpenAPI 3.1 specification covering all primary Refweir REST routes (repos, pulls, issues, releases, commits, webhooks, audit, agent activities, health) served at `GET /api/openapi.yaml` with passing test coverage.

- [ ] Inspect existing `api/openapi.yaml` and expand paths to cover:
  - `/api/v1/repos/{owner}/{name}/commits`
  - `/api/v1/repos/{owner}/{name}/tree`
  - `/api/v1/repos/{owner}/{name}/blob`
  - `/api/v1/repos/{owner}/{name}/issues`
  - `/api/v1/repos/{owner}/{name}/issues/{number}`
  - `/api/v1/repos/{owner}/{name}/releases`
  - `/api/v1/repos/{owner}/{name}/webhooks`
  - `/api/v1/audit/events`
  - `/api/v1/cluster/metrics`
- [ ] Add `mux.HandleFunc("GET /api/openapi.yaml", ...)` and `"GET /api/v1/openapi.yaml"` in `internal/app/app.go` serving `api/openapi.yaml` with `Content-Type: application/yaml; charset=utf-8`.
- [ ] Write `internal/app/openapi_forge_test.go` verifying the route returns HTTP 200, valid OpenAPI 3.1 declaration, and covers key paths.
- [ ] Run `go test -v ./internal/app` and commit.

---

### Task 2: TanStack Query Keys & Query Hooks Architecture

**Files:**
- Create: `web/src/lib/queries/queryKeys.ts`
- Create: `web/src/lib/queries/useRepoQueries.ts`
- Create: `web/src/lib/queries/useAgentQueries.ts`
- Create: `web/src/lib/queries/useClusterQueries.ts`
- Create: `web/src/lib/queries/index.ts`
- Test: `web/src/lib/queries/__tests__/queryKeys.test.ts`
- Test: `web/src/lib/queries/__tests__/useRepoQueries.test.tsx`

**Deliverable:** A type-safe query layer with query key factories, custom hooks for fetching repositories, pull requests, issues, commits, file trees, and cluster health, complete with abort signal propagation.

- [ ] Create `queryKeys.ts` with hierarchical keys:
  - `repos.all`, `repos.detail(owner, name)`, `repos.files(owner, name, ref, path)`, `repos.commits(owner, name, ref)`, `repos.pulls(owner, name)`, `repos.issues(owner, name)`, `repos.releases(owner, name)`
  - `agent.activities`
  - `cluster.info`, `cluster.metrics`
- [ ] Implement query hooks in `useRepoQueries.ts`:
  - `useRepoInfoQuery(owner, name)`
  - `useFilesQuery(owner, name, ref, path)`
  - `useCommitsQuery(owner, name, ref)`
  - `usePullRequestsQuery(owner, name)`
  - `useIssuesQuery(owner, name)`
  - `useReleasesQuery(owner, name)`
- [ ] Pass `signal` to `api.ts` fetchers to support request cancellation.
- [ ] Wire window event listener for `refweir_mode_changed` in a query sync helper to reset the `QueryClient` cache on mode change.
- [ ] Write unit tests in `queryKeys.test.ts` and `useRepoQueries.test.tsx` verifying queries, caching, and signal passing.
- [ ] Run vitest and commit.

---

### Task 3: Mutation Hooks with Targeted Cache Invalidation

**Files:**
- Create: `web/src/lib/queries/useRepoMutations.ts`
- Modify: `web/src/lib/queries/index.ts`
- Test: `web/src/lib/queries/__tests__/useRepoMutations.test.tsx`

**Deliverable:** TanStack Query mutation hooks that execute writes and automatically invalidate associated queries.

- [ ] Implement `useCreatePullRequestMutation(owner, name)`:
  - Invokes `createPullRequest`
  - On success, invalidates `queryKeys.repos.pulls(owner, name)`
- [ ] Implement `useMergePullRequestMutation(owner, name)`:
  - Invokes `mergePullRequest`
  - On success, invalidates `queryKeys.repos.pulls(owner, name)` and `queryKeys.repos.commits(owner, name, defaultBranch)`
- [ ] Implement `useCreateIssueMutation(owner, name)`:
  - Invokes `createIssue`
  - On success, invalidates `queryKeys.repos.issues(owner, name)`
- [ ] Implement `useCreateReleaseMutation(owner, name)`:
  - Invokes `createRelease`
  - On success, invalidates `queryKeys.repos.releases(owner, name)`
- [ ] Write unit tests verifying cache invalidation in `useRepoMutations.test.tsx`.
- [ ] Run vitest and commit.

---

### Task 4: Migrate Key Views to TanStack Query

**Files:**
- Modify: `web/src/components/views/RepoOverview.tsx`
- Modify: `web/src/components/views/PullRequestsView.tsx`
- Test: `web/src/components/views/__tests__/RepoOverview.test.tsx`
- Test: `web/src/components/views/__tests__/PullRequestsView.test.tsx`

**Deliverable:** Refactored views utilizing the query layer for cleaner code, reduced re-renders, and instant cached navigation.

- [ ] In `PullRequestsView.tsx`:
  - Replace manual `loadPullRequests()` state handling with `usePullRequestsQuery(owner, repo)`.
  - Use `useMergePullRequestMutation` and `useCreatePullRequestMutation` for merge/create actions with automatic cache invalidation.
- [ ] In `RepoOverview.tsx`:
  - Wire `useRepoInfoQuery` and `useReleasesQuery` for repo metadata and releases.
- [ ] Ensure all 191+ tests in Vitest pass without regressions.
- [ ] Commit view migration.

---

### Task 5: Full Verification, Documentation & Merge Gate

**Files:**
- Verify: Full Go test suite (`go test ./...`)
- Verify: Full Vitest test suite (`npm test`)
- Verify: Production bundle build (`npm run build`)
- Verify: Capability counts (`gencounts`)
- Verify: Site docs & links (`npm run verify` in `site/`)

**Deliverable:** 100% green test suite across all verification gates, clean git status, and merge to `main`.
