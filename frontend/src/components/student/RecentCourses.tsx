'use client'

import Link from 'next/link'
import { useParams } from 'next/navigation'
import { BookOpen, ArrowRight } from 'lucide-react'
import { CourseDTO } from '@/types/course'
import { Skeleton } from '@/components/ui/skeleton'
import { useI18n } from '@/lib/i18n'

interface RecentCoursesProps {
  courses: CourseDTO[]
  isLoading: boolean
}

export function RecentCourses({ courses, isLoading }: RecentCoursesProps) {
  const { org } = useParams<{ org: string }>()
  const { t } = useI18n()
  const published = courses.filter((c) => c.status === 'published').slice(0, 5)

  return (
    <div className="flex flex-col">
      <div className="flex items-center justify-between mb-3">
        <p className="text-xs font-semibold text-[#6b6b6b] uppercase tracking-widest">{t('courses')}</p>
        <Link
          href={`/${org}/student/courses`}
          className="text-xs text-[#9b9b9b] hover:text-[#1a1a1a] transition-colors"
        >
          {t('viewAll')}
        </Link>
      </div>

      {isLoading ? (
        <div className="space-y-1">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-9 rounded-md" />)}
        </div>
      ) : published.length === 0 ? (
        <p className="text-sm text-[#9b9b9b] py-2">{t('noCoursesYet')}</p>
      ) : (
        <div className="divide-y divide-[#e8e8e6] border border-[#e8e8e6] rounded-lg overflow-hidden">
          {published.map((course) => (
            <Link
              key={course.id}
              href={`/${org}/student/courses/${course.id}`}
              className="group flex items-center gap-3 px-3.5 py-2.5 hover:bg-[#f7f7f5] transition-colors"
            >
              <BookOpen className="h-3.5 w-3.5 shrink-0 text-[#9b9b9b]" />
              <span className="flex-1 text-sm text-[#3b3b3b] group-hover:text-[#1a1a1a] transition-colors truncate">
                {course.title}
              </span>
              <ArrowRight className="h-3.5 w-3.5 text-[#c9c9c9] opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
