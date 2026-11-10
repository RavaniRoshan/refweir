# Refweir — one design system for the forge, the site, and the docs

The product (`web/`) and the documentation (`site/docs`) keep one palette, one
typeface, and one set of surfaces. **The public landing (`site/app/page.tsx`)
has its own approved marketing identity — "Ledger of record" — scoped to its
own CSS module.** Docs and product keep their existing scopes; the landing does
not reach into their tokens, and they do not reach into the landing's.

Approved 2026-11-10 by the owner from three rendered concept routes
(A · Ledger of record — chosen; B · Clearance document; C · Shutdown desk —
rejected). The previous marketing presentation (pale-gray grid canvas, thin
Mona Sans display, coral patina highlight, mono eyebrows, live-style demo
widgets, agent-traffic band) was rejected wholesale and has been removed.

## Source of truth

| Layer | File | Owns |
|---|---|---|
| Product tokens | `web/src/styles/tokens.css` | OKLCH palette, radii, font stacks, motion |
| Product semantics | `web/src/styles/theme.css` | shadcn-style semantic aliases |
| Docs theme | `site/app/global.css` | canvas/text tokens + `--color-fd-*` docs mapping (unchanged) |
| Landing tokens | `site/components/marketing/ledger/ledger.module.css` | scoped "Ledger of record" palette + type roles |
| Landing copy/facts | `site/components/marketing/ledger/facts.ts` | the one verified fact set, traceable to CLAIMS.md |
| Site fonts | `site/app/layout.tsx` | Mona Sans Variable via `@fontsource-variable/mona-sans` (body/UI role on the landing; docs unchanged) |

If a colour changes in the product, it changes in `web/src/styles/tokens.css`
first, then the matching values in `site/app/global.css`. Landing values live
only in the landing module and are independent of the `.dark` flip (the
landing is a fixed light surface; docs keep their theme toggle).

## The landing palette — "Ledger of record"

Warm paper, ink text, oxblood accent, one green for allowed states. Authorized
values (hex, from measured OKLCH); no other hues, no gradients on text, no
grid textures, no glows.

| Token | Value | Role |
|---|---|---|
| `--a-canvas` | `#fbf5eb` | page (warm paper) |
| `--a-slab` | `#fffdf9` | figure sheets |
| `--a-ink` | `#29211a` | primary text |
| `--a-ink2` | `#322820` | body/secondary text |
| `--a-ox` | `#710909` | accent + CTA + denied text |
| `--a-ox-hover` | `#5c0505` | CTA hover |
| `--a-green` | `#003d01` | allowed-state text only (never a fill) |
| rules | `color-mix(#29211a 26%/12%, transparent)` | hairlines / soft rules |

Contrast was measured (WCAG 2.x, computed directly from the shipped hex
values; script `contrast-hex.mjs` in the plan workspace): ink/canvas 14.59:1 ·
ink2/canvas 13.26:1 · ox/canvas 11.09:1 · ox/slab 11.84:1 ·
CTA text/ox fill 11.42:1 · green/slab 12.39:1 · green/canvas 11.60:1. Status
meaning is never color-only: every row spells out
granted/accepted/denied/opened/passing/merged/revoked in text.

## Landing typography

- **Display (serif role):** `Iowan Old Style, 'Palatino Linotype', Palatino,
  Georgia, 'Times New Roman', serif` — system stack, weight 400, display
  clamp(2.45rem → 4.4rem), line-height 1.04, tracking −0.012em, `text-wrap: balance`.
- **Body/UI role:** Mona Sans Variable (already local; kept after the rendered
  comparison — display serif and body sans roles are distinct identities).
- **Code/technical role:** `ui-monospace, 'SFMono-Regular', 'SF Mono', Menlo,
  Consolas, 'Liberation Mono', monospace` — labels, indexes, run records,
  install block. Mono meta is capped (11–13px), uppercase + 0.08–0.16em
  tracking for statuses.

## Landing layout and content rules

- One 880px narrative column (`.wrap`), sections separated by soft hairlines,
  numbered like a ledger (`fig 1`, `§1…§7`). No cards, no grids, no
  window-chrome simulations, no status rails.
- Landing order: `Nav → Hero → Proof (run record) → Clauses (facts) →
  Install → Footer`. Nothing repeats the lease/merge story per-section.
- The single proof element is fig 1: a static transaction table
  (lease → allowed push → denied protected push → PR → checks → human merge →
  revoke). It is labeled *illustrative* with an explicit "not a live
  connection" caption. The word LIVE is reserved for genuinely live
  connections and does not appear on the page.
- Primary CTA → the on-page install block (verified quickstart commands from
  `scripts/install.sh` + docs: install via
  `raw.githubusercontent.com/RavaniRoshan/refweir/...`, bootstrap with
  `refweird bootstrap --email admin@corp.refweir.dev --namespace acme
  --data-dir ./data`, open `http://localhost:8080`).
- No client-side JavaScript is added for the landing (static server
  components only). Motion: none — no marquees, reveals, carets, or pulses.
  `prefers-reduced-motion` is trivially satisfied.
- Accessibility: skip link, one `h1`, section `aria-labelledby`, natural link
  tab order, in-module `:focus-visible` (2px oxblood, 3px offset).

## The docs identity (unchanged scope)

- Fumadocs tokens derive from the `global.css` palette:
  `--color-fd-background: var(--canvas)`, `--color-fd-primary: var(--patina)`,
  etc. The docs sidebar, TOC, search, and code blocks keep today's identity.
- `lib/layout.shared.tsx` and every marketing/docs link route through
  `asset()` / `basePath` so production URLs stay under `/refweir/`.
- Docs MDX exposes the shared `Kbd` primitive; figures use fumadocs `Frame`.
- The Fumadocs provider owns `next-themes` (`class` attribute); `.dark`
  flips every token including the `--color-fd-*` variables. There is no
  `prefers-color-scheme` CSS branch.

## Product screenshots (docs imagery)

```bash
bash scripts/capture/seed.sh /tmp/refweir-capture 18081
# Playwright: 1440×900 CSS viewport, deviceScaleFactor 2, dark scheme,
# token injected into localStorage, onboarding tour suppressed
```

| Asset | View |
|---|---|
| `site/public/product/forge-home.png` | repository overview + activity feed |
| `site/public/product/forge-agents.png` | agent console / MCP + leases |
| `site/public/product/forge-pr.png` | pull request + checks + merge gate |
| `site/public/forge-pr.png` | docs hero (same capture) |
| `site/public/og-refweir-os.png` | 1200×630 social card (legacy from the previous presentation) |

The landing page itself renders no product imagery; docs keep their captures.

## Accessibility (docs + landing)

- Tabs/docs components come from Fumadocs and keep their keyboard behavior.
- Global `:focus-visible` gives docs a patina ring; the landing module
  overrides it with oxblood inside `.root`.
- Contrast: docs body copy uses `--text-2` on `--canvas`; landing pairs are
  measured above. Both palettes carry text labels for all statuses.

## Guardrails

- No stale screenshots, fake product UI, stock photos, invented metrics, or
  simulated terminals on the landing.
- No claim without a `CLAIMS.md` entry; `npm run claims-audit` reads the built
  HTML and must stay green (10 required needles — headline, install/BASH
  command, bootstrap command, lease isolation, version, license, ports, MCP
  sandbox — 19 forbidden phrases). Also enforced in Go via
  `internal/claims/coherence_test.go` over `site/{content,app,components}`.
- No animation library; the landing adds none.
- Every push gate: `site npm run verify`, `web npm test && npm run build`,
  `go test ./...`.

## Legacy notes

- `site/app/global.css` device classes used only by the deleted 2026-11-10
  marketing widgets (`.window`, `.cmd`, `.gate-item`, `.plate`, `.ruler`,
  motion keyframes, …) were pruned. The surviving `components/ui/*` wrappers
  belong to the Fumadocs/docs scope.
- `og-refweir-os.png` still carries the previous presentation's look; a new
  social card in the approved identity is pending launch work (owner).
