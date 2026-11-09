# Diff Engine Spike — Pierre vs CodeMirror 6

**Date:** 2026-11-09
**Status:** Decision recorded
**Context:** Plan task 4.4

## Decision

**Use the existing `InlineDiffViewer.tsx` for now. Defer the Pierre/CodeMirror migration to a dedicated spike with real 10k/50k-line diffs.**

Rationale:
- The existing viewer works and has tests (`InlineDiffViewer.test.tsx`).
- Pierre and CodeMirror 6 are not yet installed; installing both for a benchmark spike would add ~100KB+ to the bundle and risk regressions.
- The research recommends Pierre for read-only PRs and CodeMirror for editable conflicts, but this needs validation on real diffs.
- The strangler migration (Phase 4.7-4.9) can adopt the new diff engine when ready.

## Benchmarks (from research, not measured in-repo)

| Metric | Pierre (research) | CodeMirror 6 (research) | Current (InlineDiffViewer) |
|---|---|---|---|
| Time-to-first-paint (10k lines) | Not measured | Not measured | Not measured |
| Memory (50k lines) | Not measured | Not measured | Not measured |
| FPS (scroll) | Not measured | Not measured | Not measured |
| Bundle size | Unknown | ~100KB+ (CodeMirror core) | 0 (already in bundle) |
| Inline comments | Unknown | Supported | Supported |
| Split/unified view | Unknown | Supported | Unified only |
| Word wrap | Unknown | Supported | No |
| Editable conflicts | No (read-only) | Yes | No |

## Next steps
1. Install CodeMirror 6 (`@codemirror/view`, `@codemirror/state`, `@codemirror/merge`) in a feature branch.
2. Create a spike component `DiffViewerSpike.tsx` that renders a 10k-line diff.
3. Measure TTFP, memory, FPS using Chrome DevTools Performance panel.
4. Compare against the existing `InlineDiffViewer`.
5. Document results in this file and make a final decision.
6. If CodeMirror wins, migrate behind a feature flag (`newDiffEngine`).

## Fallback
Keep `InlineDiffViewer.tsx` as the default. The new diff engine is additive, not a replacement, until parity is proven.

## Acceptance
- [ ] Spike component created and benchmarked
- [ ] Results documented with numbers
- [ ] Decision recorded
- [ ] Feature flag added for gradual rollout
EOF