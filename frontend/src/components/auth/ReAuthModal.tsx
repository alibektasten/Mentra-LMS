'use client'

import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { useReAuthStore } from '@/lib/stores/reauth.store'
import { useAuthStore } from '@/lib/stores/auth.store'
import { reAuthSchema, ReAuthFormData } from '@/lib/validators/auth.schema'
import { login, superAdminLogin } from '@/lib/api/auth'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export function ReAuthModal() {
  const router = useRouter()
  const { isOpen, prefillEmail, resolve, reject } = useReAuthStore()
  const { orgSlug, role, setSession } = useAuthStore()

  const { register, handleSubmit, formState: { errors, isSubmitting }, reset } = useForm<ReAuthFormData>({
    resolver: zodResolver(reAuthSchema),
  })

  const onSubmit = async (data: ReAuthFormData) => {
    try {
      const isSuperAdmin = role === 'super_admin'
      if (!orgSlug && !isSuperAdmin) return
      const result = isSuperAdmin
        ? await superAdminLogin({ email: prefillEmail, password: data.password })
        : await login(orgSlug!, { email: prefillEmail, password: data.password })
      setSession({
        token: result.token,
        user: result.user,
        role: result.role,
        orgSlug: isSuperAdmin ? '' : orgSlug!,
        expiresAt: result.expiresAt,
      })
      reset()
      resolve(result.token)
      toast.success('Session renewed')
    } catch {
      toast.error('Invalid password. Please try again.')
    }
  }

  const handleCancel = () => {
    reject()
    router.push('/login')
  }

  return (
    <Dialog open={isOpen} onOpenChange={() => {}}>
      <DialogContent
        className="sm:max-w-sm"
        onInteractOutside={(e) => e.preventDefault()}
        onEscapeKeyDown={handleCancel}
      >
        <DialogHeader>
          <DialogTitle>Session Expired</DialogTitle>
          <DialogDescription>
            Your session has expired. Enter your password to continue.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="reauth-email">Email</Label>
            <Input id="reauth-email" value={prefillEmail} readOnly className="bg-muted" />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="reauth-password">Password</Label>
            <Input
              id="reauth-password"
              type="password"
              autoFocus
              {...register('password')}
            />
            {errors.password && (
              <p className="text-xs text-destructive">{errors.password.message}</p>
            )}
          </div>
          <div className="flex gap-2 justify-end">
            <Button type="button" variant="ghost" onClick={handleCancel}>
              Sign out
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Signing in...' : 'Continue'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
