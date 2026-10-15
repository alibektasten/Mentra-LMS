'use client'

import { useMemo } from 'react'
import { useQueries } from '@tanstack/react-query'
import { useCourses } from '@/lib/queries/courses.queries'
import { useProgress } from '@/lib/queries/progress.queries'
import { useStudentLessonContextsState } from '@/lib/queries/student-lesson-context.queries'
import { getMyGrades, getUpcomingDeadlines } from '@/lib/api/grades'
import { gradeKeys } from '@/lib/queries/grades.queries'
import { WelcomeHero } from '@/components/student/WelcomeHero'
import { StatsStrip } from '@/components/student/StatsStrip'
import { AIStudyPanel } from '@/components/student/AIStudyPanel'
import { QuickActions } from '@/components/student/QuickActions'
import { MiniCalendar } from '@/components/student/MiniCalendar'
import { RecentCourses } from '@/components/student/RecentCourses'
import { ActivityFeed } from '@/components/student/ActivityFeed'
import { NextStudyRecommendation } from '@/components/student/NextStudyRecommendation'
import { getNextStudyRecommendation } from '@/lib/personalization/next-study-recommendation'

export default function StudentDashboard() {
  const { data: coursesData, isLoading: coursesLoading, isError: coursesError } = useCourses({ page: 1, page_size: 20 })
  const { data: progress = [], isLoading: progressLoading, isError: progressError } = useProgress()

  const courses = coursesData?.data ?? []
  const lessonContextState = useStudentLessonContextsState(courses)
  const lessonContexts = lessonContextState.contexts
  const publishedCourses = useMemo(() => courses.filter((course) => course.status === 'published'), [courses])
  const firstCourse = courses.find((c) => c.status === 'published')
  const deadlineQueries = useQueries({
    queries: publishedCourses.map((course) => ({
      queryKey: gradeKeys.deadlines(course.id),
      queryFn: () => getUpcomingDeadlines(course.id),
      staleTime: 60_000,
    })),
  })
  const gradeQueries = useQueries({
    queries: publishedCourses.map((course) => ({
      queryKey: gradeKeys.myGrades(course.id),
      queryFn: () => getMyGrades(course.id),
      staleTime: 60_000,
    })),
  })
  const gradesByCourse = useMemo(() => Object.fromEntries(
    publishedCourses.map((course, index) => [course.id, gradeQueries[index]?.data])
  ), [gradeQueries, publishedCourses])
  const recommendationLoading = coursesLoading || progressLoading || lessonContextState.isLoading || gradeQueries.some((query) => query.isLoading || query.isPending)
  const recommendationUnavailable = coursesError || progressError || lessonContextState.isError || gradeQueries.some((query) => query.isError)
  const nextStudyRecommendation = useMemo(
    () => getNextStudyRecommendation(publishedCourses, gradesByCourse, lessonContexts, progress),
    [gradesByCourse, lessonContexts, progress, publishedCourses],
  )
  const deadlines = useMemo(() => deadlineQueries.flatMap((query, index) => {
    const course = publishedCourses[index]
    if (!course) return []
    return (query.data ?? []).map((deadline) => ({
      id: `${course.id}-${deadline.item_id}`,
      title: deadline.title,
      dueDate: deadline.due_date,
      courseID: course.id,
      courseTitle: course.title,
      submitted: deadline.submitted,
      itemType: deadline.item_type,
    }))
  }), [deadlineQueries, publishedCourses])

  const lastProgress = useMemo(
    () =>
      [...progress]
        .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
        .find((p) => !p.completed_at),
    [progress]
  )

  const streak = useMemo(() => {
    const days = new Set(
      progress.map((p) => {
        const d = new Date(p.created_at)
        return `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`
      })
    )
    let count = 0
    const cursor = new Date()
    while (true) {
      const key = `${cursor.getFullYear()}-${cursor.getMonth()}-${cursor.getDate()}`
      if (days.has(key)) {
        count++
        cursor.setDate(cursor.getDate() - 1)
      } else {
        break
      }
    }
    return count
  }, [progress])

  return (
    <div className="max-w-5xl mx-auto">
      {/* Page header */}
      <div className="mb-8">
        <WelcomeHero streak={streak} />
      </div>

      {/* Stats row */}
      <div className="mb-8">
        <StatsStrip progress={progress} courses={courses} />
      </div>

      {/* Main content: two-column */}
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_280px] lg:gap-8">
        {/* Left column — primary content */}
        <div className="space-y-8">
          <NextStudyRecommendation
            recommendation={nextStudyRecommendation}
            state={recommendationLoading ? 'loading' : recommendationUnavailable ? 'unavailable' : 'ready'}
          />
          <RecentCourses courses={courses} isLoading={coursesLoading} gradesByCourse={gradesByCourse} />
          <ActivityFeed progress={progress} lessonContexts={lessonContexts} />
        </div>

        {/* Right column — secondary / tools */}
        <div className="space-y-8">
          <QuickActions lastProgress={lastProgress} />
          <MiniCalendar progress={progress} deadlines={deadlines} />
          <AIStudyPanel courseId={firstCourse?.id} courseTitle={firstCourse?.title} />
        </div>
      </div>
    </div>
  )
}
