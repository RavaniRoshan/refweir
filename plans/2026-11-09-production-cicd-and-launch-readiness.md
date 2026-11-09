# Production CI/CD & Launch Readiness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish an authentic, professional multi-stage CI/CD pipeline (Development vs Staging/Test vs Production) across the entire Refweir product ecosystem, harden container/Helm deployment assets, and verify 100% launch readiness.

**Architecture:** 
1. Three-tier CI/CD workflow architecture: Fast Developer PR Check (`ci-dev.yml`), Deep Staging/Security/Integration Gate (`ci-test-staging.yml`), and Production Multi-Arch Release & Container Publishing (`release-prod.yml`).
2. Dual-transport containerization: Hardened multi-stage Dockerfile and Docker Compose exposing both Smart-HTTP (`:8080`) and embedded pure-Go SSH (`:2222`), non-root security context.
3. UI Environment Differentiation: Tasteful `[DEV]` badge and demo controls enabled strictly in development, with zero-mock invariants and clean production states.
4. Comprehensive E2E Smoke Test: `scripts/smoke_test_lifecycle.sh` verifying end-to-end bootstrap, lease issue, Smart-HTTP/SSH transport, pre-receive hook isolation, and audit trail verification.

**Tech Stack:** GitHub Actions, Go 1.26, Docker Buildx, Helm 3, React 19, TypeScript, Vitest, SQLite WAL, Fumadocs.

**Spec:** `AGENTS.md` (Operational Invariants), `CLAIMS.md` (Claims Ledger), `README.md` (Quickstart), `deploy/helm/refweir-enterprise/`.

## Global Constraints

- Backend must compile with `CGO_ENABLED=0` across all architectures (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`).
- Zero mock data in live production (`isDemoModeActive() === false`).
- Both port `:8080` (HTTP/REST/MCP) and `:2222` (SSH) must be properly configured, exposed, and tested.
- All claims-audit needles in `site/scripts/audit-claims.mjs` must remain intact.
- Every task must end with automated tests passing with 100% success before committing.

## Review Focus

1. **Dual Transport Exposure**: Dockerfile and Docker Compose must expose `:2222` for SSH alongside `:8080` for HTTP; container health checks must verify daemon responsiveness.
2. **Environment Separation**: Development-only features (mock data toggle, debug banners) must never leak into production builds (`isDemoModeActive() === false`).
3. **Multi-Arch Binary Integrity**: Release workflow must build and cryptographically hash binaries for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64`.
4. **Security Scanner Hardening**: Static analysis (`govulncheck`, `gosec`, `gitleaks`) must run in CI with zero unhandled critical vulnerabilities.
5. **Durable Audit Verification**: Pre-receive hook must intercept over-reach (`GH007`) and record append-only JSONL audit logs.

---

## File Structure

```text
.github/workflows/
├── ci-dev.yml                     # Fast PR and feature branch workflow (< 2 min)
├── ci-test-staging.yml            # Deep integration, race detector, security scans & smoke test
├── release-prod.yml               # Multi-arch binary builder, GHCR container publisher, Helm packager
└── pages-deploy.yml               # Static documentation and marketing deployment to GitHub Pages

deploy/
├── Dockerfile                     # Multi-stage production container exposing :8080 and :2222
├── docker-compose.enterprise.yml   # Multi-service stack (Refweir + MinIO + Prometheus)
└── helm/refweir-enterprise/       # Production Kubernetes Helm chart (synced to v0.6.0-beta)
    ├── Chart.yaml
    └── values.yaml

scripts/
└── smoke_test_lifecycle.sh        # Turnkey automated lifecycle smoke test

web/
└── src/components/views/Header.tsx # Environment-aware visual indicator ([DEV] badge in non-prod)
```

---

## Tasks

### Task 1: Harden Container Packaging & Docker Compose for Dual Transports (:8080 & :2222)

**Files:**
- Modify: `Dockerfile`
- Modify: `docker-compose.enterprise.yml`
- Test: `scripts/test_docker_build.sh` (or local verification)

- [ ] **Step 1: Write verification test for container configuration**
  Create `scripts/test_container_contract.sh` verifying that `Dockerfile` exposes ports 8080 and 2222 and sets default `REFWEIR_SSH_ADDR`.
- [ ] **Step 2: Run verification test to confirm failure**
  Run: `bash scripts/test_container_contract.sh`
- [ ] **Step 3: Update `Dockerfile` & `docker-compose.enterprise.yml`**
  - Add `REFWEIR_SSH_ADDR=0.0.0.0:2222` to `Dockerfile` environment.
  - Add `EXPOSE 8080 2222` to `Dockerfile`.
  - Update `docker-compose.enterprise.yml` to map `"2222:2222"`.
  - Update image tag in `docker-compose.enterprise.yml` to `v0.6.0-beta`.
- [ ] **Step 4: Run verification test to confirm pass**
  Run: `bash scripts/test_container_contract.sh`
- [ ] **Step 5: Commit changes**
  Commit: `fix(docker): expose dual transports :8080 and :2222 in Dockerfile and Compose`

---

### Task 2: Create Turnkey End-to-End Lifecycle Smoke Test (`scripts/smoke_test_lifecycle.sh`)

**Files:**
- Create: `scripts/smoke_test_lifecycle.sh`

- [ ] **Step 1: Implement lifecycle smoke test script**
  Test covers:
  1. Bootstrapping daemon in ephemeral temp directory (`REFWEIR_DATA_DIR`).
  2. Healthcheck ping to `:8080/healthz` and Prometheus endpoint `:8080/metrics`.
  3. Generating ephemeral capability lease via REST API (`POST /api/v1/leases`).
  4. Repository creation via API (`POST /api/v1/repos`).
  5. Git clone via Smart-HTTP with lease credentials.
  6. Git push to `refs/heads/agent/smoke-test` (verifies `200 ALLOW`).
  7. Forbidden push to `refs/heads/main` (verifies `GH007 DENY`).
  8. Audit log query to verify both ALLOW and DENY decisions are recorded.
  9. Clean daemon termination.
- [ ] **Step 2: Execute smoke test locally**
  Run: `bash scripts/smoke_test_lifecycle.sh`
- [ ] **Step 3: Verify zero errors and clean exit**
- [ ] **Step 4: Commit changes**
  Commit: `test(e2e): add automated lifecycle smoke test script`

---

### Task 3: Build Professional Multi-Stage GitHub Actions Pipelines

**Files:**
- Create: `.github/workflows/ci-dev.yml` (Fast PR check)
- Create: `.github/workflows/ci-test-staging.yml` (Deep integration & security)
- Update: `.github/workflows/release-prod.yml` (Multi-arch matrix + GHCR push)
- Update: `.github/workflows/pages-deploy.yml` (Pages deploy with claims gate)
- Remove/Consolidate: Legacy `.github/workflows/ci.yml` and `.github/workflows/release.yml`

- [ ] **Step 1: Define `ci-dev.yml` for rapid PR iteration**
  - Triggers on `pull_request` to `main`.
  - Concurrently runs Go formatting/vet, Web UI Vitest unit tests, and Site verification.
  - Target runtime: < 90 seconds.
- [ ] **Step 2: Define `ci-test-staging.yml` for staging/integration verification**
  - Triggers on `push` to `main` and scheduled daily cron.
  - Runs full `go test -race ./...` across all 35 Go packages.
  - Runs static security scanners: `govulncheck`, `gosec`, `gitleaks`.
  - Runs the new `scripts/smoke_test_lifecycle.sh`.
  - Validates Helm chart with `helm lint deploy/helm/refweir-enterprise`.
- [ ] **Step 3: Update `release-prod.yml` for multi-arch builds & GHCR publishing**
  - Multi-arch binary builder: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`.
  - Multi-arch Docker buildx container push to `ghcr.io/ravaniroshan/refweir:${VERSION}`.
  - Checksum generation (`dist/release/SHA256SUMS`).
  - GitHub release draft/publish with extracted changelog notes.
- [ ] **Step 4: Update `pages-deploy.yml` for static export**
  - Verify `npm run verify` passes with claims audit and link check before deployment.
- [ ] **Step 5: Verify workflow files with actionlint or syntax checks**
- [ ] **Step 6: Commit changes**
  Commit: `ci: establish professional dev, staging, release, and pages workflows`

---

### Task 4: Synchronize Kubernetes Helm Chart to v0.6.0-beta

**Files:**
- Modify: `deploy/helm/refweir-enterprise/Chart.yaml`
- Modify: `deploy/helm/refweir-enterprise/values.yaml`
- Modify: `deploy/helm/refweir-enterprise/templates/service.yaml`
- Modify: `deploy/helm/refweir-enterprise/templates/deployment.yaml`

- [ ] **Step 1: Check Helm templates for SSH port exposure**
  Ensure `service.yaml` exposes port 2222 (`name: ssh`) and `deployment.yaml` has container port 2222.
- [ ] **Step 2: Update `Chart.yaml` and `values.yaml`**
  - Sync `appVersion` and `image.tag` to `v0.6.0-beta`.
  - Ensure `REFWEIR_SSH_ADDR` and SSH ports are documented.
- [ ] **Step 3: Verify Helm chart syntax**
  Check templates format and values structure.
- [ ] **Step 4: Commit changes**
  Commit: `fix(helm): sync chart to v0.6.0-beta and expose pure-Go SSH port 2222`

---

### Task 5: Add UI Environment Indicator ([DEV] Badge) in Web UI

**Files:**
- Modify: `web/src/components/views/Header.tsx`
- Modify: `web/src/components/views/__tests__/Header.test.tsx`

- [ ] **Step 1: Write the failing test**
  Add test in `web/src/components/views/__tests__/Header.test.tsx` asserting that when `isDemoModeActive()` or dev environment is active, a clear `[DEV]` badge is rendered.
- [ ] **Step 2: Run test to confirm it fails**
  Run: `cd web && npx vitest run src/components/views/__tests__/Header.test.tsx`
- [ ] **Step 3: Implement environment badge in `Header.tsx`**
  Render a discreet, authentic Primer badge:
  `<span className="mono text-[10px] px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-500 border border-amber-500/30">DEV</span>`
  only when in development / demo mode.
- [ ] **Step 4: Run test to confirm pass**
  Run: `cd web && npx vitest run src/components/views/__tests__/Header.test.tsx`
- [ ] **Step 5: Commit changes**
  Commit: `feat(web): add development environment indicator badge to header`

---

### Task 6: Full-Product Verification Gate & Launch Readiness Audit

- [ ] **Step 1: Run all backend tests (`go test ./...`)**
- [ ] **Step 2: Run frontend test suites & production build (`cd web && npm test && npm run build`)**
- [ ] **Step 3: Run marketing & docs gate (`cd site && npm run verify`)**
- [ ] **Step 4: Run README audit (`python3 audit_readme.py README.md`)**
- [ ] **Step 5: Run E2E smoke test (`bash scripts/smoke_test_lifecycle.sh`)**
- [ ] **Step 6: Push all commits to `origin/main`**
