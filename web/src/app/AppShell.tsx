import React, { useState, useEffect } from 'react';
import {
  Inbox,
  Play,
  CheckSquare,
  Shield,
  FolderGit2,
  ChevronLeft,
  ChevronRight,
  Sun,
  Moon,
} from 'lucide-react';
import { useUIStore } from '@/lib/store';
import { CommandPalette } from '@/components/CommandPalette';
import { cn } from '@/lib/utils';

export interface AppShellProps {
  children: React.ReactNode;
  activeNav?: string;
  onNavChange?: (nav: string) => void;
  onTabChange?: (tab: string) => void;
}

const NAV_ITEMS = [
  { id: 'inbox', label: 'Inbox', icon: Inbox },
  { id: 'runs', label: 'Runs', icon: Play },
  { id: 'reviews', label: 'Reviews', icon: CheckSquare },
  { id: 'policies', label: 'Policies', icon: Shield },
  { id: 'repos', label: 'Repos', icon: FolderGit2 },
] as const;

export function AppShell({ children, activeNav, onNavChange, onTabChange }: AppShellProps) {
  const { sidebarOpen, setSidebarOpen, theme, setTheme } = useUIStore();
  const [internalNav, setInternalNav] = useState<string>('repos');
  const [paletteOpen, setPaletteOpen] = useState(false);

  const currentNav = activeNav ?? internalNav;

  // Determine current active dark mode status
  const isDark =
    theme === 'dark' ||
    (theme === 'system' &&
      (typeof document !== 'undefined'
        ? document.documentElement.classList.contains('dark')
        : false));

  // Sync document root class when theme state updates
  useEffect(() => {
    if (theme === 'system') return;
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [theme]);

  const toggleTheme = () => {
    const nextIsDark = !isDark;
    setTheme(nextIsDark ? 'dark' : 'light');
    if (nextIsDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  };

  const handleNav = (id: string) => {
    if (onNavChange) {
      onNavChange(id);
    } else {
      setInternalNav(id);
    }

    if (onTabChange) {
      if (id === 'runs') onTabChange('agents');
      else if (id === 'reviews') onTabChange('prs');
      else if (id === 'policies') onTabChange('rulesets');
      else if (id === 'inbox') onTabChange('home');
      else if (id === 'repos') onTabChange('code');
    }
  };

  const handleCommandPaletteTab = (tab: string) => {
    if (onTabChange) {
      onTabChange(tab);
    }
    if (tab === 'agents') handleNav('runs');
    else if (tab === 'prs') handleNav('reviews');
    else if (tab === 'rulesets') handleNav('policies');
    else if (tab === 'home') handleNav('inbox');
    else handleNav('repos');
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      if (target?.closest('input, textarea, select, [contenteditable="true"], [role="dialog"]')) return;
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setPaletteOpen((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <div className="flex h-screen w-full overflow-hidden bg-background text-foreground font-sans">
      {/* Global rail */}
      <aside
        className={cn(
          'flex flex-col border-r border-border bg-slab transition-all duration-200 shrink-0 select-none',
          sidebarOpen ? 'w-56' : 'w-14',
        )}
      >
        <div className="flex h-12 items-center justify-between px-3 border-b border-border">
          <span className={cn('font-semibold text-sm text-text-1 tracking-tight', !sidebarOpen && 'hidden')}>
            Refweir
          </span>
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="text-text-2 hover:text-text-1 text-xs cursor-pointer p-1 rounded-md hover:bg-raised transition-colors flex items-center justify-center focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-primary"
            aria-label="Toggle sidebar"
            aria-expanded={sidebarOpen}
          >
            {sidebarOpen ? <ChevronLeft className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
          </button>
        </div>

        <nav
          className="flex-1 overflow-y-auto py-2 space-y-0.5 px-1.5"
          role="tablist"
          aria-label="Global navigation"
        >
          {NAV_ITEMS.map((item) => {
            const Icon = item.icon;
            const isActive = currentNav === item.id;
            return (
              <button
                key={item.id}
                role="tab"
                aria-selected={isActive}
                aria-label={item.label}
                onClick={() => handleNav(item.id)}
                title={!sidebarOpen ? item.label : undefined}
                className={cn(
                  'flex w-full items-center gap-3 px-2.5 py-2 text-sm rounded-md transition-colors cursor-pointer text-left focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-primary',
                  isActive
                    ? 'bg-secondary text-foreground font-medium'
                    : 'text-text-2 hover:bg-raised hover:text-text-1',
                  !sidebarOpen && 'justify-center px-0'
                )}
              >
                <Icon className={cn('w-4 h-4 shrink-0', isActive ? 'text-primary' : 'text-text-2')} />
                {sidebarOpen && <span>{item.label}</span>}
              </button>
            );
          })}
        </nav>

        <div className="border-t border-border p-2">
          <button
            onClick={toggleTheme}
            className={cn(
              'w-full text-xs text-text-2 hover:text-text-1 flex items-center gap-2 cursor-pointer p-2 rounded-md hover:bg-raised transition-colors focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-primary',
              !sidebarOpen && 'justify-center p-1.5'
            )}
            aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
            title={!sidebarOpen ? (isDark ? 'Switch to light mode' : 'Switch to dark mode') : undefined}
          >
            {isDark ? (
              <Sun className="w-4 h-4 shrink-0 text-signal" />
            ) : (
              <Moon className="w-4 h-4 shrink-0 text-text-2" />
            )}
            {sidebarOpen && <span>{isDark ? 'Light' : 'Dark'}</span>}
          </button>
        </div>
      </aside>

      {/* Main content container (div rather than nested main to preserve single main landmark) */}
      <div className="flex-1 overflow-y-auto flex flex-col min-w-0">
        {/* Top bar */}
        <header className="sticky top-0 z-10 flex h-12 shrink-0 items-center justify-between border-b border-border bg-canvas/80 px-4 backdrop-blur">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-text-1 capitalize">{currentNav}</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPaletteOpen(true)}
              className="flex items-center gap-1 rounded-md border border-border px-2 py-1 text-xs text-text-3 hover:text-text-1 hover:border-hairline cursor-pointer transition-colors focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-primary"
              aria-label="Open command palette"
            >
              <kbd className="font-mono text-xs">⌘K</kbd>
            </button>
          </div>
        </header>

        <div className="flex-1 min-h-0">{children}</div>
      </div>

      {/* Integrated Command Palette */}
      <CommandPalette
        open={paletteOpen}
        onOpenChange={setPaletteOpen}
        onTabChange={handleCommandPaletteTab}
      />
    </div>
  );
}
