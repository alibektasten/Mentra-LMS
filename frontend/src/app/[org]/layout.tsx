'use client'

import { useEffect } from 'react'
import { useParams } from 'next/navigation'
import { useAuthStore } from '@/lib/stores/auth.store'
import { ReAuthModal } from '@/components/auth/ReAuthModal'
import { LocaleProvider } from '@/lib/i18n'

export default function OrgLayout({ children }: { children: React.ReactNode }) {
  const { org } = useParams<{ org: string }>()
  const setOrgSlug = useAuthStore((s) => s.setOrgSlug)
  const user = useAuthStore((s) => s.user)

  useEffect(() => {
    if (org) setOrgSlug(org)
  }, [org, setOrgSlug])

  useEffect(() => {
    if (!user?.id) return
    try {
      const preferences = JSON.parse(localStorage.getItem(`mentra:preferences:${user.id}`) ?? '{}')
      document.documentElement.dataset.theme = preferences.theme ?? 'light'
      document.documentElement.dataset.uiSize = preferences.uiSize ?? 'medium'
      document.documentElement.lang = preferences.language ?? 'ru'
    } catch { /* Use the default appearance. */ }
  }, [user?.id])

  return (
    <>
      <LocaleProvider userId={user?.id}>{children}</LocaleProvider>
      <ReAuthModal />
    </>
  )
}
