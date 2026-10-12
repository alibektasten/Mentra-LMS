'use client'

import { useRouter, usePathname } from 'next/navigation'
import { useMemo, useRef, useState } from 'react'
import { useI18n } from '@/lib/i18n'
import { useMe } from '@/lib/queries/users.queries'
import { Bell, ChevronDown, LogOut, Menu, User, X } from 'lucide-react'
import { useAuthStore } from '@/lib/stores/auth.store'
import { getRoleLabel } from '@/lib/utils/format'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { useNotifications } from '@/lib/queries/notifications.queries'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

const PAGE_TITLES: Record<string, string> = {
  '': 'Dashboard',
  'courses': 'My Courses',
  'progress': 'Progress',
  'grades': 'My grades',
  'ai': 'Study AI',
  'certificates': 'Certificates',
  'analytics': 'Analytics',
  'students': 'Students',
  'settings': 'Settings',
}

function usePageTitle(): string {
  const pathname = usePathname()
  const segments = pathname.split('/').filter(Boolean)
  const rest = segments.slice(2)
  if (rest.length === 0) return 'Dashboard'
  if (rest[0] === 'courses' && rest.length === 2) return 'Course Details'
  if (rest[0] === 'courses' && rest.length >= 4 && rest[2] === 'lessons') return 'Lesson'
  return PAGE_TITLES[rest[0]] ?? 'Page'
}

interface TopBarProps {
  onMenuClick?: () => void
}

export function TopBar({ onMenuClick }: TopBarProps) {
  const router = useRouter()
  const pathname = usePathname()
  const { user, role, clearSession } = useAuthStore()
  const pageTitle = usePageTitle()
  const [notificationsOpen, setNotificationsOpen] = useState(false)
  const notificationTouchStart = useRef<number | null>(null)
  const { t } = useI18n()
  const { data: profile } = useMe()
  const { data: notifications = [] } = useNotifications()
  const [dismissedNotifications, setDismissedNotifications] = useState<Set<string>>(new Set())
  const visibleNotifications = useMemo(() => notifications.filter((item) => !dismissedNotifications.has(item.id)), [dismissedNotifications, notifications])

  const initials = user
    ? `${user.first_name[0] ?? ''}${user.last_name[0] ?? ''}`.toUpperCase()
    : 'U'

  const handleLogout = () => {
    clearSession()
    router.push('/login')
  }

  return (
    <header className="relative z-40 flex h-[72px] shrink-0 items-center justify-between border-b border-[#e7e6e0] bg-[#fbfbf9]/90 px-5 backdrop-blur-xl sm:px-7">
      <div className="flex min-w-0 items-center gap-2.5">
        <Button variant="ghost" size="icon" className="h-9 w-9 shrink-0 rounded-xl text-[#77766f] lg:hidden" aria-label="Open navigation menu" onClick={onMenuClick}>
          <Menu className="h-4 w-4" />
        </Button>
        <div className="min-w-0">
        <p className="mb-0.5 text-[10px] font-semibold uppercase tracking-[0.14em] text-[#aaa9a2]">Mentra · {t('workspace')}</p>
        <h1 className="truncate text-[17px] font-semibold tracking-[-0.025em] text-[#1a1a17]">{pageTitle}</h1>
        </div>
      </div>

      <div className="flex items-center gap-1.5 sm:gap-2">
        <Button variant="ghost" size="icon" className="relative h-9 w-9 rounded-xl text-[#77766f]" aria-label="Notifications" onClick={() => setNotificationsOpen((open) => !open)}>
          <Bell className="h-4 w-4" />
          {visibleNotifications.length > 0 && <span className="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-emerald-500 ring-2 ring-[#fbfbf9]" />}
        </Button>
        <div className="hidden h-6 w-px bg-[#e4e3dd] sm:block" />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              variant="ghost"
              className="h-10 gap-2 rounded-xl px-2 text-[#54544e] hover:bg-[#f1f0ec] hover:text-[#1a1a17]"
            >
              <Avatar className="h-7 w-7">
                {profile?.avatar_url && <AvatarImage src={profile.avatar_url} alt="Фото профиля" />}
                <AvatarFallback className="bg-[#dff4e9] text-[#087252] text-[10px] font-bold">
                  {initials}
                </AvatarFallback>
              </Avatar>
              <span className="hidden max-w-28 truncate text-xs font-semibold sm:block">{user?.first_name} {user?.last_name}</span>
              <ChevronDown className="h-3.5 w-3.5 opacity-60" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56 rounded-xl border-[#e5e4df] p-1.5 shadow-xl">
            <DropdownMenuLabel className="px-2.5 py-2">
              <p className="text-xs font-semibold text-[#1a1a17]">{user?.first_name} {user?.last_name}</p>
              <p className="mt-0.5 truncate text-[11px] font-normal text-[#92918a]">{user?.email}</p>
              <p className="mt-1.5 text-[10px] font-semibold uppercase tracking-[0.1em] text-[#07855f]">{getRoleLabel(role ?? 'student')}</p>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => router.push(`/${pathname.split('/').filter(Boolean)[0]}/${role === 'admin' ? 'admin/profile' : role === 'teacher' ? 'teacher/settings' : 'student/settings'}`)} className="gap-2 rounded-lg px-2.5 py-2 text-xs">
              <User className="h-3.5 w-3.5" />
              {t('personalSettings')}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={handleLogout}
              className="gap-2 rounded-lg px-2.5 py-2 text-xs text-destructive focus:text-destructive"
            >
              <LogOut className="h-3.5 w-3.5" />
              {t('signOut')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      {notificationsOpen && <div id="notification-panel" onTouchStart={(event) => { notificationTouchStart.current = event.touches[0]?.clientX ?? null }} onTouchEnd={(event) => { const start = notificationTouchStart.current; const end = event.changedTouches[0]?.clientX; if (start !== null && end !== undefined && Math.abs(end - start) > 56) setNotificationsOpen(false); notificationTouchStart.current = null }} className="fixed right-4 top-[76px] z-[100] w-[min(22rem,calc(100vw-2rem))] touch-pan-y overflow-hidden rounded-2xl border border-border bg-popover text-popover-foreground shadow-[0_18px_44px_rgba(20,28,22,0.18)] sm:right-6"><div className="flex items-center justify-between border-b border-border px-4 py-3"><p className="text-sm font-semibold">{t('notifications')}</p><span className="rounded-full bg-emerald-50 px-2 py-0.5 text-[10px] font-semibold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">{visibleNotifications.length} new</span></div>{visibleNotifications.length === 0 ? <div className="px-4 py-5 text-center"><Bell className="mx-auto h-6 w-6 text-emerald-600" /><p className="mt-2 text-sm font-semibold">Вы в курсе событий</p></div> : <div className="max-h-[min(32rem,calc(100dvh-6rem))] overflow-y-auto">{visibleNotifications.map((item) => <div key={item.id} className="group flex gap-3 border-b border-border px-4 py-3 last:border-0"><Bell className="mt-0.5 h-4 w-4 shrink-0 text-emerald-600" /><button className="min-w-0 flex-1 text-left" onClick={() => { setNotificationsOpen(false); router.push(`/${pathname.split('/').filter(Boolean)[0]}/${role}/courses/${item.course_id}`) }}><p className="truncate text-xs font-semibold">{item.title}</p><p className="mt-0.5 text-[11px] font-medium text-emerald-700 dark:text-emerald-300">{item.course_title}</p><p className="mt-1 line-clamp-2 text-xs text-muted-foreground">{item.body}</p></button><button aria-label="Dismiss notification" className="h-6 w-6 shrink-0 rounded text-muted-foreground opacity-70 hover:bg-muted hover:text-foreground" onClick={() => setDismissedNotifications((previous) => new Set(previous).add(item.id))}><X className="h-3.5 w-3.5" /></button></div>)}</div>}</div>}
    </header>
  )
}
