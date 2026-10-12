'use client'

import { useMemo, useState } from 'react'
import { ArrowDownAZ, ArrowUpAZ, CalendarArrowDown, Mail, UserRound } from 'lucide-react'
import { useAllUsers } from '@/lib/queries/super-admin.queries'
import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent } from '@/components/ui/card'

export default function SuperAdminUsersPage() {
  const [page, setPage] = useState(1)
  const [sortBy, setSortBy] = useState<'name' | 'email' | 'created_at'>('created_at')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc')
  const pageSize = 20
  const { data, isLoading } = useAllUsers({ page, page_size: pageSize })

  const users = useMemo(() => [...(data?.data ?? [])].sort((a, b) => {
    const aValue = sortBy === 'created_at' ? new Date(a.created_at).getTime() : a[sortBy].toLocaleLowerCase()
    const bValue = sortBy === 'created_at' ? new Date(b.created_at).getTime() : b[sortBy].toLocaleLowerCase()
    const comparison = aValue < bValue ? -1 : aValue > bValue ? 1 : 0
    return sortDirection === 'asc' ? comparison : -comparison
  }), [data?.data, sortBy, sortDirection])
  const total = data?.meta?.total ?? 0
  const totalPages = Math.ceil(total / pageSize)
  const changeSort = (field: typeof sortBy) => {
    if (field === sortBy) setSortDirection((direction) => direction === 'asc' ? 'desc' : 'asc')
    else { setSortBy(field); setSortDirection(field === 'created_at' ? 'desc' : 'asc') }
  }
  const sortIcon = (field: typeof sortBy) => field === sortBy
    ? (sortDirection === 'asc' ? <ArrowUpAZ className="h-3.5 w-3.5" /> : <ArrowDownAZ className="h-3.5 w-3.5" />)
    : field === 'created_at' ? <CalendarArrowDown className="h-3.5 w-3.5" /> : field === 'email' ? <Mail className="h-3.5 w-3.5" /> : <ArrowDownAZ className="h-3.5 w-3.5" />

  return (
    <div className="space-y-7">
      <PageHeader title="All users" description={`${total} accounts across every organisation`} />
      <Card className="sa-card overflow-hidden rounded-2xl border-0">
        <CardContent className="p-0">
          {isLoading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
            </div>
          ) : (
            <div className="overflow-x-auto">
            <table className="w-full min-w-[640px] text-sm">
              <thead>
                <tr className="border-b border-slate-100 bg-slate-50/70 text-left text-[11px] font-bold uppercase tracking-[0.1em] text-slate-400 dark:border-slate-700 dark:bg-slate-800/60">
                  {([['name', 'Account'], ['email', 'Email'], ['created_at', 'Created']] as const).map(([field, label]) => (
                    <th key={field} className="px-5 py-4">
                      <button onClick={() => changeSort(field)} className="inline-flex items-center gap-1.5 transition hover:text-slate-700 dark:hover:text-white" title={`Sort by ${label.toLowerCase()}`}>
                        {label} {sortIcon(field)}
                      </button>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {users.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="px-4 py-8 text-center text-ink-muted">No users found</td>
                  </tr>
                ) : (
                  users.map((user) => (
                    <tr key={user.id} className="border-b border-slate-100 last:border-0 transition hover:bg-violet-50/50 dark:border-slate-700 dark:hover:bg-slate-800/50">
                      <td className="px-5 py-4"><div className="flex items-center gap-3"><span className="grid h-9 w-9 place-items-center rounded-full bg-violet-100 text-xs font-bold text-violet-600">{user.name.charAt(0).toUpperCase() || <UserRound className="h-4 w-4" />}</span><span className="font-semibold text-slate-700 dark:text-slate-100">{user.name}</span></div></td>
                      <td className="px-5 py-4 text-slate-500 dark:text-slate-300">{user.email}</td>
                      <td className="px-5 py-4 text-slate-400">{new Date(user.created_at).toLocaleDateString()}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
            </div>
          )}
        </CardContent>
      </Card>
      {totalPages > 1 && (
        <div className="flex items-center justify-end gap-2">
          <Button variant="outline" size="sm" className="rounded-lg" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Previous
          </Button>
          <span className="text-sm text-slate-400">Page {page} of {totalPages}</span>
          <Button variant="outline" size="sm" className="rounded-lg" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            Next
          </Button>
        </div>
      )}
    </div>
  )
}
