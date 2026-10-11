'use client'

import { useRef, useState } from 'react'
import { toast } from 'sonner'
import { Upload, CheckCircle, XCircle } from 'lucide-react'
import { useBulkImportMembers } from '@/lib/queries/members.queries'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { CSVImportResult } from '@/types/member'
import { useI18n } from '@/lib/i18n'

export function ImportCSVDialog() {
  const { t } = useI18n()
  const [open, setOpen] = useState(false)
  const [result, setResult] = useState<CSVImportResult | null>(null)
  const fileRef = useRef<HTMLInputElement>(null)
  const importCSV = useBulkImportMembers()

  const handleImport = async () => {
    const file = fileRef.current?.files?.[0]
    if (!file) {
      toast.error(t('selectCsv'))
      return
    }
    try {
      const res = await importCSV.mutateAsync(file)
      setResult(res)
      if (res.imported.length > 0) {
        toast.success(`Imported ${res.imported.length} member(s)`)
      }
    } catch {
      toast.error(t('importFailed'))
    }
  }

  const handleClose = (v: boolean) => {
    if (!v) {
      setResult(null)
      if (fileRef.current) fileRef.current.value = ''
    }
    setOpen(v)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogTrigger asChild>
        <Button variant="outline" className="gap-1.5 border-[#e4e2de]">
          <Upload className="h-4 w-4" />
          {t('importCsv')}
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{t('importMembersCsv')}</DialogTitle>
        </DialogHeader>
        {!result ? (
          <div className="space-y-4 mt-2">
            <p className="text-xs text-[#6b6b6b]">
              CSV must have columns: <code className="bg-[#f0eeeb] px-1 rounded">name</code>,{' '}
              <code className="bg-[#f0eeeb] px-1 rounded">email</code>,{' '}
              <code className="bg-[#f0eeeb] px-1 rounded">role</code>{' '}
              (admin / teacher / student)
            </p>
            <input
              ref={fileRef}
              type="file"
              accept=".csv"
              className="block w-full text-sm text-[#6b6b6b] file:mr-3 file:py-1.5 file:px-3 file:rounded-md file:border file:border-[#e4e2de] file:text-xs file:font-medium file:bg-[#f0eeeb] file:text-[#1a1a1a] hover:file:bg-[#e4e2de]"
            />
            <Button
              onClick={handleImport}
              disabled={importCSV.isPending}
              className="w-full bg-[#059669] hover:bg-[#047857] text-white"
            >
              {importCSV.isPending ? t('importing') : t('import')}
            </Button>
          </div>
        ) : (
          <div className="space-y-4 mt-2 max-h-80 overflow-y-auto">
            {result.imported.length > 0 && (
              <div>
                <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide mb-2 flex items-center gap-1">
                  <CheckCircle className="h-3.5 w-3.5 text-emerald-600" />
                  {t('imported')} ({result.imported.length})
                </p>
                <div className="space-y-1">
                  {result.imported.map((u) => (
                    <div key={u.email} className="flex items-center justify-between rounded-md bg-emerald-50 px-3 py-1.5 text-xs">
                      <span className="font-medium text-[#1a1a1a]">{u.name} <span className="text-[#6b6b6b]">({u.email})</span></span>
                      <span className="text-emerald-700 font-semibold capitalize">{u.role}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
            {result.errors.length > 0 && (
              <div>
                <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-wide mb-2 flex items-center gap-1">
                  <XCircle className="h-3.5 w-3.5 text-red-500" />
                  {t('errors')} ({result.errors.length})
                </p>
                <div className="space-y-1">
                  {result.errors.map((e, i) => (
                    <div key={i} className="rounded-md bg-red-50 px-3 py-1.5 text-xs">
                      <span className="font-medium text-red-700">Row {e.row}</span>
                      {e.email && <span className="text-[#6b6b6b]"> · {e.email}</span>}
                      <span className="text-red-600"> — {e.error}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
            <Button
              variant="outline"
              className="w-full border-[#e4e2de]"
              onClick={() => handleClose(false)}
            >
              {t('done')}
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
