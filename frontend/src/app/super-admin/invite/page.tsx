'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { useAdminOrgs, useInviteOrgAdmin } from '@/lib/queries/super-admin.queries'
import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent } from '@/components/ui/card'
import { Building2, CheckCircle2, Eye, EyeOff, KeyRound, UserPlus } from 'lucide-react'

const schema = z.object({
  email: z.string().email('Invalid email address'),
  name: z.string().min(1, 'Name is required'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  org_id: z.string().min(1, 'Organization is required'),
})

type FormData = z.infer<typeof schema>

export default function SuperAdminInvitePage() {
  const { data: orgsData } = useAdminOrgs({ page: 1, page_size: 100 })
  const orgs = orgsData?.data ?? []
  const inviteAdmin = useInviteOrgAdmin()
  const [showPassword, setShowPassword] = useState(false)

  const { register, handleSubmit, reset, setValue, formState: { errors, isSubmitting } } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  const onSubmit = async (data: FormData) => {
    try {
      const result = await inviteAdmin.mutateAsync(data)
      toast.success(`Invited ${result.name} as org admin`)
      reset()
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: { message?: string } } } }
      toast.error(e?.response?.data?.error?.message ?? 'Failed to invite admin')
    }
  }

  const createPassword = async () => {
    const generated = Array.from(crypto.getRandomValues(new Uint32Array(4))).map((value) => value.toString(36)).join('').slice(0, 16) + 'A!7'
    setValue('password', generated, { shouldValidate: true })
    setShowPassword(true)
    try { await navigator.clipboard.writeText(generated); toast.success('Secure password generated and copied') }
    catch { toast.success('Secure password generated') }
  }

  return (
    <div className="space-y-7">
      <PageHeader title="Invite an organisation admin" description="Create an administrator account and give it access to one workspace." />
      <div className="grid max-w-5xl gap-6 lg:grid-cols-[minmax(0,1fr)_280px]">
      <Card className="sa-card rounded-2xl border-0">
        <CardContent className="p-6 sm:p-7">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5 sm:col-span-2">
              <Label htmlFor="name">Full Name</Label>
              <Input id="name" autoComplete="name" placeholder="Jane Doe" {...register('name')} />
              {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" autoComplete="email" placeholder="jane@company.com" {...register('email')} />
              {errors.email && <p className="text-xs text-destructive">{errors.email.message}</p>}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <div className="flex gap-2"><div className="relative flex-1"><Input id="password" type={showPassword ? 'text' : 'password'} autoComplete="new-password" {...register('password')} /><button type="button" aria-label={showPassword ? 'Hide password' : 'Show password'} onClick={() => setShowPassword((value) => !value)} className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-slate-400 hover:text-slate-700"><>{showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}</></button></div><Button type="button" variant="outline" className="shrink-0" onClick={createPassword}><KeyRound className="h-4 w-4" /><span className="hidden sm:inline">Generate</span></Button></div>
              {errors.password && <p className="text-xs text-destructive">{errors.password.message}</p>}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="org_id">Organization</Label>
              <select
                id="org_id"
                {...register('org_id')}
                disabled={orgsData === undefined}
                className="flex h-10 w-full rounded-lg border border-input bg-background px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-2 focus:ring-ring/20 disabled:cursor-not-allowed disabled:opacity-60"
              >
                <option value="">{orgsData === undefined ? 'Loading organisations…' : orgs.length === 0 ? 'No organisations available' : 'Select an organisation…'}</option>
                {orgs.map((org) => (
                  <option key={org.id} value={org.id}>{org.name} ({org.slug})</option>
                ))}
              </select>
              {errors.org_id && <p className="text-xs text-destructive">{errors.org_id.message}</p>}
            </div>
            </div>
            <Button type="submit" className="w-full bg-blue-600 hover:bg-blue-700" disabled={isSubmitting || orgs.length === 0}>
              <UserPlus className="h-4 w-4" /> {isSubmitting ? 'Creating administrator…' : 'Create administrator'}
            </Button>
          </form>
        </CardContent>
      </Card>
      <Card className="sa-card h-fit rounded-2xl border-0">
        <CardContent className="p-6">
          <div className="grid h-10 w-10 place-items-center rounded-xl bg-blue-50 text-blue-600"><Building2 className="h-5 w-5" /></div>
          <h2 className="mt-4 text-base font-semibold text-slate-800 dark:text-white">What happens next</h2>
          <ul className="mt-4 space-y-3 text-sm leading-5 text-slate-500 dark:text-slate-300">
            {['The account is created immediately.', 'The administrator can sign in at the selected workspace URL.', 'They can manage users, courses and groups for that organisation.'].map((item) => <li key={item} className="flex gap-2"><CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />{item}</li>)}
          </ul>
        </CardContent>
      </Card>
      </div>
    </div>
  )
}
