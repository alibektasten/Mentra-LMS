'use client'

import { useMyGrades } from '@/lib/queries/grades.queries'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils/cn'
import { CheckCircle2, Clock, MessageSquareText, Send, XCircle } from 'lucide-react'
import { useI18n } from '@/lib/i18n'

interface MyGradesProps {
  courseId: string
}

function letterGrade(pct: number): string {
  if (pct >= 90) return 'A'
  if (pct >= 80) return 'B'
  if (pct >= 70) return 'C'
  if (pct >= 60) return 'D'
  return 'F'
}

export function MyGrades({ courseId }: MyGradesProps) {
  const { t } = useI18n()
  const { data, isLoading } = useMyGrades(courseId)

  if (isLoading) {
    return <div className="space-y-2">{[1, 2, 3].map(i => <Skeleton key={i} className="h-12 rounded-lg" />)}</div>
  }

  if (!data || data.items.length === 0) {
    return <p className="text-sm text-ink-muted py-4">{t('noGradedItems')}</p>
  }

  const letter = letterGrade(data.percentage)
  const gradedCount = data.items.filter((item) => item.status === 'graded').length
  const submittedCount = data.items.filter((item) => item.status === 'submitted').length

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="rounded-xl border bg-surface-50 p-4 sm:p-5">
        <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-ink-muted">{t('overallGrade')}</p>
          <p className="text-2xl font-bold text-ink">{data.percentage.toFixed(1)}%</p>
          <p className="text-xs text-ink-subtle">{data.total_earned} / {data.total_possible} pts</p>
        </div>
        <div className="text-4xl font-bold text-accent">{letter}</div>
        </div>
        <div className="mt-4 grid grid-cols-3 gap-2 border-t pt-4 text-center text-xs">
          <div><p className="font-semibold text-emerald-700">{gradedCount}</p><p className="mt-0.5 text-ink-subtle">Graded</p></div>
          <div><p className="font-semibold text-amber-700">{submittedCount}</p><p className="mt-0.5 text-ink-subtle">In review</p></div>
          <div><p className="font-semibold text-slate-600">{data.items.length - gradedCount - submittedCount}</p><p className="mt-0.5 text-ink-subtle">Not submitted</p></div>
        </div>
      </div>

      {/* Items */}
      <div className="divide-y divide-border rounded-lg border">
        {data.items.map((item) => (
          <div key={item.item_id} className="px-4 py-3.5">
          <div className="flex items-center gap-3">
            {item.status === 'graded' ? (
              <CheckCircle2 className="h-4 w-4 text-green-600 shrink-0" />
            ) : item.status === 'submitted' ? (
              <Clock className="h-4 w-4 text-amber-500 shrink-0" />
            ) : (
              <XCircle className="h-4 w-4 text-slate-400 shrink-0" />
            )}
            <div className="min-w-0 flex-1"><span className="block text-sm font-medium text-ink">{item.title}</span>{item.graded_at && <span className="mt-0.5 block text-xs text-ink-subtle">Graded {new Date(item.graded_at).toLocaleDateString()}</span>}</div>
            <span className={cn(
              'text-xs px-2 py-0.5 rounded capitalize',
              item.item_type === 'assignment' ? 'text-amber-700 bg-amber-50' : 'text-purple-700 bg-purple-50'
            )}>
              {item.item_type}
            </span>
            <span className="text-sm font-medium min-w-[80px] text-right">
              {item.status === 'graded' && item.score !== null
                ? `${item.score} / ${item.max_points}`
                : item.status === 'submitted' ? <span className="text-amber-700">In review</span> : <span className="text-ink-subtle">Not submitted</span>
              }
            </span>
          </div>
          {item.feedback && (
            <div className="ml-7 mt-3 flex gap-2 rounded-lg border border-emerald-100 bg-emerald-50/70 px-3 py-2.5 text-sm text-emerald-950">
              <MessageSquareText className="mt-0.5 h-4 w-4 shrink-0 text-emerald-700" />
              <div><p className="text-xs font-semibold text-emerald-800">Instructor feedback</p><p className="mt-0.5 whitespace-pre-wrap leading-relaxed">{item.feedback}</p></div>
            </div>
          )}
          {item.status === 'submitted' && <p className="ml-7 mt-2 flex items-center gap-1.5 text-xs text-amber-700"><Send className="h-3.5 w-3.5" />Submitted and waiting for instructor review</p>}
          </div>
        ))}
      </div>
    </div>
  )
}
