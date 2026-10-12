'use client'

import { useDeferredValue, useEffect, useState } from 'react'
import { ArrowDown, ArrowUp, Building2, Search, SlidersHorizontal, UserRound, X } from 'lucide-react'
import { useAdminOrgs, useAllUsers } from '@/lib/queries/super-admin.queries'
import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent } from '@/components/ui/card'
import type { SystemUserFilters } from '@/types/super-admin'

const roleLabels: Record<string, string> = { super_admin: 'System admin', admin: 'Organisation admin', teacher: 'Teacher', student: 'Student', unassigned: 'Unassigned' }
const roleStyles: Record<string, string> = { super_admin: 'bg-violet-100 text-violet-700', admin: 'bg-blue-100 text-blue-700', teacher: 'bg-amber-100 text-amber-700', student: 'bg-emerald-100 text-emerald-700', unassigned: 'bg-slate-100 text-slate-600' }

export default function SuperAdminUsersPage() {
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [role, setRole] = useState('')
  const [orgID, setOrgID] = useState('')
  const [sortBy, setSortBy] = useState<NonNullable<SystemUserFilters['sort_by']>>('created_at')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const deferredSearch = useDeferredValue(search)
  const filters: SystemUserFilters = { page, page_size: 20, search: deferredSearch || undefined, role: role || undefined, org_id: orgID || undefined, sort_by: sortBy, sort_order: sortOrder }
  const { data: orgsData } = useAdminOrgs({ page: 1, page_size: 100 })
  const { data, isLoading, isFetching } = useAllUsers(filters)
  useEffect(() => { setPage(1) }, [deferredSearch, role, orgID, sortBy, sortOrder])

  const users = data?.data ?? []
  const total = data?.meta?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / 20))
  const hasFilters = Boolean(search || role || orgID)
  const toggleSort = (field: NonNullable<SystemUserFilters['sort_by']>) => {
    if (field === sortBy) setSortOrder((value) => value === 'asc' ? 'desc' : 'asc')
    else { setSortBy(field); setSortOrder(field === 'created_at' ? 'desc' : 'asc') }
  }
  const sortIndicator = (field: NonNullable<SystemUserFilters['sort_by']>) => field !== sortBy ? null : sortOrder === 'asc' ? <ArrowUp className="h-3.5 w-3.5" /> : <ArrowDown className="h-3.5 w-3.5" />

  return <div className="space-y-6">
    <PageHeader title="All users" description={`${total} account${total === 1 ? '' : 's'} across every organisation`} />
    <Card className="sa-card rounded-2xl border-0"><CardContent className="p-4 sm:p-5">
      <div className="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center">
        <div className="relative min-w-0 flex-1"><Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" /><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search name or email…" className="h-10 w-full rounded-xl border border-slate-200 bg-white pl-9 pr-3 text-sm outline-none transition focus:border-blue-400 focus:ring-4 focus:ring-blue-100 dark:border-slate-700 dark:bg-slate-900" /></div>
        <div className="flex flex-col gap-2 sm:flex-row">
          <select value={role} onChange={(event) => setRole(event.target.value)} aria-label="Filter by role" className="h-10 rounded-xl border border-slate-200 bg-white px-3 text-sm text-slate-700 outline-none focus:border-blue-400 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"><option value="">All roles</option><option value="super_admin">System admin</option><option value="admin">Organisation admin</option><option value="teacher">Teacher</option><option value="student">Student</option></select>
          <select value={orgID} onChange={(event) => setOrgID(event.target.value)} aria-label="Filter by organisation" className="h-10 rounded-xl border border-slate-200 bg-white px-3 text-sm text-slate-700 outline-none focus:border-blue-400 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"><option value="">All organisations</option>{(orgsData?.data ?? []).map((org) => <option key={org.id} value={org.id}>{org.name}</option>)}</select>
          {hasFilters && <Button variant="outline" size="sm" className="h-10 rounded-xl px-3" onClick={() => { setSearch(''); setRole(''); setOrgID('') }}><X className="h-4 w-4" /> Clear</Button>}
        </div>
      </div>
      <div className="flex items-center gap-2 border-t border-slate-100 pt-3 text-xs text-slate-500 dark:border-slate-800"><SlidersHorizontal className="h-3.5 w-3.5" /> Filters apply across all pages {isFetching && <span className="text-blue-600">Updating…</span>}</div>
    </CardContent></Card>
    <Card className="sa-card overflow-hidden rounded-2xl border-0"><CardContent className="p-0">
      {isLoading ? <div className="space-y-2 p-4">{Array.from({ length: 5 }).map((_, index) => <Skeleton key={index} className="h-12" />)}</div> : <div className="overflow-x-auto"><table className="w-full min-w-[860px] text-sm"><thead><tr className="border-b border-slate-100 bg-slate-50/70 text-left text-[11px] font-bold uppercase tracking-[0.1em] text-slate-400 dark:border-slate-700 dark:bg-slate-800/60">{([['name', 'Account'], ['email', 'Email'], ['role', 'Role'], ['org_name', 'Organisation'], ['created_at', 'Created']] as const).map(([field, label]) => <th key={field} className="px-5 py-4"><button onClick={() => toggleSort(field)} className="inline-flex items-center gap-1.5 transition hover:text-slate-700 dark:hover:text-white">{label}{sortIndicator(field)}</button></th>)}</tr></thead><tbody>
        {users.length === 0 ? <tr><td colSpan={5} className="px-4 py-12 text-center text-slate-500">No accounts match these filters.</td></tr> : users.map((user) => <tr key={`${user.id}-${user.org_id}`} className="border-b border-slate-100 last:border-0 transition hover:bg-blue-50/50 dark:border-slate-700 dark:hover:bg-slate-800/50"><td className="px-5 py-4"><div className="flex items-center gap-3"><span className="grid h-9 w-9 place-items-center rounded-full bg-violet-100 text-xs font-bold text-violet-600">{user.name.charAt(0).toUpperCase() || <UserRound className="h-4 w-4" />}</span><span className="font-semibold text-slate-700 dark:text-slate-100">{user.name}</span></div></td><td className="px-5 py-4 text-slate-500 dark:text-slate-300">{user.email}</td><td className="px-5 py-4"><span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${roleStyles[user.role] ?? roleStyles.unassigned}`}>{roleLabels[user.role] ?? user.role}</span></td><td className="px-5 py-4"><span className="inline-flex items-center gap-1.5 text-slate-600 dark:text-slate-300"><Building2 className="h-3.5 w-3.5 text-slate-400" />{user.org_name}</span></td><td className="px-5 py-4 text-slate-400">{new Date(user.created_at).toLocaleDateString()}</td></tr>)}
      </tbody></table></div>}
    </CardContent></Card>
    {totalPages > 1 && <div className="flex items-center justify-end gap-2"><Button variant="outline" size="sm" className="rounded-lg" disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>Previous</Button><span className="text-sm text-slate-500">Page {page} of {totalPages}</span><Button variant="outline" size="sm" className="rounded-lg" disabled={page >= totalPages} onClick={() => setPage((value) => value + 1)}>Next</Button></div>}
  </div>
}
