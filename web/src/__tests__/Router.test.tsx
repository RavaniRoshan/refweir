import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { RouterProvider, createRouter, createMemoryHistory } from '@tanstack/react-router';
import { QueryClientProvider } from '@tanstack/react-query';
import { router } from '@/router';
import { routeTree } from '@/routeTree.gen';
import { queryClient } from '@/lib/queryClient';

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<any>('@/lib/api');
  return {
    ...actual,
    checkActiveSession: vi.fn().mockResolvedValue({
      authenticated: true,
      user: { id: 'u-1', name: 'Roshan Ravani', login: 'shiva' },
    }),
    getRepoInfo: vi.fn().mockResolvedValue({
      ownerSlug: 'shiva',
      name: 'refweir.dev',
      defaultBranch: 'main',
    }),
    getRepositories: vi.fn().mockResolvedValue([]),
    getFiles: vi.fn().mockResolvedValue([]),
    getFileContent: vi.fn().mockResolvedValue(''),
    getCommits: vi.fn().mockResolvedValue([]),
    getPullRequests: vi.fn().mockResolvedValue([]),
    getIssues: vi.fn().mockResolvedValue([]),
  };
});

describe('Root Router & QueryClient Provider Integration', () => {
  it('configures TanStack router with intent preloading', () => {
    expect(router.options.defaultPreload).toBe('intent');
    expect(router.options.defaultPreloadStaleTime).toBe(30_000);
  });

  it('renders root router tree with QueryClientProvider and mounts index route App', async () => {
    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getAllByText('refweir.dev').length).toBeGreaterThan(0);
    });
  });

  it('switches views when router navigates with search params', async () => {
    const memoryRouter = createRouter({
      routeTree,
      history: createMemoryHistory({ initialEntries: ['/?tab=issues'] }),
    });

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={memoryRouter} />
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/is:issue/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /new issue/i })).toBeInTheDocument();
    });
  });
});
