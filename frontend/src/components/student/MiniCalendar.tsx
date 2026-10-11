'use client'

import { useState } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { ProgressDTO } from '@/types/progress'
import { cn } from '@/lib/utils/cn'
import { useI18n } from '@/lib/i18n'

interface MiniCalendarProps {
  progress: ProgressDTO[]
}

export function MiniCalendar({ progress }: MiniCalendarProps) {
  const { locale } = useI18n()
  const languageTag = locale === 'kk' ? 'kk-KZ' : locale === 'ru' ? 'ru-RU' : 'en-US'
  const days = Array.from({ length: 7 }, (_, index) => new Intl.DateTimeFormat(languageTag, { weekday: 'short' }).format(new Date(2024, 0, index + 1)))
  const today = new Date()
  const [viewDate, setViewDate] = useState(new Date(today.getFullYear(), today.getMonth(), 1))

  const year = viewDate.getFullYear()
  const month = viewDate.getMonth()
  const monthLabel = viewDate.toLocaleDateString(languageTag, { month: 'long', year: 'numeric' })

  const activeDays = new Set(
    progress.map((p) => {
      const d = new Date(p.created_at)
      return `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`
    })
  )

  const firstDay = new Date(year, month, 1)
  const startOffset = (firstDay.getDay() + 6) % 7
  const daysInMonth = new Date(year, month + 1, 0).getDate()

  const cells: (number | null)[] = [
    ...Array(startOffset).fill(null),
    ...Array.from({ length: daysInMonth }, (_, i) => i + 1),
  ]
  while (cells.length % 7 !== 0) cells.push(null)

  const prevMonth = () => setViewDate(new Date(year, month - 1, 1))
  const nextMonth = () => setViewDate(new Date(year, month + 1, 1))

  const thisMonthActivity = progress.filter((p) => {
    const d = new Date(p.created_at)
    return d.getMonth() === month && d.getFullYear() === year
  })
  const completedThisMonth = thisMonthActivity.filter((p) => !!p.completed_at).length

  return (
    <div className="flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-widest">Activity</p>
        <div className="flex items-center gap-1">
          <button
            onClick={prevMonth}
            className="flex h-5 w-5 items-center justify-center rounded hover:bg-[#f0efed] transition-colors"
          >
            <ChevronLeft className="h-3 w-3 text-[#9b9b9b]" />
          </button>
          <span className="text-xs text-[#6b6b6b] w-24 text-center">{monthLabel}</span>
          <button
            onClick={nextMonth}
            className="flex h-5 w-5 items-center justify-center rounded hover:bg-[#f0efed] transition-colors"
          >
            <ChevronRight className="h-3 w-3 text-[#9b9b9b]" />
          </button>
        </div>
      </div>

      {/* Day headers */}
      <div className="grid grid-cols-7 mb-1">
        {days.map((d) => (
          <div key={d} className="text-center text-[10px] font-medium text-[#c9c9c9] py-1">
            {d}
          </div>
        ))}
      </div>

      {/* Day grid */}
      <div className="grid grid-cols-7 gap-y-0.5">
        {cells.map((day, i) => {
          if (!day) return <div key={`empty-${i}`} />
          const isToday =
            day === today.getDate() && month === today.getMonth() && year === today.getFullYear()
          const key = `${year}-${month}-${day}`
          const hasActivity = activeDays.has(key)

          return (
            <div
              key={key}
              className={cn(
                'mx-auto flex h-6 w-6 items-center justify-center rounded text-[11px] font-medium transition-colors cursor-default',
                isToday
                  ? 'bg-[#1a1a1a] text-white'
                  : hasActivity
                  ? 'bg-[#059669]/10 text-[#059669] font-semibold'
                  : 'text-[#9b9b9b]'
              )}
            >
              {day}
            </div>
          )
        })}
      </div>

      {/* Summary */}
      <div className="mt-4 pt-3 border-t border-[#e8e8e6]">
        <p className="text-xs text-[#9b9b9b]">
          <span className="font-semibold text-[#1a1a1a]">{completedThisMonth}</span> lessons completed this month
        </p>
      </div>
    </div>
  )
}
