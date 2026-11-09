import { useState, useEffect, lazy, Suspense } from 'react'
import {
  Activity,
  Server,
  ShieldAlert,
  KeyRound,
  Users,
  Loader2,
} from 'lucide-react'
import { Header } from '@/components/views/Header'
import { DemoBanner } from '@/components/views/DemoBanner'
import { RepoOverview } from '@/components/views/RepoOverview'
import { GlobalHomeView } from '@/components/views/GlobalHomeView'
import { defaultRepo, getRepoInfo, getRepositories, isDemoModeActive } from '@/lib/api'
import { cn } from '@/lib/utils'
import { AppShell } from '@/app/AppShell'
import { useUIStore } from '@/lib/store'
import { ViewErrorBoundary } from '@/components/ui/ViewErrorBoundary'

// Lazy-loaded views for bundle optimization & code splitting
const PullRequestsView = lazy(() =>
  import('@/components/views/PullRequestsView').then((m) => ({ default: m.PullRequestsView }))
)
const CommitsView = lazy(() =>
  import('@/components/views/CommitsView').then((m) => ({ default: m.CommitsView }))
)
const IssuesView = lazy(() =>
  import('@/components/views/IssuesView').then((m) => ({ default: m.IssuesView }))
)
const ReleasesView = lazy(() =>
  import('@/components/views/ReleasesView').then((m) => ({ default: m.ReleasesView }))
)
const AgentConsoleView = lazy(() =>
  import('@/components/views/AgentConsoleView').then((m) => ({ default: m.AgentConsoleView }))
)
const RulesetsView = lazy(() =>
  import('@/components/views/RulesetsView').then((m) => ({ default: m.RulesetsView }))
)
const WebhooksView = lazy(() =>
  import('@/components/views/WebhooksView').then((m) => ({ default: m.WebhooksView }))
)
const ProductTour = lazy(() =>
  import('@/components/views/ProductTour').then((m) => ({ default: m.ProductTour }))
)
const UserManualModal = lazy(() =>
  import('@/components/views/UserManualModal').then((m) => ({ default: m.UserManualModal }))
)
const RepoSettingsView = lazy(() =>
  import('@/components/views/RepoSettingsView').then((m) => ({ default: m.RepoSettingsView }))
)
const UserProfileView = lazy(() =>
  import('@/components/views/UserProfileView').then((m) => ({ default: m.UserProfileView }))
)
const ClusterHealth = lazy(() =>
  import('@/components/views/admin/ClusterHealth').then((m) => ({ default: m.ClusterHealth }))
)
const AuditLogs = lazy(() =>
  import('@/components/views/admin/AuditLogs').then((m) => ({ default: m.AuditLogs }))
)
const SsoConfig = lazy(() =>
  import('@/components/views/admin/SsoConfig').then((m) => ({ default: m.SsoConfig }))
)
const EnterpriseUsers = lazy(() =>
  import('@/components/views/admin/EnterpriseUsers').then((m) => ({ default: m.EnterpriseUsers }))
)

export const ViewFallback = () => (
  <div
    role="status"
    aria-live="polite"
    className="min-h-[450px] flex flex-col items-center justify-center p-8 text-center text-xs text-muted-foreground border border-dashed border-border/50 rounded-md bg-card/20 animate-pulse"
  >
    <Loader2 className="w-5 h-5 animate-spin mx-auto mb-2 text-primary" aria-hidden="true" />
    <span>Loading view...</span>
  </div>
)

export interface AppProps {
  initialTab?: 'home' | 'profile' | 'code' | 'issues' | 'prs' | 'commits' | 'releases' | 'agents' | 'rulesets' | 'webhooks' | 'admin' | 'settings'
  initialShell?: boolean
}

export default function App({ initialTab, initialShell }: AppProps = {}) {
  const [activeTab, setActiveTab] = useState<'home' | 'profile' | 'code' | 'issues' | 'prs' | 'commits' | 'releases' | 'agents' | 'rulesets' | 'webhooks' | 'admin' | 'settings'>(() => {
    if (initialTab && ['home', 'profile', 'code', 'issues', 'prs', 'commits', 'releases', 'agents', 'rulesets', 'webhooks', 'admin', 'settings'].includes(initialTab)) {
      return initialTab
    }
    if (typeof window !== 'undefined') {
      const p = new URLSearchParams(window.location.search)
      const t = p.get('tab')
      if (t && ['home', 'profile', 'code', 'issues', 'prs', 'commits', 'releases', 'agents', 'rulesets', 'webhooks', 'admin', 'settings'].includes(t)) {
        return t as any
      }
    }
    return 'code'
  })

  useEffect(() => {
    if (initialTab && ['home', 'profile', 'code', 'issues', 'prs', 'commits', 'releases', 'agents', 'rulesets', 'webhooks', 'admin', 'settings'].includes(initialTab)) {
      setActiveTab(initialTab)
    }
  }, [initialTab])
  const [adminSection, setAdminSection] = useState<'cluster' | 'users' | 'sso' | 'audit'>('cluster')
  const [tourOpen, setTourOpen] = useState(false)
  const [systemInfo, setSystemInfo] = useState<{
    version?: string
    mcpMountedToolsCount?: number
    demoMode?: boolean
    demoRepo?: string
  }>({})

  useEffect(() => {
    fetch('/api/v1/system/info')
      .then((response) => (response.ok ? response.json() : null))
      .then((data) => { if (data) setSystemInfo(data) })
      .catch(() => {})
  }, [])

  const [manualOpen, setManualOpen] = useState(false)
  const [currentRepo, setCurrentRepo] = useState(defaultRepo)

  const handleSelectRepo = (slug?: string) => {
    if (!slug) {
      setActiveTab('code')
      return
    }
    let owner = 'shiva'
    let name = slug
    if (slug.includes('/')) {
      const parts = slug.split('/')
      owner = parts[0]
      name = parts[1]
    }
    getRepoInfo(owner, name).then((info) => {
      setCurrentRepo(info)
    })
    setActiveTab('code')
  }

  useEffect(() => {
    // If first-time user hasn't completed tour yet, auto-open tour (unless ?notour=1)
    const p = new URLSearchParams(window.location.search)
    if (p.get('notour') !== '1') {
      const completed = localStorage.getItem('refweir_onboarding_tour_completed')
      if (completed !== 'true') {
        setTourOpen(true)
      }
    }

    // In live mode, dynamically resolve the active primary repository from cluster
    if (!isDemoModeActive()) {
      getRepositories().then((repos) => {
        if (repos && repos.length > 0) {
          const first = repos[0]
          let owner = 'shiva'
          let name = first.name
          if (first.full_name && first.full_name.includes('/')) {
            const parts = first.full_name.split('/')
            owner = parts[0]
            name = parts[1]
          }
          getRepoInfo(owner, name).then((info) => {
            setCurrentRepo(info)
          })
        }
      })
    }
  }, [])

  const adminNavItems = [
    { id: 'cluster', label: 'Cluster Health & Metrics', icon: <Server className="w-4 h-4" /> },
    { id: 'users', label: 'User Directory & SCIM', icon: <Users className="w-4 h-4" /> },
    { id: 'sso', label: 'Single Sign-On (SAML/OIDC)', icon: <KeyRound className="w-4 h-4" /> },
    { id: 'audit', label: 'Audit Log & Security', icon: <ShieldAlert className="w-4 h-4" /> },
  ] as const

  const { newShell } = useUIStore()
  const [isShellParamActive, setIsShellParamActive] = useState(() => {
    if (initialShell !== undefined) return initialShell
    if (typeof window !== 'undefined') {
      const p = new URLSearchParams(window.location.search)
      return p.get('shell') === '1' || p.get('shell') === 'true'
    }
    return false
  })

  useEffect(() => {
    if (initialShell !== undefined) {
      setIsShellParamActive(initialShell)
    }
  }, [initialShell])

  const isShellActive = newShell || isShellParamActive

  const shellNav = (() => {
    if (activeTab === 'home') return 'inbox'
    if (activeTab === 'agents') return 'runs'
    if (activeTab === 'prs') return 'reviews'
    if (activeTab === 'rulesets') return 'policies'
    return 'repos'
  })()

  const handleShellNavChange = (nav: string) => {
    if (nav === 'inbox') setActiveTab('home')
    else if (nav === 'runs') setActiveTab('agents')
    else if (nav === 'reviews') setActiveTab('prs')
    else if (nav === 'policies') setActiveTab('rulesets')
    else if (nav === 'repos') setActiveTab('code')
  }

  const content = (
    <div className={cn("bg-background text-foreground flex flex-col font-sans selection:bg-primary/25 selection:text-primary", isShellActive ? "min-h-full" : "min-h-screen")}>
      {/* Hosted demo banner — only when the server reports REFWEIR_DEMO_MODE=1 */}
      {systemInfo.demoMode ? <DemoBanner repo={systemInfo.demoRepo} /> : null}

      {/* Top Application Header & UnderlineNav */}
      <Header
        repo={currentRepo}
        activeTab={activeTab}
        suppressCommandPalette={isShellActive}
        onTabChange={(tab) => {
          if (tab === 'admin') {
            setActiveTab('admin')
          } else {
            setActiveTab(tab as any)
          }
        }}
        onOpenTour={() => setTourOpen(true)}
        onOpenManual={() => setManualOpen(true)}
      />

      {/* Main Content Area */}
      <main className="flex-1 max-w-[1440px] w-full mx-auto px-5 lg:px-8 py-6">
        <ViewErrorBoundary>
          {/* 0. Enterprise Home Activity Feed */}
        {activeTab === 'home' && (
          <GlobalHomeView
            onSelectRepo={(repoSlug) => handleSelectRepo(repoSlug)}
            onSelectTab={(tab) => setActiveTab(tab as any)}
            onOpenTour={() => setTourOpen(true)}
            onOpenManual={() => setManualOpen(true)}
          />
        )}

        {/* 0.1 User Profile & Keys Management */}
        {activeTab === 'profile' && (
          <Suspense fallback={<ViewFallback />}>
            <UserProfileView
              onSelectRepo={(repoSlug) => handleSelectRepo(repoSlug)}
            />
          </Suspense>
        )}

        {/* 1. Code & File Tree View */}
        {activeTab === 'code' && (
          <RepoOverview
            repo={currentRepo}
            onSelectCommits={() => setActiveTab('commits')}
            onSelectAgentConsole={() => setActiveTab('agents')}
          />
        )}

        {/* 1.5. Issues View */}
        {activeTab === 'issues' && (
          <Suspense fallback={<ViewFallback />}>
            <IssuesView
              repo={currentRepo}
              onSelectPullRequest={() => setActiveTab('prs')}
            />
          </Suspense>
        )}

        {/* 2. Pull Requests View */}
        {activeTab === 'prs' && (
          <Suspense fallback={<ViewFallback />}>
            <PullRequestsView
              repo={currentRepo}
              onSelectAgentConsole={() => setActiveTab('agents')}
            />
          </Suspense>
        )}

        {/* 3. Commits History View */}
        {activeTab === 'commits' && (
          <Suspense fallback={<ViewFallback />}>
            <CommitsView repo={currentRepo} />
          </Suspense>
        )}

        {/* 3.5. Releases & Tags View */}
        {activeTab === 'releases' && (
          <Suspense fallback={<ViewFallback />}>
            <ReleasesView repo={currentRepo} />
          </Suspense>
        )}

        {/* 4. Autonomous Agent Console & MCP Telemetry */}
        {activeTab === 'agents' && (
          <Suspense fallback={<ViewFallback />}>
            <AgentConsoleView />
          </Suspense>
        )}

        {/* 5. Branch Rulesets & Policy View */}
        {activeTab === 'rulesets' && (
          <Suspense fallback={<ViewFallback />}>
            <RulesetsView repo={currentRepo} />
          </Suspense>
        )}

        {/* 6. Webhooks Delivery View */}
        {activeTab === 'webhooks' && (
          <Suspense fallback={<ViewFallback />}>
            <WebhooksView />
          </Suspense>
        )}

        {/* 6.5. Repository Settings & Collaborators */}
        {activeTab === 'settings' && (
          <Suspense fallback={<ViewFallback />}>
            <RepoSettingsView
              repo={currentRepo}
              onRepoUpdated={() => {
                getRepoInfo(currentRepo.ownerSlug, currentRepo.name).then(setCurrentRepo)
              }}
              onRepoDeleted={() => {
                handleSelectRepo('')
                setActiveTab('home')
              }}
            />
          </Suspense>
        )}

        {/* 7. Enterprise Admin & Settings View (Classic 2-column layout) */}
        {activeTab === 'admin' && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 items-start">
            {/* Left Settings Sidebar (25%) */}
            <aside className="md:col-span-1 space-y-1">
              <div className="px-3 pb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Enterprise Settings
              </div>
              <nav className="space-y-0.5">
                {adminNavItems.map((item) => (
                  <button
                    key={item.id}
                    onClick={() => setAdminSection(item.id)}
                    className={cn(
                      'w-full flex items-center gap-2.5 px-3 py-2 text-xs font-medium rounded-md transition-colors text-left cursor-pointer',
                      adminSection === item.id
                        ? 'bg-secondary text-foreground font-semibold border-l-2 border-primary'
                        : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
                    )}
                  >
                    <span className={cn('w-4 h-4 shrink-0', adminSection === item.id ? 'text-primary' : 'text-muted-foreground')}>
                      {item.icon}
                    </span>
                    <span>{item.label}</span>
                  </button>
                ))}
              </nav>
            </aside>

            {/* Right Settings Content (75%) */}
            <section className="md:col-span-3">
              <Suspense fallback={<ViewFallback />}>
                {adminSection === 'cluster' && <ClusterHealth />}
                {adminSection === 'users' && <EnterpriseUsers />}
                {adminSection === 'sso' && <SsoConfig />}
                {adminSection === 'audit' && <AuditLogs />}
              </Suspense>
            </section>
          </div>
        )}
        </ViewErrorBoundary>
      </main>

      {/* GitHub Primer Footer */}
      <footer className="border-t border-border bg-card/60 py-5 mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-muted-foreground font-mono">
          <div className="flex items-center gap-3">
            <span className="flex items-center gap-1.5 text-foreground">
              <span className="relative flex h-2 w-2">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-400" />
              </span>
              <span>refweird (127.0.0.1:8080)</span>
            </span>
            <span>•</span>
            <span className="text-foreground flex items-center gap-1">
              <Activity className="w-3.5 h-3.5 text-primary" />
              <span>SQLite WAL</span>
            </span>
            <span>•</span>
            <span className="text-foreground">{systemInfo.mcpMountedToolsCount ?? '—'} MCP Tools Mounted</span>
          </div>

          <div className="flex items-center gap-4">
            <a
              href="https://ravaniroshan.github.io/refweir/docs/"
              target="_blank"
              rel="noreferrer"
              className="hover:text-primary transition-colors hover:underline"
            >
              Docs & Guide
            </a>
            <a
              href="https://github.com/RavaniRoshan/refweir"
              target="_blank"
              rel="noreferrer"
              className="hover:text-primary transition-colors hover:underline"
            >
              GitHub
            </a>
            <span className="text-border">|</span>
            <span className="text-foreground font-semibold">
              refweir.dev {systemInfo.version ?? ''}
            </span>
          </div>
        </div>
      </footer>
    </div>
  )

  return (
    <>
      {isShellActive ? (
        <AppShell
          activeNav={shellNav}
          onNavChange={handleShellNavChange}
          onTabChange={(tab) => setActiveTab(tab as any)}
        >
          {content}
        </AppShell>
      ) : (
        content
      )}

      {/* Interactive First-Time Onboarding Tour */}
      {tourOpen && (
        <ViewErrorBoundary fallbackTitle="Failed to load onboarding tour">
          <Suspense fallback={null}>
            <ProductTour
              isOpen={tourOpen}
              onClose={() => setTourOpen(false)}
              onOpenManual={() => setManualOpen(true)}
            />
          </Suspense>
        </ViewErrorBoundary>
      )}

      {/* In-App System Architecture User Manual */}
      {manualOpen && (
        <ViewErrorBoundary fallbackTitle="Failed to load architecture manual">
          <Suspense fallback={null}>
            <UserManualModal
              isOpen={manualOpen}
              onClose={() => setManualOpen(false)}
            />
          </Suspense>
        </ViewErrorBoundary>
      )}
    </>
  )
}
