# Refweir Interface & UI Architectural Deep Research Report

**Topic:** Interface, UX/UI, and Frontend Ergonomics for Refweir (Web Forge & Marketing Site)  
**Date:** 2026-11-24  
**Research Mode:** Deep Multi-Agent Research with Adversarial Verification  
**Repositories Audited:**  
- `refweir` (Forge daemon `refweird`, Web UI `web/`, docs `site/`)  
- `refweir-site` (Landing page & visual storytelling site)  

---

## Executive Summary

This deep research investigation evaluates Refweir's human-agent interface and user experience across seven critical dimensions: Git forge review ergonomics, autonomous agent co-work observability, design system tokens and WCAG 2.2 accessibility, high-performance diff rendering in React 19, developer storytelling, mobile and small-viewport usability, and command palette micro-interactions.

While Refweir's frontend demonstrates an exceptionally clean, cohesive visual aesthetic inspired by GitHub and Linear—backed by modern React 19, Tailwind CSS v4, and Radix/Aria primitives—our multi-agent research and adversarial verification uncovered critical UX bottlenecks:
1. **Unbuffered Eager Form Clear in PR Discussions:** `PullRequestsView.tsx` clears comment input state prior to network resolution, destroying user drafts on transient network failure.
2. **Review Blocker Omission in MergeBox:** `MergeBox.tsx` accepts a generic `checksPassed?: boolean` and hardcodes `checksPassed={true}` in `PullRequestsView.tsx`, failing to display blocker warnings when the automated Review Agent (`agentreviewer`) requests changes or flags security vulnerabilities.
3. **Static Lease Observation & Missing Revocation:** `AgentConsoleView.tsx` renders static string timestamps for capability leases without active countdown timers, TTL progress rings, or a 1-click revocation trigger.
4. **Mobile Layout Lock:** On viewports below 768px, `AppShell.tsx` locks the sidebar rail at a fixed 56px (`w-14`) with an inert toggle button, and modal dialogs lack viewport-bounded scrollbars, clipping action buttons when mobile virtual keyboards appear.
5. **Diff Keystroke Lag:** `InlineDiffViewer.tsx` holds line comment composer state in the parent view, forcing all diff hunks, lines, and suggestions across the entire pull request to re-render on every keystroke.
6. **Dark-Mode Badge Contrast:** `StateLabel` in `Badge.tsx` places white text on bright magenta/purple `--chart-4` in dark mode, producing an inaccessible 2.82:1 contrast ratio.

---

## 1. Findings Organized by Sub-Question

### 1.1 Git Forge Review & Code Navigation Ergonomics
- **Multi-File Diff Management:** High-velocity developer platforms (GitHub, GitLab, Linear) provide collapsible file headers and file jump drawers. In `refweir/web`, `InlineDiffViewer.tsx` stacks all diffs vertically without file collapse toggles or "Viewed" checkboxes, requiring long sequential scrolling on large PRs.
- **Diff View Strategy:** Unified single-column diffing is optimal for responsive viewports (<1440px) and mobile devices. Rather than forcing complex side-by-side tables that induce line wrapping, developer tools prioritize unified diffs with expandable context lines.
- **Draft Persistence:** In `InlineDiffViewer.tsx`, draft state is preserved within `try/catch` blocks. However, in `PullRequestsView.tsx:301`, `setNewCommentText('')` is called before awaiting `mutateAsync`, dropping unsubmitted drafts on failure.

### 1.2 Agent-Human Co-work Observability
- **Capability Lease Visualization:** `AgentConsoleView.tsx` lacks live time-to-live (TTL) countdown badges. Best-practice HITL (Human-In-The-Loop) design utilizes a 3-state hybrid badge with `tabular-nums` (Active/Emerald $\rightarrow$ Pulsing Amber <2 min $\rightarrow$ Expired/Slate).
- **Lease Revocation Trigger:** While `internal/store/agent_leases.go` exposes `RevokeAgentLease`, the web UI lacked a visual revocation kill-switch.
- **Review Gating:** `agentreviewer` outputs structured security violations (AWS secrets, private keys, SQLi, unsafe innerHTML/eval). In `MergeBox.tsx`, these violations must explicitly block merge confirmation until resolved or overridden by an administrator.

### 1.3 Accessibility (WCAG 2.2) & Design System Tokens
- **Semantic Text Contrast:** Secondary text token `--text-3: oklch(0.60 0.012 252)` achieves 4.96:1 on canvas but drops to 4.34:1 on `--color-raised` (`#181c22`), narrowly missing WCAG AA (4.5:1). Elevating to `oklch(0.64 0.012 252)` restores compliance across all surfaces.
- **Badge Foreground Contrast:** In `Badge.tsx`, `StateLabel` sets `merged` to `bg-chart-4 text-white`. In dark mode, `--chart-4` has a relative luminance of ~0.32, creating a 2.82:1 failure. Switching to `text-zinc-950 font-medium` achieves AAA compliance (>7:1).
- **Focus Visibility:** `UnderlineNav.tsx` tab triggers defined `outline-none` without `focus-visible:` ring indicators, failing WCAG 2.4.7 (Focus Visible).

### 1.4 High-Performance Code & Diff Rendering in React 19
- **Diff Composer Re-render Isolation:** In `InlineDiffViewer.tsx`, typing into `composerText` triggered full PR re-renders. Isolating line commenting into a dedicated `InlineCommentComposer` component with local state eliminates typing latency.
- **Language Breakdown Memoization:** In `RepoOverview.tsx`, `calculateLanguages()` ran synchronously on every render pass. Wrapping it in `useMemo` eliminates redundant file iterations.
- **CSS Layout Skipping:** Utilizing `content-visibility: auto` with `contain-intrinsic-size: 1px 32px` on diff hunks allows the browser to skip layout and paint calculations for off-screen lines while preserving native `Cmd+F` page searchability.
- **Concurrent Tab Transitions:** Wrapping tab switches in `App.tsx` with `startTransition` enables non-blocking background rendering during navigation.

### 1.5 Marketing UI & Developer Storytelling (`refweir-site`)
- **Interactive Policy Sandbox:** `Policy.tsx` generates dynamic `refweir.toml` configurations based on user selections. Adding a 1-click "Copy refweir.toml" button closes the gap between interactive exploration and local adoption.
- **Demo Tab Ergonomics:** In `workflow-demos.tsx` (`ConnectDemo`), auto-rotating client tabs every 3 seconds caused disruption when reading code blocks. Making tabs user-controlled or pausing on hover preserves developer reading flow.
- **Marquee Verification:** Adversarial cross-checking verified that `.marquee-host:focus-within` is already implemented in `refweir-site/src/index.css:182-185`, properly pausing animations during keyboard navigation.

### 1.6 Mobile & Responsive Viewport Ergonomics
- **Responsive Drawer:** Replacing the locked 56px left rail with a mobile slide-out drawer (or hidden rail on `< md` with hamburger trigger) restores 100% horizontal width on 320px–640px devices.
- **Modal Viewport Bounding:** Adding `max-h-[calc(100dvh-2rem)] overflow-y-auto` to `Dialog.tsx` ensures modal forms and submit buttons remain scrollable and accessible when mobile soft keyboards open.
- **Collapsible File Diffs:** Adding accordion collapse/expand toggles to `BoxHeader` in `InlineDiffViewer.tsx` allows mobile engineers to collapse reviewed files.
- **Mobile Input Zoom Prevention:** Setting inputs to `text-base md:text-xs` prevents iOS Safari from forcibly zooming viewports on input focus.

### 1.7 Command Palettes, Keyboard Shortcuts, & Micro-Interactions
- **Shortcut Hints:** `CommandPalette.tsx` defines `{item.hint && <Kbd ...>}`, but navigation entries had no hints defined. Adding standard hotkey hints (`g c`, `g p`, `g i`, `g a`) reinforces muscle memory.
- **Empty State Depth:** Elevating generic text empty states in `IssuesView.tsx` with contextual illustrations, micro-copy, and action buttons ("Create first issue").
- **Skeleton Placeholders:** Utilizing Refweir's existing `Skeleton.tsx` primitive during data loading prevents abrupt layout shifts.

---

## 2. Adversarial Verification & Resolved Disagreements

During the verification phase (Step 4), several claims were cross-examined and resolved:
1. **Marquee Focus Pause:** An initial researcher claimed `Proof.tsx` failed to pause on keyboard focus. Verifier proved this was **false**; CSS already contained `.marquee-host:focus-within .marquee { animation-play-state: paused; }`.
2. **Comment Draft Loss:** An initial researcher claimed diff comment posting wiped drafts. Verifier proved `InlineDiffViewer.tsx` preserves drafts safely; the unbuffered draft loss was isolated strictly to `PullRequestsView.tsx:301`.
3. **Keyboard Shortcut Conflict (`j`/`k`):** Resolved by establishing contextual modality: `j`/`k` navigates lists (issues, PRs, commits), while `[`/`]` jumps between diff hunks.
4. **Virtualization vs Searchability:** Resolved by rejecting full DOM unmounting virtualization (which destroys `Cmd+F` and screen-reader accessibility) in favor of CSS `content-visibility: auto` paired with isolated React component memoization.

---

## 3. Concrete Implementation Action Plan

### Phase 1: High-Impact Stability & Security Polish (Immediate)
1. **`PullRequestsView.tsx`:** Guard comment draft state; clear input only upon successful mutation response.
2. **`MergeBox.tsx` & `PullRequestsView.tsx`:** Integrate `agentReview` blocker state; display warning banner and disable merge confirmation when `agentreviewer` requests changes.
3. **`Badge.tsx` (`StateLabel`):** Fix dark mode contrast on `merged` state by using `bg-chart-4 text-zinc-950 font-medium`.
4. **`tokens.css`:** Elevate `--text-3` to `oklch(0.64 0.012 252)` to guarantee WCAG AA on raised cards.

### Phase 2: Performance & Observability Enhancements
1. **`AgentConsoleView.tsx`:** Add live countdown timer, progress ring, scope chips, and "Revoke Lease" button for active capability leases.
2. **`InlineDiffViewer.tsx`:** Isolate `InlineCommentComposer` with `React.memo` to eliminate diff re-renders during typing; add file header collapse toggles.
3. **`RepoOverview.tsx`:** Wrap `calculateLanguages()` in `useMemo`.
4. **`App.tsx`:** Wrap tab transitions in React 19 `startTransition`.

### Phase 3: Mobile & Command Palette Polish
1. **`AppShell.tsx` & `Dialog.tsx`:** Make mobile sidebar an off-canvas drawer; bound dialog heights with `max-h-[calc(100dvh-2rem)] overflow-y-auto`.
2. **`CommandPalette.tsx`:** Add shortcut hints (`g c`, `g p`, `g i`, etc.).
3. **`refweir-site/src/sections/Policy.tsx`:** Add 1-click "Copy refweir.toml" button.
4. **`refweir-site/src/sections/workflow-demos.tsx`:** Pause `ConnectDemo` auto-rotation on hover/focus.
