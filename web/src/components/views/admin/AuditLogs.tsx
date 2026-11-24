import React, { useState, useEffect, useMemo } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  getPaginationRowModel,
  type ColumnDef,
  type SortingState,
  type RowSelectionState,
} from '@tanstack/react-table'
import {
  ShieldAlert,
  Download,
  Search,
  AlertCircle,
  Info,
  Bot,
  User,
  Terminal,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  ChevronLeft,
  ChevronRight,
  CheckSquare,
  Square,
} from 'lucide-react'
import { Box, BoxHeader, BoxBody } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { getAuditEvents, type AuditEvent } from '@/lib/api'

export const AuditLogs: React.FC = () => {
  const [events, setEvents] = useState<AuditEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [globalFilter, setGlobalFilter] = useState('')
  const [severityFilter, setSeverityFilter] = useState<'all' | 'info' | 'warning' | 'critical'>('all')
  const [sorting, setSorting] = useState<SortingState>([])
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})

  useEffect(() => {
    async function load() {
      setLoading(true)
      try {
        const data = await getAuditEvents()
        setEvents(data)
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  // Severity-filtered dataset for table
  const data = useMemo(() => {
    if (severityFilter === 'all') return events
    return events.filter((ev) => ev.severity === severityFilter)
  }, [events, severityFilter])

  const columns = useMemo<ColumnDef<AuditEvent>[]>(
    () => [
      {
        id: 'select',
        header: ({ table }) => (
          <button
            type="button"
            onClick={table.getToggleAllRowsSelectedHandler()}
            className="text-muted-foreground hover:text-foreground cursor-pointer flex items-center justify-center p-0.5"
            aria-label="Select all rows"
          >
            {table.getIsAllRowsSelected() ? (
              <CheckSquare className="w-3.5 h-3.5 text-primary" />
            ) : (
              <Square className="w-3.5 h-3.5" />
            )}
          </button>
        ),
        cell: ({ row }) => (
          <button
            type="button"
            onClick={row.getToggleSelectedHandler()}
            className="text-muted-foreground hover:text-foreground cursor-pointer flex items-center justify-center p-0.5"
            aria-label={`Select row ${row.original.id}`}
          >
            {row.getIsSelected() ? (
              <CheckSquare className="w-3.5 h-3.5 text-primary" />
            ) : (
              <Square className="w-3.5 h-3.5" />
            )}
          </button>
        ),
        enableSorting: false,
      },
      {
        accessorKey: 'severity',
        header: 'Severity',
        cell: ({ row }) => {
          const sev = row.original.severity
          if (sev === 'critical') {
            return (
              <Badge variant="closed" className="text-[10px] py-0 gap-1 whitespace-nowrap">
                <AlertCircle className="w-3 h-3" /> Critical
              </Badge>
            )
          }
          if (sev === 'warning') {
            return (
              <Badge variant="warning" className="text-[10px] py-0 gap-1 whitespace-nowrap">
                <AlertCircle className="w-3 h-3" /> Warning
              </Badge>
            )
          }
          return (
            <Badge variant="default" className="text-[10px] py-0 gap-1 whitespace-nowrap">
              <Info className="w-3 h-3 text-primary" /> Info
            </Badge>
          )
        },
      },
      {
        accessorKey: 'action',
        header: 'Action',
        cell: ({ row }) => (
          <span className="font-mono text-xs font-semibold text-foreground">
            {row.original.action}
          </span>
        ),
      },
      {
        accessorKey: 'actor',
        header: 'Actor',
        cell: ({ row }) => {
          const { actor, actorType } = row.original
          return (
            <span className="flex items-center gap-1.5 text-xs font-mono text-foreground">
              {actorType === 'agent' ? (
                <Bot className="w-3.5 h-3.5 text-primary shrink-0" />
              ) : actorType === 'user' ? (
                <User className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
              ) : (
                <Terminal className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
              )}
              <span className="truncate max-w-[120px]">{actor}</span>
            </span>
          )
        },
      },
      {
        accessorKey: 'resource',
        header: 'Resource',
        cell: ({ row }) => (
          <span className="font-mono text-[11px] text-muted-foreground truncate max-w-[150px] block">
            {row.original.resource}
          </span>
        ),
      },
      {
        accessorKey: 'details',
        header: 'Details',
        cell: ({ row }) => (
          <div className="text-xs text-foreground/90 max-w-md line-clamp-2 leading-relaxed">
            {row.original.details}
          </div>
        ),
      },
      {
        accessorKey: 'clientIp',
        header: 'Client IP',
        cell: ({ row }) => (
          <span className="font-mono text-[11px] text-muted-foreground whitespace-nowrap">
            {row.original.clientIp}
          </span>
        ),
      },
      {
        accessorKey: 'timestamp',
        header: 'Timestamp',
        cell: ({ row }) => (
          <span className="font-mono text-[11px] text-muted-foreground whitespace-nowrap">
            {row.original.timestamp}
          </span>
        ),
      },
    ],
    []
  )

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      globalFilter,
      rowSelection,
    },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    onRowSelectionChange: setRowSelection,
    globalFilterFn: (row, _columnId, filterValue) => {
      const q = String(filterValue).toLowerCase()
      const ev = row.original
      return (
        ev.action.toLowerCase().includes(q) ||
        ev.actor.toLowerCase().includes(q) ||
        ev.resource.toLowerCase().includes(q) ||
        ev.clientIp.toLowerCase().includes(q) ||
        ev.details.toLowerCase().includes(q)
      )
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

  // Selected items or all filtered items for export
  const getExportData = () => {
    const selectedRows = table.getSelectedRowModel().rows
    if (selectedRows.length > 0) {
      return selectedRows.map((r) => r.original)
    }
    return table.getFilteredRowModel().rows.map((r) => r.original)
  }

  const exportToJson = () => {
    const targetData = getExportData()
    const blob = new Blob([JSON.stringify(targetData, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `refweir-audit-log-${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
  }

  const exportToCsv = () => {
    const targetData = getExportData()
    const headers = ['ID', 'Timestamp', 'Actor', 'ActorType', 'Action', 'Resource', 'ClientIP', 'Severity', 'Details']
    const rows = targetData.map((e) => [
      e.id,
      e.timestamp,
      e.actor,
      e.actorType,
      e.action,
      e.resource,
      e.clientIp,
      e.severity,
      `"${e.details.replace(/"/g, '""')}"`,
    ])
    const csvContent = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n')
    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `refweir-audit-log-${Date.now()}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  const exportFullSIEM = async (format: 'ndjson' | 'csv' = 'ndjson') => {
    try {
      const resp = await fetch(`/api/v1/admin/audit-logs/export?format=${format}`)
      if (!resp.ok) {
        throw new Error(`Export failed with HTTP ${resp.status}`)
      }
      const blob = await resp.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `refweir-audit-siem-${Date.now()}.${format === 'ndjson' ? 'ndjson' : 'csv'}`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      console.error('Failed to export SIEM audit logs', err)
      if (format === 'csv') {
        exportToCsv()
      } else {
        exportToJson()
      }
    }
  }

  const selectedCount = Object.keys(rowSelection).length

  return (
    <div className="space-y-4">
      {/* Header & Export Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
            <ShieldAlert className="w-4 h-4 text-primary" />
            <span>Security & Compliance Audit Log</span>
          </h2>
          <p className="text-xs text-muted-foreground">
            Immutable audit trail of Git pushes, access grants, ruleset bypasses, and MCP agent calls. Powered by TanStack Table.
          </p>
        </div>

        {/* Export Buttons */}
        <div className="flex items-center gap-2 self-start sm:self-auto flex-wrap">
          {selectedCount > 0 && (
            <span className="text-xs font-mono text-primary font-medium mr-1">
              {selectedCount} selected
            </span>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={() => exportFullSIEM('ndjson')}
            className="text-xs gap-1.5"
            title="Download full streaming newline-delimited JSON audit logs for SIEM ingestion"
          >
            <Download className="w-3.5 h-3.5 text-primary" />
            <span>Export SIEM (NDJSON)</span>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={exportToJson}
            disabled={table.getFilteredRowModel().rows.length === 0}
            className="text-xs gap-1.5"
          >
            <Download className="w-3.5 h-3.5" />
            <span>{selectedCount > 0 ? `Export JSON (${selectedCount})` : 'Export JSON'}</span>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={exportToCsv}
            disabled={table.getFilteredRowModel().rows.length === 0}
            className="text-xs gap-1.5"
          >
            <Download className="w-3.5 h-3.5" />
            <span>{selectedCount > 0 ? `Export CSV (${selectedCount})` : 'Export CSV'}</span>
          </Button>
        </div>
      </div>

      {/* Filter & Search Bar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
        {/* Severity Tabs */}
        <div className="flex items-center gap-1 p-0.5 bg-secondary border border-border rounded-md text-xs w-full sm:w-auto">
          {(['all', 'info', 'warning', 'critical'] as const).map((sev) => (
            <button
              key={sev}
              type="button"
              onClick={() => {
                setSeverityFilter(sev)
                table.setPageIndex(0)
              }}
              className={`px-2.5 py-1 rounded capitalize transition-colors cursor-pointer ${
                severityFilter === sev
                  ? 'bg-background text-foreground font-semibold shadow-2xs'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {sev === 'warning' ? 'Warnings' : sev}
            </button>
          ))}
        </div>

        {/* Search Input */}
        <div className="relative w-full sm:max-w-xs">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search action, actor, IP, details..."
            value={globalFilter ?? ''}
            onChange={(e) => {
              setGlobalFilter(e.target.value)
              table.setPageIndex(0)
            }}
            className="w-full h-8 pl-8 pr-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
          />
        </div>
      </div>

      {/* Events Table */}
      <Box>
        <BoxHeader className="bg-secondary/35 py-2.5 flex items-center justify-between">
          <span className="font-semibold text-foreground text-xs">Audit Events Stream</span>
          <span className="text-muted-foreground font-mono text-[11px]">
            {table.getFilteredRowModel().rows.length} {table.getFilteredRowModel().rows.length === 1 ? 'event' : 'events'}
          </span>
        </BoxHeader>

        <BoxBody className="p-0 overflow-x-auto">
          {loading ? (
            <div className="p-8 text-center text-xs text-muted-foreground font-mono">
              Loading audit stream...
            </div>
          ) : table.getRowModel().rows.length === 0 ? (
            <div className="p-8 text-center space-y-2">
              <ShieldAlert className="w-6 h-6 text-muted-foreground mx-auto" />
              <h4 className="text-xs font-semibold text-foreground">No Security Audit Events</h4>
              <p className="text-xs text-muted-foreground max-w-sm mx-auto">
                No audit logs recorded matching your filter. Live events stream automatically when users or autonomous agents interact with the forge.
              </p>
            </div>
          ) : (
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
                          className="px-3 py-2 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider select-none"
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
                {table.getRowModel().rows.map((row) => (
                  <tr
                    key={row.id}
                    className={`hover:bg-secondary/20 transition-colors ${
                      row.getIsSelected() ? 'bg-primary/5' : ''
                    }`}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-3 py-2.5 align-top">
                        {typeof cell.column.columnDef.cell === 'function'
                          ? cell.column.columnDef.cell(cell.getContext())
                          : (cell.getValue() as React.ReactNode)}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </BoxBody>

        {/* Pagination Bar */}
        {!loading && table.getPageCount() > 1 && (
          <div className="flex items-center justify-between px-3 py-2 border-t border-border bg-secondary/15 text-xs text-muted-foreground">
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
    </div>
  )
}
