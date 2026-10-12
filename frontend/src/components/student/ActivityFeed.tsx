'use client'

import { CheckCircle2, Circle } from 'lucide-react'
import { ProgressDTO } from '@/types/progress'
import { formatDate, formatPercent } from '@/lib/utils/format'
import { useI18n } from '@/lib/i18n'
import { StudentLessonContext } from '@/lib/queries/student-lesson-context.queries'

interface ActivityFeedProps {
  progress: ProgressDTO[]
  lessonContexts?: Map<string, StudentLessonContext>
}

export function ActivityFeed({ progress, lessonContexts = new Map() }: ActivityFeedProps) {
  const { t } = useI18n()
  const recent = [...progress]
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 8)

  return (
    <div className="flex flex-col">
      <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-widest mb-3">{t('recentActivity')}</p>

      {recent.length === 0 ? (
        <p className="text-sm text-[#9b9b9b] py-2">{t('noActivityYet')}</p>
      ) : (
        <div className="divide-y divide-[#e8e8e6] border border-[#e8e8e6] rounded-lg overflow-hidden">
          {recent.map((p) => {
            const context = lessonContexts.get(p.lesson_id)
            return <div key={p.id} className="flex items-center gap-3 px-3.5 py-2.5">
              {p.completed_at
                ? <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-[#059669]" />
                : <Circle className="h-3.5 w-3.5 shrink-0 text-[#c9c9c9]" />
              }
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-[#3b3b3b]">{context?.lessonTitle ?? (p.completed_at ? t('completedLesson') : t('startedLesson'))}</p>
                <p className="truncate text-[11px] text-[#8f8f89]">{context ? `${context.courseTitle} · ${context.moduleTitle}` : (p.completed_at ? t('completedLesson') : t('startedLesson'))}</p>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                {p.score != null && (
                  <span className="text-xs font-semibold tabular-nums text-[#059669]">{formatPercent(p.score)}</span>
                )}
                <span className="text-[11px] text-[#9b9b9b]">{formatDate(p.created_at)}</span>
              </div>
            </div>
          })}
        </div>
      )}
    </div>
  )
}
