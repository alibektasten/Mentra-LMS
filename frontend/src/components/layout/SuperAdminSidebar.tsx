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
    <aside className={`sa-sidebar fixed inset-y-0 left-0 z-50 flex h-[100dvh] w-[272px] shrink-0 flex-col !bg-slate-950 shadow-2xl transition-transform duration-300 lg:static lg:h-auto lg:min-h-screen lg:self-stretch lg:translate-x-0 lg:shadow-none ${mobileOpen ? 'translate-x-0' : '-translate-x-full'}`}>
      <div className="flex h-[84px] items-center gap-3 px-6">
        <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-cyan-400 to-blue-600 text-sm font-black text-white shadow-lg shadow-blue-950/30">
          M
        </div>
        <div>
          <span className="block text-[17px] font-semibold tracking-[-0.025em] text-white">Mentra</span>
          <span className="block text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-400">System console</span>
        </div>
        <button aria-label="Close navigation menu" onClick={onClose} className="ml-auto rounded-lg p-2 text-slate-400 hover:bg-white/10 hover:text-white lg:hidden">
          <X className="h-4 w-4" />
        </button>
      </div>

      <nav className="flex-1 overflow-y-auto px-3 py-5" onClick={onClose}>
        <p className="mb-3 px-3 text-[10px] font-bold uppercase tracking-[0.16em] text-slate-500">Workspace</p>
        {navItems.map((item) => (
          <NavItem key={item.href} {...item} />
        ))}
      </nav>

      <div className="space-y-2 border-t border-white/10 p-4">
        <div className="flex items-center gap-2.5 rounded-xl bg-white/5 px-2.5 py-2.5">
          <Avatar className="h-8 w-8 shrink-0">
            <AvatarFallback className="bg-cyan-400 text-slate-950 text-[10px] font-bold">
              {initials}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <span className="block truncate text-xs font-semibold text-white">{user?.first_name} {user?.last_name}</span>
            <span className="block truncate text-[10px] text-slate-400">Platform administrator</span>
          </div>
        </div>
        <button
          onClick={handleLogout}
          className="flex w-full items-center gap-2 rounded-xl px-2.5 py-2 text-xs font-medium text-slate-400 transition-colors hover:bg-white/10 hover:text-white"
        >
          <LogOut className="h-4 w-4 shrink-0" />
          Sign out
        </button>
      </div>
    </aside>
  )
}
