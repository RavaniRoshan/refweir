import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

if (typeof window !== 'undefined') {
  window.addEventListener('refweir_mode_changed', () => {
    queryClient.clear();
  });
}
