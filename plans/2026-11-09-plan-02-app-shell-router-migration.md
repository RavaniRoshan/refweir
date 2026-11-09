# Authenticated App Shell & TanStack Router Migration Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Mount `RouterProvider` with TanStack Query in `web/src/main.tsx`, integrate `AppShell` with full GitHub Primer accessibility and Lucide SVG icons, wire the `newShell` layout flag to views, and ensure 100% test pass rate across all suites.

**Architecture:** TanStack Router (`createRouter`) manages client-side file-based routes (`web/src/routes/`) wrapped in `QueryClientProvider`. `web/src/routes/index.tsx` delegates to `web/src/App.tsx`. `App.tsx` supports both classic Primer tabbed layout and the new authenticated global rail `AppShell.tsx` gated by the `newShell` Zustand state and `?shell=1` search param. `AppShell` implements Primer accessibility (`role="tablist"`, `role="tab"`, `aria-selected`, explicit button `aria-label`s) and integrates the command palette (`⌘K`).

**Tech Stack:** React 19, TypeScript 6, Vite 8, `@tanstack/react-router`, `@tanstack/react-query`, Zustand, Tailwind CSS v4, Lucide React, Vitest, Testing Library.

**Spec:** `plan.md` (Phase 4: §4.1 Add TanStack Router, §4.5 New app shell, §4.8 Testing, performance, accessibility, §4.9 Strangler cutover).

## Global Constraints

- Pure-Go backend invariants remain intact: zero cgo, zero host OpenSSH/Git dependencies, pure-Go SQLite WAL.
- Zero mock data in live mode (`isDemoModeActive() === false`).
- Design system tokens strictly follow `web/src/styles/tokens.css` (OKLCH primitives) and `theme.css`.
- Primer accessibility: `role="tablist"`, `role="tab"`, `aria-selected`, 6px squircle radius, 1px hairline borders.
- All 45+ existing Vitest test files must remain 100% green without regressions.
- `npm run build`, `go run ./internal/tools/gencounts`, and `go test ./...` must pass with exit code 0.

## Review Focus

1. Rail navigation keyboard accessibility: navigating between Inbox, Runs, Reviews, Policies, and Repos via tab/focus states.
2. Theme toggle button state: must expose valid `aria-label` in both collapsed (`w-14`) and expanded (`w-56`) sidebar states.
3. Command palette trigger: clicking the top bar `⌘K` badge or pressing `Meta+K`/`Ctrl+K` opens `CommandPalette`.
4. URL and Query param persistence: loading `/?shell=1&tab=agents` activates the new rail shell and directs to Agent Console.
5. Backwards compatibility: rendering `<App />` in legacy unit tests without router context must still pass without exceptions.

---

### Task 1: Primer Accessibility & Polish in `web/src/app/AppShell.tsx`

**Files:**
- Modify: `web/src/app/AppShell.tsx`
- Test: `web/src/app/__tests__/AppShell.test.tsx`

**Interfaces:**
- Consumes: `useUIStore` from `@/lib/store`, `CommandPalette` from `@/components/CommandPalette`, icons from `lucide-react` (`Inbox`, `Play`, `CheckSquare`, `Shield`, `FolderGit2`, `ChevronLeft`, `ChevronRight`, `Sun`, `Moon`).
- Produces: `AppShell({ children, activeNav, onNavChange }: AppShellProps)` where `activeNav?: string` and `onNavChange?: (nav: string) => void`.

- [ ] **Step 1: Write the failing unit tests for accessible AppShell**

Add tests to `web/src/app/__tests__/AppShell.test.tsx` verifying:
1. Nav items render Lucide SVG icons instead of raw emojis.
2. Rail navigation uses `role="tablist"` with each item having `role="tab"` and `aria-selected`.
3. Sidebar toggle button has `aria-label="Toggle sidebar"` and `aria-expanded={sidebarOpen}`.
4. Theme toggle button has explicit `aria-label` (e.g. `Switch to light theme` or `Switch to dark theme`).
5. Top bar renders a button with `<kbd>⌘K</kbd>` that has `aria-label="Open command palette"` and clicking it opens the `CommandPalette`.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx vitest run src/app/__tests__/AppShell.test.tsx`
Expected: FAIL on missing `role="tablist"` / `role="tab"` and missing `aria-label` on theme button.

- [ ] **Step 3: Implement accessible Primer `AppShell` in `web/src/app/AppShell.tsx`**

- Replace emojis in `NAV_ITEMS` with Lucide icons: `Inbox`, `Play`, `CheckSquare`, `Shield`, `FolderGit2`.
- Wrap navigation buttons in `<nav role="tablist" aria-label="Global navigation">` and give each button `role="tab"` with `aria-selected={currentNav === item.id}`.
- Update sidebar toggle button with `aria-label="Toggle sidebar"` and `aria-expanded={sidebarOpen}`.
- Update theme toggle button with `aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}`.
- Wire `CommandPalette` into `AppShell`: triggerable via clicking the `⌘K` badge button (with `aria-label="Open command palette"`) and via `useEffect` listener for `Meta+K` / `Ctrl+K`.
- Support optional `activeNav` and `onNavChange` props with internal fallback state.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx vitest run src/app/__tests__/AppShell.test.tsx`
Expected: PASS (all tests in `AppShell.test.tsx` green).

- [ ] **Step 5: Commit**

```bash
git add web/src/app/AppShell.tsx web/src/app/__tests__/AppShell.test.tsx
git commit -m "feat(web): polish AppShell Primer accessibility, Lucide icons, and command palette"
```

---

### Task 2: Route Search Validation & App Integration in `web/src/routes/index.tsx`

**Files:**
- Modify: `web/src/routes/index.tsx`
- Modify: `web/src/routes/__root.tsx`
- Test: `web/src/routes/__tests__/index.test.tsx`

**Interfaces:**
- Consumes: `createFileRoute` from `@tanstack/react-router`, `App` from `@/App`.
- Produces: `Route` with typed search params `{ tab?: string, shell?: string }` rendering `<App />`.

- [ ] **Step 1: Write the failing test for index route**

Create `web/src/routes/__tests__/index.test.tsx` verifying:
1. `Route` configuration defines `validateSearch` parsing `tab` and `shell` params.
2. The index route component renders `<App />`.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx vitest run src/routes/__tests__/index.test.tsx`
Expected: FAIL because `index.test.tsx` does not find `<App />` rendered by `IndexPage`.

- [ ] **Step 3: Implement index route search validation and App rendering**

In `web/src/routes/index.tsx`:
- Add `validateSearch` to validate `{ tab?: string; shell?: string }`.
- Import and render `<App />` inside `IndexPage`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx vitest run src/routes/__tests__/index.test.tsx`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/routes/index.tsx web/src/routes/__tests__/index.test.tsx
git commit -m "feat(web): connect IndexRoute to App with validated search params"
```

---

### Task 3: Shell Feature Flag & Rail Layout in `web/src/App.tsx`

**Files:**
- Modify: `web/src/App.tsx`
- Test: `web/src/__tests__/App.test.tsx`

**Interfaces:**
- Consumes: `useUIStore` from `@/lib/store`, `AppShell` from `@/app/AppShell`.
- Produces: Adaptive layout in `App` that mounts inside `AppShell` when `newShell === true` or URL search contains `shell=1` / `shell=true`.

- [ ] **Step 1: Write the failing test for AppShell mode in App.test.tsx**

Add test to `web/src/__tests__/App.test.tsx`:
`renders inside AppShell when newShell is active in store, and maps rail navigation to views`

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx vitest run src/__tests__/App.test.tsx -t "renders inside AppShell"`
Expected: FAIL because `App` does not currently render `AppShell`.

- [ ] **Step 3: Implement conditional AppShell wrapping in `web/src/App.tsx`**

- In `web/src/App.tsx`, read `newShell` from `useUIStore` and check URL search parameter `?shell=1` / `?shell=true`.
- When `isShellActive` is true:
  - Map `AppShell` active nav items (`inbox` -> `'home'`, `runs` -> `'agents'`, `reviews` -> `'prs'`, `policies` -> `'rulesets'`, `repos` -> `'code'`).
  - Wrap the main forge content inside `<AppShell activeNav={shellNav} onNavChange={handleShellNavChange}>`.
- When `isShellActive` is false:
  - Render the existing GitHub Primer layout (top `Header`, `<main>`, `footer`) exactly as before.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx vitest run src/__tests__/App.test.tsx`
Expected: PASS (all 10 tests in `App.test.tsx` green).

- [ ] **Step 5: Commit**

```bash
git add web/src/App.tsx web/src/__tests__/App.test.tsx
git commit -m "feat(web): wire AppShell rail layout to App behind newShell feature flag"
```

---

### Task 4: Mount TanStack Router & Query Client in `web/src/main.tsx`

**Files:**
- Modify: `web/src/main.tsx`
- Test: `web/src/__tests__/Router.test.tsx`

**Interfaces:**
- Consumes: `RouterProvider` from `@tanstack/react-router`, `router` from `@/router`, `QueryClientProvider` from `@tanstack/react-query`, `queryClient` from `@/lib/queryClient`.
- Produces: Root application mounting `<QueryClientProvider client={queryClient}><RouterProvider router={router} /></QueryClientProvider>`.

- [ ] **Step 1: Write the failing test for Router root mounting**

Create `web/src/__tests__/Router.test.tsx` verifying:
1. Mounting `RouterProvider` with `router` and `QueryClientProvider` renders the default route `/` successfully without crashing.
2. Deep navigation to `/?tab=issues` or `/?tab=prs` resolves to the corresponding forge view.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web && npx vitest run src/__tests__/Router.test.tsx`
Expected: FAIL because `Router.test.tsx` is newly created.

- [ ] **Step 3: Update `web/src/main.tsx` to mount RouterProvider**

In `web/src/main.tsx`:
```tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from '@tanstack/react-router'
import '@fontsource-variable/mona-sans'
import './index.css'
import { router } from './router'
import { queryClient } from './lib/queryClient'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd web && npx vitest run src/__tests__/Router.test.tsx`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/main.tsx web/src/__tests__/Router.test.tsx
git commit -m "feat(web): mount TanStack RouterProvider and QueryClientProvider in main entry"
```

---

### Task 5: End-to-End Verification, Capability Counts & Production Build

**Files:**
- Verify: `web/src/**`
- Update: `docs/generated/capability_counts.md` (via `gencounts`)

- [ ] **Step 1: Run full frontend Vitest test suite**

Run: `cd web && npm test`
Expected: 100% pass across all test files (≥47 test files, ≥175 tests).

- [ ] **Step 2: Run web production build**

Run: `cd web && npm run build`
Expected: PASS with 0 errors, output in `internal/webui/dist/`.

- [ ] **Step 3: Regenerate and verify capability counts**

Run: `go run ./internal/tools/gencounts && git diff --exit-code docs/generated/capability_counts.md`
If diff detected: `git add docs/generated/capability_counts.md && git commit -m "docs: update capability counts for router and shell tests"`

- [ ] **Step 4: Run Go backend test suite**

Run: `go test ./...`
Expected: PASS across all packages.

- [ ] **Step 5: Verify docs & marketing site**

Run: `cd site && npm run verify`
Expected: PASS across all 28 static pages and links.

- [ ] **Step 6: Commit and merge feature branch**

```bash
git checkout main
git merge --ff-only feat/app-shell-router
git branch -d feat/app-shell-router
```

---

## Self-Review

1. **Spec coverage:** Covers `plan.md` §4.1 (TanStack Router, typed search params, compatibility wrapper) and §4.5 (AppShell global rail, top bar, command palette, theme toggle, feature flag `newShell`, authentic accessibility).
2. **Step scan:** Every step contains exact files, test commands, expected outputs, and commit messages.
3. **Type consistency:** Matches existing types in `web/src/lib/types.ts`, `web/src/lib/store.ts`, `web/src/router.ts`, and `web/src/components/CommandPalette.tsx`.
4. **Review Focus:**
   - Rail navigation accessibility tested with `role="tablist"` / `role="tab"`.
   - Theme toggle button tested for `aria-label`.
   - Command palette trigger tested via button click and keyboard shortcut.
   - Persistence of `?shell=1` and `?tab=` tested.
   - Backward compatibility of `<App />` tested in existing suites.
5. **Proportion:** Concise, focused, high-density plan without bloated transcripts.
