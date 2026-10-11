'use client'

import { useParams } from 'next/navigation'
import { useCourses } from '@/lib/queries/courses.queries'
import { CourseGrid } from '@/components/courses/CourseGrid'
import { useI18n } from '@/lib/i18n'

export default function StudentCoursesPage() {
  const { org } = useParams<{ org: string }>()
  const { t } = useI18n()
  const { data, isLoading } = useCourses()

  const publishedCourses = (data?.data ?? []).filter((c) => c.status === 'published')

  return (
    <div className="max-w-5xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold tracking-tight text-[#1a1a1a]">{t('myCourses')}</h1>
        <p className="mt-1 text-sm text-[#9b9b9b]">
          {publishedCourses.length > 0
            ? `${publishedCourses.length} ${t('availableCourses').toLowerCase()}`
            : t('noCoursesYet')}
        </p>
      </div>

      <CourseGrid
        courses={publishedCourses}
        isLoading={isLoading}
        basePath={`/${org}/student`}
      />
    </div>
  )
}
