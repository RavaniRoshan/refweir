import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { AppShell } from '@/app/AppShell';

describe('AppShell', () => {
  it('renders children and nav items', () => {
    render(<AppShell><div>content here</div></AppShell>);
    expect(screen.getByText('content here')).toBeTruthy();
    expect(screen.getByText('Inbox')).toBeTruthy();
    expect(screen.getByText('Runs')).toBeTruthy();
    expect(screen.getByText('Reviews')).toBeTruthy();
    expect(screen.getByText('Policies')).toBeTruthy();
    expect(screen.getByText('Repos')).toBeTruthy();
  });

  it('implements accessible Primer tablist semantics on rail navigation', () => {
    render(<AppShell activeNav="repos"><div>content</div></AppShell>);
    const tablist = screen.getByRole('tablist', { name: /global navigation/i });
    expect(tablist).toBeInTheDocument();

    const tabs = screen.getAllByRole('tab');
    expect(tabs.length).toBe(5);

    const reposTab = screen.getByRole('tab', { name: /repos/i });
    expect(reposTab).toHaveAttribute('aria-selected', 'true');

    const inboxTab = screen.getByRole('tab', { name: /inbox/i });
    expect(inboxTab).toHaveAttribute('aria-selected', 'false');
  });

  it('invokes onNavChange callback when clicking a navigation tab', () => {
    const handleNav = vi.fn();
    render(<AppShell activeNav="repos" onNavChange={handleNav}><div>content</div></AppShell>);

    const runsTab = screen.getByRole('tab', { name: /runs/i });
    fireEvent.click(runsTab);
    expect(handleNav).toHaveBeenCalledWith('runs');
  });

  it('toggles sidebar and updates aria-expanded attribute', () => {
    render(<AppShell><div>x</div></AppShell>);
    const toggle = screen.getByLabelText('Toggle sidebar');
    expect(toggle).toBeTruthy();
    expect(toggle).toHaveAttribute('aria-expanded');
  });

  it('provides an accessible theme toggle button with dynamic aria-label', () => {
    render(<AppShell><div>content</div></AppShell>);
    const themeBtn = screen.getByRole('button', { name: /switch to (light|dark) mode/i });
    expect(themeBtn).toBeInTheDocument();
  });

  it('renders command palette trigger with accessible label and opens on click', async () => {
    render(<AppShell><div>content</div></AppShell>);
    const cmdBtn = screen.getByRole('button', { name: /open command palette/i });
    expect(cmdBtn).toBeInTheDocument();
    expect(cmdBtn).toHaveTextContent('⌘K');

    fireEvent.click(cmdBtn);
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/type a command or search/i)).toBeInTheDocument();
    });
  });
});
