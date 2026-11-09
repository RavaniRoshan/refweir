import { useContext } from 'react';
import {
  useQuery,
  QueryClientContext,
  type QueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  getRepoInfo,
  getRepositories,
  getCommits,
  getFiles,
  getFileContent,
  getPullRequests,
  getIssues,
  getReleases,
  getBranches,
  type BranchItem,
} from '@/lib/api';
import type {
  RepoInfo,
  RepositoryRecord,
  CommitItem,
  FileItem,
  PullRequest,
  Issue,
  ReleaseItem,
} from '@/lib/types';
import { queryClient as defaultClient } from '@/lib/queryClient';
import { queryKeys } from './queryKeys';

export function useSafeQueryClient(client?: QueryClient): QueryClient {
  const contextClient = useContext(QueryClientContext);
  return client || contextClient || defaultClient;
}

export function useRepoInfoQuery(
  owner: string,
  name: string,
  options?: Partial<UseQueryOptions<RepoInfo, Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.repos.detail(owner, name),
      queryFn: ({ signal }) => getRepoInfo(owner, name, signal),
      enabled: Boolean(owner && name),
      ...options,
    },
    queryClient
  );
}

export function useRepositoriesQuery(
  options?: Partial<UseQueryOptions<RepositoryRecord[], Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.repos.all,
      queryFn: ({ signal }) => getRepositories(signal),
      ...options,
    },
    queryClient
  );
}

export function useCommitsQuery(
  owner: string,
  name: string,
  ref = 'main',
  options?: Partial<UseQueryOptions<CommitItem[], Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.repos.commits(owner, name, ref),
      queryFn: ({ signal }) => getCommits(owner, name, ref, signal),
      enabled: Boolean(owner && name),
      ...options,
    },
    queryClient
  );
}

export function useFilesQuery(
  owner: string,
  name: string,
  ref = 'main',
  path = '',
  options?: Partial<UseQueryOptions<FileItem[], Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.repos.files(owner, name, ref, path),
      queryFn: ({ signal }) => getFiles(owner, name, ref, path, signal),
      enabled: Boolean(owner && name),
      ...options,
    },
    queryClient
  );
}

export function useFileContentQuery(
  owner: string,
  name: string,
  path: string,
  ref = 'main',
  options?: Partial<UseQueryOptions<string, Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.repos.content(owner, name, path, ref),
      queryFn: ({ signal }) => getFileContent(owner, name, path, ref, signal),
      enabled: Boolean(owner && name && path),
      ...options,
    },
    queryClient
  );
}

export function usePullRequestsQuery(
  owner: string,
  name: string,
  options?: Partial<UseQueryOptions<PullRequest[], Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.repos.pulls(owner, name),
      queryFn: ({ signal }) => getPullRequests(owner, name, signal),
      enabled: Boolean(owner && name),
      ...options,
    },
    queryClient
  );
}

export function useIssuesQuery(
  owner: string,
  name: string,
  options?: Partial<UseQueryOptions<Issue[], Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.repos.issues(owner, name),
      queryFn: ({ signal }) => getIssues(owner, name, signal),
      enabled: Boolean(owner && name),
      ...options,
    },
    queryClient
  );
}

export function useReleasesQuery(
  owner: string,
  name: string,
  options?: Partial<UseQueryOptions<ReleaseItem[], Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.repos.releases(owner, name),
      queryFn: ({ signal }) => getReleases(owner, name, signal),
      enabled: Boolean(owner && name),
      ...options,
    },
    queryClient
  );
}

export function useBranchesQuery(
  owner: string,
  name: string,
  options?: Partial<UseQueryOptions<BranchItem[], Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.repos.branches(owner, name),
      queryFn: ({ signal }) => getBranches(owner, name, signal),
      enabled: Boolean(owner && name),
      ...options,
    },
    queryClient
  );
}
