# Refweir — Design system for the forge and docs

The product (`web/`) and the documentation (`site/`) keep one cohesive palette, one
typeface, and one set of surfaces.

Per the decision on 2026-11-11, **the product does not maintain an independent marketing landing site.**
Only two frontend surfaces are maintained:
1. **Forge UI (`web/`)**: dense, authenticated Primer/React Aria application embedded directly inside the `refweird` daemon.
2. **Documentation (`site/`)**: technical reference, guides, and architecture documentation powered by Fumadocs, with the root (`/`) redirecting cleanly to `/docs/`.

---

## Source of truth

| Layer | File | Owns |
|---|---|---|
| Product tokens | `web/src/styles/tokens.css` | OKLCH palette, radii, font stacks, motion |
| Product semantics | `web/src/styles/theme.css` | shadcn-style semantic aliases |
| Docs theme | `site/app/global.css` | canvas/text tokens + `--color-fd-*` docs mapping |
| Site fonts | `site/app/layout.tsx` | Mona Sans Variable via `@fontsource-variable/mona-sans` |
| Verification gate | `site/scripts/audit-claims.mjs` | verifiable claims check against `out/docs/index.html` and `CLAIMS.md` |

If a colour changes in the product, it changes in `web/src/styles/tokens.css`
first, then the matching values in `site/app/global.css`.

---

## The docs identity

- Fumadocs tokens derive from the `global.css` palette:
  `--color-fd-background: var(--canvas)`, `--color-fd-primary: var(--patina)`,
  etc. The docs sidebar, TOC, search, and code blocks keep this identity.
- `lib/layout.shared.tsx` routes navigation to `/docs/` and GitHub repository.
- Docs MDX exposes the shared `Kbd` primitive; figures use Fumadocs `Frame`.
- The Fumadocs provider owns `next-themes` (`class` attribute); `.dark`
  flips every token including the `--color-fd-*` variables.

---

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

---

## Guardrails

- No stale screenshots, fake product UI, stock photos, or invented metrics.
- No claim without a `CLAIMS.md` entry; `npm run claims-audit` reads the built
  documentation HTML and must stay green (10 required needles — headline, install/bash
  command, bootstrap command, lease isolation, version, license, ports, MCP
  sandbox — and 19 forbidden phrases). Also enforced in Go via
  `internal/claims/coherence_test.go` over `site/{content,app,components}`.
- Every push gate: `site npm run verify`, `web npm test && npm run build`,
  `go test ./...`.
