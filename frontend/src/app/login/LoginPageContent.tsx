'use client'

import { useSearchParams } from 'next/navigation'
import { LoginForm } from '@/components/auth/LoginForm'

export function LoginPageContent() {
  const searchParams = useSearchParams()
  const org = searchParams.get('org') ?? ''
  const returnTo = searchParams.get('returnTo') ?? ''

  return <LoginForm orgSlug={org} returnTo={returnTo} isSuperAdmin={!org} />
}
