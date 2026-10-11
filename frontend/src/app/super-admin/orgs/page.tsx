'use client'

import { useState } from 'react'
import { Building2, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { useAdminOrgs, useCreateOrg, useDeleteOrg } from '@/lib/queries/super-admin.queries'
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
  const pageSize = 20
  const { data, isLoading } = useAdminOrgs({ page, page_size: pageSize })
  const deleteOrg = useDeleteOrg()
  const createOrg = useCreateOrg()

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

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <PageHeader title="Organizations" description={`${total} total`} />
        <Button onClick={() => setCreateOpen(true)} className="gap-2">
          <Plus className="h-4 w-4" /> Create organization
        </Button>
      </div>
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left text-ink-muted">
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Slug</th>
                  <th className="px-4 py-3 font-medium">Created</th>
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
                    <tr key={org.id} className="border-b last:border-0 hover:bg-muted/30">
                      <td className="px-4 py-3 font-medium text-ink">{org.name}</td>
                      <td className="px-4 py-3 text-ink-muted font-mono text-xs">{org.slug}</td>
                      <td className="px-4 py-3 text-ink-muted">{new Date(org.created_at).toLocaleDateString()}</td>
                      <td className="px-4 py-3">
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-7 w-7 p-0 text-destructive hover:text-destructive"
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
          <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
            Previous
          </Button>
          <span className="text-sm text-ink-muted">Page {page} of {totalPages}</span>
          <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
            Next
          </Button>
        </div>
      )}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2"><Building2 className="h-5 w-5" /> New organization</DialogTitle>
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
            <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
            <Button disabled={createOrg.isPending || name.trim().length < 2 || slug.trim().length < 2} onClick={handleCreate}>
              {createOrg.isPending ? 'Creating…' : 'Create organization'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
