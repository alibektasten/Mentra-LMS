'use client'

import { useMemo, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { CalendarDays, ChevronLeft, ChevronRight, CircleDot, ListChecks } from 'lucide-react'
import { ProgressDTO } from '@/types/progress'
import { cn } from '@/lib/utils/cn'
import { useI18n } from '@/lib/i18n'

export interface CalendarDeadline {
  id: string
  title: string
  dueDate: string
  courseID: string
  courseTitle: string
  submitted: boolean
}

interface MiniCalendarProps {
  progress: ProgressDTO[]
  deadlines?: CalendarDeadline[]
}

const dateKey = (date: Date) => `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`

export function MiniCalendar({ progress, deadlines = [] }: MiniCalendarProps) {
  const router = useRouter()
  const { org } = useParams<{ org: string }>()
  const { locale } = useI18n()
  const languageTag = locale === 'kk' ? 'kk-KZ' : locale === 'ru' ? 'ru-RU' : 'en-US'
  const copy = locale === 'ru'
    ? { activity: 'Активность', today: 'Сегодня', completed: 'Завершено уроков', noEvents: 'На эту дату событий нет', open: 'Открыть курс' }
    : locale === 'kk'
      ? { activity: 'Белсенділік', today: 'Бүгін', completed: 'Аяқталған сабақтар', noEvents: 'Бұл күні оқиғалар жоқ', open: 'Курсты ашу' }
      : { activity: 'Activity', today: 'Today', completed: 'Lessons completed', noEvents: 'No events on this date', open: 'Open course' }
  const today = new Date()
  const [viewDate, setViewDate] = useState(new Date(today.getFullYear(), today.getMonth(), 1))
  const [selectedKey, setSelectedKey] = useState(dateKey(today))
  const year = viewDate.getFullYear()
  const month = viewDate.getMonth()
  const days = Array.from({ length: 7 }, (_, index) => new Intl.DateTimeFormat(languageTag, { weekday: 'short' }).format(new Date(2024, 0, index + 1)))
  const monthLabel = viewDate.toLocaleDateString(languageTag, { month: 'long', year: 'numeric' })

  const activityByDay = useMemo(() => {
    const result = new Map<string, ProgressDTO[]>()
    progress.forEach((item) => {
      const key = dateKey(new Date(item.created_at))
      result.set(key, [...(result.get(key) ?? []), item])
    })
    return result
  }, [progress])
  const deadlinesByDay = useMemo(() => {
    const result = new Map<string, CalendarDeadline[]>()
    deadlines.forEach((item) => {
      const key = dateKey(new Date(item.dueDate))
      result.set(key, [...(result.get(key) ?? []), item])
    })
    return result
  }, [deadlines])

  const firstDay = new Date(year, month, 1)
  const startOffset = (firstDay.getDay() + 6) % 7
  const daysInMonth = new Date(year, month + 1, 0).getDate()
  const cells: (number | null)[] = [...Array(startOffset).fill(null), ...Array.from({ length: daysInMonth }, (_, index) => index + 1)]
  while (cells.length % 7 !== 0) cells.push(null)

  const selectDay = (day: number) => setSelectedKey(`${year}-${month}-${day}`)
  const goToToday = () => { setViewDate(new Date(today.getFullYear(), today.getMonth(), 1)); setSelectedKey(dateKey(today)) }
  const selectedActivity = activityByDay.get(selectedKey) ?? []
  const selectedDeadlines = deadlinesByDay.get(selectedKey) ?? []
  const [selectedYear, selectedMonth, selectedDay] = selectedKey.split('-').map(Number)
  const selectedLabel = new Date(selectedYear, selectedMonth, selectedDay).toLocaleDateString(languageTag, { day: 'numeric', month: 'long' })

  return (
    <section className="flex flex-col rounded-2xl border border-[#e8e8e6] bg-white/70 p-4 shadow-sm" aria-label={copy.activity}>
      <div className="mb-3 flex items-center justify-between gap-2">
        <div className="flex items-center gap-2"><CalendarDays className="h-4 w-4 text-emerald-600" /><p className="text-xs font-semibold uppercase tracking-widest text-[#6b6b6b]">{copy.activity}</p></div>
        <button onClick={goToToday} className="rounded-md px-2 py-1 text-[11px] font-semibold text-emerald-700 hover:bg-emerald-50">{copy.today}</button>
      </div>
      <div className="mb-3 flex items-center justify-between">
        <button aria-label="Previous month" onClick={() => setViewDate(new Date(year, month - 1, 1))} className="flex h-7 w-7 items-center justify-center rounded-lg hover:bg-[#f0efed]"><ChevronLeft className="h-3.5 w-3.5 text-[#6b6b6b]" /></button>
        <span className="text-xs font-semibold capitalize text-[#3b3b3b]">{monthLabel}</span>
        <button aria-label="Next month" onClick={() => setViewDate(new Date(year, month + 1, 1))} className="flex h-7 w-7 items-center justify-center rounded-lg hover:bg-[#f0efed]"><ChevronRight className="h-3.5 w-3.5 text-[#6b6b6b]" /></button>
      </div>
      <div className="mb-1 grid grid-cols-7">{days.map((day) => <div key={day} className="py-1 text-center text-[10px] font-medium text-[#9b9b9b]">{day}</div>)}</div>
      <div className="grid grid-cols-7 gap-y-1">
        {cells.map((day, index) => {
          if (!day) return <div key={`empty-${index}`} />
          const key = `${year}-${month}-${day}`
          const isToday = key === dateKey(today)
          const hasActivity = activityByDay.has(key)
          const hasDeadline = deadlinesByDay.has(key)
          const isSelected = key === selectedKey
          return <button key={key} onClick={() => selectDay(day)} aria-label={new Date(year, month, day).toLocaleDateString(languageTag)} className={cn('relative mx-auto flex h-7 w-7 items-center justify-center rounded-lg text-[11px] font-medium transition-colors', isSelected ? 'bg-emerald-600 text-white shadow-sm' : isToday ? 'border border-emerald-500 text-emerald-700' : hasActivity || hasDeadline ? 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100' : 'text-[#8f8f89] hover:bg-[#f1f0ec]')}>
            {day}{(hasActivity || hasDeadline) && !isSelected && <span className={cn('absolute bottom-0.5 h-1 w-1 rounded-full', hasDeadline ? 'bg-orange-500' : 'bg-emerald-500')} />}
          </button>
        })}
      </div>
      <div className="mt-4 border-t border-[#e8e8e6] pt-3">
        <p className="text-[11px] font-semibold text-[#55554f]">{selectedLabel}</p>
        {selectedActivity.length === 0 && selectedDeadlines.length === 0 ? <p className="mt-1 text-xs text-[#9b9b9b]">{copy.noEvents}</p> : <div className="mt-2 space-y-2">
          {selectedActivity.length > 0 && <div className="flex items-center gap-2 text-xs text-[#4d4d47]"><CircleDot className="h-3.5 w-3.5 text-emerald-600" />{selectedActivity.length} {copy.completed.toLowerCase()}</div>}
          {selectedDeadlines.map((deadline) => <button key={deadline.id} onClick={() => router.push(`/${org}/student/courses/${deadline.courseID}`)} className="flex w-full items-start gap-2 rounded-lg bg-orange-50 px-2.5 py-2 text-left hover:bg-orange-100"><ListChecks className="mt-0.5 h-3.5 w-3.5 shrink-0 text-orange-600" /><span className="min-w-0 flex-1"><span className="block truncate text-xs font-semibold text-[#4b3824]">{deadline.title}</span><span className="block truncate text-[10px] text-orange-700">{deadline.courseTitle} · {copy.open}</span></span></button>)}
        </div>}
      </div>
    </section>
  )
}
