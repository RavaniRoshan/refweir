export interface DocSection {
  id: string
  title: string
  level?: 2 | 3
  content?: string
}

export interface DocCodeSnippet {
  language: 'bash' | 'curl' | 'json' | 'go' | 'typescript' | 'yaml'
  label: string
  code: string
}

export interface DocCallout {
  type: 'note' | 'tip' | 'important' | 'warning'
  title?: string
  content: string
}

export interface DocParam {
  name: string
  type: string
  required: boolean
  description: string
  default?: string
}

export interface DocArticle {
  id: string
  categoryId: string
  title: string
  lead: string
  readTime: string
  badge?: string
  lastUpdated: string
  sections: DocSection[]
  callouts?: DocCallout[]
  codeSnippets?: DocCodeSnippet[]
  params?: DocParam[]
  bodyHtml?: string
  contentMarkdown?: string
  relatedArticles?: string[]
}

export interface DocCategory {
  id: string
  title: string
  icon: string
  description: string
  articleIds: string[]
}

export const DOCS_CATEGORIES: DocCategory[] = [
  {
    id: 'get-started',
    title: 'Get Started',
    icon: 'Rocket',
    description: 'Learn the fundamentals of Refweir, deployment models, and quick bootstrap.',
    articleIds: ['overview', 'quickstart', 'configuration', 'forks-and-upstream'],
  },
  {
    id: 'agents-mcp',
    title: 'Autonomous Agents & MCP',
    icon: 'Bot',
    description: 'Model Context Protocol endpoints, capability leases, and agent reviews.',
    articleIds: ['agentic-workflows', 'mcp-tools-reference', 'capability-leases', 'background-reviewer'],
  },
  {
    id: 'security-governance',
    title: 'Security & Governance',
    icon: 'ShieldCheck',
    description: 'Cryptographic branch rulesets, secret scanning, and tamper-evident audit logs.',
    articleIds: ['rulesets-policy', 'secret-scanning', 'ssh-ca', 'audit-logs'],
  },
  {
    id: 'git-storage',
    title: 'Git Engine & Storage',
    icon: 'HardDrive',
    description: 'Pure-Go Smart-HTTP and SSH servers, Git LFS CAS, and SQLite WAL replication.',
    articleIds: ['git-http-transport', 'git-ssh-transport', 'git-lfs-cas', 'storage-replication'],
  },
  {
    id: 'api-webhooks',
    title: 'REST API & Webhooks',
    icon: 'Webhook',
    description: 'OpenAPI 3.1 endpoints, HMAC-SHA256 webhooks, and CLI operations.',
    articleIds: ['rest-api-overview', 'webhooks-guide', 'cli-reference'],
  },
  {
    id: 'enterprise-admin',
    title: 'Enterprise Administration',
    icon: 'Building2',
    description: 'SAML 2.0 / OIDC SSO, SCIM 2.0 provisioning, and telemetry monitoring.',
    articleIds: ['sso-integration', 'scim-provisioning', 'metrics-observability'],
  },
]

export const DOCS_ARTICLES: Record<string, DocArticle> = {
  overview: {
    id: 'overview',
    categoryId: 'get-started',
    title: 'Overview of Refweir',
    lead: 'Refweir is an autonomous enterprise Git forge with built-in Model Context Protocol (MCP) tool hosting, cryptographic capability leases, and a pure-Go transport engine.',
    readTime: '3 min read',
    badge: 'Core Architecture',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'what-is-refweir', title: 'What is Refweir?' },
      { id: 'key-capabilities', title: 'Key Capabilities' },
      { id: 'single-binary-model', title: 'Single-Binary Operational Model' },
      { id: 'next-steps', title: 'Next Steps' },
    ],
    callouts: [
      {
        type: 'note',
        title: 'Zero External Dependencies',
        content: 'Refweir is distributed as a single static Go binary (`refweird`) containing the embedded web console, SQLite WAL multi-engine, and SSH daemon. No Redis, Postgres, or external daemons are required.',
      },
    ],
    codeSnippets: [
      {
        language: 'bash',
        label: 'CLI',
        code: `# Check cluster status and version
refweird version
# Output: refweird v0.2.0-enterprise (linux/amd64)`,
      },
      {
        language: 'curl',
        label: 'cURL',
        code: `curl -s http://localhost:8080/healthz | jq .
# {
#   "status": "healthy",
#   "version": "refweird v0.2.0-enterprise",
#   "mcp_tools": 55
# }`,
      },
    ],
    relatedArticles: ['quickstart', 'configuration', 'agentic-workflows'],
  },

  quickstart: {
    id: 'quickstart',
    categoryId: 'get-started',
    title: 'Quickstart: 5-Minute Bootstrap',
    lead: 'Get a production-ready Refweir instance running locally or on your enterprise cluster with a single bootstrap command.',
    readTime: '4 min read',
    badge: 'Getting Started',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'prerequisites', title: 'Prerequisites' },
      { id: 'bootstrap-daemon', title: '1. Bootstrap the Daemon' },
      { id: 'create-first-repo', title: '2. Create Your First Repository' },
      { id: 'configure-ssh-keys', title: '3. Configure SSH Git Authentication' },
      { id: 'connect-agent', title: '4. Connect an Autonomous Agent' },
    ],
    callouts: [
      {
        type: 'tip',
        title: 'Ephemeral Sudo Token',
        content: 'The initial bootstrap command prints an administrative recovery token. Store it securely in your secret manager to configure SAML/OIDC and SCIM.',
      },
      {
        type: 'important',
        title: 'Port Bindings',
        content: 'Refweir binds HTTP/REST/Web UI on port 8080 and pure-Go SSH Git on port 2222 by default.',
      },
    ],
    codeSnippets: [
      {
        language: 'bash',
        label: 'Step 1: Bootstrap',
        code: `# Initialize database, admin credentials, and default rulesets
refweird bootstrap --email admin@refweir.internal --admin-user root

# Start the cluster daemon in foreground or as a systemd service
refweird run --data-dir /var/lib/refweir --http-port 8080 --ssh-port 2222`,
      },
      {
        language: 'bash',
        label: 'Step 2: Git Clone',
        code: `# Clone over pure-Go SSH
git clone ssh://git@localhost:2222/shiva/core-engine.git

# Or clone over Smart-HTTP with Bearer token
git clone http://localhost:8080/shiva/core-engine.git`,
      },
      {
        language: 'json',
        label: 'Step 3: Claude Code / MCP Config',
        code: `{
  "mcpServers": {
    "refweir": {
      "command": "refweir",
      "args": ["mcp", "serve", "--endpoint", "http://localhost:8080/api/v1"]
    }
  }
}`,
      },
    ],
    relatedArticles: ['overview', 'configuration', 'mcp-tools-reference'],
  },

  configuration: {
    id: 'configuration',
    categoryId: 'get-started',
    title: 'Cluster Configuration & Environment',
    lead: 'Comprehensive reference of environment variables, configuration flags, and file paths for Refweir operators.',
    readTime: '5 min read',
    badge: 'Operations',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'environment-variables', title: 'Core Environment Variables' },
      { id: 'storage-layout', title: 'On-Disk Storage Layout' },
      { id: 'high-availability-flags', title: 'High Availability & Litestream Flags' },
    ],
    params: [
      {
        name: 'REFWEIR_DATA_DIR',
        type: 'string',
        required: false,
        default: './data',
        description: 'Root directory for bare Git repositories, SQLite databases, and LFS objects.',
      },
      {
        name: 'REFWEIR_HTTP_ADDR',
        type: 'string',
        required: false,
        default: ':8080',
        description: 'Listen address for Smart-HTTP, REST API, Web UI, and MCP server endpoints.',
      },
      {
        name: 'REFWEIR_SSH_ADDR',
        type: 'string',
        required: false,
        default: ':2222',
        description: 'Listen address for the pure-Go SSH transport server.',
      },
      {
        name: 'REFWEIR_JWT_SECRET',
        type: 'string',
        required: true,
        default: 'auto-generated',
        description: '256-bit cryptographic secret for signing session tokens and capability leases.',
      },
      {
        name: 'REFWEIR_LITESTREAM_URL',
        type: 'string',
        required: false,
        default: 'none',
        description: 'S3-compatible destination URL for continuous sub-second SQLite WAL replication.',
      },
    ],
    codeSnippets: [
      {
        language: 'yaml',
        label: 'systemd / Docker Env',
        code: `REFWEIR_DATA_DIR=/var/lib/refweir
REFWEIR_HTTP_ADDR=:8080
REFWEIR_SSH_ADDR=:2222
REFWEIR_JWT_SECRET=c9f7a82b3d4e5f60718293a4b5c6d7e8f90123456789abcdef0123456789abcd
REFWEIR_LOG_FORMAT=json
REFWEIR_ENABLE_BACKGROUND_REVIEWER=true`,
      },
    ],
    relatedArticles: ['overview', 'quickstart', 'storage-replication'],
  },

  'forks-and-upstream': {
    id: 'forks-and-upstream',
    categoryId: 'get-started',
    title: 'Working with Forks & Upstream Synchronization',
    lead: 'How to fork repositories, preserve parent lineage, and synchronize divergent branches with upstream using fast-forward or 3-way merges.',
    readTime: '4 min read',
    badge: 'Git Forge Lineage',
    lastUpdated: 'November 2026',
    sections: [
      {
        id: 'what-is-a-fork',
        title: 'What is a Fork in Refweir?',
        content:
          'A fork is a personal or organization-owned copy of an upstream parent repository that shares initial Git object history while living in an isolated bare repository on disk. In Refweir, forks maintain bidirectional lineage tracking via parent_repo_id and is_fork metadata flags.',
      },
      {
        id: 'forking-workflow',
        title: 'Forking Repositories with Two-Phase Concurrency Safety',
        content:
          'Forks can be initiated via the Web UI, CLI (refweir repo fork), REST API, or MCP (fork_repository). To eliminate deadlocks during concurrent multi-tenant forking, Refweir enforces deterministic two-phase repository mutex locking in alphabetical order of namespace slugs.',
      },
      {
        id: 'upstream-synchronization',
        title: 'Synchronizing Branches with Upstream',
        content:
          'When the parent repository advances, fork branches can be synchronized without a local working tree checkout. Refweir fetches upstream into refs/remotes/upstream/<branch> and tests ancestor status. If the fork is an ancestor, it performs an instant fast-forward. If non-conflicting commits exist on both sides, a 3-way merge is minted. Conflicting changes fail closed with HTTP 409 Conflict.',
      },
      {
        id: 'mcp-and-cli-tooling',
        title: 'CLI, REST, and Autonomous MCP Tooling',
        content:
          'Operators and agents can manage the entire fork lifecycle using refweir repo fork, refweir repo forks, and refweir repo sync-fork, or via the embedded MCP server using tools fork_repository, list_forks, and sync_fork.',
      },
    ],
    callouts: [
      {
        type: 'tip',
        title: 'Two-Phase Concurrency Locks',
        content:
          'Refweir always locks source and destination repositories in lexicographical order (srcKey < targetKey), completely preventing deadlocks when reciprocal forks run simultaneously.',
      },
      {
        type: 'important',
        title: 'Lineage Preservation',
        content:
          'Forks inherit all branches and tags from their parent, but commit and branch permissions remain strictly isolated within the fork owner namespace.',
      },
    ],
    codeSnippets: [
      {
        language: 'bash',
        label: 'CLI Commands',
        code: `# Fork an upstream repository
refweir repo fork upstream/core --org contributor --name core

# List all downstream forks of a repository
refweir repo forks upstream/core

# Synchronize fork branch from upstream parent
refweir repo sync-fork contributor/core --branch main`,
      },
      {
        language: 'curl',
        label: 'REST API',
        code: `# 1. Create a fork
curl -X POST http://localhost:8080/api/v1/repos/upstream/core/forks \\
  -H "Authorization: Bearer $REFWEIR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{"organization": "contributor", "name": "core"}'

# 2. Synchronize fork branch with upstream
curl -X POST http://localhost:8080/api/v1/repos/contributor/core/forks/sync \\
  -H "Authorization: Bearer $REFWEIR_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{"branch": "main"}'`,
      },
      {
        language: 'json',
        label: 'MCP Tool Calls',
        code: `{
  "method": "tools/call",
  "params": {
    "name": "sync_fork",
    "arguments": {
      "ownerSlug": "contributor",
      "repoName": "core",
      "branch": "main"
    }
  }
}`,
      },
    ],
    params: [
      {
        name: 'ownerSlug',
        type: 'string',
        required: true,
        description: 'Namespace or owner of the forked repository.',
      },
      {
        name: 'repoName',
        type: 'string',
        required: true,
        description: 'Repository name of the fork.',
      },
      {
        name: 'branch',
        type: 'string',
        required: false,
        description: 'Target branch to synchronize. Defaults to the repository default branch.',
        default: 'main',
      },
    ],
    relatedArticles: ['quickstart', 'mcp-tools-reference', 'cli-reference'],
  },

  'agentic-workflows': {
    id: 'agentic-workflows',
    categoryId: 'agents-mcp',
    title: 'Autonomous Agent Integration & Ambient Tokens',
    lead: 'How autonomous agents authenticate, discover tools, acquire ephemeral scoped leases, and safely interact with Refweir.',
    readTime: '6 min read',
    badge: 'Enterprise Agentic',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'agent-architecture', title: 'Agentic Safety Architecture' },
      { id: 'ambient-token-lifecycle', title: 'Ambient Token Lifecycle' },
      { id: 'subagent-delegation', title: 'Subagent Delegation & Concurrency' },
      { id: 'verification-gates', title: 'Verification Gates' },
    ],
    callouts: [
      {
        type: 'important',
        title: 'Zero Permanent Push Credentials',
        content: 'Autonomous coding agents never hold permanent Git write access. All modifications require an active, time-bound Capability Lease with prefix constraints (e.g. `refs/heads/agent/*`).',
      },
      {
        type: 'tip',
        title: 'Streaming Activity Ledger',
        content: 'Agent operations stream live to the Agent Console via Server-Sent Events (SSE), allowing operators to view real-time latency and symbol modifications.',
      },
    ],
    codeSnippets: [
      {
        language: 'curl',
        label: 'Request Scoped Lease',
        code: `curl -X POST http://localhost:8080/api/v1/agent/leases \\
  -H "Authorization: Bearer $REFWEIR_AGENT_TOKEN" \\
  -H "Content-Type: application/json" \\
  -d '{
    "agentId": "claude-code-worker-01",
    "repoId": "shiva/refweir",
    "allowedRefPrefixes": ["refs/heads/agent/"],
    "ttlSeconds": 1800
  }'`,
      },
    ],
    relatedArticles: ['mcp-tools-reference', 'capability-leases', 'background-reviewer'],
  },

  'mcp-tools-reference': {
    id: 'mcp-tools-reference',
    categoryId: 'agents-mcp',
    title: '58 Model Context Protocol (MCP) Tools Reference',
    lead: 'Complete specification of all 58 Model Context Protocol tools exposed by Refweir for autonomous agent navigation, code mutation, repository forking, branch synchronization, and pull request reviews.',
    readTime: '8 min read',
    badge: 'MCP Specification 2024-11-05',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'overview-mcp', title: 'MCP Protocol Overview' },
      { id: 'repository-tools', title: 'Repository & Exploration Tools (1-15)' },
      { id: 'git-mutation-tools', title: 'Branching & Mutation Tools (16-30)' },
      { id: 'pull-request-tools', title: 'Pull Request & Review Tools (31-45)' },
      { id: 'governance-tools', title: 'Rulesets & Lease Management (46-55)' },
      { id: 'forks-and-mirroring-tools', title: 'Forks & Lineage Tools (56-58)' },
    ],
    params: [
      {
        name: 'get_file_contents',
        type: 'tool',
        required: true,
        description: 'Read the contents of a file at a specific Git ref or commit SHA with base64 decoding.',
      },
      {
        name: 'search_symbols',
        type: 'tool',
        required: true,
        description: 'Semantic symbol search across Go, TypeScript, Python, and Rust AST trees.',
      },
      {
        name: 'create_pull_request',
        type: 'tool',
        required: true,
        description: 'Draft or open a pull request with automated check run triggers.',
      },
      {
        name: 'issue_capability_lease',
        type: 'tool',
        required: true,
        description: 'Generate an ephemeral cryptographic token bound to allowed ref prefixes.',
      },
      {
        name: 'trigger_agent_review',
        type: 'tool',
        required: true,
        description: 'Invoke the autonomous background reviewer on an open pull request diff.',
      },
      {
        name: 'fork_repository',
        type: 'tool',
        required: true,
        description: 'Tool 56: Fork a repository into an organization or user namespace with lineage tracking.',
      },
      {
        name: 'list_forks',
        type: 'tool',
        required: true,
        description: 'Tool 57: List downstream repositories forked from a specified parent repository.',
      },
      {
        name: 'sync_fork',
        type: 'tool',
        required: true,
        description: 'Tool 58: Synchronize a forked repository branch with its upstream parent repository via fast-forward or 3-way merge.',
      },
    ],
    codeSnippets: [
      {
        language: 'json',
        label: 'MCP JSON-RPC Example',
        code: `{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "search_symbols",
    "arguments": {
      "owner": "shiva",
      "repo": "core-engine",
      "query": "NewGitServer",
      "kind": "function"
    }
  }
}`,
      },
    ],
    relatedArticles: ['agentic-workflows', 'capability-leases', 'background-reviewer'],
  },

  'capability-leases': {
    id: 'capability-leases',
    categoryId: 'agents-mcp',
    title: 'Capability Leases & Time-Bound Scopes',
    lead: 'How Refweir uses cryptographic HMAC-SHA256 leases to enforce ephemeral permissions and eliminate autonomous agent permission drift.',
    readTime: '5 min read',
    badge: 'Security Mechanism',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'threat-model', title: 'Why Leases Are Necessary' },
      { id: 'lease-structure', title: 'Cryptographic Lease Structure' },
      { id: 'grace-periods', title: 'Expiry, Renewal, and Grace Periods' },
      { id: 'instant-revocation', title: 'Instant Revocation Mechanism' },
    ],
    callouts: [
      {
        type: 'important',
        title: 'Revocation Propagation',
        content: 'When an operator revokes a lease in the Agent Console, the revocation hash is pushed into an in-memory lock table immediately terminating active Git SSH and Smart-HTTP push sessions.',
      },
    ],
    codeSnippets: [
      {
        language: 'bash',
        label: 'CLI Revoke',
        code: `# Revoke an active agent lease immediately
refweir lease revoke --id lease_1732418000 --reason "Operator manual abort"`,
      },
    ],
    relatedArticles: ['agentic-workflows', 'mcp-tools-reference', 'rulesets-policy'],
  },

  'background-reviewer': {
    id: 'background-reviewer',
    categoryId: 'agents-mcp',
    title: 'Background Agent Reviewer & Automated Check Runs',
    lead: 'Configure the built-in autonomous code reviewer to inspect pull requests for security flaws, ruleset violations, and performance regressions.',
    readTime: '4 min read',
    badge: 'Automated CI/CD',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'reviewer-overview', title: 'Reviewer Architecture' },
      { id: 'trigger-events', title: 'Trigger Events & Webhook Integrations' },
      { id: 'verdict-types', title: 'Verdicts: Approve, Comment, or Request Changes' },
    ],
    codeSnippets: [
      {
        language: 'curl',
        label: 'Trigger Review via API',
        code: `curl -X POST http://localhost:8080/api/v1/repos/shiva/core-engine/pulls/42/reviews/agent \\
  -H "Authorization: Bearer $REFWEIR_TOKEN"`,
      },
    ],
    relatedArticles: ['agentic-workflows', 'mcp-tools-reference', 'rulesets-policy'],
  },

  'rulesets-policy': {
    id: 'rulesets-policy',
    categoryId: 'security-governance',
    title: 'Branch Rulesets & Push Protection',
    lead: 'Enforce cryptographic push verification, mandatory PR reviews, linear commit history, and secret scanning at the Git transport layer.',
    readTime: '5 min read',
    badge: 'Compliance & Governance',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'ruleset-modes', title: 'Ruleset Modes: Active vs Evaluate' },
      { id: 'linear-history', title: 'Enforce Linear History & Fast-Forward Merges' },
      { id: 'commit-signing', title: 'SSH & GPG Commit Signing Verification' },
      { id: 'bypass-roles', title: 'Bypass List for Emergency Admin Intervention' },
    ],
    callouts: [
      {
        type: 'warning',
        title: 'Push Protection Enforcement',
        content: 'In Active mode, commits containing unencrypted private keys, cloud tokens, or high-entropy secrets are rejected during the `pre-receive` hook with line-level diagnostics.',
      },
    ],
    codeSnippets: [
      {
        language: 'json',
        label: 'Ruleset Policy Schema',
        code: `{
  "name": "Production Guard",
  "target": "refs/heads/main",
  "enforcement": "active",
  "rules": {
    "requireSignedCommits": true,
    "requirePullRequest": true,
    "requiredApprovingReviews": 1,
    "preventForcePush": true,
    "preventBranchDeletion": true,
    "secretScanning": true
  }
}`,
      },
    ],
    relatedArticles: ['secret-scanning', 'ssh-ca', 'audit-logs'],
  },

  'secret-scanning': {
    id: 'secret-scanning',
    categoryId: 'security-governance',
    title: 'Secret Scanning & High-Entropy Token Detection',
    lead: 'Protect repositories from accidental credential leakage using pure-Go regex patterns, Shannon entropy analysis, and pre-receive validation.',
    readTime: '4 min read',
    badge: 'Security Mechanism',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'how-it-works', title: 'How Secret Scanning Works' },
      { id: 'detected-patterns', title: 'Detected Secret Patterns' },
      { id: 'remediation', title: 'Remediation & Break-Glass Bypasses' },
    ],
    relatedArticles: ['rulesets-policy', 'audit-logs', 'ssh-ca'],
  },

  'ssh-ca': {
    id: 'ssh-ca',
    categoryId: 'security-governance',
    title: 'SSH Certificate Authority & PKI Integration',
    lead: 'Issue short-lived SSH certificates for human developers and autonomous agents without managing individual public keys across servers.',
    readTime: '5 min read',
    badge: 'Enterprise Security',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'ca-architecture', title: 'SSH CA Architecture' },
      { id: 'user-certificates', title: 'User Certificate Signing' },
      { id: 'host-verification', title: 'Host Key Verification' },
    ],
    codeSnippets: [
      {
        language: 'bash',
        label: 'Sign User Certificate',
        code: `# Sign a developer key valid for 8 hours with enterprise principal
ssh-keygen -s /var/lib/refweir/ca/refweir_user_ca -I dev-alice -n alice -V +8h id_ed25519.pub`,
      },
    ],
    relatedArticles: ['rulesets-policy', 'git-ssh-transport', 'sso-integration'],
  },

  'audit-logs': {
    id: 'audit-logs',
    categoryId: 'security-governance',
    title: 'Tamper-Evident SHA-256 Audit Logs',
    lead: 'Refweir records every repository creation, push event, capability lease grant, and user login in an immutable cryptographic hash chain.',
    readTime: '4 min read',
    badge: 'Audit & Compliance',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'audit-ledger', title: 'Cryptographic Audit Ledger' },
      { id: 'supported-events', title: 'Logged Security Events' },
      { id: 'export-formats', title: 'JSON & CSV Export API' },
    ],
    codeSnippets: [
      {
        language: 'curl',
        label: 'Fetch Audit Logs API',
        code: `curl -H "Authorization: Bearer $REFWEIR_TOKEN" \\
  "http://localhost:8080/api/v1/admin/audit-logs?severity=warn&limit=50" | jq .`,
      },
    ],
    relatedArticles: ['rulesets-policy', 'scim-provisioning', 'metrics-observability'],
  },

  'git-http-transport': {
    id: 'git-http-transport',
    categoryId: 'git-storage',
    title: 'Pure-Go Smart-HTTP Transport Engine',
    lead: 'Deep dive into Refweir’s native HTTP Git server implementing the git-upload-pack and git-receive-pack wire protocol on port 8080.',
    readTime: '5 min read',
    badge: 'Transport Engine',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'smart-http-spec', title: 'Smart-HTTP Wire Protocol' },
      { id: 'packet-negotiation', title: 'pkt-line Packet Negotiation' },
      { id: 'lfs-transfer', title: 'Integrated Git LFS Batch Protocol' },
    ],
    relatedArticles: ['git-ssh-transport', 'git-lfs-cas', 'storage-replication'],
  },

  'git-ssh-transport': {
    id: 'git-ssh-transport',
    categoryId: 'git-storage',
    title: 'Pure-Go SSH Transport Server',
    lead: 'Refweir includes a native, non-OpenSSH SSH server built on golang.org/x/crypto/ssh running on port 2222 with zero C dependencies.',
    readTime: '4 min read',
    badge: 'Transport Engine',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'ssh-architecture', title: 'Native SSH Server Architecture' },
      { id: 'key-auth', title: 'Public Key & Certificate Validation' },
      { id: 'command-dispatch', title: 'git-receive-pack & git-upload-pack Dispatch' },
    ],
    relatedArticles: ['git-http-transport', 'ssh-ca', 'storage-replication'],
  },

  'git-lfs-cas': {
    id: 'git-lfs-cas',
    categoryId: 'git-storage',
    title: 'Git LFS Content Addressable Storage',
    lead: 'Large file tracking, chunked upload validation, and deduplication backed by local filesystem or S3 object stores.',
    readTime: '4 min read',
    badge: 'Storage',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'lfs-protocol', title: 'Git LFS API Protocol' },
      { id: 'pointer-verification', title: 'SHA-256 Pointer Verification' },
      { id: 'garbage-collection', title: 'Orphaned Object Pruning' },
    ],
    relatedArticles: ['git-http-transport', 'storage-replication'],
  },

  'storage-replication': {
    id: 'storage-replication',
    categoryId: 'git-storage',
    title: 'Multi-Engine SQLite WAL & Litestream Replication',
    lead: 'How Refweir maintains ACID persistence, concurrent reader connections, and continuous zero-data-loss streaming replication to cloud storage.',
    readTime: '6 min read',
    badge: 'Reliability',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'sqlite-wal-engine', title: 'SQLite WAL Multi-Reader Engine' },
      { id: 'busy-timeout-locking', title: 'Busy Timeout & Write Mutex Management' },
      { id: 'litestream-sync', title: 'Litestream S3 Continuous Sync' },
      { id: 'failover-procedure', title: 'Disaster Recovery & Point-In-Time Restore' },
    ],
    callouts: [
      {
        type: 'tip',
        title: 'Zero Latency WAL Mode',
        content: 'Refweir sets `PRAGMA journal_mode=WAL` and `PRAGMA synchronous=NORMAL` providing ultra-fast write performance while safeguarding transaction integrity across abrupt crashes.',
      },
    ],
    relatedArticles: ['configuration', 'metrics-observability'],
  },

  'rest-api-overview': {
    id: 'rest-api-overview',
    categoryId: 'api-webhooks',
    title: 'REST API Overview & OpenAPI 3.1 Contract',
    lead: 'Explore all standard RESTful endpoints mounted under /api/v1 for programmatically managing repositories, issues, pull requests, and releases.',
    readTime: '6 min read',
    badge: 'REST Specification',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'base-url', title: 'Base URL and Content Negotiation' },
      { id: 'openapi-spec', title: 'Download OpenAPI 3.1 YAML Contract' },
      { id: 'rate-limiting', title: 'Rate Limiting and Sudo Headers' },
    ],
    codeSnippets: [
      {
        language: 'curl',
        label: 'Download OpenAPI Contract',
        code: `curl -s http://localhost:8080/api/v1/openapi.yaml > openapi.yaml`,
      },
      {
        language: 'curl',
        label: 'List Repositories',
        code: `curl -H "Authorization: Bearer $REFWEIR_TOKEN" \\
  http://localhost:8080/api/v1/repos`,
      },
    ],
    relatedArticles: ['webhooks-guide', 'cli-reference', 'agentic-workflows'],
  },

  'webhooks-guide': {
    id: 'webhooks-guide',
    categoryId: 'api-webhooks',
    title: 'Webhooks Delivery & HMAC-SHA256 Signatures',
    lead: 'Subscribe external CI/CD pipelines, Discord bots, and Slack notification services to repository push, issue, and pull request events.',
    readTime: '5 min read',
    badge: 'Integration',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'supported-events', title: 'Supported Webhook Events' },
      { id: 'signature-verification', title: 'Verifying the X-Refweir-Signature Header' },
      { id: 'redelivery-loop', title: 'Automated Exponential Backoff Redelivery' },
    ],
    codeSnippets: [
      {
        language: 'go',
        label: 'Go Signature Verification',
        code: `func VerifySignature(secret, body []byte, signatureHeader string) bool {
    mac := hmac.New(sha256.New, secret)
    mac.Write(body)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signatureHeader))
}`,
      },
    ],
    relatedArticles: ['rest-api-overview', 'cli-reference'],
  },

  'cli-reference': {
    id: 'cli-reference',
    categoryId: 'api-webhooks',
    title: 'Command Line Reference: refweir and refweird',
    lead: 'Comprehensive command line reference for the operator daemon (refweird) and developer workstation client (refweir).',
    readTime: '5 min read',
    badge: 'CLI Reference',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'refweird-daemon', title: 'refweird Daemon Commands' },
      { id: 'refweir-client', title: 'refweir Client Commands' },
    ],
    codeSnippets: [
      {
        language: 'bash',
        label: 'Operator Commands',
        code: `# Run cluster bootstrap
refweird bootstrap --email admin@refweir.internal

# Check database schema migrations
refweird db migrate

# Inspect cluster health
refweird health`,
      },
      {
        language: 'bash',
        label: 'Developer Commands',
        code: `# Authenticate developer workstation
refweir auth login --endpoint http://localhost:8080

# List repository pull requests
refweir pr list

# Request an ephemeral capability lease
refweir lease create --ttl 30m`,
      },
    ],
    relatedArticles: ['quickstart', 'configuration', 'rest-api-overview'],
  },

  'sso-integration': {
    id: 'sso-integration',
    categoryId: 'enterprise-admin',
    title: 'Single Sign-On (SAML 2.0 & OIDC) Integration',
    lead: 'Integrate enterprise identity providers (Okta, Microsoft Entra ID, Google Workspace, PingIdentity) for centralized user authentication.',
    readTime: '6 min read',
    badge: 'Enterprise Identity',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'supported-protocols', title: 'Supported SSO Protocols' },
      { id: 'oidc-configuration', title: 'OIDC Discovery & PKCE Setup' },
      { id: 'saml-metadata', title: 'SAML 2.0 Entity ID and ACS URL' },
      { id: 'enforce-sso', title: 'Enforce SSO & Break-Glass Accounts' },
    ],
    relatedArticles: ['scim-provisioning', 'audit-logs'],
  },

  'scim-provisioning': {
    id: 'scim-provisioning',
    categoryId: 'enterprise-admin',
    title: 'SCIM 2.0 Directory Sync & User Lifecycle',
    lead: 'Automate user onboarding, role assignments, team memberships, and immediate offboarding via RFC 7643 / 7644 SCIM 2.0 endpoints.',
    readTime: '5 min read',
    badge: 'Identity Governance',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'scim-endpoints', title: 'Mounted SCIM Endpoints' },
      { id: 'attribute-mapping', title: 'User & Group Attribute Mappings' },
      { id: 'deprovisioning', title: 'Instant Deprovisioning & Token Invalidation' },
    ],
    relatedArticles: ['sso-integration', 'audit-logs'],
  },

  'metrics-observability': {
    id: 'metrics-observability',
    categoryId: 'enterprise-admin',
    title: 'Metrics, Ring-Buffer History & Telemetry',
    lead: 'Monitor cluster throughput, Git transfers, MCP tool invocations, and memory allocation via Prometheus endpoints and live charts.',
    readTime: '4 min read',
    badge: 'Observability',
    lastUpdated: 'November 2026',
    sections: [
      { id: 'prometheus-scrape', title: 'Prometheus Scrape Endpoint' },
      { id: 'historical-retention', title: '120-Sample Historical Ring Buffer' },
      { id: 'health-probes', title: 'Liveness & Readiness Probes (/healthz)' },
    ],
    codeSnippets: [
      {
        language: 'curl',
        label: 'Prometheus Scrape',
        code: `curl -s http://localhost:8080/metrics
# HELP refweir_http_requests_total Total HTTP requests received
# TYPE refweir_http_requests_total counter
# HELP refweir_mcp_tool_invocations_total Total MCP tool executions
# TYPE refweir_mcp_tool_invocations_total counter`,
      },
      {
        language: 'curl',
        label: 'Historical Metrics API',
        code: `curl -H "Authorization: Bearer $REFWEIR_TOKEN" \\
  http://localhost:8080/api/v1/admin/metrics/history | jq .`,
      },
    ],
    relatedArticles: ['configuration', 'storage-replication'],
  },
}
