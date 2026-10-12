'use client'

import Link from 'next/link'
import { Activity, ArrowUpRight, Building2, CircleCheckBig, Plus, ShieldCheck, Users } from 'lucide-react'
import { useSystemStats } from '@/lib/queries/super-admin.queries'
import { PageHeader } from '@/components/shared/PageHeader'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

function StatCard({ title, value, caption, icon: Icon, tint, href }: {
  title: string
  value: string | number
  caption: string
  icon: React.ElementType
  tint: string
  href: string
}) {
  return (
    <Link href={href} className="group block rounded-2xl focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2">
    <Card className="sa-card h-full overflow-hidden rounded-2xl border-0 transition duration-200 group-hover:-translate-y-0.5 group-hover:border-blue-200 group-hover:shadow-lg">
      <CardHeader className="flex flex-row items-start justify-between pb-3">
        <div>
          <CardTitle className="text-xs font-bold uppercase tracking-[0.12em] text-slate-400">{title}</CardTitle>
          <div className="mt-3 text-3xl font-bold tracking-tight text-slate-800 dark:text-slate-100">{value}</div>
        </div>
        <div className={`rounded-xl p-2.5 ${tint}`}><Icon className="h-5 w-5" /></div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between gap-2"><p className="text-xs font-medium text-slate-400">{caption}</p><ArrowUpRight className="h-4 w-4 text-blue-600 opacity-0 transition group-hover:opacity-100" /></div>
      </CardContent>
    </Card></Link>
  )
}

export default function SuperAdminDashboard() {
  const { data: stats, isLoading } = useSystemStats()

  return (
    <div className="space-y-7">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <PageHeader title="Platform overview" description="Monitor organisations, access and overall platform health." />
        <Link href="/super-admin/orgs" className="inline-flex h-10 items-center gap-2 rounded-xl bg-blue-600 px-4 text-sm font-semibold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-700">
          <Plus className="h-4 w-4" /> New organization
        </Link>
      </div>
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3">
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-36 rounded-2xl" />)
        ) : (
          <>
            <StatCard title="Organizations" value={stats?.total_orgs ?? 0} caption="Active workspaces" icon={Building2} tint="bg-blue-50 text-blue-600" href="/super-admin/orgs" />
            <StatCard title="Platform users" value={stats?.total_users ?? 0} caption="Accounts across all tenants" icon={Users} tint="bg-violet-50 text-violet-600" href="/super-admin/users" />
            <StatCard title="System status" value="Healthy" caption="API and access controls operational" icon={Activity} tint="bg-emerald-50 text-emerald-600" href="/super-admin/dashboard#readiness" />
          </>
        )}
      </div>
      <div className="grid gap-5 lg:grid-cols-[1.45fr_1fr]">
        <Card id="readiness" className="sa-card rounded-2xl border-0">
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle className="text-base text-slate-800 dark:text-slate-100">Platform operations</CardTitle>
              <p className="mt-1 text-sm text-slate-400">Common administrative workflows</p>
            </div>
            <ShieldCheck className="h-5 w-5 text-blue-600" />
          </CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-2">
            <Link href="/super-admin/orgs" className="group rounded-xl border border-slate-100 bg-slate-50/70 p-4 transition hover:-translate-y-0.5 hover:border-blue-200 hover:bg-blue-50/60 dark:border-slate-700 dark:bg-slate-800/60">
              <Building2 className="h-5 w-5 text-blue-600" />
              <p className="mt-4 font-semibold text-slate-800 dark:text-slate-100">Manage organizations</p>
              <p className="mt-1 text-xs leading-5 text-slate-400">Create workspaces and review tenant access.</p>
              <ArrowUpRight className="mt-3 h-4 w-4 text-blue-600 transition group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
            </Link>
            <Link href="/super-admin/invite" className="group rounded-xl border border-slate-100 bg-slate-50/70 p-4 transition hover:-translate-y-0.5 hover:border-violet-200 hover:bg-violet-50/60 dark:border-slate-700 dark:bg-slate-800/60">
              <Users className="h-5 w-5 text-violet-600" />
              <p className="mt-4 font-semibold text-slate-800 dark:text-slate-100">Assign an org admin</p>
              <p className="mt-1 text-xs leading-5 text-slate-400">Give a tenant a responsible administrator.</p>
              <ArrowUpRight className="mt-3 h-4 w-4 text-violet-600 transition group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
            </Link>
          </CardContent>
        </Card>
        <Card className="sa-card rounded-2xl border-0">
          <CardHeader>
            <CardTitle className="text-base text-slate-800 dark:text-slate-100">Readiness</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {['Role-based access enabled', 'Tenant data isolation enabled', 'Health endpoint available'].map((item) => (
              <div key={item} className="flex items-center gap-3 text-sm text-slate-600 dark:text-slate-300">
                <CircleCheckBig className="h-4 w-4 shrink-0 text-emerald-500" /> {item}
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
