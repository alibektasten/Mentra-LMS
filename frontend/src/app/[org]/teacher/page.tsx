'use client'

import Link from 'next/link'
import { useParams } from 'next/navigation'
import { ArrowRight, BarChart3, BookOpen, BookOpenCheck, ClipboardCheck, Plus, Sparkles, Users } from 'lucide-react'
import { useTeacherOverview, useTeacherReviewQueue } from '@/lib/queries/teacher.queries'
import { PageHeader } from '@/components/shared/PageHeader'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

const quickLinks = [
  { key: 'courses', title: 'Мои курсы', description: 'Материалы, группы и оценки', icon: BookOpen },
  { key: 'students', title: 'Студенты', description: 'Участники ваших групп', icon: Users },
  { key: 'analytics', title: 'Аналитика', description: 'Обзор курсов и групп', icon: BarChart3 },
  { key: 'ai', title: 'AI-инструменты', description: 'Помощь в подготовке занятий', icon: Sparkles },
]

export default function TeacherDashboard() {
  const { org } = useParams<{ org: string }>()
  const { data: overview, isLoading: isOverviewLoading } = useTeacherOverview()
  const { data: reviewQueue, isLoading: isReviewQueueLoading } = useTeacherReviewQueue()
  const stats = [
    { label: 'Назначенные курсы', value: overview?.total_courses ?? 0, tone: 'text-[#1a1a1a]' },
    { label: 'Студенты', value: overview?.total_students ?? 0, tone: 'text-[#059669]' },
    { label: 'Группы', value: overview?.total_groups ?? 0, tone: 'text-[#2563eb]' },
    { label: 'На проверке', value: reviewQueue?.length ?? 0, tone: 'text-amber-700' },
  ]

  return (
    <div className="space-y-7">
      <PageHeader title="Панель преподавателя" description="Ключевые задачи и состояние ваших учебных курсов." actions={<Button asChild className="bg-[#059669] text-white hover:bg-[#047857]"><Link href={`/${org}/teacher/courses/new`}><Plus className="mr-2 h-4 w-4" />Создать курс</Link></Button>} />
      <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {isOverviewLoading || isReviewQueueLoading ? Array.from({ length: 4 }).map((_, index) => <Skeleton key={index} className="h-28 rounded-xl" />) : stats.map((stat) => <div key={stat.label} className="rounded-xl border border-[#e4e2de] bg-white px-5 py-4 shadow-sm"><p className="text-xs font-medium text-[#6b6b6b]">{stat.label}</p><p className={`mt-2 text-3xl font-semibold tracking-tight ${stat.tone}`}>{stat.value}</p></div>)}
      </section>
      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {quickLinks.map(({ key, title, description, icon: Icon }) => <Link key={key} href={`/${org}/teacher/${key}`} className="group rounded-xl border border-[#e4e2de] bg-white p-5 shadow-sm transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-md"><div className="mb-5 flex h-9 w-9 items-center justify-center rounded-lg bg-[#f0eeeb] text-[#4b5563] group-hover:bg-emerald-50 group-hover:text-emerald-700"><Icon className="h-4 w-4" /></div><div className="flex items-end justify-between gap-3"><div><h2 className="text-sm font-semibold text-[#1a1a1a]">{title}</h2><p className="mt-1 text-xs leading-relaxed text-[#6b6b6b]">{description}</p></div><ArrowRight className="h-4 w-4 shrink-0 text-[#9b9b9b] transition group-hover:translate-x-0.5 group-hover:text-emerald-700" /></div></Link>)}
      </section>
      <section className="rounded-xl border border-[#e4e2de] bg-white shadow-sm">
        <div className="flex items-center justify-between gap-4 border-b border-[#f0eeeb] px-5 py-4"><div className="flex items-center gap-2.5"><div className="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-50"><ClipboardCheck className="h-4 w-4 text-amber-700" /></div><div><h2 className="text-sm font-semibold text-[#1a1a1a]">Работы на проверку</h2><p className="text-xs text-[#9b9b9b]">Несоценённые задания из ваших курсов</p></div></div>{!isReviewQueueLoading && <span className="rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700">{reviewQueue?.length ?? 0}</span>}</div>
        {isReviewQueueLoading ? <div className="space-y-3 p-5"><Skeleton className="h-12 w-full" /><Skeleton className="h-12 w-full" /></div> : !reviewQueue?.length ? <div className="flex items-center gap-2 p-5 text-sm text-[#6b6b6b]"><BookOpenCheck className="h-4 w-4 text-emerald-600" />Все отправленные работы уже проверены.</div> : <div className="divide-y divide-[#f0eeeb]">{reviewQueue.slice(0, 5).map((item) => <div key={item.submission_id} className="flex items-center justify-between gap-4 px-5 py-3"><div className="min-w-0"><p className="truncate text-sm font-medium text-[#1a1a1a]">{item.assignment_title}</p><p className="truncate text-xs text-[#6b6b6b]">{item.student_name || 'Студент'} · {item.course_title}</p></div><Link href={`/${org}/teacher/courses/${item.course_id}`} className="shrink-0 text-xs font-medium text-[#059669] hover:underline">Проверить</Link></div>)}</div>}
      </section>
    </div>
  )
}
