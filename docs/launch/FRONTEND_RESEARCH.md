# Frontend research — best-in-class for a Product Hunt launch

**Status:** Research input for Phase 3 (agents-first frontend redesign)
**Owner:** Frontend / Design
**Created:** 2026-11-09
**Companion:** [`plan.md`](../../plan.md) §3, [`AGENTS.md`](../../AGENTS.md) §2/§4

This document captures the research the launch plan asked for: what the best
launching products in 2026 actually ship for their marketing site and their app
shell, which libraries they use, and how the top Product Hunt launches present
themselves. It is a decision aid, not a spec. Every recommendation must still
pass the invariants in `AGENTS.md` §2 (zero mock data, design tokens, zero-CGO
is irrelevant here, test integrity).

---

## 1. Executive summary

1. **Split the surfaces.** The highest-performing 2026 launches (Linear, Notion,
   Framer, Cursor, Lovable, Raycast, Vercel) run a **marketing site** and an
   **app shell** as two different frontends with two different stacks. The
   marketing site is content-heavy and SSR/SSG; the app shell is a typed SPA.
   Do not try to make one Vite SPA serve both.
2. **Keep our base.** The plan's decision to keep **Patina + React Aria +
   Tailwind v4 + Vite 8** is correct and aligned with 2026 best practice. The
   consensus 2026 React stack is exactly this: Vite 8 (Rolldown) + TypeScript +
   Tailwind v4 + shadcn/ui-style copy-paste components. We are not behind.
3. **Add TanStack Router + TanStack Query** for the app shell (typed routes,
   typed search params, server state). This is the 2026 consensus for SPAs
   behind auth, and it is what the plan already calls for.
4. **Keep React Aria as the primitive layer, not Radix/shadcn.** We already
   invested in React Aria. The research says React Aria is the
   best-in-class choice for max accessibility and i18n, which is exactly what
   a regulated/air-gapped buyer needs. Do not switch to Radix or Base UI mid-
   project; that would throw away tested work for marginal gain.
5. **Motion: Motion (formerly Framer Motion) for marketing, CSS for the app.**
   Research: heavy JS animation hurts Core Web Vitals and is the #1 cause of
   "looks nice but feels slow" PH feedback. Use CSS transitions + `transform`
   for the app shell; reserve Motion (Framer Motion) for the marketing hero and
   scroll storytelling, and respect `prefers-reduced-motion`.
6. **Marketing site: Next.js (or Astro) + MDX, deployed to Vercel/Cloudflare.**
   The research is unambiguous that content-heavy marketing sites ship on
   Next.js (App Router, RSC, image optimization) or Astro. Our existing `site/`
   is Next.js static export to GitHub Pages — keep it, but move the marketing
   hero to a richer animation and add an interactive demo embed.
7. **Do not rebuild the app shell in one shot.** Strangler fig: new routes
   render in the new shell behind a flag; old views stay until parity. This is
   what the plan already says. The research confirms it is how Linear/Notion
   migrations were done.

---

## 2. Marketing site vs. app shell

| Concern | Marketing site (`site/`) | App shell (`web/`) |
|---|---|---|
| Audience | Visitors, evaluators, PH voters | Logged-in users, agents |
| Rendering | SSG/SSR, SEO-critical | SPA behind auth |
| Stack | Next.js (App Router) or Astro + MDX | Vite 8 + React 19 + TS |
| Styling | Tailwind v4 + design tokens | Tailwind v4 + design tokens (same) |
| Motion | Motion (Framer Motion) + scroll triggers | CSS transitions only |
| Data | Static + a few live widgets | TanStack Query + SSE |
| Deploy | Vercel / Cloudflare Pages / GitHub Pages | Static bundle behind the Go daemon |
| Bundle budget | < 200 KB first load | Route-level code splitting, lazy diff/editor |

The single most common mistake in PH launches is one app serving both: the
marketing page inherits the app's bundle, or the app inherits marketing
animations. Keep them separate. Our repo already does (`site/` vs `web/`).

---

## 3. Landing page patterns that convert in 2026 (research findings)

Sources: SaaSFrame "10 SaaS landing page trends 2026", Moburst, Framer trend
report, Product Hunt launch guide 2026.

1. **Hero = product, not illustration.** The top launches put a live or
   animated product surface in the hero, not a stock illustration. Linear,
   Vercel, Raycast, Cursor all do this. Our containment demo (82 s, captioned)
   is the hero asset; embed it above the fold with a "Run it" CTA.
2. **Interactive product demo in the hero.** SaaSFrame: "screenshots are
   evolving into interactive components; embedded product previews and guided
   tours right in the hero." We already have the hosted demo with a one-click
   agent run — surface it on the landing page, not just at `/`.
3. **Micro-animations with purpose.** Motion that demonstrates functionality
   (hover reveals, scroll progress, animated counters, pulsing CTA on viewport
   entry). Decorative motion is out; purposeful motion is expected. Keep it
   lightweight (CSS + `transform`), lazy-load heavy assets, respect
   `prefers-reduced-motion`, and watch Core Web Vitals (no layout shift).
4. **Bento grids for feature sections.** The dominant 2026 layout for feature
   lists is the bento grid (Linear, Vercel, Notion). Use it for the containment
   proof points: lease, ref scope, audit, archived freeze, deploy keys.
5. **Social proof with numbers, not logos-only.** "3 quotes and one mini-case
   study with a number" (plan §2.2). Numbers beat logos. Show the containment
   benchmark (4 ms lease p50, 2.94 s full loop) as proof.
6. **One primary CTA, repeated.** "Try the demo" / "Self-host" / "Read the
   threat model" — but one dominant action per viewport. Our banner already does
   this; the landing page should too.
7. **Mobile-first, fast.** PH voters are on phones. Gallery images must be
   readable on mobile; video must load fast with captions (we have captions).
8. **Tagline under 60 characters, plain language.** Already in our positioning
   ("Contain every agent. Prove every merge.").
9. **Landing page matches PH copy exactly.** SmolLaunch checklist: "Landing
   page matches the Product Hunt copy." Keep one source of truth (POSITIONING.md).
10. **Demo video: 60–90 s, captions burned in, loads fast.** We have this.

---

## 4. Component library and design system decisions

Research consensus (2026):
- **shadcn/ui** is the top pick for new React + Tailwind projects (copy-paste,
  zero runtime, Radix primitives).
- **React Aria** is the best choice for max accessibility + i18n, and for
  regulated/enterprise buyers.
- **Base UI** (from Radix/MUI/Floating UI creators) is the new foundational
  primitive layer; shadcn/ui is moving to it.
- **HeroUI (NextUI v3)** is the best "ready-made + Tailwind + React Aria" option.
- **MUI** still wins for data-heavy enterprise (data grids).
- **Motion (Framer Motion)** is the animation standard for marketing.

**Decision for Refweir (no change to existing plan):**
- Keep **React Aria** as the primitive layer. We already built on it; it is the
  best-in-class for accessibility and i18n, which is our differentiator for
  regulated buyers. Switching to Radix/Base UI mid-project loses tested work.
- Keep **shadcn-compatible copy-paste patterns** in `web/src/components/ui/`
  (the compatibility layer already does this). New components should follow the
  shadcn convention (copy into repo, Tailwind classes, Radix-style APIs) so we
  can adopt community components without a dependency.
- Keep **Tailwind v4 + OKLCH tokens** (`tokens.css`, `theme.css`,
  `typography.css`). This is exactly what 2026 best practice recommends.
- Add **Motion (Framer Motion)** only to the marketing site (`site/`), not the
  app shell. The app shell uses CSS transitions.
- Add **TanStack Router + TanStack Query** for the app shell (typed routes,
  typed search params, server state, SSE).
- Add **Zustand** for UI state (already planned).
- Add **Pierre / CodeMirror 6** for diffs (already planned; spike on 10k/50k
  line diffs first).
- Keep **Vitest + Testing Library + MSW + Playwright** (already planned).

---

## 5. Animation and motion rules

- Marketing site: Motion (Framer Motion) for hero, scroll storytelling, bento
  grid reveals. Keep animations < 300 ms, use `transform`/`opacity` only,
  lazy-load below the fold.
- App shell: CSS transitions only (`transition-colors`, `transition-opacity`).
  No JS animation runtime in the app bundle (keeps bundle small, avoids jank
  on low-end machines, which matters for self-hosted/air-gapped buyers).
- Always honor `prefers-reduced-motion` (already in motion.css).
- Never animate layout properties (width/height/top/left) — use transform.
- Core Web Vitals budget: LCP < 2.5 s, CLS < 0.1, INP < 200 ms. The demo
  video must not block LCP (poster image + lazy video).

---

## 6. Top Product Hunt launch patterns (research findings)

Sources: smollaunch.com PH guide 2026, saascity.io, dev.to, medium (Mohit).

1. **30-day prep, 3 phases.** Foundation/narrative (week 1), audience growth
   (weeks 2–3), assets/QA/logistics (week 4). Skipping any phase reduces
   top-5 chances.
2. **Launch Tue–Thu, 12:01 AM PST.** Target 200+ upvotes and 30+ comments in
   the first 6 hours (algorithm sets initial ranking).
3. **Maker profile matters.** Complete profile, photo, bio, previous work.
   First thing upvoters see.
4. **Gallery images readable on mobile** (research confirms).
5. **Demo video loads fast with captions** (we have this).
6. **Landing page matches PH copy exactly** (one source of truth:
   POSITIONING.md).
7. **Respond to every comment within 30 minutes** on launch day. 12–16 hour
   marathon.
8. **Pre-built momentum.** Start building audience 30 days before. Our content
   engine (2 posts/week) and design-partner outreach are this.
9. **Don't treat launch as a single event.** Spread across multiple platforms
   over 3 weeks (PH, then Show HN, Uneed, Fazier, Peerlist). PH link is
   nofollow; the others give dofollow backlinks.
10. **Show HN for technical audience.** Highest variance (1,906 points or 3
    points). For a self-hosted forge, Show HN is a strong fit — technical,
    blunt, but high ceiling. Plan it as a separate launch 1–2 weeks after PH.
11. **Badges and embeds.** Startup Fame, Twelve Tools, TinyLaunch give
    dofollow links for embedding a badge (~$0). Worth ~$115 total for
    guaranteed visibility across a week.
12. **Landing page analyzer / QA checklist.** Run smollaunch's analyzer before
    launch. Add to our pre-launch checklist.

---

## 7. Recommended architecture for Refweir (Phase 3)

Keep the strangler-fig plan from AGENTS.md §4/plan.md §3. Concretely:

```
site/                          # Next.js static export → GitHub Pages (marketing)
  app/                         # landing, docs, blog
  components/marketing/        # hero, bento, testimonials, pricing (no prices)
  # NEW: interactive demo embed (iframe to hosted demo or a replay)
  # NEW: Motion for hero + scroll storytelling
  # NEW: bento grid for containment proof points

web/                           # Vite 8 + React 19 + TS (app shell)
  src/
    routes/                    # TanStack Router, typed
    components/primitives/     # React Aria (keep)
    components/ui/             # shadcn-compat layer (keep)
    lib/
      api.ts                  # typed client (keep, add generated types)
      query.ts                # TanStack Query
      store.ts                # Zustand
    features/
      agents/                 # NEW: agents-first IA (Inbox, Runs, Reviews, Policies, Repos)
      runs/                   # run detail: Overview, Changes, Checks, Trace, Audit
      fleet/                  # fleet table with saved views
    # NEW: SSE client (fetch-event-source) for agent activity stream
    # NEW: Pierre/CodeMirror 6 for diffs (lazy-loaded route)
```

Key additions:
- **OpenAPI codegen** (Hey API) from the Go server's OpenAPI spec → typed
  client. Eliminates hand-written `api.ts` drift.
- **SSE** endpoints on Go side + `fetch-event-source` client with heartbeats
  and reconnect (for the agent activity stream).
- **Route-level JS budgets** and lazy-loading for diff/editor/chart routes.
- **MSW** for tests (already planned).
- **Playwright** for E2E of the new shell (already planned).

---

## 8. Migration plan (strangler fig) — unchanged from plan

1. Add TanStack Router + Query + Zustand + OpenAPI codegen foundation.
2. Spike Pierre and CodeMirror 6 on real 10k/50k-line diffs (measure memory,
   FPS, TTFP). Do not commit to either until measured.
3. Build the new shell (global rail: Inbox, Runs, Reviews, Policies, Repos) in
   parallel; keep existing views behind the new shell until parity.
4. Migrate view by view; delete old view only after parity + tests.
5. Keep `npm run build` (tsc -b && vite build) and `go test ./...` green at
   every commit.

---

## 9. What NOT to do (anti-patterns from research)

- Do not switch from React Aria to Radix/Base UI mid-project.
- Do not put Motion (Framer Motion) in the app shell bundle.
- Do not rebuild the marketing site as a Vite SPA.
- Do not add a heavy component library (MUI/Ant) — our design system is
  custom and token-based; MUI would fight it.
- Do not claim "first Git forge for agents" (already in anti-claims).
- Do not publish prices before validation (PRICING.md §8).
- Do not use mock data in production (AGENTS.md invariant #1).
- Do not ship a marketing page without a fast, captioned demo video.

---

## 10. Sources

- "My Frontend Stack In 2026" — thetshaped.dev (Next.js for content, Vite 8
  for SPAs, Tailwind + shadcn, TanStack Query/Router, Zustand, Zod, Oxlint,
  Storybook, MSW, Vitest, Playwright).
- "The Modern React Stack in 2026" — LinkedIn / jsdev.space (Next.js, Vite,
  TanStack Query, Zustand, Tailwind, shadcn/ui, Zod, Vitest, Playwright).
- "Best React Component Libraries 2026" — designrevision.com, uxpin.com,
  untitledui.com, builder.io, supernova.io, greatfrontend.com, telerik.com
  (shadcn/ui top pick; React Aria best for a11y/i18n; Base UI foundational;
  HeroUI v3 on React Aria + Tailwind v4; Motion for marketing).
- "10 SaaS Landing Page Trends for 2026" — saasframe.io (interactive demos,
  micro-animations, bento grids, personalization).
- "Landing Page Design Trends 2026" — moburst.com, tncflow.com (Framer trends,
  lightweight CSS/JS animation, Core Web Vitals).
- "Product Hunt Launch Guide 2026" — smollaunch.com (30-day prep, 3 phases,
  200+ upvotes in 6 h, gallery readable on mobile, captioned demo video,
  landing matches PH copy).
- "6 Best Places to Launch in 2026" — medium.com/@hii_mohit (spread across
  platforms; Show HN high variance; badges for dofollow links).
- "Top 50 PH Startup Websites" — Figma Community (reference designs).
- Product Hunt leaderboard / categories pages (current top products).
- Existing repo: AGENTS.md §2 (design system), plan.md §3 (frontend redesign),
  web/src/styles/* (tokens).

---

## 11. Next actions (frontend)

1. Spike Pierre vs CodeMirror 6 on 10k/50k-line diffs; measure memory, FPS,
   TTFP. (Plan §3.1)
2. Add TanStack Router + Query + Zustand + OpenAPI codegen foundation.
   (Plan §3.2)
3. Build the new shell (global rail) in parallel; keep old views until parity.
4. Add SSE endpoints on Go side + fetch-event-source client.
5. Build the marketing hero with the containment demo embed + Motion.
6. Run the PH landing-page analyzer (smollaunch) before launch.
7. Keep `npm run build` and `go test ./...` green at every commit.
