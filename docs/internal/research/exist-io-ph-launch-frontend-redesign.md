# exist.io — Product Hunt Launch & Agents-First Frontend Rebuild

**Deep research report** · 8 November 2026
**Scope:** OSS core + hosted tier · radical agents-first frontend redesign · 3+ month runway
**Method:** 8 parallel research streams → 1 adversarial verifier → 2 resolution passes
**Confidence:** Medium-high on launch mechanics and naming; medium on competitive pricing and IA; medium-low on willingness-to-pay and PH conversion for self-hosted OSS.

---

## Executive summary

The launch is winnable, but not under the current name or the current positioning. `exist.io` is the domain of a 13-year-old Melbourne life-tracking product, and the forge has 0 GitHub stars, 0 forks, and no owned brand. Renaming is the first launch task, not a branding nicety.

Product Hunt's math has changed. In late 2026, the observed overall Top 10 cutoff is roughly **90–160 points** (plan for **150–250** to be safe), a dev-tool #1 is roughly **300–450**, and the old "1,000 upvotes = #1" rule is obsolete. PH referrals for a top-3 launch are about **1,000–2,500 visitors**, converting at **1–3%** for B2B dev tools; Hacker News and niche Reddit repeatedly outperform PH for OSS installs. PH is a coordination event, not a dependable signup engine.

"First Git forge for agents" is false as of November 2026. Cursor Origin, Entire, GitHub's Agents tab, and Bitbucket's agentic pipelines all exist. The defensible gap is narrower and more valuable: a **self-hosted, air-gapped, auditable, vendor-neutral forge for agent fleets**, with per-agent identity, scoped leases, and policy at the forge layer. GitLab's **$299/user/month** air-gapped agent SKU is the price umbrella to undercut.

A radical agents-first redesign is justified, but the evidence favors a **hybrid IA**: a global Inbox/Runs/Reviews/Policies rail for cross-repo triage, a repo-scoped Agents tab beside Code/Issues/PRs, and a run detail where **the diff is truth and chat is annotation**. Cursor 3 shows the cost of hiding the diff: users report changes are harder to see and whole files render as new.

Visually and technically, do not restart from zero. The existing Patina-on-Primer system is already close to the 2026 mood; keep it, add a signature mono/display pairing and one tactile signature (a lease seal or merge gate), and photograph a dark UI on a warm neutral gallery surround. On the stack, keep React Aria and Vite 8, add TanStack Router + Query, go SSE-first, use Pierre for large read-only diffs and CodeMirror 6 for editable conflicts, and migrate route-by-route behind the existing shell.

---

## Scope and method

This report answers two linked questions:

1. **How does exist.io reach Top 10 on Product Hunt launch day?**
2. **How should the frontend be recreated around an agents-first information architecture without losing the Git mental model?**

Decisions locked before research:

| Decision | Choice |
|---|---|
| Launch subject | OSS core **and** a hosted tier |
| Frontend direction | **Radical redesign, new IA** (agents-first) |
| Timeline | **3+ months**, no fixed date |
| Research depth | **Deep** — 8 independent streams, adversarial verification, resolution pass |

Eight sub-questions were researched in parallel: PH ranking mechanics; dev-tool launch playbooks; competitive/pricing landscape; practitioner pain and objections; naming/brand/PH assets; agents-first IA; visual design direction; frontend architecture. A separate verifier then checked every finding against its cited source, looking for unsupported inference, contradictions, stale evidence, and vendor marketing. Two resolution agents re-checked the decision-critical errors. The report carries the corrected numbers and flags the remaining uncertainty rather than silently dropping it.

**What this report is not:** legal advice on the name, a guarantee of PH placement, or a substitute for 10–15 customer interviews on willingness to pay. Those gaps are called out explicitly.

---

## Background: the product, the market, and the terms

**exist.io** is an Apache-2.0, self-hosted Git forge written in pure Go. It ships as a single binary with an embedded web UI, Smart-HTTP Git on `:8080`, an embedded pure-Go SSH server on `:2222`, Git LFS v1 CAS, OpenSSH/GPG signature verification, OIDC/SAML/SCIM, a native MCP server at `/mcp` (plus a read-only sandbox), ephemeral HMAC-SHA256 agent leases with `refs/heads/agent/*` isolation, an autonomous PR review agent, semantic AST symbol search, and Raft/object-shard replication work. The repository is three weeks old (created 14 October 2026), at v0.6.0-beta, with **0 stars, 0 forks, and 0 open issues**.

The market context that matters:

- **GitHub** is the default and is bundling agents into the forge (Agent HQ, Copilot coding agent, repository Agents tab, mission control, `actor_is_agent` audit fields). Its self-hosted product, GHES, is structurally split from the cloud agent features.
- **GitLab** has a mature Duo Agent Platform, but its air-gapped self-managed agent SKU starts at **$299/user/month**, creating an explicit price umbrella.
- **Cursor** now owns a Git forge (Origin, early beta) and was acquired by SpaceX in a **$60B** all-stock deal that closed 14 August 2026.
- **Entire**, founded by former GitHub CEO Thomas Dohmke with a **$60M seed**, launched an agent-native forge on 26 October 2026 that replaces pull requests with "Trails".
- **Open-source forges** (Forgejo, Gitea, GitLab CE) are free or cheap and adding MCP, but are thin on per-agent identity, scoped leases, and audit.
- **MCP is table stakes.** Gitea shipped an official MCP server in September 2026; Forgejo has community servers; GitHub has a registry; GitLab's is in beta. MCP alone is not a moat.

**Terms used in this report:**

| Term | Meaning |
|---|---|
| **PH** | Product Hunt |
| **Points** | PH's engagement-weighted score; not the same as raw upvotes |
| **Featured** | PH's editorial gate; only featured launches compete for the daily leaderboard |
| **ICP** | Ideal customer profile |
| **Lease** | exist.io's short-lived, HMAC-signed capability token for an agent, scoped to ref prefixes |
| **Run / session** | One agent task from intent to merge, with logs, diff, checks, and audit |
| **Air-gapped** | Infrastructure with no internet access; cannot use cloud-hosted agents |
| **WTP** | Willingness to pay |

---

## Findings

### 1. Product Hunt mechanics: how Top 10 is actually decided

**PH ranks by points, not upvotes.** Its own 2025 documentation says every launch earns points from genuine engagement — upvotes plus comments and other signals — so displayed points and upvote counts can legitimately diverge. The current guide says the leaderboard moves on upvotes, comments, "time since launch", and other undisclosed signals, and it deliberately withholds the formula. A lower-upvote launch can outrank a higher-upvote one. ([PH points](https://help.producthunt.com/en/articles/10275873-what-are-points), [launch guide](https://help.producthunt.com/en/articles/17734397-6-what-to-do-during-and-after-your-launch))

**Vote quality dominates vote quantity.** PH's current FAQ calls "every upvote counts the same" a myth: established, active members' upvotes earn "far more points" than brand-new or low-quality accounts. A wave of fresh signups barely moves a launch. Older PH copy said all upvotes were equal; the newer page should win, but this is exactly the kind of rule that changes quietly, so verify again 30 days before launch. ([PH FAQ](https://help.producthunt.com/en/articles/17734646-8-launch-faq-and-myths))

**Featuring is the real gate.** Only featured launches compete for the daily leaderboard. Featuring is an editorial judgment on useful, novel, high-craft, creative products; waitlists, directories, templates, boilerplates, courses, services, and similar are explicitly excluded. An unfeatured launch is still live via direct link and the "See all of today's products" button, but it cannot win the day. For a self-hosted OSS forge at v0.6-beta, the risk is not the category — it is looking unfinished or like a boilerplate. The live demo and screenshots have to clear the "high craft" bar. ([Featuring guidelines](https://help.producthunt.com/en/articles/9883485-product-hunt-featuring-guidelines))

**Timing is a fixed Pacific-time day.** Go live at **12:01 AM PT** and run until midnight PT. Scheduling opens the next day and runs up to **30 days ahead**. There is no official golden weekday: PH says weekends get **15% more "Visit" clicks**, while larger companies cluster on weekdays and smaller teams on weekends. A founder postmortem argues Tuesday–Thursday bring the most traffic but the stiffest competition, and Friday–Saturday are quietest; PH sends no Top Products newsletter on Friday/Saturday, so a Sunday #1 gets Monday's email. Treat day choice as goal-dependent, not doctrinal. ([PH FAQ](https://help.producthunt.com/en/articles/17734646-8-launch-faq-and-myths), [PH preparing](https://www.producthunt.com/launch/preparing-for-launch), [Flexprice](https://flexprice.io/blog/how-we-ranked-product-of-the-day-on-product-hunt))

**Early votes are deliberately noisy, and the public evidence is stale.** A January 2024 founder analysis quotes PH's notice that upvotes are hidden for the first four hours; PH's "Randomized Leaderboard Day" post describes a 25-minute hidden / 5-minute visible loop. Neither is documented in the current guide, so 2026 behavior is unconfirmed. Do not build the launch plan around a specific first-hour visibility rule. ([Indie Hackers](https://www.indiehackers.com/post/what-i-learned-from-the-first-hour-of-launching-on-product-hunt-6bc788313c))

**There is no frozen final score.** PH's leaderboard states that "score is the current active score as of now, while rank reflects ordering by score at the end of the viewed period." Scores rise during the day and decay after it. Mid-day snapshots are not comparable to end-of-day medians, and archived numbers drift. This is why the "top 10 threshold" is a range, not a number. ([Nov 8 leaderboard](https://www.producthunt.com/leaderboard/daily/2026/11/8), [Nov 6 leaderboard](https://www.producthunt.com/leaderboard/daily/2026/11/6))

**The late-2026 math.** LaunchPact's dataset (self-interested, but the most granular public source) gives all-topic medians by rank band across 11,158 settled launches from September 2024 to 7 November 2026: **#1 = 539; #2–3 = 363; #4–5 = 247; #6–10 = 156**. Developer-tool months are lower and volatile: September 2026 median **95**, day-#1 **439**, top **622**; October 2026 median **86**, day-#1 **314**, top **436**, with **60.9%** of dev-tool launches under 100. The practical late-2026 range: **overall Top 10 ≈ 90–160; dev-tool #1 ≈ 300–450; overall #1 ≈ 350–550**. ([LaunchPact statistics](https://www.launchpact.io/product-hunt-statistics), [Sept 2026](https://www.launchpact.io/product-hunt-launches/september-2026/developer-tools), [Oct 2026](https://www.launchpact.io/product-hunt-launches/october-2026/developer-tools))

**The category is saturated.** October 2026 had roughly **322 featured Developer Tools launches, about 10.4 per day**; September had about 230, or 7.7 per day. A "top 10 dev tool of the day" is therefore close to "every dev tool that launched." The meaningful target is **overall daily rank**, not the dev-tool subcategory. ([LaunchPact October](https://www.launchpact.io/product-hunt-launches/october-2026/developer-tools))

**Self-hunting is allowed; paying a hunter is not.** PH says **79%** of featured posts were self-hunted and **60%** of #1 Product of the Day winners self-hunted. Paying a hunter breaks the rules and can get the product unfeatured and the maker permanently banned. ([PH before launch](https://www.producthunt.com/launch/before-launch))

**Two content signals correlate with winning.** About **70%** of Product of the Day/Week/Month products had a maker first comment, and about **53%** of Product of the Day products since 2021 include a video. PH presents these as correlations, not proven causes, but they are cheap to satisfy. ([PH preparing](https://www.producthunt.com/launch/preparing-for-launch))

**Ads do not buy rank.** Promoted Products are paid placements marked "PROMOTED", unavailable for brand-new launches, and separate from the leaderboard. Appwrite reports a 30% lift in GitHub visits and 3× mentions from PH ads, but that is awareness, not organic rank. ([PH promoted](https://help.producthunt.com/en/articles/1444961-how-do-i-get-my-product-promoted), [PH sponsor](https://www.producthunt.com/sponsor))

**Implication for exist.io:** set the goal as **overall Top 10**, which is plausible at **150–250 points** with a coordinated campaign, and a stretch goal of **Top 3 at 300+**. Do not promise #1. With 0 stars and no owned audience today, the 3-month runway is mostly an audience-building runway.

---

### 2. Launch playbooks: what actually worked for dev tools

**The benchmark set is survivorship-biased but instructive.** Dub.co (March 2024) hit **1,085 upvotes and 210 comments**, #1 day/week/month, with 25k email subscribers and 15k GitHub stars; it drove about **2,000 unique visitors and 663 signups** on launch day, 8× its daily average. Appwrite Sites (May 2025) hit **971 upvotes and 203 comments**, #1 day and #4 week/month, on its sixth PH launch with a page already at 5.4k followers. Daytona (March 2024) hit **#2 with 522 upvotes and 144 comments**, entering with 5k GitHub stars and one week of prep. Flexprice (April 2025) hit **#1 with 488 upvotes and 81 comments** and no existing community, but converted only **50+ signups**. ([Dub](https://dub.co/blog/product-hunt), [Appwrite](https://www.producthunt.com/products/appwrite/launches/appwrite-sites), [Daytona](https://www.daytona.io/dotfiles/how-we-missed-1-on-product-hunter-but-still-won-big), [Flexprice](https://flexprice.io/blog/how-we-ranked-product-of-the-day-on-product-hunt))

**The "1,000 upvotes" rule is dead.** Stackfix's December 2024 case study set the rule at **~1k upvotes for a high-confidence #1** and **~2k for Product of the Month**. By late 2026, the dev-tool day-#1 median is **314–499**, and the all-topic #1 median is **539**. The old number describes a different competition. ([Stackfix](https://tomdekan.com/articles/product-hunt-launch-guide), [LaunchPact](https://www.launchpact.io/product-hunt-statistics))

**Team and network breadth beat product quality on PH.** LaunchPact's 3,069-launch study (September 2025–July 2026) found solo makers reached a day's Top 10 **35.8%** of the time versus **74.2%** for teams of 5+; median votes were **107 vs 215**; launches with video won **59.0% vs 44.8%**; **43.6%** of 3,534 reciprocal support pacts failed, and only **3.0%** of fulfilled pacts left a comment. The strongest predictor was team/network breadth, not product quality. This is the single most uncomfortable finding for a solo/small-team launch. ([LaunchPact study](https://www.launchpact.io/research/product-hunt-launch-study))

**Pre-launch has changed.** PH retired "Coming Soon"/notify-me teaser pages on **28 August 2025**; new products get a permanent **Forum Thread** and a Pre-Launch Dashboard instead. There is no universal audience threshold: Dub had 25k emails, Flexprice had no community but a three-week warm-up, and Stackfix treats three weeks as the minimum. The common thread is an owned list that can be activated in the first four hours, not a static teaser page. ([PH changelog](https://www.producthunt.com/p/producthunt/changelog-august-28-no-more-coming-soon-and-a-couple-other-shiny-updates), [Infrasity](https://www.infrasity.com/blog/product-hunt-launch-for-developer-tools))

**Launch day is an operations problem.** Winners describe a 12:01 AM PT start, a first-four-hours war room, staggered outreach rather than one spike, replies to every comment, an in-app popup, a README banner, and a maker first comment that asks for feedback rather than upvotes. Kilo Code treats launches as repeatable: feature-led 60-character tagline, three screenshots, no marketing fluff, feedback-first comment, reply to every comment, then post-launch ads. ([Kilo Code](https://www.producthunt.com/p/kilocode/reverse-engineering-kilo-s-recent-product-hunt-launch), [Awesome PH teardown](https://github.com/fmerian/awesome-product-hunt/blob/main/teardowns/appwrite.md))

**Post-launch conversion is real but thin.** First-party data is lower than the folklore: Plausible's #2 launch drove **1,490 unique visitors and 22 trial signups on launch day** (601 upvotes); across the month, PH referred **2,399 visitors and 33 trials**, a **1.38%** conversion, versus **3.52%** from Google organic. Tally's #1 launch saw about **1,400 PH-referred visitors** on launch day (9.8k total site traffic). Dub is the outlier at **2,000+ visitors and 663 signups**, because it had a large owned list. Realistic PH-referred ranges: **#1–3 ≈ 1,000–2,500; #4–10 ≈ 300–1,000; outside Top 10 ≈ <300–500**. At 1–3% signup, that is **20–150 signups** for a typical B2B dev tool; 300–800+ requires a large pre-built list. ([Plausible](https://plausible.io/blog/product-hunt-launch), [Tally](https://blog.tally.so/product-hunt-launch-results/), [Dub](https://dub.co/blog/product-hunt))

**Hacker News and niche Reddit often beat PH for OSS installs.** Wasp/Open SaaS finished **#7 with ~400 upvotes** despite 100 upvotes in the first four hours; the 500k-subscriber PH newsletter produced only ~20 upvotes, while HN delivered **>3× PH's traffic** and the combined push produced **2.5k GitHub stars in a week**. Watermelon's HN front-page #2 (107 points) produced **50+ stars and 100+ installs**, while its PH #14 (193 votes) produced **243 visitors, ~30 installs, and 10 stars**. Postiz, a self-hosted product, grew from **3k to 14k GitHub stars in three months** by launching OSS and posting changelogs to r/selfhosted every two weeks — not by a single PH spike. ([Open SaaS](https://docs.opensaas.sh/blog/2025-05-07-you-should-still-launch-your-product-on-ph/), [Watermelon](https://medium.com/@baristaGeek/lessons-launching-a-developer-tool-on-hacker-news-vs-product-hunt-and-other-channels-27be8784338b), [Postiz](https://gitroom.com/blog/from-3k-stars-to-14k-in-3-months))

**PH is best understood as a 24-hour coordination event.** It provides awareness, a DR91 backlink, social proof, and an excuse to synchronize HN, Reddit, X, and GitHub trending. It is not a dependable signup engine. The teams that win treat it as one channel in a week-long campaign, not the campaign itself. ([HN discussion](https://news.ycombinator.com/item?id=41773355), [Hackmamba](https://hackmamba.io/developer-marketing/how-to-launch-on-product-hunt/))

**Implication for exist.io:** with 0 stars today, the launch runway must produce **1,000–2,000 engaged developers**, **5k+ GitHub stars**, **3–5 design-partner logos**, and a **working hosted demo** before launch day. Otherwise Top 10 is a coin flip, and the post-launch flatline is likely regardless of rank.

---

### 3. Competitive and pricing landscape

The category is no longer empty. By November 2026, every major forge has an agent story, two well-funded startups have shipped agent-native Git, and open-source forges have MCP. The honest map:

| Vendor | Entry price (Nov 2026) | Deploy | Native agents | MCP |
|---|---|---|---|---|
| **GitHub** | Team **$4**/user/mo; Enterprise **$21**/user/mo; Copilot Business **$19**, Enterprise **$39** per seat | Cloud + GHES (agent features cloud-first) | Agent HQ, Copilot coding agent, repository Agents tab | Registry + server |
| **GitLab** | Premium **$29**/user/mo annual, **12 credits**; Ultimate **custom**, **24 credits**; extra credit **$1**; air-gapped Duo Agent Platform **from $299**/user/mo | SaaS, self-managed, Dedicated | Duo Agent Platform (GA Jan 2026) | Client + server (beta) |
| **Bitbucket** | Free ≤5 users; Standard **$3.65**; Premium **$8** | Cloud + Data Center (quote) | Agentic Pipelines (beta) | Atlassian MCP |
| **Azure DevOps** | Basic **$6**/user/mo after 5 free; Test Plans **$52** | Cloud + Server | Via GitHub Copilot | Partial |
| **Cursor** | Plus **$8**; Premium **$30**; Super **$100**; Ultra **$200**; Teams **$30/$100/$200** per seat | SaaS + Origin (beta) | Cloud agents, Bugbot | Yes |
| **Codex / Claude** | **$20–$200** consumer/prosumer; Codex Business **$25/$20** standard, **$125/$100** premium per seat | SaaS | Cloud coding agents | Yes |
| **Devin** | **$20** Pro; **$200** Max; Teams **$80**/mo + **$40**/seat | SaaS + VPC | Devin agents | Yes |
| **Kiro** | **$20/$40/$100/$200**; Teams **$20**/seat + 500 credits; **$0.04** overage | SaaS + GovCloud | Agent IDE | Yes |
| **Factory** | **$20/$100/$200**; Teams **$60** + **$40**/seat | SaaS + on-prem Enterprise | Agent IDE | Yes |
| **Augment** | **$20/$100** flat up to 50 seats; usage at list **+40%** | SaaS | Agent IDE | Yes |
| **CodeRabbit** | **$30/$60/$90** monthly (**$24/$48/$72** annual); self-host only Enterprise | SaaS + Enterprise self-host | PR review | Partial |
| **Graphite** | **$20/$40** annual; GHES support | SaaS + GHES | Review | Partial |
| **Qodo / Sourcery** | Qodo **$0.015**/credit, on-prem Enterprise; Sourcery **$15/$30** (**$12/$24** annual), self-host Enterprise | SaaS + Enterprise | Review | Partial |
| **Gitea** | Free (MIT); Enterprise **$9.50 promo / $19 list** per user/mo | Self-host | None native | Official MCP (57 tools at v1.8.0; 48 after consolidation) |
| **Forgejo** | Free | Self-host | None native | Community MCP |
| **GitLab CE** | Free | Self-host | Duo requires paid tiers | Beta |
| **SourceHut** | **€4/8/12** per month | Hosted/self-host | Explicitly "No AI features whatsoever" | No |
| **RhodeCode** | CE AGPL; Enterprise **$75**/user/yr; Cloud from **$80**/mo | Self-host, air-gap | Agent features mostly **roadmap "In build · 5.15"**, not shipped | Token/API |
| **Entire** | Hosting free; agents PAYG at model provider rates | Hosted only; self-hosted Git nodes "upcoming" | Trails, review agents | Unclear |
| **Cursor Origin** | Included from Premium; no separate price | SaaS, early beta | Cloud agents on Origin repos | Yes |

Sources: [GitHub pricing](https://github.com/pricing), [GitHub Copilot billing](https://docs.github.com/en/copilot/concepts/billing/organizations-and-enterprises), [GitLab pricing](https://about.gitlab.com/pricing/), [GitLab credits for self-managed](https://support.gitlab.com/hc/en-us/articles/27614164750364-How-GitLab-Credits-Work-for-Self-Managed), [Bitbucket pricing](https://www.atlassian.com/software/bitbucket/pricing), [Azure DevOps pricing](https://azure.microsoft.com/en-us/pricing/details/devops/azure-devops-services/), [Cursor pricing](https://cursor.com/pricing), [Cursor Origin](https://cursor.com/docs/origin), [Devin pricing](https://devin.ai/pricing), [Kiro pricing](https://kiro.dev/pricing/), [Factory pricing](https://factory.com/pricing), [Augment pricing](https://www.augmentcode.com/pricing), [CodeRabbit pricing](https://www.coderabbit.ai/pricing), [Graphite pricing](https://graphite.dev/pricing), [Gitea](https://gitea.com/), [SourceHut pricing](https://sourcehut.org/pricing/), [RhodeCode roadmap](https://rhodecode.com/roadmap/), [Entire pricing](https://entire.com/pricing).

**The incumbents are structurally weak in four places.**

1. **Cloud-only agent execution for regulated code.** GHES runs repositories on-prem, but "most Copilot features require a presence on GitHub Enterprise Cloud"; the coding agent assumes GitHub.com repos; air-gapped Copilot CLI on GHES 3.22 is only a technical preview. GitLab's offline agent SKU is **$299/user/month**. That is a price and capability umbrella for a self-hosted, air-gapped forge. ([GitHub on GHES](https://docs.github.com/en/enterprise-server@3.18/copilot/copilot-on-ghes/about-copilot-on-ghes), [GitLab self-managed credits](https://support.gitlab.com/hc/en-us/articles/27614164750364-How-GitLab-Credits-Work-for-Self-Managed))

2. **Agent work priced like human seats or opaque credits.** GitHub's new model is a **$0 Copilot seat** for GitHub Enterprise Cloud licensed users, but **metered usage is billed separately** with **no included AI credits**; Copilot Business is **$19** with 1,900 pooled credits, Enterprise **$39** with 3,900. GitLab meters Duo work in **$1 credits**. For teams running many agents, fixed self-hosted compute is a budget argument, not just a sovereignty argument. ([GitHub Copilot access](https://docs.github.com/en/copilot/get-started/whats-changing-with-github-enterprise-and-copilot), [GitLab credits](https://about.gitlab.com/blog/introducing-gitlab-credits))

3. **Repo-scoped PATs, not per-agent identity.** GitHub's own MCP server now has read-only mode, toolsets, per-tool include/exclude, and lockdown mode, but **per-tool rate limits/quotas are still missing**; the configuration is static, not a runtime policy engine. A Claude Code, Cursor, or Codex process pushing through a PAT still looks like the token's owner. GitHub's `actor_is_agent` and `agent_session_id` audit fields are real progress, but they cover GitHub's own agent, not cross-tool identity. ([GitHub MCP configuration](https://github.com/github/github-mcp-server/blob/main/docs/server-configuration.md), [agentic audit events](https://docs.github.com/en/copilot/reference/agentic-audit-log-events))

4. **Model and vendor lock-in.** GitHub Agent HQ is an "any agent" control plane only on paying tiers; third-party Claude Code/Codex delegation is excluded from Copilot Pro and included only on Pro+/Max/Enterprise. GitLab's agent platform is GitLab-shaped. A vendor-neutral MCP forge with BYO models is a real alternative for teams that do not want their agent layer tied to one cloud. ([GitHub Copilot plans](https://github.com/features/copilot/plans))

**A small self-hosted forge is structurally weak in four other places.**

- **No CI marketplace or Actions ecosystem.** Gitea reuses GitHub Actions' 20k+ plugins; Forgejo has its own runner ecosystem. exist.io has checks and a reviewer, but not the long tail of integrations.
- **No network effect.** GitHub claims 180M developers, FedRAMP Tailored ATO, and SOC 2. At launch, exist.io has none of that paper.
- **Real ops burden.** Patching, backups, runners, storage, HA, LFS at scale, and upgrades are continuous work. Forgejo's own guidance says: "None of this is difficult. All of it is continuous."
- **Compliance is a multi-month project.** SOC 2, FedRAMP, pen tests, and a trust page are not launch-week deliverables. Regulated buyers will ask.

**Direct "forge for agents" entrants are already here.**

- **Cursor Origin** (early beta, 17 August 2026): Cursor's own Git forge with repos, clone/push/pull, GitHub mirroring, PRs, code search, CLI, API, and cloud agents on Origin repos. Included from Premium; no separate price; no GA date. This is the clearest "agent platform ships a Git backend" threat. ([Cursor Origin docs](https://cursor.com/docs/origin), [changelog](https://cursor.com/changelog/origin-code-hosting))
- **Entire** (product launch 26 October 2026): founded by former GitHub CEO Thomas Dohmke; **$60M seed at a $300M valuation** (February 2026); replaces the pull request with "Trails"; distributed Git hosting in the US, EU, India, and Australia; free hosting, agents pay-as-you-go at model provider rates; self-hosting of Git nodes is "upcoming." The funding and founder distribution make it the most credible direct threat. ([Entire launch](https://entire.com/news/entire-launches-an-independent-git-forge-for-agent-human-collaboration), [seed round](https://entire.com/news/former-github-ceo-thomas-dohmke-raises-60-million-seed-round))
- **Open-source challengers**: **Stratum** (evaluation-gated merges, policy-as-code, ephemeral workspaces, MCP, layer-over-GitHub or full forge, self-host on Cloudflare), **Diffinery** (GitHub-API-compatible forge with agent accounts, sandboxes, official-registry MCP, Cloudflare-hosted), and **Citadel** (self-hostable Go monolith with agent identity, knowledge graph, MCP, air-gap for regulated/defense). Their existence validates the thesis and narrows the window. ([Stratum](https://usestratum.dev/), [Diffinery](https://diffinery.com/agents), [Citadel](https://we.rethunk.tech/citadel))

**RhodeCode is the closest roadmap analogue, but not a shipped one.** Its Agentic AI page markets first-class agent accounts, path-level ACLs with deny-wins, IP binding, rate/spend caps, one-action revoke, and a line-level provenance ledger. The verifier found that path ACLs, IP binding, self-hosted deployment, and BYO-model review are shipped, but **agent accounts, spend caps, one-action revoke, and the provenance ledger are marked "In build · 5.15"** and are not in the 5.15.0 release notes. Treat RhodeCode as validation of the feature set, not proof that a competitor has already shipped it. ([RhodeCode Agentic](https://rhodecode.com/agentic/), [roadmap](https://rhodecode.com/roadmap/))

**Where the defensible gap is.** Not "agents in Git." The gap is **proof and isolation for agents you do not fully trust, on infrastructure you control**:

- **Per-agent identity and scoped leases**, not shared PATs. An agent gets a short-lived credential bound to a repo, a ref prefix, an action set, and an expiry. Revocation is one action.
- **Policy and audit at the forge layer**, not the harness. The forge can reject a push to `main` regardless of which agent or harness is used, and can record who (human or agent) did what, with model, prompt hash, cost, and approver.
- **Vendor-neutral MCP and BYO models.** Any agent (Claude Code, Codex, Cursor, Cline, custom) can connect; no inference resale, no model lock-in.
- **Air-gap and single-binary operations.** Pure Go, zero CGO, embedded UI, SQLite WAL, no host OpenSSH. This is the deployment story that GitLab charges **$299/user/month** for.

**ICP, in order of urgency:**

1. **Air-gapped / defense / federal / national labs.** Cannot use cloud agents; already pay enterprise prices for isolation; procurement language eliminates US-hosted SaaS.
2. **EU / sovereign platform teams.** Want Apache-2.0, data residency, and no GitLab Ultimate pricing; care about audit and vendor neutrality.
3. **Agent-heavy platform engineering teams.** Run many concurrent coding agents; need per-agent scopes, ref isolation, review gates, and cost control more than a CI marketplace.
4. **Cost-sensitive mid-market.** Replacing per-agent seats or credits with self-hosted compute; the unit economics matter more than the feature list.

**Competitive risks.** GitHub can bundle Agent HQ into Enterprise at $21 and erode willingness to pay separately. Gitea/Forgejo can add native agent identity for free. Cursor Origin and Entire normalize "forge + agent + memory" as one product. The category may commoditize into a feature of CI/identity platforms. The window is real but not indefinite; 6–12 months is the planning horizon.

**Positioning implication:** retire "the first Git forge for agents." Replace it with a claim that is true and defensible: **"The self-hosted forge for agent fleets that need proof."** Proof of identity, proof of scope, proof of review, proof of audit. The demo must show an agent being contained, not just an agent being productive.

---

### 4. Demand, pain points, and objections

**The strongest open evidence is agent git-footgun and isolation failure, not forge-level branch-protection bypass.** This distinction matters for credibility.

- Claude Code issue **#48474** ("Claude repeatedly commits and pushes directly to main despite explicit instructions") is **closed as not planned/stale**. It contains the reporter's assertion that branch protection was bypassed, but no protection config or push-rejection output. A commenter notes the push "would be caught by branch protection rules," implying protection may not have been configured. It proves an agent pushed to `main`; it does **not** prove a forge bypass. ([#48474](https://github.com/anthropics/claude-code/issues/48474))
- Issue **#81115** is also **closed**. The agent ran `git checkout -b feature-branch origin/dev`, silently setting upstream to `origin/dev`; the **human then ran `git push`**, landing commits on shared `dev` with no PR. It proves a git upstream footgun and misconfiguration; it does not prove the agent bypassed protection. ([#81115](https://github.com/anthropics/claude-code/issues/81115))
- Issue **#83311** is **open** and is the strongest evidence: with five `isolation:"worktree"` agents, two isolated correctly and **three contaminated** — one agent's commit landed on another's branch, stashes were created in the parent repo, and the user's uncommitted work was stashed/reverted. Two commenters reproduced it. This is **local worktree isolation failure**, not server-side branch-protection bypass, but it is exactly the class of problem a forge-level lease and ref isolation can solve. ([#83311](https://github.com/anthropics/claude-code/issues/83311))

**Other real failure modes:**

- **Worktrees are not a security boundary.** An agent can write `.git/hooks` and persist code execution into later commits. Teams block `.git/hooks` and proxy GitHub tokens so they are never exposed to the agent. ([HN](https://news.ycombinator.com/item?id=49110389))
- **Repo-scoped agents do not stay repo-scoped.** On the GitLost prompt-injection thread: "The fallacy here is expecting an agent that has access to ALL your repos to respect the singular repo it's in. It won't." The root cause is failure to scope permissions. ([HN](https://news.ycombinator.com/item?id=48829383))
- **Review, not generation, is the bottleneck.** Godot's 2026 contribution policy bans autonomous agents and says "PR review is the largest bottleneck in the engine right now." A forge that adds review capacity sells into the real constraint; one that adds PR volume is a liability. ([Godot](https://godotengine.org/article/contribution-policy-2026/))
- **Agent retry loops blow up CI spend.** An r/cicd report: "roughly $100 to $120 used to cover four or five devs. It rose to $500 to $600 per dev" as agents pushed, waited, failed, and re-pushed; every agent commit retriggers workflows. Self-hosted runners change the unit economics, not the loop. ([r/cicd](https://www.reddit.com/r/cicd/comments/1wurwho/github_actions_burned_through_my_ci_budget/))
- **Secret leakage rises with agent-authored commits.** GitGuardian reports "Claude Code-assisted commits showed a 3.2% secret-leak rate, versus a 1.5% baseline," and found 24,008 secrets in public MCP configs, 2,117 of them valid. A forge with lease-scoped credentials and pre-receive scanning has a concrete security story. ([GitGuardian](https://blog.gitguardian.com/the-state-of-secrets-sprawl-2026))

**The buying case is strongest where compliance or air-gap is non-negotiable.** A compliance consultancy describes the procurement clause: "That single line in a procurement contract eliminates GitHub.com, GitLab.com, Bitbucket Cloud, and every other US-hosted source control platform from the architecture." But self-hosting then becomes "production infrastructure for the development lifecycle," inheriting patching, backup, access review, and log-forwarding burden. The buyer is not choosing "free vs paid"; they are choosing "cloud agent I cannot use" vs "self-hosted agent stack I must operate." ([Truvo](https://truvocyber.com/blog/soc2-sdlc-self-hosted-gitlab))

**Vendors already charge enterprise prices for this wedge.** IBM launched self-hosted Bob for "AI sovereignty and governance"; Helix sells an eight-week self-hosted agent pilot **from $75K** with "ephemeral per-task git credentials" and "branch-scoped access — agents can only push to the branch they were assigned. No accidental access to main." The pain is budgeted. The open question is who owns the Git layer. ([IBM](https://uk.newsroom.ibm.com/IBM-Introduces-Self-Hosted-Deployment-for-IBM-Bob-to-Help-Enterprises-Advance-AI-Sovereignty-and-Governance), [Helix](https://www.helix.ml/use-cases/enterprise-coding-agents))

**The counter-evidence is serious.**

- Many teams describe GitHub's built-in agentic tooling as "excellent integration"; their complaints are credit pricing and review load, not a missing forge. Pain that can be fixed with harness hooks does not create a migration. ([r/ExperiencedDevs](https://www.reddit.com/r/ExperiencedDevs/comments/1tw5622/for_folks_heavily_using_a_agentic_engineering/))
- Blocking force-pushes and pushes to `main` can be "about a 30sec configuration" in the agent harness. Teams already wrap agents in OS-level sandboxes and token proxies. If the fix is a config file, a new forge is a hard sell. ([HN](https://news.ycombinator.com/item?id=49110389))
- The self-hosted ecosystem gap is real: "GitHub for community. Forgejo for control." Teams keep GitHub for reach and CI, and self-host only where control justifies the cost. ([Derek Armstrong](https://derekarmstrong.dev/blog/why-forgejo-git-ci-and-you-should-too))

**Messaging implication:** do not say "GitHub is broken" or "agents bypass branch protection." Say: **"Agents are powerful but over-permissioned. Give them a lease, not your token."** Then prove it with a reproducible demo: an agent attempts `main`, gets rejected; another pushes to `agent/feature-123`, gets reviewed and merged; the audit log shows model, scope, expiry, and approver. That claim is true, differentiated, and does not depend on a closed issue.

---

### 5. Naming, brand, and Product Hunt listing assets (stop-ship)

**The name is a launch blocker, not a branding exercise.** `exist.io` is the exact-match domain of **Exist**, a live 13-year-old personal analytics product from Hello Code Pty Ltd, a two-person Melbourne company. It sells at **$6.99/month** across web, iOS, and Android, publicly commits to "No AI" and "No investors," and its App Store listing (id986201088) has 47 US ratings at 4.0. The domain was created on **19 June 2013**, expires **19 June 2027**, and is **clientTransferProhibited**. There is no normal registration path. ([Exist values](https://exist.io/about/values/), [App Store](https://apps.apple.com/us/app/exist/id986201088), [RDAP](https://rdap.identitydigital.services/rdap/domain/exist.io))

**The forge does not own its own name.** Its canonical URL is `ravaniroshan.github.io/exist.io` and its repo is `github.com/RavaniRoshan/exist.io`. Every user who types the product name as a URL reaches the life tracker. The repo was created on **14 October 2026** and has **0 stars, 0 forks, and 0 open issues** as of 8 November 2026. Renaming is cheapest right now; every star, backlink, and press mention raises the cost. ([Site](https://ravaniroshan.github.io/exist.io/), [GitHub API](https://api.github.com/repos/RavaniRoshan/exist.io))

**The legal picture is nuanced — get counsel.** A live USPTO search found no EXIST mark in classes 9 or 42 and no marks owned by Hello Code Pty Ltd, but the search is JavaScript-dependent and the negative is unverified. IP Australia lists three EXIST word marks, only one live, in **class 25 apparel** (Spendless Shoes) — not a software mark. Registration is not the test: conflict turns on **likelihood of confusion**, and identical names can coexist in unrelated fields but conflict when goods/services overlap. Exist ships downloadable apps and a hosted service, the same broad software/SaaS territory as a forge, so "different industry" is a weak defence. Do not treat the class 25 mark as a legal block; treat the domain, 13 years of SEO, and app-store identity as the practical block. ([USPTO](https://tmsearch.uspto.gov/search/search-results?q=EXIST), [IP Australia](https://search.ipaustralia.gov.au/trademarks/search/view/773543), [Trama](https://www.tramatm.com/blog/category/legal/is-your-business-name-already-taken-how-to-avoid-trademark-conflicts-before))

**Keeping the word forces a lesser address.** `exist.com` and `exist.dev` are registered; `exist.tools`, `exist.software`, and `existforge.com` were unregistered in spot checks. None removes the collision or gives a clean branded search. If leadership insists on keeping "exist," launch under a distinct product name and use `getexist.dev`-style addresses only as redirects; expect to lose every direct-navigation and branded-search query.

**Rename precedent.** OpenTF became **OpenTofu** in September 2023 "due to trademark concerns," renamed the binary to `tofu`, the org, and the domain. Its public naming criteria demanded a name "distinct enough that it can be effectively googled." The real cost is surface area — binary, Go module path, docs, org, backlinks, install script, socials, PH slug — not the logo. A 2-week rename sprint is realistic now; a 2-month rename after launch is not. ([OpenTofu issue](https://github.com/opentofu/opentofu/issues/296))

**Naming criteria for the replacement:**

- 2–3 syllables, pronounceable, spellable after hearing it once.
- Not generic "forge/git/agent/repo" (Forgejo, Gitea, GitLab already own the vocabulary; descriptive names are crowded and hard to search).
- `.com` or `.dev` available, plus GitHub org, X/LinkedIn handles, and PH slug.
- No live software/SaaS trademark in the US, EU, UK, or Australia; screen before committing.
- Works as a binary name (`existd` → `newd`), Go module path, Docker image, Helm chart, and CLI verb.
- Does not collide with an existing dev tool, crypto project, or AI product.

**TLD economics.** `.com` costs **$13–20/year** and carries the highest trust; `.io` costs **$44–60/year**, is strong with developers, but is a ccTLD with long-term Chagos/Mauritius tail risk; `.dev` is Google-run, a top-20 gTLD, HSTS-preloaded (HTTPS mandatory), and explicitly marketed for developer projects. For a self-hosted infrastructure product, `.dev` is a credible primary if `.com` is taken. ([HostAccent](https://www.hostaccent.com/blog/com-vs-net-vs-io), [get.dev](https://get.dev/))

**Product Hunt listing specs (2026):**

| Field | Limit / spec |
|---|---|
| Name | ≤40 characters |
| Tagline | ≤60 characters |
| Description | ≤500 characters |
| Logo | 240×240 JPG/PNG/GIF, ≤2MB |
| Gallery | ≥1 image; 3+ recommended; **1270×760** (1.67:1); viewable at 2+ |
| Video | Optional; YouTube link; PH endorses a simple screen recording over a trailer |
| Tags | Up to 3 required launch tags |
| Pricing | Free / Paid / Free options |
| Promo code | 25-char text / 15-char code |
| First comment | ≤5,000 characters |
| Scheduling | Up to 30 days ahead; day starts **12:01 AM PT** |
| Account | Personal account required; company accounts cannot post; co-makers need accounts |
| Open source | Company info can mark the product open source and link the repo |
| Product link | Homepage, app store page, or GitHub repo |

Sources: [PH product page guide](https://help.producthunt.com/en/articles/17834537-4-build-your-product-page), [featuring guidelines](https://help.producthunt.com/en/articles/9883485-product-hunt-featuring-guidelines).

**The gallery is the pitch.** Top dev-tool galleries read as a story: hero/OG frame → real UI → proof → CTA. Aikido opened with its OG image and closed with a CTA; Appwrite anchored on "The open-source Vercel alternative"; Ronak's teardown used custom UI images and technical bullets, no stock. PH asks for product-in-use, before/after, or a key-moment GIF; GIFs animate on hover, so frame one must stand alone. Recommended five-frame sequence for exist.io:

1. **Hero**: "Your agents get a lease, not your token." A run detail showing repo, ref `agent/feature-123`, lease scope, policy verdict, and a merge button.
2. **Fleet**: the global Runs/Inbox view with 6–8 concurrent agents, status, repo, cost, and risk.
3. **Proof**: the diff review with inline comments, a blocked push to `main`, and the audit entry showing model, scope, expiry, approver.
4. **Deploy**: the self-hosted single binary / air-gap architecture; terminal shot with `existd`, SQLite WAL, SSH `:2222`, MCP `/mcp`.
5. **CTA**: "Self-host in 5 minutes. Or try the hosted demo." QR/URL, open-source badge, GitHub stars.

**The maker first comment is prime conversion surface.** Who you are, why you built it, what it does, who it is for, what is different, one specific question, any offer. Aikido and Appwrite replied to every comment. PH says ask for feedback, never upvotes. ([Aikido teardown](https://github.com/fmerian/awesome-product-hunt/blob/main/teardowns/aikido.md), [Appwrite teardown](https://github.com/fmerian/awesome-product-hunt/blob/main/teardowns/appwrite.md))

**Brand system benchmark.** Coder uses a monochrome-only logo (white or black), a shorthand mark for favicon/social, clear space equal to logo height, a mono typeface for technical labels, and an explicit **80/20 dark/light split** — dark for developers, light for enterprise buyers. Dark-mode marks should be designed, not inverted; contrast beats hue fidelity. exist.io's Patina palette can carry this; the logo needs a shorthand mark that survives 240×240 and a favicon. ([Coder brand](https://coder.com/brand))

**Recommendation:** **rename before launch.** Run a 10–14 day sprint: 20 candidates → trademark/domain/handle screen → 3 finalists → pick one → rename repo, Go module, binary, docs, install script, socials, PH slug. If the rename is rejected, launch under a distinct product name, keep `exist.io` only as a legacy redirect, and document the branded-search loss as an accepted risk.

---

### 6. Agents-first information architecture

**The industry has not replaced Git; it has wrapped agents around it.** GitHub's stated direction is continuity: agents should be "native to the GitHub flow," and mission control is "not a single destination" but a consistent interface across GitHub, VS Code, mobile, and CLI, with Git, PRs, and issues as unchanged primitives. In January 2026, GitHub then re-contained its own global mission control into a repository **Agents** tab, placing sessions "alongside your code, pull requests, and issues." The repo, not the fleet, remained the anchor. ([GitHub welcome home agents](https://github.blog/news-insights/company-news/welcome-home-agents/), [Agents tab](https://github.blog/changelog/2026-01-26-introducing-the-agents-tab-in-your-repository/))

**Five IA archetypes are visible in shipped products:**

| Archetype | Examples | Strength | Weakness | Fit for a forge |
|---|---|---|---|---|
| **A. Repo-anchored agent tab** | GitHub Agents tab | Git permissions, branches, diffs, review remain the frame; familiar | Weak cross-repo fleet visibility; duplicate global/repo lists | **Best default** |
| **B. Agent-first multi-repo workspace** | Cursor 3, Codex Cloud | Powerful parallelism; local/cloud/SSH/Slack/Linear in one sidebar | Loses the repo as spatial anchor; review can regress | **Partial — for fleet view only** |
| **C. Delegation inbox / issue-native** | Linear agents, Jules | Strong triage and intake; plan approval before edits | Weak deep-review depth; Jules shuts down 31 Dec 2026 | **Partial — for inbox** |
| **D. Run/control-plane console** | Temporal UI | Excellent audit/debug; filters, saved views, history, relationships | Operator jargon; dashboard sprawl | **Partial — for ops/audit** |
| **E. Governance control plane** | GitHub enterprise AI controls | Central policy, audit, MCP allowlists, agent sessions | Should not share the daily work surface | **Yes — separate admin IA** |

Sources: [GitHub Agents tab](https://github.blog/changelog/2026-01-26-introducing-the-agents-tab-in-your-repository/), [Cursor agents window](https://cursor.com/docs/agent/agents-window), [Codex Cloud](https://developers.openai.com/codex/cloud), [Linear AIG](https://linear.app/developers/aig), [Jules docs](https://jules.google/docs), [Temporal UI](https://docs.temporal.io/web-ui), [GitHub agent control plane](https://github.blog/changelog/2025-10-28-enterprise-ai-controls-the-agent-control-plane-are-in-public-preview/).

**Cursor 3 is the cautionary tale.** Its agent-first rebuild made changes harder to see, required opening each diff bubble, rendered whole files as new, and removed `@Docs` from the new surface. Agent-first shells can regress review when the diff is not the spine. A forge cannot afford that: review is the product. ([Cursor forum](https://forum.cursor.com/t/cursor-3-agents-window/156509))

**The trust ladder is the best pattern for leases.** Claude Code's permission modes — Manual (reads only, approved edits) → acceptEdits → plan → auto (classifier-reviewed) → dontAsk → bypass — with a persistent status indicator and plan approval that explicitly chooses auto versus per-edit review, are the clearest available model for rendering agent capability. Map each lease to a visible mode, scope, expiry, and policy verdict. ([Claude Code permission modes](https://code.claude.com/docs/en/permission-modes))

**Logs should be one stream with filters, not many tabs.** Devin collapsed Progress and Shell into one searchable Activity stream with type filters and a pinnable detail drawer; rows fold under the user message, shell cards list edited files, and the task list filters activity. It also attaches provenance to the exact commit: "latest changes haven't been reviewed," the last reviewed commit, and one-off or every-push review. Its notification inbox has All/Unread, per-type switches, row mute, hover preview, and automation failures notify on the 1st, 3rd, and 9th consecutive failure rather than every run. That is the right alert-fatigue model. ([Devin release notes](https://docs.devin.ai/release-notes/overview))

**Progressive disclosure from story to span.** LangSmith makes the thread the navigation unit with three lenses — Trajectory (conversation), Turns (per-turn cards), Details (inputs/outputs/timing/tokens/errors) — plus grouped tool calls, inline subagents, and per-block token/cost/model. Vercel AI Gateway separates fleet economics (requests by model, TTFT, tokens, spend) from a filterable request log with per-request routing drilldown. Cost and token meters belong on an ops surface, not the primary run list. ([LangSmith](https://docs.langchain.com/langsmith/view-traces), [Vercel AI observability](https://vercel.com/docs/ai-gateway/observability-and-spend/observability))

**A useful 2026 pattern library** models delegation in six phases — intent, plan, execution, checkpoint, result, audit — with an editable plan, live step ledger, tool-call cards, tiered gates, batch approval, interrupt/redirect, diff-first summary, a verified/assumed/could-not-do footer, an action trail with undo, and human takeover. Its anti-patterns are the ones to design out: infinite spinner, approval fatigue, confident wrong result, chat log as audit trail, hidden autonomy, fake progress. ([Setproduct](https://www.setproduct.com/blog/ai-agent-ui-design-patterns))

**Accessibility and agent-legibility converge.** NN/g argues that descriptive labels (no icon-only controls), predictable patterns, logical hierarchy, and semantic HTML/ARIA help screen readers and agents alike. Dense professional tables need those semantics and keyboard paths, not lower density. ([NN/g](https://www.nngroup.com/articles/ai-agents-as-users/))

**Recommended IA for exist.io: a hybrid, with the run as the stable object and the diff as truth.**

**Global shell (left rail):**

- **Inbox** — cross-repo triage: runs needing input, review, approval, or a failed check.
- **Runs** — fleet table: agent, repo, ref, state, policy verdict, cost, duration; saved views; bulk actions.
- **Reviews** — queue sorted by risk/policy; diff-first; inline comments; merge gate.
- **Policies** — lease scopes, ref rules, MCP allowlists, approval thresholds, audit export.
- **Repos** — the existing repo catalog; a repo opens into a repo-scoped shell.
- **Search / Command** — RAC Autocomplete + Menu + Dialog; keyboard-first.

**Repo shell (tabs):** Code · Issues · Pull requests · **Agents** · Checks · Settings. The Agents tab is a sibling, not a replacement. It shows sessions filtered to this repo, with one-click PR links and archive.

**Run detail (the core screen):**

- **Header**: state (queued/running/waiting/review/merged/failed), repo, branch/ref, commit SHA, lease chip (scope + expiry), policy verdict, cost/tokens, human owner.
- **Tabs**: Overview (intent, plan, summary, verified/assumed/could-not-do), **Changes** (diff-first, inline comments, split/unified, word wrap), Checks, Trace (tool calls, logs, timing), Audit (who/what/when, model, prompt hash, approver).
- **Action bar**: approve, request changes, steer/interrupt (if live), merge, revoke lease. Merge is gated by policy and review.
- **Every run has a stable ID** that resolves to a commit, PR, and diff. Chat is annotation; the diff is the artifact.

**Fleet view:** a virtualized table with columns for run, agent, repo, ref, state, policy, review, cost, duration, last activity. Filters and saved views (Needs me, Failing checks, Over budget, Unreviewed). Bulk actions (cancel, revoke, archive). This is where a platform team manages 50 agents; it should not be the default landing for a single-repo developer.

**Onboarding:** connect an agent via an MCP snippet, create a first lease, run a seeded demo, see the diff and merge. No mock data. Empty states teach the next action.

**What not to build:** a chat-only interface; a dashboard that hides the diff; a separate "agent world" disconnected from repos; a global fleet view with no repo anchor; approval modals for every action; an audit trail made of chat logs.

---

### 7. Visual design direction

**The category has converged, and the convergence is the problem.** Linear, Vercel, Supabase, Raycast, and Resend all use a near-black or true-white canvas, one chromatic accent, hairline borders, and mono metadata. The palette is table stakes. LogRocket argues the "Linear look" is saturated and generic; differentiation now has to come from type, texture, proprietary data views, and product imagery. exist.io should keep a dark shell but not make indigo the brand. ([LogRocket](https://blog.logrocket.com/ux-design/linear-design))

**What the best systems actually do:**

- **Linear**: 23 colors, 13 type tokens, 8px radius, accent only on CTA/focus/link, 150–200ms `cubic-bezier(0.2,0,0,1)` motion, display type with measured negative tracking, 13px dense UI. Its near-black is not `#000`, and borders are 5–8% white. The closest template for a dense forge — and the most copied. ([Shadcn Linear](https://www.shadcn.io/design/linear))
- **Vercel Geist**: 10-step color scales with explicit roles (100–300 backgrounds, 400–600 borders, 700–800 high-contrast fills, 900–1000 text/icons), P3 support, Geist Sans + Mono, a 72→12 type scale. Rules: draw each border once, keep material shadows translucent, use gray-alpha only when the surface must show through. ([Geist colors](https://vercel.com/geist/colors))
- **Supabase**: "timelessness" ("will this still feel good in a few years?"), "less is more", brand green only for CTA; Figma variables export to Tailwind so design and code cannot drift. Copy the token pipeline, not just the dark green. ([Supabase design](https://supabase.com/blog/how-design-works-at-supabase))
- **Sentry's 2025 redesign**: tactile depth — buttons lift on hover, inputs recess, switches have physical state — after concluding "polish is the baseline. Every app looks good." It took 18 months and hundreds of components. One or two tactile signatures will photograph better than another gradient. ([Sentry](https://blog.sentry.io/sentry-has-a-bold-new-look))
- **PostHog**: the counterexample that proves personality can beat category convention — "opinionated, human, slightly weird, playful"; it competes with "cat videos and TikTok, not B2B SaaS competitors"; cream canvas, hedgehog mascot, single yellow-orange CTA. Patina is a similar asset; use it as a material metaphor, not a label. ([PostHog brand](https://posthog.com/handbook/brand/foundations))

**Typography: a sans + mono pairing, not mono-everything.** Vercel uses Geist/Geist Mono, Neon uses Inter/GeistMono, Resend uses Domaine + Commit Mono, Warp uses Inter + Instrument Serif italic + DM Mono. Mono is reserved for SHAs, refs, diffs, commands, timestamps, and IDs; the sans carries hierarchy. Dense scales run 11–13px labels, 14px body, 16px modal copy, with tabular numerals wherever values change. Mona Sans is credible, but exist.io lacks a signature mono/display pairing. ([Refero](https://styles.refero.design/style/0d914ef0-fa84-4c60-a9aa-cef0b5eb6e5d))

**Color: OKLCH is the right authoring space, but perceptual lightness is not contrast.** OKLCH gives stable hue, predictable lightness, and P3 gamut. Use APCA to check contrast, and give dark mode a narrower L range and lower chroma rather than inverting light values. Patina's OKLCH base is an asset; it needs documented ramps, not hand-picked swatches. ([Evil Martians](https://evilmartians.com/chronicles/oklch-in-css-why-quit-rgb-hsl))

**Accessibility is a launch asset, not a compliance checkbox.** WCAG 3/APCA is still draft as of April 2026; ship **WCAG 2.2 AA** as the legal baseline and use APCA as a supplementary readability check. Never encode CI, agent, or review status by hue alone: pair color with icon, text, or shape (pass ✓, fail ✕, running spinner, blocked lock). Red/green diff states are the classic failure; a forge also needs ± markers and labels in diffs. ([Roselli](http://adrianroselli.com/2026/04/wcag3-contrast-as-of-april-2026.html), [Smashing](https://www.smashingmagazine.com/2024/02/designing-for-colorblindness))

**Dark vs light is unresolved.** NN/g finds light mode generally better for reading speed and small text, dark mode preferred in low light, and recommends user choice. Best-in-class tools ship both or choose dark-first for the product and light for docs. For PH, a dark app screenshot on PH's white page has more pop; the safest gallery composition is a light or warm-neutral surround with a dark UI inside. ([NN/g dark mode](https://www.nngroup.com/articles/dark-mode))

**Density is contextual.** Dashboard systems recommend comfortable and compact modes; analysts on large monitors choose density, casual viewers need whitespace. For a forge, default to **compact 32px rows** with a comfortable toggle, and spend whitespace around primary actions and empty states. ([Fuselab](https://fuselabcreative.com/dashboard-design-system))

**Motion should be short, purposeful, and optional.** About 100ms for toggles/checks, 200–300ms for modals and large transitions; ease-out on entry, ease-in on exit; the more frequent the interaction, the shorter it should be. Excessive motion is an accessibility issue; `prefers-reduced-motion` must remove non-essential animation. Reserve animation for the demo video, not the screenshot. ([NN/g animation](https://www.nngroup.com/articles/animation-duration))

**Glassmorphism is an accent, not a foundation.** Translucent text over uncontrolled backgrounds fails contrast, and `backdrop-filter` creates GPU compositing layers that can drop frames on large scrolling surfaces. Use solid surfaces for nav, tables, and text; reserve blur for the command palette, a floating agent panel, or one hero overlay. Static grain is cheaper than animated blur. ([Setproduct glass](https://www.setproduct.com/blog/liquid-glass-vs-glassmorphism))

**The 2026 palette mood is a reaction against neon AI gradients.** Warm earthy tones, oxidized/patina colors, paper/stone surfaces, grain, and tactile texture are rising, driven by AI-content fatigue and a craving for "reality." Patina is on-trend — but earthy palettes can read lifestyle rather than infrastructure. Keep terminal/mono cues and one cool signal color so it still reads as a forge. ([Stauffer](https://www.stauffer.com/news/blog/why-2026-color-trends-are-signaling-a-craving-for-reality))

**Visual quality is a trust proxy, not proof.** NN/g's credibility research says design quality/organization, up-front disclosure, completeness, and external presence drive trust; typos and broken links degrade it fast. The popular "75% judge credibility by design" Stanford statistic is widely repeated but often decontextualized. A beautiful forge still needs real repos, real activity, and real security copy in the screenshots. ([NN/g trust](https://www.nngroup.com/articles/trustworthy-design))

**Recommended visual direction: "Patina, proven."**

- **Keep the Patina palette** (oxidized copper/teal) and Primer structure: 6px controls, 8px slabs, 1px hairlines, dark-first app, light docs.
- **Add a signature pairing**: Mona Sans for UI, a distinctive mono for refs/SHAs/logs, and a display cut for headlines. Do not switch to Geist or Linear's indigo.
- **One tactile signature**: a **lease seal** (a stamped chip with scope, expiry, and issuer) and a **merge gate** (a physical "click" when policy passes). These are the screenshots people remember.
- **Proof language**: ledger lines, stamps, ref prefixes, audit trails. The brand metaphor is patina — age, trust, evidence — not "AI magic."
- **Status system**: icon + text + color; ± diff markers; no hue-only meaning.
- **PH gallery**: dark UI on warm neutral surround; 1270×760; one idea per frame; captions in the image; no stock.

---

### 8. Frontend architecture

The existing app is **React 19 + Vite 8.3 + TypeScript + Tailwind 4.3 + React Aria Components 1.17 + motion**, with large hand-built views and a typed fetch client. There is **no router** and **no TanStack Query** in `package.json`. This is not a greenfield. Any recommendation that ignores the installed stack creates a migration project on top of a redesign.

**Routing: TanStack Router is the strongest fit; React Router is the safer ecosystem bet.** TanStack derives path params, validated search params, and loader data through TypeScript inference over its route tree. React Router 7's strongest typing requires framework-mode route-module codegen and a watch step. A forge is full of filters, tabs, diff anchors, and pagination, so typed search params are core UX, not polish. The tradeoff: React Router has **75.2M** weekly downloads vs TanStack Router's **23.9M** (Nov 2026), so TanStack has a smaller ecosystem of examples, integrations, and hires. The old "20–30s IntelliSense stalls at ~100 routes" complaint is **misread**: the 2024 issue described about 20 pages and ~1s per-file type-check hotspots, and it was closed after a fix. Re-measure on this route tree rather than trusting either the old complaint or the fix. ([TanStack vs React Router](https://vercel.com/i/tanstack-router-vs-react-router), [issue #1091](https://github.com/TanStack/router/issues/1091), [npm React Router](https://api.npmjs.org/downloads/point/last-week/react-router), [npm TanStack Router](https://api.npmjs.org/downloads/point/last-week/@tanstack/react-router))

**Do not put all server state in route loaders.** A production team after eight months with TanStack Router used loaders for page-level data, Suspense queries for component-level data, and Query for interaction/deferred data. TanStack Router's cache does per-route dedupe, preload, stale-while-revalidate, and GC, but it explicitly lacks shared cross-route caching/dedupe, mutation APIs, and cache-level optimistic APIs. Use **TanStack Query** for shared server state, with `ensureQueryData` in loaders, and one cache. ([Swizec](https://swizec.com/blog/tips-from-8-months-of-tan-stack-router-in-production/), [Router data loading](https://tanstack.com/router/latest/docs/guide/data-loading))

**Server state: TanStack Query v5, with a known optimistic-update sharp edge.** It has **84.0M** weekly downloads. A documented v5 report shows `cancelQueries()` does not always stop an already-resolved stale refetch from overwriting optimistic data; merge the server response and avoid blanket invalidation. ([npm Query](https://api.npmjs.org/downloads/point/last-week/@tanstack/react-query), [discussion](https://github.com/TanStack/query/discussions/10712))

**API: OpenAPI + Hey API, not tRPC.** tRPC assumes a TypeScript server; the backend is Go. OpenAPI + Hey API generates typed fetch clients, TanStack Query hooks, and Zod schemas. Hey API claims Vercel, OpenCode, and PayPal usage, but that is vendor marketing; validate the Go OpenAPI spec's completeness before committing. If the spec is incomplete, keep the hand-written typed client and add codegen incrementally. ([tRPC](https://encore.dev/articles/trpc), [Hey API](https://heyapi.dev/))

**Real-time: SSE-first.** Agent token streams are server-to-client; SSE runs over plain HTTP and auto-reconnects, while WebSocket needs a custom upgrade path. Use WebSocket only when the client must send mid-stream (cancel, steer, tool approval) or needs low-latency bidirectional control. SSE fails at proxies, not at protocol throughput: Nginx buffers by default, so set `proxy_buffering off` and `X-Accel-Buffering: no`, send `data:\n\n` heartbeats every 15–30s, and remember Cloudflare's 100s idle timeout can inject an HTML error page into the stream. HTTP/2 avoids the 6-connection-per-origin cap. Native `EventSource` cannot set auth headers or POST bodies; use `@microsoft/fetch-event-source` for headers, POST, and custom retry. Budget for backpressure: SSE has no app-level backpressure; WebSocket servers need `bufferedAmount` caps and write timeouts. ([SSE vs WebSocket](https://tianpan.co/blog/2026/04/19/sse-websockets-grpc-streaming-llm-applications), [fetch-event-source](https://github.com/Azure/fetch-event-source), [websocket.org](https://websocket.org/guides/use-cases/ai-streaming/))

**Diff rendering is the highest-risk frontend decision.** It needs a virtualization-first engine. **Pierre** is the strongest 2026 candidate: `CodeView` virtualizes files and lines, uses inverse-sticky positioning to avoid blanking, runs Shiki in a worker pool, has an LRU highlight cache and DOM pooling. Its vendor benchmark cut a Linux v6→v7 diff from **2.4GB to 1.15GB** and parse time by ~80%. But `@pierre/diffs` first shipped on **10 December 2025**, is at v1.6.4, and has **76 open issues**: strong momentum, no multi-year forge track record. ([Pierre](https://pierre.computer/writing/on-rendering-diffs), [repo](https://github.com/pierrecomputer/pierre))

**Monaco is usually too heavy for a forge.** Sourcegraph measured Monaco at **2.4MB and 40% of search-page JS**; replacing it with CodeMirror cut page JS from **6MB to 3.4MB (−43%)** and fixed global-config, single-line, and CSS-theming problems. This is 2022 evidence, so re-benchmark, but the architectural lesson holds. Whole-DOM diff libraries collapse on large patches: an independent 2026 benchmark measured `react-diff-view` at **1.4s/133MB for 10k lines**, **7.6s/631MB/~13 FPS at 50k**, and **17s/1.3GB at 100k**. Caveat: the benchmark author maintains a competing library and did not test Pierre, so treat the numbers as directional. Recommended split: **Pierre for huge read-only PRs, CodeMirror 6 merge view for editable conflicts.** ([Sourcegraph](https://sourcegraph.com/blog/migrating-monaco-codemirror), [benchmark](https://dev.to/jiahangzhang/our-release-tool-crashed-on-a-large-pr-heres-the-react-diff-library-i-ended-up-writing-6i5))

**Keep React Aria Components; Base UI is not a migration mandate.** RAC documents 50+ components including Table, Tree, GridList, and NavigationTree, with 32 languages and 13 calendar systems, and normalizes mouse, touch, keyboard, and screen-reader input. Base UI is newer and became shadcn's default in July 2026, but shadcn explicitly says Radix is not deprecated and existing apps need not migrate; its migration path is progressive, one component per commit, not a codemod. RAC's stricter composition and `motion.create()` integration are real DX costs, so switch only for a concrete capability gap. ([Untitled UI](https://www.untitledui.com/blog/base-ui-vs-react-aria), [shadcn changelog](https://ui.shadcn.com/docs/changelog/2026-07-base-ui-default), [RAC DX](https://github.com/shadcn-ui/ui/issues/1622))

**Build the command palette on RAC's own example, not cmdk by default.** RAC ships a command-palette example using Autocomplete + Menu + Dialog, with filtering and keyboard semantics handled. cmdk is far more downloaded (46.4M/week vs RAC 9.0M), but its repo was last pushed on 2025-10-29 and has 80 open issues; acceptable only for a standalone palette with maintenance risk accepted. Use RAC for the forge palette. ([RAC example](https://reactspectrum.blob.core.windows.net/reactspectrum/78a69daa88c1e2bd0cba1f5bb7c69024596d05bf/docs/react-aria/examples/command-palette.html), [cmdk](https://github.com/dip/cmdk))

**Virtualization should be split by surface.** TanStack Virtual is headless and better for tables, trees, and logs where you own layout; react-virtuoso has built-in dynamic heights, sticky headers, grouping, and prepend-without-jump for agent feeds and chat. Using one for everything forces either custom measurement or unnecessary API. ([PkgPulse](https://www.pkgpulse.com/guides/tanstack-virtual-vs-react-window-vs-react-virtuoso-2026))

**Charts: uPlot for telemetry, Recharts for product metrics.** uPlot is ~22KB gzip, renders 166,650 points in 25ms cold, and updates 3,600 points at 60fps at ~10% CPU/12.3MB per its README (vendor, older benchmark). Recharts (64.6M/week, ~136–150KB gzip) is fine for ordinary SVG dashboards under ~5k points; ECharts only for exotic or very large data. ([uPlot](https://github.com/leeoniya/uPlot), [chart survey](https://www.usedatabrain.com/blog/react-chart-libraries))

**Client state: Zustand for UI state, Query for server state.** Zustand has 92.1M weekly downloads, ~0.5KB gzip, two open issues, and active maintenance; Jotai fits fine-grained derived state; RTK only if the team is already standardized on it. Do not create a second server cache. ([Zustand](https://api.npmjs.org/downloads/point/last-week/zustand), [repo](https://github.com/pmndrs/zustand))

**Testing: Vitest + Testing Library + MSW + Playwright is the pragmatic default.** Vitest has 180M weekly downloads, Playwright 116M, MSW 33M. Vitest browser `toMatchScreenshot` handles visual regression, but screenshots need a pinned CI/browser environment; Storybook's Vitest addon did not support it as of Nov 2025, so do not make Storybook the visual source of truth. ([Vitest](https://main.vitest.dev/guide/browser/visual-regression-testing), [Storybook](https://github.com/storybookjs/storybook/discussions/32930))

**Vite 8 + React Compiler are worth adopting, but only with measurement.** Vite 8's Rolldown claims 10–30× faster builds (Linear 46s→6s, Ramp −57%, Beehiiv −64%) at the cost of a larger toolchain install; plugin-react v6 uses Oxc and exposes a React Compiler preset. React Compiler 1.0 is stable; Meta reports up to 12% faster initial loads/navigation and >2.5× on some interactions, Wakelet LCP −10%/INP −15%. These are vendor benchmarks. Set route-level JS budgets, lazy-load diff/editor/chart routes, and pin the compiler if test coverage is weak. The repo already ships Vite 8.3, so this is an incremental adoption, not a migration. ([Vite 8](https://vite.dev/blog/announcing-vite8), [React Compiler](https://react.dev/blog/2025/10/07/react-compiler-1))

**Migration: strangler fig, route by route.** Mount the new React 19 shell beside the existing app, move high-value read-only routes first, and keep old views until parity and rollback are proven. For design-system work, migrate one component per commit rather than a codemod. ([Microsoft](https://learn.microsoft.com/en-us/azure/architecture/patterns/strangler-fig), [shadcn](https://ui.shadcn.com/docs/changelog/2026-07-base-ui-default))

**Recommended stack:**

| Area | Default pick | Why / caveat |
|---|---|---|
| Routing | **TanStack Router** (library mode) | Typed search params; React Router 7 if ecosystem/familiarity dominates |
| Server state | **TanStack Query v5** | `ensureQueryData` in loaders; one cache |
| API | **OpenAPI + Hey API** | Go backend; skip tRPC; validate spec completeness |
| Realtime | **SSE-first** | WebSocket only for mid-stream control |
| Diff | **Pierre + CodeMirror 6** | Huge read-only PRs vs editable merge; prototype both |
| UI | **React Aria Components** | Keep; Base UI not a mandatory upgrade |
| Lists | **TanStack Virtual / react-virtuoso** | Tables/trees/logs vs agent feed |
| Palette | **RAC Autocomplete + Menu** | cmdk only with maintenance risk accepted |
| Charts | **uPlot + Recharts** | Telemetry vs product dashboards |
| Client state | **Zustand** | Query owns server state |
| Testing | **Vitest + Playwright + MSW** | Screenshot tests in pinned CI |
| Build | **Vite 8 + React Compiler** | Measure; pin compiler |
| Migration | **Strangler fig** | Route-by-route; old views until parity |

---

## Analysis & discussion

### Seven patterns that should shape the plan

1. **Product Hunt rewards coordination, not product quality alone.** Team/network breadth was the strongest predictor in the largest public dataset. A solo launch with 0 stars is playing a different game than Appwrite's sixth launch with 5.4k followers. exist.io needs to build the network before it needs the launch page.

2. **PH is a spike, not a funnel.** The median dev-tool launch converts 1–3% of a modest referral stream. The durable channels for OSS are Hacker News, niche Reddit, GitHub trending, changelogs, docs, and word of mouth. Plan PH as the loudest day in a 90-day campaign, not the campaign.

3. **The category is crowding fast.** GitHub, GitLab, Bitbucket, Cursor, Entire, and three open-source challengers all have an agent-forge story. "Agents in Git" is no longer a differentiator. The differentiator is **containment with proof**: per-agent identity, scoped leases, policy at the forge, audit, air-gap, vendor neutrality.

4. **Self-hosting is both a moat and a tax.** Compliance and air-gap buyers will pay; everyone else sees patching, backups, runners, and upgrades as a reason to stay on GitHub. The product must make operations boring: single binary, SQLite WAL, embedded UI, one-command bootstrap, backup/restore, upgrade path, Helm chart. The ops story is part of the sales pitch.

5. **The diff must remain the spine of the UI.** Cursor 3's regression is the warning. Agents can have a fleet view, an inbox, a trace, and a chat, but the artifact humans approve is the diff. Every run must resolve to a commit, a PR, and a reviewable change.

6. **Visual differentiation is compressing.** The Linear/Vercel look is everywhere. Patina is a genuine asset because it is warm, tactile, and unusual in infrastructure; the risk is reading as lifestyle rather than engineering. Keep mono/terminal cues and one cool signal color. The signature should be a lease seal and a merge gate, not another gradient.

7. **The frontend should evolve, not restart.** The existing stack is modern (React 19, Vite 8, RAC, Tailwind 4, Vitest 5). The gaps are routing, server-state, real-time, and diff rendering — four adoptions, not a rewrite. A strangler migration preserves the working views while the new shell and IA land.

### Strengths, weaknesses, opportunities, threats

**Strengths**

- Technically broad, hermetic, pure-Go, zero CGO, single binary.
- Native MCP plus read-only sandbox — vendor-neutral by design.
- Ephemeral HMAC leases with ref-prefix isolation — a real security primitive, not a marketing claim.
- Autonomous PR reviewer and policy/checks — directly addresses the review bottleneck.
- Apache-2.0 and self-hosted — credible for sovereign/air-gapped buyers.
- A working product at v0.6 with 137 frontend tests and a Go test suite, not a mockup.

**Weaknesses**

- The name is someone else's domain and brand.
- 0 stars, 0 forks, 0 open issues: no audience, no social proof, no community.
- No CI marketplace or Actions ecosystem; checks are not a substitute for the long tail.
- No SOC 2, FedRAMP, pen-test report, or trust page yet.
- Small team; the frontend is large hand-built views with no router or query layer.
- No hosted demo; PH visitors will not self-host on launch day.
- No published security proof; the wedge currently rests on competitor issues, some closed.

**Opportunities**

- GitLab's **$299/user/month** air-gapped agent SKU is a price umbrella.
- GHES cannot run most Copilot agent features on-prem; GitHub's cloud agent assumes GitHub.com.
- Codeberg now restricts LLM-generated content; Google is winding Jules down on 31 December 2026; category volatility favors portable, self-hosted infrastructure.
- Agent CI cost blowups ($100–120 for 4–5 devs → $500–600/dev) create a budget argument for self-hosted compute.
- Secret leakage in agent-authored commits (3.2% vs 1.5%) creates a security argument for lease-scoped credentials and pre-receive scanning.
- Regulated/air-gapped buyers already pay $75K+ for agent pilots; a forge that owns the Git layer can be the durable substrate.

**Threats**

- GitHub bundles Agent HQ into Enterprise at $21 and erodes willingness to pay separately.
- Gitea/Forgejo add native agent identity for free and own the "free self-hosted" mindshare.
- Cursor Origin and Entire normalize "forge + agent + memory" as one product; Entire has $60M and a former GitHub CEO.
- The category commoditizes into a feature of CI/identity platforms.
- PH's featured gate can make a good launch invisible.
- A name collision or trademark dispute can consume the launch window.
- The security wedge can be dismissed as "a config file" if the demo is not reproducible.

### The strategic choice

There are three plausible strategies:

| Strategy | What it is | Pros | Cons | Verdict |
|---|---|---|---|---|
| **A. Sovereign agent forge** | Self-hosted, air-gapped, auditable forge for regulated/sovereign teams | Clear ICP, high willingness to pay, defensible vs cloud-only, undercuts GitLab $299 | Smaller market, long sales cycles, compliance burden | **Recommended** |
| **B. Hosted agent forge for SMB** | Cloud product competing with GitHub/Cursor on developer experience | Large market, self-serve | Head-on with GitHub/Cursor, no distribution advantage, price pressure | **Avoid at launch** |
| **C. OSS infrastructure layer** | Apache-2.0 forge that other agent platforms embed | Distribution via OSS, vendor-neutral, ecosystem | Monetization is indirect, support burden, clones | **Use as distribution, not the whole business** |

The recommended shape is **A with C as distribution and a hosted tier as convenience**: Apache-2.0 core, free self-host, paid hosted sandbox for evaluation, enterprise license for air-gap/SSO/audit/support. Do not try to out-GitHub GitHub on developer experience in year one.

### What "Top 10" actually requires

Top 10 overall in late 2026 is roughly **90–160 points**, but the distribution is wide and the safe target is **150–250**. That requires:

- **1,000–2,000 engaged developers** who can act in the first four hours.
- **5k+ GitHub stars** or an equivalent proof of community.
- **3–5 design-partner logos** and 2–3 quotable testimonials.
- A **working hosted demo** that a PH visitor can try in under 60 seconds.
- A **featured-worthy listing**: real product, real screenshots, high craft, no vaporware.
- A **coordinated week**: PH, HN, r/selfhosted, r/devops, X, LinkedIn, newsletters, Discord/Slack.
- A **war room** for the first four hours and a reply to every comment.

With 0 stars today, this is a 90-day audience build, not a 2-week asset sprint. That is the honest gap between the current state and the goal.

### The frontend thesis

The radical redesign should be **radical in IA, conservative in stack**. Radically reorganize around the run, the lease, and the diff; keep the familiar Git primitives; keep Patina, RAC, Tailwind, and Vite; add the four missing layers (router, query, SSE, diff engine); migrate route-by-route. The result should feel like a control plane for agent fleets that happens to be a Git forge — not a GitHub clone with an "Agents" tab bolted on, and not a chat app that forgot the code.

---

## Conclusions & implications

1. **Rename first.** The current name is not ownable, not searchable, and sends users to a different product. The repo is three weeks old with zero stars; this is the cheapest it will ever be. If the rename is rejected, launch under a distinct product name and treat `exist.io` as legacy.

2. **Reposition from "first" to "proof."** The first-mover claim is false. The defensible claim is self-hosted, air-gapped, auditable, vendor-neutral containment for agent fleets. That claim is true today, matters to a buyer with budget, and is hard for GitHub/Cursor to copy without cannibalizing their cloud model.

3. **Treat PH as a campaign, not a day.** Build 1,000–2,000 engaged developers, 5k+ stars, 3–5 design partners, and a hosted demo before launch. Target overall Top 10 at 150–250 points; stretch Top 3 at 300+. Do not promise #1.

4. **Make the demo the product.** The launch asset is not a landing page; it is a reproducible containment demo and a hosted sandbox. A PH visitor must see an agent get a lease, push to an isolated ref, get reviewed, and merge — in under 60 seconds, without installing anything.

5. **Redesign the IA around the run, but keep Git as the substrate.** Global Inbox/Runs/Reviews/Policies + repo Agents tab + run detail with diff as truth. Do not replace PRs with a novel object; do not hide the diff; do not build a chat app.

6. **Keep Patina and the existing stack; add four layers.** TanStack Router, TanStack Query, SSE, and a virtualization-first diff engine. Migrate route-by-route. The radical part is the IA and the proof language, not a greenfield rewrite.

7. **The business model is a hypothesis, not a finding.** The research supports a wedge and a price umbrella, not a specific price. Validate willingness to pay with 10–15 ICP interviews before publishing a pricing page. The likely shape is free Apache-2.0 core, paid hosted evaluation, enterprise license for air-gap/SSO/audit/support.

---

## Recommendations

### Immediate — next 14 days

**1. Make the name decision and start the sprint.**
Pick a launch name that is ownable and searchable. Run 20 candidates through: trademark screen (US, EU, UK, Australia), `.com`/`.dev` availability, GitHub org, X/LinkedIn handles, PH slug, and a Google/DDG search for existing dev tools. Pick three finalists, then one. Rename the repo, Go module path, binary (`existd` → new name), Docker image, Helm chart, install script, docs, socials, and PH draft. Budget 10–14 days. If leadership rejects the rename, document the branded-search and direct-navigation loss as an accepted risk, and use a distinct product name for the PH listing.

**2. Rewrite the positioning in one page.**
Use this structure: *For [ICP] who need [job], [product] is a [category] that [differentiator]. Unlike [alternatives], it [proof].* Example: "For platform teams running autonomous coding agents on regulated or air-gapped infrastructure, [name] is a self-hosted Git forge that gives every agent a scoped, expiring lease instead of a shared token. Unlike GitHub or GitLab, it enforces ref isolation, policy, and audit at the forge layer — with any agent, any model, one binary." Retire "first," "GitHub is broken," and "MCP is the moat."

**3. Build the containment demo.**
Two agents, one repo, one policy. Agent A gets a lease for `refs/heads/agent/feature-123` and succeeds. Agent B attempts `main` and is rejected. The audit log shows agent identity, model, prompt hash, scope, expiry, and approver. Record a 60–90 second screen capture with captions. Publish the reproducible script and a blog post. This is the proof asset for PH, HN, and sales.

**4. Stand up the hosted demo.**
Seed a realistic repo with commits, issues, a PR, and an agent run. One-click "run an agent" with a pre-approved lease. Show the diff, review, checks, audit, and merge. No mock data, no signup wall for the first run; capture email only for the waitlist or a second run. Put the demo URL in the PH listing, README, and docs.

**5. Recruit 5–8 design partners.**
Target the ICP in order: air-gapped/defense/federal, EU/sovereign platform teams, agent-heavy platform engineering, cost-sensitive mid-market. Offer white-glove install, a private Slack channel, and a roadmap call. Ask for three things: a quote, a logo (if allowed), and a written mini-case study with a number (time saved, agents contained, cost reduced).

**6. Instrument the baseline and set the goal.**
Track weekly: GitHub stars, unique clones, installs (curl script + binary downloads), waitlist signups, demo runs, activation (first repo, first agent run, first lease, first merge), and 7/30-day retention. Set the launch goal as **overall Top 10 (150–250 points)** and a stretch of **Top 3 (300+)**. Share the dashboard with the team; do not optimize for upvotes alone.

### 30–60 days

**7. Prototype and test three IA archetypes.**
Build clickable prototypes for (A) repo-anchored Agents tab, (B) agent-first fleet workspace, and (C) hybrid rail + repo tab + run detail. Test with 8–12 developers who use coding agents daily. Measure: time to find a running agent, time to review a diff, time to identify who approved a merge, trust rating, and preference. Pick the hybrid unless the data strongly says otherwise.

**8. Extend the design system.**
Document Patina ramps in OKLCH; add a signature mono and a display cut; define semantic tokens for run states, lease states, policy verdicts, diff states, and cost. Build the **lease seal** and **merge gate** as signature components. Ship dark and light themes; verify WCAG 2.2 AA and APCA; add colorblind-safe diff markers (± and labels). Add a compact/comfortable density toggle.

**9. Build the new shell.**
Global rail (Inbox, Runs, Reviews, Policies, Repos), repo Agents tab, run detail (Overview, Changes, Checks, Trace, Audit), fleet table with saved views, review queue, policy/audit admin, command palette, onboarding. Keep the existing views behind the new shell until parity.

**10. Lay the frontend foundation.**
Adopt TanStack Router + TanStack Query + OpenAPI/Hey API codegen. Add SSE endpoints on the Go side and a fetch-event-source client with heartbeats and reconnect. Spike Pierre and CodeMirror 6 on real 10k/50k-line diffs; measure memory, FPS, and time-to-first-paint. Add Zustand for UI state. Add Vitest/Testing Library/MSW/Playwright coverage for the new shell and the run detail. Set route-level JS budgets and lazy-load the diff/editor/chart routes.

**11. Start the content engine.**
Two posts per week: build-in-public, benchmark (agent containment, CI cost, secret scanning), architecture deep-dives, and customer stories. Post changelogs to r/selfhosted every two weeks (Postiz's compounding playbook). Publish docs that a skeptical platform engineer can follow: install, air-gap, SSO, backup/restore, upgrade, MCP setup, lease model, audit export.

**12. Prepare the PH pre-launch.**
Claim the product page, create the Forum Thread (Coming Soon pages are retired), and build the gallery and video. Draft the tagline (≤60), description (≤500), first comment (≤5,000), and the five-frame gallery. Schedule 30 days ahead once the date is fixed. Do not publish a teaser page that promises something the demo cannot do.

### 60–90 days

**13. Run a private beta with 20–30 teams.**
Instrument activation and the top five friction points. Fix onboarding, install, upgrade, and the first agent run. Publish a "known limitations" page; nothing kills trust faster than a launch demo that hides the sharp edges.

**14. Ship the security/compliance pack.**
Threat model, audit log documentation, SBOM, signed releases, security policy, responsible disclosure, backup/restore guide, air-gap guide, SSO/SCIM setup, HA/DR runbook. Start SOC 2 readiness (not certification) and publish a trust page with what is done, what is in progress, and what is not claimed. This is table stakes for the ICP.

**15. Validate pricing with 10–15 ICP interviews.**
Test three packages: (a) free OSS + paid hosted evaluation; (b) per-seat hosted + enterprise self-host license; (c) per-agent or per-run usage. Anchor below GitLab's **$299/user/month** air-gap SKU, but do not publish a number until you hear the same willingness-to-pay signal three times. Ask about budget owner, procurement path, and what they currently pay for the workaround.

**16. Finalize launch assets.**
Five screenshots at **1270×760** (hero, fleet, proof/diff, self-host/air-gap, CTA), a **240×240** logo, a 60–90 second video with burned-in captions, a maker first comment, an open-source badge, and a demo link. Test the gallery on a phone and at 240×240 thumbnail size. If the first image does not answer "what does this do?" in two seconds, redo it.

**17. Orchestrate launch week.**
Choose Tuesday–Thursday (more traffic, more competition) or Sunday (more Visit clicks, Monday newsletter). Go live at **12:01 AM PT**. Same day: Show HN at 6–9 AM PT, r/selfhosted and r/devops posts (follow each community's rules), X/LinkedIn threads, newsletters, Discord/Slack communities. War room for the first four hours; reply to every comment; no vote buying; no paid hunter. Have the hosted demo and status page monitored.

**18. Measure, publish, and iterate.**
Track PH rank/points, PH referrals, signups, stars, installs, demo runs, activation, and retention. Publish a transparent launch retrospective with the numbers. Convert the launch spike into a durable channel: newsletter, changelog cadence, design-partner stories, and a monthly agent-containment benchmark.

### Launch-week checklist

| When | Action |
|---|---|
| **T-14 days** | Freeze name, listing, screenshots, video, demo. Brief 30–50 supporters individually. Confirm PH date and schedule. |
| **T-7 days** | Test the demo on fresh devices and slow networks. Prepare HN/Reddit/X drafts. Confirm newsletter and community slots. Assign war-room roles. |
| **T-1 day** | Final QA: links, pricing, install command, MCP snippet, status page. Schedule the maker comment. Sleep. |
| **T-0, 12:01 AM PT** | Go live. Post the maker comment. Email the list. X/LinkedIn. Monitor comments and demo errors. |
| **T-0, 6–9 AM PT** | Show HN. Reddit posts. Community posts. Second outreach wave. |
| **T-0, all day** | Reply to every comment within 15 minutes. Fix demo issues live. Do not ask for upvotes. |
| **T+1 day** | Thank-you post, retrospective draft, convert waitlist, follow up with every commenter who asked a question. |
| **T+7 days** | Webinar or live demo, case study, changelog, second HN/Reddit wave if warranted. |

### What not to do

- **Do not launch as `exist.io`.** The domain, brand, and search results belong to another product.
- **Do not claim "first Git forge for agents."** It is false in November 2026.
- **Do not claim "GitHub agents bypass branch protection" without qualification.** The open evidence is worktree contamination, not a proven server-side bypass.
- **Do not buy votes or pay a hunter.** PH can unfeature the product and ban the maker.
- **Do not ship mock data or a demo mode in production.** The launch demo must use a real repo and real agent runs.
- **Do not replace pull requests with a novel object on day one.** Entire is making that bet with $60M; a small team should not.
- **Do not hide the diff behind chat.** Cursor 3 is the cautionary tale.
- **Do not adopt Linear purple, Geist scales, neon gradients, or glassmorphism.** Patina is the asset.
- **Do not rewrite the frontend greenfield.** Adopt four layers and migrate route-by-route.
- **Do not make PH the only channel.** HN and niche Reddit outperform it for OSS installs.
- **Do not publish enterprise pricing without willingness-to-pay data.**
- **Do not overclaim security.** Publish a reproducible demo and a threat model, not adjectives.

---

## Disagreements & open questions

### Where the evidence conflicts

| Disagreement | Positions | How this report resolves it |
|---|---|---|
| **Is PH worth it for dev tools?** | Believers cite Dub's 663 signups and #1 launches; skeptics cite Plausible's 1.38%, Wasp's #7, and HN outperforming PH for installs. | Both are true at different funnel stages. PH is worth it for awareness, SEO, social proof, and a coordination moment; it is not a dependable signup engine. Budget accordingly. |
| **Best day of week** | PH says no golden day and weekends get 15% more Visit clicks; Flexprice says Tue–Thu have the most traffic but the most competition, and Fri–Sat are quietest with no newsletter. | Goal-dependent. For maximum overall rank, Tuesday–Thursday; for newsletter carryover and lower competition, Sunday. Decide two weeks out based on the leaderboard. |
| **What is a Top 10 score?** | Old rule: ~1k upvotes. Late-2026 medians: overall #1 539, #6–10 156; dev-tool #1 314–499. Live leaderboards show 90–150 on a quiet Sunday. | Use ranges, not thresholds. Target 150–250 for overall Top 10; 300+ for Top 3. PH ranking is engagement-weighted, not pure score. |
| **Does "Top 10" mean overall or dev-tool category?** | Dev-tool category is ~10 launches/day, so a "top 10 dev tool" is close to "every dev tool." | Target **overall daily rank**. The dev-tool badge is not the achievement. |
| **Is the security wedge real?** | The strongest open evidence (#83311) is local worktree contamination; #48474/#81115 are closed and do not prove a forge bypass. Counter-evidence says harness config fixes it in 30 seconds. | The wedge is real but narrower: per-agent identity, scoped leases, and forge-level policy solve a class of failures that harness configs do not solve consistently across agents. Prove it with a reproducible demo; do not claim GitHub is broken. |
| **Is self-hosted demand durable?** | Compliance/air-gap buyers pay; everyone else sees ops burden. Forgejo's own guidance says stay hosted unless you need control. | Durable for the ICP, niche for the mainstream. Do not build a mass-market self-hosted business; build a high-value sovereign one. |
| **Repo-anchored vs agent-first IA** | GitHub re-contained agents into a repo tab; Cursor/Codex bet on an agent-first workspace; Cursor 3 shows the review regression. | Hybrid: global rail for fleet triage, repo Agents tab for context, run detail with diff as truth. Test with users before committing. |
| **Dark vs light** | NN/g finds light better for reading; PH dark screenshots pop; best tools ship both. | Dark-first app, light docs, light/warm-neutral PH gallery surround with dark UI inside. |
| **TanStack Router vs React Router** | TanStack has better typed search params; React Router has 3× the downloads and a larger ecosystem. | TanStack for this URL-heavy app, unless the team values ecosystem/familiarity more. Re-measure type performance on the real route tree. |
| **Pierre vs CodeMirror vs Monaco** | Pierre is fast and young; CodeMirror is proven and editable; Monaco is heavy. | Pierre for large read-only PRs, CodeMirror 6 for editable conflicts; prototype on real diffs. |
| **RAC vs Base UI** | RAC has broader components and accessibility; Base UI is shadcn's new default and lighter in some patterns. | Keep RAC; switch only for a concrete capability gap. |
| **GitLab Ultimate price** | Prior research said $24/seat; the verifier found it is custom-priced, with 24 credits included. | Use "custom" and "24 credits" — never quote $24 as the seat price. |
| **LaunchPact data** | It is the most granular public dataset, but LaunchPact sells launch support and the sample is self-selected. | Use it for ranges and patterns, not as a guarantee. Cross-check with first-party founder numbers. |

### Open questions that need primary research

1. **Will PH feature a self-hosted OSS forge at v0.6-beta?** The featured gate is editorial and the category is crowded. Submit early; have a fallback plan if unfeatured.
2. **What is the real PH-to-install/star conversion for a curl-install OSS product with no signup wall?** No public source measures this funnel.
3. **What will air-gapped/regulated buyers pay per seat per year** for agent-scoped Git, versus extending GitLab or Gitea with proxies and harness rules? Interview 10–15.
4. **Will GitHub bring the coding agent to GHES-only repos, and at what price?** This is the biggest competitive unknown.
5. **Will Gitea/Forgejo ship native agent identity for free?** If yes, the "free self-hosted" wedge narrows.
6. **Does Entire ship self-hosted Git nodes or open-source the forge?** If yes, the direct competition intensifies.
7. **Can exist.io obtain SOC 2/FedRAMP fast enough for regulated buyers?** If not, the ICP is limited to buyers who accept a security questionnaire and a roadmap.
8. **What is TanStack Router's actual type-check/IntelliSense cost at this route count?** Re-measure after the 2024 fixes.
9. **Does Pierre handle forge-scale PRs** (thousands of files, inline comments, split/unified, word wrap) without regressions? No public long-term evidence.
10. **Does RAC's composition/motion DX cost justify a Base UI migration?** Likely no, but test on the run detail and fleet table.
11. **What are the real JS/INP budgets on low-end hardware?** Measure before setting marketing claims.
12. **Is the Go OpenAPI spec complete enough for generated clients?** If not, keep the typed fetch client and codegen incrementally.
13. **What is the CI cost of Playwright and visual regression at this test volume?** Budget before mandating screenshot tests on every PR.
14. **Does the name conflict extend to UK/EU/WIPO registrations or common-law rights?** Only US and Australian registers were screened; get counsel.
15. **Is `exist.io` for sale or lease?** No public signal; an acquisition approach is possible but likely expensive and still leaves the brand collision.

---

## Full source list

### Product Hunt — official documentation

- Points: https://help.producthunt.com/en/articles/10275873-what-are-points
- Launch guide: https://help.producthunt.com/en/articles/17734397-6-what-to-do-during-and-after-your-launch
- Launch FAQ and myths: https://help.producthunt.com/en/articles/17734646-8-launch-faq-and-myths
- Featuring guidelines: https://help.producthunt.com/en/articles/9883485-product-hunt-featuring-guidelines
- Product of the Day/Week/Month: https://help.producthunt.com/en/articles/11751186-product-of-the-day-week-month
- Before launch: https://www.producthunt.com/launch/before-launch
- Preparing for launch: https://www.producthunt.com/launch/preparing-for-launch
- Why points go down: https://help.producthunt.com/en/articles/4853541-why-did-my-launch-points-go-down
- Promoted products: https://help.producthunt.com/en/articles/1444961-how-do-i-get-my-product-promoted
- Build your product page: https://help.producthunt.com/en/articles/17834537-4-build-your-product-page
- Sponsorship: https://www.producthunt.com/sponsor
- Newsletters: https://www.producthunt.com/newsletters
- Coming Soon retired (28 Aug 2025): https://www.producthunt.com/p/producthunt/changelog-august-28-no-more-coming-soon-and-a-couple-other-shiny-updates
- Daily leaderboard, 8 Nov 2026: https://www.producthunt.com/leaderboard/daily/2026/11/8
- Daily leaderboard, 6 Nov 2026: https://www.producthunt.com/leaderboard/daily/2026/11/6
- Appwrite launch: https://www.producthunt.com/products/appwrite/launches/appwrite-sites
- Kilo Code launch retrospective: https://www.producthunt.com/p/kilocode/reverse-engineering-kilo-s-recent-product-hunt-launch

### Product Hunt — data, teardowns, and launch case studies

- LaunchPact, Dec 2025 dev tools: https://www.launchpact.io/product-hunt-launches/december-2025/developer-tools
- LaunchPact, Sept 2025 dev tools: https://www.launchpact.io/product-hunt-launches/september-2025/developer-tools
- LaunchPact, Sept 2026 dev tools: https://www.launchpact.io/product-hunt-launches/september-2026/developer-tools
- LaunchPact, Oct 2026 dev tools: https://www.launchpact.io/product-hunt-launches/october-2026/developer-tools
- LaunchPact, Dec 2024 dev tools: https://www.launchpact.io/product-hunt-launches/december-2024/developer-tools
- LaunchPact statistics: https://www.launchpact.io/product-hunt-statistics
- LaunchPact 3,069-launch study: https://www.launchpact.io/research/product-hunt-launch-study
- Stackfix launch guide: https://tomdekan.com/articles/product-hunt-launch-guide
- Dub.co launch: https://dub.co/blog/product-hunt
- Daytona launch: https://www.daytona.io/dotfiles/how-we-missed-1-on-product-hunter-but-still-won-big
- Flexprice launch: https://flexprice.io/blog/how-we-ranked-product-of-the-day-on-product-hunt
- Postiz growth: https://gitroom.com/blog/from-3k-stars-to-14k-in-3-months
- Open SaaS/Wasp counter-case: https://docs.opensaas.sh/blog/2025-05-07-you-should-still-launch-your-product-on-ph/
- Watermelon HN vs PH: https://medium.com/@baristaGeek/lessons-launching-a-developer-tool-on-hacker-news-vs-product-hunt-and-other-channels-27be8784338b
- Plausible launch numbers: https://plausible.io/blog/product-hunt-launch
- Tally launch numbers: https://blog.tally.so/product-hunt-launch-results/
- Hackmamba PH guide: https://hackmamba.io/developer-marketing/how-to-launch-on-product-hunt/
- Ronak Ganatra dev-tool PH guide: https://ronakganatra.com/posts/successfully-launch-dev-tools-on-producthunt
- Aikido teardown: https://github.com/fmerian/awesome-product-hunt/blob/main/teardowns/aikido.md
- Appwrite teardown: https://github.com/fmerian/awesome-product-hunt/blob/main/teardowns/appwrite.md
- Awesome Product Hunt discussion: https://github.com/fmerian/awesome-product-hunt/discussions/11
- HN discussion on PH value: https://news.ycombinator.com/item?id=41773355
- Infrasity PH guide for dev tools: https://www.infrasity.com/blog/product-hunt-launch-for-developer-tools
- Marketing Ideas PH critique: https://www.marketingideas.com/p/how-to-successfully-launch-on-product

### Competitive and pricing landscape

- GitHub pricing: https://github.com/pricing
- GitHub Copilot billing: https://docs.github.com/en/copilot/concepts/billing/organizations-and-enterprises
- GitHub Copilot access changes: https://docs.github.com/en/copilot/get-started/whats-changing-with-github-enterprise-and-copilot
- GitHub Copilot plans: https://github.com/features/copilot/plans
- GitHub Copilot on GHES: https://docs.github.com/en/enterprise-server@3.18/copilot/copilot-on-ghes/about-copilot-on-ghes
- GitHub Agent HQ: https://github.blog/news-insights/company-news/welcome-home-agents/
- GitHub repository Agents tab: https://github.blog/changelog/2026-01-26-introducing-the-agents-tab-in-your-repository/
- GitHub mission control: https://github.blog/changelog/2025-10-28-a-mission-control-to-assign-steer-and-track-copilot-coding-agent-tasks/
- GitHub enterprise AI controls: https://github.blog/changelog/2025-10-28-enterprise-ai-controls-the-agent-control-plane-are-in-public-preview/
- GitHub agentic audit events: https://docs.github.com/en/copilot/reference/agentic-audit-log-events
- GitHub MCP server configuration: https://github.com/github/github-mcp-server/blob/main/docs/server-configuration.md
- GitHub MCP server issue #2233: https://github.com/github/github-mcp-server/issues/2233
- GitHub MCP server issue #1683: https://github.com/github/github-mcp-server/issues/1683
- GitLab pricing: https://about.gitlab.com/pricing/
- GitLab credits for self-managed: https://support.gitlab.com/hc/en-us/articles/27614164750364-How-GitLab-Credits-Work-for-Self-Managed
- GitLab credits announcement: https://about.gitlab.com/blog/introducing-gitlab-credits
- Bitbucket pricing: https://www.atlassian.com/software/bitbucket/pricing
- Bitbucket lifecycle note: https://www.software-lifecycle.com/sites/bitbucket
- Rovo Dev pricing: https://www.atlassian.com/software/rovo-dev/pricing
- Rovo Dev licensing: https://www.atlassian.com/licensing/rovo-dev
- Azure DevOps pricing: https://azure.microsoft.com/en-us/pricing/details/devops/azure-devops-services/
- SpaceX/Anysphere SEC filing: https://www.sec.gov/Archives/edgar/data/1181412/000162828026056945/spcx-20260814.htm
- Cursor joining SpaceX: https://cursor.com/blog/joining-spacex
- Cursor pricing: https://cursor.com/pricing
- Cursor models and pricing: https://cursor.com/docs/models-and-pricing
- Cursor Origin docs: https://cursor.com/docs/origin
- Cursor Origin changelog: https://cursor.com/changelog/origin-code-hosting
- Reuters SpaceX/Anysphere (401 at check): https://www.reuters.com/legal/transactional/spacex-buy-anysphere-60-billion-2026-06-16/
- Anthropic pricing: https://www.anthropic.com/pricing
- ChatGPT Business overview: https://help.openai.com/en/articles/8792828-chatgpt-business-overview
- Devin pricing: https://devin.ai/pricing
- Kiro pricing: https://kiro.dev/pricing/
- Factory pricing: https://factory.com/pricing
- Augment pricing: https://www.augmentcode.com/pricing
- CodeRabbit pricing: https://www.coderabbit.ai/pricing
- Greptile pricing: https://www.greptile.com/pricing
- Graphite pricing: https://graphite.dev/pricing
- Gitea: https://gitea.com/
- Gitea MCP releases: https://gitea.com/gitea/gitea-mcp/releases
- Gitea MCP v1.8.0 README: https://gitea.com/gitea/gitea-mcp/raw/tag/v1.8.0/README.md
- Forgejo MCP guide: https://fjord.sh/guides/forgejo-gitea-mcp
- SourceHut pricing: https://sourcehut.org/pricing/
- RhodeCode Agentic AI: https://rhodecode.com/agentic/
- RhodeCode roadmap: https://rhodecode.com/roadmap/
- RhodeCode 5.15.0 release notes: https://docs.rhodecode.com/5.x/rce/release-notes/release-notes-5.15.0.html
- Entire launch: https://entire.com/news/entire-launches-an-independent-git-forge-for-agent-human-collaboration
- Entire seed round: https://entire.com/news/former-github-ceo-thomas-dohmke-raises-60-million-seed-round
- Entire pricing: https://entire.com/pricing
- Stratum: https://usestratum.dev/
- Diffinery: https://diffinery.com/agents
- Citadel: https://we.rethunk.tech/citadel
- Jules changelog (wind-down): https://jules.google/docs/changelog/2026-10-29
- Codeberg: https://codeberg.org/

### Demand, security, and community sentiment

- Claude Code issue #48474: https://github.com/anthropics/claude-code/issues/48474
- Claude Code issue #81115: https://github.com/anthropics/claude-code/issues/81115
- Claude Code issue #83311: https://github.com/anthropics/claude-code/issues/83311
- HN worktree isolation thread: https://news.ycombinator.com/item?id=49110389
- HN GitLost prompt-injection thread: https://news.ycombinator.com/item?id=48829383
- HN GitHub lock-in thread: https://news.ycombinator.com/item?id=49483788
- Godot 2026 contribution policy: https://godotengine.org/article/contribution-policy-2026/
- r/cicd CI cost thread: https://www.reddit.com/r/cicd/comments/1wurwho/github_actions_burned_through_my_ci_budget/
- r/ExperiencedDevs agentic engineering thread: https://www.reddit.com/r/ExperiencedDevs/comments/1tw5622/for_folks_heavily_using_a_agentic_engineering/
- Forgejo "what is Forgejo": https://www.forgejo.ch/what-is-forgejo
- Truvo self-hosted GitLab/SOC 2: https://truvocyber.com/blog/soc2-sdlc-self-hosted-gitlab
- IBM self-hosted Bob: https://uk.newsroom.ibm.com/IBM-Introduces-Self-Hosted-Deployment-for-IBM-Bob-to-Help-Enterprises-Advance-AI-Sovereignty-and-Governance
- Helix enterprise coding agents: https://www.helix.ml/use-cases/enterprise-coding-agents
- GitGuardian secrets sprawl 2026: https://blog.gitguardian.com/the-state-of-secrets-sprawl-2026
- Derek Armstrong, Forgejo: https://derekarmstrong.dev/blog/why-forgejo-git-ci-and-you-should-too

### Naming, brand, and legal

- Exist values: https://exist.io/about/values/
- Exist App Store: https://apps.apple.com/us/app/exist/id986201088
- exist.io RDAP: https://rdap.identitydigital.services/rdap/domain/exist.io
- exist.io site (forge): https://ravaniroshan.github.io/exist.io/
- GitHub repo API: https://api.github.com/repos/RavaniRoshan/exist.io
- USPTO search: https://tmsearch.uspto.gov/search/search-results?q=EXIST
- IP Australia mark: https://search.ipaustralia.gov.au/trademarks/search/view/773543
- Trama trademark conflicts: https://www.tramatm.com/blog/category/legal/is-your-business-name-already-taken-how-to-avoid-trademark-conflicts-before
- exist.dev RDAP: https://rdap.org/domain/exist.dev
- OpenTofu rename: https://github.com/opentofu/opentofu/issues/296
- Forgejo: https://forgejo.org
- agentforge.com RDAP: https://rdap.org/domain/agentforge.com
- HostAccent TLD comparison: https://www.hostaccent.com/blog/com-vs-net-vs-io
- get.dev: https://get.dev/
- Coder brand: https://coder.com/brand
- Framed Shot PH gallery sizes: https://framed-shot.com/guides/product-hunt-gallery-screenshots-sizes

### Agents-first information architecture

- Linear Agent Interaction Guidelines: https://linear.app/developers/aig
- Cursor agents window: https://cursor.com/docs/agent/agents-window
- Cursor 3 forum feedback: https://forum.cursor.com/t/cursor-3-agents-window/156509
- Codex Cloud: https://developers.openai.com/codex/cloud
- Jules docs: https://jules.google/docs
- Claude Code on the web: https://code.claude.com/docs/en/claude-code-on-the-web
- Claude Code permission modes: https://code.claude.com/docs/en/permission-modes
- Devin release notes: https://docs.devin.ai/release-notes/overview
- Temporal Web UI: https://docs.temporal.io/web-ui
- LangSmith traces: https://docs.langchain.com/langsmith/view-traces
- Vercel AI Gateway observability: https://vercel.com/docs/ai-gateway/observability-and-spend/observability
- Setproduct agent UI patterns: https://www.setproduct.com/blog/ai-agent-ui-design-patterns
- NN/g AI agents as users: https://www.nngroup.com/articles/ai-agents-as-users/

### Visual design

- LogRocket Linear design critique: https://blog.logrocket.com/ux-design/linear-design
- Shadcn Linear design: https://www.shadcn.io/design/linear
- Vercel Geist colors: https://vercel.com/geist/colors
- Supabase design: https://supabase.com/blog/how-design-works-at-supabase
- Sentry redesign: https://blog.sentry.io/sentry-has-a-bold-new-look
- PostHog brand: https://posthog.com/handbook/brand/foundations
- Refero style reference: https://styles.refero.design/style/0d914ef0-fa84-4c60-a9aa-cef0b5eb6e5d
- Evil Martians OKLCH: https://evilmartians.com/chronicles/oklch-in-css-why-quit-rgb-hsl
- GitHub Primer inclusive color: https://github.blog/engineering/user-experience/unlocking-inclusive-design-how-primers-color-system-is-making-github-com-more-inclusive
- Roselli WCAG 3 contrast: http://adrianroselli.com/2026/04/wcag3-contrast-as-of-april-2026.html
- Smashing colorblindness: https://www.smashingmagazine.com/2024/02/designing-for-colorblindness
- NN/g dark mode: https://www.nngroup.com/articles/dark-mode
- Fuselab dashboard density: https://fuselabcreative.com/dashboard-design-system
- NN/g animation duration: https://www.nngroup.com/articles/animation-duration
- Setproduct glassmorphism: https://www.setproduct.com/blog/liquid-glass-vs-glassmorphism
- Stauffer 2026 color trends: https://www.stauffer.com/news/blog/why-2026-color-trends-are-signaling-a-craving-for-reality
- NN/g trustworthy design: https://www.nngroup.com/articles/trustworthy-design

### Frontend architecture

- TanStack Router vs React Router: https://vercel.com/i/tanstack-router-vs-react-router
- TanStack Router production notes: https://swizec.com/blog/tips-from-8-months-of-tan-stack-router-in-production/
- TanStack Router issue #1091: https://github.com/TanStack/router/issues/1091
- TanStack Router data loading: https://tanstack.com/router/latest/docs/guide/data-loading
- React Router downloads: https://api.npmjs.org/downloads/point/last-week/react-router
- TanStack Router downloads: https://api.npmjs.org/downloads/point/last-week/@tanstack/react-router
- TanStack Query downloads: https://api.npmjs.org/downloads/point/last-week/@tanstack/react-query
- TanStack Query optimistic update discussion: https://github.com/TanStack/query/discussions/10712
- tRPC critique: https://encore.dev/articles/trpc
- Hey API: https://heyapi.dev/
- SSE vs WebSocket for LLM apps: https://tianpan.co/blog/2026/04/19/sse-websockets-grpc-streaming-llm-applications
- fetch-event-source: https://github.com/Azure/fetch-event-source
- WebSocket AI streaming: https://websocket.org/guides/use-cases/ai-streaming/
- Pierre diff rendering: https://pierre.computer/writing/on-rendering-diffs
- Pierre repo: https://github.com/pierrecomputer/pierre
- Sourcegraph Monaco→CodeMirror: https://sourcegraph.com/blog/migrating-monaco-codemirror
- React diff library benchmark: https://dev.to/jiahangzhang/our-release-tool-crashed-on-a-large-pr-heres-the-react-diff-library-i-ended-up-writing-6i5
- Base UI vs React Aria: https://www.untitledui.com/blog/base-ui-vs-react-aria
- shadcn Base UI default: https://ui.shadcn.com/docs/changelog/2026-07-base-ui-default
- RAC DX issue: https://github.com/shadcn-ui/ui/issues/1622
- TanStack Virtual vs react-virtuoso: https://www.pkgpulse.com/guides/tanstack-virtual-vs-react-window-vs-react-virtuoso-2026
- RAC command palette example: https://reactspectrum.blob.core.windows.net/reactspectrum/78a69daa88c1e2bd0cba1f5bb7c69024596d05bf/docs/react-aria/examples/command-palette.html
- cmdk: https://github.com/dip/cmdk
- uPlot: https://github.com/leeoniya/uPlot
- React chart survey: https://www.usedatabrain.com/blog/react-chart-libraries
- Zustand downloads: https://api.npmjs.org/downloads/point/last-week/zustand
- Zustand repo: https://github.com/pmndrs/zustand
- Vitest visual regression: https://main.vitest.dev/guide/browser/visual-regression-testing
- Storybook visual regression discussion: https://github.com/storybookjs/storybook/discussions/32930
- Vite 8: https://vite.dev/blog/announcing-vite8
- React Compiler 1.0: https://react.dev/blog/2025/10/07/react-compiler-1
- Strangler fig: https://learn.microsoft.com/en-us/azure/architecture/patterns/strangler-fig

---

## Appendices

### Appendix A — Corrected Product Hunt launch math

**All-topic medians by rank band** (11,158 settled launches, Sep 2024–7 Nov 2026):

| Rank band | Median score |
|---|---|
| #1 | **539** |
| #2–3 | **363** |
| #4–5 | **247** |
| #6–10 | **156** |

**Developer-tool months:**

| Month | Launches | Median upvotes / comments | Day-#1 median | Top launch | % under 100 |
|---|---|---|---|---|---|
| Dec 2024 | 52 | — | **636** | — | 3.8% |
| Sept 2025 | 98 | **168 / 15** | **499** | **862** (Aikido) | 15.3% |
| Dec 2025 | 86 | **125 / 11** | **442** | **570** | 27.9% |
| Sept 2026 | — | **95** | **439** | **622** | — |
| Oct 2026 | ~322 | **86** | **314** | **436** | **60.9%** |

**Live leaderboard snapshots (not frozen finals):**

| Day | #1 | #2 | #3 | #10 | Dev-tool tags in top 10 |
|---|---|---|---|---|---|
| Sun 8 Nov 2026, 09:00 PST | **150** | 141 | 128 | 85 | **6 / 10** |
| Fri 6 Nov 2026 | **379** | 268 | 170 | 89 | **2 / 10** |

**Caveats:** PH says score is "the current active score as of now," while rank reflects end-of-period ordering; scores drift during and after the day. LaunchPact monthly figures are last-sync counts, not documented end-of-day medians, and its day-#1 median is conditional on dev-tool launches that finished #1. Use ranges.

**Planning targets for exist.io:**

| Goal | Target points | What it requires |
|---|---|---|
| Overall Top 10 | **150–250** | 1,000–2,000 engaged developers; coordinated PH + HN + Reddit + X; working demo |
| Overall Top 3 | **300–350+** | 2,000–5,000 engaged developers; 5k+ stars; design-partner proof; strong maker comment |
| #1 | **400–550** | Large owned audience, repeat launch experience, or a major news hook |

---

### Appendix B — Competitive pricing (corrected)

| Vendor | Current price (Nov 2026) | Notes |
|---|---|---|
| GitHub Team | **$4**/user/mo | Repos, Actions, Packages |
| GitHub Enterprise Cloud | **$21**/user/mo | Copilot seat **$0** for GHE Cloud licensed users, but **metered usage billed separately**, no included AI credits |
| Copilot Business / Enterprise | **$19 / $39** per seat/mo | 1,900 / 3,900 pooled AI credits; new model staged from 28 Oct 2026 |
| GitLab Premium | **$29**/user/mo annual | **12 credits**/user/mo included; credits $1 each; promotional |
| GitLab Ultimate | **Custom** | **24 credits**/user/mo included; not a $24 seat price |
| GitLab air-gapped Duo Agent Platform | **From $299**/user/mo | Offline-licensed self-managed SKU |
| Bitbucket Standard / Premium | **$3.65 / $8** per user/mo | Premium rose ~10% on 13 Oct 2026 |
| Azure DevOps Basic / Test Plans | **$6 / $52** per user/mo | Five free Basic users |
| Cursor Plus / Premium / Super / Ultra | **$8 / $30 / $100 / $200** per month | Teams **$30 / $100 / $200** per seat; Origin included from Premium |
| Claude Pro / Max / Ultra | **$17–20 / $100 / $200** per month | Usage limits apply |
| Codex Business | **$25/$20** standard; **$125/$100** premium per seat | OpenAI estimates ~$100–200/dev/mo in practice |
| Devin Pro / Max | **$20 / $200** per month | Teams **$80/mo + $40/seat**; Enterprise VPC |
| Kiro | **$20 / $40 / $100 / $200** per month | Teams $20/seat + 500 credits; $0.04 overage |
| Factory | **$20 / $100 / $200** per month | Teams $60 + $40/seat; on-prem Enterprise |
| Augment | **$20 / $100** flat up to 50 seats | Usage at list +40% |
| CodeRabbit | **$30 / $60 / $90** monthly | $24/$48/$72 annual; self-host Enterprise only |
| Graphite | **$20 / $40** annual | GHES support |
| Qodo | **$0.015**/credit | On-prem Enterprise |
| Sourcery | **$15 / $30** monthly | $12/$24 annual; self-host Enterprise |
| Gitea | Free; Enterprise **$9.50 promo / $19 list** per user/mo | SAML, audit logs, SLA |
| Forgejo | Free | Codeberg e.V. |
| GitLab CE | Free | Duo requires paid tiers |
| SourceHut | **€4 / €8 / €12** per month | "No AI features whatsoever" |
| RhodeCode Enterprise | **$75**/user/year; Cloud from **$80**/mo | Agent accounts/spend caps/provenance **roadmap, not shipped** |
| Entire | Hosting free; agents PAYG at model rates | Self-hosted Git nodes "upcoming" |
| Cursor Origin | Included from Premium | Early beta; no separate price |

---

### Appendix C — Product Hunt listing checklist

**Page**

- [ ] Name ≤40 characters; tagline ≤60; description ≤500.
- [ ] Logo 240×240, JPG/PNG/GIF, ≤2MB.
- [ ] Gallery ≥3 images (recommend 5), **1270×760**, PNG, 10–20% padding, browser frame for web UI.
- [ ] First image answers "what does this do?" in under 2 seconds.
- [ ] Video 60–90s, YouTube, burned-in captions, hook in first 3–5s, real UI, end CTA.
- [ ] Up to 3 tags; primary tag "Developer Tools".
- [ ] Pricing: Free / Paid / Free options; open-source badge; repo link.
- [ ] Demo link works on mobile and desktop; no signup wall for the first run.
- [ ] Maker first comment ≤5,000 characters: who, why, what, who for, differentiator, one question, offer.
- [ ] Co-makers added before launch; personal account can post.

**Launch day**

- [ ] Live at **12:01 AM PT**; scheduled ≤30 days ahead.
- [ ] First-four-hours war room; reply to every comment within 15 minutes.
- [ ] X/LinkedIn threads; Show HN at 6–9 AM PT; r/selfhosted and r/devops posts; newsletters; Discord/Slack.
- [ ] No paid hunter; no vote buying; no "upvote for upvote" pacts.
- [ ] Status page and demo monitoring; rollback plan.
- [ ] Thank-you post and retrospective within 48 hours.

**Featured-gate pre-check**

- [ ] Real product, not a waitlist, directory, template, boilerplate, course, or service.
- [ ] High-craft screenshots and a working demo.
- [ ] Clear novelty: self-hosted, air-gapped, auditable agent containment.
- [ ] No vaporware claims; the demo does what the listing says.

---

### Appendix D — IA archetypes

| Archetype | Primary object | Best for | Main risk | Recommendation |
|---|---|---|---|---|
| **A. Repo-anchored agent tab** | Session inside a repo | Developers working in one repo; Git-native review | Weak fleet visibility | **Use for repo context** |
| **B. Agent-first multi-repo workspace** | Thread/session | Platform teams running many agents across repos | Review regression; repo anchor lost | **Use for fleet view only** |
| **C. Delegation inbox / issue-native** | Work item | Triage, intake, plan approval | Weak deep review | **Use for Inbox** |
| **D. Run/control-plane console** | Run | Audit, debug, ops | Jargon, dashboard sprawl | **Use for ops/audit** |
| **E. Governance control plane** | Policy/audit | Admin, compliance, security | Separated from daily work | **Use for admin** |

**Recommended composition:** global rail (Inbox, Runs, Reviews, Policies, Repos) + repo Agents tab + run detail (Overview, Changes, Checks, Trace, Audit) + separate admin governance.

---

### Appendix E — Frontend migration sequence

| Phase | Work | Exit criteria |
|---|---|---|
| **0. Baseline** | Measure current JS, LCP, INP, test count, route inventory | Budgets and baseline recorded |
| **1. Shell** | New app shell, rail, top bar, command palette, theme | Shell usable; old routes mount inside |
| **2. Routing** | TanStack Router, typed search params, route tree | All existing routes mapped; deep links work |
| **3. Data** | TanStack Query + OpenAPI codegen; `ensureQueryData` in loaders | No duplicate caches; optimistic updates tested |
| **4. Read-only routes** | Repo overview, commits, issues, releases, profile | Parity + visual regression pass |
| **5. Run detail** | SSE stream, Pierre/CodeMirror spike, diff review, audit | 10k/50k-line diff budgets met |
| **6. Fleet/Inbox** | Virtualized table, saved views, bulk actions | 50+ runs usable at 60fps |
| **7. Review/merge** | Queue, inline comments, policy gate, merge | End-to-end agent run → merge |
| **8. Admin** | Policies, leases, audit export, SSO/SCIM | Security review sign-off |
| **9. Cutover** | Remove old shell, keep rollback branch | No regressions; tests green |

**Non-negotiables:** no mock data in production; authentic empty states; WCAG 2.2 AA; reduced motion; keyboard-first; dark/light; route-level JS budgets; visual regression in pinned CI.

---

### Appendix F — Verification log and limitations

**Blockers found by the verifier and resolved**

1. **Name is a stop-ship.** Confirmed: `exist.io` is a 13-year-old life-tracking product; the domain is owned by Hello Code Pty Ltd; the forge has 0 stars and does not own the name. The IP Australia class 25 footwear mark is **not** a software block; the practical block is domain + brand + app-store + SEO. Recommendation: rename.
2. **PH math was stale.** The old "1k upvotes = #1" rule is obsolete. Corrected to late-2026 ranges: overall Top 10 **90–160**, overall #1 **~539 median**, dev-tool #1 **300–450**. Live leaderboard snapshots are mid-day/current, not final; use ranges.
3. **"First forge for agents" is false.** Cursor Origin (Aug 2026), Entire (Oct 2026), GitHub Agents tab (Jan 2026), and Bitbucket agentic pipelines all exist. Repositioned to self-hosted/air-gapped/auditable containment.

**Majors corrected**

- **GitLab Ultimate** is custom-priced; **$24** is the included credit allotment, not the seat price. Premium is **$29** with **12 credits**; extra credits **$1**; credits are promotional.
- **TanStack Router issue #1091** described ~20 pages and ~1s per-file hotspots, not "20–30s at ~100 routes."
- **The React diff benchmark** was written by the maintainer of a competing library and did not test Pierre; treat as directional.
- **Claude Code #48474 and #81115 are closed** and do not prove a forge-level branch-protection bypass. The strongest open evidence is **#83311** (worktree contamination).
- **GitHub MCP #2233** is a vendor-authored suggestion, not a GitHub admission; GitHub already ships read-only mode, toolsets, include/exclude, and lockdown mode. Missing: per-tool rate limits/quotas.
- **Plausible** drove **2,399 visitors / 33 trials** across the month (1.38%), not 2,490; launch day was **1,490 visitors / 22 trials**. The 5k–15k "top-3" figure is total site traffic, not PH referrals.
- **LaunchPact** is a vendor source; use it for ranges, not guarantees.
- **Jules** is winding down (read-only 11 Nov 2026; shutdown 31 Dec 2026), so it is not a durable IA exemplar.
- **Design tokens** from Linear/Geist conflict with the existing Patina system ("No other hues"). Use them as moodboards, not palettes.
- **The frontend is not greenfield**: React 19, Vite 8.3, RAC 1.17, Tailwind 4.3, Vitest 5; no router, no TanStack Query. Recommendations were translated to the installed stack.

**Minors corrected**

- Exist is 13 years old, not 12.
- PH day starts **12:01 AM PT**, not 00:00.
- IP Australia's live EXIST mark is class 25 apparel, not software.
- GitHub Enterprise's "$0 Copilot seat" is a licensing nuance with metered usage, not a free Copilot seat.
- Rovo Dev's standalone EOL date is TBA; **3 Dec 2026** is the Code Reviewer metering start, not the EOL.
- Gitea MCP v1.8.0 lists **57 tools**, not "40+"; latest v1.13.0 has 48 after consolidation.
- RhodeCode's agent accounts/spend caps/provenance are **roadmap**, not shipped.

**Remaining limitations**

- No legal advice on the name; only US and Australian registers were screened.
- No primary willingness-to-pay data; pricing is a hypothesis.
- No public data on PH-to-install/star conversion for curl-install OSS.
- No public usage data on whether repo-anchored or agent-first IA improves review quality; the hybrid recommendation is judgment plus vendor evidence.
- Pierre and Base UI are young; prototype before committing.
- LaunchPact's monthly medians do not publish sample sizes; month-to-month swings may be noise.
- PH has no frozen final scores; all live numbers drift.
- Reddit evidence came partly through search/RSS excerpts; exact context and vote weight are partially verified.
- Competitive pricing changes frequently; re-check 30 days before launch.

---

### Appendix G — Research provenance

**Primary streams (8):** PH mechanics; dev-tool launch playbooks; competitive/pricing; demand/security sentiment; naming/brand/PH assets; agents-first IA; visual design; frontend architecture.

**Verification:** one adversarial verifier checked every finding against its cited source, flagged contradictions, stale sources, single-source claims, and vendor marketing. Two resolution agents re-checked the decision-critical pricing, launch math, and security evidence.

**Confidence by area:**

| Area | Confidence | Why |
|---|---|---|
| PH mechanics | **High** | PH's own 2025–2026 documentation |
| PH score ranges | **Medium-high** | LaunchPact + live leaderboards; vendor sample |
| Naming collision | **High** | RDAP, live product, App Store, repo data |
| Competitive pricing | **Medium** | Primary pages fetched; some tiers custom/promotional |
| Demand/security | **Medium** | Multiple independent sources; some closed issues |
| IA patterns | **Medium** | Shipped products documented; little usability data |
| Visual direction | **Medium** | Design systems + PH guidance; conversion claims anecdotal |
| Frontend architecture | **Medium** | Production postmortems + adoption data; some vendor benchmarks |
| Willingness to pay | **Low** | No primary interviews |

**Bottom line:** the launch plan is not decision-grade until the name is resolved, the positioning is corrected, and the security proof is reproducible. The redesign plan is usable now if translated to the existing Patina/RAC/Vite stack and tested with users. The research is strong enough to drive the next 90 days, not strong enough to promise a Top 10 finish.
