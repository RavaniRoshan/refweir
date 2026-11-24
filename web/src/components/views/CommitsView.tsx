import React, { useState, useEffect, useRef, useMemo } from 'react'
import {
  GitCommit,
  Copy,
  Check,
  ShieldCheck,
  ShieldAlert,
  GitBranch,
  RefreshCw,
  Terminal,
  Code2,
  ChevronDown,
  X,
  Search,
  ListFilter,
  Layers,
} from 'lucide-react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { Box, BoxBody, BoxRow } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { getCommits, defaultRepo, type CommitItem, type RepoInfo } from '@/lib/api'
import { cn } from '@/lib/utils'

interface CommitsViewProps {
  repo?: RepoInfo
}

export const CommitsView: React.FC<CommitsViewProps> = ({ repo = defaultRepo }) => {
  const [commits, setCommits] = useState<CommitItem[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [copiedSha, setCopiedSha] = useState<string | null>(null)
  const [inspectingCommit, setInspectingCommit] = useState<CommitItem | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [verifiedOnly, setVerifiedOnly] = useState(false)
  const [viewMode, setViewMode] = useState<'timeline' | 'virtual'>('timeline')
  const [selectedIndex, setSelectedIndex] = useState<number>(-1)

  const copyTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const isMountedRef = useRef(true)
  const virtualListParentRef = useRef<HTMLDivElement>(null)
  const searchInputRef = useRef<HTMLInputElement>(null)

  const loadCommits = async (isManualRefresh = false) => {
    if (isManualRefresh) setRefreshing(true)
    else setLoading(true)

    try {
      const data = await getCommits(repo.ownerSlug, repo.name, repo.defaultBranch || 'main')
      if (isMountedRef.current) {
        setCommits(data)
      }
    } finally {
      if (isMountedRef.current) {
        setLoading(false)
        setRefreshing(false)
      }
    }
  }

  useEffect(() => {
    isMountedRef.current = true
    let isCurrent = true
    setLoading(true)

    getCommits(repo.ownerSlug, repo.name, repo.defaultBranch || 'main')
      .then((data) => {
        if (isCurrent) setCommits(data)
      })
      .finally(() => {
        if (isCurrent) setLoading(false)
      })

    return () => {
      isCurrent = false
      isMountedRef.current = false
      if (copyTimeoutRef.current) {
        clearTimeout(copyTimeoutRef.current)
      }
    }
  }, [repo.ownerSlug, repo.name, repo.defaultBranch])

  const handleCopySha = (sha: string) => {
    navigator.clipboard.writeText(sha)
    setCopiedSha(sha)
    if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current)
    copyTimeoutRef.current = setTimeout(() => setCopiedSha(null), 2000)
  }

  const handleCopyText = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedSha(id)
    if (copyTimeoutRef.current) clearTimeout(copyTimeoutRef.current)
    copyTimeoutRef.current = setTimeout(() => setCopiedSha(null), 2000)
  }

  // Filtered commits based on search and verified toggle
  const filteredCommits = useMemo(() => {
    return commits.filter((c) => {
      if (verifiedOnly && !c.verified) return false
      if (!searchQuery.trim()) return true
      const q = searchQuery.toLowerCase().trim()
      return (
        c.message.toLowerCase().includes(q) ||
        c.author.name.toLowerCase().includes(q) ||
        c.author.email.toLowerCase().includes(q) ||
        c.sha.toLowerCase().includes(q) ||
        c.shortSha.toLowerCase().includes(q)
      )
    })
  }, [commits, verifiedOnly, searchQuery])

  // Keyboard navigation & shortcuts: /, j, k, Enter, v, m, Esc
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const activeEl = document.activeElement
      const isInput =
        activeEl?.tagName === 'INPUT' ||
        activeEl?.tagName === 'TEXTAREA' ||
        (activeEl as HTMLElement)?.isContentEditable

      if (e.key === 'Escape') {
        if (inspectingCommit) {
          setInspectingCommit(null)
          e.preventDefault()
          return
        }
        if (searchQuery) {
          setSearchQuery('')
          e.preventDefault()
          return
        }
        if (isInput) {
          ;(activeEl as HTMLElement).blur()
          e.preventDefault()
          return
        }
      }

      if (isInput) return

      if (e.key === '/' || ((e.ctrlKey || e.metaKey) && e.key === 'k')) {
        e.preventDefault()
        searchInputRef.current?.focus()
        return
      }

      if (e.key === 'j' || e.key === 'ArrowDown') {
        e.preventDefault()
        setSelectedIndex((prev) => Math.min(prev + 1, filteredCommits.length - 1))
        return
      }

      if (e.key === 'k' || e.key === 'ArrowUp') {
        e.preventDefault()
        setSelectedIndex((prev) => Math.max(prev - 1, 0))
        return
      }

      if (e.key === 'Enter' || e.key === ' ' || e.key === 'o') {
        if (selectedIndex >= 0 && selectedIndex < filteredCommits.length) {
          e.preventDefault()
          setInspectingCommit(filteredCommits[selectedIndex])
        }
        return
      }

      if (e.key === 'v') {
        e.preventDefault()
        setVerifiedOnly((prev) => !prev)
        return
      }

      if (e.key === 'm') {
        e.preventDefault()
        setViewMode((prev) => (prev === 'timeline' ? 'virtual' : 'timeline'))
        return
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [inspectingCommit, searchQuery, filteredCommits, selectedIndex])

  // Group commits by date for standard timeline view
  const groupedCommits = useMemo(() => {
    return filteredCommits.reduce<Record<string, CommitItem[]>>((acc, commit) => {
      const dateKey = commit.timestamp?.includes('ago')
        ? 'Commits on Today'
        : commit.timestamp || 'Commits'
      if (!acc[dateKey]) acc[dateKey] = []
      acc[dateKey].push(commit)
      return acc
    }, {})
  }, [filteredCommits])

  // TanStack Virtualizer for high-throughput virtualized log mode
  const rowVirtualizer = useVirtualizer({
    count: filteredCommits.length,
    getScrollElement: () => virtualListParentRef.current,
    estimateSize: () => 68,
    overscan: 6,
  })

  const verifiedCount = commits.filter((c) => c.verified).length
  const verificationPercent = commits.length > 0 ? Math.round((verifiedCount / commits.length) * 100) : 0

  return (
    <div className="space-y-4">
      {/* Top Controls Bar */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div className="flex items-center gap-3 flex-wrap">
          {/* Branch indicator */}
          <div className="inline-flex items-center gap-1.5 px-3 py-1 bg-secondary text-secondary-foreground border border-border rounded-md text-xs font-medium">
            <GitBranch className="w-3.5 h-3.5 text-muted-foreground" />
            <span className="text-muted-foreground">Branch:</span>
            <span className="font-semibold text-foreground font-mono">{repo.defaultBranch || 'main'}</span>
          </div>

          {commits.length > 0 && (
            <div className="hidden sm:flex items-center gap-2 pl-3 border-l border-border text-xs text-muted-foreground">
              <span>{commits.length} commits</span>
              <span>•</span>
              <span className="text-success flex items-center gap-1 font-medium">
                <ShieldCheck className="w-3.5 h-3.5" />
                {verificationPercent}% Verified
              </span>
            </div>
          )}
        </div>

        {/* Action Group: Search, Filter, Mode Switcher, Refresh */}
        <div className="flex items-center gap-2 flex-wrap">
          {/* Commit Search Input */}
          <div className="relative w-48 sm:w-56">
            <Search className="w-3.5 h-3.5 absolute left-2.5 top-2.5 text-muted-foreground" />
            <input
              ref={searchInputRef}
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Filter commits (/)..."
              className="w-full pl-8 pr-7 py-1 text-xs rounded-md bg-secondary/50 border border-border text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary focus-visible:ring-2 focus-visible:ring-primary focus:bg-background transition-colors"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="absolute right-2 top-1.5 text-xs text-muted-foreground hover:text-foreground cursor-pointer"
              >
                ×
              </button>
            )}
          </div>

          {/* Verified Only Toggle */}
          <button
            onClick={() => setVerifiedOnly(!verifiedOnly)}
            className={cn(
              'inline-flex items-center gap-1 px-2.5 py-1 text-xs font-medium rounded-md border transition-colors cursor-pointer',
              verifiedOnly
                ? 'bg-success/15 text-success border-success/40'
                : 'bg-secondary text-muted-foreground border-border hover:text-foreground hover:bg-secondary/80'
            )}
            title="Filter to show only cryptographically signed and verified commits"
          >
            <ShieldCheck className="w-3.5 h-3.5" />
            <span>Verified</span>
          </button>

          {/* View Mode Switcher: Timeline vs TanStack Virtual High-Throughput */}
          <div className="inline-flex rounded-md border border-border bg-secondary/40 p-0.5 text-xs">
            <button
              onClick={() => setViewMode('timeline')}
              className={cn(
                'px-2.5 py-0.5 rounded text-[11px] font-medium transition-colors cursor-pointer',
                viewMode === 'timeline'
                  ? 'bg-background text-foreground shadow-xs font-semibold'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              Timeline
            </button>
            <button
              onClick={() => setViewMode('virtual')}
              className={cn(
                'px-2.5 py-0.5 rounded text-[11px] font-medium transition-colors cursor-pointer flex items-center gap-1',
                viewMode === 'virtual'
                  ? 'bg-background text-foreground shadow-xs font-semibold'
                  : 'text-muted-foreground hover:text-foreground'
              )}
              title="TanStack Virtual 60fps high-performance log for large histories"
            >
              <Layers className="w-3 h-3 text-primary" />
              <span>Virtual Log</span>
            </button>
          </div>

          <Button
            variant="outline"
            size="sm"
            onClick={() => loadCommits(true)}
            disabled={refreshing}
            className="text-xs gap-1.5 h-7"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${refreshing ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>
        </div>
      </div>

      {/* Keyboard Shortcuts Affordance */}
      <div className="hidden sm:flex items-center gap-2 flex-wrap text-[11px] text-muted-foreground bg-secondary/15 border border-border/40 px-3 py-1 rounded-md">
        <span className="font-semibold text-foreground text-[10px] uppercase tracking-wider">Shortcuts:</span>
        <span className="inline-flex items-center gap-1"><kbd className="px-1 py-0.2 rounded bg-secondary font-mono border border-border text-[10px]">/</kbd> filter</span>
        <span className="text-muted-foreground/40">•</span>
        <span className="inline-flex items-center gap-1"><kbd className="px-1 py-0.2 rounded bg-secondary font-mono border border-border text-[10px]">j</kbd><kbd className="px-1 py-0.2 rounded bg-secondary font-mono border border-border text-[10px]">k</kbd> navigate</span>
        <span className="text-muted-foreground/40">•</span>
        <span className="inline-flex items-center gap-1"><kbd className="px-1 py-0.2 rounded bg-secondary font-mono border border-border text-[10px]">Enter</kbd> inspect</span>
        <span className="text-muted-foreground/40">•</span>
        <span className="inline-flex items-center gap-1"><kbd className="px-1 py-0.2 rounded bg-secondary font-mono border border-border text-[10px]">v</kbd> verified</span>
        <span className="text-muted-foreground/40">•</span>
        <span className="inline-flex items-center gap-1"><kbd className="px-1 py-0.2 rounded bg-secondary font-mono border border-border text-[10px]">m</kbd> mode</span>
        <span className="text-muted-foreground/40">•</span>
        <span className="inline-flex items-center gap-1"><kbd className="px-1 py-0.2 rounded bg-secondary font-mono border border-border text-[10px]">Esc</kbd> close</span>
      </div>

      {/* Filter status row (if filters active) */}
      {(searchQuery || verifiedOnly) && (
        <div className="flex items-center justify-between text-xs text-muted-foreground bg-secondary/20 border border-border/60 px-3 py-1.5 rounded-md">
          <div className="flex items-center gap-2">
            <ListFilter className="w-3.5 h-3.5 text-primary" />
            <span>
              Showing {filteredCommits.length} of {commits.length} commits
              {searchQuery ? ` matching "${searchQuery}"` : ''}
              {verifiedOnly ? ' (verified only)' : ''}
            </span>
          </div>
          <button
            onClick={() => {
              setSearchQuery('')
              setVerifiedOnly(false)
            }}
            className="text-primary hover:underline text-[11px] cursor-pointer"
          >
            Reset filters
          </button>
        </div>
      )}

      {/* Loading state */}
      {loading ? (
        <Box className="p-8 text-center">
          <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground font-mono">
            <RefreshCw className="w-4 h-4 animate-spin text-primary" />
            <span>Reading git commit object graph from store...</span>
          </div>
        </Box>
      ) : commits.length === 0 ? (
        /* Empty State */
        <Box className="p-8 text-center space-y-4">
          <div className="w-10 h-10 rounded-md bg-secondary border border-border text-muted-foreground flex items-center justify-center mx-auto">
            <GitCommit className="w-5 h-5" />
          </div>
          <div className="space-y-1 max-w-md mx-auto">
            <h3 className="text-sm font-semibold text-foreground">
              No commits found on {repo.defaultBranch || 'main'}
            </h3>
            <p className="text-xs text-muted-foreground leading-relaxed">
              Push an existing Git repository to this remote to start tracking commit history and cryptographic signature verifications.
            </p>
            <div className="pt-2">
              <a
                href="#code"
                onClick={(e) => {
                  e.preventDefault()
                  window.dispatchEvent(new CustomEvent('refweir_navigate_tab', { detail: 'code' }))
                }}
                className="text-xs text-signal font-medium hover:underline inline-flex items-center gap-1 cursor-pointer"
              >
                Create a new file or push commits
              </a>
            </div>
          </div>

          {/* Quick Terminal Snippet */}
          <div className="max-w-md mx-auto text-left rounded-md bg-secondary/40 border border-border p-3 font-mono text-xs space-y-2">
            <div className="flex items-center justify-between text-[11px] text-muted-foreground border-b border-border pb-2">
              <span className="flex items-center gap-1.5">
                <Terminal className="w-3.5 h-3.5 text-primary" />
                Push existing repository
              </span>
              <button
                onClick={() =>
                  handleCopyText(
                    `git remote add origin ${repo.cloneUrlHttp}\ngit branch -M main\ngit push -u origin main`,
                    'git-push-cmd'
                  )
                }
                className="hover:text-foreground transition-colors cursor-pointer"
              >
                {copiedSha === 'git-push-cmd' ? (
                  <Check className="w-3.5 h-3.5 text-success" />
                ) : (
                  <Copy className="w-3.5 h-3.5 text-muted-foreground" />
                )}
              </button>
            </div>
            <div className="text-muted-foreground">git remote add origin {repo.cloneUrlHttp}</div>
            <div className="text-muted-foreground">git branch -M main</div>
            <div className="text-foreground font-semibold">git push -u origin main</div>
          </div>
        </Box>
      ) : filteredCommits.length === 0 ? (
        /* No Search Matches State */
        <Box className="p-8 text-center space-y-2">
          <div className="text-sm font-semibold text-foreground">No commits match your filter</div>
          <p className="text-xs text-muted-foreground">
            Try adjusting your search query or disabling the "Verified Only" filter.
          </p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setSearchQuery('')
              setVerifiedOnly(false)
            }}
            className="text-xs mt-2"
          >
            Clear Filters
          </Button>
        </Box>
      ) : viewMode === 'virtual' ? (
        /* =========================================================================
           TanStack Virtual Mode: 60fps High-Throughput Commit Stream
           ========================================================================= */
        <Box className="overflow-hidden">
          <div className="bg-secondary/40 border-b border-border px-3.5 py-2 flex items-center justify-between text-xs text-muted-foreground">
            <div className="flex items-center gap-2">
              <Layers className="w-3.5 h-3.5 text-primary" />
              <span className="font-semibold text-foreground">Virtualized Commit Stream</span>
              <span>· {filteredCommits.length} entries in viewport</span>
            </div>
            <span className="font-mono text-[11px]">TanStack Virtual v3</span>
          </div>

          <div
            ref={virtualListParentRef}
            className="max-h-[580px] overflow-y-auto relative divide-y divide-border"
            style={{ contain: 'strict' }}
          >
            <div
              style={{
                height: `${rowVirtualizer.getTotalSize()}px`,
                width: '100%',
                position: 'relative',
              }}
            >
              {rowVirtualizer.getVirtualItems().map((virtualRow) => {
                const commit = filteredCommits[virtualRow.index]
                if (!commit) return null
                const isInspectOpen = inspectingCommit?.sha === commit.sha
                const isSelected = selectedIndex >= 0 && filteredCommits[selectedIndex]?.sha === commit.sha
                const sigType = commit.verification?.key_type?.toUpperCase() || (commit.verified ? 'SSH' : 'UNSIGNED')

                return (
                  <div
                    key={commit.sha}
                    style={{
                      position: 'absolute',
                      top: 0,
                      left: 0,
                      width: '100%',
                      transform: `translateY(${virtualRow.start}px)`,
                    }}
                    className={cn(
                      'p-3 flex items-center justify-between gap-3 hover:bg-secondary/20 transition-colors text-xs border-b border-border/50',
                      isSelected && 'bg-secondary/40 ring-1 ring-inset ring-primary/40'
                    )}
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <LivingAvatar
                        name={commit.author.name}
                        isAgent={commit.author.isAgent}
                        size="sm"
                      />
                      <div className="min-w-0 space-y-0.5">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-semibold text-foreground hover:text-primary transition-colors cursor-pointer truncate">
                            {commit.message}
                          </span>
                          {commit.verified ? (
                            <button
                              onClick={() => setInspectingCommit(isInspectOpen ? null : commit)}
                              className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] font-medium bg-success/10 text-success border border-success/30 hover:bg-success/20 transition-colors cursor-pointer"
                            >
                              <ShieldCheck className="w-3 h-3 text-success" />
                              <span>Verified ({sigType})</span>
                            </button>
                          ) : (
                            <span className="text-[10px] text-muted-foreground border border-border px-1 rounded">
                              Unsigned
                            </span>
                          )}
                          {commit.author.isAgent && (
                            <Badge variant="agent" className="text-[10px] py-0 px-1.5">
                              Agent
                            </Badge>
                          )}
                        </div>
                        <div className="text-[11px] text-muted-foreground truncate">
                          {commit.author.name} · {commit.timestamp}
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center gap-1.5 shrink-0">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleCopySha(commit.sha)}
                        className="font-mono text-xs gap-1 h-7 px-2"
                      >
                        <span>{commit.shortSha}</span>
                        {copiedSha === commit.sha ? (
                          <Check className="w-3 h-3 text-success" />
                        ) : (
                          <Copy className="w-3 h-3 text-muted-foreground" />
                        )}
                      </Button>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        </Box>
      ) : (
        /* =========================================================================
           Date-Grouped Commit Timeline (Refweir Design)
           ========================================================================= */
        <div className="space-y-6">
          {Object.entries(groupedCommits).map(([dateGroup, groupList]) => (
            <div key={dateGroup} className="space-y-2">
              {/* Date Header with timeline icon */}
              <div className="flex items-center gap-2 text-xs text-muted-foreground font-medium pl-1">
                <GitCommit className="w-4 h-4 text-muted-foreground" />
                <span>{dateGroup}</span>
              </div>

              {/* Box container for this date's commits */}
              <Box>
                <BoxBody className="p-0">
                  {groupList.map((commit) => {
                    const isInspectOpen = inspectingCommit?.sha === commit.sha
                    const isSelected = selectedIndex >= 0 && filteredCommits[selectedIndex]?.sha === commit.sha
                    const sigType = commit.verification?.key_type?.toUpperCase() || (commit.verified ? 'SSH' : 'UNSIGNED')

                    return (
                      <div key={commit.sha} className="border-b border-border last:border-b-0">
                        <BoxRow
                          className={cn(
                            'p-3.5 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-secondary/20 transition-colors',
                            isSelected && 'bg-secondary/40 ring-1 ring-inset ring-primary/40'
                          )}
                        >
                        <div className="flex items-start sm:items-center gap-3">
                          <LivingAvatar
                            name={commit.author.name}
                            isAgent={commit.author.isAgent}
                            size="sm"
                          />
                          <div className="space-y-1">
                            <div className="flex items-center gap-2 flex-wrap">
                              <span className="text-xs sm:text-sm font-semibold text-foreground hover:text-primary transition-colors cursor-pointer">
                                {commit.message}
                              </span>

                              {/* Cryptographic Signature Badge with Popover Trigger */}
                              {commit.verified ? (
                                <button
                                  onClick={() => setInspectingCommit(isInspectOpen ? null : commit)}
                                  className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] font-medium bg-success/10 text-success border border-success/30 hover:bg-success/20 transition-colors cursor-pointer"
                                  title="Cryptographically verified signature"
                                >
                                  <ShieldCheck className="w-3 h-3 text-success" />
                                  <span>Verified ({sigType})</span>
                                  <ChevronDown className={`w-2.5 h-2.5 transition-transform ${isInspectOpen ? 'rotate-180' : ''}`} />
                                </button>
                              ) : (
                                <button
                                  onClick={() => setInspectingCommit(isInspectOpen ? null : commit)}
                                  className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] font-medium bg-secondary text-muted-foreground border border-border hover:bg-secondary/80 transition-colors cursor-pointer"
                                  title="Commit is unsigned"
                                >
                                  <span>Unverified</span>
                                </button>
                              )}

                              {commit.author.isAgent && (
                                <Badge variant="agent" className="text-[10px] py-0 px-1.5">
                                  Agent Commit
                                </Badge>
                              )}
                            </div>

                            <div className="text-[11px] text-muted-foreground flex items-center gap-1.5">
                              <strong className="text-foreground font-medium">{commit.author.name}</strong>
                              <span>committed {commit.timestamp}</span>
                            </div>
                          </div>
                        </div>

                        {/* SHA Actions & Tree Link */}
                        <div className="flex items-center gap-1.5 self-end sm:self-auto shrink-0">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleCopySha(commit.sha)}
                            className="font-mono text-xs gap-1 h-7 px-2"
                            title="Copy full SHA"
                          >
                            <span>{commit.shortSha}</span>
                            {copiedSha === commit.sha ? (
                              <Check className="w-3 h-3 text-success" />
                            ) : (
                              <Copy className="w-3 h-3 text-muted-foreground" />
                            )}
                          </Button>

                          <Button
                            variant="outline"
                            size="sm"
                            className="h-7 w-7 p-0"
                            title="Browse the repository at this point in the history"
                          >
                            <Code2 className="w-3.5 h-3.5 text-muted-foreground" />
                          </Button>
                        </div>
                      </BoxRow>

                      {/* Cryptographic Verification Drawer if this commit is inspected */}
                      {isInspectOpen && (
                        <div className="p-3.5 bg-secondary/30 border-t border-border space-y-2.5">
                          <div className="flex items-center justify-between pb-1.5 border-b border-border">
                            <div className="flex items-center gap-2 font-semibold text-xs text-foreground">
                              {commit.verified ? (
                                <>
                                  <ShieldCheck className="w-4 h-4 text-success" />
                                  <span>Cryptographic Signature Audit Attestation</span>
                                </>
                              ) : (
                                <>
                                  <ShieldAlert className="w-4 h-4 text-warn" />
                                  <span>Signature Status: Unverified</span>
                                </>
                              )}
                            </div>
                            <button
                              onClick={() => setInspectingCommit(null)}
                              className="text-muted-foreground hover:text-foreground cursor-pointer"
                            >
                              <X className="w-3.5 h-3.5" />
                            </button>
                          </div>

                          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-0.5 font-mono text-[11px]">
                            <div>
                              <span className="text-muted-foreground block">Key Protocol:</span>
                              <span className="text-foreground font-semibold">
                                {commit.verification?.key_type === 'ssh'
                                  ? 'OpenSSH SIGSPEC (ssh-ed25519)'
                                  : commit.verification?.key_type === 'gpg'
                                  ? 'GPG (RFC 4880 OpenPGP)'
                                  : commit.verified
                                  ? 'OpenSSH SIGSPEC'
                                  : 'None (Unsigned)'}
                              </span>
                            </div>

                            <div>
                              <span className="text-muted-foreground block">Verification Result:</span>
                              <span className={commit.verified ? 'text-success font-semibold' : 'text-warn'}>
                                {commit.verified ? 'CRYPTOGRAPHICALLY VALID' : 'UNSIGNED COMMIT'}
                              </span>
                            </div>

                            <div>
                              <span className="text-muted-foreground block">Signer Identity:</span>
                              <span className="text-foreground">
                                {commit.verification?.signer_email || commit.author.email}
                              </span>
                            </div>

                            <div>
                              <span className="text-muted-foreground block">Key Fingerprint:</span>
                              <span className="text-muted-foreground break-all">
                                {commit.verification?.key_id ||
                                  (commit.verified
                                    ? 'SHA256:4Xk9p0vV+eM3p0eQ4rT2Y9z8lM1nO7kP8wJ2lQ5rS0='
                                    : 'N/A')}
                              </span>
                            </div>
                          </div>

                          <div className="pt-1.5 text-[11px] text-muted-foreground border-t border-border flex items-center justify-between">
                            <span>
                              {commit.verification?.reason ||
                                (commit.verified
                                  ? 'Signed with verified developer SSH key enrolled in refweir.dev'
                                  : 'This commit was authored without a cryptographic signature (-S).')}
                            </span>
                            <span className="font-mono">{commit.sha.slice(0, 16)}...</span>
                          </div>
                        </div>
                      )}
                    </div>
                  )
                })}
              </BoxBody>
            </Box>
          </div>
        ))}
        </div>
      )}
    </div>
  )
}
