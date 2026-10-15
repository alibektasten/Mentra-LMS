'use client'

import { useEffect } from 'react'
import { useParams, useRouter, useSearchParams } from 'next/navigation'
import { toast } from 'sonner'
import { LoginForm } from '@/components/auth/LoginForm'
import { parseGoogleOAuthPayload } from '@/lib/api/auth'
import { useAuthStore } from '@/lib/stores/auth.store'
import { getRoleBasePath } from '@/lib/utils/role-guard'

export function OrgLoginContent() {
  const { org } = useParams<{ org: string }>()
  const router = useRouter()
  const searchParams = useSearchParams()
  const returnTo = searchParams.get('returnTo') ?? ''
  const setSession = useAuthStore((state) => state.setSession)

  useEffect(() => {
    const encoded = new URLSearchParams(window.location.hash.slice(1)).get('oauth')
    if (!encoded) return
    try {
      const result = parseGoogleOAuthPayload(encoded)
      setSession({ token: result.token, user: result.user, role: result.role, orgSlug: org, expiresAt: result.expiresAt, authProvider: 'google' })
      window.history.replaceState(null, '', window.location.pathname + window.location.search)
      router.replace(returnTo || `/${org}/${getRoleBasePath(result.role)}`)
    } catch {
      window.history.replaceState(null, '', window.location.pathname + window.location.search)
      toast.error('Google sign-in could not be completed. Please try again.')
    }
  }, [org, returnTo, router, setSession])

  return <LoginForm orgSlug={org} returnTo={returnTo} />
}
