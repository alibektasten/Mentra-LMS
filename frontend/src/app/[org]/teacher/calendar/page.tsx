'use client'

import Link from 'next/link'
import { useMemo } from 'react'
import { useParams } from 'next/navigation'
import { useQueries } from '@tanstack/react-query'
import { AlertCircle, ArrowRight, CalendarDays, Clock3, GraduationCap } from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { Skeleton } from '@/components/ui/skeleton'
import * as examsApi from '@/lib/api/exams'
import { examKeys } from '@/lib/queries/exams.queries'
import { useCourses } from '@/lib/queries/courses.queries'
import { useI18n } from '@/lib/i18n'
import { CourseDTO } from '@/types/course'
import { ExamListItemDTO } from '@/types/exam'

type CalendarEvent = {
  exam: ExamListItemDTO
  course: CourseDTO
  dueDate: Date
}

const copyByLocale = {
  ru: {
    title: 'Календарь преподавателя',
    description: 'Ближайшие контрольные точки по вашим курсам.',
    upcoming: 'Ближайшие события',
    scheduled: 'запланировано',
    noEvents: 'Ближайших экзаменов пока нет.',
    noEventsDescription: 'Назначьте даты Midterm, Endterm или Final в настройках экзаменов курса.',
    noCourses: 'У вас пока нет назначенных курсов.',
    loading: 'Загружаем расписание…',
    error: 'Не удалось загрузить часть расписания. Попробуйте обновить страницу.',
    course: 'Открыть курс',
    midterm: 'Рубежный контроль',
    endterm: 'Конец семестра',
    final: 'Финальный экзамен',
    other: 'Экзамен',
  },
  kk: {
    title: 'Оқытушы күнтізбесі',
    description: 'Курстарыңыз бойынша жақын бақылау кезеңдері.',
    upcoming: 'Жақын оқиғалар',
    scheduled: 'жоспарланған',
    noEvents: 'Жақын емтихандар әзірге жоқ.',
    noEventsDescription: 'Курс емтихандары баптауларында Midterm, Endterm немесе Final күндерін белгілеңіз.',
    noCourses: 'Сізге әлі курс тағайындалмаған.',
    loading: 'Кесте жүктелуде…',
    error: 'Кестенің бір бөлігін жүктеу мүмкін болмады. Бетті жаңартып көріңіз.',
    course: 'Курсты ашу',
    midterm: 'Аралық бақылау',
    endterm: 'Семестр соңы',
    final: 'Қорытынды емтихан',
    other: 'Емтихан',
  },
  en: {
    title: 'Teacher calendar',
    description: 'Upcoming assessment milestones across your courses.',
    upcoming: 'Upcoming events',
    scheduled: 'scheduled',
    noEvents: 'No upcoming exams yet.',
    noEventsDescription: 'Set Midterm, Endterm, or Final dates in a course’s exam settings.',
    noCourses: 'You do not have any assigned courses yet.',
    loading: 'Loading schedule…',
    error: 'Part of the schedule could not be loaded. Try refreshing the page.',
    course: 'Open course',
    midterm: 'Midterm',
    endterm: 'Endterm',
    final: 'Final exam',
    other: 'Exam',
  },
} as const

const dateKey = (date: Date) => `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`

export default function TeacherCalendarPage() {
  const { org } = useParams<{ org: string }>()
  const { locale } = useI18n()
  const copy = copyByLocale[locale]
  const languageTag = locale === 'kk' ? 'kk-KZ' : locale === 'ru' ? 'ru-RU' : 'en-US'
  const { data: coursesResponse, isLoading: coursesLoading, isError: coursesError } = useCourses({ page: 1, page_size: 100 })
  const courses = coursesResponse?.data ?? []
  const examQueries = useQueries({
    queries: courses.map((course) => ({
      queryKey: examKeys.byCourse(course.id),
      queryFn: () => examsApi.listExams(course.id),
      enabled: Boolean(course.id),
      retry: false,
    })),
  })

  const events = useMemo<CalendarEvent[]>(() => {
    const now = new Date()
    return examQueries.flatMap((query, index) => {
      const course = courses[index]
      if (!course || !query.data) return []
      return query.data.flatMap((exam) => {
        if (!exam.due_date) return []
        const dueDate = new Date(exam.due_date)
        if (Number.isNaN(dueDate.getTime()) || dueDate < now) return []
        return [{ exam, course, dueDate }]
      })
    }).sort((left, right) => left.dueDate.getTime() - right.dueDate.getTime())
  }, [courses, examQueries])

  const eventsByDay = useMemo(() => {
    const grouped = new Map<string, CalendarEvent[]>()
    events.forEach((event) => {
      const key = dateKey(event.dueDate)
      grouped.set(key, [...(grouped.get(key) ?? []), event])
    })
    return Array.from(grouped.entries())
  }, [events])

  const eventsLoading = coursesLoading || (courses.length > 0 && examQueries.some((query) => query.isLoading))
  const hasExamError = coursesError || examQueries.some((query) => query.isError)

  return (
    <div className="space-y-6">
      <PageHeader
        title={copy.title}
        description={copy.description}
        actions={<div className="flex items-center gap-2 rounded-lg border border-[#e4e2de] bg-white px-3 py-2 text-xs font-medium text-[#55554f]"><CalendarDays className="h-4 w-4 text-emerald-600" />{events.length} {copy.scheduled}</div>}
      />

      {hasExamError && <div className="flex items-start gap-2 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900"><AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />{copy.error}</div>}

      <section className="rounded-2xl border border-[#e4e2de] bg-white shadow-sm">
        <div className="flex items-center gap-3 border-b border-[#f0eeeb] px-5 py-4">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-emerald-50 text-emerald-700"><CalendarDays className="h-4 w-4" /></div>
          <div><h2 className="text-sm font-semibold text-[#1a1a1a]">{copy.upcoming}</h2><p className="text-xs text-[#8a8983]">{copy.description}</p></div>
        </div>

        {eventsLoading ? <div className="space-y-3 p-5">{Array.from({ length: 4 }).map((_, index) => <Skeleton key={index} className="h-20 w-full rounded-xl" />)}</div>
          : courses.length === 0 ? <EmptyState title={copy.noCourses} description={copy.noEventsDescription} />
            : eventsByDay.length === 0 ? <EmptyState title={copy.noEvents} description={copy.noEventsDescription} />
              : <div className="divide-y divide-[#f0eeeb]">
                {eventsByDay.map(([key, dayEvents]) => {
                  const date = dayEvents[0].dueDate
                  return <div key={key} className="px-5 py-5">
                    <div className="mb-3 flex items-center gap-2"><div className="flex h-8 min-w-8 flex-col items-center justify-center rounded-lg bg-[#f0faf5] px-1 text-emerald-700"><span className="text-[10px] font-semibold uppercase leading-none">{date.toLocaleDateString(languageTag, { month: 'short' })}</span><span className="mt-0.5 text-sm font-bold leading-none">{date.getDate()}</span></div><p className="text-sm font-semibold capitalize text-[#3c3c36]">{date.toLocaleDateString(languageTag, { weekday: 'long', day: 'numeric', month: 'long' })}</p></div>
                    <div className="space-y-2 pl-0 sm:pl-10">
                      {dayEvents.map(({ exam, course, dueDate }) => {
                        const assessmentLabel = copy[exam.assessment_type]
                        return <div key={exam.id} className="flex flex-col gap-3 rounded-xl border border-[#e8e7e2] bg-[#fcfcfa] p-4 sm:flex-row sm:items-center sm:justify-between">
                          <div className="min-w-0"><div className="flex flex-wrap items-center gap-2"><span className="rounded-full bg-violet-50 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-violet-700">{assessmentLabel}</span><span className="text-xs text-[#8f8f89]"><Clock3 className="mr-1 inline h-3 w-3" />{dueDate.toLocaleTimeString(languageTag, { hour: '2-digit', minute: '2-digit' })}</span></div><p className="mt-2 truncate text-sm font-semibold text-[#23231f]">{exam.title}</p><p className="mt-0.5 truncate text-xs text-[#6f6f68]">{course.title}</p></div>
                          <Link href={`/${org}/teacher/courses/${course.id}`} className="inline-flex shrink-0 items-center gap-1 text-xs font-semibold text-emerald-700 hover:text-emerald-800 hover:underline">{copy.course}<ArrowRight className="h-3.5 w-3.5" /></Link>
                        </div>
                      })}
                    </div>
                  </div>
                })}
              </div>}
      </section>
    </div>
  )
}

function EmptyState({ title, description }: { title: string; description: string }) {
  return <div className="flex flex-col items-center justify-center px-6 py-16 text-center"><div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#f0faf5]"><GraduationCap className="h-6 w-6 text-emerald-600" /></div><p className="mt-4 text-sm font-semibold text-[#3b3b36]">{title}</p><p className="mt-1 max-w-md text-xs leading-relaxed text-[#898981]">{description}</p></div>
}
