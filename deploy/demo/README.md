# Hosted demo

**Status:** Scaffolding complete; **public URL not deployed yet.**
**Owner:** Backend + Security/Infra
**Gate:** G1 — hosted demo live (blocked on infrastructure + DNS/TLS)
**Plan:** [`plan.md`](../../plan.md) §1.4

This directory runs a public, single-tenant Refweir instance for the launch. It
is a **real** instance: a real Git forge, a real repository, real leases, real
pushes, and a real audit log. There is no mock data and no sandbox demo mode.

The public URL is not live yet. Do not put a demo link on Product Hunt, the site,
or the README until [§8 Verification](#8-verification-before-announcing) passes on
the deployed host.

---

## 1. What the demo shows

A visitor lands on a seeded repository, `demo/agent-lab`, with:

- a short human commit history and a runbook;
- an issue about agent retry storms and an issue about audit export;
- a pull request opened by `claude-code` from `agent/retry-budget`, pushed under a
  real HMAC-signed capability lease;
- a `v0.3.0` release;
- an audit log containing the lease, the accepted push, and the PR;
- a persistent **Live demo** banner.

The banner has two actions:

1. **Try an agent run** — the server issues a fresh scoped lease, clones the demo
   repo over Smart-HTTP with that lease, commits a note on a new `agent/demo-*`
   branch, pushes it, and opens a real pull request. The lease cannot write
   `main`. The action is rate-limited at the app and at the edge.
2. **Merge the PR** — after a run, the banner can merge only that `agent/*`
   pull request into the default branch through `POST /api/v1/demo/merge`. This
   is demo-only, scoped to the demo repository, and rate-limited. It exists so a
   visitor can complete the loop without an account; it is not a general
   unauthenticated merge path.
3. **Self-host Refweir** — links to the quickstart.

The first run needs no signup. The demo exposes no admin credentials and no
write path to `main`.

---

## 2. Architecture

```text
                 TLS (operator LB / proxy)
                           │
                  ┌────────▼────────┐
                  │  nginx :8080    │  edge rate limits, SSE + Git buffering off
                  └────────┬────────┘
                           │
                  ┌────────▼────────┐
                  │ refweird :8080  │  REFWEIR_DEMO_MODE=1
                  │                 │  demo/agent-lab seed
                  │  SQLite + git   │
                  │  audit.log      │
                  └─────────────────┘
                           │
                    demo-data volume
```

- One instance per demo. There is **no tenant isolation inside the OSS binary**;
  isolation is the container and the host. Do not put multiple customers on this
  deployment.
- The `refweird` container is not published directly. Only the nginx edge is.
- TLS is intentionally not in this directory. Terminate it at your load balancer
  or add an upstream proxy. Do not expose plain HTTP on the public internet.
- `REFWEIR_DEMO_SELF_URL=http://refweird:8080` lets the agent-run handler reach its
  own Git transport without going back through the public edge.

---

## 3. Quick start (local)

Prerequisites: Docker with Compose v2, `git`, `curl`, `python3`, and ~3 GB free
disk for the image build.

```bash
cd deploy/demo
./reset.sh
```

`reset.sh` generates `.env`, removes the old volume, bootstraps a fresh admin,
starts the stack, waits for health, and seeds the repository. Open:

```text
http://localhost:8088
```

The banner should appear. Click **Try an agent run**; within a few seconds you
should get a PR number and a branch name. Then check:

```bash
docker compose exec refweird sh -c 'tail -n 20 /var/lib/refweir/audit.log'
```

You should see `agent.lease_issued`, `git.receive_pack` with actor type `agent`,
and `demo.agent_run`.

### Without Docker

Run a binary directly with the demo surface enabled:

```bash
CGO_ENABLED=0 go build -o /tmp/refweird ./cmd/refweird
REFWEIR_TELEMETRY_OPT_IN=0 /tmp/refweird bootstrap \
  --email demo-admin@refweir.dev --namespace demo --data-dir /tmp/refweir-demo
REFWEIR_LISTEN_ADDR=127.0.0.1:8088 \
REFWEIR_PUBLIC_URL=http://127.0.0.1:8088 \
REFWEIR_GIT_URL=127.0.0.1:8088 \
REFWEIR_SSH_ADDR=none \
REFWEIR_DATA_DIR=/tmp/refweir-demo \
REFWEIR_DEMO_MODE=1 \
REFWEIR_DEMO_REPO=demo/agent-lab \
  /tmp/refweird
```

Then, from another shell, seed with the API key printed by bootstrap:

```bash
deploy/demo/seed.sh http://127.0.0.1:8088 <API_KEY>
```

---

## 4. Seeding

`seed.sh` is idempotent at the repository level: it deletes `demo/agent-lab` and
`demo/demo` (the repository created by bootstrap), recreates `demo/agent-lab`, and
rebuilds the history. It requires an admin API key.

What it creates:

| Artifact | How |
|---|---|
| Human commit history | Real `git` commits with three authors |
| Agent branch and push | Real lease issued over REST, then `git push` with the lease as the credential |
| Pull request | `POST /api/v1/repos/demo/agent-lab/pulls` |
| Two issues | `POST /api/v1/repos/demo/agent-lab/issues` |
| Release `v0.3.0` | `POST /api/v1/repos/demo/agent-lab/releases` |
| Audit trail | Written by the daemon and the pre-receive hook |

The seed does **not** call the demo agent-run endpoint. That endpoint is reserved
for visitors, so the seed does not consume their rate budget.

---

## 5. Reset schedule

The demo must reset on a schedule: visitors create branches and PRs, and a
months-old demo rots. The included systemd timer resets every six hours.

```bash
sudo cp deploy/demo/systemd/refweir-demo-reset.* /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now refweir-demo-reset.timer
systemctl list-timers refweir-demo-reset.timer
```

`reset.service` assumes the repo is at `/opt/refweir` and sets
`RESET_SKIP_BUILD=1` so the six-hour reset does not rebuild the image. Rebuild
manually after a code change:

```bash
cd deploy/demo && docker compose build && ./reset.sh
```

Cron alternative:

```cron
0 */6 * * * cd /opt/refweir/deploy/demo && RESET_SKIP_BUILD=1 ./reset.sh >> /var/log/refweir-demo-reset.log 2>&1
```

The reset is destructive by design. It removes the `refweir-demo-data` volume,
including the audit log. Export logs before a reset if you need launch-week
evidence.

---

## 6. Rate limits and abuse controls

| Layer | Control |
|---|---|
| App | One agent run per client per 20 seconds; five per hour; one run in flight |
| App | One merge per client per 20 seconds; five per hour |
| App | Agent runs are capped to a single demo repository and `refs/heads/agent/*` |
| App | Merges are capped to the demo repo, an `agent/*` head, and the default base |
| App | A fresh lease expires after 10 minutes; the lease cannot touch `main`/`master` |
| nginx | `10r/s` general API, `30r/m` for demo runs/merges, connection caps, 100 MB body |
| Container | `no-new-privileges`, all capabilities dropped, 2 GB / 2 CPU / 512 PID limits |

The app limiter is in-memory, so it resets when the process restarts. The nginx
limiter is the durable edge. If you put a CDN in front, preserve
`X-Forwarded-For` and make sure the proxy overwrites it rather than appending an
attacker-controlled value.

WAF/edge rules are out of scope here. Add them at the load balancer for a public
launch.

---

## 7. No mock data, by construction

- `REFWEIR_DEMO_MODE=1` does not switch the UI into `isDemoModeActive()`. The web
  client still reads the live API.
- The banner renders from `GET /api/v1/system/info` (`demoMode: true`), not from a
  local flag.
- The agent run performs a real clone, commit, push, and PR insert. The response
  includes the real lease ID and branch name.
- Empty API results render the normal Primer empty states; they are not replaced
  with fixtures.

If the demo is broken, it must look broken. Do not add a fallback that papers over
a failed API call.

---

## 8. Verification before announcing

Run these on the deployed host, from a phone and a desktop, on a fresh network:

- [ ] `https://<demo-host>/healthz` returns `200`.
- [ ] The banner says **Live demo** and the self-host link works.
- [ ] The seeded PR, issues, and release are visible.
- [ ] **Try an agent run** returns a PR number in under 60 seconds.
- [ ] The PR's head branch is `agent/demo-*`; `main` is unchanged.
- [ ] **Merge the PR** succeeds, records a merge commit, and closes the PR.
- [ ] A second run inside 20 seconds returns `429` with a retry hint.
- [ ] `audit.log` contains `agent.lease_issued`, `git.receive_pack`,
      `demo.agent_run`, and the lease ID.
- [ ] A manual push to `main` under a demo lease is rejected with `GH007`
      (use `examples/containment/run.sh` as the reference).
- [ ] The site looks correct at 390 px and 1440 px wide.
- [ ] No console errors, no failed requests in the network tab.
- [ ] The instance was reset at least once and still seeds correctly.
- [ ] A status page or uptime check is pointed at `/healthz`.

Until this list is complete, the public URL stays out of launch assets.

---

## 9. Launch-day operations

- Keep the reset timer disabled during the launch window if a reset mid-demo would
  be disruptive; reset manually before the window and after it.
- Watch `docker stats`, `nginx` 4xx/5xx, and `/healthz`.
- Watch the app's `429` rate. A spike means the demo is popular or being abused;
  tighten the nginx zone, do not remove the limit.
- Keep a shell open with `docker compose logs -f refweird proxy`.
- If the agent run fails, the banner shows the server error. Do not replace it
  with a canned success.
- Have `examples/containment/run.sh` ready as a fallback terminal demo if the
  hosted instance is down. Say it is a local demo, not the hosted one.

---

## 10. Teardown

```bash
cd deploy/demo
docker compose down -v --remove-orphans
rm -f .env
```

This removes the demo volume and the generated secret. It does not touch any
other Refweir deployment on the host.

---

## 11. Environment reference

| Variable | Default | Purpose |
|---|---|---|
| `REFWEIR_DEMO_MODE` | off | Enables the banner flag and the agent-run route |
| `REFWEIR_DEMO_REPO` | `demo/agent-lab` | The only repository the demo run may use |
| `REFWEIR_DEMO_SELF_URL` | derived from `REFWEIR_LISTEN_ADDR` | Base URL for the runner's own Git transport |
| `REFWEIR_JWT_SECRET` | required (`.env`) | Lease signing key; 32+ bytes |
| `REFWEIR_AUDIT_LOG` | `<data-dir>/audit.log` | Durable JSONL audit file |
| `DEMO_PORT` | `8088` | Host port for the nginx edge |
| `DEMO_PUBLIC_URL` | `http://localhost:8088` | Public base URL shown in clone links |
| `DEMO_GIT_URL` | `localhost:8088` | Git host shown in the UI |
| `RESET_SKIP_BUILD` | unset | `1` reuses the image during reset |

The agent-run route is registered only when `REFWEIR_DEMO_MODE=1`. A normal
self-hosted deployment must leave it unset; the route then does not exist.
