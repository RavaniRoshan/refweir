import { useQuery, useMutation, type QueryClient, type UseQueryOptions, type UseMutationOptions } from '@tanstack/react-query';
import {
  fetchClusterInfo,
  getClusterMetrics,
  fetchLiveServerStatus,
  fetchClusterReplication,
  triggerClusterCheckpoint,
  triggerClusterSnapshot,
  type ClusterSnapshotMetadata,
  type ClusterReplicationStatus,
} from '@/lib/api';
import type { ClusterSystemInfo, ClusterMetrics } from '@/lib/types';
import { useSafeQueryClient } from './useRepoQueries';
import { queryKeys } from './queryKeys';

export function useClusterInfoQuery(
  options?: Partial<UseQueryOptions<ClusterSystemInfo, Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.cluster.info,
      queryFn: ({ signal }) => fetchClusterInfo(signal),
      ...options,
    },
    queryClient
  );
}

export function useClusterMetricsQuery(
  options?: Partial<UseQueryOptions<ClusterMetrics, Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.cluster.metrics,
      queryFn: ({ signal }) => getClusterMetrics(signal),
      ...options,
    },
    queryClient
  );
}

export function useServerStatusQuery(
  options?: Partial<UseQueryOptions<{ ok: boolean; status: string; latencyMs: number }, Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.cluster.status,
      queryFn: ({ signal }) => fetchLiveServerStatus(signal),
      ...options,
    },
    queryClient
  );
}

export function useClusterReplicationQuery(
  options?: Partial<UseQueryOptions<ClusterReplicationStatus, Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: ['cluster', 'replication'],
      queryFn: ({ signal }) => fetchClusterReplication(signal),
      ...options,
    },
    queryClient
  );
}

export function useClusterCheckpointMutation(
  options?: UseMutationOptions<{ status: string; mode?: string }, Error, void>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useMutation(
    {
      mutationFn: () => triggerClusterCheckpoint(),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: queryKeys.cluster.metrics });
      },
      ...options,
    },
    queryClient
  );
}

export function useClusterSnapshotMutation(
  options?: UseMutationOptions<ClusterSnapshotMetadata, Error, void>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useMutation(
    {
      mutationFn: () => triggerClusterSnapshot(),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: queryKeys.cluster.metrics });
      },
      ...options,
    },
    queryClient
  );
}
