'use client'

import Link from 'next/link'
import { useMemo, useState } from 'react'
import { Menu, ShieldCheck, Bell, Building2, Users, X, CheckCheck } from 'lucide-react'
import { SuperAdminSidebar } from '@/components/layout/SuperAdminSidebar'
import { useSystemStats } from '@/lib/queries/super-admin.queries'
import { ReAuthModal } from '@/components/auth/ReAuthModal'

export default function SuperAdminLayout({ children }: { children: React.ReactNode }) {
  const [mobileNavOpen, setMobileNavOpen] = useState(false)
  const [notificationsOpen, setNotificationsOpen] = useState(false)
  const [dismissed, setDismissed] = useState<Set<string>>(new Set())
  const { data: stats } = useSystemStats()
  const notifications = useMemo(() => [
    { id: 'platform-health', title: 'Platform is operational', body: 'API, access controls and data isolation are healthy.', href: '/super-admin/dashboard', icon: ShieldCheck },
    { id: 'organisations', title: `${stats?.total_orgs ?? 0} active organisations`, body: 'Review workspaces, create a new organisation or manage access.', href: '/super-admin/orgs', icon: Building2 },
    { id: 'users', title: `${stats?.total_users ?? 0} platform accounts`, body: 'Search and filter users across every organisation.', href: '/super-admin/users', icon: Users },
  ].filter((item) => !dismissed.has(item.id)), [dismissed, stats?.total_orgs, stats?.total_users])

  return (
    <div className="sa-shell flex min-h-screen overflow-x-hidden bg-slate-100 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
      {mobileNavOpen && (
        <button
          aria-label="Close navigation menu"
          className="fixed inset-0 z-40 bg-[#171714]/30 backdrop-blur-[1px] lg:hidden"
          onClick={() => setMobileNavOpen(false)}
        />
      )}
      <SuperAdminSidebar mobileOpen={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />
      <main className="min-w-0 flex-1 overflow-y-auto">
        <header className="flex h-[72px] items-center justify-between border-b border-slate-200 bg-white px-4 sm:px-6 lg:px-9 dark:border-slate-800 dark:bg-slate-900">
          <div className="flex items-center gap-3">
            <button aria-label="Open navigation menu" className="rounded-xl border border-slate-200 bg-white p-2.5 text-slate-600 shadow-sm lg:hidden" onClick={() => setMobileNavOpen(true)}>
              <Menu className="h-4 w-4" />
            </button>
            <div className="hidden items-center gap-2 text-sm font-medium text-slate-500 sm:flex dark:text-slate-300">
              <ShieldCheck className="h-4 w-4 text-blue-600" /> System administration
            </div>
          </div>
          <div className="relative">
            <button aria-label="Notifications" aria-expanded={notificationsOpen} onClick={() => setNotificationsOpen((open) => !open)} className="relative rounded-xl border border-slate-200 bg-white p-2.5 text-slate-500 transition hover:bg-slate-50 hover:text-slate-900 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700">
              <Bell className="h-4 w-4" />
              {notifications.length > 0 && <span className="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-cyan-500 ring-2 ring-white dark:ring-slate-900" />}
            </button>
            {notificationsOpen && <div id="notification-panel" className="absolute right-0 top-12 z-50 w-[min(23rem,calc(100vw-2rem))] overflow-hidden rounded-2xl border border-slate-200 bg-white text-slate-900 shadow-xl dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100">
              <div className="flex items-center justify-between border-b border-slate-100 px-4 py-3 dark:border-slate-800"><div><p className="text-sm font-semibold">Notifications</p><p className="text-xs text-slate-500">Platform activity and shortcuts</p></div>{notifications.length > 0 && <button onClick={() => setDismissed(new Set(['platform-health', 'organisations', 'users']))} className="inline-flex items-center gap-1 text-xs font-semibold text-blue-600 hover:text-blue-700"><CheckCheck className="h-3.5 w-3.5" /> Mark all read</button>}</div>
              {notifications.length === 0 ? <div className="px-4 py-8 text-center"><CheckCheck className="mx-auto h-6 w-6 text-emerald-500" /><p className="mt-2 text-sm font-semibold">You are up to date</p><button onClick={() => setDismissed(new Set())} className="mt-2 text-xs font-semibold text-blue-600">Restore notifications</button></div> : <div>{notifications.map((item) => { const Icon = item.icon; return <div key={item.id} className="flex gap-3 border-b border-slate-100 px-4 py-3 last:border-0 dark:border-slate-800"><div className="mt-0.5 rounded-lg bg-blue-50 p-2 text-blue-600 dark:bg-blue-950"><Icon className="h-4 w-4" /></div><Link href={item.href} onClick={() => setNotificationsOpen(false)} className="min-w-0 flex-1"><p className="text-sm font-semibold">{item.title}</p><p className="mt-0.5 text-xs leading-5 text-slate-500 dark:text-slate-300">{item.body}</p></Link><button aria-label={`Dismiss ${item.title}`} onClick={() => setDismissed((previous) => new Set(previous).add(item.id))} className="h-7 w-7 rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-700 dark:hover:bg-slate-800"><X className="mx-auto h-4 w-4" /></button></div> })}</div>}
            </div>}
          </div>
        </header>
        <div className="mx-auto max-w-[1440px] p-4 sm:p-6 lg:p-8">
          {children}
        </div>
      </main>
      <ReAuthModal />
    </div>
  )
}
