# plan.md — Launch & Frontend Rebuild Execution Plan

**Source of truth:** `research/exist-io-ph-launch-frontend-redesign.md`
**Goal:** Reach **Product Hunt overall Top 10** on launch day (target **150–250 points**; stretch Top 3 at **300+**) and ship a radical agents-first frontend redesign.
**Launch model:** Apache-2.0 OSS core + hosted demo/tier.
**Timeline:** 13 weeks from Week 1, with a 1-week buffer. Target window: **early February 2027**.
**Rule:** Tasks are ordered. Do not start a later phase until its gate passes. Tasks marked **P0** are launch blockers; **P1** is launch-critical; **P2** can slip past launch.

---

## How to use this plan

- Work top to bottom. The phase order is the dependency order.
- Each task has an ID, owner role, output, and acceptance criteria.
- Roles are hats, not hires: **Founder/PM**, **Backend**, **Frontend**, **Design**, **DevRel/Content**, **Security/Infra**.
- Keep a single launch dashboard with the metrics in the **Metrics and targets** section.
- When a task is done, check the box and link the artifact (PR, doc, recording, URL).
- If a gate fails, fix it before moving on. Do not accumulate launch debt.

---

## Start here — the next 7 days

These are the only tasks that matter this week:

1. **Decide the name.** Rename, or launch under a distinct product name. Do not launch as `exist.io`.
2. **Open `docs/launch/NAMING.md`.** List 20 candidates, screen domains/handles/trademarks, pick 3 finalists.
3. **Write the one-page positioning.** ICP, job, category, differentiator, proof. Put it in `docs/launch/POSITIONING.md`.
4. **Build the containment demo script.** Agent A succeeds on `refs/heads/agent/*`; Agent B is blocked from `main`; audit shows identity, scope, expiry.
5. **Record a 60–90s screen capture** of that demo with captions.
6. **Stand up a throwaway hosted demo** with a seeded repo and one-click agent run. No mock data.
7. **List 50 ICP targets** for design partners and send the first 10 emails.

If those seven are not done in 7 days, the 13-week plan is already late.

---

## Critical path (must be sequential)

```text
Name decision ──► Rename surfaces ──► Positioning ──► Containment demo ──► Hosted demo
      │                                                                        │
      └──────────────────────────────► Design partners ◄──────────────────────┘
                                             │
IA prototype ──► Usability test ──► IA spec ──► Design system ──► New shell ──► Run detail
                                                                        │
                                              Frontend foundation ◄─────┘
                                                       │
Security/compliance ◄──────────────────────────────────┘
      │
Private beta ──► Pricing decision ──► Launch assets ──► Launch rehearsal ──► LAUNCH
                                                                              │
                                                                       Post-launch
```

**Parallel tracks that can run alongside the critical path:**

- Content engine and community building (Weeks 2–12).
- Security/compliance documentation (Weeks 7–11).
- Pricing interviews (Weeks 9–11).
- Launch asset production (Weeks 10–12).

---

## Phase gates

| Gate | When | Go criteria |
|---|---|---|
| **G0 — Identity** | End of Week 1 | Name locked; repo/module/binary/docs renamed or distinct launch name approved; baseline metrics recorded |
| **G1 — Proof** | End of Week 2 | Containment demo reproducible and recorded; hosted demo live; positioning one-pager approved |
| **G2 — Audience** | End of Week 4 | 5 design partners signed; 250+ waitlist/engaged devs; 250+ GitHub stars |
| **G3 — IA** | End of Week 6 | 3 prototypes tested with 8–12 devs; IA spec signed off; design tokens/components specified |
| **G4 — Shell** | End of Week 9 | New shell + run detail + fleet + review behind a flag; no regressions; budgets met |
| **G5 — Trust** | End of Week 10 | Security page, threat model, air-gap/backup docs, SBOM; pricing decision by Week 11 |
| **G6 — Beta** | End of Week 12 | 20–30 beta teams; activation instrumented; top 5 friction points fixed |
| **G7 — Launch-ready** | End of Week 12 | PH assets, demo load test, war-room runbook, 1,000+ engaged devs, 5k stars |
| **G8 — Launch** | Week 13 | Live at 12:01 AM PT; war room staffed; reply to every comment |
| **G9 — Post-launch** | Weeks 14–16 | Retrospective, conversion, case studies, next-channel plan |

---

## Phase 0 — Identity and stop-ship decisions (Week 1)

### 0.1 Name decision — P0
**Owner:** Founder/PM
**Output:** `docs/launch/NAMING.md` with a decision record.
**Tasks:**
- [x] Present the collision facts: `exist.io` is a 13-year-old life-tracking product; domain owned by Hello Code Pty Ltd; repo has 0 stars; Go module is `exist.io`. → `docs/launch/NAMING.md` §1
- [x] Choose one: **rename** (recommended) or **distinct product name for launch** with `exist.io` as a legacy redirect. → **Refweir** selected 2026-11-09
- [x] If rename: define the naming criteria (2–3 syllables, searchable, `.com`/`.dev`, GitHub org, handles, PH slug, no software/SaaS trademark). → §3
- [x] Generate 20 candidates; screen domains, handles, USPTO/EUIPO/IP Australia, GitHub org, PH slug. → 76 screened; domains/GitHub/npm/PyPI checked, trademark pending counsel (§4–§5, Appendix A)
- [x] Pick 3 finalists, then 1. Record the rationale. → finalists **Refweir**, **Refseal**, **Refhaven**; **Refweir** selected (§6–§7)
- [ ] Get legal counsel on the finalist before committing.

**Acceptance:** A signed decision; no public launch asset uses `exist.io` as the product name.

### 0.2 Rename surfaces — P0
**Owner:** Backend + Frontend + DevRel
**Depends on:** 0.1
**Output:** One PR (or a small series) that renames every surface.
**Checklist:**
- [x] `go.mod` module path → `refweir.dev`; all imports updated.
- [x] `cmd/existd` → `cmd/refweird`; `cmd/exist` → `cmd/refweir`.
- [x] Binary names, Docker image, compose service, Helm chart/release names.
- [x] `scripts/install.sh`, `Dockerfile`, `docker-compose.enterprise.yml`, `deploy/helm/refweir-enterprise`.
- [x] Env vars (`REFWEIR_*`), `EXISTD_PATH` → `REFWEIRD_PATH`; data dir/DB filename unchanged.
- [x] `README.md`, `AGENTS.md`, `ROADMAP.md`, `CHANGELOG.md`, `LICENSE` references.
- [x] `site/` metadata, `assets/refweir-logo.svg`, `web/index.html`.
- [x] `docs/` links, badges, screenshots, install commands.
- [ ] GitHub repo/org rename, social handles, PH slug, npm registry reservations — **manual, outside the repo**.
- [x] Add redirects from old doc URLs; `MIGRATION.md` added.
- [ ] Re-render explainer videos and terminal demos (rendered footage still shows the old name).

**Acceptance:** `go test ./...`, `npm test`, `npm run build`, and `cd site && npm run verify` pass. `grep -ri "exist.io"` returns only historical/redirect references.

### 0.3 Baseline metrics — P0
**Owner:** Founder/PM
**Output:** `docs/launch/METRICS.md` updated with a baseline table.
**Tasks:**
- [ ] Record GitHub stars, forks, clones, installs, waitlist, demo runs, activation, retention.
- [ ] Add UTM conventions for PH, HN, Reddit, X, newsletters.
- [ ] Create a single dashboard (sheet or Grafana) with weekly snapshots.
- [ ] Set the launch target: Top 10 overall, 150–250 points; stretch Top 3, 300+.

**Acceptance:** Baseline recorded; dashboard shared; targets agreed.

### 0.4 War-room roles — P1
**Owner:** Founder/PM
**Tasks:**
- [ ] Assign: launch lead, comment lead, demo/ops lead, social lead, metrics lead.
- [ ] Create a private launch channel and a shared runbook.
- [ ] Define decision rights for last-minute changes.

**Acceptance:** Named owners; runbook skeleton exists.

**Gate G0:** Name locked; rename or distinct launch name approved; baseline recorded.

---

## Phase 1 — Proof, demo, and positioning (Weeks 1–2)

### 1.1 Positioning one-pager — P0
**Owner:** Founder/PM + DevRel
**Output:** `docs/launch/POSITIONING.md`
**Structure:**
- **For** [ICP] **who** [job to be done], **[product]** is a [category] **that** [differentiator].
- **Unlike** [alternatives], it [proof].
- Three pillars: per-agent identity/leases; policy + audit at the forge; vendor-neutral MCP/any agent.
- Three proof points with numbers or demos.
- Three anti-claims: not "first," not "GitHub is broken," not "MCP is the moat."
- Objection handling: "why not GitHub?", "MCP already works," "self-hosting is a liability," "another forge?"

**Acceptance:** One page, approved, reused in README, site hero, PH tagline/description, and sales deck.

### 1.2 Containment threat model — P0
**Owner:** Security/Infra + Backend
**Output:** `docs/security/THREAT_MODEL.md`
**Tasks:**
- [x] Define assets: repos, refs, tokens, leases, audit logs, secrets. → §2
- [x] Define adversaries: over-permissioned agent, prompt-injected agent, compromised agent, malicious dependency, insider. → §4
- [x] Map controls: lease scope, ref-prefix isolation, pre-receive hook, read-only mode, secret scanning, audit, revocation, expiry. → §5
- [x] List residual risks and non-goals. → §7–§8

**Acceptance:** Reviewed; controls map to tests or demo steps. → every control in
`docs/security/THREAT_MODEL.md` §5 names a test or demo step; security review
checklist in §10 remains open.

### 1.3 Reproducible containment demo — P0
**Owner:** Backend + DevRel
**Depends on:** 1.2
**Output:** `examples/containment/` + `docs/launch/CONTAINMENT_DEMO.md`
**Scenario:**
1. Bootstrap a demo repo and two agents.
2. Issue Agent A a lease for `refs/heads/agent/feature-123` (scope, expiry, actions).
3. Agent A pushes successfully.
4. Agent B attempts `refs/heads/main` and is rejected.
5. Attempt an expired/revoked lease and show rejection.
6. Show the durable audit log: actor, actor type, lease ID, ref scope, expiry, and rejection reason. Model, prompt hash, and approver are **not** recorded — see the threat-model non-goals.

**Implementation:** `examples/containment/run.sh` + `docs/launch/CONTAINMENT_DEMO.md`.

**Acceptance:** A fresh machine can run `examples/containment/run.sh` and reproduce every step; a 60–90s recording with captions exists. → script passes locally; recording at `docs/launch/assets/containment-demo.mp4` (82s, captions burned in) produced by `docs/launch/record-containment.sh`; re-run after any script change.

### 1.4 Hosted demo — P0
**Owner:** Backend + Security/Infra
**Depends on:** 1.3
**Output:** `deploy/demo/` + a public demo URL
**Tasks:**
- [x] Seed a realistic repo: commits, issues, a PR, a release, an agent run, a lease, an audit trail. → `deploy/demo/seed.sh` (verified locally)
- [x] Add a one-click "Try an agent run" flow; no signup wall for the first run. → `POST /api/v1/demo/agent-run` + `POST /api/v1/demo/merge`, `DemoBanner` buttons; performs a real lease-backed push, PR, and scoped merge
- [x] Reset the demo on a schedule; rate-limit; isolate tenants. → `deploy/demo/reset.sh`, systemd timer, app + nginx limits; single-tenant instance, isolation is the container/host (documented)
- [x] Add a persistent "Demo" banner and a clear "self-host" CTA. → `web/src/components/views/DemoBanner.tsx`, gated by `demoMode` from `/api/v1/system/info`
- [ ] Load-test the demo for launch day; add a status page. → blocked on deployment host
- [x] No mock data; no demo mode in production. → demo mode is a server flag; the web client stays on live APIs; `REFWEIR_DEMO_MODE` is off by default

**Acceptance:** A PH visitor can run an agent, see the diff, review, audit, and merge in under 60 seconds on a fresh phone and desktop. → local run returns a real PR in ~1s and a 429 on the second immediate run; **public URL pending deployment**, so the phone/desktop acceptance is not yet met. See `deploy/demo/README.md` §8 for the pre-announcement checklist.

### 1.5 Pricing hypothesis — P1
**Owner:** Founder/PM
**Output:** `docs/launch/PRICING.md` (hypothesis only)
**Tasks:**
- [x] Draft three packages: free OSS; hosted evaluation; enterprise self-host (SSO/audit/support). → `docs/launch/PRICING.md` §2
- [x] Anchor against GitLab's **$299/user/mo** air-gapped agent SKU, GitHub Enterprise + Copilot, and Gitea Enterprise. → `docs/launch/PRICING.md` §4
- [x] Write the 10–15 interview script (budget owner, current spend, procurement path, WTP). → `docs/launch/PRICING.md` §7
- [x] Do **not** publish prices yet. → rules in `docs/launch/PRICING.md` §8–§9; no price appears on the site.

**Acceptance:** Hypothesis documented; interview script ready. → `docs/launch/PRICING.md`; interviews pending Phase 2.

**Gate G1:** Containment demo reproducible and recorded; hosted demo live; positioning approved.

**G1 status (2026-11-09):** containment demo reproducible and recorded (script + 82s captioned asset); hosted demo scaffolding complete and locally verified, public URL pending infrastructure/DNS/TLS; positioning drafted, approval outstanding.

---

## Phase 2 — Audience and design partners (Weeks 2–8, parallel)

### 2.1 ICP list of 50 targets — P0
**Owner:** Founder/PM + DevRel
**Output:** `docs/launch/DESIGN_PARTNERS.md` with a CRM-style table.
**Status:** `docs/launch/DESIGN_PARTNERS.md` written — 50 named targets across the four segments, qualification rules, outreach cadence, email template, and exit criteria. Contacts are role-based; verify before sending.
**Segments, in order:**
1. Air-gapped/defense/federal/national labs.
2. EU/sovereign platform teams.
3. Agent-heavy platform engineering teams.
4. Cost-sensitive mid-market replacing agent seats/credits.

**Acceptance:** 50 named targets with contact, segment, trigger, and next action.

### 2.2 Design partner outreach — P0
**Owner:** Founder/PM
**Tasks:**
- [ ] Send 10 personalized emails per week; track replies.
- [ ] Offer white-glove install, private channel, roadmap call.
- [ ] Sign 5–8 partners by end of Week 4.
- [ ] Get 3 quotes/logos and one mini-case study with a number by Week 10.

**Acceptance:** 5 signed partners; 3 quotes; 1 case study.

### 2.3 Content engine — P1
**Owner:** DevRel/Content
**Output:** `docs/launch/CONTENT_CALENDAR.md`
**Cadence:** 2 posts/week.
**Topics:**
- Agent containment benchmark (the demo).
- CI cost of agent retry loops.
- Secret leakage in agent-authored commits.
- Air-gap install in 10 minutes.
- MCP setup for Claude Code, Codex, Cursor, Cline.
- Lease model and ref isolation.
- Autonomous PR review that gates merge.
- Migrating from GitHub/Gitea/Forgejo.
- Terraform/Helm/air-gap deployment.

**Channels:** blog, docs, r/selfhosted changelog every 2 weeks, X/LinkedIn, newsletters, Discord/Slack.

**Acceptance:** 16+ posts by launch; changelog cadence maintained; top 3 posts repurposed into launch assets.

### 2.4 Community and stars — P1
**Owner:** DevRel/Content
**Tasks:**
- [ ] Post a useful changelog to r/selfhosted every 2 weeks (not just launch spam).
- [ ] Answer questions in r/devops, r/ExperiencedDevs, MCP/agent communities.
- [ ] Ship a "good first issue" set and a contributing guide.
- [ ] Target **5k GitHub stars** and **1,000–2,000 engaged developers** by launch.

**Acceptance:** Stars/waitlist tracked weekly; 5k stars and 1,000+ engaged devs by Week 12.

### 2.5 Waitlist and newsletter — P1
**Owner:** DevRel/Content
**Tasks:**
- [ ] Add a waitlist to the site and docs; one field, clear value.
- [ ] Send a weekly build-in-public update.
- [ ] Segment: self-host, hosted, design partner.
- [ ] Target 1,000+ waitlist by launch.

**Acceptance:** Waitlist live; weekly send; 1,000+ by Week 12.

**Gate G2:** 5 design partners; 250+ waitlist; 500+ stars. (These are early gates; the launch targets are higher.)

---

## Phase 3 — IA and design system (Weeks 3–6)

### 3.1 Three IA prototypes — P0
**Owner:** Design + Frontend
**Output:** `docs/design/IA.md` + clickable prototypes
**Archetypes to test:**
- **A. Repo-anchored Agents tab** (GitHub-like).
- **B. Agent-first fleet workspace** (Cursor/Codex-like).
- **C. Hybrid**: global Inbox/Runs/Reviews/Policies rail + repo Agents tab + run detail.

**Prototype scope:** shell, fleet table, run detail, diff review, policy/audit, onboarding.

**Acceptance:** Three prototypes clickable; each covers the same five tasks.

### 3.2 Usability tests — P0
**Owner:** Design
**Depends on:** 3.1
**Tasks:**
- [ ] Recruit 8–12 developers who use coding agents daily.
- [ ] Test: find a running agent; review a diff; identify who approved a merge; revoke a lease; find the audit entry.
- [ ] Measure time-on-task, errors, trust rating, preference.
- [ ] Record findings in `docs/design/USABILITY_TEST.md`.

**Acceptance:** Results summarized; a clear winning archetype or a hybrid recommendation with evidence.

### 3.3 IA specification — P0
**Owner:** Design + Frontend
**Depends on:** 3.2
**Output:** `docs/design/IA.md`
**Must define:**
- Sitemap and route map.
- Object model: run, lease, repo, ref, commit, PR, review, check, policy, audit.
- Global rail; repo tabs; run detail tabs.
- States: empty, loading, error, blocked, expired, revoked.
- Keyboard map and command palette actions.
- Permissions and policy verdicts.
- Mobile/responsive behavior.
- Migration mapping from current `?tab=` views.

**Acceptance:** Signed off by Founder/PM + Frontend; no open questions on primary navigation.

### 3.4 Design system extension — P0
**Owner:** Design + Frontend
**Depends on:** 3.3
**Output:** Updated `web/src/styles/tokens.css`, `theme.css`, `typography.css`, `site/DESIGN.md`
**Tasks:**
- [ ] Document Patina ramps in OKLCH; keep "No other hues."
- [ ] Add a signature mono (tabular numerals, Apache-2.0-compatible license) and a display cut; keep Mona Sans for UI.
- [ ] Define semantic tokens: run states, lease states, policy verdicts, diff states, cost, risk.
- [ ] Build signature components: **LeaseSeal**, **MergeGate**, **RunStatus**, **DiffMarker**.
- [ ] Dark and light themes; WCAG 2.2 AA; APCA check; colorblind-safe diff markers.
- [ ] Density toggle: compact 32px rows default, comfortable optional.
- [ ] Motion: 100ms toggles, 200–300ms dialogs, reduced-motion support.
- [ ] No glassmorphism/neon/Linear purple/Geist palette.

**Acceptance:** Tokens and components documented; visual regression tests; accessibility audit passes.

### 3.5 Component inventory and migration map — P1
**Owner:** Frontend
**Tasks:**
- [ ] Inventory current `web/src/components/ui/`, `primitives/`, `views/`, `modals/`.
- [ ] Map each to keep / restyle / replace / delete.
- [ ] Identify the 10 highest-risk components (diff, tables, dialogs, forms, navigation, charts, toasts, command palette, code blocks, empty states).
- [ ] Write the strangler migration order.

**Acceptance:** `docs/design/COMPONENT_MAP.md`; migration order agreed.

**Gate G3:** IA spec signed off; design tokens/components specified; usability results documented.

---

## Phase 4 — Frontend foundation and new shell (Weeks 5–10)

### 4.1 Add TanStack Router — P0
**Owner:** Frontend
**Depends on:** 3.3
**Tasks:**
- [ ] Add `@tanstack/react-router`; create `web/src/routes/` and a route tree.
- [ ] Map existing `?tab=` views to routes with typed search params.
- [ ] Add deep links for repo, PR, commit, run, diff anchor, filters.
- [ ] Keep `App.tsx` as a compatibility wrapper until cutover.
- [ ] Add route-level code splitting.

**Acceptance:** All existing views reachable by URL; back/forward works; no dead links; tests pass.

### 4.2 Add TanStack Query + OpenAPI codegen — P0
**Owner:** Frontend + Backend
**Tasks:**
- [ ] Add `@tanstack/react-query`; create `web/src/lib/queryClient.ts`.
- [ ] Add an OpenAPI spec (`api/openapi.yaml`) covering auth, repos, commits, PRs, reviews, checks, leases, audit, agent activities.
- [ ] Generate a typed client with Hey API; keep the hand-written client as fallback.
- [ ] Use `ensureQueryData` in loaders; one cache; no duplicate server state.
- [ ] Test optimistic updates (merge, comment, lease revoke) and invalidation.

**Acceptance:** ≥80% of used endpoints generated; typed client compiles; no duplicate caches.

### 4.3 SSE for agent activity — P0
**Owner:** Backend + Frontend
**Tasks:**
- [ ] Add `GET /api/v1/agent/activities/stream` (SSE) in `internal/agentactivity` or `internal/app`.
- [ ] Send heartbeats every 15–30s; set `Cache-Control: no-cache`, `X-Accel-Buffering: no`.
- [ ] Add `web/src/lib/events.ts` using `@microsoft/fetch-event-source` (auth headers, POST, retry).
- [ ] Handle reconnect, backoff, dedupe, and fallback to polling.
- [ ] Add tests for stream ordering and reconnect.
- [ ] Document Nginx/Cloudflare proxy settings.

**Acceptance:** Live run updates appear without refresh; reconnect works through a proxy; fallback works.

### 4.4 Diff engine spike — P0
**Owner:** Frontend
**Tasks:**
- [ ] Benchmark **Pierre** and **CodeMirror 6 merge view** on real 1k/10k/50k-line diffs.
- [ ] Measure time-to-first-paint, memory, FPS, bundle size, inline-comment support, split/unified, word wrap.
- [ ] Compare against current `InlineDiffViewer.tsx`.
- [ ] Decide: Pierre for read-only PRs; CodeMirror for editable conflicts; keep a feature flag.
- [ ] Document results in `docs/design/DIFF_SPIKE.md`.

**Acceptance:** Decision recorded with numbers; fallback path exists; no regression on 10k-line diffs.

### 4.5 New app shell — P0
**Owner:** Frontend + Design
**Depends on:** 4.1, 3.4
**Tasks:**
- [ ] Create `web/src/app/AppShell.tsx` with global rail (Inbox, Runs, Reviews, Policies, Repos), top bar, command palette.
- [ ] Add namespace/cluster switcher, agent activity indicator, notifications, theme toggle.
- [ ] Add keyboard shortcuts and a help overlay.
- [ ] Add authentic empty states and onboarding.
- [ ] Feature flag: `newShell`.

**Acceptance:** Shell usable; old routes mount inside; no layout shift; keyboard navigation works.

### 4.6 Run detail, fleet, inbox, review — P0
**Owner:** Frontend
**Depends on:** 4.5
**Tasks:**
- [ ] Run detail: header (state, repo, ref, SHA, lease, policy, cost), tabs Overview/Changes/Checks/Trace/Audit, action bar.
- [ ] Fleet: virtualized table, filters, saved views, bulk actions, cost/duration columns.
- [ ] Inbox: needs-me, failing checks, unreviewed, over-budget.
- [ ] Review: queue sorted by risk, diff-first, inline comments, merge gate.
- [ ] Stable run ID resolving to commit/PR/diff.
- [ ] Chat as annotation; diff as truth.

**Acceptance:** End-to-end: agent run → diff → review → merge; 50+ runs usable at 60fps.

### 4.7 Migrate read-only routes — P1
**Owner:** Frontend
**Tasks:**
- [ ] Repo overview, commits, issues, releases, profile, search.
- [ ] Keep old views until parity; visual regression and Playwright tests.
- [ ] Delete old views only after two clean weeks.

**Acceptance:** Parity checklist complete; no regressions.

### 4.8 Testing, performance, accessibility — P0
**Owner:** Frontend
**Tasks:**
- [ ] Vitest + Testing Library + MSW for new shell/run detail.
- [ ] Playwright for critical flows (login, repo, run, review, merge, lease revoke).
- [ ] Visual regression in pinned CI (no Storybook as source of truth).
- [ ] Route-level JS budgets; lazy-load diff/editor/chart routes.
- [ ] axe/accessibility checks; keyboard-only pass; reduced-motion pass.
- [ ] Measure LCP/INP on low-end hardware.

**Acceptance:** CI green; budgets documented; a11y audit passes; no P0 regressions.

### 4.9 Strangler cutover — P1
**Owner:** Frontend
**Depends on:** 4.7, 4.8
**Tasks:**
- [ ] Enable `newShell` for internal users, then design partners, then all.
- [ ] Keep a rollback branch and a kill switch.
- [ ] Remove the old shell after two clean weeks.
- [ ] Update docs/screenshots.

**Acceptance:** All users on the new shell; rollback tested; old code removed or archived.

**Gate G4:** New shell + run detail + fleet + review behind a flag; no regressions; budgets met.

---
## Phase 5 — Security, compliance, and pricing (Weeks 7–11)

### 5.1 Security pack — P0
**Owner:** Security/Infra
**Output:** `SECURITY.md`, `docs/security/`
**Tasks:**
- [ ] Threat model (from 1.2) reviewed and published.
- [ ] Security policy, disclosure process, supported versions, SLA.
- [ ] Air-gap install guide: no internet, offline license, local models, no telemetry.
- [ ] Backup/restore guide: SQLite WAL, bare repos, LFS CAS, encryption, restore drill.
- [ ] Upgrade guide: schema migrations, rollback, compatibility.
- [ ] HA/DR runbook: replication, failover, snapshots, integrity checks.
- [ ] SSO/SCIM setup guide with a test IdP.
- [ ] Audit log documentation: events, retention, export, `actor_is_agent` equivalent.
- [ ] Known limitations and non-goals.

**Acceptance:** A skeptical platform engineer can deploy, back up, restore, upgrade, and audit without asking a question.

### 5.2 Supply chain — P0
**Owner:** Security/Infra
**Tasks:**
- [ ] Generate an SBOM per release (CycloneDX or SPDX).
- [ ] Sign releases and checksums (cosign/sigstore or GPG).
- [ ] Add `govulncheck`, `npm audit`, and dependency review to CI.
- [ ] Pin GitHub Actions by SHA; least-privilege tokens.
- [ ] Publish checksums and verification instructions.
- [ ] Add a `SECURITY.md` contact and a 90-day disclosure policy.

**Acceptance:** A release can be verified end-to-end; CI fails on known critical vulns.

### 5.3 Trust page — P0
**Owner:** Founder/PM + Security/Infra
**Output:** `site/` trust/security page or `docs/security/TRUST.md`
**Tasks:**
- [ ] State what is done: signed releases, SBOM, audit, SSO, air-gap, encryption, backups.
- [ ] State what is in progress: SOC 2 readiness, pen test, FedRAMP gap analysis.
- [ ] State what is not claimed: no SOC 2 certificate yet, no FedRAMP ATO, no ISO 27001.
- [ ] Publish a security questionnaire pre-answer.
- [ ] Add a responsible-disclosure link.

**Acceptance:** No overclaiming; a procurement reviewer can self-serve the basics.

### 5.4 Pricing interviews — P0
**Owner:** Founder/PM
**Depends on:** 1.5, 2.2
**Tasks:**
- [ ] Interview 10–15 ICP buyers (not users only).
- [ ] Ask: current spend, budget owner, procurement path, must-have controls, WTP at 3 price points.
- [ ] Test packaging: free OSS; hosted evaluation; enterprise self-host (SSO/audit/support); per-seat vs per-agent vs per-run.
- [ ] Anchor below GitLab **$299/user/mo** air-gap; compare GitHub Enterprise + Copilot and Gitea Enterprise.
- [ ] Record the decision and the evidence in `docs/launch/PRICING.md`.

**Acceptance:** A pricing/packaging decision with 3+ consistent WTP signals; no public price until then.

### 5.5 Legal and privacy — P1
**Owner:** Founder/PM
**Tasks:**
- [ ] Privacy policy, terms, DPA, subprocessors, data residency.
- [ ] Trademark clearance for the final name (US, EU, UK, Australia).
- [ ] Open-source license review (Apache-2.0 + dependencies).
- [ ] Export/air-gap compliance note if targeting defense/federal.

**Acceptance:** Counsel-reviewed; linked from the site and PH listing.

**Gate G5:** Security page, threat model, air-gap/backup docs, SBOM, pricing decision.

---

## Phase 6 — Private beta and launch assets (Weeks 9–12)

### 6.1 Private beta — P0
**Owner:** Founder/PM + Backend
**Depends on:** G4 (Week 9) + security pack (Week 10)
**Tasks:**
- [ ] Start with design partners in Week 9; expand to 20–30 teams from the waitlist by Week 12.
- [ ] White-glove onboarding: install, SSO, first repo, first agent, first lease, first merge.
- [ ] Instrument activation events and a feedback form.
- [ ] Weekly beta call; shared channel; public issue tracker.
- [ ] Fix the top 5 friction points before launch.
- [ ] Publish a "known limitations" page.

**Acceptance:** 20+ teams onboarded; activation instrumented; top 5 fixes shipped.

### 6.2 Activation metrics — P0
**Owner:** Founder/PM + Backend
**Tasks:**
- [ ] Define activation: first repo + first agent run + first lease + first merge within 24h.
- [ ] Add events and a dashboard.
- [ ] Set a target activation rate (e.g., 40% of demo visitors start a run; 25% complete a merge).
- [ ] Review weekly; fix the biggest drop-off.

**Acceptance:** Dashboard live; target agreed; weekly review cadence.

### 6.3 Product Hunt listing — P0
**Owner:** DevRel/Content + Design
**Depends on:** G1, G5
**Output:** Updated `docs/launch/` assets
**Tasks:**
- [ ] Name ≤40, tagline ≤60, description ≤500.
- [ ] Logo 240×240; gallery 5 images at 1270×760; video 60–90s with captions.
- [ ] Up to 3 tags; primary "Developer Tools"; pricing field; open-source badge; repo link.
- [ ] First comment ≤5,000: who, why, what, who for, differentiator, one question, offer.
- [ ] Demo link; no signup wall for the first run.
- [ ] Claim the page; create the Forum Thread (Coming Soon pages are retired).
- [ ] Schedule ≤30 days ahead; day starts 12:01 AM PT.

**Acceptance:** Listing reviewed by 5 outsiders; first image answers "what does this do?" in 2 seconds; all links work.

### 6.4 Landing and docs — P0
**Owner:** Design + DevRel
**Tasks:**
- [ ] Update `site/` hero, product preview, proof section, CTA, pricing/trust links.
- [ ] Update `README.md` with the new name, one-liner, demo, quickstart, MCP config.
- [ ] Update docs: install, air-gap, MCP, leases, review, audit, migration.
- [ ] Add a "Why self-hosted?" page and a "vs GitHub/GitLab/Gitea" comparison (factual, no FUD).
- [ ] Add a changelog and a roadmap page.

**Acceptance:** `cd site && npm run verify` green; all links valid; claims audit passes.

### 6.5 Launch video and screenshots — P0
**Owner:** Design + DevRel
**Tasks:**
- [ ] 60–90s video: hook in 3–5s, real UI, captions, end CTA.
- [ ] Five screenshots: hero (run/lease/diff/merge), fleet, proof (blocked `main` + audit), self-host/air-gap, CTA.
- [ ] Dark UI on warm neutral surround; one idea per frame; legible at 240×240.
- [ ] Capture at 2x; export PNG; compress.
- [ ] Store in `assets/launch/`; update `docs/launch/SCREENSHOTS.md`.

**Acceptance:** Assets approved; mobile and thumbnail legibility tested.

### 6.6 Press, newsletters, communities — P1
**Owner:** DevRel/Content
**Tasks:**
- [ ] Build a list of 30 newsletters, 20 communities, 10 podcasts, 10 journalists.
- [ ] Personalize 50 outreach messages; no mass BCC.
- [ ] Prepare an embargoed press kit: logo, screenshots, video, facts, quotes.
- [ ] Line up 3–5 friendly amplifiers who will share on launch day.
- [ ] Do not pay for votes or use a paid hunter.

**Acceptance:** 50 personalized outreaches; 5 amplifiers confirmed; press kit ready.

### 6.7 Launch rehearsal — P0
**Owner:** Founder/PM + Security/Infra
**Tasks:**
- [ ] Load-test the hosted demo and API.
- [ ] Test the signup, install, MCP, lease, review, merge flows on fresh devices.
- [ ] Verify status page, alerting, rollback, and backups.
- [ ] Write comment reply templates for 20 likely questions.
- [ ] Run a 60-minute dry run with the war room.
- [ ] Freeze code 24 hours before launch; only critical fixes.

**Acceptance:** Dry run passes; no P0 bugs; war room knows its roles.

**Gate G6:** 20–30 beta teams; activation instrumented; top 5 friction fixed.
**Gate G7:** PH assets, demo load test, war-room runbook, 1,000+ engaged devs, 5k stars.

---

## Phase 7 — Launch week (Week 13)

### 7.1 T-14 days
- [ ] Freeze name, listing, screenshots, video, demo.
- [ ] Confirm PH date; schedule the launch.
- [ ] Brief 30–50 supporters individually; no vote-buying pacts.
- [ ] Confirm newsletter/community slots.
- [ ] Finalize the war-room roster and runbook.

### 7.2 T-7 days
- [ ] Test the demo on fresh devices, slow networks, and mobile.
- [ ] Prepare HN Show HN, Reddit posts, X/LinkedIn threads.
- [ ] Confirm 5 amplifiers.
- [ ] Check status page, alerts, on-call.
- [ ] Run a second dry run.

### 7.3 T-1 day
- [ ] Final QA: links, pricing, install command, MCP snippet, PH fields.
- [ ] Schedule the maker first comment.
- [ ] Prepare a launch-day dashboard.
- [ ] Sleep; no late-night deploys.

### 7.4 T-0 — 12:01 AM PT
| Time (PT) | Action |
|---|---|
| 12:01 AM | Go live; post the maker comment; email the list. |
| 12:15 AM | X/LinkedIn threads; community posts. |
| 6:00–9:00 AM | Show HN; Reddit posts; second outreach wave. |
| 9:00 AM–12:00 PM | Reply to every comment within 15 minutes; monitor demo. |
| 12:00–5:00 PM | Third outreach wave; address objections; post updates. |
| 5:00–11:59 PM | Final replies; capture metrics; thank supporters. |

**Rules:**
- Reply to every comment; ask for feedback, never upvotes.
- No paid hunter; no vote buying; no "upvote for upvote."
- Fix demo issues live; communicate status.
- Do not change the listing mid-day unless it is broken.

### 7.5 T+1 to T+7
- [ ] Thank-you post and transparent retrospective draft.
- [ ] Convert waitlist; follow up with every commenter who asked a question.
- [ ] Publish a launch recap with numbers.
- [ ] Run a webinar/live demo.
- [ ] Publish a design-partner case study.
- [ ] Plan the next channel push (HN/Reddit/newsletter), not another PH launch immediately.

**Gate G8:** Live at 12:01 AM PT; war room staffed; every comment answered.

---

## Phase 8 — Post-launch (Weeks 14–16)

- [ ] Publish a transparent retrospective: rank, points, referrals, signups, stars, installs, activation, retention.
- [ ] Interview 10 launch-day signups; fix the top 3 activation blockers.
- [ ] Publish 2 case studies with numbers.
- [ ] Run a monthly agent-containment benchmark; make it a recurring content asset.
- [ ] Keep the changelog cadence; do not disappear after launch.
- [ ] Decide the next bet: hosted tier, enterprise air-gap, or ecosystem integrations.
- [ ] Update the roadmap with what was learned.
- [ ] If Top 10 was missed, run a second coordinated push (HN/Reddit/newsletter) rather than a duplicate PH launch.

**Gate G9:** Retrospective published; conversion/retention baseline; next 90-day plan.

---

## Metrics and targets

| Metric | Baseline (Week 1) | Launch target | Source |
|---|---|---|---|
| GitHub stars | 0 | **5,000** | GitHub API |
| Engaged developers / waitlist | 0 | **1,000–2,000** | Waitlist + newsletter |
| Design partners | 0 | **5–8** | CRM |
| Hosted demo runs | 0 | **250+ pre-launch; 1,000+ launch week** | Analytics |
| PH points | — | **150–250** (stretch 300+) | PH leaderboard |
| PH rank | — | **Overall Top 10** | PH leaderboard |
| PH referrals | — | **1,000–2,500** | UTM |
| Signups | — | **50–150** (more with a large list) | Auth |
| Activation (repo → run → lease → merge) | — | **25–40%** of demo visitors | Product analytics |
| 30-day retention | — | **40%+** of activated teams | Product analytics |
| Install → first repo | — | **< 10 minutes** | Telemetry/opt-in |
| Docs → successful quickstart | — | **> 60%** | Docs analytics |

**Cadence:** weekly snapshot every Monday; launch-day dashboard refreshed hourly.

---

## Risks and mitigations

| Risk | Likelihood | Impact | Mitigation | Owner |
|---|---|---|---|---|
| Name/trademark dispute | High | High | Rename before launch; counsel; distinct launch name fallback | Founder/PM |
| PH does not feature the product | Medium | High | Submit early; make the demo high-craft; run HN/Reddit as primary; direct-link campaign | DevRel |
| Demo fails on launch day | Medium | High | Load test; status page; fallback video; on-call; rate limits | Security/Infra |
| Not enough audience for Top 10 | High | High | 12-week audience build; design partners; newsletters; HN/Reddit; 5k stars | Founder/PM |
| GitHub bundles agent features | High | Medium | Emphasize air-gap, audit, vendor neutrality, per-agent leases; target regulated buyers | Founder/PM |
| Gitea/Forgejo ship free agent identity | Medium | Medium | Speed; enterprise support; compliance; proof | Backend |
| Security wedge dismissed as "config" | Medium | High | Reproducible containment demo; third-party validation; threat model | Security/Infra |
| No SOC 2/FedRAMP at launch | High | Medium | Target buyers who accept questionnaires; publish roadmap; trust page; no overclaim | Security/Infra |
| Frontend rewrite slips | Medium | High | Strangler migration; feature flags; cut P2; keep old views until parity | Frontend |
| Diff engine immature (Pierre) | Medium | Medium | CodeMirror fallback; benchmark on real diffs; feature flag | Frontend |
| Pricing is wrong | Medium | High | 10–15 interviews; do not publish until validated; flexible packaging | Founder/PM |
| Small team bandwidth | High | High | Ruthless P0 focus; cut P2; one launch, not three | Founder/PM |
| Launch-day low engagement | Medium | Medium | War room; amplifier list; paid ads only for awareness; extend campaign | DevRel |

---

## What not to do

- Do not launch as `exist.io`.
- Do not claim "first Git forge for agents."
- Do not claim "GitHub agents bypass branch protection" without qualification.
- Do not buy votes or pay a hunter.
- Do not ship mock data or demo mode in production.
- Do not replace pull requests with a novel object on day one.
- Do not hide the diff behind chat.
- Do not adopt Linear purple, Geist scales, neon gradients, or glassmorphism.
- Do not rewrite the frontend greenfield.
- Do not make PH the only channel.
- Do not publish enterprise pricing without WTP data.
- Do not overclaim security.
- Do not schedule the launch before G7 passes.

---

## File map — where the work lands

| Path | Purpose | Phase |
|---|---|---|
| `research/exist-io-ph-launch-frontend-redesign.md` | Research source of truth | 0 |
| `docs/launch/NAMING.md` | Name candidates, screens, decision | 0 |
| `docs/launch/POSITIONING.md` | One-pager, ICP, proof, objections | 1 |
| `docs/security/THREAT_MODEL.md` | Assets, adversaries, controls, residual risk | 1 |
| `examples/containment/` | Reproducible containment demo | 1 |
| `docs/launch/CONTAINMENT_DEMO.md` | Demo script and recording notes | 1 |
| `deploy/demo/` | Hosted demo compose/config/seed | 1 |
| `docs/launch/PRICING.md` | Pricing hypothesis and interview evidence | 1, 5 |
| `docs/launch/DESIGN_PARTNERS.md` | ICP list, outreach, quotes | 2 |
| `docs/launch/CONTENT_CALENDAR.md` | 2 posts/week, channels, owners | 2 |
| `docs/design/IA.md` | Sitemap, object model, states, keyboard map | 3 |
| `docs/design/USABILITY_TEST.md` | Test plan and findings | 3 |
| `docs/design/COMPONENT_MAP.md` | Keep/restyle/replace/delete | 3 |
| `docs/design/DIFF_SPIKE.md` | Pierre/CodeMirror benchmark and decision | 4 |
| `web/src/routes/` | TanStack Router route tree | 4 |
| `web/src/app/AppShell.tsx` | New shell | 4 |
| `web/src/lib/queryClient.ts` | TanStack Query setup | 4 |
| `web/src/lib/events.ts` | SSE client | 4 |
| `web/src/components/primitives/LeaseSeal.tsx` | Signature lease component | 3 |
| `web/src/components/primitives/MergeGate.tsx` | Signature merge gate | 3 |
| `internal/agentactivity/stream.go` | SSE endpoint | 4 |
| `api/openapi.yaml` | OpenAPI spec for codegen | 4 |
| `SECURITY.md` | Disclosure policy and supported versions | 5 |
| `docs/security/AIRGAP.md` | Air-gap install and operation | 5 |
| `docs/security/BACKUP_RESTORE.md` | Backup/restore drill | 5 |
| `docs/security/SBOM.md` | SBOM and signing | 5 |
| `docs/security/TRUST.md` | Trust page content | 5 |
| `docs/launch/SCREENSHOTS.md` | 5-frame gallery plan | 6 |
| `docs/launch/LAUNCH_RUNBOOK.md` | T-14 to T+7, roles, templates | 6, 7 |
| `docs/launch/BETA_TESTING.md` | Beta cohort and feedback | 6 |
| `docs/launch/RETROSPECTIVE.md` | Post-launch numbers and lessons | 8 |

---

## Definition of done — launch-ready checklist

**Identity**
- [ ] Name locked; all surfaces renamed or distinct launch name approved.
- [ ] Trademark/domain/handles screened; counsel sign-off.
- [ ] No public asset uses `exist.io` as the product name.

**Proof**
- [ ] Containment demo reproducible on a fresh machine.
- [ ] 60–90s recording with captions.
- [ ] Hosted demo live, load-tested, no mock data.
- [ ] Threat model and security page published.

**Audience**
- [ ] 5–8 design partners; 3 quotes/logos.
- [ ] 1,000–2,000 engaged developers / waitlist.
- [ ] 5,000 GitHub stars.
- [ ] 16+ content posts; r/selfhosted changelog cadence.

**Product**
- [ ] New shell + run detail + fleet + review behind a flag.
- [ ] Diff performance meets budget on 10k/50k lines.
- [ ] SSE live updates with reconnect and fallback.
- [ ] Activation instrumented; 20–30 beta teams onboarded.
- [ ] Top 5 friction points fixed.

**Trust**
- [ ] SBOM, signed releases, checksums.
- [ ] Air-gap, backup/restore, upgrade, HA/DR docs.
- [ ] Security questionnaire and disclosure policy.
- [ ] Trust page with done/in-progress/not-claimed.

**Commercial**
- [ ] Pricing validated with 10–15 interviews.
- [ ] Packaging decision documented.
- [ ] Legal/privacy/DPA ready.

**Launch**
- [ ] PH listing approved by 5 outsiders.
- [ ] 5 screenshots, 240×240 logo, 60–90s video.
- [ ] Maker first comment, FAQ, objection handling.
- [ ] 50 personalized press/newsletter/community outreaches.
- [ ] 5 amplifiers confirmed.
- [ ] Launch rehearsal passed; status page and rollback ready.
- [ ] War-room roles and schedule confirmed.

**Post-launch**
- [ ] Retrospective published.
- [ ] Conversion/retention baseline.
- [ ] 2 case studies.
- [ ] Next 90-day plan.

---

## Master sequence (one line per task)

This is the ordered index. Detail is in the phase sections above.

1. Name decision (0.1)
2. Rename surfaces (0.2)
3. Baseline metrics (0.3)
4. War-room roles (0.4)
5. Positioning one-pager (1.1)
6. Containment threat model (1.2)
7. Reproducible containment demo (1.3)
8. Hosted demo (1.4)
9. Pricing hypothesis (1.5)
10. ICP list of 50 (2.1)
11. Design partner outreach (2.2)
12. Content engine (2.3)
13. Community and stars (2.4)
14. Waitlist and newsletter (2.5)
15. Three IA prototypes (3.1)
16. Usability tests (3.2)
17. IA specification (3.3)
18. Design system extension (3.4)
19. Component inventory and migration map (3.5)
20. TanStack Router (4.1)
21. TanStack Query + OpenAPI (4.2)
22. SSE agent activity (4.3)
23. Diff engine spike (4.4)
24. New app shell (4.5)
25. Run detail, fleet, inbox, review (4.6)
26. Migrate read-only routes (4.7)
27. Testing, performance, accessibility (4.8)
28. Strangler cutover (4.9)
29. Security pack (5.1)
30. Supply chain (5.2)
31. Trust page (5.3)
32. Pricing interviews (5.4)
33. Legal and privacy (5.5)
34. Private beta (6.1)
35. Activation metrics (6.2)
36. Product Hunt listing (6.3)
37. Landing and docs (6.4)
38. Launch video and screenshots (6.5)
39. Press, newsletters, communities (6.6)
40. Launch rehearsal (6.7)
41. T-14 launch prep (7.1)
42. T-7 launch prep (7.2)
43. T-1 final QA (7.3)
44. Launch day (7.4)
45. T+1 to T+7 follow-up (7.5)
46. Post-launch retrospective (8)
47. Conversion and retention fixes (8)
48. Case studies and benchmark (8)
49. Next 90-day plan (8)

**The only three things that can kill the launch:** the name, the demo, and the audience. Everything else is recoverable.
