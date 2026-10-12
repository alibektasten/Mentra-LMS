export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 7) return formatDate(dateStr)
  if (days > 0) return `${days}d ago`
  if (hours > 0) return `${hours}h ago`
  if (minutes > 0) return `${minutes}m ago`
  return 'just now'
}

export function formatPercent(value: number): string {
  return `${normalizePercentage(value)}%`
}

/** Keeps progress values safe for both labels and progress-bar widths. */
export function normalizePercentage(value: number): number {
  if (!Number.isFinite(value)) return 0
  return Math.min(100, Math.max(0, Math.round(value)))
}

export function truncate(str: string, length: number): string {
  return str.length > length ? `${str.slice(0, length)}...` : str
}

export function getRoleLabel(role: string): string {
  const labels: Record<string, string> = {
    super_admin: 'Super Admin',
    admin: 'Admin',
    teacher: 'Teacher',
    student: 'Student',
  }
  return labels[role] ?? role
}
