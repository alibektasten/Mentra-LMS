'use client'

import { useMyGrades } from '@/lib/queries/grades.queries'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils/cn'
import { CheckCircle2, Clock } from 'lucide-react'
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

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="rounded-lg border bg-surface-50 p-4 flex items-center justify-between">
        <div>
          <p className="text-sm text-ink-muted">{t('overallGrade')}</p>
          <p className="text-2xl font-bold text-ink">{data.percentage.toFixed(1)}%</p>
          <p className="text-xs text-ink-subtle">{data.total_earned} / {data.total_possible} pts</p>
        </div>
        <div className="text-4xl font-bold text-accent">{letter}</div>
      </div>

      {/* Items */}
      <div className="divide-y divide-border rounded-lg border">
        {data.items.map((item) => (
          <div key={item.item_id} className="flex items-center gap-3 px-4 py-3">
            {item.score !== null ? (
              <CheckCircle2 className="h-4 w-4 text-green-600 shrink-0" />
            ) : (
              <Clock className="h-4 w-4 text-amber-500 shrink-0" />
            )}
            <span className="flex-1 text-sm">{item.title}</span>
            <span className={cn(
              'text-xs px-2 py-0.5 rounded capitalize',
              item.item_type === 'assignment' ? 'text-amber-700 bg-amber-50' : 'text-purple-700 bg-purple-50'
            )}>
              {item.item_type}
            </span>
            <span className="text-sm font-medium min-w-[80px] text-right">
              {item.score !== null
                ? `${item.score} / ${item.max_points}`
                : <span className="text-ink-subtle">{t('notGraded')}</span>
              }
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
