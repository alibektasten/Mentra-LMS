'use client'

import { useState } from 'react'
import { Menu, Search, Bell } from 'lucide-react'
import { SuperAdminSidebar } from '@/components/layout/SuperAdminSidebar'

export default function SuperAdminLayout({ children }: { children: React.ReactNode }) {
  const [mobileNavOpen, setMobileNavOpen] = useState(false)

  return (
    <div className="sa-shell flex min-h-screen overflow-hidden">
      {mobileNavOpen && (
        <button
          aria-label="Close navigation menu"
          className="fixed inset-0 z-40 bg-[#171714]/30 backdrop-blur-[1px] lg:hidden"
          onClick={() => setMobileNavOpen(false)}
        />
      )}
      <SuperAdminSidebar mobileOpen={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />
      <main className="flex-1 overflow-y-auto">
        <header className="flex h-[84px] items-center justify-between border-b border-slate-200/80 bg-white/80 px-4 backdrop-blur-xl sm:px-6 lg:px-9">
          <div className="flex items-center gap-3">
            <button aria-label="Open navigation menu" className="rounded-xl border border-slate-200 bg-white p-2.5 text-slate-600 shadow-sm lg:hidden" onClick={() => setMobileNavOpen(true)}>
              <Menu className="h-4 w-4" />
            </button>
            <div className="hidden items-center gap-2 rounded-xl bg-slate-100 px-3 py-2 text-xs text-slate-400 sm:flex">
              <Search className="h-3.5 w-3.5" /> Search platform
              <kbd className="ml-8 rounded border border-slate-200 bg-white px-1.5 py-0.5 text-[10px]">⌘ K</kbd>
            </div>
          </div>
          <button aria-label="Notifications" className="relative rounded-xl border border-slate-200 bg-white p-2.5 text-slate-500 shadow-sm">
            <Bell className="h-4 w-4" />
            <span className="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-cyan-500" />
          </button>
        </header>
        <div className="mx-auto max-w-[1520px] p-4 sm:p-6 lg:p-9">
          {children}
        </div>
      </main>
    </div>
  )
}
