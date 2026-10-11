'use client'

import { useRouter } from 'next/navigation'
import { LayoutDashboard, Building2, Users, UserPlus, LogOut, X } from 'lucide-react'
import { useAuthStore } from '@/lib/stores/auth.store'
import { NavItem } from './NavItem'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'

const navItems = [
  { href: '/super-admin/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { href: '/super-admin/orgs', label: 'Organizations', icon: Building2 },
  { href: '/super-admin/users', label: 'All Users', icon: Users },
  { href: '/super-admin/invite', label: 'Invite Admin', icon: UserPlus },
]

interface SuperAdminSidebarProps {
  mobileOpen?: boolean
  onClose?: () => void
}

export function SuperAdminSidebar({ mobileOpen = false, onClose }: SuperAdminSidebarProps) {
  const router = useRouter()
  const { user, clearSession } = useAuthStore()

  const initials = user
    ? `${user.first_name[0] ?? ''}${user.last_name[0] ?? ''}`.toUpperCase()
    : 'SA'

  const handleLogout = () => {
    clearSession()
    router.push('/login')
  }

  return (
    <aside className={`fixed inset-y-0 left-0 z-50 flex h-screen w-[264px] shrink-0 flex-col border-r border-[#e7e6e0] bg-[#fbfbf9] shadow-2xl transition-transform duration-300 lg:static lg:z-auto lg:translate-x-0 lg:shadow-none ${mobileOpen ? 'translate-x-0' : '-translate-x-full'}`}>
      <div className="flex h-[72px] items-center gap-3 px-5">
        <div className="flex h-8 w-8 items-center justify-center rounded-[10px] bg-gradient-to-br from-[#111110] to-[#3a3a35] text-white text-xs font-bold shadow-[0_5px_12px_rgba(17,17,16,0.16)]">
          M
        </div>
        <div>
          <span className="block text-[15px] font-semibold tracking-[-0.025em] text-[#171714]">Mentra</span>
          <span className="block text-[10px] font-medium uppercase tracking-[0.14em] text-[#a09f99]">Platform control</span>
        </div>
        <button aria-label="Close navigation menu" onClick={onClose} className="ml-auto rounded-lg p-2 text-[#77766f] hover:bg-[#f1f0ec] lg:hidden">
          <X className="h-4 w-4" />
        </button>
      </div>

      <nav className="flex-1 overflow-y-auto px-3 py-4" onClick={onClose}>
        <p className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-[0.14em] text-[#aaa9a2]">Administration</p>
        {navItems.map((item) => (
          <NavItem key={item.href} {...item} />
        ))}
      </nav>

      <div className="space-y-1 border-t border-[#e7e6e0] p-3">
        <div className="flex items-center gap-2.5 rounded-xl px-2.5 py-2">
          <Avatar className="h-8 w-8 shrink-0">
            <AvatarFallback className="bg-[#dff4e9] text-[#087252] text-[10px] font-bold">
              {initials}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <span className="block truncate text-xs font-semibold text-[#393934]">{user?.first_name} {user?.last_name}</span>
            <span className="block truncate text-[10px] text-[#96958e]">Platform administrator</span>
          </div>
        </div>
        <button
          onClick={handleLogout}
          className="flex w-full items-center gap-2 rounded-xl px-2.5 py-2 text-xs font-medium text-[#77766f] transition-colors hover:bg-[#f1f0ec] hover:text-[#282824]"
        >
          <LogOut className="h-4 w-4 shrink-0" />
          Sign out
        </button>
      </div>
    </aside>
  )
}
