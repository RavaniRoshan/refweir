import { useQuery, type QueryClient, type UseQueryOptions } from '@tanstack/react-query';
import { getAgentActivities, queryLocalMcpTools } from '@/lib/api';
import type { AgentActivityItem } from '@/lib/types';
import { useSafeQueryClient } from './useRepoQueries';
import { queryKeys } from './queryKeys';

export function useAgentActivitiesQuery(
  options?: Partial<UseQueryOptions<AgentActivityItem[], Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.agent.activities,
      queryFn: ({ signal }) => getAgentActivities(signal),
      ...options,
    },
    queryClient
  );
}

export function useMcpToolsQuery(
  options?: Partial<UseQueryOptions<any[], Error>>,
  client?: QueryClient
) {
  const queryClient = useSafeQueryClient(client);
  return useQuery(
    {
      queryKey: queryKeys.agent.mcpTools,
      queryFn: () => queryLocalMcpTools(),
      ...options,
    },
    queryClient
  );
}
