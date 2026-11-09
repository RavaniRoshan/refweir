# Frontend templates & motion — lowest effort, highest ROI

**Status:** Research follow-up to `FRONTEND_RESEARCH.md`
**Owner:** Frontend / Design
**Created:** 2026-11-09
**Goal:** Ship a top-tier, motion-driven, agentic-looking landing page with the
least possible engineering effort, using free templates and copy-paste
component libraries instead of designing from scratch.

---

## 1. The "agentic" look in 2026 (what top designers actually ship)

Research sources: toimi.pro "Top 10 SaaS website designs 2026", Linear's
2025–2026 homepage, Vercel, Raycast, Spell UI's "25 Best SaaS Landing Pages".

The defining traits of a 2026 top-tier B2B SaaS landing page (Linear is the
most-copied example):

1. **Dark-as-default with one neon accent.** ~75% of top design-led SaaS sites.
   Linear-purple, Raycast-red, Mercury-lime, Cursor-cyan. Pick ONE accent and
   use it with restraint. Our OKLCH tokens already support this; choose a single
   `--signal` accent and audit every component for stray colors.
2. **Kinetic / motion-driven typography.** Headlines that swap words, stretch,
   and respond to scroll. Linear swaps "teams" ↔ "agents". This is the single
   biggest "premium" signal. Implement with CSS `@keyframes` word-swap or
   Motion's `AnimatePresence`.
3. **Shader-based WebGL backgrounds.** Now commodity (Framer ships a Shaders
   library, Vercel ships v0 shader hero templates). One shader in the hero,
   nothing else. Do not overdo it.
4. **Product is the demo.** The hero shows the actual product UI animating
   autonomously (Linear shows an agent picking up a task, commenting, opening a
   PR). Our containment demo video IS this asset — embed it in the hero.
5. **Bento grids** for feature sections (Linear, Vercel, Notion).
6. **Scroll-triggered reveals** with `animation-timeline: view()` (Josh
   Comeau's 2026 guide) or Motion's `whileInView`. Progressive disclosure.
7. **Micro-interactions with purpose.** Hover reveals, pulsing CTA on viewport
   entry, animated counters. Not decoration.
8. **Single CTA repeated.** One dominant action per viewport.
9. **Named testimonials with photo/role/company.** Anonymous testimonials
   convert worse than none.
10. **Fast.** Core Web Vitals: LCP < 2.5 s, CLS < 0.1, INP < 200 ms. Heavy
    animation must be lazy and interruptible.

---

## 2. Free template shortlist (lowest effort, highest ROI)

Researched from wrap pixel, aniq-ui, adminlte, colorlib, aceternity guides,
prebuiltui, ui.aceternity. All MIT or free for commercial use unless noted.

### Tier 1 — copy-paste component libraries (free, no install, shadcn-compatible)

These are the highest-ROI tools: you copy a component into your repo and it
works. No dependency, no version lock. Our `web/src/components/ui/` compat
layer and `site/` can absorb them directly.

| Library | What it is | Free? | Best for | URL |
|---|---|---|---|---|
| **Aceternity UI** | 118+ free animated components + 183 blocks + 17 Next.js templates. Tailwind + Motion. The "shadcn/ui for magic effects." | Yes (free tier) | Cinematic hero, shaders, bento, 3D cards, background beams, text effects | ui.aceternity.com |
| **Magic UI** | 150+ free animated components. shadcn/ui companion. "Design Engineer" craft. | Yes (MIT) | Polished micro-interactions, marketing animations, animated beams, grids | magicui.design |
| **21st.dev** | Registry/aggregator of animated components (Aceternity, Magic UI, etc.) with live previews | Yes | Browsing + copy-paste; AI-agent install (MCP) | 21st.dev |
| **Animata** | 155+ free MIT components, copy-and-own, a11y + reduced-motion included | Yes (MIT) | Product UI polish you own | animata.design |
| **PrebuiltUI** | Free Next.js/React/HTML templates (Agentix = dark AI-agent SaaS template) | Yes | Whole-page starting point | prebuiltui.com |
| **MVPBlocks** | Free copy-paste blocks/sections/templates, Next.js + Tailwind + Framer Motion, v0-compatible | Yes | Fast section assembly | (via shadcn.io) |

**Recommendation:** Use **Aceternity UI** for the marketing hero + bento +
motion, **Magic UI** for micro-interactions, and **Animata** for any product-UI
polish. All three are copy-paste, Tailwind + Motion, shadcn-compatible — they
drop into our existing stack with zero new dependencies.

### Tier 2 — free full-page templates (Next.js 16 + Tailwind v4 + Motion)

| Template | Stack | License | Best for |
|---|---|---|---|
| **Cruip "Simple"** (Simple Light / Simple) | Next.js + Tailwind v4 + animated hero, features, testimonials, newsletter | Free (GPL) / 4.6k stars | Most polished free SaaS landing. Cruip design sensibility. |
| **Velora UI** | Next.js 16 + React 19 + Tailwind v4 + shadcn/ui + Motion | MIT (free, Pro tier announced) | Whole site: home, pricing, blog (MDX), about, contact, login, signup, changelog, 404. 64 animated components. |
| **Launch UI** | Next.js 16 + React 19 + Tailwind v4 + TS + Framer Motion 12 + GSAP + WebGL particles | Free | Glass effects, dashboard mockups, animated numbers, pricing cards, world map, RTL/i18n. |
| **ShipFast**-style starters (various) | Next.js + Tailwind + Framer Motion | Free (MIT, many) | Opinionated full-stack starter with auth/payments. |
| **Aceternity "Nodus Agent" template** | Next.js + Tailwind + Framer Motion | Free (aceternity templates) | AI-agent / agentic SaaS template — bento grids, micro-interactions. Closest to our positioning. |
| **Aceternity "Agenforce" template** | Next.js + Tailwind + Framer Motion | Free | Agentic marketing template. |

**Recommendation for our marketing site (`site/`):** Start from **Velora UI**
(MIT, whole-site, Tailwind v4, Motion, MDX blog) for the shell, then layer
**Aceternity's "Nodus Agent"** hero/bento sections on top. This gives us a
complete, animated, dark-mode marketing site in days instead of weeks, while
keeping our own content and tokens.

### Tier 3 — paid (only if we want to skip even more work)

- **Magic UI Pro**: $49/template or $199 lifetime All-Access. 9 templates
  incl. "AI Agent", "Dev Tool", "Codeforge". Next.js + Tailwind + Framer Motion.
- **Tailwind Plus (Catalyst)**: paid, but the official Tailwind application UI
  kit; integrates perfectly with Tailwind v4.
- **Framer templates**: $49–$59 each (MNC, Solarium, etc.) — only if we
  switch the marketing site to Framer (not recommended; keep Next.js).

**Recommendation:** Stay free. Velora + Aceternity + Magic UI gives 90% of the
paid look at $0. Revisit paid only if the free templates block a specific
launch-critical section.

---

## 3. Motion stack (research findings)

| Layer | Tool | Use |
|---|---|---|
| Marketing hero / scroll storytelling | **Motion (Framer Motion)** + CSS | Hero word-swap, scroll reveals, bento entrance, shader background |
| Scroll-driven reveals | Native CSS `animation-timeline: view()` (Josh Comeau 2026 guide) | Lightweight, no JS, hardware-accelerated; fallback for older browsers |
| Micro-interactions | CSS transitions (`transition-colors`, `transform`) | Hover, focus, press — no JS runtime |
| Smooth scroll | **Lenis** (or native) | Premium feel, pairs with scroll-driven animations |
| Shader backgrounds | **WebGL / Framer Shaders** or a single canvas shader | One hero shader only |
| App shell | CSS only | No Motion in app bundle (bundle size + jank on low-end self-hosted) |

Rules (from research + AGENTS.md motion.css):
- Only animate `transform` and `opacity`. Never layout props.
- Honor `prefers-reduced-motion`.
- Lazy-load below-the-fold motion.
- Keep entrance animations < 300 ms.
- Interruptible: if user scrolls fast, jump to end state (no catch-up).
- Core Web Vitals: LCP < 2.5 s, CLS < 0.1, INP < 200 ms. The demo video
  must not block LCP (poster + lazy).

---

## 4. Lowest-effort, highest-ROI plan for our marketing site

Estimated effort: **3–5 days** instead of 3–4 weeks of custom design.

1. **Day 1 — Shell.** Fork Velora UI (MIT) into `site/`. Replace its tokens
   with our OKLCH tokens from `web/src/styles/tokens.css` (port the CSS vars).
   Swap the accent to our single neon accent. Dark mode default.
2. **Day 1–2 — Hero.** Use Aceternity's "hero with fluid gradient" / "hero with
   logo ticker" / "hero section with hub" component. Embed the 82 s containment
   demo video as the hero product surface (poster image + lazy video + captions).
   Add kinetic typography (word swap "teams" ↔ "agents") via Motion.
3. **Day 2 — Bento.** Use Aceternity's "feature grid with motion and tabs" /
   "feature grid with bento" for the containment proof points (lease, ref scope,
   audit, archived freeze, deploy keys, benchmarks).
4. **Day 2 — Micro-interactions.** Magic UI animated beams, number counters,
   logo cloud (marquee) for social proof.
5. **Day 3 — Sections.** Testimonials (named, with photo/role/company),
   pricing placeholder (no prices — "Contact us" / "Self-host"), FAQ accordion,
   final CTA.
6. **Day 3 — Motion polish.** Lenis smooth scroll, scroll-triggered reveals
   via `animation-timeline: view()`, one hero shader (optional, test perf).
7. **Day 4 — QA.** Run smollaunch landing-page analyzer, Lighthouse, mobile
   check at 390 px / 1440 px, Core Web Vitals, reduced-motion check.
8. **Day 5 — Integrate.** Wire the hosted demo embed ("Try an agent run" CTA
   opens the live demo in a new tab or iframe). Keep `npm run build` (site)
   green.

What we keep from our existing `site/`:
- Next.js static export to GitHub Pages (no infra change).
- Fumadocs docs under `/docs`.
- The claims audit (`npm run claims-audit`) — every new claim must pass it.
- Our positioning one-pager as the copy source of truth.

What we replace:
- The current minimal hero/marketing sections → Aceternity/Magic UI components.
- Any custom animation code → library components (less code, more polish).

---

## 5. Specific free components to grab first (copy-paste)

From Aceternity UI (ui.aceternity.com):
- Hero with fluid gradient + planner / hero with logo ticker + workspace
- Background beams / background cubes grid
- Bento grid (feature grid with motion + tabs)
- 3D card effect / wobble card / apple cards carousel
- Floating navbar (sticky, hides on scroll)
- Text generate effect / typewriter effect (for kinetic headline)
- Infinite moving cards (logo cloud / testimonials)
- Animated modal (for demo embed)

From Magic UI (magicui.design):
- Animated beam (for "agent → forge → audit" flow diagram)
- Number ticker (for benchmark numbers: 4 ms, 2.94 s)
- Marquee (logo cloud)
- Border beam (card highlight)
- Shimmer button (primary CTA)
- Globe (optional, for "self-hosted anywhere" visual)

From Animata (animata.design):
- Reduced-motion-aware versions of above
- Copy-and-own source for anything we need to customize heavily

---

## 6. What we should NOT spend time on

- Custom shader work beyond one hero shader (use Framer's Shaders or skip).
- Building our own component library (use shadcn + Aceternity/Magic UI).
- Rebuilding the app shell (strangler fig; keep existing views until parity).
- Paid templates (free tier covers 90%).
- A separate design tool pipeline (Figma → code). Use the templates' Figma
  files only for reference; ship from code.

---

## 7. Sources

- Aceternity UI (ui.aceternity.com): 118+ free components, 183 blocks,
  17 templates, Tailwind + Motion, shadcn-compatible.
- Magic UI (magicui.design): 150+ free animated components, shadcn companion.
- Animata (animata.design): 155+ free MIT, reduced-motion-first.
- PrebuiltUI (prebuiltui.com): free Next.js/React/HTML templates incl.
  "Agentix" dark AI-agent SaaS template.
- MVPBlocks: free copy-paste blocks, v0-compatible.
- 21st.dev: registry/aggregator with live previews + MCP install.
- Cruip Simple (4.6k stars): most polished free Next.js SaaS landing.
- Velora UI: MIT, whole-site (home/pricing/blog/about/contact/login/signup/
  changelog/404), 64 animated components, Tailwind v4.
- Launch UI: Next.js 16 + Tailwind v4 + Framer Motion 12 + GSAP + WebGL.
- Aceternity guides: best-nextjs-templates-2026, best-ai-saas-templates.
- toimi.pro "Top 10 SaaS Website Designs 2026" (Linear kinetic typography,
  dark-as-default + neon accent, shader backgrounds).
- SaaSFrame "10 SaaS Landing Page Trends 2026" (interactive demos,
  micro-animations, bento, personalization).
- Moburst, tncflow (Framer trends 2026).
- Josh Comeau "Scroll-Driven Animations" (animation-timeline API).
- spell.sh "25 Best SaaS Landing Pages" (Raycast, Resend, etc.).
- plinthstudio.dev SaaS landing best practices (logo bar, named testimonials,
  old-way/new-way, alternative positioning).
- 2POINT Agency "Website Animation Complete Guide 2026".
- smollaunch.com PH guide 2026 (landing matches PH copy, mobile gallery,
  captioned video).
- Existing: AGENTS.md §2 (design system), FRONTEND_RESEARCH.md, plan.md §3.

---

## 8. Next actions

1. Spike Velora UI fork in a branch; port our OKLCH tokens; verify
   `npm run build` + claims audit still pass.
2. Copy 5 Aceternity components into `site/components/marketing/` and render
   them in a draft page. Measure LCP/CLS/INP.
3. Embed the containment demo video in the hero (poster + lazy + captions).
4. Add kinetic typography headline ("Contain every agent. Prove every merge.").
5. Run smollaunch landing-page analyzer + Lighthouse before launch.
6. Keep `npm run verify` (site) green at every commit.
