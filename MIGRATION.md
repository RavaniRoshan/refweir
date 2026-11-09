# Migration Guide — `exist.io` → Refweir

This repository was renamed from **exist.io** to **Refweir** before public launch. The old name
collided with a live 13-year-old life-tracking product and its domain. The code, APIs, and data
model are unchanged; only names, paths, and identifiers moved.

- **Product name:** Refweir
- **Canonical site:** <https://refweir.dev>
- **Repository:** <https://github.com/RavaniRoshan/refweir>
- **Docs:** <https://ravaniroshan.github.io/refweir/docs/>
- **License:** Apache-2.0 (unchanged)

## 1. Old → new map

### URLs and identity

| Before | After |
|---|---|
| `https://exist.io` | `https://refweir.dev` |
| `https://ravaniroshan.github.io/exist.io/` | `https://ravaniroshan.github.io/refweir/` |
| `https://github.com/RavaniRoshan/exist.io` | `https://github.com/RavaniRoshan/refweir` |
| Docs base path `/exist.io/docs/` | `/refweir/docs/` |
| `notifications@exist.io` | `notifications@refweir.dev` |
| `security@exist.io` | `security@refweir.dev` |
| `conduct@exist.io` | `conduct@refweir.dev` |
| `maintainers@exist.io` | `maintainers@refweir.dev` |
| `noreply@exist.io` | `noreply@refweir.dev` |

### Code and binaries

| Before | After |
|---|---|
| Go module `exist.io` | `refweir.dev` |
| Import prefix `exist.io/internal/...` | `refweir.dev/internal/...` |
| `cmd/exist` → `bin/exist` | `cmd/refweir` → `bin/refweir` (operator CLI) |
| `cmd/existd` → `bin/existd` | `cmd/refweird` → `bin/refweird` (server daemon) |
| `terraform-provider-exist` | `terraform-provider-refweir` |

Update build commands:

```bash
# Before
go build -o bin/exist ./cmd/exist
go build -o bin/existd ./cmd/existd

# After
go build -o bin/refweir ./cmd/refweir
go build -o bin/refweird ./cmd/refweird
```

Update Go imports:

```go
// Before
import "exist.io/internal/cli"

// After
import "refweir.dev/internal/cli"
```

### Environment variables

Every `EXIST_*` variable became `REFWEIR_*`. The mapping is mechanical:

| Before | After |
|---|---|
| `EXIST_DATA_DIR` | `REFWEIR_DATA_DIR` |
| `EXIST_LISTEN_ADDR` | `REFWEIR_LISTEN_ADDR` |
| `EXIST_PUBLIC_URL` | `REFWEIR_PUBLIC_URL` |
| `EXIST_GIT_URL` | `REFWEIR_GIT_URL` |
| `EXIST_SSH_ADDR` | `REFWEIR_SSH_ADDR` |
| `EXIST_SSH_HOST_KEY` | `REFWEIR_SSH_HOST_KEY` |
| `EXIST_CLUSTER_ROLE` | `REFWEIR_CLUSTER_ROLE` |
| `EXIST_REPLICA_DIR` | `REFWEIR_REPLICA_DIR` |
| `EXIST_NODE_ID` | `REFWEIR_NODE_ID` |
| `EXIST_SMTP_*` | `REFWEIR_SMTP_*` |
| `EXIST_TELEMETRY_OPT_IN` | `REFWEIR_TELEMETRY_OPT_IN` |
| `EXIST_TELEMETRY_ENDPOINT` | `REFWEIR_TELEMETRY_ENDPOINT` |
| `EXIST_VERSION` | `REFWEIR_VERSION` |
| `EXISTD_PATH` | `REFWEIRD_PATH` |

A one-shot shell migration for an existing deployment:

```bash
sed -i 's/^EXIST_/REFWEIR_/' .env
sed -i 's/^EXISTD_PATH=/REFWEIRD_PATH=/' .env
```

### Runtime state

| Before | After | Notes |
|---|---|---|
| Session cookie `exist_session` | `refweir_session` | **Breaking:** existing browser sessions are invalidated and users must sign in again. |
| `localStorage` keys `exist_token`, `exist_sandbox_demo_mode`, `exist_mode_changed`, `exist_onboarding_tour_completed` | `refweir_*` | **Breaking:** users re-enter a token once; other keys reset. |
| Data directory (default `./data`) | `./data` | Unchanged. |
| Database file `meta.db` | `meta.db` | Unchanged. Move the existing file as-is. |
| Docker data path `/var/lib/exist` | `/var/lib/refweir` | Update volumes and `REFWEIR_DATA_DIR`. |
| Docker user/group `exist` | `refweir` | Ownership must be updated (`chown -R 1000:1000`). |

The SQLite schema is unchanged. Do **not** re-run migrations into a new database unless you intend
to start fresh; point `REFWEIR_DATA_DIR` at the existing `data/` directory.

### MCP client configuration

Rename the client-side server key; the endpoint is unchanged.

```json
{
  "mcpServers": {
    "refweir": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer <your-admin-token>"
      }
    }
  }
}
```

### Deploy surfaces

| Before | After |
|---|---|
| Docker image `exist-io/refweird` | `refweir/refweird` |
| Helm chart `deploy/helm/exist-enterprise` | `deploy/helm/refweir-enterprise` |
| Helm release name `exist-cluster` | `refweir-cluster` |
| Terraform provider `exist` | `refweir` |
| Terraform resources `exist_*` | `refweir_*` |
| Release archives `exist_<version>_<os>_<arch>` | `refweir_<version>_<os>_<arch>` |
| Install script `RavaniRoshan/exist.io` | `RavaniRoshan/refweir` |

Terraform users must update both the provider source and every resource prefix, or use `moved`
blocks for a staged migration. Helm users should install the new chart under the new release name
and then delete the old release.

## 2. What did not change

- REST paths under `/api/v1/...`.
- Smart-HTTP and SSH Git protocol behaviour.
- MCP endpoint paths `/mcp` and `/mcp/readonly`.
- Repository, PR, issue, lease, and audit data models.
- SQLite schema and `meta.db` format.
- Apache-2.0 licensing.

## 3. Redirect plan

| Old URL | Treatment |
|---|---|
| `https://exist.io/*` | The domain is owned by a third party and is **not** controlled by this project. It cannot be redirected. |
| `https://ravaniroshan.github.io/exist.io/*` | GitHub Pages path changes when the repository is renamed. Add a static redirect page at the old project path if the host still serves it. |
| `/exist.io/<page>/` doc URLs | `site/scripts/redirect-stubs.mjs` emits stubs for the pre-unification doc layout under the new base path. |
| `github.com/RavaniRoshan/exist.io` | GitHub redirects renamed repositories automatically. Keep the old remote working; update CI and install scripts to the new URL. |

Because `exist.io` is not ours, the only durable fix is to publish the new canonical URLs and keep
the old GitHub repository redirect alive. Do not promise a redirect we cannot control.

## 4. Manual steps still required

These are outside the repository and cannot be completed by editing files:

- [ ] Rename the GitHub repository `RavaniRoshan/exist.io` → `RavaniRoshan/refweir`.
- [ ] Register `refweir.com`, `refweir.dev`, `refweir.io`, plus defensive `refwier.com` and `refweird.com`.
- [ ] Point `refweir.dev` DNS at the site host and set the GitHub Pages custom domain.
- [ ] Reserve the GitHub org `refweir`, npm `refweir`, PyPI `refweir`, and Docker Hub `refweir`.
- [ ] Check the Product Hunt slug `refweir` and X/LinkedIn handles manually.
- [ ] Re-render `videos/exist-agent-git-explainer/renders/video.mp4` and the vertical explainer; the rendered footage still shows the old name.
- [ ] Regenerate `assets/demo.gif` and `assets/demo.mp4` if their on-screen text mentions the old name.
- [ ] Publish the trademark policy after counsel clears the name.

## 5. Rollback

The rename touched identifiers, not schema or protocol. To roll back a deployment, restore the
previous binary and set the old `EXIST_*` environment variables. Session cookies and browser
storage keys will not round-trip automatically; users sign in again either way.
