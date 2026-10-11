'use client'

import Link from 'next/link'
import { useParams } from 'next/navigation'
import {
  LayoutDashboard, BookOpen, Users, BarChart2, Sparkles, Settings,
  GraduationCap, Award, User, X,
} from 'lucide-react'
import { useAuthStore } from '@/lib/stores/auth.store'
import { NavItem } from './NavItem'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { useI18n } from '@/lib/i18n'
import { useMe } from '@/lib/queries/users.queries'

interface SidebarProps {
  mobileOpen?: boolean
  onClose?: () => void
}

export function Sidebar({ mobileOpen = false, onClose }: SidebarProps) {
  const { org } = useParams<{ org: string }>()
  const { role, user } = useAuthStore()
  const { t } = useI18n()
  const { data: profile } = useMe()

  const adminNav = [
    { href: `/${org}/admin`, label: t('dashboard'), icon: LayoutDashboard },
    { href: `/${org}/admin/courses`, label: t('courses'), icon: BookOpen },
    { href: `/${org}/admin/students`, label: t('students'), icon: Users },
    { href: `/${org}/admin/analytics`, label: t('analytics'), icon: BarChart2 },
    { href: `/${org}/admin/ai`, label: t('ai'), icon: Sparkles },
    { href: `/${org}/admin/settings`, label: t('settings'), icon: Settings },
  ]

  const teacherNav = [
    { href: `/${org}/teacher`, label: t('dashboard'), icon: LayoutDashboard },
    { href: `/${org}/teacher/courses`, label: t('myCourses'), icon: BookOpen },
    { href: `/${org}/teacher/students`, label: t('students'), icon: Users },
    { href: `/${org}/teacher/analytics`, label: t('analytics'), icon: BarChart2 },
    { href: `/${org}/teacher/ai`, label: t('ai'), icon: Sparkles },
    { href: `/${org}/teacher/settings`, label: t('personalSettings'), icon: Settings },
  ]

  const studentNav = [
    { href: `/${org}/student`, label: t('dashboard'), icon: LayoutDashboard },
    { href: `/${org}/student/courses`, label: t('myCourses'), icon: BookOpen },
    { href: `/${org}/student/progress`, label: t('progress'), icon: GraduationCap },
    { href: `/${org}/student/ai`, label: t('ai'), icon: Sparkles },
    { href: `/${org}/student/certificates`, label: t('certificates'), icon: Award },
    { href: `/${org}/student/settings`, label: t('personalSettings'), icon: Settings },
  ]

  const navItems =
    role === 'admin' || role === 'super_admin' ? adminNav :
    role === 'teacher' ? teacherNav :
    studentNav

  const initials = user
    ? `${user.first_name[0] ?? ''}${user.last_name[0] ?? ''}`.toUpperCase()
    : 'U'

  const fullName = user ? `${user.first_name} ${user.last_name}` : 'User'
  const roleLabel = role === 'super_admin' ? 'Platform administrator' : role === 'admin' ? 'Organization administrator' : role === 'teacher' ? 'Instructor' : 'Student'

  return (
    <aside className={`fixed inset-y-0 left-0 z-50 flex h-screen w-[264px] shrink-0 flex-col border-r border-[#e7e6e0] bg-[#fbfbf9] shadow-2xl transition-transform duration-300 lg:static lg:z-auto lg:translate-x-0 lg:shadow-none ${mobileOpen ? 'translate-x-0' : '-translate-x-full'}`}>
      {/* Brand */}
      <div className="flex h-[72px] items-center gap-3 px-5">
        <div className="flex h-8 w-8 items-center justify-center rounded-[10px] bg-gradient-to-br from-[#111110] to-[#3a3a35] text-white text-xs font-bold shadow-[0_5px_12px_rgba(17,17,16,0.16)]">
          M
        </div>
        <div>
          <span className="block text-[15px] font-semibold tracking-[-0.025em] text-[#171714]">Mentra</span>
          <span className="block text-[10px] font-medium uppercase tracking-[0.14em] text-[#a09f99]">Learning space</span>
        </div>
        <button aria-label="Close navigation menu" onClick={onClose} className="ml-auto rounded-lg p-2 text-[#77766f] hover:bg-[#f1f0ec] lg:hidden">
          <X className="h-4 w-4" />
        </button>
      </div>

      {/* Nav */}
      <nav className="flex-1 overflow-y-auto px-3 py-4" onClick={onClose}>
        <p className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-[0.14em] text-[#aaa9a2]">{t('workspace')}</p>
        {navItems.map((item) => (
          <NavItem key={item.href} {...item} />
        ))}
      </nav>

      {/* User footer */}
      <div className="border-t border-[#e7e6e0] p-3">
        <div className="flex items-center gap-2.5 rounded-xl border border-transparent px-2.5 py-2.5 transition-colors hover:border-[#e7e6e0] hover:bg-white cursor-pointer">
          <Avatar className="h-8 w-8 shrink-0">
            {profile?.avatar_url && <AvatarImage src={profile.avatar_url} alt="Фото профиля" />}
            <AvatarFallback className="bg-[#dff4e9] text-[#087252] text-[10px] font-bold">
              {initials}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <span className="block truncate text-xs font-semibold text-[#393934]">{fullName}</span>
            <span className="block truncate text-[10px] text-[#96958e]">{roleLabel}</span>
          </div>
        </div>
      </div>
    </aside>
  )
}
