import React, { useState } from 'react';
import { useUIStore } from '@/lib/store';
import { cn } from '@/lib/utils';

export interface AppShellProps {
  children: React.ReactNode;
}

const NAV_ITEMS = [
  { id: 'inbox', label: 'Inbox', icon: '📥' },
  { id: 'runs', label: 'Runs', icon: '▶' },
  { id: 'reviews', label: 'Reviews', icon: '✓' },
  { id: 'policies', label: 'Policies', icon: '🛡' },
  { id: 'repos', label: 'Repos', icon: '📁' },
] as const;

export function AppShell({ children }: AppShellProps) {
  const { sidebarOpen, setSidebarOpen, theme, setTheme } = useUIStore();
  const [activeNav, setActiveNav] = useState<string>('repos');

  return (
    <div className="flex h-screen w-full overflow-hidden bg-background text-foreground">
      {/* Global rail */}
      <aside
        className={cn(
          'flex flex-col border-r border-border bg-slab transition-all duration-200',
          sidebarOpen ? 'w-56' : 'w-14',
        )}
      >
        <div className="flex h-12 items-center justify-between px-3 border-b border-border">
          <span className={cn('font-semibold text-sm', !sidebarOpen && 'hidden')}>Refweir</span>
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="text-text-2 hover:text-text-1 text-xs"
            aria-label="Toggle sidebar"
          >
            {sidebarOpen ? '◀' : '▶'}
          </button>
        </div>
        <nav className="flex-1 overflow-y-auto py-2">
          {NAV_ITEMS.map((item) => (
            <button
              key={item.id}
              onClick={() => setActiveNav(item.id)}
              className={cn(
                'flex w-full items-center gap-3 px-3 py-2 text-sm transition-colors',
                activeNav === item.id
                  ? 'bg-accent text-accent-foreground'
                  : 'text-text-2 hover:bg-raised hover:text-text-1',
              )}
            >
              <span className="text-base">{item.icon}</span>
              {sidebarOpen && <span>{item.label}</span>}
            </button>
          ))}
        </nav>
        <div className="border-t border-border p-3">
          <button
            onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
            className="text-xs text-text-2 hover:text-text-1"
          >
            {sidebarOpen ? (theme === 'dark' ? '☀ Light' : '🌙 Dark') : (theme === 'dark' ? '☀' : '🌙')}
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto">
        {/* Top bar */}
        <header className="sticky top-0 z-10 flex h-12 items-center justify-between border-b border-border bg-canvas/80 px-4 backdrop-blur">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-text-1">{activeNav}</span>
          </div>
          <div className="flex items-center gap-2">
            <kbd className="rounded border border-border px-1.5 py-0.5 text-xs text-text-3">⌘K</kbd>
          </div>
        </header>
        <div className="p-4">{children}</div>
      </main>
    </div>
  );
}
