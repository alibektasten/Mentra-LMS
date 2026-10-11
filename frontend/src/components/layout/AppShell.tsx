'use client'

import { useState } from 'react'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'

interface AppShellProps {
  children: React.ReactNode
}

export function AppShell({ children }: AppShellProps) {
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
      <Sidebar mobileOpen={mobileNavOpen} onClose={() => setMobileNavOpen(false)} />
      <div className="flex min-w-0 flex-1 flex-col overflow-hidden">
        <TopBar onMenuClick={() => setMobileNavOpen(true)} />
        <main className="flex-1 overflow-y-auto px-4 py-5 sm:px-6 lg:px-8 lg:py-7">
          <div className="content-surface mx-auto min-h-full max-w-[1440px] rounded-2xl border border-[#ebeae5] bg-white/70 p-5 backdrop-blur-[2px] sm:p-7 lg:p-8">
            {children}
          </div>
        </main>
      </div>
    </div>
  )
}
