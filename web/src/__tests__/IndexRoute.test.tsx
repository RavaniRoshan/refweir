import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { Route } from '@/routes/index';

vi.mock('@/lib/api', async () => {
  const actual = await vi.importActual<any>('@/lib/api');
  return {
    ...actual,
    checkActiveSession: vi.fn().mockResolvedValue({
      authenticated: true,
      user: { id: 'u-1', name: 'Test User', login: 'tester' },
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
  };
});

describe('IndexRoute', () => {
  it('validates search parameters for tab and shell', () => {
    const options = Route.options;
    expect(options.validateSearch).toBeDefined();

    const parsed = options.validateSearch?.({ tab: 'agents', shell: 'true', unknown: 123 });
    expect(parsed).toEqual({
      tab: 'agents',
      shell: 'true',
    });

    const empty = options.validateSearch?.({});
    expect(empty).toEqual({
      tab: undefined,
      shell: undefined,
    });
  });

  it('renders Refweir App as the index page component', async () => {
    const Component = Route.options.component;
    expect(Component).toBeDefined();
    if (!Component) throw new Error('Route component is undefined');

    render(React.createElement(Component));
    await waitFor(() => {
      expect(screen.getAllByText('refweir.dev').length).toBeGreaterThan(0);
    });
  });
});
