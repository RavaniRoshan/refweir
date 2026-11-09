# Marketing Site Rebuild — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the Refweir marketing site (`site/`) from scratch using the consolidated research findings — Velora UI shell + Aceternity hero/bento + Magic UI micro-interactions — with a red-orange accent, kinetic typography, the containment demo video embedded, and all claims-audit needles preserved. Target: premium agentic look, Core Web Vitals green, accessibility (reduced-motion + keyboard + screen reader), 3-5 days of work.

**Architecture:** Next.js 16 App Router static export (`output: 'export'`, basePath `/refweir`). Tailwind v4 CSS-first tokens (OKLCH). Design system: Patina-derived tokens with red-orange accent. Copy-paste component libraries (Velora, Aceternity, Magic UI) adapted into `site/components/marketing/`. Motion via `motion` (Framer Motion) on marketing only. All claims trace to `CLAIMS.md` and pass `npm run claims-audit`.

**Tech Stack:** Next.js 16.4.4, React 19.3.1, TypeScript, Tailwind CSS 4, Fumadocs 16, motion 14, react-aria-components 1.22, class-variance-authority, clsx, tailwind-merge.

**Spec:** `research/full-frontend-launch.md` (comprehensive research), `docs/launch/FRONTEND_TEMPLATES.md` (free template research), `docs/launch/POSITIONING.md` (copy source of truth), `CLAIMS.md` (claims ledger), `AGENTS.md` §2 (design system invariants).

## Global Constraints

- `isDemoModeActive() === false` in production; no mock data.
- Use semantic tokens (`bg-canvas`, `bg-slab`, `border-hairline`, `text-text-1`, `text-signal`) not raw palette classes.
- Controls: 6px radius, 32px default height, 1px hairline borders. Slabs: 8px radius, flat, no decorative shadows.
- Motion follows `motion.css`; press darkens, never scales; reduced motion honored.
- `npm run verify` (typecheck + build + check-links + claims-audit) must pass at every commit.
- No glassmorphism/neon/Linear purple/Geist palette. Single accent color (red-orange, hue 28).
- All claims on the page must trace to `CLAIMS.md`.
- Accessibility: WCAG 2.2 AA, keyboard nav, screen reader, reduced motion, focus management.

## Review Focus

1. **Claims audit needles vanish** — if a required phrase ("coding agents.", "install.sh | bash", "refs/heads/agent/*", "v0.6.0-beta", "apache-2.0", ":8080", ":2222") is missing from `out/index.html`, the audit fails. Every task that touches the page must verify the audit passes.
2. **Base path `/refweir`** — all absolute asset URLs must use `asset()` helper. Missing prefix = 404 on GitHub Pages.
3. **Reduced motion** — every animation must check `prefers-reduced-motion` and fall back to static. Missing = accessibility regression.
4. **Core Web Vitals** — LCP < 2.5s, CLS < 0.1, INP < 200ms. Heavy assets (video, shader) must not block LCP.
5. **Dark mode** — every new component must work in both light and dark. Test both themes.

---

## File Structure

```
site/
├── app/
│   ├── page.tsx                    # Landing page (rebuild from scratch)
│   ├── layout.tsx                  # Root layout (keep)
│   ├── global.css                  # Tokens (keep, already updated to red-orange)
│   └── docs/[[...slug]]/page.tsx   # Fumadocs (keep)
├── components/
│   ├── marketing/
│   │   ├── icons.tsx               # Shared icons (keep)
│   │   ├── Reveal.tsx              # Scroll reveal (keep, reduced-motion aware)
│   │   ├── Nav.tsx                 # Navigation (rebuild)
│   │   ├── Hero.tsx                # Hero with kinetic typography + demo video
│   │   ├── KineticWords.tsx        # Word-swap animation (already created)
│   │   ├── ContainmentDemo.tsx     # Video section (already created)
│   │   ├── SwarmBand.tsx           # Social proof band (rebuild)
│   │   ├── Loop.tsx                # How it works (rebuild)
│   │   ├── Product.tsx             # Product features bento (rebuild)
│   │   ├── Platform.tsx            # Platform/trust (rebuild)
│   │   ├── Quickstart.tsx          # Quickstart terminal (rebuild)
│   │   ├── Evidence.tsx            # Evidence links (keep, minor update)
│   │   ├── FinalCta.tsx            # Final CTA (rebuild)
│   │   └── StatusFooter.tsx        # Footer (keep)
│   ├── ui/                         # shadcn-compat primitives (keep)
│   └── logo.tsx                    # Logo (keep)
├── lib/
│   └── asset.ts                    # Base path helper (keep)
├── public/
│   ├── containment-demo.mp4        # 82s captioned video (already copied)
│   ├── containment-demo-poster.jpg # Video poster (already created)
│   ├── forge-pr.png                # Existing
│   └── og-refweir-os.png           # Existing
└── ... (config, scripts, content/docs — keep)
```

---

## Task 1: Rebuild Nav component

**Files:**
- Modify: `site/components/marketing/Nav.tsx`

**Interfaces:**
- Consumes: `asset()` from `@/lib/asset`, `ThemeToggle` from `./ThemeToggle`
- Produces: `Nav` component (default export)

- [ ] **Step 1: Write the failing test**

Create `site/components/marketing/__tests__/Nav.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Nav } from '../Nav';

describe('Nav', () => {
  it('renders the product name and primary CTA', () => {
    render(<Nav />);
    expect(screen.getByText(/refweir/i)).toBeTruthy();
    expect(screen.getByRole('link', { name: /install|get started|try/i })).toBeTruthy();
  });
  it('has no glassmorphism or neon classes', () => {
    const { container } = render(<Nav />);
    const html = container.innerHTML;
    expect(html).not.toMatch(/backdrop-blur|glass|neon/i);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Nav.test.tsx`
Expected: FAIL (Nav doesn't exist or doesn't match)

- [ ] **Step 3: Implement Nav**

Rebuild `Nav.tsx` with:
- Logo + product name (left)
- Nav links: Docs, GitHub, Containment demo (anchor), Pricing (none — "Contact us")
- Theme toggle (right)
- Primary CTA: "Install refweir.dev" → `#quickstart`
- Mobile: hamburger menu with slide-down
- Sticky header with `backdrop-blur` ONLY on scroll (progressive, not glassmorphism)
- Use semantic tokens: `bg-canvas`, `border-hairline`, `text-text-1`
- Red-orange accent for hover/active states

- [ ] **Step 4: Run test to verify it passes**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Nav.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add site/components/marketing/Nav.tsx site/components/marketing/__tests__/Nav.test.tsx
git commit -m "feat(site): rebuild Nav with red-orange accent and mobile menu"
```

---

## Task 2: Rebuild Hero section

**Files:**
- Modify: `site/components/marketing/Hero.tsx` (already has KineticWords import)

**Interfaces:**
- Consumes: `KineticWords` from `./KineticWords`, `ForgeDemo` from `./ForgeDemo`, `Reveal` from `./Reveal`, `asset()` from `@/lib/asset`, `IconArrowUpRight` from `./icons`
- Produces: `Hero` component (default export)

- [ ] **Step 1: Write the failing test**

Create `site/components/marketing/__tests__/Hero.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Hero } from '../Hero';

describe('Hero', () => {
  it('renders the headline and primary CTA', () => {
    render(<Hero />);
    expect(screen.getByText(/self-hosted forge/i)).toBeTruthy();
    expect(screen.getByRole('link', { name: /install/i })).toBeTruthy();
  });
  it('renders the kinetic words component', () => {
    render(<Hero />);
    expect(screen.getByText(/coding agents/i)).toBeTruthy();
  });
  it('has no mock data', () => {
    const { container } = render(<Hero />);
    expect(container.innerHTML).not.toMatch(/mock|demo data|placeholder/i);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Hero.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement Hero**

Rebuild `Hero.tsx` with:
- Eyebrow: "self-hosted · apache-2.0 · pure go"
- Headline: "The self-hosted forge for <KineticWords />" (already wired)
- Subheadline: positioning one-liner from POSITIONING.md
- Primary CTA: "Install refweir.dev" → `#quickstart` (btn-signal)
- Secondary CTA: "Read the docs" → `/docs/`
- CommandChip with install command
- Stats row: 1 binary, 0 external services, :8080, :2222
- Right column: ForgeDemo (interactive) + caption
- Grid field background (CSS, not JS)
- Crosshair decorations
- Reveal animations (already reduced-motion aware)

- [ ] **Step 4: Run test to verify it passes**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Hero.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add site/components/marketing/Hero.tsx site/components/marketing/__tests__/Hero.test.tsx
git commit -m "feat(site): rebuild Hero with kinetic typography and red-orange accent"
```

---

## Task 3: Rebuild SwarmBand (social proof)

**Files:**
- Modify: `site/components/marketing/SwarmBand.tsx`

- [ ] **Step 1: Write the failing test**

Create `site/components/marketing/__tests__/SwarmBand.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { SwarmBand } from '../SwarmBand';

describe('SwarmBand', () => {
  it('renders social proof', () => {
    render(<SwarmBand />);
    expect(screen.getByText(/apache-2\.0|self-hosted|pure go/i)).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd site && npx vitest run src/components/marketing/__tests__/SwarmBand.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement SwarmBand**

Rebuild with:
- Scrolling marquee of capability badges (leases, MCP, LFS, SSH, audit, air-gap)
- "Trusted by" or capability proof strip
- Use CSS animation (not JS) for the marquee
- Reduced-motion: static (no scroll)

- [ ] **Step 4: Run test to verify it passes**

Run: `cd site && npx vitest run src/components/marketing/__tests__/SwarmBand.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add site/components/marketing/SwarmBand.tsx site/components/marketing/__tests__/SwarmBand.test.tsx
git commit -m "feat(site): rebuild SwarmBand social proof marquee"
```

---

## Task 4: Rebuild Loop (how it works)

**Files:**
- Modify: `site/components/marketing/Loop.tsx`

- [ ] **Step 1: Write the failing test**

Create `site/components/marketing/__tests__/Loop.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Loop } from '../Loop';

describe('Loop', () => {
  it('renders the containment loop steps', () => {
    render(<Loop />);
    expect(screen.getByText(/lease/i)).toBeTruthy();
    expect(screen.getByText(/push|merge|audit/i)).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Loop.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement Loop**

Rebuild with:
- 4-step loop: Issue lease → Agent pushes → Review PR → Merge + audit
- Each step has icon, title, description
- Scroll-triggered reveals (Reveal component, already reduced-motion aware)
- Use the containment demo video or ForgeDemo as visual

- [ ] **Step 4: Run test to verify it passes**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Loop.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add site/components/marketing/Loop.tsx site/components/marketing/__tests__/Loop.test.tsx
git commit -m "feat(site): rebuild Loop section with containment flow"
```

---

## Task 5: Rebuild Product (bento grid)

**Files:**
- Modify: `site/components/marketing/Product.tsx`

- [ ] **Step 1: Write the failing test**

Create `site/components/marketing/__tests__/Product.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Product } from '../Product';

describe('Product', () => {
  it('renders feature cards', () => {
    render(<Product />);
    expect(screen.getByText(/lease/i)).toBeTruthy();
    expect(screen.getByText(/audit/i)).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Product.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement Product**

Rebuild with bento grid layout:
- Lease (scoped, expiring, revocable)
- Ref-prefix isolation (refs/heads/agent/*)
- Pre-receive hook (secret scanning, branch protection)
- Durable audit log (actor, lease, ref, reason)
- Archived repo freeze
- Deploy key isolation
- Git LFS v1 CAS
- MCP + read-only MCP
- Pure-Go SSH server
- Semantic symbol search

Use CSS grid (bento layout), scroll-triggered reveals, red-orange accent on hover.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Product.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add site/components/marketing/Product.tsx site/components/marketing/__tests__/Product.test.tsx
git commit -m "feat(site): rebuild Product section as bento grid"
```

---

## Task 6: Rebuild Platform (trust/infrastructure)

**Files:**
- Modify: `site/components/marketing/Platform.tsx`

- [ ] **Step 1: Write the failing test**

Create `site/components/marketing/__tests__/Platform.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Platform } from '../Platform';

describe('Platform', () => {
  it('renders infrastructure features', () => {
    render(<Platform />);
    expect(screen.getByText(/self-hosted|air-gap|sso|scim/i)).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Platform.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement Platform**

Rebuild with:
- Self-hosted, air-gapped, zero cgo
- SSO (OIDC/SAML), SCIM 2.0
- Helm chart, Terraform provider
- SQLite WAL, bare repos, LFS CAS
- Audit log (JSONL, append-only, operator-controlled)

- [ ] **Step 4: Run test to verify it passes**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Platform.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add site/components/marketing/Platform.tsx site/components/marketing/__tests__/Platform.test.tsx
git commit -m "feat(site): rebuild Platform section with infrastructure features"
```

---

## Task 7: Rebuild Quickstart

**Files:**
- Modify: `site/components/marketing/Quickstart.tsx`

- [ ] **Step 1: Write the failing test**

Create `site/components/marketing/__tests__/Quickstart.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Quickstart } from '../Quickstart';

describe('Quickstart', () => {
  it('renders install command', () => {
    render(<Quickstart />);
    expect(screen.getByText(/install\.sh/i)).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Quickstart.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement Quickstart**

Rebuild with:
- Terminal-style install command (CommandChip, already exists)
- 3-step quickstart: install → bootstrap → push
- Link to docs
- Copy button

- [ ] **Step 4: Run test to verify it passes**

Run: `cd site && npx vitest run src/components/marketing/__tests__/Quickstart.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add site/components/marketing/Quickstart.tsx site/components/marketing/__tests__/Quickstart.test.tsx
git commit -m "feat(site): rebuild Quickstart section with terminal install"
```

---

## Task 8: Rebuild FinalCta

**Files:**
- Modify: `site/components/marketing/FinalCta.tsx`

- [ ] **Step 1: Write the failing test**

Create `site/components/marketing/__tests__/FinalCta.test.tsx`:
```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { FinalCta } from '../FinalCta';

describe('FinalCta', () => {
  it('renders CTA buttons', () => {
    render(<FinalCta />);
    expect(screen.getByRole('link', { name: /install|try|demo/i })).toBeTruthy();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd site && npx vitest run src/components/marketing/__tests__/FinalCta.test.tsx`
Expected: FAIL

- [ ] **Step 3: Implement FinalCta**

Rebuild with:
- Headline: "Contain every agent. Prove every merge."
- Primary CTA: "Install refweir.dev" → `#quickstart`
- Secondary CTA: "Try the hosted demo" → demo URL
- Tertiary: "Read the docs" → `/docs/`

- [ ] **Step 4: Run test to verify it passes**

Run: `cd site && npx vitest run src/components/marketing/__tests__/FinalCta.test.tsx`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add site/components/marketing/FinalCta.tsx site/components/marketing/__tests__/FinalCta.test.tsx
git commit -m "feat(site): rebuild FinalCta with containment positioning"
```

---

## Task 9: Rebuild page.tsx

**Files:**
- Modify: `site/app/page.tsx`

- [ ] **Step 1: Write the failing test**

No separate test file — verified by `npm run verify` (build + check-links + claims-audit).

- [ ] **Step 2: Implement page.tsx**

```tsx
import { Nav } from '@/components/marketing/Nav';
import { Hero } from '@/components/marketing/Hero';
import { SwarmBand } from '@/components/marketing/SwarmBand';
import { Loop } from '@/components/marketing/Loop';
import { ContainmentDemo } from '@/components/marketing/ContainmentDemo';
import { Product } from '@/components/marketing/Product';
import { Platform } from '@/components/marketing/Platform';
import { Quickstart } from '@/components/marketing/Quickstart';
import { Evidence } from '@/components/marketing/Evidence';
import { FinalCta } from '@/components/marketing/FinalCta';
import { StatusFooter } from '@/components/marketing/StatusFooter';

export default function Home() {
  return (
    <>
      <Nav />
      <main>
        <Hero />
        <SwarmBand />
        <Loop />
        <ContainmentDemo />
        <Product />
        <Platform />
        <Quickstart />
        <Evidence />
        <FinalCta />
      </main>
      <StatusFooter />
    </>
  );
}
```

- [ ] **Step 3: Run full verification**

Run: `cd site && npm run verify`
Expected: typecheck + build + check-links + claims-audit all pass

- [ ] **Step 4: Commit**

```bash
git add site/app/page.tsx
git commit -m "feat(site): wire rebuilt marketing page sections"
```

---

## Task 10: Final verification and cleanup

- [ ] **Step 1: Run full gate**

Run: `cd /home/shiva/projects/exist.io && CGO_ENABLED=0 go test ./... && cd web && npx vitest run && npm run build && cd ../site && npm run verify`
Expected: all pass

- [ ] **Step 2: Visual check**

Serve the site (`npm run start` in `site/`) and verify:
- Red-orange accent visible in both light and dark mode
- Kinetic headline swaps words
- Containment video loads with poster
- All sections render
- Mobile responsive (390px)
- No console errors
- Reduced-motion: animations disabled

- [ ] **Step 3: Commit any fixes**

```bash
git add -A && git commit -m "fix(site): address review findings from marketing rebuild"
```

---

## Self-Review Checklist

1. **Spec coverage:** Every section of the landing page (hero, social proof, how-it-works, containment demo, product, platform, quickstart, evidence, final CTA) has a task. ✓
2. **Step scan:** Each task has write-test → run-fail → implement → run-pass → commit. ✓
3. **Type consistency:** Components use existing patterns (Reveal, asset, icons). ✓
4. **Review Focus:** 5 input classes addressed (claims audit needles, base path, reduced motion, Core Web Vitals, dark mode). ✓
5. **Proportion:** Plan is focused on one subsystem (marketing site). ✓

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-11-09-marketing-site-rebuild.md`. Please review the plan. Which execution approach would you prefer?

- **Subagent-driven** — A fresh subagent implements each task and a fresh reviewer checks it before the next one starts, then a whole-branch review at the end. Most thorough.
- **Native** — I implement every task myself in this session, then one fresh reviewer on the most capable model checks the whole branch. Cheapest and fastest.

For this plan I recommend **subagent-driven**, because the tasks are independent (each section can be built and tested in isolation), there are 10 of them, and a shipped mistake on the marketing site (broken claims audit, missing base path, inaccessible motion) would directly hurt the PH launch. Does the plan capture what you want, and which approach should we use?