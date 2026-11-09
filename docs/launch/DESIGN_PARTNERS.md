# Design partners

**Status:** Phase 2.1/2.2 working document (draft, in progress)
**Owner:** Founder/PM + DevRel
**Plan:** [`plan.md`](../../plan.md) §2.1–2.2
**Created:** 2026-11-09

This is the living CRM for design-partner outreach. Names are organizations and
roles, not personal email addresses. Verify a real contact before sending; do
not invent people. Update this file as interviews happen; keep the signal log in
[`PRICING.md`](PRICING.md) §8.

---

## 1. ICP and segment priority

1. **Air-gapped / defense / federal / national labs** — highest willingness to
   pay, strongest data-residency constraint, longest cycle, highest proof
   requirement.
2. **EU / sovereign platform teams** — data-residency and audit requirements,
   mid-cycle, strong fit for "self-hosted, auditable, vendor-neutral."
3. **Agent-heavy platform engineering teams** — fastest activation, best demo
   conversion, good for case studies.
4. **Cost-sensitive mid-market** — replaces agent seats/credits, price-sensitive,
   good for volume and OSS adoption.

Target mix for the first 50: 13 / 13 / 13 / 11 (approx). Sign 5–8 by end of
Week 4, 3 quotes and 1 mini-case study by Week 10.

## 2. Qualification rules

**Strong fit:**
- 25–500 engineers; multiple AI coding agents on shared repos.
- A platform/DevEx or security owner who already owns "how do agents touch Git."
- Self-host or air-gap requirement, or a data-residency constraint.
- Existing spend on agent seats or a forge; a budget line exists.

**Weak fit (do not force):**
- Solo developer or a team that has not run an agent against a shared repo.
- Fully committed to GitHub Cloud and comfortable with vendor-side audit.
- Wants only a cheaper GitHub clone with no agent-containment need.
- Needs CI marketplace parity, native mobile, or 10k-seat scale.

**Disqualifiers for the enterprise tier:** no security owner, no procurement path,
no self-host requirement, or a hard requirement for features not yet built
(per-repo rulesets, audit shipping, HA/DR).

## 3. Target list (50)

Status key: `new` = not contacted; `sent` = outreach sent; `reply` = replied;
`call` = discovery call done; `partner` = signed design partner.

### Segment 1 — Air-gapped / defense / federal / national labs (13)

| # | Organization | Contact (role) | Segment | Trigger | Next action | Status |
|---|---|---|---|---|---|---|
| 1 | US Air Force Software Factory (Platform One) | Platform engineering lead | 1 | Air-gapped DoD software delivery; agents under evaluation | Email: ask how agent commits are attributed today | new |
| 2 | US Navy / NAVWAR | DevSecOps platform owner | 1 | Air-gapped Git and CI; agent governance question | Email with containment demo link | new |
| 3 | Defense Innovation Unit (DIU) | Engineering lead | 1 | Rapid software acquisition; agent tooling pilots | Email: offer 30-min containment walkthrough | new |
| 4 | Johns Hopkins APL | Secure development platform lead | 1 | Air-gapped research code; audit requirements | Email: threat model + demo | new |
| 5 | MITRE | Software engineering center lead | 1 | Internal agent tooling; containment gap | Email: lease model discussion | new |
| 6 | Sandia National Laboratories | DevOps/platform lead | 1 | Classified-network Git; agent policy under review | Email: air-gap install offer | new |
| 7 | Los Alamos National Laboratory | Platform engineering lead | 1 | Regulated code; data-residency constraint | Email: demo + install offer | new |
| 8 | Oak Ridge National Laboratory | Research software engineering lead | 1 | Multi-agent workflows on HPC-adjacent code | Email: containment demo | new |
| 9 | National Security Agency / CNSA-adjacent | Platform security lead | 1 | Air-gapped forge replacement evaluation | Email: security questionnaire pack | new |
| 10 | NATO CCDCOE | Cyber defence platform lead | 1 | Allied software sharing; audit requirements | Email: positioning one-pager | new |
| 11 | UK Ministry of Defence (DIO) | Platform engineering lead | 1 | Sovereign code hosting; agent governance | Email: air-gap install offer | new |
| 12 | Bundeswehr Cyber and Information Domain | Platform lead | 1 | EU sovereign deployment; agent containment | Email: EU data residency angle | new |
| 13 | Lockheed Martin (internal platform) | DevEx lead | 1 | Air-gapped enterprise Git; agent pilot | Email: containment + audit export | new |

### Segment 2 — EU / sovereign platform teams (13)

| # | Organization | Contact (role) | Segment | Trigger | Next action | Status |
|---|---|---|---|---|---|---|
| 14 | SAP (internal platform) | DevEx / platform engineering lead | 2 | Sovereign cloud + AI coding agents | Email: self-hosted forge angle | new |
| 15 | Siemens | Platform engineering lead | 2 | On-prem Git; agent governance gap | Email: containment demo | new |
| 16 | Bosch | DevOps platform lead | 2 | Regulated automotive code; agent pilots | Email: audit trail + leases | new |
| 17 | Ericsson | Platform engineering lead | 2 | Telecom code; data residency | Email: self-host + air-gap | new |
| 18 | Nokia | DevEx lead | 2 | Sovereign deployment requirement | Email: positioning one-pager | new |
| 19 | Deutsche Telekom (internal) | Platform engineering lead | 2 | Sovereign cloud; agent governance | Email: demo link | new |
| 20 | Capgemini (internal platform) | Platform engineering lead | 2 | EU client data residency | Email: air-gap install offer | new |
| 21 | Atos / Eviden | Platform lead | 2 | Sovereign cloud offering | Email: partnership angle | new |
| 22 | Thales | Secure development lead | 2 | Defense-adjacent regulated code | Email: threat model + demo | new |
| 23 | Airbus (internal) | Platform engineering lead | 2 | Air-gapped engineering code | Email: containment demo | new |
| 24 | European Space Agency (ESOC) | Software infrastructure lead | 2 | Sovereign research code | Email: air-gap install | new |
| 25 | CERN (IT) | Version control platform lead | 2 | Open science + sovereign compute | Email: containment demo | new |
| 26 | Fraunhofer Society | Research software lead | 2 | Multi-site research code | Email: self-host offer | new |

### Segment 3 — Agent-heavy platform engineering teams (13)

| # | Organization | Contact (role) | Segment | Trigger | Next action | Status |
|---|---|---|---|---|---|---|
| 27 | Stripe (internal platform) | DevEx lead | 3 | High agent usage; containment question | Email: demo + lease model | new |
| 28 | Shopify (internal) | Platform engineering lead | 3 | Agent-authored PRs at scale | Email: containment demo | new |
| 29 | Cloudflare (internal) | Platform engineering lead | 3 | Agent workflows; self-host preference | Email: demo link | new |
| 30 | Vercel (internal) | Platform engineering lead | 3 | Agent-heavy engineering org | Email: containment demo | new |
| 31 | Supabase (internal) | Platform engineering lead | 3 | Agent-authored changes; audit need | Email: demo + case study ask | new |
| 32 | GitLab (internal platform) | DevEx lead | 3 | Competitor-adjacent; agent governance | Email: positioning one-pager | new |
| 33 | Sourcegraph | Platform engineering lead | 3 | Code intelligence + agents | Email: containment demo | new |
| 34 | Cursor (internal) | Platform engineering lead | 3 | Agent fleet; self-host gap | Email: vendor-neutral angle | new |
| 35 | Replit (internal) | Platform engineering lead | 3 | Agent-authored code at scale | Email: containment demo | new |
| 36 | Cognition (Devin) | Platform engineering lead | 3 | Agent fleet needs forge containment | Email: partnership + demo | new |
| 37 | Anthropic (internal) | Platform engineering lead | 3 | Agent tooling; self-hosted preference | Email: demo + threat model | new |
| 38 | OpenAI (internal) | Platform engineering lead | 3 | Agent workflows; audit requirements | Email: demo link | new |
| 39 | Anysphere (Cursor) | Platform engineering lead | 3 | Agent fleet; vendor-neutral forge | Email: demo + containment | new |

### Segment 4 — Cost-sensitive mid-market (11)

| # | Organization | Contact (role) | Segment | Trigger | Next action | Status |
|---|---|---|---|---|---|---|
| 40 | PostHog (internal) | Platform engineering lead | 4 | Cost-sensitive; OSS-friendly | Email: OSS + containment | new |
| 41 | Cal.com | Platform engineering lead | 4 | OSS company; agent governance | Email: demo + OSS angle | new |
| 42 | Plausible | Platform engineering lead | 4 | Small team; agent usage | Email: self-host offer | new |
| 43 | Ghost | Platform engineering lead | 4 | OSS; cost-sensitive infra | Email: demo link | new |
| 44 | n8n | Platform engineering lead | 4 | Automation-heavy; agent usage | Email: containment demo | new |
| 45 | Appwrite | Platform engineering lead | 4 | OSS; agent-authored PRs | Email: demo + OSS angle | new |
| 46 | Supabase (smaller team) | Platform engineering lead | 4 | Cost of agent seats | Email: containment demo | new |
| 47 | Railway | Platform engineering lead | 4 | Infra platform; agent usage | Email: demo link | new |
| 48 | Render | Platform engineering lead | 4 | Infra platform; agent usage | Email: demo link | new |
| 49 | Fly.io | Platform engineering lead | 4 | Infra platform; agent usage | Email: demo link | new |
| 50 | HashiCorp (internal, pre-IBM) | Platform engineering lead | 4 | Infra tooling; agent governance | Email: demo + containment | new |

## 4. Outreach cadence (Phase 2.2)

- **Week 2:** Send 10 personalized emails (5 segment 1, 3 segment 2, 2 segment 3).
- **Weeks 3–4:** 10 per week, rotating segments. Track replies in this file.
- **Week 4:** Target 5–8 signed partners.
- **Weeks 5–8:** White-glove installs, private channels, roadmap calls.
- **Week 10:** 3 quotes/logos and one mini-case study with a number.

## 5. Email template (personalized, 4–6 sentences)

Subject: `How does <company> stop an agent from pushing to main?`

Body:

> Hi <name>,
>
> I'm researching how teams run AI coding agents against private code, and what
> they do about branch access and audit. I'm not selling anything today.
>
> Refweir is a self-hosted Git forge that gives every agent a short-lived,
> ref-scoped lease instead of a long-lived token, and records every Git write in
> an append-only audit log. Here is a 82-second containment demo:
> <link to containment-demo.mp4>.
>
> Two questions: (1) what credential does your agent use today, and (2) who
> would sign off on a $10K–$50K tool like this?
>
> Happy to do a 30-minute walkthrough or a white-glove install on your own
> hardware. No cost during the design-partner phase.
>
> — <founder>

Rules:
- Ask about the past and present, not the hypothetical future.
- Ask for a number before showing ours (see [`PRICING.md`](PRICING.md) §7).
- Do not attach a price list. Do not claim SOC 2, ISO 27001, or
  "enterprise-ready."
- Send from a personal address, not a marketing domain.

## 6. Tracking fields (update per contact)

For each reply, record in this file or in `docs/launch/PRICING_INTERVIEWS.md`:
role, company band, current spend, budget owner, procurement path, WTP range,
must-have feature, and verbatim quotes. Do not paraphrase away disagreement.

## 7. Exit criteria for Phase 2.2

- [ ] 50 named targets in §3 (done).
- [ ] 10 personalized emails/week sent for 4 consecutive weeks.
- [ ] 5 signed design partners.
- [ ] 3 quotes/logos.
- [ ] 1 mini-case study with a number.
- [ ] At least 10 of 15 pricing interviews in the target segment.
- [ ] Three-repeat rule met before publishing any price
  (see [`PRICING.md`](PRICING.md) §8).

## 8. Notes

- "Named target" means an organization plus a role, not a verified individual.
  Verify the contact before sending; do not fabricate personal emails.
- If a target is clearly a weak fit on first contact, mark it `dropped` with a
  one-line reason rather than forcing the conversation.
- Air-gapped/defense targets need the security questionnaire pack and the threat
  model before a call; send those, not a price.
