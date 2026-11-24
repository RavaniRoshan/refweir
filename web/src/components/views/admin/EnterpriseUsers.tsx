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
  Users,
  UserPlus,
  Mail,
  RefreshCw,
  Search,
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
import { Dialog } from '@/components/ui/Dialog'
import { Input } from '@/components/ui/Input'
import { LivingAvatar } from '@/components/ui/LivingAvatar'
import { getEnterpriseUsers, createEnterpriseUser, type EnterpriseUser } from '@/lib/api'

export const EnterpriseUsers: React.FC = () => {
  const [users, setUsers] = useState<EnterpriseUser[]>([])
  const [loading, setLoading] = useState(true)
  const [globalFilter, setGlobalFilter] = useState('')
  const [roleFilter, setRoleFilter] = useState<'all' | 'admin' | 'contributor' | 'member'>('all')
  const [sorting, setSorting] = useState<SortingState>([])
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({})

  const [inviteModalOpen, setInviteModalOpen] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Form fields
  const [email, setEmail] = useState('')
  const [login, setLogin] = useState('')
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<'admin' | 'member'>('member')

  useEffect(() => {
    loadUsers()
  }, [])

  const loadUsers = async () => {
    setLoading(true)
    const list = await getEnterpriseUsers()
    setUsers(list)
    setLoading(false)
  }

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email.trim() || !password.trim()) {
      setErrorMsg('Email and password are required')
      return
    }
    if (password.length < 8) {
      setErrorMsg('Password must be at least 8 characters long')
      return
    }

    setSubmitting(true)
    setErrorMsg(null)

    const res = await createEnterpriseUser({
      email: email.trim(),
      login: (login || email.split('@')[0]).trim(),
      name: (name || login || email.split('@')[0]).trim(),
      password: password,
      role: role,
    })

    setSubmitting(false)
    if (res.ok) {
      setInviteModalOpen(false)
      setEmail('')
      setLogin('')
      setName('')
      setPassword('')
      setRole('member')
      loadUsers()
    } else {
      setErrorMsg(res.error || 'Failed to create user')
    }
  }

  // Filtered dataset for TanStack Table
  const data = useMemo(() => {
    if (roleFilter === 'all') return users
    return users.filter((u) => u.role === roleFilter)
  }, [users, roleFilter])

  const columns = useMemo<ColumnDef<EnterpriseUser>[]>(
    () => [
      {
        id: 'select',
        header: ({ table }) => (
          <button
            type="button"
            onClick={table.getToggleAllRowsSelectedHandler()}
            className="text-muted-foreground hover:text-foreground cursor-pointer flex items-center justify-center p-0.5"
            aria-label="Select all members"
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
            aria-label={`Select ${row.original.name}`}
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
        accessorKey: 'name',
        header: 'User',
        cell: ({ row }) => {
          const u = row.original
          return (
            <div className="flex items-center gap-3">
              <LivingAvatar name={u.name} size="md" />
              <div className="space-y-0.5">
                <div className="flex items-center gap-1.5">
                  <span className="text-sm font-semibold text-foreground">{u.name}</span>
                  <span className="text-xs text-muted-foreground font-mono">@{u.username}</span>
                </div>
                <div className="text-[11px] font-mono text-muted-foreground">
                  ID: {u.id}
                </div>
              </div>
            </div>
          )
        },
      },
      {
        accessorKey: 'role',
        header: 'Role',
        cell: ({ row }) => {
          const r = row.original.role
          return (
            <Badge
              variant={r === 'admin' ? 'agent' : 'outline'}
              className="text-[10px] py-0 capitalize"
            >
              {r}
            </Badge>
          )
        },
      },
      {
        accessorKey: 'status',
        header: 'Status',
        cell: ({ row }) => {
          const s = row.original.status
          return (
            <Badge
              variant={s === 'active' ? 'success' : 'closed'}
              className="text-[10px] py-0 capitalize"
            >
              {s}
            </Badge>
          )
        },
      },
      {
        accessorKey: 'email',
        header: 'Email',
        cell: ({ row }) => (
          <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <Mail className="w-3 h-3 text-muted-foreground" />
            <span>{row.original.email}</span>
          </div>
        ),
      },
      {
        accessorKey: 'lastActive',
        header: 'Activity',
        cell: ({ row }) => (
          <span className="text-[11px] text-muted-foreground whitespace-nowrap">
            {row.original.lastActive}
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
      const u = row.original
      return (
        u.name.toLowerCase().includes(q) ||
        u.email.toLowerCase().includes(q) ||
        u.username.toLowerCase().includes(q)
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

  const selectedCount = Object.keys(rowSelection).length

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-border">
        <div className="space-y-0.5">
          <div className="flex items-center gap-2">
            <h2 className="text-base font-semibold text-foreground flex items-center gap-2">
              <Users className="w-4 h-4 text-primary" />
              <span>Enterprise User Management & RBAC</span>
            </h2>
            <Badge variant="outline" className="text-[10px]">
              {users.length} {users.length === 1 ? 'account' : 'accounts'}
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground">
            Manage developer memberships, access roles, and platform administrator privileges. Powered by TanStack Table.
          </p>
        </div>

        <div className="flex items-center gap-2">
          {selectedCount > 0 && (
            <span className="text-xs font-mono text-primary font-medium mr-1">
              {selectedCount} selected
            </span>
          )}
          <Button variant="outline" size="sm" onClick={loadUsers} disabled={loading} className="gap-1.5 text-xs">
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </Button>
          <Button variant="primary" size="sm" onClick={() => setInviteModalOpen(true)} className="gap-1.5 text-xs">
            <UserPlus className="w-3.5 h-3.5" />
            <span>Add User</span>
          </Button>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
        {/* Role Filters */}
        <div className="flex items-center gap-1 p-0.5 bg-secondary border border-border rounded-md text-xs w-full sm:w-auto">
          {(['all', 'admin', 'contributor', 'member'] as const).map((r) => (
            <button
              key={r}
              type="button"
              onClick={() => {
                setRoleFilter(r)
                table.setPageIndex(0)
              }}
              className={`px-2.5 py-1 rounded capitalize transition-colors cursor-pointer ${
                roleFilter === r
                  ? 'bg-background text-foreground font-semibold shadow-2xs'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {r === 'all' ? 'All Roles' : r}
            </button>
          ))}
        </div>

        {/* Search Input */}
        <div className="relative flex-1 max-w-sm w-full">
          <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search users by name, username, or email..."
            value={globalFilter ?? ''}
            onChange={(e) => {
              setGlobalFilter(e.target.value)
              table.setPageIndex(0)
            }}
            className="w-full h-8 pl-8 pr-3 text-xs bg-input border border-border rounded-md text-foreground placeholder:text-muted-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
          />
        </div>
      </div>

      {/* Users Table */}
      <Box>
        <BoxHeader className="bg-secondary/35 py-2.5 flex items-center justify-between">
          <span className="font-semibold text-foreground text-xs">Registered Members</span>
          <span className="text-muted-foreground font-mono text-[11px]">
            {table.getFilteredRowModel().rows.length} {table.getFilteredRowModel().rows.length === 1 ? 'account' : 'accounts'}
          </span>
        </BoxHeader>

        <BoxBody className="p-0 overflow-x-auto">
          {loading ? (
            <div className="p-8 text-center text-xs text-muted-foreground font-mono flex items-center justify-center gap-2">
              <RefreshCw className="w-4 h-4 animate-spin text-primary" />
              <span>Loading user accounts...</span>
            </div>
          ) : table.getRowModel().rows.length === 0 ? (
            <div className="p-8 text-center space-y-3">
              <Users className="w-6 h-6 text-muted-foreground mx-auto" />
              <h4 className="text-xs font-semibold text-foreground">No users found</h4>
              <p className="text-xs text-muted-foreground max-w-sm mx-auto">
                {globalFilter ? 'No accounts matched your search query.' : 'Create an account or connect your SCIM/SSO provider to sync corporate users.'}
              </p>
              {!globalFilter && (
                <Button size="sm" variant="primary" onClick={() => setInviteModalOpen(true)} className="mt-2 text-xs">
                  Add First User
                </Button>
              )}
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
                {table.getRowModel().rows.map((row) => (
                  <tr
                    key={row.id}
                    className={`hover:bg-secondary/20 transition-colors ${
                      row.getIsSelected() ? 'bg-primary/5' : ''
                    }`}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-3.5 py-3 align-middle">
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

        {/* Pagination */}
        {!loading && table.getPageCount() > 1 && (
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

      {/* Add User Modal */}
      <Dialog
        isOpen={inviteModalOpen}
        onClose={() => setInviteModalOpen(false)}
        title="Add Enterprise Member"
        description="Create an authenticated user account with role-based access controls."
      >
        <form onSubmit={handleCreateUser} className="space-y-3 pt-2">
          {errorMsg && (
            <div className="p-2.5 rounded-md bg-destructive/10 border border-destructive/30 text-destructive text-xs">
              {errorMsg}
            </div>
          )}

          <div className="space-y-1">
            <label className="text-xs font-medium text-foreground">Email Address *</label>
            <Input
              type="email"
              placeholder="developer@enterprise.internal"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value)
                if (!login && e.target.value.includes('@')) {
                  setLogin(e.target.value.split('@')[0])
                }
              }}
              required
              className="text-xs"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1">
              <label className="text-xs font-medium text-foreground">Username *</label>
              <Input
                placeholder="alice"
                value={login}
                onChange={(e) => setLogin(e.target.value)}
                required
                className="text-xs font-mono"
              />
            </div>
            <div className="space-y-1">
              <label className="text-xs font-medium text-foreground">Full Name</label>
              <Input
                placeholder="Alice Walker"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="text-xs"
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-foreground">Initial Password *</label>
            <Input
              type="password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="text-xs font-mono"
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-medium text-foreground">Access Role</label>
            <select
              value={role}
              onChange={(e) => setRole(e.target.value as any)}
              className="w-full h-8 px-2.5 text-xs bg-input border border-border rounded-md text-foreground outline-none focus:border-ring focus:ring-1 focus:ring-ring"
            >
              <option value="member">Member (Repository developer & code author)</option>
              <option value="admin">Platform Administrator (Full tenant & forge authority)</option>
            </select>
          </div>

          <div className="pt-2 flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setInviteModalOpen(false)}
              className="text-xs"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="sm"
              disabled={submitting}
              className="text-xs gap-1.5"
            >
              {submitting ? 'Creating...' : 'Create Account'}
            </Button>
          </div>
        </form>
      </Dialog>
    </div>
  )
}
