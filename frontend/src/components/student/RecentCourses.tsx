'use client'

import Link from 'next/link'
import { useParams } from 'next/navigation'
import { BookOpen, ArrowRight } from 'lucide-react'
import { CourseDTO } from '@/types/course'
import { StudentGradeDTO } from '@/types/grade'
import { Skeleton } from '@/components/ui/skeleton'
import { useI18n } from '@/lib/i18n'

interface RecentCoursesProps {
  courses: CourseDTO[]
  isLoading: boolean
  gradesByCourse?: Record<string, StudentGradeDTO | undefined>
}

export function RecentCourses({ courses, isLoading, gradesByCourse = {} }: RecentCoursesProps) {
  const { org } = useParams<{ org: string }>()
  const { t, locale } = useI18n()
  const copy = locale === 'ru'
    ? { complete: 'Итоговая оценка', progress: 'Оценивание в процессе', components: 'компонента учтено' }
    : locale === 'kk'
      ? { complete: 'Қорытынды баға', progress: 'Бағалау жалғасуда', components: 'құрамдас есептелді' }
      : { complete: 'Final grade', progress: 'Assessment in progress', components: 'components recorded' }
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
          {published.map((course) => {
            const grades = gradesByCourse[course.id]
            const completedComponents = grades ? [grades.assignments_percent, grades.midterm_percent, grades.endterm_percent, grades.final_exam_percent].filter((value) => value != null).length : 0
            const finalPublished = Boolean(grades?.is_published && grades.final_grade !== undefined)
            return <Link
              key={course.id}
              href={`/${org}/student/courses/${course.id}`}
              className="group flex items-center gap-3 px-3.5 py-3 hover:bg-[#f7f7f5] transition-colors"
            >
              <BookOpen className="h-3.5 w-3.5 shrink-0 text-[#9b9b9b]" />
              <span className="min-w-0 flex-1"><span className="block truncate text-sm text-[#3b3b3b] group-hover:text-[#1a1a1a] transition-colors">{course.title}</span>{grades && <span className={finalPublished ? 'mt-0.5 block text-[11px] font-medium text-emerald-700' : 'mt-0.5 block text-[11px] text-[#8b7b58]'}>{finalPublished ? `${copy.complete}: ${grades.final_grade?.toFixed(1)}%` : grades.is_final_grade_ready ? 'Final result awaiting publication' : `${copy.progress} · ${completedComponents}/4 ${copy.components}`}</span>}</span>
              <ArrowRight className="h-3.5 w-3.5 text-[#c9c9c9] opacity-0 group-hover:opacity-100 transition-opacity shrink-0" />
            </Link>
          })}
        </div>
      )}
    </div>
  )
}
