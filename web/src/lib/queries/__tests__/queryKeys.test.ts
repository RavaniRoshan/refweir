import { describe, it, expect } from 'vitest';
import { queryKeys } from '../queryKeys';

describe('queryKeys factory', () => {
  it('constructs predictable hierarchical keys for repo queries', () => {
    expect(queryKeys.repos.all).toEqual(['repos']);
    expect(queryKeys.repos.detail('shiva', 'exist')).toEqual(['repos', 'detail', 'shiva', 'exist']);
    expect(queryKeys.repos.commits('shiva', 'exist', 'feat')).toEqual(['repos', 'commits', 'shiva', 'exist', 'feat']);
    expect(queryKeys.repos.files('shiva', 'exist', 'main', 'src')).toEqual(['repos', 'files', 'shiva', 'exist', 'main', 'src']);
    expect(queryKeys.repos.content('shiva', 'exist', 'README.md', 'main')).toEqual(['repos', 'content', 'shiva', 'exist', 'main', 'README.md']);
    expect(queryKeys.repos.pulls('shiva', 'exist')).toEqual(['repos', 'pulls', 'shiva', 'exist']);
    expect(queryKeys.repos.pullDetail('shiva', 'exist', 42)).toEqual(['repos', 'pulls', 'shiva', 'exist', 42]);
    expect(queryKeys.repos.issues('shiva', 'exist')).toEqual(['repos', 'issues', 'shiva', 'exist']);
    expect(queryKeys.repos.releases('shiva', 'exist')).toEqual(['repos', 'releases', 'shiva', 'exist']);
  });

  it('constructs agent and cluster query keys', () => {
    expect(queryKeys.agent.activities).toEqual(['agent', 'activities']);
    expect(queryKeys.agent.mcpTools).toEqual(['agent', 'mcpTools']);
    expect(queryKeys.cluster.info).toEqual(['cluster', 'info']);
    expect(queryKeys.cluster.metrics).toEqual(['cluster', 'metrics']);
    expect(queryKeys.cluster.status).toEqual(['cluster', 'status']);
  });
});
