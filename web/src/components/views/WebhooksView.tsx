import React, { useState, useEffect, useMemo } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  getPaginationRowModel,
  type ColumnDef,
  type SortingState,
} from '@tanstack/react-table'
import {
  Webhook,
  RefreshCw,
  CheckCircle2,
  ChevronDown,
  ChevronRight,
  Send,
  AlertCircle,
  Copy,
  Check,
  Search,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  ChevronLeft,
} from 'lucide-react'
import { Box, BoxHeader, BoxBody } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { fetchWebhookDeliveries, pingWebhook, batchRedeliverWebhooks } from '@/lib/api'

export const WebhooksView: React.FC = () => {
  const [deliveries, setDeliveries] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [pinging, setPinging] = useState(false)
  const [redeliveringId, setRedeliveringId] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [feedbackMsg, setFeedbackMsg] = useState<string | null>(null)

  const [globalFilter, setGlobalFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'success' | 'failure'>('all')
  const [sorting, setSorting] = useState<SortingState>([])

  const showFeedback = (msg: string) => {
    setFeedbackMsg(msg)
    setTimeout(() => setFeedbackMsg(null), 3000)
  }

  const loadDeliveries = async (isManual = false) => {
    if (isManual) setRefreshing(true)
    else setLoading(true)

    try {
      const data = await fetchWebhookDeliveries()
      setDeliveries(data)
      if (data.length > 0 && !expandedId) {
        setExpandedId(data[0].id)
      }
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    loadDeliveries()
  }, [])

  const handlePing = async () => {
    setPinging(true)
    try {
      const res = await pingWebhook()
      if (res.ok) {
        showFeedback('Test ping dispatched successfully')
        loadDeliveries(true)
      } else {
        showFeedback('Ping failed: ' + res.error)
      }
    } finally {
      setPinging(false)
    }
  }

  const handleRedeliver = async (id: string) => {
    setRedeliveringId(id)
    try {
      const ok = await batchRedeliverWebhooks([id])
      if (ok) {
        showFeedback('Webhook redelivery queued')
        loadDeliveries(true)
      } else {
        showFeedback('Redelivery failed')
      }
    } finally {
      setRedeliveringId(null)
    }
  }

  const handleCopy = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const data = useMemo(() => {
    if (statusFilter === 'all') return deliveries
    return deliveries.filter((del) => {
      const st = del.statusCode || del.status || 200
      if (statusFilter === 'success') return st >= 200 && st < 300
      return st >= 400 || st === 0
    })
  }, [deliveries, statusFilter])

  const columns = useMemo<ColumnDef<any>[]>(
    () => [
      {
        id: 'expand',
        header: '',
        cell: ({ row }) => {
          const isExpanded = expandedId === row.original.id
          return (
            <button
              type="button"
              onClick={() => setExpandedId(isExpanded ? null : row.original.id)}
              className="text-muted-foreground hover:text-foreground cursor-pointer p-0.5"
              aria-label={isExpanded ? 'Collapse payload' : 'Expand payload'}
            >
              {isExpanded ? (
                <ChevronDown className="w-4 h-4" />
              ) : (
                <ChevronRight className="w-4 h-4" />
              )}
            </button>
          )
        },
        enableSorting: false,
      },
      {
        accessorKey: 'event',
        header: 'Event',
        cell: ({ row }) => (
          <span className="font-mono text-xs font-semibold text-foreground bg-secondary px-2 py-0.5 rounded border border-border shrink-0 inline-block">
            {row.original.event || 'push'}
          </span>
        ),
      },
      {
        accessorKey: 'targetUrl',
        header: 'Endpoint URL',
        cell: ({ row }) => {
          const url = row.original.targetUrl || row.original.url || 'https://ci.internal.refweir.dev/hooks/forge'
          return (
            <span className="text-xs text-foreground font-mono truncate max-w-xs md:max-w-md block" title={url}>
              {url}
            </span>
          )
        },
      },
      {
        id: 'status',
        accessorFn: (row) => row.statusCode || row.status || 200,
        header: 'Status',
        cell: ({ row }) => {
          const status = row.original.statusCode || row.original.status || 200
          const isOk = status >= 200 && status < 300
          return (
            <Badge variant={isOk ? 'success' : 'closed'} className="text-[10px] font-mono py-0 gap-1 whitespace-nowrap">
              {isOk ? <CheckCircle2 className="w-3 h-3 text-success" /> : <AlertCircle className="w-3 h-3 text-destructive" />}
              <span>HTTP {status}</span>
            </Badge>
          )
        },
      },
      {
        id: 'latency',
        accessorFn: (row) => row.durationMs || 32,
        header: 'Duration',
        cell: ({ row }) => {
          const lat = row.original.latency || `${row.original.durationMs || 32}ms`
          return <span className="text-xs font-mono text-muted-foreground whitespace-nowrap">{lat}</span>
        },
      },
      {
        accessorKey: 'timestamp',
        header: 'Time',
        cell: ({ row }) => (
          <span className="text-[11px] text-muted-foreground whitespace-nowrap">
            {row.original.timestamp || 'Recently'}
          </span>
        ),
      },
      {
        id: 'actions',
        header: '',
        cell: ({ row }) => {
          const id = row.original.id
          return (
            <div className="flex justify-end">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleRedeliver(id)}
                disabled={redeliveringId === id}
                className="gap-1 h-7 px-2 text-xs"
                title="Redeliver webhook payload"
              >
                <RefreshCw className={`w-3 h-3 ${redeliveringId === id ? 'animate-spin' : ''}`} />
                <span className="hidden sm:inline">Redeliver</span>
              </Button>
            </div>
          )
        },
        enableSorting: false,
      },
    ],
    [expandedId, redeliveringId]
  )

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      globalFilter,
    },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const q = String(filterValue).toLowerCase()
      const d = row.original
      const evt = String(d.event || '').toLowerCase()
      const url = String(d.targetUrl || d.url || '').toLowerCase()
      return evt.includes(q) || url.includes(q)
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    initialState: {
      pagination: {
        pageSize: 10,
      },
    },
  })

  return (
    <div className="space-y-4">
      {/* Header Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
            <Webhook className="w-4 h-4 text-primary" />
            <span>Webhook Delivery Log</span>
          </h2>
          <p className="text-xs text-muted-foreground">
            Audit, inspect, and redeliver dispatched JSON webhooks with TanStack Table.
          </p>
        </div>

        <div className="flex items-center gap-2 self-start sm:self-auto">
          {feedbackMsg && (
            <Badge variant="outline" className="text-xs font-mono text-success border-success/40 gap-1">
              <Check className="w-3 h-3" />
              <span>{feedbackMsg}</span>
            </Badge>
          )}

          <Button
            variant="outline"
            size="sm"
            onClick={handlePing}
            disabled={pinging}
            className="text-xs gap-1.5"
          >
            <Send className="w-3.5 h-3.5" />
            <span>Send Test Ping</span>
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={() => loadDeliveries(true)}
            disabled={refreshing}
            className="text-xs gap-1.5"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${refreshing ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
        {/* Status Filters */}
        <div className="flex items-center gap-1 p-0.5 bg-secondary border border-border rounded-md text-xs w-full sm:w-auto">
          {(['all', 'success', 'failure'] as const).map((st) => (
            <button
              key={st}
              type="button"
              onClick={() => {
                setStatusFilter(st)
                table.setPageIndex(0)
              }}
              className={`px-2.5 py-1 rounded capitalize transition-colors cursor-pointer ${
                statusFilter === st
                  ? 'bg-background text-foreground font-semibold shadow-2xs'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {st === 'all' ? 'All Deliveries' : st === 'success' ? '2xx Success' : 'Failures'}
            </button>
          ))}
        </div>

        {/* Search */}
        <div className="relative flex-1 max-w-sm w-full">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search by event or endpoint URL..."
            value={globalFilter ?? ''}
            onChange={(e) => {
              setGlobalFilter(e.target.value)
              table.setPageIndex(0)
            }}
            className="w-full h-8 pl-8 pr-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
          />
        </div>
      </div>

      {loading ? (
        <Box className="p-8 text-center">
          <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground font-mono">
            <RefreshCw className="w-4 h-4 animate-spin text-primary" />
            <span>Loading webhook deliveries from daemon...</span>
          </div>
        </Box>
      ) : deliveries.length === 0 ? (
        /* Empty State */
        <Box className="p-8 text-center space-y-4">
          <div className="w-10 h-10 rounded-md bg-secondary border border-border text-muted-foreground flex items-center justify-center mx-auto">
            <Webhook className="w-5 h-5" />
          </div>
          <div className="space-y-1 max-w-md mx-auto">
            <h4 className="text-sm font-semibold text-foreground">No Webhook Deliveries Logged</h4>
            <p className="text-xs text-muted-foreground leading-relaxed">
              When Git pushes, pull requests, or CI checks occur, automated JSON webhooks are signed with HMAC-SHA256 and dispatched to external endpoints.
            </p>
          </div>
          <Button variant="secondary" size="sm" onClick={handlePing} disabled={pinging} className="gap-1.5 text-xs">
            <Send className="w-3.5 h-3.5" />
            <span>Trigger Test Ping Event</span>
          </Button>
        </Box>
      ) : (
        /* Deliveries Table Box */
        <Box>
          <BoxHeader className="bg-secondary/35 py-2.5 flex items-center justify-between">
            <span className="font-semibold text-foreground text-xs">Recent Deliveries</span>
            <span className="text-muted-foreground font-mono text-[11px]">
              {table.getFilteredRowModel().rows.length} records
            </span>
          </BoxHeader>

          <BoxBody className="p-0 overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                {table.getHeaderGroups().map((headerGroup) => (
                  <tr key={headerGroup.id} className="border-b border-border bg-secondary/20">
                    {headerGroup.headers.map((header) => {
                      const canSort = header.column.getCanSort()
                      const isSorted = header.column.getIsSorted()
                      return (
                        <th
                          key={header.id}
                          className="px-3.5 py-2 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider select-none"
                        >
                          {header.isPlaceholder ? null : canSort ? (
                            <button
                              type="button"
                              onClick={header.column.getToggleSortingHandler()}
                              className="flex items-center gap-1 hover:text-foreground transition-colors cursor-pointer"
                            >
                              {typeof header.column.columnDef.header === 'function'
                                ? header.column.columnDef.header(header.getContext())
                                : header.column.columnDef.header}
                              {isSorted === 'asc' ? (
                                <ArrowUp className="w-3 h-3 text-primary" />
                              ) : isSorted === 'desc' ? (
                                <ArrowDown className="w-3 h-3 text-primary" />
                              ) : (
                                <ArrowUpDown className="w-2.5 h-2.5 opacity-40 hover:opacity-100" />
                              )}
                            </button>
                          ) : typeof header.column.columnDef.header === 'function' ? (
                            header.column.columnDef.header(header.getContext())
                          ) : (
                            header.column.columnDef.header
                          )}
                        </th>
                      )
                    })}
                  </tr>
                ))}
              </thead>
              <tbody className="divide-y divide-border/60">
                {table.getRowModel().rows.map((row) => {
                  const del = row.original
                  const isExpanded = expandedId === del.id
                  return (
                    <React.Fragment key={row.id}>
                      <tr className="hover:bg-secondary/20 transition-colors">
                        {row.getVisibleCells().map((cell) => (
                          <td key={cell.id} className="px-3.5 py-3 align-middle">
                            {typeof cell.column.columnDef.cell === 'function'
                              ? cell.column.columnDef.cell(cell.getContext())
                              : (cell.getValue() as React.ReactNode)}
                          </td>
                        ))}
                      </tr>
                      {isExpanded && (
                        <tr>
                          <td colSpan={columns.length} className="p-4 bg-secondary/15 border-t border-border font-mono text-xs">
                            <div className="space-y-4">
                              <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-2">
                                <div className="flex items-center gap-2">
                                  <span className="font-semibold text-foreground text-xs font-sans">Delivery Details:</span>
                                  <code className="text-primary text-[11px] font-mono">{del.id}</code>
                                </div>
                                <div className="flex items-center gap-2">
                                  <Button
                                    variant="outline"
                                    size="xs"
                                    onClick={() => handleRedeliver(del.id)}
                                    disabled={redeliveringId === del.id}
                                    className="gap-1 h-6 text-xs"
                                  >
                                    <RefreshCw className={`w-3 h-3 ${redeliveringId === del.id ? 'animate-spin' : ''}`} />
                                    <span>Redeliver Event</span>
                                  </Button>
                                  <Button
                                    variant="ghost"
                                    size="xs"
                                    onClick={() => handleCopy(JSON.stringify(del.payload || del.request?.body || {}, null, 2), del.id)}
                                    className="gap-1 h-6 text-xs"
                                  >
                                    {copiedId === del.id ? (
                                      <Check className="w-3 h-3 text-success" />
                                    ) : (
                                      <Copy className="w-3 h-3" />
                                    )}
                                    <span>Copy JSON</span>
                                  </Button>
                                </div>
                              </div>

                              <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-[11px]">
                                {/* Request Headers */}
                                <div className="space-y-1.5 p-3 rounded-md bg-background border border-border">
                                  <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block font-sans">
                                    Request Headers
                                  </span>
                                  <div className="space-y-1 text-muted-foreground">
                                    <div><span className="text-foreground">X-Refweir-Event:</span> {del.event || 'push'}</div>
                                    <div><span className="text-foreground">X-Hub-Signature-256:</span> sha256={del.signature || 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'}</div>
                                    <div><span className="text-foreground">X-Refweir-Delivery:</span> {del.id}</div>
                                    <div><span className="text-foreground">Content-Type:</span> application/json</div>
                                    <div><span className="text-foreground">User-Agent:</span> Refweir-Hookshot/1.0</div>
                                  </div>
                                </div>

                                {/* Response Details */}
                                <div className="space-y-1.5 p-3 rounded-md bg-background border border-border">
                                  <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block font-sans">
                                    Target Response
                                  </span>
                                  <div className="space-y-1 text-muted-foreground">
                                    <div className="flex items-center gap-1.5">
                                      <span className="text-foreground">Status:</span>
                                      <Badge variant={(del.statusCode || 200) >= 200 && (del.statusCode || 200) < 300 ? 'success' : 'closed'} className="text-[9px] py-0 font-mono">
                                        HTTP {del.statusCode || 200}
                                      </Badge>
                                    </div>
                                    <div><span className="text-foreground">Duration:</span> {del.durationMs || 32}ms</div>
                                    <div className="truncate"><span className="text-foreground">Response Preview:</span> {del.responseBody || '200 OK - Webhook acknowledged'}</div>
                                  </div>
                                </div>
                              </div>

                              {/* Payload JSON */}
                              <div className="space-y-1.5">
                                <span className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider block font-sans">
                                  Request Payload (JSON)
                                </span>
                                <pre className="text-foreground overflow-x-auto p-3 bg-background rounded-md border border-border text-[11px] max-h-56">
                                  {JSON.stringify(del.payload || del.request?.body || { event: del.event, id: del.id }, null, 2)}
                                </pre>
                              </div>
                            </div>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  )
                })}
              </tbody>
            </table>
          </BoxBody>

          {/* Pagination */}
          {table.getPageCount() > 1 && (
            <div className="flex items-center justify-between px-3.5 py-2 border-t border-border bg-secondary/15 text-xs text-muted-foreground">
              <div>
                Page <strong className="text-foreground">{table.getState().pagination.pageIndex + 1}</strong> of{' '}
                <strong className="text-foreground">{table.getPageCount()}</strong>
              </div>
              <div className="flex items-center gap-1.5">
                <Button
                  variant="outline"
                  size="xs"
                  onClick={() => table.previousPage()}
                  disabled={!table.getCanPreviousPage()}
                  className="gap-1 text-xs"
                >
                  <ChevronLeft className="w-3 h-3" />
                  <span>Previous</span>
                </Button>
                <Button
                  variant="outline"
                  size="xs"
                  onClick={() => table.nextPage()}
                  disabled={!table.getCanNextPage()}
                  className="gap-1 text-xs"
                >
                  <span>Next</span>
                  <ChevronRight className="w-3 h-3" />
                </Button>
              </div>
            </div>
          )}
        </Box>
      )}
    </div>
  )
}
