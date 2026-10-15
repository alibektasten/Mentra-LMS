'use client'

import { useEffect, useState } from 'react'
import { FileText, ExternalLink } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { fetchPrivateFile } from '@/lib/api/private-files'

interface PDFViewerProps {
  lessonId: string
  fileUrl: string
  title: string
}

export function PDFViewer({ lessonId, fileUrl, title }: PDFViewerProps) {
  const [previewURL, setPreviewURL] = useState('')
  const [loadError, setLoadError] = useState(false)
  useEffect(() => {
    if (!fileUrl) return
    if (!fileUrl.includes('/uploads/') && !fileUrl.startsWith('uploads/')) {
      setPreviewURL(fileUrl)
      return
    }
    let active = true
    let blobURL = ''
    fetchPrivateFile('lessons', lessonId).then((blob) => {
      if (!active) return
      blobURL = URL.createObjectURL(blob)
      setPreviewURL(blobURL)
      setLoadError(false)
    }).catch(() => { if (active) setLoadError(true) })
    return () => { active = false; if (blobURL) URL.revokeObjectURL(blobURL) }
  }, [fileUrl, lessonId])

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <FileText className="h-5 w-5 text-[#9b9b9b]" />
          <span className="text-sm font-medium text-[#1a1a1a]">{title}</span>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => { if (previewURL) window.open(previewURL, '_blank', 'noopener,noreferrer') }}
          disabled={!previewURL}
          className="gap-1.5"
        >
          <ExternalLink className="h-3.5 w-3.5" />
          Open PDF
        </Button>
      </div>
      <div className="rounded-xl border border-[#e4e2de] overflow-hidden bg-[#f7f6f3]" style={{ height: '70vh' }}>
        {previewURL ? <iframe src={previewURL} title={title} className="w-full h-full" loading="lazy" /> : <p className="p-6 text-sm text-muted-foreground">{loadError ? 'PDF is unavailable.' : 'Loading PDF…'}</p>}
      </div>
    </div>
  )
}
