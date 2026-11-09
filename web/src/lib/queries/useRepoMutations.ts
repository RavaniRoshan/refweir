import { useMutation, type QueryClient } from '@tanstack/react-query';
import {
  createPullRequest,
  mergePullRequest,
  createIssue,
  createRelease,
  createBranch,
} from '@/lib/api';
import { useSafeQueryClient } from './useRepoQueries';
import { queryKeys } from './queryKeys';

export function useCreatePullRequestMutation(owner: string, name: string, client?: QueryClient) {
  const queryClient = useSafeQueryClient(client);

  return useMutation(
    {
      mutationFn: (params: {
        title: string;
        body?: string;
        head?: string;
        base?: string;
        headBranch?: string;
        baseBranch?: string;
        draft?: boolean;
      }) =>
        createPullRequest(owner, name, {
          title: params.title,
          body: params.body,
          head: params.head || params.headBranch || 'feature',
          base: params.base || params.baseBranch || 'main',
          draft: params.draft,
        }),
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: queryKeys.repos.pulls(owner, name),
        });
      },
    },
    queryClient
  );
}

export function useMergePullRequestMutation(owner: string, name: string, client?: QueryClient) {
  const queryClient = useSafeQueryClient(client);

  return useMutation(
    {
      mutationFn: (params: {
        prNumber: number;
        mergeMethod?: 'merge' | 'squash' | 'rebase';
      }) =>
        mergePullRequest(
          owner,
          name,
          params.prNumber,
          params.mergeMethod === 'rebase' ? 'squash' : params.mergeMethod
        ),
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: queryKeys.repos.pulls(owner, name),
        });
        queryClient.invalidateQueries({
          queryKey: ['repos', 'commits', owner, name],
        });
      },
    },
    queryClient
  );
}

export function useCreateIssueMutation(owner: string, name: string, client?: QueryClient) {
  const queryClient = useSafeQueryClient(client);

  return useMutation(
    {
      mutationFn: (params: { title: string; body?: string }) =>
        createIssue(owner, name, params.title, params.body || ''),
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: queryKeys.repos.issues(owner, name),
        });
      },
    },
    queryClient
  );
}

export function useCreateReleaseMutation(owner: string, name: string, client?: QueryClient) {
  const queryClient = useSafeQueryClient(client);

  return useMutation(
    {
      mutationFn: (params: {
        tagName: string;
        name: string;
        body?: string;
        commitSha?: string;
      }) => createRelease(owner, name, params),
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: queryKeys.repos.releases(owner, name),
        });
      },
    },
    queryClient
  );
}

export function useCreateBranchMutation(owner: string, name: string, client?: QueryClient) {
  const queryClient = useSafeQueryClient(client);

  return useMutation(
    {
      mutationFn: (params: { branchName: string; from?: string }) =>
        createBranch(owner, name, params.branchName, params.from),
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: queryKeys.repos.branches(owner, name),
        });
      },
    },
    queryClient
  );
}
