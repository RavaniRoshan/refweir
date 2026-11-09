# Pricing Hypothesis

**Status:** Internal hypothesis — **do not publish prices**
**Owner:** Founder/PM
**Gate:** G1 input; validated by Phase 2 interviews
**Created:** 2026-11-09
**Last updated:** 2026-11-09

This document records what we *think* the commercial model should be, what it is
anchored against, and how we will falsify it. It is not a price list. Nothing here
may appear on the site, the Product Hunt listing, the README, or in a sales
conversation as a commitment until the validation rules in §8 are met.

The product is OSS-first. The paid tiers exist to fund the OSS core, not to
cripple it.

---

## 1. What we are actually selling

Refweir's differentiated value is **containment with evidence**:

- every agent gets a short-lived, ref-scoped, revocable lease instead of a
  long-lived repo-wide token;
- the pre-receive gate enforces the lease at the Git boundary, not in the UI;
- every lease lifecycle event, accepted push, and rejection lands in a durable
  audit log the operator controls.

That value lands differently for three buyers:

| Buyer | What they need | What they will pay for |
|---|---|---|
| Individual / small team | A forge they can run on a VPS, with agent containment | Little or nothing; they self-host the OSS core |
| Platform/DevEx team at a mid-size company | Rollout to 10–100 engineers and a mixed agent fleet | Operational support, upgrade path, hosted evaluation |
| Regulated / air-gapped enterprise | Procurement, SSO/SCIM, audit export, HA/DR, indemnity | A contract, a support SLA, and a named owner |

The free tier is not a demo. It must be the real product for the first two rows,
or the OSS positioning is dishonest.

---

## 2. Packaging hypothesis (three packages)

> Working names only. Features listed as "existing" are in the repo today;
> "planned" means not built and must not be sold yet.

### Package A — Community (free, OSS)

**Target:** individuals, small teams, air-gapped labs, evaluators.

- Existing: single hermetic binary, embedded UI, Smart-HTTP + pure-Go SSH, LFS,
  capability leases, ref-prefix isolation, pre-receive secret scanning, protected
  branches, MCP + read-only MCP, archived-repo freeze, deploy-key isolation,
  durable JSONL audit log, Docker/Compose/Helm/Terraform.
- Existing: community support via GitHub issues/discussions.
- No: SSO/SAML, SCIM, audit shipping/rotation, support SLA, indemnity, HA/DR
  runbook, per-repo configurable rulesets, hosted operation.
- Limits: one trust domain per daemon; sized for small-to-medium teams (see
  `docs/benchmarks.md`).

**Price:** $0. No seat cap, no repo cap, no agent cap, no feature flag that turns
off leases or audit. This is the distribution and trust engine.

### Package B — Hosted evaluation / Hosted team

**Target:** teams that want to try the workflow before self-hosting, and teams
that do not want to operate a server.

- Existing (demo only): a seeded, reset-on-schedule public demo with a one-click
  agent run; no signup wall for the first run.
- Planned: isolated per-tenant instances, managed upgrades, managed backups,
  audit export, SSO, usage dashboard.
- Planned: a time-boxed evaluation instance (14–30 days) with a real repo and a
  real agent, not a canned video.

**Price hypothesis:** free public demo; evaluation instance free during beta;
paid hosted team tier **only after** the isolation and reset story is real. Do not
quote a hosted price until §8 passes.

### Package C — Enterprise self-host

**Target:** regulated, air-gapped, or procurement-bound organizations running
multiple agents against private code.

- Existing: SSO (OIDC/SAML), SCIM 2.0, RBAC, audit log, Helm chart, Terraform
  provider, air-gap deployment, signed releases (planned), security policy.
- Planned and required before selling: audit export/shipping and rotation, HA/DR
  runbook, upgrade/rollback guide, support SLA, security questionnaire pack,
  named support contact, license/key mechanism (if any), indemnity terms.
- Optional services: onboarding, air-gap install assistance, policy design
  workshop, quarterly containment review.

**Price hypothesis:** annual per-instance license with an included engineer band,
plus support. Not per-agent. Not a percentage of cloud spend.

---

## 3. Why not per-agent pricing

Agents are ephemeral, bursty, and often unnamed. Per-agent pricing would:

- punish the exact behavior we want (many short-lived, well-scoped agents);
- require a metering and billing surface we do not have and cannot yet make
  trustworthy;
- create an incentive to reuse one over-broad lease instead of issuing tight ones;
- be trivially gamed by rotating `agentId` values.

Per-seat pricing is understandable, but it mis-measures value: a team of five
engineers can run fifty agents. Per-instance pricing matches the deployment
(one daemon, one trust domain) and is the simplest thing to buy. The hypothesis is
**per-instance enterprise license + included engineer band**, with a hosted
per-seat option only if interviews show seat-based budgeting is mandatory.

---

## 4. Anchors (November 2026)

These are competitors' list prices, used only to bound our hypothesis. They are
not our prices, and several cover a different scope (agent execution, not forge
containment).

| Product | Entry price | Per seat / year | What it includes | Relevance |
|---|---|---|---|---|
| **GitLab air-gapped Duo Agent Platform** | **from $299/user/mo** | **~$3,588** | Air-gapped agent platform on top of GitLab | The explicit price umbrella for sovereign agents |
| **GitHub Enterprise** | **$21/user/mo** | **~$252** | Forge; agent features are cloud-first | Baseline self-hosted forge |
| **GitHub Copilot Business / Enterprise** | **$19 / $39 per seat/mo** | **~$228 / ~$468** | Agent seats, metered usage separate | Agent budget already exists |
| **GitLab Premium** | **$29/user/mo** annual | **~$348** | Forge + some Duo credits | Mid-market forge |
| **Gitea Enterprise** | **$9.50 promo / $19 list per user/mo** | **~$114 / ~$228** | Self-hosted forge, no native agent identity | The "cheap self-hosted" floor |
| **RhodeCode Enterprise** | **$75/user/yr** | **~$75** | Self-hosted forge, agent features mostly roadmap | Low-end enterprise forge |
| **Helix self-hosted agent pilot** | **from $75K** | — | Eight-week pilot, branch-scoped agent credentials | Evidence the pain is budgeted |

**Reading the anchors:**

- Combined GitHub Enterprise + Copilot Business is roughly **$480/seat/year**.
  That is the realistic "do nothing different" budget for a mid-size team.
- GitLab's air-gapped agent SKU is roughly **$3,588/seat/year** — an order of
  magnitude higher, because it bundles agent execution and a sovereign deployment.
- Gitea Enterprise is roughly **$228/seat/year** and has no native per-agent
  lease model. We must be meaningfully better than "free Gitea + a bot PAT" to
  charge anything.
- Helix's $75K pilot says a regulated buyer will pay a real number for
  branch-scoped agent credentials, but that is a services-led motion, not a
  self-serve price.

---

## 5. Internal price bands (hypothesis only)

Use these to steer interviews. Never state them to a prospect as an offer.

| Package | Hypothesis | Rationale | Confidence |
|---|---|---|---|
| Community | **$0** | Distribution, trust, air-gap labs | High |
| Hosted evaluation | **$0 during beta** | Validate activation before charging | High |
| Hosted team (later) | **$20–$30 per engineer/mo** | Below GitHub Enterprise + Copilot (~$40/mo combined) | Low |
| Enterprise self-host | **$15K–$40K/year** for a 25–100 engineer band (~$400–$800/seat/yr at 50 seats), plus support | Above Gitea floor, far below GitLab's $3,588/seat air-gap SKU | Low |
| Enterprise support add-on | **15–25% of license/year** | Standard infrastructure-software ratio | Low |
| Air-gap onboarding / policy workshop | **$5K–$15K one-time** | Services, not product | Low |

Rules for the bands:

- The enterprise band must be **defensible against the workaround**: GitLab
  extension + proxies + policy glue + audit stitching is engineer time, and
  engineer time is the real competitor.
- Do not discount below the Gitea floor plus support cost; a $5K contract with a
  security questionnaire and an SLA is a loss.
- Do not anchor the first offer at $299/user/mo just because GitLab does. That
  number buys agent execution; we sell the containment layer.
- If the buyer's alternative is "GitHub Enterprise + Copilot + branch protection
  + a bot PAT," our reference price is the delta in risk and engineer time, not
  the full GitLab SKU.

---

## 6. Segment and qualification

**Strong fit:**

- 25–500 engineers, multiple AI coding agents (Claude Code, Codex, Cline, Cursor,
  internal agents), private or regulated code.
- A platform/DevEx or security owner who already owns "how do agents touch Git?"
- Self-hosted or air-gap requirement, or a data-residency constraint.
- Existing spend on agent seats or a forge; a budget line already exists.

**Weak fit (do not force):**

- Solo developer or a team that has not run an agent against a shared repo.
- A team fully committed to GitHub Cloud and comfortable with vendor-side audit.
- A team that only wants a cheaper GitHub clone and does not care about agent
  containment.
- A buyer who needs CI marketplace parity, native mobile apps, or 10k-seat scale.

**Disqualifiers for the enterprise tier:** no security owner, no procurement path,
no self-host requirement, or a hard requirement for features we have not built
(per-repo rulesets, audit shipping, HA/DR).

---

## 7. Interview script (10–15 interviews)

**Format:** 30–40 minutes, remote, recorded with permission, one interviewer.
**Target roles:** platform engineering lead, DevEx lead, security architect,
engineering manager running agents, CISO delegate. Mix company sizes (50–5,000).
**Goal:** budget owner, current spend, procurement path, and willingness to pay.
**Rule:** ask about the past and the present, not the hypothetical future. Ask for
a number before showing ours.

### 0. Frame (2 min)

> "I'm researching how teams run AI coding agents against private code, and what
> they do about branch access and audit. I'm not selling anything today. I'll ask
> about what you do now, what it costs, and how a purchase like this would happen.
> There are no right answers."

### 1. Context and role (4 min)

1. What's your role, and who owns the developer platform / Git infrastructure?
2. How many engineers, and how many AI agents run against your repos in a typical
   week? Which agents/tools?
3. Are your repos cloud-hosted, self-hosted, or air-gapped? Which forge?

### 2. Current agent workflow (6 min)

4. Walk me through what happens today when an agent produces a change. Who or
   what reviews it, and how does it reach `main`?
5. What credential does the agent use — a bot PAT, a GitHub App, a deploy key, an
   SSH key, something else? Who owns it, and how long does it live?
6. Has an agent ever pushed somewhere it shouldn't, force-pushed, or leaked a
   secret? What happened?
7. What stops an agent from touching `main` today — branch protection, a proxy, a
   harness rule, code review, or nothing?

### 3. Current spend and workaround (6 min)

8. What do you pay today for the forge and for agent seats? (Push for list vs
   negotiated, annual vs monthly.)
9. How much engineer time goes into the workaround — maintaining proxies,
   rotating bot tokens, stitching audit logs, reviewing agent PRs?
10. Have you bought or trialed anything specifically for agent governance
    (GitLab Duo, Copilot, Helix, a security vendor)? What did it cost, and why did
    it or didn't it stick?
11. If you had to justify this internally, which budget line would it come from —
    platform, security, or AI tooling?

### 4. Pain and risk (5 min)

12. What would a regulator or auditor ask you about agent commits that you cannot
    answer today?
13. What's the cost of the worst realistic incident — a bad push to `main`, a
    leaked key, an unattributable change?
14. If you could change one thing about how agents touch Git, what would it be?

### 5. Procurement (4 min)

15. Who signs off on a $10K–$50K tool? Who signs off above that?
16. What does the path look like — security review, legal, vendor questionnaire,
    SOC 2 requirement, procurement portal? How long?
17. Do you require self-host or air-gap, or is a managed instance acceptable?
18. What would make you reject a vendor in the first 15 minutes?

### 6. Willingness to pay (6 min)

Ask for a range before revealing any of ours.

19. For a self-hosted forge that gives every agent a scoped, expiring lease and a
    durable audit trail, what would you expect to pay per year for a team your
    size?
20. Is that per seat, per instance, per agent, or something else? Why?
21. At what price would it be an easy yes without procurement? At what price does
    it need a business case? At what price is it clearly too expensive?
22. If the OSS core were free and the paid tier added SSO, audit shipping, HA/DR,
    and support, would you still buy? Which of those is the must-have?
23. Would you pay for a paid pilot or onboarding, or does it have to be free to
    start?
24. Who else should I talk to?

### 7. Close (2 min)

25. Anything I should have asked?
26. Can I follow up when we have something to show?

---

## 8. Validation rules and decision gates

**Signal log:** record each interview in `docs/launch/PRICING_INTERVIEWS.md`
(create when the first interview happens) with role, company band, current spend,
budget owner, procurement path, WTP range, must-have feature, and verbatim quotes.
Do not paraphrase away disagreement.

**Three-repeat rule:** do not publish a price until the same willingness-to-pay
band and the same must-have feature appear **three separate times** from
independent buyers.

**Proceed to publish when all of these are true:**

- [ ] At least 10 of 15 interviews are in the target segment.
- [ ] At least 3 buyers name a budget owner and a procurement path without
      prompting.
- [ ] At least 3 buyers state a per-instance or per-seat range inside the
      hypothesis band.
- [ ] At least 3 buyers name the same must-have paid feature (most likely SSO,
      audit export, or support).
- [ ] At least 1 buyer agrees to a paid pilot or a design-partner agreement.
- [ ] The required enterprise features in Package C are built or have dated
      commitments.
- [ ] Legal review of the OSS license, enterprise terms, and support SLA is done.

**Change the model when:**

- Buyers consistently reject per-instance and insist on per-seat → test hosted
  per-seat and an enterprise per-seat license.
- Buyers say the OSS core is enough → the paid tier is support/services, not
  product features.
- Buyers compare us only to Gitea's price → the containment value is not landing;
  fix the demo and positioning before pricing.
- Buyers demand per-agent metering → do not build it; revisit when there is a
  trustworthy run identity and a billing surface.

**Kill criteria for the enterprise tier at launch:** if fewer than 3 interviews
produce a budget owner and a real range, launch with OSS + hosted demo only and
keep the enterprise tier as a waitlist. Do not invent a price to look established.

---

## 9. What we will not do

- Publish a price before §8 passes.
- Paywall leases, ref isolation, the pre-receive gate, or the audit log. Those are
  the product's integrity; gating them makes the OSS tier a demo.
- Charge per agent at launch.
- Sell HA/DR, audit shipping, SCIM, or a support SLA before they are built and
  tested.
- Claim SOC 2, ISO 27001, or "enterprise-ready." Say what is done and what is not.
- Use a fake "contact us" price to imply demand we have not measured.
- Discount to win a logo in a way that sets a reference price we cannot support.

---

## 10. Open questions

1. Does the hosted tier need to be multi-tenant, or is one instance per customer
   acceptable at launch? (Affects isolation cost and the hosted price.)
2. Is audit **export** (SIEM/syslog) the must-have paid feature, or is SSO?
3. Will buyers accept a per-instance license, or does procurement require
   per-seat because that is how forge budgets are structured?
4. What is the real cost of the workaround? No interview has quantified it yet.
5. Does air-gap installation justify a services fee, or is it table stakes?
6. Is there a design-partner price (discounted, reference-able) that does not
   poison the list price?
