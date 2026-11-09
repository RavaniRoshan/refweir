import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
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
  it('toggles sidebar', () => {
    render(<AppShell><div>x</div></AppShell>);
    const toggle = screen.getByLabelText('Toggle sidebar');
    expect(toggle).toBeTruthy();
  });
});
