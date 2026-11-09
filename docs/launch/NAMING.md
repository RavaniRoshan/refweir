# Naming Decision Record

**Status:** Founder decision made — **Refweir**; legal sign-off pending
**Owner:** Founder/PM
**Gate:** G0 — Identity (end of Week 1)
**Created:** 2026-11-09
**Decision date:** 2026-11-09

---

## 1. Why this is stop-ship

The product cannot launch as `exist.io`.

| Fact | Evidence |
|---|---|
| `exist.io` is the exact-match domain of **Exist**, a live 13-year-old life-tracking product from Hello Code Pty Ltd (Melbourne) | `exist.io/about/values/`; App Store id986201088 |
| The domain was created **2013-06-19**, expires **2027-06-19**, and is `clientTransferProhibited` | RDAP |
| The forge repo was created **2026-10-14** and has **0 stars, 0 forks, 0 issues** | GitHub API, 2026-11-08 |
| The Go module path is literally `exist.io` | `go.mod:1` |
| Every user who types the product name as a URL reaches the life tracker, not the forge | `https://exist.io` |

The legal picture is nuanced, but the *practical* block is not: domain ownership, 13 years of SEO, and app-store identity. Renaming is cheapest now; every star, backlink, and press mention raises the cost. See `research/exist-io-ph-launch-frontend-redesign.md` §"The name is a launch blocker".

---

## 2. Decision required

| Option | What it means | Cost | Recommendation |
|---|---|---|---|
| **A. Full rename now** | New product name across repo, module, binaries, env vars, docs, site, socials, PH slug. `exist.io` becomes a redirect only. | 1–2 weeks of mechanical work; one PR series | **Recommended** |
| **B. Distinct launch name, keep `exist.io` internally** | Public brand is new; Go module/binary/env vars stay `exist` for now. | Cheaper now, but the install command, repo URL, and docs leak the old name; guarantees a second rename | Not recommended |
| **C. Launch as `exist.io`** | Compete with a 13-year-old product for the same query. | Brand confusion, lost search, possible trademark conflict | Reject |

**Decision:** **A. Full rename now**, to **Refweir**, selected 2026-11-09 — see §11.

---

## 3. Naming criteria

Weighted, in order:

1. **Ownable** — `.com` or `.dev`, GitHub org, npm/PyPI, X/LinkedIn, Product Hunt slug.
2. **Searchable** — exact-match query returns the product, not a dictionary word, person, or unrelated company.
3. **Sayable and spellable** — 2–3 syllables; a developer hears it once and types it correctly.
4. **On-message** — evokes containment, scoped authority, audit, or the forge; not "AI magic".
5. **Durable** — not tied to a single agent vendor, model, or protocol.
6. **Legal** — no confusingly similar software/SaaS mark in classes 9/42.
7. **Neutral** — works for defense, EU sovereign, and platform-engineering buyers; no cutesy or aggressive connotations.

---

## 4. What was actually checked

Checked on **2026-11-09**:

- **`.com`** — Verisign RDAP (`404` = unregistered).
- **`.dev` / `.io`** — Google DNS-over-HTTPS NS lookup (`NXDOMAIN` = very likely available; confirm at registrar).
- **GitHub** — `https://github.com/<name>` (`404` = org/user free). Confirmed against a known free and known taken name.
- **npm** — `https://registry.npmjs.org/<name>` (`404` = unscoped package free).
- **PyPI** — `https://pypi.org/pypi/<name>/json` (`404` = free).
- **Docker Hub** — no public repositories under the namespace (not authoritative; namespace reservation is separate).
- **Web/trademark** — exact-string web search for direct software collisions. **This is not a USPTO/EUIPO/TMview database search.** Treat every "no collision" as an unverified negative until counsel screens classes 9 and 42.

**Not checked:** Product Hunt slug (Cloudflare blocks automated checks — check manually), X/LinkedIn handles (login-walled), App Store/Play Store, EUIPO/IP Australia databases.

---

## 5. Screening results

**76 names screened** across three batches. The headline: every common dictionary word is already taken on `.com`, `.dev`, `.io`, and GitHub. Only coined compounds and rare words survived.

### Batch 1 — dictionary words (24)

`corral`, `muster`, `drover`, `wardline`, `portcullis`, `redoubt`, `mandate`, `writ`, `bailiff`, `curfew`, `chaperone`, `steward`, `handler`, `fold`, `rein`, `keep`, `notary`, `witness`, `airwall`, `moat`, `rampart`, `bulwark`, `tumbler`, `sallyport`

**Result:** 23 of 24 taken on at least four of five checks. `steward.io`, `airwall.dev`, `tumbler.dev`, and `tumbler.io` were the only free slots, and none had a coherent full stack.

### Batch 2 — compounds and rare words (25)

| Name | .com | .dev | .io | GitHub | npm | Note |
|---|---|---|---|---|---|---|
| refward | taken | free | free | free | free | **Rejected:** `refward.me` is an active London software company (founded 2018) |
| refkeep | taken | free | free | free | free | **Rejected:** existing SourceForge project; phonetically collides with `rekeep.one` |
| herdgate | taken | free | free | free | free | Whimsical; weak for defense/enterprise |
| keephold | taken | free | free | free | free | Reads as a storage product |
| reflock | taken | free | free | taken | free | GitHub org taken |
| wardforge | taken | free | free | taken | free | GitHub org taken |
| gitward | taken | taken | free | taken | free | "Git" prefix narrows the brand |
| keyway | taken | taken | free | taken | taken | Generic |
| sealmark | taken | taken | taken | free | taken | Weak |

The remaining 16 (`bulkhead`, `keyway`, `pinfold`, `hirsel`, `muniment`, `motte`, `postern`, `bastle`, `airlock`, `deadbolt`, `agentary`, `mergesafe`, `handrail`, `slipway`, `drydock`, `gangway`, `holdfast`) were taken on at least three checks.

### Batch 3 — coined compounds and rare containment words (27)

See Appendix A for the full table. The names that cleared **all five** checks:

| Name | .com | .dev | .io | GitHub | npm | PyPI |
|---|---|---|---|---|---|---|
| **refweir** | free | free | free | free | free | free |
| **refhaven** | free | free | free | free | free | free |
| **moorward** | free | free | free | free | free | free |

Near-misses with a full technical stack but no `.com`:

| Name | .com | .dev | .io | GitHub | npm | PyPI |
|---|---|---|---|---|---|---|
| **refseal** | taken (2010, expires 2027-04) | free | free | free | free | free |
| **refnest** | taken | free | free | free | free | free |
| **refhold** | taken | free | free | free | free | free |
| **keepward** | taken | free | free | free | free | free |

---

## 6. Finalists

### 6.1 Refweir — recommended, conditional

**Pronunciation:** REF-weer (`/ˈrɛf.wɪər/`)
**Metaphor:** a weir is a low barrier that controls the flow of a river without damming it. Refweir controls the flow of agent commits into refs without stopping the work.
**Why it wins:** the only finalist with `.com`, `.dev`, `.io`, GitHub, npm, and PyPI all free; no exact-match software collision in search; the metaphor maps directly to leases, ref-prefix isolation, and merge gates.
**Risks:**
- "Weir" is uncommon; users may type `refwier` or `refwire`. Mitigate with defensive domains and a persistent pronunciation line.
- Weir Group IP Limited owns the `.weir` gTLD and WEIR engineering marks. A compound in software is likely distinguishable, but this is a **yellow flag for counsel**.

**Defensive domains (checked free):** `refwier.com`, `refweird.com`, `refweir.co`, `refweir.app`.

### 6.2 Refseal — strongest fallback

**Pronunciation:** REF-seel
**Metaphor:** a seal authenticates and closes a record. Refseal is the seal on an agent's merge: identity, scope, approval, audit.
**Why it works:** easiest to say and spell of the three; directly carries the audit/approval pillar; clean on `.dev`, `.io`, GitHub, npm, PyPI.
**Risk:** `refseal.com` is registered (2010, expires 2027-04) and the name is close to the well-known Seal Software (DocuSign) mark. Counsel should screen "SEAL" in classes 9/42. No `.com` weakens enterprise credibility.

### 6.3 Refhaven — safest full-stack fallback

**Pronunciation:** REF-hay-ven
**Metaphor:** a safe harbor for agent work.
**Why it works:** all five checks free, zero collisions found, friendly and easy to spell.
**Risk:** "haven" is generic and warm; the name reads more like a documentation or backup product than an enforcement layer. It undersells the containment wedge.

### 6.4 Refnest — alternate

**Pronunciation:** REF-nest
**Metaphor:** Git refs are literally nested under `refs/`; a nest is a contained, protected place.
**Why it works:** Git-native, easy to spell, clean technical stack.
**Risk:** no `.com`; may read as a birding or notes app.

### Comparison

| Criterion | Refweir | Refseal | Refhaven | Refnest |
|---|---|---|---|---|
| `.com` | ✅ | ❌ | ✅ | ❌ |
| `.dev` / `.io` | ✅ | ✅ | ✅ | ✅ |
| GitHub / npm / PyPI | ✅ | ✅ | ✅ | ✅ |
| Exact-match search clean | ✅ | ✅ (near-miss: Seal Software) | ✅ | ✅ |
| Spelling friction | Medium | Low | Low | Low |
| Containment metaphor | Excellent | Good (audit) | Medium | Good (Git) |
| Enterprise/defense tone | Strong | Strong | Medium | Medium |
| Trademark yellow flags | WEIR Group | Seal Software | None found | None found |

---

## 7. Recommendation

**Choose Refweir, subject to two conditions:**

1. **Spell-back test passes.** Show the name to 5 developers who use coding agents; ask them to say it and type it. If more than one types something other than `refweir`, fall back to **Refseal**.
2. **Counsel clears WEIR in classes 9/42.** If counsel flags a likelihood-of-confusion risk with Weir Group, fall back to **Refseal** (accept `.dev` as primary) or **Refhaven** (keep `.com`).

If both conditions pass, register the full stack the same day:

- `refweir.com`, `refweir.dev`, `refweir.io`
- `refwier.com`, `refweird.com` (typo defense)
- GitHub org `refweir`, npm `refweir`, PyPI `refweir`
- Product Hunt slug `refweir` (manual check)
- X/LinkedIn handles `refweir` (manual check)

**Working one-liner:** "Refweir is the self-hosted Git forge that gives every AI agent an identity, a scoped lease, and an audit trail — so agent fleets can ship without touching `main`."

**Working tagline:** "Contain every agent. Prove every merge."

---

## 8. Brand test script (30 minutes)

Recruit 5 developers who use Claude Code, Codex, Cursor, or Cline weekly.

1. Say: "The product is called Refweir. Can you spell that?"
2. Say: "What do you think it does?"
3. Show the one-liner. Ask: "What category is this?"
4. Ask: "Would you type `refweir.com` or search for it? What would you search?"
5. Ask: "Does the name feel like a security product, a Git host, or an AI tool?"

**Pass:** 4/5 spell it correctly; 3/5 place it in "Git/developer tooling"; none describe it as a life-tracking or notes app.
**Fail:** 2+ spell it wrong, or the category read is scattered. Move to Refseal.

---

## 9. Legal next steps (counsel, not the founder)

- [ ] Screen **REFWEIR**, **REFSEAL**, and **REFHAVEN** in USPTO classes 9 and 42.
- [ ] Screen EUIPO, UKIPO, and IP Australia.
- [ ] Check for confusingly similar marks: `REF WEIR`, `WEIR`, `REFSEAL`, `SEAL`, `REFHAVEN`, `HAVEN`.
- [ ] Confirm the `.weir` gTLD is irrelevant to a `refweir.com` registration (it is a closed brand TLD, not a naming restriction by itself).
- [ ] Confirm Apache-2.0 project naming and trademark policy: the code license does not grant trademark rights; publish a short trademark policy.
- [ ] Do not file before the decision; filing is ~$250–$350 per class per jurisdiction and should follow the final choice.

---

## 10. Rename surfaces (execution checklist)

Once the name is chosen, Phase 0.2 applies. Full checklist is in `plan.md` §0.2. Summary:

- `go.mod` module path and every import.
- `cmd/exist`, `cmd/existd` → `cmd/<new>`, `cmd/<new>d`.
- Binaries, Docker image, compose service, Helm chart/release.
- `scripts/install.sh`, `Dockerfile`, `docker-compose.enterprise.yml`, `deploy/helm/exist-enterprise`.
- `EXIST_*` env vars, data dir, DB filename, config keys.
- README, AGENTS.md, ROADMAP, CHANGELOG, LICENSE references.
- `site/`, `web/index.html`, logo, favicon, OG image.
- Docs links, badges, screenshots, install commands.
- GitHub repo/org, socials, PH slug, npm/Go module references.
- Redirects from old URLs + `MIGRATION.md`.

**Acceptance:** `go test ./...`, `npm test`, `npm run build`, and `cd site && npm run verify` pass; `grep -ri "exist.io"` returns only historical/redirect references.

---

## 11. Decision record

| Field | Value |
|---|---|
| Selected name | **Refweir** |
| Fallback | **Refseal** (if the spell-back test or WEIR clearance fails) |
| Decided by | Founder/PM |
| Date | 2026-11-09 |
| Legal sign-off | **Pending** — counsel to screen WEIR in classes 9/42 |
| Domains registered | **Pending** — `refweir.com`, `refweir.dev`, `refweir.io`, `refwier.com`, `refweird.com` |
| GitHub org reserved | **Pending** — repo rename to `RavaniRoshan/refweir` is a manual step |
| npm / PyPI reserved | **Pending** — names verified free on 2026-11-09 |
| PH slug checked | **Pending** — manual check |
| Public announcement | **Pending** — target early February 2027 |

**Signature:** ______________________  **Date:** ____________

---

## Appendix A — Batch 3 screening table

Legend: `F` = free, `T` = taken, `?` = inconclusive.

| Name | .com | .dev | .io | GitHub | npm |
|---|---|---|---|---|---|
| sluice | T | T | T | T | T |
| sluicegate | T | F | F | T | F |
| **refweir** | **F** | **F** | **F** | **F** | **F** |
| refyard | T | T | F | F | T |
| **refhaven** | **F** | **F** | **F** | **F** | **F** |
| refvault | T | F | T | T | F |
| **refseal** | **T** | **F** | **F** | **F** | **F** |
| refhold | T | F | F | F | F |
| refkey | T | F | F | T | F |
| refnest | T | F | F | F | F |
| keyholder | T | T | T | T | T |
| keyward | T | T | T | T | F |
| wardian | T | T | T | T | F |
| bailiwick | T | T | T | T | T |
| stanchion | T | T | T | T | T |
| fairlead | T | T | T | T | T |
| windlass | T | T | T | T | T |
| penstock | T | T | T | T | T |
| stockyard | T | T | T | T | F |
| lariat | T | T | T | T | T |
| vallum | T | T | T | T | T |
| claustra | T | T | T | T | T |
| custodex | T | T | T | T | T |
| **moorward** | **F** | **F** | **F** | **F** | **F** |
| keepward | T | F | F | F | F |
| gateward | T | F | F | F | F |
| wardlock | T | F | F | T | F |

---

## Appendix B — Caveats

- Domain availability can change between this check and registration. Register the finalist the day it is chosen.
- `NXDOMAIN` on `.dev`/`.io` is a strong but not authoritative availability signal. `.com` was checked through the authoritative Verisign RDAP.
- GitHub returns `404` for both free names and some suspended accounts; verify visually before announcing.
- Web search is not a trademark clearance search. The report's USPTO search was JavaScript-dependent and its negative was unverified; counsel must confirm.
- `exist.io` is not for sale on any public signal, and an acquisition approach would still leave the brand collision. Do not plan around buying it.
