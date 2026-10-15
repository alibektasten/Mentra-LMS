import { apiClient, API_URL } from '@/lib/api/client'

export type PrivateFileKind = 'lessons' | 'attachments' | 'submissions' | 'exam-attempts'

export async function fetchPrivateFile(kind: PrivateFileKind, id: string): Promise<Blob> {
  const response = await apiClient.get<Blob>(`/files/${kind}/${encodeURIComponent(id)}`, { responseType: 'blob' })
  return response.data
}

export async function openPrivateFile(kind: PrivateFileKind, id: string, filename: string) {
  const blob = await fetchPrivateFile(kind, id)
  const blobUrl = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = blobUrl
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(blobUrl), 30_000)
}

export function publicAvatarURL(userID: string, savedURL: string): string {
  if (!savedURL) return ''
  if (savedURL.startsWith(`avatars/${userID}/`)) return `${API_URL}/avatars/${encodeURIComponent(userID)}?v=${encodeURIComponent(savedURL.split('/').pop() ?? '')}`
  if (savedURL.includes('/uploads/') || savedURL.startsWith('uploads/')) return ''
  return savedURL
}

export async function uploadAvatar(file: File): Promise<{ path: string; url: string }> {
  const body = new FormData()
  body.append('file', file)
  const response = await apiClient.post<{ path: string; url: string }>('/me/avatar', body)
  return response.data
}
