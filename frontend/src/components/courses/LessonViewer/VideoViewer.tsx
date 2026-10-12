'use client'

import { ExternalLink, PlayCircle } from 'lucide-react'

interface VideoViewerProps {
  videoUrl: string
  title: string
}

function getYouTubeEmbedURL(value: string): string | null {
  try {
    const url = new URL(value)
    const host = url.hostname.replace(/^www\./, '').toLowerCase()
    let id = ''
    if (host === 'youtu.be') id = url.pathname.split('/').filter(Boolean)[0] ?? ''
    if (host === 'youtube.com' || host === 'm.youtube.com') {
      if (url.pathname === '/watch') id = url.searchParams.get('v') ?? ''
      else if (url.pathname.startsWith('/embed/') || url.pathname.startsWith('/shorts/')) id = url.pathname.split('/')[2] ?? ''
    }
    return /^[A-Za-z0-9_-]{11}$/.test(id) ? `https://www.youtube-nocookie.com/embed/${id}?rel=0&modestbranding=1` : null
  } catch { return null }
}

export function VideoViewer({ videoUrl, title }: VideoViewerProps) {
  const youtubeEmbedURL = getYouTubeEmbedURL(videoUrl)

  if (!videoUrl) return <div className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted-foreground">This lesson has no video source yet.</div>

  return <div className="space-y-3">
    <div className="aspect-video w-full overflow-hidden rounded-xl bg-black shadow-sm">
      {youtubeEmbedURL ? (
        <iframe src={youtubeEmbedURL} title={title} className="h-full w-full border-0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowFullScreen referrerPolicy="strict-origin-when-cross-origin" />
      ) : (
        <video controls playsInline className="h-full w-full" title={title} preload="metadata"><source src={videoUrl} />Your browser does not support the video element.</video>
      )}
    </div>
    <a href={videoUrl} target="_blank" rel="noreferrer" className="inline-flex items-center gap-1.5 text-sm font-medium text-accent hover:underline"><PlayCircle className="h-4 w-4" />Open video in a new tab <ExternalLink className="h-3.5 w-3.5" /></a>
  </div>
}
