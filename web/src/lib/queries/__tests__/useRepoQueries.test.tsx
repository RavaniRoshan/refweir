import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useRepoInfoQuery,
  usePullRequestsQuery,
  useIssuesQuery,
  useReleasesQuery,
} from '../useRepoQueries';
import * as api from '@/lib/api';

vi.mock('@/lib/api', () => ({
  getRepoInfo: vi.fn(),
  getPullRequests: vi.fn(),
  getIssues: vi.fn(),
  getReleases: vi.fn(),
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('useRepoQueries', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches and returns repository details', async () => {
    const mockRepo = {
      ownerSlug: 'shiva',
      name: 'refweir',
      description: 'Enterprise forge',
      defaultBranch: 'main',
      isPrivate: false,
      isArchived: false,
      starsCount: 10,
      forksCount: 2,
      updatedAt: 'Just now',
      cloneUrlHttp: 'http://localhost/repo.git',
      cloneUrlSsh: 'git@localhost:repo.git',
    };
    vi.mocked(api.getRepoInfo).mockResolvedValueOnce(mockRepo);

    const { result } = renderHook(() => useRepoInfoQuery('shiva', 'refweir'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockRepo);
    expect(api.getRepoInfo).toHaveBeenCalledWith('shiva', 'refweir', expect.anything());
  });

  it('fetches and returns pull requests', async () => {
    const mockPulls = [
      {
        id: 'pr-1',
        number: 1,
        title: 'Initial PR',
        state: 'open' as const,
        author: 'shiva',
        createdAt: 'Just now',
        headRef: 'feat',
        baseRef: 'main',
        branch: 'feat',
      },
    ];
    vi.mocked(api.getPullRequests).mockResolvedValueOnce(mockPulls as any);

    const { result } = renderHook(() => usePullRequestsQuery('shiva', 'refweir'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockPulls);
    expect(api.getPullRequests).toHaveBeenCalledWith('shiva', 'refweir', expect.anything());
  });

  it('fetches and returns issues', async () => {
    const mockIssues = [
      {
        id: 'issue-1',
        number: 10,
        title: 'Bug report',
        state: 'open' as const,
        author: 'user',
        createdAt: 'Just now',
      },
    ];
    vi.mocked(api.getIssues).mockResolvedValueOnce(mockIssues as any);

    const { result } = renderHook(() => useIssuesQuery('shiva', 'refweir'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockIssues);
  });

  it('fetches and returns releases', async () => {
    const mockReleases = [
      {
        id: 'rel-1',
        tagName: 'v1.0.0',
        name: 'Release 1.0',
        body: 'First stable release',
        author: 'shiva',
        createdAt: 'Just now',
      },
    ];
    vi.mocked(api.getReleases).mockResolvedValueOnce(mockReleases as any);

    const { result } = renderHook(() => useReleasesQuery('shiva', 'refweir'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockReleases);
  });
});
