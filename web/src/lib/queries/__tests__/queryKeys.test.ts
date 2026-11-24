import { describe, it, expect } from 'vitest';
import { queryKeys } from '../queryKeys';

describe('queryKeys factory', () => {
  it('constructs predictable hierarchical keys for repo queries', () => {
    expect(queryKeys.repos.all).toEqual(['repos']);
    expect(queryKeys.repos.detail('shiva', 'refweir')).toEqual(['repos', 'detail', 'shiva', 'refweir']);
    expect(queryKeys.repos.commits('shiva', 'refweir', 'feat')).toEqual(['repos', 'commits', 'shiva', 'refweir', 'feat']);
    expect(queryKeys.repos.files('shiva', 'refweir', 'main', 'src')).toEqual(['repos', 'files', 'shiva', 'refweir', 'main', 'src']);
    expect(queryKeys.repos.content('shiva', 'refweir', 'README.md', 'main')).toEqual(['repos', 'content', 'shiva', 'refweir', 'main', 'README.md']);
    expect(queryKeys.repos.pulls('shiva', 'refweir')).toEqual(['repos', 'pulls', 'shiva', 'refweir']);
    expect(queryKeys.repos.pullDetail('shiva', 'refweir', 42)).toEqual(['repos', 'pulls', 'shiva', 'refweir', 42]);
    expect(queryKeys.repos.issues('shiva', 'refweir')).toEqual(['repos', 'issues', 'shiva', 'refweir']);
    expect(queryKeys.repos.releases('shiva', 'refweir')).toEqual(['repos', 'releases', 'shiva', 'refweir']);
  });

  it('constructs agent and cluster query keys', () => {
    expect(queryKeys.agent.activities).toEqual(['agent', 'activities']);
    expect(queryKeys.agent.mcpTools).toEqual(['agent', 'mcpTools']);
    expect(queryKeys.cluster.info).toEqual(['cluster', 'info']);
    expect(queryKeys.cluster.metrics).toEqual(['cluster', 'metrics']);
    expect(queryKeys.cluster.status).toEqual(['cluster', 'status']);
  });
});
