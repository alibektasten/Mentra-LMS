'use client'

import { useState } from 'react'
import { Menu, ShieldCheck, Bell } from 'lucide-react'
import { SuperAdminSidebar } from '@/components/layout/SuperAdminSidebar'

export default function SuperAdminLayout({ children }: { children: React.ReactNode }) {
  const [mobileNavOpen, setMobileNavOpen] = useState(false)

  return (
    <div className="sa-shell flex min-h-screen overflow-hidden bg-slate-100 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
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
          <button aria-label="Notifications" title="System notifications are coming soon" className="relative rounded-xl border border-slate-200 bg-white p-2.5 text-slate-500 transition hover:bg-slate-50 hover:text-slate-900 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700">
            <Bell className="h-4 w-4" />
            <span className="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-cyan-500" />
          </button>
        </header>
        <div className="mx-auto max-w-[1440px] p-4 sm:p-6 lg:p-8">
          {children}
        </div>
      </main>
    </div>
  )
}
