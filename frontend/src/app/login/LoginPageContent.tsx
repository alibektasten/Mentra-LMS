'use client'

import { useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { LoginForm } from '@/components/auth/LoginForm'
import { parseGoogleOAuthPayload } from '@/lib/api/auth'
import { useAuthStore } from '@/lib/stores/auth.store'
import { toast } from 'sonner'

export function LoginPageContent() {
	const searchParams = useSearchParams()
	const router = useRouter()
	const setSession = useAuthStore((state) => state.setSession)
	const org = searchParams.get('org') ?? ''
	const returnTo = searchParams.get('returnTo') ?? ''

	useEffect(() => {
		const encoded = new URLSearchParams(window.location.hash.slice(1)).get('oauth')
		if (!encoded || org) return
		try {
			const result = parseGoogleOAuthPayload(encoded)
			if (result.role !== 'super_admin') throw new Error('Unexpected role')
			setSession({ token: result.token, user: result.user, role: result.role, orgSlug: '', expiresAt: result.expiresAt, authProvider: 'google' })
			window.history.replaceState(null, '', window.location.pathname + window.location.search)
			router.replace('/super-admin/dashboard')
		} catch {
			window.history.replaceState(null, '', window.location.pathname + window.location.search)
			toast.error('Google sign-in could not be completed. Please try again.')
		}
	}, [org, router, setSession])

	return <LoginForm orgSlug={org} returnTo={returnTo} isSuperAdmin={!org} />
}
