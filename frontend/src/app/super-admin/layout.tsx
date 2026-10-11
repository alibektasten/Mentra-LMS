'use client'

import { useState } from 'react'
import { Menu } from 'lucide-react'
import { SuperAdminSidebar } from '@/components/layout/SuperAdminSidebar'

export default function SuperAdminLayout({ children }: { children: React.ReactNode }) {
  const [mobileNavOpen, setMobileNavOpen] = useState(false)

  return (
    <div className="app-canvas flex min-h-screen overflow-hidden">
      {mobileNavOpen && (
        <button
          aria-label="Close navigation menu"
          className="fixed inset-0 z-40 bg-[#171714]/30 backdrop-blur-[1px] lg:hidden"
          onClick={() => setMobileNavOpen(false)}
        />
      )}
      <SuperAdminSidebar mobileOpen={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />
      <main className="flex-1 overflow-y-auto px-4 py-5 sm:px-6 lg:px-8 lg:py-7">
        <div className="content-surface mx-auto min-h-full max-w-[1440px] rounded-2xl border border-[#ebeae5] bg-white/70 p-5 backdrop-blur-[2px] sm:p-7 lg:p-8">
          <button
            aria-label="Open navigation menu"
            className="mb-5 rounded-xl border border-[#e7e6e0] bg-white p-2.5 text-[#77766f] shadow-sm hover:bg-[#f7f6f2] lg:hidden"
            onClick={() => setMobileNavOpen(true)}
          >
            <Menu className="h-4 w-4" />
          </button>
          {children}
        </div>
      </main>
    </div>
  )
}
