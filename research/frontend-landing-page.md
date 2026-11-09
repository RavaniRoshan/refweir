# Deep Research: Best-in-Class Marketing Landing Page for an Agentic Git Forge (Refweir)

**Topic slug:** `frontend-landing-page`
**Depth:** deep
**Date:** 2026-11-09
**Companion docs:** `docs/launch/FRONTEND_RESEARCH.md`, `docs/launch/FRONTEND_TEMPLATES.md`, `docs/launch/POSITIONING.md`, `docs/launch/PRICING.md`, `plan.md` §3, `AGENTS.md` §2
**Decision supported:** Phase 3 frontend redesign — specifically the marketing site (`site/`) for the Product Hunt launch, and the "lowest effort, highest ROI" path to a premium motion-driven, agentic-looking landing page.

---

## 1. Executive summary

1. **Split the surfaces.** The highest-performing 2026 launches run a **marketing site** (Next.js, SSG/SSR, content-heavy) and an **app shell** (Vite + React SPA behind auth) as two separate frontends. Our repo already does this (`site/` vs `web/`). Do not collapse them.
2. **Keep our base.** Vite 8 + React 19 + TypeScript + Tailwind v4 + React Aria + OKLCH tokens is exactly the 2026 consensus stack. We are not behind. The migration is additive (TanStack Router/Query, SSE, Zustand, Motion on marketing only), not a rewrite.
3. **The "agentic" look is well-defined and reproducible** with free copy-paste libraries: dark-as-default + one neon accent, kinetic typography, bento grids, product-as-demo hero, scroll-triggered reveals. Linear is the most-copied exemplar; our containment demo video is the hero asset.
4. **Lowest-effort path:** fork Velora UI (MIT, whole-site, Tailwind v4, Motion) for the shell, layer Aceternity UI components (hero, bento, beams) and Magic UI (micro-interactions) on top. Estimated 3-5 days instead of 3-4 weeks. **But** Velora's accessibility scores (32/100 ARIA) and tiny community (61 stars) are real risks; Aceternity's are worse (4/112 reduced-motion, 11/112 ARIA). Accessibility must be added deliberately.
5. **Motion is a performance budget, not decoration.** Linear animates only `transform`/`opacity` at 0.1-0.35 s. Vercel's WebGL hero costs hundreds of KB and requires mobile fallbacks. The best motion is "invisible." Marketing-agency claims that "motion lifts conversion" are uncontrolled anecdotes; perf-focused sources disagree.
6. **Conversion evidence is weaker than it looks.** Most cited lift percentages (testimonials +34%, video +80-86%, social proof below CTA +68%, etc.) are single case studies from marketing blogs, not controlled experiments. Only 12% of A/B tests reach significance. Treat benchmarks as directional, not targets.
7. **Accessibility is non-negotiable and largely unaddressed by free libraries.** WCAG 2.3.3 (AAA), 2.2.2 (A), `prefers-reduced-motion`, focus management, and screen-reader support must be added by us. The free libraries mostly don't do it.
8. **Ship path:** 4 days — Day 1 shell + tokens, Day 2 hero (Aceternity + demo video embed), Day 3 bento + micro-interactions (Magic UI), Day 4 Lighthouse + reduced-motion + keyboard QA. Keep `npm run verify` green throughout.

---

## 2. Background and key terms

- **OKLCH**: perceptually uniform color space; our tokens use it. Some combinations are out of sRGB gamut; Chrome/Safari still use fast clipping for some cases — wrap P3 colors in `@media (color-gamut: p3)`.
- **Tailwind v4**: CSS-first (`@theme`/`@theme inline`), auto content detection, no JS config. Dark mode defaults to `@media` (class-based needs `@variant dark`).
- **shadcn/ui**: copy-paste component model (not a package). Components live in your repo. Built on Radix primitives (historically); multi-primitive support emerging.
- **React Aria (Adobe)**: headless primitives with best-in-class accessibility and i18n. Our existing layer.
- **Base UI**: unstyled primitives from Radix/MUI/Floating UI creators; shadcn/ui is adding multi-primitive support.
- **Motion (formerly Framer Motion)**: animation library. Full `motion` component cannot tree-shake below ~43 KB; LazyMotion + `m` = ~4.6 KB initial.
- **Core Web Vitals**: LCP ≤2.5 s, CLS ≤0.1, INP ≤200 ms. Performance budget for the marketing site.
- **Bento grid**: modular tile layout for feature sections; Apple popularized, Ramp formalized.
- **Kinetic typography**: motion-driven variable-font headlines (word swap, stretch, scroll response).
- **Product-as-demo**: hero shows the actual product UI animating autonomously.
- **Strangler fig**: incremental migration; new shell runs in parallel with old views until parity.
- **WCAG 2.3.3 (AAA)**: motion triggered by interaction can be disabled unless essential. WCAG 2.2.2 (A): auto-moving content >5 s needs pause/stop/hide.

---

## 3. Findings by sub-question

### SQ1 — Landing page structure (conversion-backed)

**Canonical section order** (SaaSHero 10-section framework, corroborated by Orbix/Flowout 6-section model):
1. Minimal nav
2. Hero (headline + subheadline + CTA + product visual)
3. Initial social proof (logo bar)
4. Problem & solution
5. How it works
6. Features & benefits
7. Deep social proof & case studies
8. Pricing preview
9. FAQ
10. Final CTA + footer

Each section "earns the next" by resolving a specific objection before the visitor names it. (source: saashero.net, orbix.studio, flowout.com)

**Hero:** Headline = Outcome + Pain Point; subheadline = Mechanism; single low-friction CTA; visual = product in context (screenshot/GIF/video, not stock illustration). Outcome-anchored headlines beat category claims (Shop Boss 305% lift from headline rewrite — single case study). (source: flowout.com, saashero.net)

**Reading level:** 5th–7th grade copy → 11.1% median CVR vs 5.3% college-level (Unbounce, 41k pages). **Caveat:** Unbounce study is 2023–2024 (Jul 2023–Jul 2024), not 2026; reading-level correlations may have shifted with mobile-first consumption and AI-generated copy. (source: digitalapplied.com, unbounce.com)

**Page length:** 250–725 words optimal (Unbounce). Tension: long-form pages with mid-page CTA converted 220% better for complex/high-consideration B2B. Reconciliation: short for simple products, long-form for complex/regulated ones (we are the latter).

**Social proof:** Logo bar immediately after hero (above the fold). Three tiers: logo bars (instant recognition), metric-based proof mid-page, detailed testimonials/case studies near feature claims. Testimonials +34% (VWO/WikiJobs); video testimonials 80–86% better than text; social proof below CTA +68%; reviews up to 270% for premium (Northwestern Spiegel, 2017 — "up to 270%" is the maximum for exactly 5 reviews on premium products, easily misread); Docsend +260% from enterprise logos; Basecamp +102% from real reviews; real-time social proof +98%; 92% read testimonials; only 23.2% of marketers include social proof (competitive gap). **Caveat:** all single-source case studies, not controlled experiments. (sources: flint.com, userevidence.com, apexure.com, quora.com)

**CTA repetition:** One conversion goal reinforced at multiple scroll depths beats competing CTAs by 20–30% (Apexure, 3,000+ pages). CTAs at every major fold +20–35%. Single repeated CTA at 3 points = 51.78% CVR (Radicle Science). "Book a Demo" 13.1% vs "Contact Us" 4.8% (Docket). First-person copy ("Get My Free Audit") +10–25% vs second-person. Specific copy +30–40% vs "Submit". CTA above fold converts ~2x vs below fold. Pages burying CTA below fold convert at ~half the rate. **Caveat:** "above fold" concept increasingly contested in scroll-heavy mobile UX.

**Bento grid:** No data-backed conversion evidence found. Described as a visual trend (lapa.ninja, LinkedIn). **Conclusion: bento is a 2026 aesthetic convention, not a conversion-proven pattern.** Use it for visual polish, not because it converts.

**Benchmarks:** Median CVR 6.6% all industries (Unbounce, 41k pages, 464M visits, Jul 2023–Jul 2024); SaaS median 3.8%; top quartile 11.4%+. Only 12% of A/B tests reach significance (Optimizely, 173k experiments). 83% mobile visits; desktop converts 8% better. 5 or fewer form fields convert significantly better. **Caveat:** Unbounce benchmark is 2023–2024, not 2026.

**Developer-tool specific:** Dense product screens above the fold (Linear, Cursor, Posthog, Vercel); dark mode by default; interactive in-page demos cited as "the conversion lift of 2026" (no quantified study found). (source: designkey.studio, adminlte.io)

**Performance as design constraint:** Linear/Vercel/Framer score >90 Lighthouse accessibility; every 100KB hero media increases bounce ~1.8%; 1s load delay reduces conversions ~7% (Akamai 2017 retail — increasingly outdated). First aesthetic impression forms in 17–50ms (Google); users decide to stay within 10–20s; 57% of viewing time is above the fold (NN/g). (sources: digitalapplied.com, designkey.studio, thrivethemes.com)

### SQ2 — Free templates and libraries

**Tier 1 (copy-paste, free, shadcn-compatible):**
- **shadcn/ui**: MIT, ~123k stars, 8.7M CLI downloads/week. Tailwind v4 + React 19 since Feb 2025. July 2026: multi-primitive support (Radix default, Base UI and React Aria available). Known: no marketing sections (~30 blocks, mostly sidebar/login + 1 dashboard), generic default look, `init --force` path issues, 1007 open issues.
- **Magic UI**: MIT free core; Pro $199. Tailwind v4 + React 19 by default. 6/78 components handle `prefers-reduced-motion` (accessibility risk). No Figma kit.
- **Velora UI**: MIT, 61 stars, 100 components, 31 free blocks. Next.js 16 + Tailwind v4 + Motion. 89/100 reduced-motion, 32/100 ARIA, 51/100 zero deps, <3KB gzipped. Pro tier $149. Very new (small community, unproven long-term maintenance).
- **Launch UI**: MIT free + Pro. Next.js 16 + Tailwind v4 + shadcn + Motion. Smaller ecosystem.
- **Animata**: MIT, ~3.0k stars. 156+ components. Next.js 16 + Tailwind v4 + Storybook 10 confirmed. Smaller community.
- **Aceternity UI**: ambiguous license (no SPDX). 118+ free components. 4/112 reduced-motion, 11/112 ARIA. Custom license for Pro. Keyframe migration needed for v4. **Accessibility risk is severe.**
- **PrebuiltUI**: proprietary. Free tier allows end products but no source redistribution. Tailwind v4.
- **21st.dev**: community registry, inconsistent quality, free plan capped.
- **MVPBlocks**: BSD-3-Clause, 834 stars. Tailwind v4 not explicitly documented.
- **Cruip Simple Light/Open**: GPL-3.0 stated in README but NO LICENSE file (legal ambiguity). Next.js 15, Tailwind v4. 4.5k stars. Most-polished free SaaS landing.

**Context:** Tailwind CSS remains free MIT, actively maintained (v4.2, Apr 2026). Tailwind v4 migration friction common across libraries: dark mode defaults to `@media`, `@tailwind` removed, opacity modifiers, `@source` for monorepos, Turbopack HMR, `next/font` CSS variables.

**Key risk:** Velora UI is recommended as primary choice despite 32/100 ARIA and 61 stars. Its accessibility gaps are real but smaller than Aceternity's. Accessibility must be added deliberately regardless of library choice.

### SQ3 — Motion techniques and performance

**Linear:** Next.js static export, styled-components, inline SVG sprite. Animates only composited properties (`transform`, `opacity`) plus occasional paint (`background-color`, `border-color`). Never layout properties. Durations: quick .1s, regular .25s, slow .35s. Asymmetric enter (instant) / exit (150ms). No transitions on list rows. Motion is spatial (origin-referenced). Inlined app-shell CSS with `logoBackgroundPulse` keyframe. Single Inter Variable woff2 with `font-display: swap`. (source: performance.dev reverse-engineering article)

**Vercel:** WebGL via Three.js/R3F, custom GLSL shaders (rainbow/iridescence spectrum, raymarched reflections), Bloom + LUT postprocessing, instanced meshes with additive blending. Mobile mitigations: SSR off, `antialias: false`, textures instead of extra rendering, postprocessing merged into single pass. (source: Vercel blog)

**Raycast:** No technical source found. Only marketing copy ("Your shortcut to everything"). **Unverified.**

**Cursor:** "Spring-based layout animations and shared element transitions" in marketing copy only. No verified implementation. **Unverified.**

**CSS scroll-driven animations:** Baseline 2026 (~85% support, Firefox behind flag). Compositor thread, no JS. (MDN, Josh Comeau)

**Framer Motion bundle sizes:** Full `motion` component cannot tree-shake below ~43 KB. LazyMotion + `m` = ~4.6 KB initial (domAnimation). Full package = 75.4 KB min+gz. Motion One base = 4 KB + 2 KB React. **Note:** The "3.12 KB min vs GSAP 27.29 KB" comparison from the fast-ship-path research is not apples-to-apples (min vs min+gz). Motion's actual gzipped React bundle is larger than 3.12 KB.

**State of Web Animation 2026:** 91.7% of mobile pages use CSS transitions; 18.4% use JS animation library; 40% (mobile) / 44% (desktop) have jank-prone non-composited animations. **Caveat:** No publication by this exact name is known; these precise figures may be misattributed. The HTTP Archive Web Almanac covers animation but typically lags by 1-2 years.

**Performance principles:** Only `transform`/`opacity` on compositor; other properties force layout/paint. Animations <300ms for UI feedback. JS animations must use `requestAnimationFrame` (throttled in background tabs). `will-change` targeted, not blanket. Lazy-load below-the-fold via IntersectionObserver.

**Disagreement:** Marketing-agency sources claim animated landing pages outperform static ones when animation clarifies value prop. Performance sources argue motion often adds jank/bundle weight; best animation is "invisible." No controlled causal data found.

**Tension to reconcile:** "Best animation is invisible" (Linear) vs. "kinetic typography and product-as-demo are growing trends" (agentic visual language). These pull in opposite directions on motion intensity. Reconciliation: motion should serve comprehension (product-as-demo, kinetic type that clarifies), not decoration. Linear's own site uses scroll-driven reveals and kinetic type but keeps durations short and properties composited. The difference is intensity and property choice, not presence/absence of motion.

### SQ4 — Product Hunt launch patterns

**PH listing as mini pitch deck:** 5-slide gallery storyboard (hero → core workflow → outcome → proof → CTA). Thumbnail 240×240 (or 640×480). Tagline <60 chars, benefit-driven. 5-10 gallery images. 30-90s demo video (1080p, captions, YouTube/Vimeo). First comment (maker story). External link with UTMs.

**Dedicated website structure:** Clear headline naming concrete outcome (not "AI platform to turbocharge…"). Concise value prop. Product-forward hero (real screenshot, not illustration). Single-field or single CTA above fold. Social proof above fold. Trust signals near CTA ("free trial", "cancel anytime", "no credit card"). Annotated screenshots with captions. Demo video (product walkthrough, outcome shown, not empty state). PH-specific landing page with offer/discount + PH badge. FAQ. Message match between PH copy and landing page. Mobile-readable.

**Conversion patterns:** Email capture is primary conversion goal (target 1000+ emails; PH is top-of-funnel). Avg 2-5% CVR, strong 5-10%, exceptional 10%+. Avg 1,000-5,000 PH visits (strong 5,000-15,000). "3-second comprehension" rule. Screenshot the moment after action completes. 3 feature bullets max above fold. No navigation on campaign pages. Repeat CTA at multiple points. Power-user/community reviews drive conversion. Load-tested, stable onboarding. PH-specific offer on dedicated landing page increases engagement. PH badge embed showing live upvotes.

**Tagline:** Action verb + outcome + time/ease. Clarity beats grandeur. PH CEO: no vague/grander taglines.

**Benchmarks:** Self-reported, varying by product type (B2B vs B2C). No controlled studies found.

**Launch logistics:** 30-day prep, 3 phases. Launch Tue–Thu, 12:01 AM PST. 200+ upvotes and 30+ comments in first 6 hours. Respond to every comment within 30 min. 12-16 hour marathon. Pre-built momentum (audience 30 days before). Spread across platforms over 3 weeks (PH, then Show HN, Uneed, Fazier, Peerlist). PH link is nofollow; others give dofollow backlinks. Show HN for technical audience (highest variance). Badges/embeds for dofollow links (~$115 total). Run landing-page analyzer (smollaunch) before launch.

### SQ5 — Agentic visual language

**Dark-mode-by-default + one neon accent:** ~75% of top design-led SaaS sites (toimi.pro — but this is an unmeasured design-trend observation, not a survey; precision is implausible). Named "one-color ownership": Linear (purple), Raycast (red), Cursor (cyan), Mercury (lime), Vercel (deep near-black). Counter-movement: Anthropic, PostHog, Notion (cream/serif/warm illustration). (sources: toimi.pro, Behance/Contra case studies)

**Bento grids:** Apple popularized; Ramp formalized into internal design system with Bakken & Baeck. 55+ SaaS examples (Unkey, Vercel, Customer.io, Mintlify, Attio, Huly, WorkOS, Cal.com). Attio uses hover micro-interactions (scale + shadow + content reveal). Good for compact overview of equal-weight capabilities; poor for sequential stories or long text. (sources: toimi.pro, saasframe.io, muz.li)

**Product-as-demo:** "Website is the product demo." Linear (ENG-2703 agent picks up task, comments, opens PR, animating autonomously). Attio ("Ask Attio" natural language query returning filtered data). ElevenLabs (audio on hover). Lovable (prompt input in hero). Figma (rotating community files). Ungated interactive demos generate 2x engagement and 20-25% faster close. **Caveat:** "2x engagement" is single-source marketing analysis, not peer-reviewed. (sources: toimi.pro, userpilot.com)

**Kinetic typography:** Linear swaps "teams"↔"agents". Framer motion-driven display type. Growing from decorative to functional. Motion/Framer Motion. (sources: toimi.pro, Medium/Bradley Kinnard, b12 guide)

**Agentic UX patterns (product-level):** Presence/activity stream, plan-and-progress, confidence signals, receipts/undo, stop button (Design Accelerator). Seven AI-first SaaS trends (Outcrowd): chat→agent, trust/control, explainability/audit, multimodal, adaptive interface, AI search, invisible AI. (sources: designaccelerator.substack.com, medium.com/outcrowd)

**Quantitative claims** (2x engagement, 127% lift, 3.8% median CVR) are single-source marketing analyses, not peer-reviewed.

### SQ6 — Shader/WebGL/canvas backgrounds

**Measured bundle sizes (gzipped JS, one gradient):** Paper Shaders MeshGradient 8.4 KB (WebGL2, Apache-2.0, zero deps, 20+ shaders, pin version — 0.0.x breaking changes). Hazeglow 14.0 KB (MIT, prefers-reduced-motion check, onUnsupported fallback). NEAT (FireCMS) 27.2 KB (MIT + Commons Clause). gradients.sh 52.3 KB (WebGPU→WebGL2 fallback, server-rendered CSS fallback). InstantShader Flow 58.7 KB (MIT). ShaderGradient 293.6 KB (three.js + R3F). React Bits Aurora ~18 KB (no deps, reduced-motion check, pauses off-screen). Aceternity CSS blur: no bundle cost but per-frame repaints.

**WebGL libraries:** OGL ~14-33 KB gzipped (numbers disagree by measurement method/version). three.js 139 KB gzip (556 KB min); R3F aliasing can reduce to 50-80 KB. Framer Shader Backgrounds (free, WebGL, optimized Three.js, pauses off-screen).

**CSS-only:** `@property` animated gradients, zero bundle cost. Animating color stops triggers paint per frame. CSS orbs with transform/opacity cheapest.

**Canvas 2D:** quantize alpha, cap dpr at 2, fillRect over arc.

**Core Web Vitals:** LCP ≤2.5s (heavy JS delays parse), CLS ≤0.1 (reserve dimensions), INP ≤200ms (main-thread loops block). Mitigations: dynamic import, CSS fallback first-paint, pause on hidden tab/off-screen, cap dpr at 1.5-2, OffscreenCanvas + worker.

**Disagreements:** OGL size numbers conflict; three.js tree-shaking claims improved but R3F still recommends aliasing; CSS approaches trade bundle for repaint; gradients.sh is vendor source.

**Open:** VFX-JS size unknown; whether canvas can become LCP; WebGPU overhead.

### SQ7 — Accessibility and reduced-motion

**WCAG criteria:** 2.3.3 (AAA) Animation from Interactions — motion triggered by interaction can be disabled unless essential. 2.2.2 (A) Pause, Stop, Hide — auto-moving content >5s needs pause/stop/hide. 2.3.1 (A) Three Flashes or Below Threshold. `prefers-reduced-motion` is primary mechanism (C39 CSS, SCR40 JS). Reduce, don't remove: keep subtle fades/color changes, remove large-scale motion (parallax, zooming, sliding panels, spinning, scroll-triggered displacement). Essential animation exception.

**Focus:** 2.4.7 Focus Visible (AA), 2.4.11 Focus Not Obscured (AA, new in 2.2), 2.4.13 Focus Appearance (AAA), 2.4.1 Bypass Blocks (A) skip link, 2.1.1 Keyboard (A). Modal focus trap + return focus. SPA route changes: move focus to main heading + aria-live.

**Screen reader:** Semantic HTML before ARIA, aria-live regions (polite/assertive), headings hierarchy, accessible names, alt text, toasts need aria-live.

**Implementation:** CSS opt-out (`@media (prefers-reduced-motion: reduce)`) or progressive (`no-preference`). Conditional stylesheet loading. JS: `matchMedia` with 'change' listener (parentheses obligatory). Motion.dev: `MotionConfig reducedMotion="user"` auto-disables transform/layout but keeps opacity/bgColor. `useReducedMotion` hook. Replace transform with opacity. Parallax: `y: shouldReduceMotion ? 0 : y`. Autoplay video: `autoplay={!shouldReduceMotion}`. `<picture>` with source media for animated images. Manual in-site toggle per WCAG 2.3.3 example.

**Testing:** Chrome DevTools Rendering panel, OS settings (macOS/Windows/iOS/Android).

**Emerging:** `prefers-reduced-transparency`, `prefers-contrast`, `inverted-colors`, `prefers-reduced-data` (Media Queries Level 5).

**Browser support:** Chrome 74, Edge 79, Firefox 63, Safari 10.1.

**Open:** Whether 2.3.3 (AAA) becomes legal requirement under ADA/EAA; WCAG 3.0 draft may restructure; whether `prefers-reduced-motion` alone satisfies 2.3.3 without in-site toggle (debated, not resolved); no quantitative compliance data.

### SQ8 — Fastest ship path

**Day 1:** Static export (`output: 'export'`, `images.unoptimized: true`, `compress: true`, `poweredByHeader: false`). Tailwind v4 CSS-first `@theme`/`@theme inline`. OKLCH tokens with `-foreground` pairs.

**Day 2:** Component sourcing. Velora UI recommended (MIT, 100 components, 31 free blocks, Next.js 16 + Tailwind v4 + Motion, 89/100 reduced-motion, 51/100 zero deps, <3KB gzipped). Install via `npx shadcn@latest add @velora/...`. Aceternity: 112 free components but 4/112 reduced-motion, 11/112 ARIA, 6 CLI installs missing keyframes. Custom license.

**Day 3:** Motion (hardware-accelerated via WAAPI, MIT). Gate Lenis behind reduced-motion + mobile breakpoint, destroy on cleanup.

**Day 4:** Self-host fonts via `next/font/google` with `display: 'swap'`, preload hero images per breakpoint, prevent FOUC with blocking script in `<head>` + `suppressHydrationWarning`, verify reduced-motion + keyboard.

**Pitfalls:** Hydration mismatch (browser APIs in `useEffect`, dynamic imports `ssr: false`). FOUC (blocking inline script in `<head>`, `try` for Safari private mode). Token drift (edit tokens not component files, `@theme inline` for dark mode). Missing keyframes (Aceternity 6 components). Reduced-motion/ARIA gaps (Aceternity 4/112, 11/112). Static export + dynamic routes (`generateStaticParams`, no `use client` + `generateStaticParams`). Image optimization disabled (`unoptimized` or custom loader). Bundle size (import only what you use). `@theme inline` resolution. OKLCH gamut clipping (wrap P3 in `@media (color-gamut: p3)`). Async params in Next.js 15 (Promise).

**Stack:** Next.js 15 App Router static export, Tailwind v4 CSS-first, Velora UI, Motion, Lenis (gated), `next/font/google` self-hosted.

**Confidence:** High on architecture/pitfalls, medium on day-by-day timeline estimates.

**Open:** Does existing site use shadcn primitives? Target Lighthouse score and WebGL desired? Dark-mode toggle vs system-only? Analytics/cookie scripts blocking theme script?

---

## 4. Analysis and discussion

### Patterns across sub-questions

1. **The "agentic" aesthetic is convergent and reproducible.** Dark-as-default + single neon accent, bento grids, product-as-demo hero, kinetic typography. Linear is the most-copied exemplar. This is a visual language, not a technology — it can be reproduced with free tools.

2. **Motion is the differentiator but also the risk.** The top sites use motion extensively (Vercel WebGL, Linear scroll-driven, Framer kinetic type), but the best ones keep it compositor-only and short. The "invisible animation" principle (Linear) and "motion serves comprehension" (product-as-demo) reconcile the apparent contradiction with "kinetic typography is growing." The difference is intent and property choice.

3. **Free libraries can produce a premium look but not premium accessibility.** Velora (89/100 reduced-motion but 32/100 ARIA) and Aceternity (4/112 reduced-motion, 11/112 ARIA) are visually impressive but accessibility-poor. For a regulated/air-gapped buyer, accessibility is a procurement requirement. We must add a11y deliberately.

4. **Conversion evidence is weaker than it appears.** Most cited lift percentages are single case studies from marketing agencies. Only 12% of A/B tests reach significance. Bento grids have no controlled conversion data. Treat all benchmarks as directional.

5. **The fastest path is real but has pitfalls.** 4 days is achievable with Velora + Aceternity + Magic UI, but hydration, FOUC, token drift, missing keyframes, and OKLCH gamut clipping are real failure modes. The blocking `<head>` script for FOUC is the #1 missed step.

6. **Performance is a design constraint, not an afterthought.** Every 100KB hero media increases bounce ~1.8%. Paper Shaders at 8.4 KB is the right price for a hero shader. three.js at 139 KB is too expensive for a marketing page unless code-split and lazy-loaded.

### Multiple perspectives

**Perspective A (design-led):** Motion and visual polish are the primary differentiators. Linear, Vercel, Raycast, Cursor all invest heavily in hero animation, kinetic type, and scroll storytelling. The "agentic" aesthetic signals premium quality and builds trust. Use Aceternity/Magic UI for maximum visual impact at minimal effort.

**Perspective B (performance-led):** Motion adds bundle weight and jank. 40-44% of pages have jank-prone animations. The best animation is invisible. Keep motion compositor-only, short, and reduced-motion-aware. Use CSS scroll-driven animations instead of JS libraries where possible.

**Reconciliation:** Both are right about different things. Visual polish matters for first impressions (17-50ms). Performance matters for conversion (1s delay = 7% drop). The synthesis: use motion for comprehension (product-as-demo, kinetic type that clarifies), keep it compositor-only and short, lazy-load heavy assets, respect reduced-motion, and measure Core Web Vitals. Do not use motion for decoration.

### Strengths and weaknesses of the recommended approach

**Strengths:**
- Free templates + copy-paste libraries = 3-5 days instead of 3-4 weeks.
- Tailwind v4 + OKLCH tokens already in place; no migration needed.
- Velora's accessibility scores (89/100 reduced-motion) are better than Aceternity's.
- Motion is MIT and hardware-accelerated.
- Static export = zero cold start, CDN-hostable.

**Weaknesses:**
- Velora UI is very new (61 stars) — long-term maintenance unproven.
- Aceternity's license is ambiguous (no SPDX) — legal risk for commercial shipping.
- Free libraries' accessibility scores are poor; we must add a11y ourselves.
- Conversion evidence is weak (single case studies, not controlled experiments).
- The "State of Web Animation 2026" stats may be misattributed.
- OKLCH gamut clipping needs careful handling.

---

## 5. Conclusions and implications

1. **The marketing site should be rebuilt using free templates (Velora shell + Aceternity hero/bento + Magic UI micro-interactions) in 3-5 days, not 3-4 weeks.** This is the highest-ROI path to a premium agentic look.
2. **Keep React Aria as the primitive layer.** Do not switch to Radix/Base UI mid-project. The accessibility investment is already made and is the right choice for regulated buyers.
3. **Motion goes on the marketing site only.** CSS transitions in the app shell. No Motion in the app bundle.
4. **Accessibility must be added deliberately.** Free libraries don't do it. Budget 1 day for reduced-motion, focus management, and screen-reader support.
5. **Performance is a design constraint.** Paper Shaders (8.4 KB) for hero shader. Dynamic-import heavy assets. Measure Core Web Vitals before launch.
6. **Conversion claims are directional, not targets.** Use them to prioritize (social proof near CTA, repeat CTA, product-in-hero) but don't expect the cited lift percentages.
7. **The containment demo video is the hero asset.** Embed it above the fold with poster + lazy load + captions. This is the "product is the demo" pattern.
8. **Launch logistics matter as much as the page.** 30-day prep, Tue-Thu launch, 200+ upvotes in 6 hours, respond to every comment within 30 min, spread across platforms.

---

## 6. Recommendations

### Immediate actions (before Phase 3 coding starts)
1. **Spike Velora UI fork** in a branch. Port OKLCH tokens. Verify `npm run build` + claims audit pass. Measure reduced-motion and ARIA gaps.
2. **Copy 5 Aceternity components** into `site/components/marketing/`. Render in a draft page. Measure LCP/CLS/INP.
3. **Embed the containment demo video** in the hero (poster + lazy + captions).
4. **Add kinetic typography headline** ("Contain every agent. Prove every merge.").
5. **Run smollaunch landing-page analyzer + Lighthouse** before launch.
6. **Verify Tailwind Plus availability** directly with Tailwind Labs (do not assume the Shopify acquisition claim is true).
7. **Confirm shadcn/ui's default primitive layer** (Radix vs Base UI) before adopting new third-party registries.

### Further research needed
1. **Visual audit of Linear, Vercel, Raycast, Cursor landing pages** — all current findings are textual; no screenshot-level evidence. A visual audit would verify specific implementations.
2. **Independent bundle-size verification** of Paper Shaders, OGL, and Velora's "<3KB gzipped" claim.
3. **Accessibility audit of Velora UI and Aceternity UI** against WCAG 2.2 AA — the self-reported scores need independent verification.
4. **Developer-tool-specific conversion benchmarks** — all available data is generic B2B SaaS.
5. **WCAG 3.0 draft** — may restructure motion criteria; monitor for changes.

---

## 7. Disagreements and open questions

### Resolved disagreements
1. **"Best animation is invisible" vs "kinetic typography is growing"** — Reconciled: motion should serve comprehension, not decoration. Linear uses both (short, composited, purposeful). The difference is intent and property choice.
2. **"Short page (250-725 words) vs long-form (+220% for complex B2B)"** — Reconciled: short for simple products, long-form for complex/regulated ones (we are the latter).
3. **"Bento grid has no conversion evidence" vs "bento is a trend"** — Reconciled: use bento for visual polish, not because it converts.

### Unresolved disagreements
1. **Motion helps conversion vs motion hurts performance** — No controlled causal data. Marketing agencies claim lift; performance sources claim jank. Treat as: motion for comprehension, measured by Core Web Vitals.
2. **Dark-mode-by-default (~75%) vs editorial counter-camp (Anthropic/PostHog/Notion)** — Unquantified trend; both are valid positioning choices. Our positioning (self-hosted, air-gapped, auditable) leans dark/technical, which aligns with the 75%.
3. **CSS-only vs WebGL for hero backgrounds** — Tradeoff: zero bundle vs visual fidelity. Decision depends on brand positioning and performance budget.

### Open questions (carried from subagent findings)
1. No first-party Core Web Vitals published for linear.app, vercel.com, raycast.com, cursor.com.
2. Whether Vercel's hero still uses Three.js or has migrated to WebGPU/TSL.
3. Raycast and Cursor animation stacks are unverified (no technical sources).
4. Whether CSS scroll-driven animations improve CWV vs JS libraries in practice (theoretical only).
5. Whether WCAG 2.3.3 (AAA) becomes a legal requirement under ADA/EAA in 2026.
6. Whether `prefers-reduced-motion` alone satisfies 2.3.3 without an in-site toggle (debated, not resolved).
7. No quantitative compliance data for motion-heavy landing pages.
8. Velora UI long-term maintenance (61 stars, very new).
9. Aceternity UI license status (no SPDX).
10. Cruip Simple Light/Open legal status (GPL stated but no LICENSE file).
11. MVPBlocks and Animata explicit Tailwind v4 support per-component not documented.
12. Whether shadcn/ui's multi-primitive support breaks existing Radix-based third-party registries.

### Verification corrections applied
- **Tailwind Plus/Shopify acquisition claim: REMOVED.** This was fabricated by the subagent. Tailwind CSS remains free MIT and actively maintained. Verify Tailwind Plus availability directly with Tailwind Labs.
- **shadcn/ui "switched default to Base UI" claim: CORRECTED.** shadcn/ui is adding multi-primitive support (Radix default, Base UI and React Aria available), not switching the default. The "July 2026 default switch" claim is unverified.
- **"State of Web Animation 2026" stats (91.7%/18.4%/40%/44%): FLAGGED.** No publication by this exact name is known. Treat as unverified.
- **"~75% dark-mode-by-default" precision: FLAGGED.** Unmeasured design-trend observation, not a survey. Treat as directional.
- **Conversion-lift percentages: FLAGGED as single-source anecdotes.** Not controlled experiments. Use as prioritization hints, not targets.
- **Unbounce reading-level study (2018-ish) and Akamai 7% delay (2010-2016): FLAGGED as stale.** Directionally useful but not 2026-precise.
- **Northwestern Spiegel "270%" (2017): FLAGGED.** Maximum effect for exactly 5 reviews on premium products. Easily misread.
- **Motion vs GSAP "3.12 KB vs 27.29 KB": CORRECTED.** Not apples-to-apples (min vs min+gz). Motion's actual gzipped React bundle is larger than 3.12 KB.
- **Velora UI recommendation: FLAGGED as risk.** 32/100 ARIA and 61 stars. Accessibility must be added deliberately.
- **Aceternity UI: FLAGGED.** 4/112 reduced-motion, 11/112 ARIA. Severe accessibility gap.
- **gradients.sh size claims: FLAGGED as vendor source.** Treat comparisons as marketing.
- **Velora self-reported audit scores: FLAGGED.** No independent verification.
- **"2x engagement from ungated interactive demos": FLAGGED as single-source.**
- **"CTA above fold converts ~2x": FLAGGED as single-source, increasingly contested on mobile.**
- **Paper Shaders 8.4 KB: FLAGGED as unverified.** Real library, but exact size should be independently confirmed.
- **OGL size "14-33 KB": FLAGGED as unresolved.** Critical bundle-budget question remains open.

---

## 8. Full source list

### Landing page structure
- SaaSHero, "Effective SaaS Landing Page Structure" — https://www.saashero.net/content/effective-saas-landing-page-structure/
- Orbix Studio, "25 High-Converting SaaS Landing Page Examples 2026" — https://www.orbix.studio/blogs/saas-landing-page-examples
- Flowout, "SaaS Landing Page Design" — https://www.flowout.com/blog/saas-landing-page-design
- Digital Applied, "Landing Page Conversion Study" — https://www.digitalapplied.com/blog/landing-page-conversion-study-2000-pages-tested-2026
- Unbounce Conversion Benchmark Report — https://unbounce.com/conversion-benchmark-report/
- DesignKey, "SaaS Website Design Patterns That Convert 2026" — https://www.designkey.studio/post/saas-website-design-patterns-that-convert-2026
- AdminLTE, "SaaS Landing Page Design Examples" — https://adminlte.io/blog/saas-landing-page-design-examples/
- ThriveThemes, "Hero Section Examples" — https://thrivethemes.com/hero-section-examples/
- Flint, "Landing Page Social Proof Element Performance Statistics" — https://www.flint.com/articles/landing-page-social-proof-element-performance-statistics
- UserEvidence, "How to Increase Landing Page Conversions by Adding 3 Types of Social Proof" — https://userevidence.com/blog/how-to-increase-your-landing-page-conversions-by-adding-3-types-of-social-proof/
- Apexure, "Landing Page Call to Action Button Tips" — https://www.apexure.com/blog/landing-page-call-to-action-button-tips
- SaaSHero, "Landing Page Design CTA Placement" — https://www.saashero.net/strategy/landing-page-design-cta-placement/
- Skymoon Infotech — https://skymooninfotech.com/blogs/landing-page-best-practices/
- Flowtrix, "12 B2B Landing Page Design Examples for 2026" — https://www.flowtrix.co/blogs/12-b2b-landing-page-design-examples-for-2026
- MarketWithBoost, "10 Best Practices for Landing Pages That Convert in 2026" — https://www.marketwithboost.com/insights/best-practices-for-landing-pages
- Lapa.ninja Bento Grid — https://www.lapa.ninja/category/bento-grid/
- LinkedIn bento grid post — https://www.linkedin.com/posts/namyakhan_landing-page-tip-2-your-bento-grid-isn-activity-7338935652811448322-GuXD
- Avidclan Technologies, "Bento Grid vs Card Layout Conversion Rate Reality" — https://avidclantechnologies.hashnode.dev/bento-grid-vs-card-layout-conversion-rate-reality

### Free templates and libraries
- shadcn/ui — https://ui.shadcn.com
- Magic UI — https://magicui.design
- Velora UI — https://velora.colorlib.com (ColorlibHQ/velora-ui on GitHub)
- Launch UI — shadcn/ui blocks for landing pages
- Animata — https://animata.design (codse/animata on GitHub)
- Aceternity UI — https://ui.aceternity.com
- PrebuiltUI — https://prebuiltui.com
- MVPBlocks — https://github.com/subhadeeproy3902/mvpblocks
- Cruip — https://cruip.com (cruip/tailwind-landing-page-template)
- 21st.dev — https://21st.dev
- Velora UI comparison page (accessibility audit) — https://velora.colorlib.com/compare
- Tailwind CSS v4 release — https://tailwindcss.com/blog/tailwindcss-v4

### Motion and performance
- Performance.dev, "How's Linear so fast?" — https://performance.dev/how-is-linear-so-fast-a-technical-breakdown
- Vercel blog, "Building an Interactive WebGL Experience in Next.js" — https://vercel.com/blog/building-an-interactive-webgl-experience-in-next-js
- Motion.dev, "Reduce Bundle Size" — https://motion.dev/docs/react-reduce-bundle-size
- Motion.dev, "Accessibility" — https://motion.dev/docs/react-accessibility
- MDN, "Scroll-driven Animations" — https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Scroll-driven_animations
- Josh Comeau, "Scroll-Driven Animations" — https://www.joshwcomeau.com/animation/scroll-driven-animations/
- 2pointagency, "Website Animation Complete Guide 2026" — https://www.2pointagency.com/guides/website-animation-the-complete-guide-to-motion-performance-and-ux-in-2026/
- Annhydrate, "State of Web Animation" — https://annnimate.com/state-of-web-animation (unverified exact name)
- Bundlephobia — https://bundlephobia.com
- Hontran.dev Framer Motion vs GSAP comparison
- Evil Martians, "Faster WebGL with OffscreenCanvas and Web Workers" — https://evilmartians.com/chronicles/faster-webgl-three-js-3d-graphics-with-offscreencanvas-and-web-workers

### Product Hunt launch
- Smol Launch, "Product Hunt Launch Guide 2026" — https://smollaunch.com/guides/launching-on-product-hunt
- Product Hunt, "Sharing Your Launch" — https://www.producthunt.com/launch/sharing-your-launch
- Lenny's Newsletter (Leo Bosuener), "How to Successfully Launch on Product Hunt" — https://www.lennysnewsletter.com/p/how-to-successfully-launch-on-product
- Demand Curve, "Product Hunt Launch" playbook — https://www.demandcurve.com/playbooks/product-hunt-launch
- Swipefiles (Corey Haines), "How to Launch on Product Hunt" — https://www.swipefiles.com/articles/how-to-launch-on-product-hunt
- RocketDevs, "Product Hunt Launch: How These 15 Companies Got #1 in 2025" — https://rocketdevs.com/blog/how-to-launch-on-product-hunt
- Whale, "The Ultimate Product Hunt Launch Checklist 2026" — https://usewhale.io/blog/product-hunt-launch-checklist
- Product Hunt, "Launch" (official) — https://www.producthunt.com/launch
- Arounda, "From Idea to Product Hunt Launch" — https://arounda.agency/blog/from-idea-to-product-hunt-launch-secrets-of-winning
- Instapage, "10 Landing Page Best Practices for 2026" — https://instapage.com/blog/landing-page-best-practices
- Orbix Studio — https://www.orbix.studio/blogs/saas-landing-page-examples
- MarketWithBoost — https://www.marketwithboost.com/insights/best-practices-for-landing-pages
- Skymoon Infotech — https://skymooninfotech.com/blogs/landing-page-best-practices/

### Agentic visual language
- Toimi, "Top 10 Best SaaS Website Designs 2026" — https://toimi.pro/blog/best-saas-website-designs
- SaaSFrame, "Bento Grid Pattern" — https://www.saasframe.io/patterns/bento-grid
- Muz.li, "Bento Grid Inspiration" — https://muz.li/inspiration/bento-grid/
- Userpilot, "SaaS Landing Pages" — https://userpilot.com/blog/saas-landing-pages/
- Design Accelerator (Substack), "The New Era of Agentic Product Design" — https://designaccelerator.substack.com/p/the-new-era-of-agentic-product-design
- Outcrowd (Medium), "2026 Design Trends for AI-First SaaS Products" — https://medium.com/outcrowd/2026-design-trends-for-ai-first-saas-products-90d8ce00391f
- Stacks, "Designing a Visual Language for Stacks Agents" — https://stacks.ai/resources/designing-a-visual-language-for-stacks-agents
- SAP Design, "A New Visual Language Defines Enterprise AI" — https://www.sap.com/design/stories-resources/a-new-visual-language-defines-enterprise-ai
- Agentic Design System case study — https://sibijoseph.com/work/agentic-design-system/
- Medium (Bradley Kinnard), "AI-Powered Web Design in 2026: Kinetic Typography" — https://medium.com/@bradkinnard/ai-powered-web-design-in-2026-kinetic-typography-and-micro-interactions-that-actually-help-178e62d6f45c
- b12, "Bento Grids and Kinetic Typography" — https://www.b12.io/resource-center/website-design/web-design-guide-bento-grids-and-kinetic-typography/
- Behance/Contra case studies (SERP Gram, Cortex AI)

### Shaders and WebGL
- gradients.sh, "React Animated Gradient Libraries" (measured 2026-10-23) — https://gradients.sh/blog/react-animated-gradient-libraries
- Paper Shaders — https://shaders.paper.design/
- React Bits Aurora — https://reactbits.dev/backgrounds/aurora
- Framer Shader Backgrounds — https://www.framer.com/marketplace/components/shader-backgrounds/
- 21st.dev, "React Animated Background Components" — https://21st.dev/blog/react-animated-background-components
- Vanta, "Animated Backgrounds" — https://vanta.supply/articles/animated-backgrounds
- Evil Martians, "Faster WebGL with OffscreenCanvas" — https://evilmartians.com/chronicles/faster-webgl-three-js-3d-graphics-with-offscreencanvas-and-web-workers
- MDN, CSS @property — https://developer.mozilla.org/en-US/docs/Web/CSS/Reference/Properties/--*
- Web.dev, prefers-reduced-motion — https://web.dev/articles/prefers-reduced-motion
- Dev.to, animating CSS gradients — https://dev.to/afif/we-can-finally-animate-css-gradient-kdk
- Josh Comeau, scroll-driven animations — https://www.joshwcomeau.com/animation/scroll-driven-animations/
- OGL comparison — https://webcodingcenter.com/three-js/three-js-and-its-alternatives--ogl-and-choosing-a-library.html
- R3F bundle size recipe — https://gracious-keller-98ef35.netlify.app/docs/recipes/reducing-bundle-size/
- GitHub VFX-JS — https://github.com/fand/vfx-js

### Accessibility
- W3C, Understanding SC 2.3.3 — https://www.w3.org/WAI/WCAG22/Understanding/animation-from-interactions.html
- W3C, Understanding 2.2.2 — https://www.w3.org/TR/WCAG22/#pause-stop-hide
- W3C, C39 (CSS technique) — https://www.w3.org/WAI/WCAG22/Techniques/css/C39
- W3C, Understanding 2.4.11 — https://www.w3.org/WAI/WCAG22/Understanding/focus-not-obscured-minimum.html
- W3C, Understanding 2.4.13 — https://www.w3.org/WAI/WCAG22/Understanding/focus-appearance.html
- MDN, prefers-reduced-motion — https://developer.mozilla.org/en-US/docs/Web/CSS/Reference/At-rules/@media/prefers-reduced-motion
- Motion.dev, Accessibility — https://motion.dev/docs/react-accessibility
- OpenReplay, prefers-reduced-motion — https://blog.openreplay.com/prefers-reduced-motion-accessible-animation/
- Section 508, Accessible Web Design Guide — https://www.section508.gov/develop/guide-accessible-web-design-development/
- Level Access, Accessibility for Developers — https://www.levelaccess.com/blog/accessibility-for-developers-the-essential-guide/
- Vision Australia, Focus Management — https://www.visionaustralia.org/business-consulting/digital-access/blog/making-dynamic-websites-accessible-using-focus-management
- AudioEye, Accessible Landing Pages — https://www.audioeye.com/post/best-practices-for-accessible-landing-pages/
- Medium, useReducedMotion 2026 — https://medium.com/@wul55267/react-usereducedmotion-hook-respect-prefers-reduced-motion-2026-7cc7525b87ce
- GitHub wcag3 issue #183 — https://github.com/w3c/wcag3/issues/183
- Reddit accessibility discussion — https://www.reddit.com/r/accessibility/comments/1ovxv7x/

### Fastest ship path
- Next.js static export docs — https://nextjs.org/docs/app/building-your-application/deploying/static-exports
- Tailwind v4 release — https://tailwindcss.com/blog/tailwindcss-v4-0
- shadcn color guide — https://ui.shadcn.com/docs/theming
- Evil Martians, OKLCH in CSS — https://evilmartians.com/chronicles/oklch-in-css-why-quit-rgb-hsl
- Nikhil Nishad, production guide — (referenced in findings)
- Velora UI comparison — https://velora.colorlib.com/compare
- Next.js 15 async params — (Next.js docs)

### General
- Spell UI, "25 Best SaaS Landing Pages" — https://spell.sh/blog/best-saas-landing-pages
- 2POINT Agency, "Website Animation Guide 2026" — https://www.2pointagency.com/guides/website-animation-the-complete-guide-to-motion-performance-and-ux-in-2026/
- YouTube, "Top Website Animations 2026" — https://www.youtube.com/watch?v=AuOUS-mMqEE
- A1 Gallery, "Scroll Animation Landing Pages" — https://www.a1.gallery/websites/scroll-animation-landing
- Plinth Studio, "SaaS Landing Page Best Practices" — https://plinthstudio.dev/saas-landing-page-best-practices
- 2POINT, "Website Animation" — https://www.2pointagency.com/guides/website-animation-the-complete-guide-to-motion-performance-and-ux-in-2026/
- Josh W Comeau, "Scroll-Driven Animations" — https://www.joshwcomeau.com/animation/scroll-driven-animations/

---

## 9. Appendices

### Appendix A: Bundle size reference table

| Option | Renderer | Gzipped JS | Notes |
|---|---|---|---|
| Paper Shaders MeshGradient | WebGL2 | 8.4 KB | Apache-2.0, zero deps, pin version (0.0.x breaking) |
| React Bits Aurora | WebGL2 | ~18 KB | No deps, reduced-motion check, pauses off-screen |
| Hazeglow | WebGL2 | 14.0 KB | MIT, prefers-reduced-motion check, onUnsupported fallback |
| NEAT (FireCMS) | WebGL | 27.2 KB | MIT + Commons Clause |
| OGL | WebGL | ~14-33 KB | Numbers disagree by measurement method |
| gradients.sh (MeshGradient) | WebGPU→WebGL2 | 52.3 KB | MIT free tier, CSS fallback |
| InstantShader Flow | WebGL | 58.7 KB | MIT, 11 shader components |
| three.js | WebGL | 139 KB | Heavy; R3F aliasing can reduce to 50-80 KB |
| ShaderGradient | three.js + R3F | 293.6 KB | Too heavy for marketing page |
| Aceternity CSS blur | CSS | 0 KB | Per-frame repaint cost |
| Motion (LazyMotion + m) | — | ~4.6 KB initial | Hardware-accelerated, MIT |
| Motion (full) | — | 75.4 KB min+gz | Cannot tree-shake below ~43 KB |
| Motion One base | — | 4 KB + 2 KB React | Lightweight alternative |

### Appendix B: Accessibility score comparison

| Library | Reduced-motion handling | ARIA coverage | Notes |
|---|---|---|---|
| Velora UI | 89/100 | 32/100 | Best among free libraries; still needs work |
| Magic UI | 6/78 | Not reported | Accessibility risk; must add manually |
| Aceternity UI | 4/112 | 11/112 | Severe accessibility gap |
| shadcn/ui | Via Radix primitives | Via Radix primitives | Good (primitives handle a11y) |
| Animata | Not reported | Not reported | Claims reduced-motion included; verify |
| React Aria | Best-in-class | Best-in-class | Our existing primitive layer |

### Appendix C: WCAG criteria reference for motion-heavy landing pages

| Criterion | Level | Requirement |
|---|---|---|
| 2.3.3 Animation from Interactions | AAA | Motion triggered by interaction can be disabled unless essential |
| 2.2.2 Pause, Stop, Hide | A | Auto-moving content >5s needs pause/stop/hide mechanism |
| 2.3.1 Three Flashes or Below Threshold | A | No content flashes >3 times/sec |
| 2.4.7 Focus Visible | AA | Keyboard focus must be visible |
| 2.4.11 Focus Not Obscured (Minimum) | AA | Focused element must remain at least partially visible |
| 2.4.13 Focus Appearance | AAA | Concrete size/contrast for focus indicators |
| 2.4.1 Bypass Blocks | A | Skip-to-main-content link |
| 2.1.1 Keyboard | A | All interactive elements operable via keyboard |

### Appendix D: Recommended stack summary

| Layer | Choice | Rationale |
|---|---|---|
| Framework | Next.js 15 App Router, `output: 'export'` | Static, zero cold start, CDN-hostable |
| Styling | Tailwind v4, CSS-first `@theme` | OKLCH tokens, auto content detection |
| Components (shell) | Velora UI (MIT) | Free blocks, reduced-motion built in, complete CLI installs |
| Components (hero/motion) | Aceternity UI (free tier) | Cinematic effects, shaders, bento |
| Components (micro-interactions) | Magic UI (MIT) | Animated beams, counters, marquee, shimmer |
| Motion | Motion (framer-motion) | Hardware-accelerated, MIT, tree-shakeable via LazyMotion |
| Smooth scroll | Lenis (gated on reduced-motion) | 60 FPS, accessibility-safe |
| Fonts | `next/font/google` self-hosted | Eliminates external roundtrips |
| Theme | Blocking `<head>` script + `suppressHydrationWarning` | No FOUC |
| Shader (hero) | Paper Shaders MeshGradient (8.4 KB) or CSS `@property` | Lightweight, reduced-motion-aware |
| Accessibility | Manual: reduced-motion, focus management, screen readers, skip link | Free libraries don't do this |

---

*This report synthesizes 8 parallel research subagents (deep depth), verified by a dedicated adversarial verification pass. All claims are sourced; limitations and corrections are explicitly stated. No fabricated claims from subagent findings were carried into the final report without flagging.*