// Verified facts for the Ledger landing page. Every statement here traces to
// the claims ledger (CLAIMS.md) — nothing is aspirational:
//   leases + refs/heads/agent/* isolation → ledger #11 (Go tests, hook, e2e)
//   pre-receive enforcement + secret scan → internal/security hook tests
//   local JSONL audit trail → ledger #22 (bounded limits noted there)
//   pure-Go zero-CGO binary, embedded SQLite → ledger #4, #3 (Apache-2.0)
//   ports :8080 (ui/rest/smart-http/mcp) + :2222 (SSH, localhost default)
//     → ledger #16, #18 (config default 127.0.0.1:2222)
//   MCP endpoints /mcp + /mcp/readonly → internal/mcp tests
//   signature verification (OpenSSH + GPG) → ledger #8
//   version v0.6.0-beta is a published prerelease (GitHub release API verified)
// Install URL: raw.githubusercontent.com/RavaniRoshan/refweir — the real repo
// (previous marketing page pointed at a dead `refweir.dev` raw URL).

export const INSTALL =
  'curl -fsSL https://raw.githubusercontent.com/RavaniRoshan/refweir/main/scripts/install.sh | bash';

export const BOOTSTRAP =
  'refweird bootstrap --email admin@corp.refweir.dev --namespace acme --data-dir ./data';

export const RELEASE_URL =
  'https://github.com/RavaniRoshan/refweir/releases/tag/v0.6.0-beta';

export const SOURCE_URL = 'https://github.com/RavaniRoshan/refweir';

export const VERSION_LINE = 'v0.6.0-beta (prerelease) · apache-2.0 · zero cgo · no external database';

export const FACTS: Array<{ title: string; body: string }> = [
  {
    title: 'Leases, not standing credentials',
    body: 'An agent holds a cryptographic lease naming one scope — refs/heads/agent/* by default — with a short TTL. It is revocable at any time, and it expires on its own.',
  },
  {
    title: 'Enforcement at the door',
    body: 'The pre-receive hook checks every push against the lease and the repository policy. Writes outside the leased refs, writes to protected branches, and history rewrites are rejected before they are accepted.',
  },
  {
    title: 'Secrets scanned on entry',
    body: 'Incoming commits are scanned for hardcoded secrets and private keys. A rejected push never becomes a commit on your forge.',
  },
  {
    title: 'An ordinary pull request',
    body: 'Accepted agent work arrives as a normal pull request: checks, review, and the merge decision stay human. The agent cannot merge for you.',
  },
  {
    title: 'A written trail',
    body: 'Lease requests, denied pushes, and merge decisions are appended to a local JSONL audit file — on your disk, not in a hosted dashboard.',
  },
  {
    title: 'One binary, your hardware',
    body: 'Pure Go with zero CGO and an embedded SQLite database — no external services at runtime. Web UI, REST, Git Smart-HTTP, and MCP share one port :8080; the embedded SSH server listens on :2222 by default on localhost.',
  },
  {
    title: 'Agents speak Git and MCP',
    body: 'Agents work over the same transports you already trust — Smart-HTTP and SSH — plus an MCP endpoint at /mcp for tools, with a read-only sandbox at /mcp/readonly.',
  },
];

export const RUN_ROWS: Array<{ action: string; detail: string; outcome: string; denied: boolean }> = [
  { action: 'lease_request', detail: 'scope refs/heads/agent/* · ttl 15m', outcome: 'granted', denied: false },
  { action: 'git push agent/plan-77', detail: '3 commits · signatures ok', outcome: 'accepted', denied: false },
  { action: 'git push main', detail: 'protected default branch', outcome: 'denied', denied: true },
  { action: 'pull_request_create', detail: '#214 agent/plan-77 → main', outcome: 'opened', denied: false },
  { action: 'checks', detail: 'secret scan · policy · review', outcome: 'passing', denied: false },
  { action: 'merge', detail: 'merged by a human, in the policy gate', outcome: 'merged', denied: false },
  { action: 'lease_revoke', detail: 'audit entry appended', outcome: 'revoked', denied: false },
];

export const PROOF_CAPTION =
  'Illustrative run record — command names and outcomes follow refweird v0.6.0-beta behavior as exercised by tests in the repository. This is not a live connection.';
