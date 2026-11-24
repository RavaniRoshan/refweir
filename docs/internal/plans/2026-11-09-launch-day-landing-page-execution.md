# Launch-Day Landing Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the state-of-the-art developer landing page enhancements identified in `research/launch-day-landing-page-architecture.md` — interactive merge-gate policy toggles, keyboard-driven terminal controls, scroll-aware performance throttling, and physical instrument craft — ensuring 100% Core Web Vitals and zero AI slop.

**Architecture:** 
Next.js 16 App Router static export (`basePath: /refweir`), Tailwind CSS v4 OKLCH tokens, Motion spring physics, React state-driven interactive sandboxes. All animations are hardware-accelerated and paused when scrolled off-screen via `IntersectionObserver`.

**Tech Stack:** Next.js 16, React 19, TypeScript, Tailwind CSS v4, Motion, Lucide icons, Vitest.

**Spec:** `research/launch-day-landing-page-architecture.md`, `CLAIMS.md`, `AGENTS.md` §2.

## Global Constraints

- `isDemoModeActive() === false` in production; zero mock data.
- All 9 claims-audit needles in `site/scripts/audit-claims.mjs` must pass.
- Press darkens, never scales (`web/src/styles/motion.css` & `AGENTS.md` §2).
- Zero layout shift (CLS 0.0); pre-allocate container bounds and aspect ratios.
- Full reduced-motion accessibility: disable continuous loops when `@media (prefers-reduced-motion: reduce)`.
- Full verification pass (`npm run verify` in `site/`, `go test ./...`, and `npm test` in `web/`) before any commit.

## Review Focus

1. **Claims Audit Needles:** "the self-hosted forge for", "coding agents.", "install.sh | bash", "refweird bootstrap --email admin@example.com", "refs/heads/agent/*", "v0.6.0-beta", "apache-2.0", ":8080", ":2222" must never be modified or removed.
2. **Scroll-Aware Ticker Throttling:** Timer intervals in `ForgeDemo`, `Loop`, and `Product` must pause when the element is off-screen.
3. **Keyboard Accessibility:** Hotkeys (`Space`, `ArrowRight`, `ArrowLeft`, `R`) must not trigger when focus is inside text inputs.
4. **Mobile Layout Integrity:** All Bento cards must stack seamlessly on 390px mobile screens without horizontal overflow.
5. **Dark and Light Theme Fidelity:** Every new component must look crisp and readable in both light and dark modes.

---

## Tasks

### Task 1: Add Scroll-Aware Pause and Keyboard Hotkeys to `ForgeDemo.tsx`

**Files:**
- Modify: `site/components/marketing/ForgeDemo.tsx`

- [ ] **Step 1: Implement keyboard hotkeys and IntersectionObserver in `ForgeDemo.tsx`**
  - Add `Space` to toggle play/pause.
  - Add `R` to replay from step 0.
  - Add `IntersectionObserver` to automatically pause animation when scrolled out of view.
  - Display subtle keyboard hint badge: `[Space] pause · [R] replay`.
- [ ] **Step 2: Verify `ForgeDemo.tsx` locally**
  Run: `cd site && npm run build`
- [ ] **Step 3: Commit changes**
  Commit: `feat(site): add scroll-aware pause and keyboard hotkeys to ForgeDemo`

---

### Task 2: Enhance `MergeGate.tsx` with Interactive Policy Rules & Visual Logic

**Files:**
- Modify: `site/components/marketing/MergeGate.tsx`

- [ ] **Step 1: Implement interactive branch protection rule toggles**
  - Rule 1: "1 Human Approval Required" (default: checked).
  - Rule 2: "Pre-Receive Secret Scan Clean" (default: checked).
  - Rule 3: "Agent Lease Scope refs/heads/agent/*" (default: checked).
  - Rule 4: "Linear History Fast-Forward Only" (default: checked).
  - When all 4 rules pass -> Animated badge displays `UNLOCKED: READY TO MERGE (200 OK)` with emerald glow.
  - When any rule is unchecked -> Animated badge displays `BLOCKED: POLICY REJECTED (GH007)` with rose alert.
- [ ] **Step 2: Verify `MergeGate.tsx` compiles cleanly**
  Run: `cd site && npm run typecheck`
- [ ] **Step 3: Commit changes**
  Commit: `feat(site): enhance MergeGate with interactive policy toggles and live status transitions`

---

### Task 3: Upgrade `Loop.tsx` with Replay Controls and Hotkey Hints

**Files:**
- Modify: `site/components/marketing/Loop.tsx`

- [ ] **Step 1: Add hotkey controls and scroll-aware pause to `Loop.tsx`**
  - Use `IntersectionObserver` to pause auto-advancing when off-screen.
  - Add keyboard navigation support (`ArrowRight`, `ArrowLeft`, `Home`, `End`).
  - Add explicit hotkey navigation hint chip in the step footer.
- [ ] **Step 2: Run site typecheck and build**
  Run: `cd site && npm run verify`
- [ ] **Step 3: Commit changes**
  Commit: `feat(site): add scroll-aware pause and accessible hotkey navigation to Loop`

---

### Task 4: Verify Product Hunt & Show HN Launch Consistency across Marketing Site

**Files:**
- Modify: `site/components/marketing/Product.tsx`
- Modify: `site/components/marketing/StatusFooter.tsx`

- [ ] **Step 1: Verify all claims needles and copy density**
  - Check that all 6 Bento cards have copyable action chips and concrete metrics.
  - Verify footer displays live UTC clock, port status tags (`:8080`, `:2222`), and copyable bootstrap command.
- [ ] **Step 2: Run full site verification**
  Run: `cd site && npm run verify`
- [ ] **Step 3: Commit changes**
  Commit: `feat(site): finalize high-craft launch-ready product and footer sections`

---

### Task 5: Full Product Gate & Verification

- [ ] **Step 1: Run Go test suite (`go test ./...`)**
- [ ] **Step 2: Run Web UI test suite and build (`cd web && npm test && npm run build`)**
- [ ] **Step 3: Run Site verification (`cd site && npm run verify`)**
- [ ] **Step 4: Run README audit (`python3 audit_readme.py README.md`)**
- [ ] **Step 5: Run E2E smoke test (`./scripts/smoke_test_lifecycle.sh`)**
- [ ] **Step 6: Push all commits to `origin/main`**
