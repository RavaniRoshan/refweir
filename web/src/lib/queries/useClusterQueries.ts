import { useQuery, type QueryClient, type UseQueryOptions } from '@tanstack/react-query';
import { fetchClusterInfo, getClusterMetrics, fetchLiveServerStatus } from '@/lib/api';
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
