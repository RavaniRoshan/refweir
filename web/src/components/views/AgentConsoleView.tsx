import React, { useState, useEffect, useRef, useMemo } from 'react'
import {
  Bot,
  Terminal,
  Activity,
  Play,
  Clock,
  ShieldCheck,
  ShieldAlert,
  ChevronRight,
  KeyRound,
  Copy,
  Check,
  RefreshCw,
  Lock,
  Search,
  CheckCircle2,
  Code2,
  GitPullRequest,
  Network,
  Shield,
  GitBranch,
} from 'lucide-react'
import { Box, BoxHeader, BoxBody } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Input } from '@/components/ui/Input'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { AgentThinking } from '@/components/agent/agent-thinking'
import {
  getAgentActivities,
  issueAgentLease,
  revokeAgentLease,
  queryLocalMcpTools,
  isDemoModeActive,
  sampleAgentActivities,
  type AgentActivityItem,
  type AgentLease,
} from '@/lib/api'
import { subscribeAgentActivity } from '@/lib/events'
import { useClusterInfoQuery } from '@/lib/queries'

export interface ManagedLeaseRecord extends AgentLease {
  revoked?: boolean
  refPrefixes?: string[]
}

const INITIAL_MOCK_LEASES: ManagedLeaseRecord[] = [
  {
    id: 'lease-act-001',
    token: 'agl_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.worker1.sig',
    agentId: 'agent:refweir-autonomous-worker',
    repoId: 'acme/rocket',
    refPrefixes: ['refs/heads/agent/'],
    allowedRefs: ['refs/heads/agent/*'],
    expiresAt: new Date(Date.now() + 1800000).toISOString(),
    issuedAt: new Date(Date.now() - 300000).toISOString(),
  },
  {
    id: 'lease-act-002',
    token: 'agl_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.reviewer.sig',
    agentId: 'agent:security-reviewer-bot',
    repoId: 'acme/rocket',
    refPrefixes: ['refs/heads/agent/review/'],
    allowedRefs: ['refs/heads/agent/review/*'],
    expiresAt: new Date(Date.now() + 45000).toISOString(), // expiring soon
    issuedAt: new Date(Date.now() - 1750000).toISOString(),
  },
  {
    id: 'lease-act-003',
    token: 'agl_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.legacy.sig',
    agentId: 'agent:triage-daemon',
    repoId: 'acme/core-api',
    refPrefixes: ['refs/heads/agent/triage/'],
    allowedRefs: ['refs/heads/agent/triage/*'],
    expiresAt: new Date(Date.now() - 3600000).toISOString(), // expired
    issuedAt: new Date(Date.now() - 7200000).toISOString(),
  },
]

export const AgentConsoleView: React.FC = () => {
  const { data: clusterInfo } = useClusterInfoQuery()
  const [activities, setActivities] = useState<AgentActivityItem[]>([])
  const [loadingActivities, setLoadingActivities] = useState(true)
  const [availableTools, setAvailableTools] = useState<any[]>([])
  const [selectedToolName, setSelectedToolName] = useState('list_pull_requests')
  const [toolArgsInput, setToolArgsInput] = useState(
    '{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "state": "open"\n}'
  )
  const [executing, setExecuting] = useState(false)
  const [executionResult, setExecutionResult] = useState<any | null>(null)
  const [executionTime, setExecutionTime] = useState<number | null>(null)
  const [activeTab, setActiveTab] = useState<'activity' | 'tester' | 'catalog' | 'leases' | 'topology'>('activity')
  const [selectedTopologyNode, setSelectedTopologyNode] = useState<{
    id: string
    type: 'agent' | 'lease' | 'gate' | 'mainline'
    title: string
    details: string
    status: string
    allowedRefs?: string[]
    token?: string
  } | null>({
    id: 'lease-act-001',
    type: 'lease',
    title: 'Lease: refs/heads/agent/*',
    details: 'Scoped HMAC capability lease issued for autonomous worker. Enforces strict isolation to agent/* branches.',
    status: 'active',
    allowedRefs: ['refs/heads/agent/*'],
    token: 'agl_eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.worker1.sig',
  })

  // Timeline Filtering State
  const [activitySearchQuery, setActivitySearchQuery] = useState('')
  const [activityCategoryFilter, setActivityCategoryFilter] = useState<'all' | 'pulls' | 'symbols' | 'reviews'>('all')

  // Lease Management State
  const [leases, setLeases] = useState<ManagedLeaseRecord[]>(INITIAL_MOCK_LEASES)
  const [leaseSearchQuery, setLeaseSearchQuery] = useState('')
  const [leaseStatusFilter, setLeaseStatusFilter] = useState<'all' | 'active' | 'expiring' | 'expired' | 'revoked'>('all')
  const [leaseAgentId, setLeaseAgentId] = useState('agent:refweir-worker')
  const [leaseRepoId, setLeaseRepoId] = useState('acme/rocket')
  const [leaseRefPrefix, setLeaseRefPrefix] = useState('refs/heads/agent/')
  const [leaseTtlMinutes, setLeaseTtlMinutes] = useState(30)
  const [generatingLease, setGeneratingLease] = useState(false)
  const [issuedLease, setIssuedLease] = useState<ManagedLeaseRecord | null>(null)
  const [leaseSecondsRemaining, setLeaseSecondsRemaining] = useState<number>(0)
  const [isRevokingLease, setIsRevokingLease] = useState(false)
  const [copiedToken, setCopiedToken] = useState(false)
  const [copiedRowTokenId, setCopiedRowTokenId] = useState<string | null>(null)

  const loadActivities = async () => {
    try {
      const items = await getAgentActivities()
      setActivities(items)
    } catch {
      if (isDemoModeActive()) {
        setActivities(sampleAgentActivities)
      } else {
        setActivities([])
      }
    } finally {
      setLoadingActivities(false)
    }
  }

  const execTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    let active = true
    let unsubscribe: (() => void) | null = null
    let fallbackInterval: ReturnType<typeof setInterval> | null = null
    let demoInterval: ReturnType<typeof setInterval> | null = null

    function cleanup() {
      if (unsubscribe) {
        unsubscribe()
        unsubscribe = null
      }
      if (fallbackInterval) {
        clearInterval(fallbackInterval)
        fallbackInterval = null
      }
      if (demoInterval) {
        clearInterval(demoInterval)
        demoInterval = null
      }
    }

    function initTransport() {
      cleanup()
      if (!active) return

      if (isDemoModeActive()) {
        loadActivities()
        demoInterval = setInterval(() => {
          if (active) loadActivities()
        }, 4000)
        return
      }

      setLoadingActivities(true)
      unsubscribe = subscribeAgentActivity({
        onSnapshot: (items) => {
          if (!active) return
          setActivities(items)
          setLoadingActivities(false)
          if (fallbackInterval) {
            clearInterval(fallbackInterval)
            fallbackInterval = null
          }
        },
        onActivity: (item) => {
          if (!active) return
          setActivities((prev) => {
            const exists = prev.some((p) => p.id === item.id)
            if (exists) return prev
            return [item, ...prev]
          })
          setLoadingActivities(false)
        },
        onError: () => {
          if (!active) return
          loadActivities()
          if (!fallbackInterval) {
            fallbackInterval = setInterval(() => {
              if (active) loadActivities()
            }, 5000)
          }
        },
      })
    }

    initTransport()

    const onModeChange = () => {
      initTransport()
    }
    window.addEventListener('refweir_mode_changed', onModeChange)

    return () => {
      active = false
      cleanup()
      window.removeEventListener('refweir_mode_changed', onModeChange)
      if (execTimeoutRef.current) clearTimeout(execTimeoutRef.current)
      if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current)
    }
  }, [])

  useEffect(() => {
    let isCurrent = true
    async function loadMcpCatalog() {
      const tools = await queryLocalMcpTools()
      if (!isCurrent) return
      if (tools.length > 0) {
        setAvailableTools(tools)
      } else {
        setAvailableTools([
          { name: 'list_pull_requests', description: 'List pull requests in repository' },
          { name: 'get_pull_request', description: 'Get details for a pull request' },
          { name: 'merge_pull_request', description: 'Merge a pull request into its base branch' },
          { name: 'search_symbols', description: 'Semantic code symbol search across AST' },
          { name: 'create_pull_request_inline_comment', description: 'Start review thread on file line' },
          { name: 'checks_read', description: 'Read check runs and suites for a commit SHA' },
          { name: 'list_rulesets', description: 'List rulesets configured on a repository' },
          { name: 'list_namespaces', description: 'List all namespaces in refweir.dev' },
        ])
      }
    }
    loadMcpCatalog()
    return () => {
      isCurrent = false
    }
  }, [])

  const handleToolSelect = (toolName: string) => {
    setSelectedToolName(toolName)
    setExecutionResult(null)
    setExecutionTime(null)

    if (toolName === 'list_pull_requests') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "state": "open"\n}')
    } else if (toolName === 'get_pull_request') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "number": 104\n}')
    } else if (toolName === 'merge_pull_request') {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket",\n  "number": 104,\n  "method": "squash"\n}')
    } else if (toolName === 'search_symbols') {
      setToolArgsInput('{\n  "query": "Lease"\n}')
    } else {
      setToolArgsInput('{\n  "ownerSlug": "acme",\n  "repoName": "rocket"\n}')
    }
  }

  const handleExecuteTool = async () => {
    setExecuting(true)
    setExecutionResult(null)
    const startTime = performance.now()

    try {
      let parsedArgs = {}
      try {
        parsedArgs = JSON.parse(toolArgsInput)
      } catch (e: any) {
        throw new Error('Invalid JSON arguments: ' + e.message)
      }

      const response = await fetch('/mcp', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: Date.now(),
          method: 'tools/call',
          params: {
            name: selectedToolName,
            arguments: parsedArgs,
          },
        }),
      })

      const elapsed = Math.round(performance.now() - startTime)
      setExecutionTime(elapsed)

      if (!response.ok) {
        throw new Error(`Server returned HTTP ${response.status}`)
      }

      const resData = await response.json()
      setExecutionResult(resData)

      if (execTimeoutRef.current) clearTimeout(execTimeoutRef.current)
      execTimeoutRef.current = setTimeout(loadActivities, 500)
    } catch (err: any) {
      setExecutionResult({ error: err.message || 'Execution failed' })
    } finally {
      setExecuting(false)
    }
  }

  const handleGenerateLease = async () => {
    setGeneratingLease(true)
    try {
      const lease = await issueAgentLease(
        leaseAgentId,
        leaseRepoId,
        [leaseRefPrefix, 'refs/heads/agent/*'],
        leaseTtlMinutes * 60
      )
      const newRecord: ManagedLeaseRecord = {
        ...lease,
        refPrefixes: [leaseRefPrefix],
        allowedRefs: lease.allowedRefs || [leaseRefPrefix, 'refs/heads/agent/*'],
        issuedAt: new Date().toISOString(),
      }
      setIssuedLease(newRecord)
      setLeases((prev) => [newRecord, ...prev])
    } catch (err: any) {
      alert('Failed to issue capability lease: ' + err.message)
    } finally {
      setGeneratingLease(false)
    }
  }

  const handleCopyToken = () => {
    if (!issuedLease) return
    navigator.clipboard.writeText(issuedLease.token)
    setCopiedToken(true)
    if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current)
    copyTimeoutRef.current = setTimeout(() => setCopiedToken(false), 2000)
  }

  const handleCopyRowToken = (leaseId: string, token: string) => {
    navigator.clipboard.writeText(token)
    setCopiedRowTokenId(leaseId)
    setTimeout(() => setCopiedRowTokenId(null), 2000)
  }

  useEffect(() => {
    if (!issuedLease || issuedLease.revoked) return
    const calcRemaining = () => {
      const exp = new Date(issuedLease.expiresAt).getTime()
      return Math.max(0, Math.floor((exp - Date.now()) / 1000))
    }
    setLeaseSecondsRemaining(calcRemaining())
    const interval = setInterval(() => {
      const rem = calcRemaining()
      setLeaseSecondsRemaining(rem)
      if (rem <= 0) clearInterval(interval)
    }, 1000)
    return () => clearInterval(interval)
  }, [issuedLease])

  const formatCountdown = (seconds: number) => {
    if (seconds <= 0) return 'Expired'
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const handleRevokeLease = async (targetLeaseId?: string) => {
    const idToRevoke = targetLeaseId || issuedLease?.id
    if (!idToRevoke) return

    setIsRevokingLease(true)
    try {
      await revokeAgentLease(idToRevoke, 'operator manual revocation')
      setLeases((prev) =>
        prev.map((l) => (l.id === idToRevoke ? { ...l, revoked: true } : l))
      )
      if (issuedLease && issuedLease.id === idToRevoke) {
        setIssuedLease((prev) => (prev ? { ...prev, revoked: true } : null))
      }
    } catch (err: any) {
      alert('Failed to revoke capability lease: ' + err.message)
    } finally {
      setIsRevokingLease(false)
    }
  }

  // Filtered Activities
  const filteredActivities = useMemo(() => {
    return activities.filter((act) => {
      const q = activitySearchQuery.toLowerCase()
      const matchesSearch =
        !q ||
        act.agentName.toLowerCase().includes(q) ||
        act.toolName.toLowerCase().includes(q) ||
        (act.responseSummary && act.responseSummary.toLowerCase().includes(q))

      if (!matchesSearch) return false

      if (activityCategoryFilter === 'pulls') {
        return act.toolName.includes('pull_request')
      }
      if (activityCategoryFilter === 'symbols') {
        return act.toolName.includes('symbols') || act.toolName.includes('ast')
      }
      if (activityCategoryFilter === 'reviews') {
        return act.toolName.includes('review') || act.toolName.includes('comment')
      }
      return true
    })
  }, [activities, activitySearchQuery, activityCategoryFilter])

  // Lease Status Calculator
  const getLeaseLifecycleState = (lease: ManagedLeaseRecord) => {
    if (lease.revoked) return { label: 'Revoked', variant: 'outline' as const, color: 'text-destructive border-destructive/40' }
    const now = Date.now()
    const exp = new Date(lease.expiresAt).getTime()
    const diffSec = Math.floor((exp - now) / 1000)
    if (diffSec <= 0) return { label: 'Expired', variant: 'outline' as const, color: 'text-muted-foreground' }
    if (diffSec < 300) return { label: `Expiring Soon (${formatCountdown(diffSec)})`, variant: 'outline' as const, color: 'text-amber-500 border-amber-500/40' }
    return { label: `Active (${formatCountdown(diffSec)})`, variant: 'success' as const, color: 'text-emerald-500' }
  }

  // Filtered Leases
  const filteredLeases = useMemo(() => {
    return leases.filter((lease) => {
      const q = leaseSearchQuery.toLowerCase()
      const matchesSearch =
        !q ||
        lease.agentId.toLowerCase().includes(q) ||
        lease.repoId.toLowerCase().includes(q) ||
        lease.id.toLowerCase().includes(q)

      if (!matchesSearch) return false

      const state = getLeaseLifecycleState(lease)
      if (leaseStatusFilter === 'active' && !state.label.startsWith('Active')) return false
      if (leaseStatusFilter === 'expiring' && !state.label.startsWith('Expiring')) return false
      if (leaseStatusFilter === 'expired' && state.label !== 'Expired') return false
      if (leaseStatusFilter === 'revoked' && state.label !== 'Revoked') return false

      return true
    })
  }, [leases, leaseSearchQuery, leaseStatusFilter])

  return (
    <div className="space-y-4">
      {/* Top Banner Hero */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <div className="flex items-center gap-2">
            <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
              <Bot className="w-4 h-4 text-primary" />
              <span>AI Agent Telemetry & Capability Leases</span>
            </h2>
            <Badge variant="agent" pulse>
              Live Protocol
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground">
            Direct Model Context Protocol server mounted on <code className="text-primary font-mono">http://localhost:8080/mcp</code> with {clusterInfo?.mcpMountedToolsCount ?? (availableTools.length > 0 ? availableTools.length : 55)} active tools.
          </p>
        </div>

        {/* Quick Metrics */}
        <div className="flex items-center gap-3 text-xs font-mono">
          <div className="px-2.5 py-1 rounded-md bg-secondary border border-border">
            <span className="text-muted-foreground mr-1">Tools:</span>
            <span className="font-semibold text-foreground">{clusterInfo?.mcpMountedToolsCount ?? (availableTools.length > 0 ? availableTools.length : 55)}</span>
          </div>
          <div className="px-2.5 py-1 rounded-md bg-secondary border border-border">
            <span className="text-muted-foreground mr-1">Avg Latency:</span>
            <span className="font-semibold text-success">12ms</span>
          </div>
        </div>
      </div>

      {/* Sub-Navigation UnderlineNav */}
      <div className="flex items-center gap-2 border-b border-border -mb-px text-xs font-medium">
        <button
          onClick={() => setActiveTab('activity')}
          className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors cursor-pointer ${
            activeTab === 'activity'
              ? 'border-primary text-foreground font-semibold'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <Activity className="w-3.5 h-3.5" />
          <span>Real-time Activity</span>
          <span className="text-[10px] bg-secondary text-secondary-foreground px-1.5 py-0.2 rounded-full font-mono">
            {activities.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab('tester')}
          className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors cursor-pointer ${
            activeTab === 'tester'
              ? 'border-primary text-foreground font-semibold'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <Terminal className="w-3.5 h-3.5" />
          <span>Live MCP Tool Runner</span>
        </button>

        <button
          onClick={() => setActiveTab('leases')}
          className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors cursor-pointer ${
            activeTab === 'leases'
              ? 'border-primary text-foreground font-semibold'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <KeyRound className="w-3.5 h-3.5" />
          <span>Capability Leases</span>
          <span className="text-[10px] bg-secondary text-secondary-foreground px-1.5 py-0.2 rounded-full font-mono">
            {leases.length}
          </span>
        </button>

        <button
          onClick={() => setActiveTab('topology')}
          className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors cursor-pointer ${
            activeTab === 'topology'
              ? 'border-primary text-foreground font-semibold'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <Network className="w-3.5 h-3.5 text-primary" />
          <span>Workflow Topology</span>
          <span className="text-[10px] bg-primary/10 text-primary border border-primary/30 px-1.5 py-0.2 rounded-full font-mono">
            Graph
          </span>
        </button>

        <button
          onClick={() => setActiveTab('catalog')}
          className={`flex items-center gap-2 px-3 py-2 border-b-2 transition-colors cursor-pointer ${
            activeTab === 'catalog'
              ? 'border-primary text-foreground font-semibold'
              : 'border-transparent text-muted-foreground hover:text-foreground'
          }`}
        >
          <ShieldCheck className="w-3.5 h-3.5" />
          <span>Tool Catalog</span>
        </button>
      </div>

      {/* Tab 1: Real-time Activity Feed & Operational Timeline */}
      {activeTab === 'activity' && (
        <Box>
          <BoxHeader className="bg-secondary/35 py-2.5 px-4 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div className="flex items-center gap-2">
              <span className="font-semibold text-foreground text-xs">Autonomous Agent Operational Timeline</span>
              <Badge variant="outline" className="text-[10px] font-mono">
                {filteredActivities.length} of {activities.length} events
              </Badge>
            </div>

            {/* Timeline Filter Controls */}
            <div className="flex flex-wrap items-center gap-2 self-start sm:self-auto">
              <div className="relative w-44">
                <Search className="w-3 h-3 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                <Input
                  type="text"
                  placeholder="Filter events..."
                  value={activitySearchQuery}
                  onChange={(e) => setActivitySearchQuery(e.target.value)}
                  className="h-7 text-[11px] pl-7 pr-2"
                />
              </div>

              <div className="flex items-center gap-1 bg-muted/60 p-0.5 rounded-lg border border-border text-[11px]">
                <button
                  type="button"
                  onClick={() => setActivityCategoryFilter('all')}
                  className={`px-2 py-0.5 rounded transition-colors ${
                    activityCategoryFilter === 'all'
                      ? 'bg-background text-foreground font-semibold shadow-xs'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  All
                </button>
                <button
                  type="button"
                  onClick={() => setActivityCategoryFilter('pulls')}
                  className={`px-2 py-0.5 rounded transition-colors flex items-center gap-1 ${
                    activityCategoryFilter === 'pulls'
                      ? 'bg-background text-foreground font-semibold shadow-xs'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <GitPullRequest className="w-2.5 h-2.5" />
                  <span>Pulls</span>
                </button>
                <button
                  type="button"
                  onClick={() => setActivityCategoryFilter('symbols')}
                  className={`px-2 py-0.5 rounded transition-colors flex items-center gap-1 ${
                    activityCategoryFilter === 'symbols'
                      ? 'bg-background text-foreground font-semibold shadow-xs'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <Code2 className="w-2.5 h-2.5" />
                  <span>Symbols</span>
                </button>
                <button
                  type="button"
                  onClick={() => setActivityCategoryFilter('reviews')}
                  className={`px-2 py-0.5 rounded transition-colors flex items-center gap-1 ${
                    activityCategoryFilter === 'reviews'
                      ? 'bg-background text-foreground font-semibold shadow-xs'
                      : 'text-muted-foreground hover:text-foreground'
                  }`}
                >
                  <CheckCircle2 className="w-2.5 h-2.5" />
                  <span>Reviews</span>
                </button>
              </div>
            </div>
          </BoxHeader>

          <BoxBody className="p-0">
            {loadingActivities && activities.length === 0 ? (
              <div className="flex items-center justify-center p-8">
                <AgentThinking variant="infinity" label="Collecting agent telemetry" showTimer={false} />
              </div>
            ) : filteredActivities.length === 0 ? (
              <div className="p-8 text-center space-y-3">
                <div className="w-10 h-10 rounded-md bg-secondary border border-border flex items-center justify-center mx-auto text-primary">
                  <Activity className="w-5 h-5" />
                </div>
                <div className="max-w-md mx-auto space-y-1">
                  <h3 className="text-sm font-semibold text-foreground">
                    {activities.length === 0 ? 'No agent operations recorded yet' : 'No events match your filter criteria'}
                  </h3>
                  <p className="text-xs text-muted-foreground leading-relaxed">
                    {activities.length === 0
                      ? 'Autonomous agents invoking tools via the JSON-RPC Model Context Protocol or executing pull request reviews will stream real-time operational telemetry into this dashboard.'
                      : 'Try resetting the search query or category filters to view all telemetry events.'}
                  </p>
                </div>
                {activities.length > 0 ? (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setActivitySearchQuery('')
                      setActivityCategoryFilter('all')
                    }}
                    className="text-xs"
                  >
                    Reset Filters
                  </Button>
                ) : (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setActiveTab('tester')}
                    className="gap-1.5 text-xs"
                  >
                    <Terminal className="w-3.5 h-3.5 text-primary" />
                    <span>Test an MCP Tool Now</span>
                  </Button>
                )}
              </div>
            ) : (
              <div className="p-4 relative before:absolute before:top-6 before:bottom-6 before:left-8 before:w-0.5 before:bg-border/60 space-y-4">
                {filteredActivities.map((act) => (
                  <div key={act.id} className="relative pl-10 group">
                    {/* Visual Timeline Marker Node */}
                    <div className="absolute left-2.5 top-1.5 -translate-x-1/2 w-6 h-6 rounded-full bg-card border-2 border-primary/50 group-hover:border-primary flex items-center justify-center shadow-xs transition-colors z-10">
                      <div className="w-2 h-2 rounded-full bg-primary" />
                    </div>

                    <div className="p-3.5 rounded-lg border border-border bg-card/60 hover:bg-secondary/20 transition-all space-y-2.5">
                      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                        <div className="flex items-center gap-2.5">
                          <LivingAvatar name={act.agentName} isAgent={true} size="sm" />
                          <div className="flex flex-wrap items-center gap-1.5 text-xs">
                            <span className="font-semibold text-foreground">{act.agentName}</span>
                            <span className="text-muted-foreground font-mono text-[11px]">executed</span>
                            <span className="font-mono text-xs text-primary bg-secondary/80 px-2 py-0.5 rounded border border-border">
                              {act.toolName}
                            </span>
                            <Badge variant={act.status === 'error' ? 'outline' : 'success'} className={`text-[9px] py-0 ${act.status === 'error' ? 'text-destructive border-destructive/40' : ''}`}>
                              {act.status || 'success'}
                            </Badge>
                          </div>
                        </div>

                        <div className="flex items-center gap-3 text-xs font-mono self-start sm:self-auto">
                          <span className={`flex items-center gap-1 text-[11px] ${
                            (act.latencyMs || 0) < 50 ? 'text-success' : (act.latencyMs || 0) < 200 ? 'text-blue-500' : 'text-amber-500'
                          }`}>
                            <Clock className="w-3 h-3" /> {act.latencyMs}ms
                          </span>
                          <span className="text-muted-foreground text-[11px]">{act.timestamp}</span>
                        </div>
                      </div>

                      {/* Request and Response Summary Cards */}
                      <div className="text-xs font-mono text-muted-foreground bg-secondary/30 p-2.5 rounded-md border border-border space-y-1">
                        <div className="text-[11px] flex items-baseline gap-2">
                          <span className="text-muted-foreground shrink-0 font-medium">Payload:</span>
                          <span className="text-foreground truncate break-all">{JSON.stringify(act.requestPayload)}</span>
                        </div>
                        <div className="text-[11px] flex items-baseline gap-2">
                          <span className="text-muted-foreground shrink-0 font-medium">Result:</span>
                          <span className="text-foreground">{act.responseSummary}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </BoxBody>
        </Box>
      )}

      {/* Tab 2: Live Tool Runner */}
      {activeTab === 'tester' && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          <Box className="p-3 space-y-2">
            <h3 className="text-xs font-semibold text-foreground">Select MCP Tool</h3>
            <div className="space-y-1 max-h-96 overflow-y-auto pr-1">
              {availableTools.map((t) => (
                <button
                  key={t.name}
                  onClick={() => handleToolSelect(t.name)}
                  className={`w-full text-left px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors flex items-center justify-between ${
                    selectedToolName === t.name
                      ? 'bg-secondary text-foreground font-semibold border border-primary/40'
                      : 'hover:bg-secondary/40 text-muted-foreground'
                  }`}
                >
                  <span className="truncate">{t.name}</span>
                  <ChevronRight className="w-3 h-3 opacity-50 shrink-0" />
                </button>
              ))}
            </div>
          </Box>

          <Box className="lg:col-span-2">
            <BoxHeader className="bg-secondary/35 py-2">
              <div>
                <h3 className="text-xs font-semibold text-foreground font-mono flex items-center gap-1.5">
                  <Terminal className="w-3.5 h-3.5 text-primary" />
                  <span>{selectedToolName}</span>
                </h3>
              </div>

              <Button
                variant="primary"
                size="sm"
                onClick={handleExecuteTool}
                disabled={executing}
                className="gap-1.5 text-xs"
              >
                {executing ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Play className="w-3.5 h-3.5" />}
                <span>{executing ? 'Executing...' : 'Run Tool'}</span>
              </Button>
            </BoxHeader>

            <BoxBody className="p-3.5 space-y-3">
              <div className="space-y-1">
                <label className="text-xs font-medium text-foreground">Tool Arguments (JSON)</label>
                <textarea
                  value={toolArgsInput}
                  onChange={(e) => setToolArgsInput(e.target.value)}
                  rows={6}
                  className="w-full bg-input border border-border rounded-md p-2.5 font-mono text-xs text-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                />
              </div>

              {executing && (
                <div className="flex items-center gap-2 rounded-md border border-border bg-secondary/30 px-3 py-2">
                  <AgentThinking variant="wave" label={`Invoking ${selectedToolName}`} />
                </div>
              )}

              {executionResult && (
                <div className="space-y-1.5 pt-2 border-t border-border">
                  <div className="flex items-center justify-between text-xs">
                    <span className="font-semibold text-foreground">Execution Response</span>
                    {executionTime && (
                      <span className="text-success font-mono">{executionTime}ms</span>
                    )}
                  </div>
                  <pre className="p-2.5 bg-background border border-border rounded-md font-mono text-xs text-foreground overflow-x-auto max-h-60 leading-relaxed">
                    {JSON.stringify(executionResult, null, 2)}
                  </pre>
                </div>
              )}
            </BoxBody>
          </Box>
        </div>
      )}

      {/* Tab 3: Capability Leases & Lifecycle Management */}
      {activeTab === 'leases' && (
        <div className="space-y-4">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Generator Box */}
            <Box>
              <BoxHeader className="bg-secondary/35 py-2.5">
                <div className="flex items-center gap-2">
                  <KeyRound className="w-4 h-4 text-primary" />
                  <span className="text-xs font-semibold text-foreground">Issue Agent Capability Lease</span>
                </div>
              </BoxHeader>

              <BoxBody className="p-4 space-y-3">
                <p className="text-xs text-muted-foreground leading-relaxed">
                  Generates a cryptographically signed HMAC-SHA256 capability token that permits an AI agent to push
                  exclusively to authorized ref prefixes (e.g. <code className="font-mono text-foreground">refs/heads/agent/*</code>)
                  with a guaranteed expiration TTL.
                </p>

                <div className="space-y-2.5 pt-1">
                  <div className="space-y-1">
                    <label className="text-xs font-medium text-foreground">Agent Identifier</label>
                    <input
                      type="text"
                      value={leaseAgentId}
                      onChange={(e) => setLeaseAgentId(e.target.value)}
                      className="w-full bg-input border border-border rounded-md p-2 text-xs text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="text-xs font-medium text-foreground">Target Repository</label>
                    <input
                      type="text"
                      value={leaseRepoId}
                      onChange={(e) => setLeaseRepoId(e.target.value)}
                      className="w-full bg-input border border-border rounded-md p-2 text-xs text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="text-xs font-medium text-foreground">Allowed Ref Prefix</label>
                    <input
                      type="text"
                      value={leaseRefPrefix}
                      onChange={(e) => setLeaseRefPrefix(e.target.value)}
                      className="w-full bg-input border border-border rounded-md p-2 text-xs text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="text-xs font-medium text-foreground">TTL Duration (Minutes)</label>
                    <input
                      type="number"
                      value={leaseTtlMinutes}
                      onChange={(e) => setLeaseTtlMinutes(parseInt(e.target.value, 10) || 15)}
                      className="w-full bg-input border border-border rounded-md p-2 text-xs text-foreground font-mono outline-none focus:border-ring focus:ring-1 focus:ring-ring"
                    />
                  </div>

                  <div className="pt-2">
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={handleGenerateLease}
                      disabled={generatingLease}
                      className="w-full gap-1.5 text-xs"
                    >
                      <Lock className="w-3.5 h-3.5" />
                      <span>{generatingLease ? 'Signing Lease...' : 'Sign & Issue Capability Lease'}</span>
                    </Button>
                  </div>
                </div>
              </BoxBody>
            </Box>

            {/* Credential Inspector Box */}
            <Box>
              <BoxHeader className="bg-secondary/35 py-2.5">
                <span className="text-xs font-semibold text-foreground">Active Lease Credential</span>
              </BoxHeader>

              <BoxBody className="p-4 space-y-3">
                {issuedLease ? (
                  <div className="space-y-3">
                    <div className="p-2.5 bg-secondary/40 border border-border rounded-md space-y-1.5">
                      <div className="flex items-center justify-between text-xs font-mono text-foreground">
                        <span>Lease ID: {issuedLease.id}</span>
                        {issuedLease.revoked ? (
                          <Badge variant="outline" className="text-[10px] text-destructive border-destructive/40">
                            Revoked
                          </Badge>
                        ) : leaseSecondsRemaining <= 0 ? (
                          <Badge variant="outline" className="text-[10px] text-muted-foreground">
                            Expired
                          </Badge>
                        ) : leaseSecondsRemaining < 60 ? (
                          <Badge variant="outline" className="text-[10px] text-amber-500 border-amber-500/40">
                            Expiring soon · {formatCountdown(leaseSecondsRemaining)}
                          </Badge>
                        ) : (
                          <Badge variant="success" className="text-[10px]">
                            Active · {formatCountdown(leaseSecondsRemaining)}
                          </Badge>
                        )}
                      </div>
                      <div className="text-[11px] text-muted-foreground font-mono">
                        Expires: {new Date(issuedLease.expiresAt).toLocaleTimeString()} ({issuedLease.expiresAt})
                      </div>
                    </div>

                    {/* Authorized Scope Chips */}
                    <div className="space-y-1.5">
                      <span className="text-[11px] font-medium text-foreground">Authorized Ref Scopes:</span>
                      <div className="flex flex-wrap gap-1.5 text-[10px] font-mono">
                        <span className="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                          ALLOW: {issuedLease.allowedRefs?.join(', ') || issuedLease.refPrefixes?.join(', ') || 'refs/heads/agent/*'}
                        </span>
                        <span className="px-2 py-0.5 rounded bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20">
                          DENY: refs/heads/main, refs/heads/release/*
                        </span>
                      </div>
                    </div>

                    <div className="space-y-1">
                      <div className="flex items-center justify-between text-xs text-muted-foreground">
                        <span>Signed Capability Token</span>
                        <button
                          onClick={handleCopyToken}
                          className="hover:text-foreground transition-colors flex items-center gap-1 text-[11px] text-primary cursor-pointer"
                        >
                          {copiedToken ? <Check className="w-3 h-3 text-success" /> : <Copy className="w-3 h-3" />}
                          <span>{copiedToken ? 'Copied' : 'Copy Token'}</span>
                        </button>
                      </div>
                      <pre className="p-2.5 bg-background border border-border rounded-md text-xs font-mono text-foreground overflow-x-auto break-all whitespace-pre-wrap">
                        {issuedLease.token}
                      </pre>
                    </div>

                    {!issuedLease.revoked && leaseSecondsRemaining > 0 && (
                      <div className="pt-1 flex items-center justify-between">
                        <Button
                          variant="ghost"
                          size="xs"
                          onClick={() => handleRevokeLease()}
                          disabled={isRevokingLease}
                          className="text-destructive hover:bg-destructive/10 text-xs gap-1"
                        >
                          <ShieldAlert className="w-3.5 h-3.5" />
                          <span>{isRevokingLease ? 'Revoking...' : 'Revoke Capability Lease'}</span>
                        </Button>
                      </div>
                    )}

                    <div className="p-2.5 bg-secondary/30 border border-border rounded-md space-y-1 text-xs text-muted-foreground">
                      <span className="font-semibold text-foreground">How to use:</span>
                      <p className="text-[11px] leading-relaxed">
                        Pass in agent HTTP requests via header:
                        <br />
                        <code className="text-primary font-mono">Authorization: Bearer {issuedLease.token.slice(0, 16)}...</code>
                      </p>
                    </div>
                  </div>
                ) : (
                  <div className="p-8 text-center text-muted-foreground text-xs font-mono">
                    No active capability lease issued in this session. Configure the parameters and click "Sign & Issue" to generate one.
                  </div>
                )}
              </BoxBody>
            </Box>
          </div>

          {/* Searchable Capability Leases Table */}
          <Box>
            <BoxHeader className="bg-secondary/35 py-2.5 px-4 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <KeyRound className="w-4 h-4 text-primary" />
                <span className="text-xs font-semibold text-foreground">Issued Agent Capability Leases & Audited State</span>
                <Badge variant="outline" className="text-[10px] font-mono">
                  {filteredLeases.length} total
                </Badge>
              </div>

              <div className="flex flex-wrap items-center gap-2 self-start sm:self-auto">
                <div className="relative w-48">
                  <Search className="w-3 h-3 absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    type="text"
                    placeholder="Search by agent or repo..."
                    value={leaseSearchQuery}
                    onChange={(e) => setLeaseSearchQuery(e.target.value)}
                    className="h-7 text-[11px] pl-7 pr-2 font-mono"
                  />
                </div>

                <div className="flex items-center gap-1 bg-muted/60 p-0.5 rounded-lg border border-border text-[11px]">
                  {(['all', 'active', 'expiring', 'expired', 'revoked'] as const).map((filter) => (
                    <button
                      key={filter}
                      type="button"
                      onClick={() => setLeaseStatusFilter(filter)}
                      className={`px-2 py-0.5 rounded capitalize transition-colors ${
                        leaseStatusFilter === filter
                          ? 'bg-background text-foreground font-semibold shadow-xs'
                          : 'text-muted-foreground hover:text-foreground'
                      }`}
                    >
                      {filter}
                    </button>
                  ))}
                </div>
              </div>
            </BoxHeader>

            <BoxBody className="p-0 overflow-x-auto">
              <table className="w-full text-left text-xs border-collapse">
                <thead>
                  <tr className="border-b border-border bg-secondary/20 text-[11px] font-medium text-muted-foreground">
                    <th className="py-2.5 px-3">Agent Identifier</th>
                    <th className="py-2.5 px-3">Target Repo</th>
                    <th className="py-2.5 px-3">Allowed Ref Prefix</th>
                    <th className="py-2.5 px-3">Lifecycle State</th>
                    <th className="py-2.5 px-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {filteredLeases.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="py-8 text-center text-muted-foreground text-xs font-mono">
                        No capability leases match the current filter.
                      </td>
                    </tr>
                  ) : (
                    filteredLeases.map((l) => {
                      const state = getLeaseLifecycleState(l)
                      return (
                        <tr key={l.id} className="hover:bg-secondary/10 transition-colors">
                          <td className="py-2.5 px-3 font-mono font-medium text-foreground">
                            <div className="flex items-center gap-2">
                              <LivingAvatar name={l.agentId} isAgent={true} size="sm" />
                              <span>{l.agentId}</span>
                            </div>
                          </td>
                          <td className="py-2.5 px-3 font-mono text-muted-foreground">{l.repoId}</td>
                          <td className="py-2.5 px-3 font-mono text-primary text-[11px]">
                            {l.refPrefixes?.join(', ') || l.allowedRefs?.join(', ') || 'refs/heads/agent/*'}
                          </td>
                          <td className="py-2.5 px-3">
                            <Badge variant={state.variant} className={`text-[10px] ${state.color}`}>
                              {state.label}
                            </Badge>
                          </td>
                          <td className="py-2.5 px-3 text-right">
                            <div className="flex items-center justify-end gap-1.5">
                              <button
                                onClick={() => handleCopyRowToken(l.id, l.token)}
                                className="p-1 rounded hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors"
                                title="Copy Capability Token"
                              >
                                {copiedRowTokenId === l.id ? (
                                  <Check className="w-3.5 h-3.5 text-success" />
                                ) : (
                                  <Copy className="w-3.5 h-3.5" />
                                )}
                              </button>
                              {!l.revoked && (
                                <button
                                  onClick={() => handleRevokeLease(l.id)}
                                  className="p-1 rounded hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                                  title="Revoke Capability Lease"
                                >
                                  <ShieldAlert className="w-3.5 h-3.5" />
                                </button>
                              )}
                            </div>
                          </td>
                        </tr>
                      )
                    })
                  )}
                </tbody>
              </table>
            </BoxBody>
          </Box>
        </div>
      )}

      {/* Tab 4: 50-Tool Catalog */}
      {activeTab === 'catalog' && (
        <Box>
          <BoxHeader className="bg-secondary/35 py-2.5">
            <span className="text-xs font-semibold text-foreground">Model Context Protocol Tools</span>
          </BoxHeader>

          <BoxBody className="p-0">
            <div className="grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-border">
              <div className="p-3.5 space-y-2">
                <h4 className="text-xs font-semibold text-foreground uppercase tracking-wider">Repository & Tree Tools</h4>
                <div className="space-y-1.5 text-xs">
                  {availableTools.slice(0, Math.ceil(availableTools.length / 2)).map((t) => (
                    <div key={t.name} className="p-2 rounded-md bg-secondary/30 border border-border space-y-0.5">
                      <span className="font-mono font-semibold text-foreground text-xs">{t.name}</span>
                      <p className="text-muted-foreground text-[11px]">{t.description}</p>
                    </div>
                  ))}
                </div>
              </div>

              <div className="p-3.5 space-y-2">
                <h4 className="text-xs font-semibold text-foreground uppercase tracking-wider">Policy & Review Tools</h4>
                <div className="space-y-1.5 text-xs">
                  {availableTools.slice(Math.ceil(availableTools.length / 2)).map((t) => (
                    <div key={t.name} className="p-2 rounded-md bg-secondary/30 border border-border space-y-0.5">
                      <span className="font-mono font-semibold text-foreground text-xs">{t.name}</span>
                      <p className="text-muted-foreground text-[11px]">{t.description}</p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </BoxBody>
        </Box>
      )}

      {/* Tab 5: Interactive Capability & Policy Workflow Topology */}
      {activeTab === 'topology' && (
        <div className="space-y-4">
          {/* Header & Metrics Strip */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            <div className="p-3 rounded-lg border border-border bg-card space-y-1">
              <div className="flex items-center justify-between text-muted-foreground text-xs">
                <span>Active Leases</span>
                <KeyRound className="w-3.5 h-3.5 text-primary" />
              </div>
              <div className="text-xl font-bold text-foreground font-mono">
                {leases.filter((l) => !l.revoked && new Date(l.expiresAt).getTime() > Date.now()).length}
              </div>
              <div className="text-[11px] text-muted-foreground">Ephemeral HMAC capability tokens</div>
            </div>

            <div className="p-3 rounded-lg border border-border bg-card space-y-1">
              <div className="flex items-center justify-between text-muted-foreground text-xs">
                <span>Ref Containment</span>
                <ShieldCheck className="w-3.5 h-3.5 text-emerald-500" />
              </div>
              <div className="text-xl font-bold text-emerald-500 font-mono">100%</div>
              <div className="text-[11px] text-muted-foreground">Restricted to refs/heads/agent/*</div>
            </div>

            <div className="p-3 rounded-lg border border-border bg-card space-y-1">
              <div className="flex items-center justify-between text-muted-foreground text-xs">
                <span>Policy Gates</span>
                <Lock className="w-3.5 h-3.5 text-amber-500" />
              </div>
              <div className="text-xl font-bold text-foreground font-mono">3 Active</div>
              <div className="text-[11px] text-muted-foreground">Rulesets, Signatures, AST Review</div>
            </div>

            <div className="p-3 rounded-lg border border-border bg-card space-y-1">
              <div className="flex items-center justify-between text-muted-foreground text-xs">
                <span>Direct Main Pushes</span>
                <ShieldAlert className="w-3.5 h-3.5 text-destructive" />
              </div>
              <div className="text-xl font-bold text-foreground font-mono">0 Allowed</div>
              <div className="text-[11px] text-muted-foreground">Hermetic forge-level quarantine</div>
            </div>
          </div>

          {/* Interactive Topology Graph Box */}
          <Box>
            <BoxHeader className="bg-secondary/35 py-2.5 px-4 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div className="flex items-center gap-2">
                <Network className="w-4 h-4 text-primary" />
                <span className="font-semibold text-foreground text-xs">Interactive Capability & Governance Topology</span>
              </div>
              <div className="flex items-center gap-2 text-xs">
                <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-500 border border-emerald-500/30 text-[10px] font-medium">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                  Weir Containment Enforced
                </span>
              </div>
            </BoxHeader>

            <BoxBody className="p-4 sm:p-6 space-y-6">
              {/* Visual 4-Tier Flow Canvas */}
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4 relative">
                {/* Tier 1: Agents */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between border-b border-border pb-1.5 text-xs font-semibold text-foreground">
                    <span className="flex items-center gap-1.5">
                      <Bot className="w-3.5 h-3.5 text-primary" />
                      <span>1. Agent Fleet</span>
                    </span>
                    <span className="text-[10px] text-muted-foreground font-mono">Identity</span>
                  </div>

                  <div className="space-y-2">
                    {[
                      { id: 'agent-1', name: 'agent:autonomous-worker', role: 'Feature Implementation', active: true },
                      { id: 'agent-2', name: 'agent:security-reviewer-bot', role: 'Policy & AST Inspection', active: true },
                      { id: 'agent-3', name: 'agent:triage-daemon', role: 'Background Indexing', active: false },
                    ].map((agent) => (
                      <button
                        key={agent.id}
                        type="button"
                        onClick={() =>
                          setSelectedTopologyNode({
                            id: agent.id,
                            type: 'agent',
                            title: agent.name,
                            details: `Agent principal actively authenticated via MCP. Designated role: ${agent.role}.`,
                            status: agent.active ? 'active' : 'idle',
                          })
                        }
                        className={`w-full text-left p-3 rounded-lg border transition-all cursor-pointer ${
                          selectedTopologyNode?.title === agent.name
                            ? 'border-primary bg-primary/10 shadow-xs'
                            : 'border-border bg-secondary/30 hover:bg-secondary/60'
                        }`}
                      >
                        <div className="flex items-center justify-between gap-1">
                          <span className="font-mono text-xs font-semibold text-foreground truncate">{agent.name}</span>
                          <span
                            className={`w-2 h-2 rounded-full shrink-0 ${
                              agent.active ? 'bg-emerald-500 animate-pulse' : 'bg-muted-foreground'
                            }`}
                          />
                        </div>
                        <div className="text-[11px] text-muted-foreground mt-0.5">{agent.role}</div>
                      </button>
                    ))}
                  </div>
                </div>

                {/* Tier 2: Capability Weir Leases */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between border-b border-border pb-1.5 text-xs font-semibold text-foreground">
                    <span className="flex items-center gap-1.5">
                      <KeyRound className="w-3.5 h-3.5 text-primary" />
                      <span>2. Capability Weir</span>
                    </span>
                    <span className="text-[10px] text-muted-foreground font-mono">HMAC Isolation</span>
                  </div>

                  <div className="space-y-2">
                    {leases.slice(0, 3).map((l) => {
                      const state = getLeaseLifecycleState(l)
                      return (
                        <button
                          key={l.id}
                          type="button"
                          onClick={() =>
                            setSelectedTopologyNode({
                              id: l.id,
                              type: 'lease',
                              title: `Lease (${l.id})`,
                              details: `Scoped capability token for ${l.agentId}. Allowed branch scope: ${(l.refPrefixes || l.allowedRefs || []).join(', ')}.`,
                              status: state.label,
                              allowedRefs: l.refPrefixes || l.allowedRefs,
                              token: l.token,
                            })
                          }
                          className={`w-full text-left p-3 rounded-lg border transition-all cursor-pointer ${
                            selectedTopologyNode?.id === l.id
                              ? 'border-primary bg-primary/10 shadow-xs'
                              : 'border-border bg-secondary/30 hover:bg-secondary/60'
                          }`}
                        >
                          <div className="flex items-center justify-between gap-1">
                            <span className="font-mono text-[11px] text-foreground font-semibold truncate">{l.id}</span>
                            <span className={`text-[10px] px-1.5 py-0.2 rounded font-medium border ${state.color}`}>
                              {state.label.split(' ')[0]}
                            </span>
                          </div>
                          <div className="text-[10px] text-primary font-mono mt-1 truncate">
                            {(l.refPrefixes || l.allowedRefs || []).join(', ')}
                          </div>
                        </button>
                      )
                    })}
                  </div>
                </div>

                {/* Tier 3: Governance Policy Gates */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between border-b border-border pb-1.5 text-xs font-semibold text-foreground">
                    <span className="flex items-center gap-1.5">
                      <Shield className="w-3.5 h-3.5 text-amber-500" />
                      <span>3. Policy Gates</span>
                    </span>
                    <span className="text-[10px] text-muted-foreground font-mono">Verification</span>
                  </div>

                  <div className="space-y-2">
                    {[
                      {
                        id: 'gate-ruleset',
                        name: 'Branch Ruleset Gate',
                        details: 'Requires pull request review, verified commit status checks, and non-linear push rejection.',
                        icon: ShieldCheck,
                      },
                      {
                        id: 'gate-sig',
                        name: 'OpenSSH Signature Attestation',
                        details: 'Cryptographically audits SSH SIGSPEC and PGP RFC 4880 signatures before merge execution.',
                        icon: Lock,
                      },
                      {
                        id: 'gate-ast',
                        name: 'Refweir Reviewer (AST & Leaks)',
                        details: 'Automated agent reviewer inspects Go/TS/Python AST diffs and flags exposed credentials.',
                        icon: CheckCircle2,
                      },
                    ].map((gate) => {
                      const Icon = gate.icon
                      return (
                        <button
                          key={gate.id}
                          type="button"
                          onClick={() =>
                            setSelectedTopologyNode({
                              id: gate.id,
                              type: 'gate',
                              title: gate.name,
                              details: gate.details,
                              status: 'enforced',
                            })
                          }
                          className={`w-full text-left p-3 rounded-lg border transition-all cursor-pointer ${
                            selectedTopologyNode?.id === gate.id
                              ? 'border-primary bg-primary/10 shadow-xs'
                              : 'border-border bg-secondary/30 hover:bg-secondary/60'
                          }`}
                        >
                          <div className="flex items-center gap-2">
                            <Icon className="w-3.5 h-3.5 text-amber-500 shrink-0" />
                            <span className="text-xs font-semibold text-foreground truncate">{gate.name}</span>
                          </div>
                          <div className="text-[10px] text-muted-foreground mt-1 line-clamp-1">{gate.details}</div>
                        </button>
                      )
                    })}
                  </div>
                </div>

                {/* Tier 4: Target Mainline */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between border-b border-border pb-1.5 text-xs font-semibold text-foreground">
                    <span className="flex items-center gap-1.5">
                      <GitBranch className="w-3.5 h-3.5 text-emerald-500" />
                      <span>4. Mainline Branch</span>
                    </span>
                    <span className="text-[10px] text-muted-foreground font-mono">Protected</span>
                  </div>

                  <div className="space-y-2">
                    <button
                      type="button"
                      onClick={() =>
                        setSelectedTopologyNode({
                          id: 'node-main',
                          type: 'mainline',
                          title: 'refs/heads/main',
                          details: 'Target mainline production branch. Direct pushes strictly prohibited for all agents without verified pull request and passing policy gates.',
                          status: 'protected',
                          allowedRefs: ['refs/heads/main'],
                        })
                      }
                      className={`w-full text-left p-4 rounded-lg border transition-all cursor-pointer ${
                        selectedTopologyNode?.id === 'node-main'
                          ? 'border-emerald-500 bg-emerald-500/10 shadow-xs'
                          : 'border-border bg-secondary/30 hover:bg-secondary/60'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-mono text-xs font-bold text-foreground">refs/heads/main</span>
                        <Badge variant="success" className="text-[10px]">Protected</Badge>
                      </div>
                      <p className="text-[11px] text-muted-foreground mt-1.5 leading-relaxed">
                        Merge transactions executed synchronously with 3-way or squash strategy once policy gates clear.
                      </p>
                    </button>

                    {/* Direct Push Prohibition Card */}
                    <div className="p-3 rounded-lg border border-destructive/30 bg-destructive/5 text-destructive text-[11px] space-y-1">
                      <div className="flex items-center gap-1.5 font-semibold">
                        <ShieldAlert className="w-3.5 h-3.5 shrink-0" />
                        <span>Direct Pushes Blocked</span>
                      </div>
                      <p className="text-[10px] leading-relaxed text-destructive/80">
                        Pre-receive hook rejects any agent token attempting push to non-prefixed refs.
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Node Inspector Panel */}
              {selectedTopologyNode && (
                <div className="mt-4 p-4 rounded-lg border border-border/80 bg-secondary/20 space-y-3">
                  <div className="flex items-center justify-between flex-wrap gap-2">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-semibold text-foreground uppercase tracking-wider">
                        Topology Inspector:
                      </span>
                      <span className="font-mono text-xs font-bold text-primary">{selectedTopologyNode.title}</span>
                      <Badge variant="outline" className="text-[10px] capitalize">
                        {selectedTopologyNode.status}
                      </Badge>
                    </div>

                    {selectedTopologyNode.type === 'lease' && selectedTopologyNode.id.startsWith('lease-') && (
                      <Button
                        size="xs"
                        variant="destructive"
                        onClick={() => handleRevokeLease(selectedTopologyNode.id)}
                        disabled={isRevokingLease}
                        className="text-xs gap-1"
                      >
                        <ShieldAlert className="w-3 h-3" />
                        <span>Revoke Lease</span>
                      </Button>
                    )}
                  </div>

                  <p className="text-xs text-foreground/90 leading-relaxed">{selectedTopologyNode.details}</p>

                  {selectedTopologyNode.token && (
                    <div className="space-y-1">
                      <span className="text-[10px] font-semibold text-muted-foreground uppercase">Token Digest</span>
                      <div className="p-2 rounded bg-zinc-950 text-zinc-300 font-mono text-[11px] flex items-center justify-between gap-2 overflow-x-auto border border-border/40">
                        <span className="truncate">{selectedTopologyNode.token}</span>
                        <button
                          type="button"
                          onClick={() => handleCopyRowToken(selectedTopologyNode.id, selectedTopologyNode.token!)}
                          className="shrink-0 p-1 hover:text-white transition-colors cursor-pointer"
                          title="Copy Token"
                        >
                          {copiedRowTokenId === selectedTopologyNode.id ? (
                            <Check className="w-3.5 h-3.5 text-emerald-400" />
                          ) : (
                            <Copy className="w-3.5 h-3.5" />
                          )}
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              )}
            </BoxBody>
          </Box>
        </div>
      )}
    </div>
  )
}
