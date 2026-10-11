'use client'

import { CheckCircle2, Circle } from 'lucide-react'
import { useProgress } from '@/lib/queries/progress.queries'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import { formatDate } from '@/lib/utils/format'
import { useI18n } from '@/lib/i18n'

export default function StudentProgressPage() {
  const { t } = useI18n()
  const { data: progress = [], isLoading } = useProgress()

  const completed = progress.filter((p) => !!p.completed_at)
  const withScore = completed.filter((p) => p.score != null)
  const avgScore = withScore.length
    ? Math.round(withScore.reduce((s, p) => s + (p.score ?? 0), 0) / withScore.length)
    : null
  const completionRate = progress.length > 0
    ? Math.round((completed.length / progress.length) * 100)
    : null

  const stats = [
    { label: t('lessonsCompleted'), value: completed.length },
    { label: t('averageScore'), value: avgScore != null ? `${avgScore}%` : '—' },
    { label: t('totalActivity'), value: progress.length },
    { label: t('completionRate'), value: completionRate != null ? `${completionRate}%` : '—' },
  ]

  const recent = [...progress]
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 20)

  return (
    <div className="max-w-3xl">
      {/* Page header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold tracking-tight text-[#1a1a1a]">{t('progress')}</h1>
        <p className="mt-1 text-sm text-[#9b9b9b]">{t('learningProgressDescription')}</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 divide-x divide-[#e8e8e6] border border-[#e8e8e6] rounded-lg overflow-hidden mb-8">
        {stats.map(({ label, value }) => (
          <div key={label} className="px-5 py-4 bg-[#fbfbfa]">
            <p className="text-[11px] text-[#9b9b9b] font-medium">{label}</p>
            <p className="mt-1 text-xl font-bold text-[#1a1a1a] tracking-tight">{value}</p>
          </div>
        ))}
      </div>

      {/* Activity */}
      <div>
        <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-widest mb-3">{t('activity')}</p>

        {isLoading ? (
          <div className="space-y-1">
            {[1, 2, 3, 4].map((i) => <Skeleton key={i} className="h-10 rounded-md" />)}
          </div>
        ) : recent.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 border border-[#e8e8e6] rounded-lg text-center">
            <p className="text-sm text-[#9b9b9b]">No activity yet — complete a lesson to see it here.</p>
          </div>
        ) : (
          <div className="divide-y divide-[#e8e8e6] border border-[#e8e8e6] rounded-lg overflow-hidden">
            {recent.map((p) => (
              <div key={p.id} className="flex items-center gap-3 px-3.5 py-2.5 hover:bg-[#f7f7f5] transition-colors">
                {p.completed_at
                  ? <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-[#059669]" />
                  : <Circle className="h-3.5 w-3.5 shrink-0 text-[#c9c9c9]" />
                }
                <span className="flex-1 text-sm text-[#3b3b3b]">
                  {p.completed_at ? t('completedLesson') : t('startedLesson')}
                </span>
                {p.score != null && (
                  <div className="flex items-center gap-2 shrink-0">
                    <Progress value={p.score} className="h-1 w-20" />
                    <span className="text-xs font-semibold text-[#3b3b3b] w-9 text-right">{p.score}%</span>
                  </div>
                )}
                <span className="text-[11px] text-[#9b9b9b] shrink-0 w-24 text-right">
                  {formatDate(p.created_at)}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
