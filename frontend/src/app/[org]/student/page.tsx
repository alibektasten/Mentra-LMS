'use client'

import { useMemo } from 'react'
import { useCourses } from '@/lib/queries/courses.queries'
import { useProgress } from '@/lib/queries/progress.queries'
import { WelcomeHero } from '@/components/student/WelcomeHero'
import { StatsStrip } from '@/components/student/StatsStrip'
import { AIStudyPanel } from '@/components/student/AIStudyPanel'
import { QuickActions } from '@/components/student/QuickActions'
import { MiniCalendar } from '@/components/student/MiniCalendar'
import { RecentCourses } from '@/components/student/RecentCourses'
import { ActivityFeed } from '@/components/student/ActivityFeed'

export default function StudentDashboard() {
  const { data: coursesData, isLoading: coursesLoading } = useCourses({ page: 1, page_size: 20 })
  const { data: progress = [] } = useProgress()

  const courses = coursesData?.data ?? []
  const firstCourse = courses.find((c) => c.status === 'published')

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
          <RecentCourses courses={courses} isLoading={coursesLoading} />
          <ActivityFeed progress={progress} />
        </div>

        {/* Right column — secondary / tools */}
        <div className="space-y-8">
          <QuickActions lastProgress={lastProgress} />
          <MiniCalendar progress={progress} />
          <AIStudyPanel courseId={firstCourse?.id} courseTitle={firstCourse?.title} />
        </div>
      </div>
    </div>
  )
}
