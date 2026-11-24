import React, { useState, useMemo, useEffect } from 'react'
import { motion, AnimatePresence } from 'motion/react'
import {
  Rocket,
  Bot,
  ShieldCheck,
  HardDrive,
  Webhook,
  Building2,
  Search,
  BookOpen,
  ChevronRight,
  ChevronDown,
  Copy,
  Check,
  ExternalLink,
  ThumbsUp,
  ThumbsDown,
  Info,
  Lightbulb,
  AlertTriangle,
  Flame,
  ArrowRight,
  ArrowLeft,
  Sparkles,
} from 'lucide-react'
import {
  DOCS_CATEGORIES,
  DOCS_ARTICLES,
  type DocArticle,
  type DocCategory,
} from './docsData'
import { Badge } from '@/components/ui/Badge'
import { cn } from '@/lib/utils'

export interface DocsViewProps {
  initialArticleId?: string
  onNavigateTab?: (tab: string) => void
}

const CATEGORY_ICONS: Record<string, React.ReactNode> = {
  Rocket: <Rocket className="w-4 h-4" />,
  Bot: <Bot className="w-4 h-4" />,
  ShieldCheck: <ShieldCheck className="w-4 h-4" />,
  HardDrive: <HardDrive className="w-4 h-4" />,
  Webhook: <Webhook className="w-4 h-4" />,
  Building2: <Building2 className="w-4 h-4" />,
}

export const DocsView: React.FC<DocsViewProps> = ({
  initialArticleId,
  onNavigateTab,
}) => {
  const [selectedArticleId, setSelectedArticleId] = useState<string>(() => {
    if (initialArticleId && DOCS_ARTICLES[initialArticleId]) {
      return initialArticleId
    }
    if (typeof window !== 'undefined') {
      const p = new URLSearchParams(window.location.search)
      const a = p.get('article')
      if (a && DOCS_ARTICLES[a]) return a
    }
    return 'quickstart'
  })

  const [expandedCategories, setExpandedCategories] = useState<Record<string, boolean>>(() => {
    const init: Record<string, boolean> = {}
    DOCS_CATEGORIES.forEach((cat) => {
      init[cat.id] = true
    })
    return init
  })

  const [searchQuery, setSearchQuery] = useState('')
  const [copiedSnippet, setCopiedSnippet] = useState<string | null>(null)
  const [activeCodeTab, setActiveCodeTab] = useState<number>(0)
  const [feedbackSubmitted, setFeedbackSubmitted] = useState<boolean>(false)
  const [activeHeadingId, setActiveHeadingId] = useState<string>('')

  // Sync article selection to URL query string
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const url = new URL(window.location.href)
      url.searchParams.set('tab', 'docs')
      url.searchParams.set('article', selectedArticleId)
      window.history.replaceState({}, '', url.toString())
    }
  }, [selectedArticleId])

  // Reset feedback state on article change
  useEffect(() => {
    setFeedbackSubmitted(false)
    setActiveCodeTab(0)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }, [selectedArticleId])

  const currentArticle: DocArticle = useMemo(() => {
    return DOCS_ARTICLES[selectedArticleId] || DOCS_ARTICLES['quickstart']
  }, [selectedArticleId])

  const currentCategory: DocCategory | undefined = useMemo(() => {
    return DOCS_CATEGORIES.find((cat) => cat.id === currentArticle.categoryId)
  }, [currentArticle])

  // Search filtering across all articles
  const searchResults = useMemo(() => {
    const q = searchQuery.trim().toLowerCase()
    if (!q) return null

    return Object.values(DOCS_ARTICLES).filter((art) => {
      return (
        art.title.toLowerCase().includes(q) ||
        art.lead.toLowerCase().includes(q) ||
        art.sections.some((s) => s.title.toLowerCase().includes(q)) ||
        art.codeSnippets?.some((cs) => cs.code.toLowerCase().includes(q))
      )
    })
  }, [searchQuery])

  const handleCopy = (code: string, id: string) => {
    navigator.clipboard.writeText(code)
    setCopiedSnippet(id)
    setTimeout(() => setCopiedSnippet(null), 2000)
  }

  const toggleCategory = (catId: string) => {
    setExpandedCategories((prev) => ({
      ...prev,
      [catId]: !prev[catId],
    }))
  }

  // Ordered list of articles for Prev / Next navigation
  const allArticleList = useMemo(() => {
    const list: DocArticle[] = []
    DOCS_CATEGORIES.forEach((cat) => {
      cat.articleIds.forEach((id) => {
        if (DOCS_ARTICLES[id]) list.push(DOCS_ARTICLES[id])
      })
    })
    return list
  }, [])

  const currentIndex = allArticleList.findIndex((a) => a.id === currentArticle.id)
  const prevArticle = currentIndex > 0 ? allArticleList[currentIndex - 1] : null
  const nextArticle = currentIndex < allArticleList.length - 1 ? allArticleList[currentIndex + 1] : null

  return (
    <div className="w-full space-y-6 pb-16">
      {/* 1. GitHub Docs Style Header & Search Bar */}
      <div className="border-b border-border/70 pb-5 pt-1">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary">
              <BookOpen className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-xl font-bold tracking-tight text-foreground">Refweir Documentation</h1>
                <Badge variant="outline" className="text-[10px] uppercase font-mono tracking-wider">
                  v0.2.0-enterprise
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground">
                Official architecture, API specification, and autonomous agent guides
              </p>
            </div>
          </div>

          {/* Quick Docs Search Input */}
          <div className="relative w-full md:w-80">
            <Search className="w-4 h-4 absolute left-3 top-2.5 text-muted-foreground" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search docs or articles... (/)"
              className="w-full pl-9 pr-9 py-1.5 text-xs rounded-md bg-secondary/50 border border-border focus:outline-none focus:ring-1 focus:ring-primary focus:bg-background transition-colors text-foreground"
            />
            {searchQuery ? (
              <button
                onClick={() => setSearchQuery('')}
                className="absolute right-2.5 top-2 text-xs text-muted-foreground hover:text-foreground"
              >
                ×
              </button>
            ) : (
              <span className="absolute right-2.5 top-2 text-[10px] text-muted-foreground border border-border px-1 rounded">
                /
              </span>
            )}
          </div>
        </div>

        {/* Breadcrumb Navigation Bar */}
        <div className="flex items-center gap-1.5 mt-4 text-xs text-muted-foreground">
          <button
            onClick={() => setSelectedArticleId('quickstart')}
            className="hover:text-foreground transition-colors cursor-pointer"
          >
            Refweir Docs
          </button>
          <ChevronRight className="w-3.5 h-3.5 text-muted-foreground/60" />
          <span className="font-medium text-muted-foreground">{currentCategory?.title || 'Guides'}</span>
          <ChevronRight className="w-3.5 h-3.5 text-muted-foreground/60" />
          <span className="font-semibold text-foreground">{currentArticle.title}</span>
        </div>
      </div>

      {/* 2. Main 3-Column Documentation Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* LEFT COLUMN (Sidebar Navigation): 3 cols */}
        <aside className="lg:col-span-3 space-y-4">
          <div className="sticky top-20 max-h-[calc(100vh-6rem)] overflow-y-auto pr-2 space-y-4 text-xs">
            {/* Search results mode */}
            {searchResults !== null ? (
              <div className="space-y-2">
                <div className="flex items-center justify-between text-[11px] font-semibold text-muted-foreground uppercase tracking-wider px-1">
                  <span>Search Results ({searchResults.length})</span>
                  <button
                    onClick={() => setSearchQuery('')}
                    className="text-primary hover:underline lowercase font-normal"
                  >
                    clear
                  </button>
                </div>
                {searchResults.length === 0 ? (
                  <div className="p-4 text-center text-muted-foreground border border-dashed border-border rounded-md">
                    No articles match "{searchQuery}"
                  </div>
                ) : (
                  <div className="space-y-1">
                    {searchResults.map((art) => (
                      <button
                        key={art.id}
                        onClick={() => {
                          setSelectedArticleId(art.id)
                          setSearchQuery('')
                        }}
                        className={cn(
                          'w-full text-left p-2 rounded-md transition-colors cursor-pointer block',
                          selectedArticleId === art.id
                            ? 'bg-primary/10 text-primary font-semibold border-l-2 border-primary'
                            : 'hover:bg-secondary/60 text-foreground'
                        )}
                      >
                        <div className="font-medium">{art.title}</div>
                        <div className="text-[11px] text-muted-foreground truncate">{art.lead}</div>
                      </button>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              /* Regular Hierarchical Navigation Tree */
              DOCS_CATEGORIES.map((cat) => {
                const isExpanded = expandedCategories[cat.id] ?? true

                return (
                  <div key={cat.id} className="border border-border/40 rounded-lg overflow-hidden bg-card/20">
                    <button
                      onClick={() => toggleCategory(cat.id)}
                      className="w-full flex items-center justify-between px-3 py-2.5 bg-secondary/30 hover:bg-secondary/60 transition-colors text-left cursor-pointer"
                    >
                      <div className="flex items-center gap-2 font-semibold text-foreground">
                        <span className="text-primary">{CATEGORY_ICONS[cat.icon]}</span>
                        <span className="text-xs">{cat.title}</span>
                      </div>
                      <div className="flex items-center gap-1.5 text-muted-foreground">
                        <span className="text-[10px] bg-secondary px-1.5 py-0.5 rounded font-mono">
                          {cat.articleIds.length}
                        </span>
                        {isExpanded ? (
                          <ChevronDown className="w-3.5 h-3.5" />
                        ) : (
                          <ChevronRight className="w-3.5 h-3.5" />
                        )}
                      </div>
                    </button>

                    <AnimatePresence initial={false}>
                      {isExpanded && (
                        <motion.div
                          initial={{ height: 0, opacity: 0 }}
                          animate={{ height: 'auto', opacity: 1 }}
                          exit={{ height: 0, opacity: 0 }}
                          transition={{ duration: 0.18, ease: 'easeInOut' }}
                          className="overflow-hidden"
                        >
                          <div className="py-1 px-1.5 space-y-0.5 border-t border-border/30">
                            {cat.articleIds.map((artId) => {
                              const art = DOCS_ARTICLES[artId]
                              if (!art) return null
                              const isSelected = selectedArticleId === art.id

                              return (
                                <button
                                  key={art.id}
                                  onClick={() => setSelectedArticleId(art.id)}
                                  className={cn(
                                    'w-full flex items-center justify-between px-2.5 py-1.5 rounded-md text-xs transition-colors text-left cursor-pointer focus-visible:ring-2 focus-visible:ring-primary focus-visible:outline-none',
                                    isSelected
                                      ? 'bg-primary/10 text-primary font-semibold border-l-2 border-primary pl-2'
                                      : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
                                  )}
                                >
                                  <span className="truncate">{art.title}</span>
                                  <span className="text-[10px] text-muted-foreground/70 shrink-0 ml-1">
                                    {art.readTime.split(' ')[0]}m
                                  </span>
                                </button>
                              )
                            })}
                          </div>
                        </motion.div>
                      )}
                    </AnimatePresence>
                  </div>
                )
              })
            )}

            {/* Quick Link Card to Agent Console & OpenAPI */}
            <div className="p-3 rounded-lg border border-border/60 bg-secondary/20 space-y-2 mt-6">
              <div className="flex items-center gap-2 text-foreground font-semibold">
                <Sparkles className="w-3.5 h-3.5 text-primary" />
                <span>Interactive Tools</span>
              </div>
              <p className="text-[11px] text-muted-foreground">
                Test MCP endpoints or explore the dynamic OpenAPI contract directly.
              </p>
              <div className="space-y-1 pt-1">
                {onNavigateTab && (
                  <button
                    onClick={() => onNavigateTab('agents')}
                    className="w-full flex items-center justify-between p-1.5 text-[11px] rounded bg-secondary hover:bg-secondary/80 text-foreground transition-colors cursor-pointer"
                  >
                    <span>Agent Console</span>
                    <ArrowRight className="w-3 h-3 text-muted-foreground" />
                  </button>
                )}
                <a
                  href="/api/v1/openapi.yaml"
                  target="_blank"
                  rel="noreferrer"
                  className="w-full flex items-center justify-between p-1.5 text-[11px] rounded bg-secondary hover:bg-secondary/80 text-foreground transition-colors"
                >
                  <span>OpenAPI 3.1 Contract</span>
                  <ExternalLink className="w-3 h-3 text-muted-foreground" />
                </a>
              </div>
            </div>
          </div>
        </aside>

        {/* CENTER COLUMN (Article Content): 7 cols */}
        <main className="lg:col-span-6 space-y-8 min-w-0" role="main">
          {/* Article Header */}
          <div className="space-y-3 pb-6 border-b border-border/60">
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant="outline" className="text-xs">
                {currentCategory?.title}
              </Badge>
              {currentArticle.badge && (
                <Badge variant="outline" className="text-xs font-mono">
                  {currentArticle.badge}
                </Badge>
              )}
              <span className="text-xs text-muted-foreground">· {currentArticle.readTime}</span>
              <span className="text-xs text-muted-foreground">· Updated {currentArticle.lastUpdated}</span>
            </div>

            <h1 className="text-3xl font-extrabold tracking-tight text-foreground">
              {currentArticle.title}
            </h1>

            <p className="text-base text-muted-foreground leading-relaxed">
              {currentArticle.lead}
            </p>
          </div>

          {/* Primer Style Callouts */}
          {currentArticle.callouts && currentArticle.callouts.length > 0 && (
            <div className="space-y-3">
              {currentArticle.callouts.map((callout, idx) => {
                const isWarning = callout.type === 'warning'
                const isImportant = callout.type === 'important'
                const isTip = callout.type === 'tip'

                const borderCol = isWarning
                  ? 'border-amber-500/40 bg-amber-500/10'
                  : isImportant
                  ? 'border-purple-500/40 bg-purple-500/10'
                  : isTip
                  ? 'border-emerald-500/40 bg-emerald-500/10'
                  : 'border-blue-500/40 bg-blue-500/10'

                const textCol = isWarning
                  ? 'text-amber-500'
                  : isImportant
                  ? 'text-purple-400'
                  : isTip
                  ? 'text-emerald-500'
                  : 'text-blue-400'

                const Icon = isWarning
                  ? AlertTriangle
                  : isImportant
                  ? Flame
                  : isTip
                  ? Lightbulb
                  : Info

                return (
                  <div
                    key={idx}
                    className={cn('p-4 rounded-lg border-l-4 text-xs space-y-1', borderCol)}
                  >
                    <div className="flex items-center gap-1.5 font-semibold text-foreground">
                      <Icon className={cn('w-4 h-4', textCol)} />
                      <span>{callout.title || callout.type.toUpperCase()}</span>
                    </div>
                    <p className="text-foreground/90 leading-relaxed pl-5.5">
                      {callout.content}
                    </p>
                  </div>
                )
              })}
            </div>
          )}

          {/* Multi-Tab Interactive Code Snippets */}
          {currentArticle.codeSnippets && currentArticle.codeSnippets.length > 0 && (
            <div className="space-y-2">
              <div className="flex items-center justify-between border-b border-border/60 pb-1">
                <div className="flex items-center gap-1">
                  {currentArticle.codeSnippets.map((snip, idx) => (
                    <button
                      key={idx}
                      onClick={() => setActiveCodeTab(idx)}
                      className={cn(
                        'relative px-3 py-1 text-xs font-medium rounded-t-md transition-colors cursor-pointer focus-visible:ring-2 focus-visible:ring-primary focus-visible:outline-none',
                        activeCodeTab === idx
                          ? 'bg-secondary text-foreground font-semibold'
                          : 'text-muted-foreground hover:text-foreground'
                      )}
                    >
                      <span>{snip.label}</span>
                      {activeCodeTab === idx && (
                        <motion.div
                          layoutId="activeCodeTabIndicator"
                          className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary"
                          transition={{ type: 'spring', stiffness: 500, damping: 35 }}
                        />
                      )}
                    </button>
                  ))}
                </div>
                <button
                  onClick={() =>
                    handleCopy(
                      currentArticle.codeSnippets![activeCodeTab].code,
                      `snippet-${activeCodeTab}`
                    )
                  }
                  className="inline-flex items-center gap-1 text-[11px] text-muted-foreground hover:text-foreground px-2 py-0.5 rounded hover:bg-secondary transition-colors cursor-pointer"
                >
                  {copiedSnippet === `snippet-${activeCodeTab}` ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-success" />
                      <span className="text-success font-medium">Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5" />
                      <span>Copy code</span>
                    </>
                  )}
                </button>
              </div>

              <pre className="p-4 rounded-lg bg-zinc-950 text-zinc-100 font-mono text-xs overflow-x-auto leading-relaxed border border-border/40 shadow-inner">
                <code>{currentArticle.codeSnippets[activeCodeTab].code}</code>
              </pre>
            </div>
          )}

          {/* Interactive Parameters Table (if defined) */}
          {currentArticle.params && currentArticle.params.length > 0 && (
            <div className="space-y-3">
              <h2 className="text-lg font-bold text-foreground">Configuration & Parameters</h2>
              <div className="border border-border rounded-lg overflow-hidden text-xs">
                <table className="w-full text-left">
                  <thead className="bg-secondary/40 text-muted-foreground font-semibold border-b border-border">
                    <tr>
                      <th className="p-2.5">Name</th>
                      <th className="p-2.5">Type</th>
                      <th className="p-2.5">Required</th>
                      <th className="p-2.5">Description</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border/60">
                    {currentArticle.params.map((param) => (
                      <tr key={param.name} className="hover:bg-secondary/20 transition-colors">
                        <td className="p-2.5 font-mono text-foreground font-semibold">{param.name}</td>
                        <td className="p-2.5 font-mono text-muted-foreground">{param.type}</td>
                        <td className="p-2.5">
                          {param.required ? (
                            <Badge variant="outline" className="text-[10px] text-primary border-primary/30">
                              required
                            </Badge>
                          ) : (
                            <span className="text-muted-foreground text-[11px]">optional</span>
                          )}
                        </td>
                        <td className="p-2.5 text-muted-foreground leading-relaxed">
                          {param.description}
                          {param.default && (
                            <div className="text-[10px] text-muted-foreground/80 mt-0.5">
                              Default: <code className="font-mono">{param.default}</code>
                            </div>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Article Detailed Sections */}
          <div className="space-y-6">
            {currentArticle.sections.map((sec) => (
              <section key={sec.id} id={sec.id} className="scroll-mt-24 space-y-2">
                <h2 className="text-lg font-bold text-foreground flex items-center gap-2 group">
                  <a href={`#${sec.id}`} className="hover:underline">
                    {sec.title}
                  </a>
                </h2>
                <div className="text-xs text-muted-foreground leading-relaxed space-y-2">
                  <p>
                    Refer to the enterprise implementation specifications and runtime configurations
                    for details regarding <strong>{sec.title}</strong>. For automated CI/CD integration,
                    configure rulesets and ambient agent leases to ensure zero-trust compliance.
                  </p>
                </div>
              </section>
            ))}
          </div>

          {/* Related Articles Cards */}
          {currentArticle.relatedArticles && currentArticle.relatedArticles.length > 0 && (
            <div className="pt-6 border-t border-border/60 space-y-3">
              <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Related Articles
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {currentArticle.relatedArticles.map((relId) => {
                  const rel = DOCS_ARTICLES[relId]
                  if (!rel) return null
                  return (
                    <button
                      key={rel.id}
                      onClick={() => setSelectedArticleId(rel.id)}
                      className="p-3 text-left rounded-lg border border-border/60 bg-secondary/20 hover:bg-secondary/50 hover:border-border transition-colors cursor-pointer group"
                    >
                      <div className="font-semibold text-xs text-foreground group-hover:text-primary transition-colors flex items-center justify-between">
                        <span>{rel.title}</span>
                        <ChevronRight className="w-3.5 h-3.5 text-muted-foreground group-hover:translate-x-0.5 transition-transform" />
                      </div>
                      <p className="text-[11px] text-muted-foreground mt-1 line-clamp-2">{rel.lead}</p>
                    </button>
                  )
                })}
              </div>
            </div>
          )}

          {/* Next / Previous Article Navigation */}
          <div className="pt-6 border-t border-border/60 flex items-center justify-between gap-4">
            {prevArticle ? (
              <button
                onClick={() => setSelectedArticleId(prevArticle.id)}
                className="flex items-center gap-2 p-2.5 rounded-lg border border-border/60 hover:bg-secondary/40 transition-colors text-left text-xs cursor-pointer max-w-[48%]"
              >
                <ArrowLeft className="w-4 h-4 text-muted-foreground shrink-0" />
                <div className="truncate">
                  <div className="text-[10px] text-muted-foreground">Previous</div>
                  <div className="font-semibold text-foreground truncate">{prevArticle.title}</div>
                </div>
              </button>
            ) : (
              <div />
            )}

            {nextArticle ? (
              <button
                onClick={() => setSelectedArticleId(nextArticle.id)}
                className="flex items-center gap-2 p-2.5 rounded-lg border border-border/60 hover:bg-secondary/40 transition-colors text-right text-xs cursor-pointer max-w-[48%] ml-auto"
              >
                <div className="truncate">
                  <div className="text-[10px] text-muted-foreground">Next</div>
                  <div className="font-semibold text-foreground truncate">{nextArticle.title}</div>
                </div>
                <ArrowRight className="w-4 h-4 text-muted-foreground shrink-0" />
              </button>
            ) : null}
          </div>

          {/* GitHub Docs Style Article Feedback Widget */}
          <div className="p-4 rounded-lg border border-border/60 bg-secondary/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
            <div>
              <div className="font-semibold text-foreground">Did this article help you?</div>
              <p className="text-[11px] text-muted-foreground">
                Your feedback helps us continuously improve the Refweir documentation.
              </p>
            </div>
            <AnimatePresence mode="wait">
              {feedbackSubmitted ? (
                <motion.div
                  key="feedback-thanks"
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  className="flex items-center gap-1.5 text-success font-medium"
                >
                  <Check className="w-4 h-4" />
                  <span>Thank you for your feedback!</span>
                </motion.div>
              ) : (
                <motion.div
                  key="feedback-actions"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="flex items-center gap-2"
                >
                  <button
                    onClick={() => setFeedbackSubmitted(true)}
                    className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-border bg-background hover:bg-secondary text-foreground transition-colors cursor-pointer focus-visible:ring-2 focus-visible:ring-primary focus-visible:outline-none"
                  >
                    <ThumbsUp className="w-3.5 h-3.5 text-primary" />
                    <span>Yes</span>
                  </button>
                  <button
                    onClick={() => setFeedbackSubmitted(true)}
                    className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-border bg-background hover:bg-secondary text-foreground transition-colors cursor-pointer focus-visible:ring-2 focus-visible:ring-primary focus-visible:outline-none"
                  >
                    <ThumbsDown className="w-3.5 h-3.5 text-muted-foreground" />
                    <span>No</span>
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </main>

        {/* RIGHT COLUMN ("In this article" / On-Page TOC): 3 cols */}
        <aside className="hidden lg:block lg:col-span-3 space-y-4">
          <div className="sticky top-20 space-y-4 text-xs">
            <div className="p-4 rounded-lg border border-border/60 bg-card/20 space-y-3">
              <div className="font-semibold text-foreground uppercase tracking-wider text-[11px]">
                In this article
              </div>
              <nav className="space-y-1.5">
                {currentArticle.sections.map((sec) => (
                  <a
                    key={sec.id}
                    href={`#${sec.id}`}
                    onClick={() => setActiveHeadingId(sec.id)}
                    className={cn(
                      'block text-muted-foreground hover:text-foreground transition-colors leading-snug',
                      activeHeadingId === sec.id ? 'text-primary font-semibold' : ''
                    )}
                  >
                    {sec.title}
                  </a>
                ))}
              </nav>
            </div>

            {/* Quick Actions Card */}
            <div className="p-3.5 rounded-lg border border-border/60 bg-secondary/30 space-y-2 text-xs">
              <div className="font-semibold text-foreground text-[11px] uppercase tracking-wider">
                Quick Resources
              </div>
              <div className="space-y-1">
                <a
                  href="/api/v1/openapi.yaml"
                  target="_blank"
                  rel="noreferrer"
                  className="flex items-center justify-between text-muted-foreground hover:text-foreground transition-colors py-1"
                >
                  <span>OpenAPI Spec (3.1)</span>
                  <ExternalLink className="w-3 h-3" />
                </a>
                <a
                  href="/healthz"
                  target="_blank"
                  rel="noreferrer"
                  className="flex items-center justify-between text-muted-foreground hover:text-foreground transition-colors py-1"
                >
                  <span>Cluster Healthz API</span>
                  <ExternalLink className="w-3 h-3" />
                </a>
                <button
                  onClick={() => {
                    navigator.clipboard.writeText(window.location.href)
                    alert('Article URL copied to clipboard')
                  }}
                  className="w-full flex items-center justify-between text-muted-foreground hover:text-foreground transition-colors py-1 cursor-pointer text-left"
                >
                  <span>Copy Page Link</span>
                  <Copy className="w-3 h-3" />
                </button>
              </div>
            </div>
          </div>
        </aside>
      </div>
    </div>
  )
}
