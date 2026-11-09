/**
 * Hierarchical TanStack Query keys factory for type-safe query caching and invalidation.
 */
export const queryKeys = {
  repos: {
    all: ['repos'] as const,
    detail: (owner: string, name: string) => ['repos', 'detail', owner, name] as const,
    commits: (owner: string, name: string, ref = 'main') => ['repos', 'commits', owner, name, ref] as const,
    files: (owner: string, name: string, ref = 'main', path = '') => ['repos', 'files', owner, name, ref, path] as const,
    content: (owner: string, name: string, path: string, ref = 'main') => ['repos', 'content', owner, name, ref, path] as const,
    pulls: (owner: string, name: string) => ['repos', 'pulls', owner, name] as const,
    pullDetail: (owner: string, name: string, number: number) => ['repos', 'pulls', owner, name, number] as const,
    pullComments: (owner: string, name: string, number: number) => ['repos', 'pulls', owner, name, number, 'comments'] as const,
    issues: (owner: string, name: string) => ['repos', 'issues', owner, name] as const,
    releases: (owner: string, name: string) => ['repos', 'releases', owner, name] as const,
    branches: (owner: string, name: string) => ['repos', 'branches', owner, name] as const,
    rulesets: (owner: string, name: string) => ['repos', 'rulesets', owner, name] as const,
    collaborators: (owner: string, name: string) => ['repos', 'collaborators', owner, name] as const,
    deployKeys: (owner: string, name: string) => ['repos', 'deployKeys', owner, name] as const,
    webhooks: (owner: string, name: string) => ['repos', 'webhooks', owner, name] as const,
  },
  agent: {
    activities: ['agent', 'activities'] as const,
    mcpTools: ['agent', 'mcpTools'] as const,
  },
  cluster: {
    info: ['cluster', 'info'] as const,
    metrics: ['cluster', 'metrics'] as const,
    status: ['cluster', 'status'] as const,
  },
  user: {
    session: ['user', 'session'] as const,
    sshKeys: ['user', 'sshKeys'] as const,
    tokens: ['user', 'tokens'] as const,
    notifications: (unreadOnly = false) => ['user', 'notifications', { unreadOnly }] as const,
    unreadNotificationCount: ['user', 'notifications', 'unreadCount'] as const,
  },
};
