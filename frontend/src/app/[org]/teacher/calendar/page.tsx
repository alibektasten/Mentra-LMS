'use client'

import Link from 'next/link'
import { useMemo, useState } from 'react'
import { useParams } from 'next/navigation'
import { ChevronLeft, ChevronRight, Clock3, GraduationCap, NotebookPen, RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils/cn'
import { useTeacherCalendar } from '@/lib/queries/teacher.queries'
import type { TeacherCalendarEventDTO } from '@/lib/api/teacher'

const weekdayLabels = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб']

function dateKey(date: Date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function monthRange(month: Date) {
  const from = new Date(month.getFullYear(), month.getMonth(), 1)
  const to = new Date(month.getFullYear(), month.getMonth() + 1, 0)
  return { from: dateKey(from), to: dateKey(to) }
}

function monthDays(month: Date) {
  const first = new Date(month.getFullYear(), month.getMonth(), 1)
  const last = new Date(month.getFullYear(), month.getMonth() + 1, 0)
  const start = new Date(first)
  start.setDate(first.getDate() - first.getDay())
  const end = new Date(last)
  end.setDate(last.getDate() + (6 - last.getDay()))

  const days: Date[] = []
  for (const day = new Date(start); day <= end; day.setDate(day.getDate() + 1)) {
    days.push(new Date(day))
  }
  return days
}

function formatTime(dateValue: string) {
  return new Date(dateValue).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

function eventTone(type: TeacherCalendarEventDTO['item_type']) {
  if (type === 'exam') return 'bg-violet-50 text-violet-800 ring-violet-100 hover:bg-violet-100'
  if (type === 'class') return 'bg-sky-50 text-sky-800 ring-sky-100 hover:bg-sky-100'
  return 'bg-amber-50 text-amber-800 ring-amber-100 hover:bg-amber-100'
}

function eventDescription(event: TeacherCalendarEventDTO) {
  if (event.item_type === 'class') {
    return `${formatTime(event.start_at)}–${event.end_at ? formatTime(event.end_at) : ''}${event.location ? ` · ${event.location}` : ''}`
  }
  if (event.item_type === 'exam') return `${formatTime(event.start_at)} · ${event.assessment_type || 'Экзамен'}`
  return `${formatTime(event.start_at)} · Дедлайн`
}

export default function TeacherCalendarPage() {
  const { org } = useParams<{ org: string }>()
  const [month, setMonth] = useState(() => new Date())
  const { from, to } = monthRange(month)
  const { data: events = [], isLoading, isError, refetch, isFetching } = useTeacherCalendar(from, to)
  const days = useMemo(() => monthDays(month), [month])
  const eventsByDay = useMemo(() => {
    const grouped = new Map<string, TeacherCalendarEventDTO[]>()
    for (const event of events) {
      const key = dateKey(new Date(event.start_at))
      grouped.set(key, [...(grouped.get(key) ?? []), event])
    }
    return grouped
  }, [events])
  const today = dateKey(new Date())
  const monthTitle = month.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })

  return (
    <div className="space-y-6">
      <PageHeader
        title="Календарь"
        description="Дедлайны, экзамены и занятия ваших групп в одном месте."
        actions={<Button variant="outline" size="sm" onClick={() => refetch()} disabled={isFetching}><RefreshCw className={cn('mr-2 h-4 w-4', isFetching && 'animate-spin')} />Обновить</Button>}
      />

      <section className="rounded-xl border border-[#e4e2de] bg-white shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-[#f0eeeb] px-4 py-3 sm:px-5">
          <div className="flex items-center gap-2">
            <Button variant="outline" size="icon" aria-label="Предыдущий месяц" onClick={() => setMonth((value) => new Date(value.getFullYear(), value.getMonth() - 1, 1))}><ChevronLeft className="h-4 w-4" /></Button>
            <Button variant="outline" size="sm" onClick={() => setMonth(new Date())}>Сегодня</Button>
            <Button variant="outline" size="icon" aria-label="Следующий месяц" onClick={() => setMonth((value) => new Date(value.getFullYear(), value.getMonth() + 1, 1))}><ChevronRight className="h-4 w-4" /></Button>
          </div>
          <h2 className="capitalize text-base font-semibold text-[#1a1a1a]">{monthTitle}</h2>
          <div className="flex items-center gap-3 text-[11px] text-[#6b6b6b]">
            <Legend tone="bg-amber-400" label="Дедлайны" />
            <Legend tone="bg-violet-400" label="Экзамены" />
            <Legend tone="bg-sky-400" label="Занятия" />
          </div>
        </div>

        {isLoading ? <CalendarSkeleton /> : isError ? (
          <div className="flex min-h-72 flex-col items-center justify-center gap-3 px-5 text-center"><p className="text-sm text-[#6b6b6b]">Не удалось загрузить календарь.</p><Button variant="outline" size="sm" onClick={() => refetch()}>Попробовать снова</Button></div>
        ) : (
          <div className="overflow-x-auto">
            <div className="min-w-[720px]">
              <div className="grid grid-cols-7 border-b border-[#f0eeeb] bg-[#fbfbf9]">
                {weekdayLabels.map((label) => <div key={label} className="px-3 py-2 text-center text-[11px] font-semibold uppercase tracking-wide text-[#85847d]">{label}</div>)}
              </div>
              <div className="grid grid-cols-7">
                {days.map((day) => {
                  const key = dateKey(day)
                  const isCurrentMonth = day.getMonth() === month.getMonth()
                  const dayEvents = eventsByDay.get(key) ?? []
                  return <div key={key} className={cn('min-h-32 border-b border-r border-[#f0eeeb] p-2', !isCurrentMonth && 'bg-[#fcfcfb] text-[#b5b4ae]')}>
                    <div className="mb-1 flex justify-end"><span className={cn('flex h-6 w-6 items-center justify-center rounded-full text-xs font-semibold', key === today && 'bg-[#059669] text-white')}>{day.getDate()}</span></div>
                    <div className="space-y-1">
                      {dayEvents.slice(0, 3).map((event) => <EventLink key={event.id} event={event} org={org} />)}
                      {dayEvents.length > 3 && <p className="px-1 text-[10px] font-medium text-[#6b6b6b]">ещё {dayEvents.length - 3}</p>}
                    </div>
                  </div>
                })}
              </div>
            </div>
          </div>
        )}
      </section>
      {!isLoading && !isError && events.length === 0 && <p className="rounded-xl border border-dashed border-[#deddd7] bg-[#fbfbf9] px-5 py-4 text-sm text-[#6b6b6b]">В этом месяце нет дедлайнов, экзаменов или занятий групп.</p>}
    </div>
  )
}

function Legend({ tone, label }: { tone: string; label: string }) {
  return <span className="flex items-center gap-1.5"><span className={cn('h-2 w-2 rounded-full', tone)} />{label}</span>
}

function EventLink({ event, org }: { event: TeacherCalendarEventDTO; org: string }) {
  const icon = event.item_type === 'class' ? <GraduationCap className="h-3 w-3" /> : event.item_type === 'exam' ? <NotebookPen className="h-3 w-3" /> : <Clock3 className="h-3 w-3" />
  return (
    <Link href={`/${org}/teacher/courses/${event.course_id}`} className={cn('block rounded-md px-1.5 py-1 text-[10px] ring-1 transition', eventTone(event.item_type))} title={`${event.title} · ${event.course_title}`}>
      <span className="flex items-center gap-1 font-semibold"><span className="shrink-0">{icon}</span><span className="truncate">{event.title}</span></span>
      <span className="mt-0.5 block truncate pl-4 text-[9px] opacity-80">{eventDescription(event)}</span>
    </Link>
  )
}

function CalendarSkeleton() {
  return <div className="grid grid-cols-7 gap-px bg-[#f0eeeb] p-px">{Array.from({ length: 35 }).map((_, index) => <div key={index} className="min-h-32 bg-white p-2"><Skeleton className="ml-auto h-6 w-6 rounded-full" /><Skeleton className="mt-3 h-8 w-full rounded-md" /></div>)}</div>
}
