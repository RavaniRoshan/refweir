# Terraform Provider for refweir.dev

The official Terraform & OpenTofu provider for managing infrastructure, security rulesets, network allowlists, and repositories on self-hosted `refweir.dev` enterprise clusters.

## Features Managed via Infrastructure-as-Code

- **`refweir_namespace`**: Organization & team namespaces.
- **`refweir_repository`**: Bare Git repositories with default branch configurations, merge policies (squash, rebase, merge commits), and branch pruning.
- **`refweir_ruleset`**: Enterprise branch protection policies:
  - Enforce signed commits (GPG & OpenSSH)
  - Require pull request reviews & minimum approvals
  - Prevent branch deletions and force pushes
  - Require status check runs (CI / security scanners)
- **`refweir_ip_allowlist`**: Network perimeter IP CIDRs (VPNs, office egress).
- **`refweir_ssh_ca`**: Enrolled enterprise SSH Certificate Authorities for keyless developer authentication.

## Authentication

Set the provider credentials via arguments or environment variables:

```bash
export REFWEIR_ENDPOINT="https://git.internal.corp"
export REFWEIR_TOKEN="<your-pat-token>"
```

## Running with OpenTofu / Terraform

```bash
cd deploy/terraform-provider-refweir/examples
tofu init
tofu plan
tofu apply
```
