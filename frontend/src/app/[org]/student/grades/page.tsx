'use client'

import Link from 'next/link'
import { useParams } from 'next/navigation'
import { useQueries } from '@tanstack/react-query'
import { Award, BookOpen, CheckCircle2, ChevronRight, ClipboardCheck, Clock, GraduationCap } from 'lucide-react'
import { useCourses } from '@/lib/queries/courses.queries'
import { gradeKeys } from '@/lib/queries/grades.queries'
import { getMyGrades } from '@/lib/api/grades'
import { Skeleton } from '@/components/ui/skeleton'

function letterGrade(percent: number) {
  if (percent >= 90) return 'A'
  if (percent >= 80) return 'B'
  if (percent >= 70) return 'C'
  if (percent >= 60) return 'D'
  return 'F'
}

export default function StudentGradesPage() {
  const { org } = useParams<{ org: string }>()
  const { data: courseResponse, isLoading: coursesLoading } = useCourses({ page: 1, page_size: 100 })
  const courses = courseResponse?.data ?? []
  const gradeQueries = useQueries({
    queries: courses.map((course) => ({
      queryKey: gradeKeys.myGrades(course.id),
      queryFn: () => getMyGrades(course.id),
      enabled: Boolean(course.id),
    })),
  })

  const isLoading = coursesLoading || gradeQueries.some((query) => query.isLoading)
  const rows = courses.map((course, index) => ({ course, grades: gradeQueries[index]?.data }))
  const gradedCourses = rows.filter(({ grades }) => grades?.is_final_grade_ready && grades.is_published)
  const totalPossible = gradedCourses.reduce((sum, { grades }) => sum + (grades?.total_possible ?? 0), 0)
  const totalEarned = gradedCourses.reduce((sum, { grades }) => sum + (grades?.total_earned ?? 0), 0)
  const overallPercent = totalPossible ? (totalEarned / totalPossible) * 100 : 0
  const awaitingReview = rows.reduce((sum, { grades }) => sum + (grades?.items.filter((item) => item.status === 'submitted').length ?? 0), 0)

  return (
    <div className="mx-auto max-w-5xl">
      <header className="mb-8 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div>
          <p className="text-xs font-bold uppercase tracking-[0.2em] text-emerald-700">Academic record</p>
          <h2 className="mt-2 text-3xl font-bold tracking-tight text-[#1a1a1a]">My grades</h2>
          <p className="mt-2 text-sm text-[#6b6b6b]">A single view of scores, feedback, and work still being reviewed.</p>
        </div>
        <Link href={`/${org}/student/courses`} className="inline-flex items-center gap-2 text-sm font-semibold text-emerald-700 hover:text-emerald-800"><BookOpen className="h-4 w-4" />Open courses</Link>
      </header>

      {isLoading ? <div className="space-y-4"><Skeleton className="h-36 rounded-2xl" />{[1, 2, 3].map((item) => <Skeleton key={item} className="h-24 rounded-xl" />)}</div> : (
        <>
          <section className="grid gap-4 sm:grid-cols-3">
            <Metric icon={Award} label="Current average" value={totalPossible ? `${overallPercent.toFixed(1)}%` : '—'} detail={totalPossible ? `${totalEarned} / ${totalPossible} points` : 'No graded work yet'} tone="emerald" />
            <Metric icon={GraduationCap} label="Academic standing" value={totalPossible ? letterGrade(overallPercent) : '—'} detail={totalPossible ? `${gradedCourses.length} course${gradedCourses.length === 1 ? '' : 's'} with grades` : 'Complete an assignment to begin'} tone="violet" />
            <Metric icon={Clock} label="Awaiting review" value={String(awaitingReview)} detail={awaitingReview ? 'Submitted work needs instructor review' : 'No work waiting for review'} tone="amber" />
          </section>

          <section className="mt-7 overflow-hidden rounded-2xl border border-[#e8e8e4] bg-white shadow-sm">
            <div className="flex items-center gap-3 border-b border-[#ecebe7] px-5 py-4"><ClipboardCheck className="h-5 w-5 text-emerald-700" /><div><h3 className="font-semibold text-[#1a1a1a]">Course results</h3><p className="text-xs text-[#8b8a84]">Open a course for detailed feedback on each assignment.</p></div></div>
            {rows.length === 0 ? <div className="py-16 text-center"><BookOpen className="mx-auto h-8 w-8 text-[#c5c4bf]" /><p className="mt-3 text-sm text-[#77766f]">No courses are assigned yet.</p></div> : (
              <div className="divide-y divide-[#efeeea]">
                {rows.map(({ course, grades }) => {
                  const graded = grades?.items.filter((item) => item.status === 'graded').length ?? 0
                  const submitted = grades?.items.filter((item) => item.status === 'submitted').length ?? 0
                  const hasGrades = Boolean(grades?.is_published && grades.final_grade !== undefined)
                  return <Link key={course.id} href={`/${org}/student/courses/${course.id}`} className="group flex items-center gap-4 px-5 py-4 transition hover:bg-[#fafaf8]">
                    <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-emerald-50 text-emerald-700"><BookOpen className="h-5 w-5" /></span>
                    <div className="min-w-0 flex-1"><p className="truncate font-semibold text-[#272722]">{course.title}</p><p className="mt-1 text-xs text-[#85847e]">{graded} graded · {submitted} awaiting review</p></div>
                    {hasGrades ? <div className="text-right"><p className="text-lg font-bold text-[#22221f]">{grades?.final_grade?.toFixed(1)}%</p><p className="text-xs font-medium text-emerald-700">{letterGrade(grades?.final_grade ?? 0)}</p></div> : <span className="text-xs font-medium text-amber-700">{grades?.is_final_grade_ready ? 'Awaiting publication' : 'Assessment in progress'}</span>}
                    <ChevronRight className="h-4 w-4 text-[#aaa9a2] transition group-hover:translate-x-0.5 group-hover:text-emerald-700" />
                  </Link>
                })}
              </div>
            )}
          </section>
        </>
      )}
    </div>
  )
}

function Metric({ icon: Icon, label, value, detail, tone }: { icon: typeof Award; label: string; value: string; detail: string; tone: 'emerald' | 'violet' | 'amber' }) {
  const tones = { emerald: 'bg-emerald-50 text-emerald-700', violet: 'bg-violet-50 text-violet-700', amber: 'bg-amber-50 text-amber-700' }
  return <div className="rounded-2xl border border-[#e8e8e4] bg-white p-5 shadow-sm"><span className={`flex h-9 w-9 items-center justify-center rounded-xl ${tones[tone]}`}><Icon className="h-4 w-4" /></span><p className="mt-4 text-xs font-semibold uppercase tracking-wider text-[#85847e]">{label}</p><p className="mt-1 text-2xl font-bold text-[#24241f]">{value}</p><p className="mt-1 text-xs text-[#85847e]">{detail}</p></div>
}
