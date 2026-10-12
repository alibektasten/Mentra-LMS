'use client'

import { useState } from 'react'
import { Building2, Pencil, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useAdminOrgs, useCreateOrg, useDeleteOrg, useUpdateOrg } from '@/lib/queries/super-admin.queries'
import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent } from '@/components/ui/card'

export default function SuperAdminOrgsPage() {
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [editingID, setEditingID] = useState<string | null>(null)
  const pageSize = 20
  const { data, isLoading } = useAdminOrgs({ page, page_size: pageSize })
  const deleteOrg = useDeleteOrg()
  const createOrg = useCreateOrg()
  const updateOrg = useUpdateOrg()

  const orgs = data?.data ?? []
  const total = data?.meta?.total ?? 0
  const totalPages = Math.ceil(total / pageSize)

  const handleDelete = (id: string, name: string) => {
    if (!confirm(`Delete organization "${name}"? This cannot be undone.`)) return
    deleteOrg.mutate(id, {
      onSuccess: () => toast.success(`Organization "${name}" deleted`),
      onError: (err: unknown) => {
        const e = err as { response?: { data?: { error?: { message?: string } } } }
        toast.error(e?.response?.data?.error?.message ?? 'Failed to delete organization')
      },
    })
  }

  const handleCreate = async () => {
    try {
      const org = await createOrg.mutateAsync({ name: name.trim(), slug: slug.trim().toLowerCase() })
      toast.success(`Organization “${org.name}” created`)
      setCreateOpen(false)
      setName('')
      setSlug('')
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: { message?: string } } } }
      toast.error(e?.response?.data?.error?.message ?? 'Failed to create organization')
    }
  }
  const openEdit = (org: { id: string; name: string; slug: string }) => { setEditingID(org.id); setName(org.name); setSlug(org.slug); setCreateOpen(true) }
  const handleSave = async () => {
    try {
      if (editingID) { const org = await updateOrg.mutateAsync({ id: editingID, name: name.trim(), slug: slug.trim().toLowerCase() }); toast.success(`Organization “${org.name}” updated`) }
      else { await handleCreate(); return }
      setCreateOpen(false); setEditingID(null); setName(''); setSlug('')
    } catch { toast.error('Could not save organization') }
  }

  return (
    <div className="space-y-7">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <PageHeader title="Organizations" description={`${total} total`} />
        <Button onClick={() => { setEditingID(null); setName(''); setSlug(''); setCreateOpen(true) }} className="gap-2 rounded-xl bg-blue-600 shadow-lg shadow-blue-600/20 hover:bg-blue-700">
          <Plus className="h-4 w-4" /> Create organization
        </Button>
      </div>
      <Card className="sa-card overflow-hidden rounded-2xl border-0">
        <CardContent className="p-0">
          {isLoading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-100 bg-slate-50/70 text-left text-[11px] font-bold uppercase tracking-[0.1em] text-slate-400 dark:border-slate-700 dark:bg-slate-800/60">
                  <th className="px-5 py-4">Organization</th>
                  <th className="px-5 py-4">Workspace</th>
                  <th className="px-5 py-4">Created</th>
                  <th className="px-4 py-3 font-medium w-16" />
                </tr>
              </thead>
              <tbody>
                {orgs.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-4 py-8 text-center text-ink-muted">No organizations found</td>
                  </tr>
                ) : (
                  orgs.map((org) => (
                    <tr key={org.id} className="border-b border-slate-100 last:border-0 transition hover:bg-blue-50/50 dark:border-slate-700 dark:hover:bg-slate-800/50">
                      <td className="px-5 py-4"><div className="flex items-center gap-3"><span className="grid h-9 w-9 place-items-center rounded-xl bg-blue-50 text-blue-600"><Building2 className="h-4 w-4" /></span><span className="font-semibold text-slate-700 dark:text-slate-100">{org.name}</span></div></td>
                      <td className="px-5 py-4"><span className="rounded-md bg-slate-100 px-2 py-1 font-mono text-xs text-slate-500 dark:bg-slate-800 dark:text-slate-300">/{org.slug}</span></td>
                      <td className="px-5 py-4 text-slate-400">{new Date(org.created_at).toLocaleDateString()}</td>
                      <td className="px-5 py-4 whitespace-nowrap">
                        <Button variant="ghost" size="sm" aria-label={`Edit ${org.name}`} className="mr-1 h-8 w-8 rounded-lg p-0 text-slate-400 hover:bg-blue-50 hover:text-blue-600" onClick={() => openEdit(org)}><Pencil className="h-3.5 w-3.5" /></Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          aria-label={`Delete ${org.name}`}
                          className="h-8 w-8 rounded-lg p-0 text-slate-400 hover:bg-red-50 hover:text-red-600"
                          onClick={() => handleDelete(org.id, org.name)}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          )}
        </CardContent>
      </Card>
      {totalPages > 1 && (
        <div className="flex items-center justify-end gap-2">
          <Button variant="outline" size="sm" className="rounded-lg" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Previous
          </Button>
          <span className="text-sm text-slate-400">Page {page} of {totalPages}</span>
          <Button variant="outline" size="sm" className="rounded-lg" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            Next
          </Button>
        </div>
      )}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2"><Building2 className="h-5 w-5" /> {editingID ? 'Edit organization' : 'New organization'}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-1.5">
              <Label htmlFor="org-name">Organization name</Label>
              <Input id="org-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="Example University" />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="org-slug">Workspace slug</Label>
              <Input id="org-slug" value={slug} onChange={(e) => setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9]/g, ''))} placeholder="example" />
              <p className="text-xs text-ink-muted">Used in the login address: /example/login</p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => { setCreateOpen(false); setEditingID(null) }}>Cancel</Button>
            <Button disabled={createOrg.isPending || updateOrg.isPending || name.trim().length < 2 || slug.trim().length < 2} onClick={handleSave}>
              {createOrg.isPending || updateOrg.isPending ? 'Saving…' : editingID ? 'Save changes' : 'Create organization'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
