'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { loginSchema, LoginFormData } from '@/lib/validators/auth.schema'
import { getGoogleOAuthStartURL, getGoogleOAuthStatus, login, superAdminLogin } from '@/lib/api/auth'
import { useAuthStore } from '@/lib/stores/auth.store'
import { getRoleBasePath } from '@/lib/utils/role-guard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Chrome } from 'lucide-react'

interface LoginFormProps {
  orgSlug: string
  returnTo?: string
  isSuperAdmin?: boolean
}

export function LoginForm({ orgSlug, returnTo, isSuperAdmin }: LoginFormProps) {
  const router = useRouter()
  const setSession = useAuthStore((s) => s.setSession)
  const [googleEnabled, setGoogleEnabled] = useState(false)

  useEffect(() => {
	if (!orgSlug && !isSuperAdmin) return
    getGoogleOAuthStatus().then(setGoogleEnabled).catch(() => setGoogleEnabled(false))
  }, [isSuperAdmin, orgSlug])

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (data: LoginFormData) => {
    try {
      const result = isSuperAdmin
        ? await superAdminLogin({ email: data.email, password: data.password })
        : await login(orgSlug, { email: data.email, password: data.password })
      setSession({
        token: result.token,
        user: result.user,
        role: result.role,
        orgSlug: isSuperAdmin ? '' : orgSlug,
        expiresAt: result.expiresAt,
        authProvider: 'password',
      })
      if (isSuperAdmin) {
        router.push('/super-admin/dashboard')
      } else if (returnTo) {
        router.push(returnTo)
      } else {
        const basePath = getRoleBasePath(result.role)
        router.push(`/${orgSlug}/${basePath}`)
      }
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: { message?: string } } } }
      toast.error(e?.response?.data?.error?.message ?? 'Login failed. Check your credentials.')
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-1.5">
        <Label htmlFor="email">Email</Label>
        <Input id="email" type="email" placeholder="you@example.com" {...register('email')} />
        {errors.email && <p className="text-xs text-destructive">{errors.email.message}</p>}
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="password">Password</Label>
        <Input id="password" type="password" {...register('password')} />
        {errors.password && <p className="text-xs text-destructive">{errors.password.message}</p>}
      </div>
      <Button type="submit" className="w-full" disabled={isSubmitting}>
        {isSubmitting ? 'Signing in...' : 'Sign in'}
      </Button>
      <div className="space-y-3 pt-1">
          <div className="flex items-center gap-3"><div className="h-px flex-1 bg-border" /><span className="text-[11px] text-muted-foreground">or</span><div className="h-px flex-1 bg-border" /></div>
          <Button type="button" variant="outline" className="w-full" disabled={!googleEnabled} onClick={() => { window.location.assign(getGoogleOAuthStartURL(orgSlug, Boolean(isSuperAdmin))) }}>
            <Chrome className="mr-2 h-4 w-4" />Continue with Google
          </Button>
          {!googleEnabled && <p className="text-center text-xs text-muted-foreground">Google sign-in will be available after the organization adds its OAuth credentials.</p>}
	      </div>
    </form>
  )
}
