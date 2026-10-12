'use client'

import { useUpcomingDeadlines } from '@/lib/queries/grades.queries'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils/cn'
import { CheckCircle2, Clock, Flag, GraduationCap } from 'lucide-react'
import { useI18n } from '@/lib/i18n'

interface UpcomingDeadlinesProps {
  courseId: string
}

export function UpcomingDeadlines({ courseId }: UpcomingDeadlinesProps) {
  const { t } = useI18n()
  const { data, isLoading } = useUpcomingDeadlines(courseId)

  if (isLoading) {
    return <div className="space-y-2">{[1, 2].map(i => <Skeleton key={i} className="h-12 rounded-lg" />)}</div>
  }

  if (!data || data.length === 0) {
    return <p className="text-sm text-ink-muted py-4">{t('noUpcomingDeadlines')}</p>
  }

  const sorted = [...data].sort((a, b) => new Date(a.due_date).getTime() - new Date(b.due_date).getTime())

  const assessmentMeta = {
    assignment: { label: 'Assignment', className: 'text-amber-700 bg-amber-50 dark:text-amber-300 dark:bg-amber-950/40' },
    quiz: { label: 'Quiz', className: 'text-purple-700 bg-purple-50 dark:text-purple-300 dark:bg-purple-950/40' },
    midterm: { label: 'Midterm', className: 'text-violet-700 bg-violet-50 dark:text-violet-300 dark:bg-violet-950/40' },
    endterm: { label: 'Endterm', className: 'text-sky-700 bg-sky-50 dark:text-sky-300 dark:bg-sky-950/40' },
    final: { label: 'Final exam', className: 'text-rose-700 bg-rose-50 dark:text-rose-300 dark:bg-rose-950/40' },
  } as const

  return (
    <div className="divide-y divide-border rounded-lg border">
      {sorted.map((item) => {
        const due = new Date(item.due_date)
        const daysLeft = Math.ceil((due.getTime() - Date.now()) / (1000 * 60 * 60 * 24))
        const urgency = daysLeft <= 2 ? 'text-red-600 bg-red-50 border-red-200' : daysLeft <= 7 ? 'text-amber-700 bg-amber-50 border-amber-200' : 'text-ink-muted bg-surface-50 border-border'
        const meta = assessmentMeta[item.item_type]
        const isExam = item.item_type === 'midterm' || item.item_type === 'endterm' || item.item_type === 'final'

        return (
          <div key={item.item_id} className="flex items-center gap-3 px-4 py-3">
            {item.submitted ? (
              <CheckCircle2 className="h-4 w-4 text-green-600 shrink-0" />
            ) : (
              <Clock className={cn('h-4 w-4 shrink-0', daysLeft <= 2 ? 'text-red-500' : 'text-amber-500')} />
            )}
            {isExam ? <GraduationCap className="h-4 w-4 shrink-0 text-primary" /> : <Flag className="h-4 w-4 shrink-0 text-amber-500" />}
            <span className="flex-1 text-sm font-medium">{item.title}</span>
            <span className={cn(
              'text-xs px-2 py-0.5 rounded',
              meta.className
            )}>
              {meta.label}
            </span>
            {item.submitted ? (
              <span className="text-xs text-green-600">{t('submitted')}</span>
            ) : (
              <span className={cn('text-xs px-2 py-0.5 rounded border', urgency)}>
                {daysLeft <= 0 ? t('dueToday') : `${daysLeft} ${t('daysLeft')}`}
              </span>
            )}
          </div>
        )
      })}
    </div>
  )
}
