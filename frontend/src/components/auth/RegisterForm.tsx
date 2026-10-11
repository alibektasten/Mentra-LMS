'use client'

import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { registerSchema, RegisterFormData } from '@/lib/validators/auth.schema'
import { register as registerUser } from '@/lib/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface RegisterFormProps {
  defaultOrgSlug?: string
}

export function RegisterForm({ defaultOrgSlug }: RegisterFormProps) {
  const router = useRouter()

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: { orgSlug: defaultOrgSlug ?? '' },
  })

  const onSubmit = async (data: RegisterFormData) => {
    try {
      await registerUser(data.orgSlug, {
        email: data.email,
        password: data.password,
        first_name: data.first_name,
        last_name: data.last_name,
      })
      toast.success('Account created! Please sign in.')
      router.push(`/${data.orgSlug}/login`)
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: { message?: string } } } }
      toast.error(e?.response?.data?.error?.message ?? 'Registration failed.')
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-1.5">
        <Label htmlFor="orgSlug">Organization</Label>
        <Input id="orgSlug" placeholder="your-org" {...register('orgSlug')} />
        {errors.orgSlug && <p className="text-xs text-destructive">{errors.orgSlug.message}</p>}
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1.5">
          <Label htmlFor="first_name">First name</Label>
          <Input id="first_name" {...register('first_name')} />
          {errors.first_name && <p className="text-xs text-destructive">{errors.first_name.message}</p>}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="last_name">Last name</Label>
          <Input id="last_name" {...register('last_name')} />
          {errors.last_name && <p className="text-xs text-destructive">{errors.last_name.message}</p>}
        </div>
      </div>
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
        {isSubmitting ? 'Creating account...' : 'Create account'}
      </Button>
    </form>
  )
}
