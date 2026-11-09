import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useCreatePullRequestMutation,
  useMergePullRequestMutation,
  useCreateIssueMutation,
  useCreateReleaseMutation,
} from '../useRepoMutations';
import * as api from '@/lib/api';

vi.mock('@/lib/api', () => ({
  createPullRequest: vi.fn(),
  mergePullRequest: vi.fn(),
  createIssue: vi.fn(),
  createRelease: vi.fn(),
}));

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('useRepoMutations', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
  });

  it('invalidates pull requests query on successful createPullRequest', async () => {
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');
    vi.mocked(api.createPullRequest).mockResolvedValueOnce({
      id: 'pr-new',
      number: 2,
      title: 'New PR',
    } as any);

    const { result } = renderHook(
      () => useCreatePullRequestMutation('shiva', 'exist'),
      { wrapper: createWrapper(queryClient) }
    );

    await act(async () => {
      await result.current.mutateAsync({
        title: 'New PR',
        head: 'feature',
        base: 'main',
      });
    });

    expect(api.createPullRequest).toHaveBeenCalledWith('shiva', 'exist', {
      title: 'New PR',
      head: 'feature',
      base: 'main',
      body: undefined,
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ['repos', 'pulls', 'shiva', 'exist'],
    });
  });

  it('invalidates pull requests and commits on successful mergePullRequest', async () => {
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');
    vi.mocked(api.mergePullRequest).mockResolvedValueOnce({
      ok: true,
      message: 'Merged',
    } as any);

    const { result } = renderHook(
      () => useMergePullRequestMutation('shiva', 'exist'),
      { wrapper: createWrapper(queryClient) }
    );

    await act(async () => {
      await result.current.mutateAsync({
        prNumber: 42,
        mergeMethod: 'squash',
      });
    });

    expect(api.mergePullRequest).toHaveBeenCalledWith('shiva', 'exist', 42, 'squash');
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ['repos', 'pulls', 'shiva', 'exist'],
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ['repos', 'commits', 'shiva', 'exist'],
    });
  });

  it('invalidates issues query on successful createIssue', async () => {
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');
    vi.mocked(api.createIssue).mockResolvedValueOnce({
      id: 'iss-1',
      number: 1,
      title: 'Bug',
    } as any);

    const { result } = renderHook(
      () => useCreateIssueMutation('shiva', 'exist'),
      { wrapper: createWrapper(queryClient) }
    );

    await act(async () => {
      await result.current.mutateAsync({
        title: 'Bug',
        body: 'Details',
      });
    });

    expect(api.createIssue).toHaveBeenCalledWith('shiva', 'exist', 'Bug', 'Details');
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ['repos', 'issues', 'shiva', 'exist'],
    });
  });

  it('invalidates releases query on successful createRelease', async () => {
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');
    vi.mocked(api.createRelease).mockResolvedValueOnce({
      id: 'rel-1',
      tagName: 'v1.1.0',
      name: 'v1.1.0',
    } as any);

    const { result } = renderHook(
      () => useCreateReleaseMutation('shiva', 'exist'),
      { wrapper: createWrapper(queryClient) }
    );

    await act(async () => {
      await result.current.mutateAsync({
        tagName: 'v1.1.0',
        name: 'v1.1.0',
      });
    });

    expect(api.createRelease).toHaveBeenCalledWith('shiva', 'exist', {
      tagName: 'v1.1.0',
      name: 'v1.1.0',
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ['repos', 'releases', 'shiva', 'exist'],
    });
  });
});
