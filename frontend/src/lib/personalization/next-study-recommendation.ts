import type { CourseDTO } from '@/types/course'
import type { GradeItemDTO, StudentGradeDTO } from '@/types/grade'
import type { ProgressDTO } from '@/types/progress'
import type { StudentLessonContext } from '@/lib/queries/student-lesson-context.queries'

export const LOW_SCORE_THRESHOLD = 70

type GradeEvidence = {
  course: CourseDTO
  item: GradeItemDTO
  percentage: number
}

export type NextStudyRecommendation =
  | {
      status: 'recommendation'
      course: CourseDTO
      assessment: GradeItemDTO
      percentage: number
      lesson: StudentLessonContext
    }
  | {
      status: 'no-graded-work'
    }
  | {
      status: 'all-current-work-mastered'
    }
  | {
      status: 'no-next-material'
      course: CourseDTO
      assessment: GradeItemDTO
      percentage: number
    }

/**
 * Creates a transparent course-level recommendation from source marks that are
 * already visible to the signed-in student. Assessment-to-topic mapping is not
 * available in the current data model, so this deliberately never infers one.
 */
export function getNextStudyRecommendation(
  courses: CourseDTO[],
  gradesByCourse: Record<string, StudentGradeDTO | undefined>,
  lessonContexts: Map<string, StudentLessonContext>,
  progress: ProgressDTO[],
): NextStudyRecommendation {
  const publishedCourses = courses.filter((course) => course.status === 'published')
  const evidence = publishedCourses.flatMap((course) => gradeEvidence(course, gradesByCourse[course.id]))

  if (evidence.length === 0) return { status: 'no-graded-work' }

  const weakest = evidence.sort(compareEvidence)[0]
  if (!weakest || weakest.percentage >= LOW_SCORE_THRESHOLD) {
    return { status: 'all-current-work-mastered' }
  }

  const completedLessonIDs = new Set(
    progress.filter((item) => Boolean(item.completed_at)).map((item) => item.lesson_id),
  )
  const nextLesson = [...lessonContexts.values()]
    .filter((lesson) => lesson.courseID === weakest.course.id && !completedLessonIDs.has(lesson.lessonID))
    .sort(compareLesson)[0]

  if (!nextLesson) {
    return {
      status: 'no-next-material',
      course: weakest.course,
      assessment: weakest.item,
      percentage: weakest.percentage,
    }
  }

  return {
    status: 'recommendation',
    course: weakest.course,
    assessment: weakest.item,
    percentage: weakest.percentage,
    lesson: nextLesson,
  }
}

function gradeEvidence(course: CourseDTO, grades?: StudentGradeDTO): GradeEvidence[] {
  return (grades?.items ?? []).flatMap((item) => {
    if (item.status !== 'graded' || item.score == null || item.max_points <= 0) return []
    return [{ course, item, percentage: (item.score / item.max_points) * 100 }]
  })
}

function compareEvidence(a: GradeEvidence, b: GradeEvidence): number {
  return a.percentage - b.percentage
    || compareText(a.course.title, b.course.title)
    || compareText(a.item.title, b.item.title)
    || compareText(a.item.item_id, b.item.item_id)
}

function compareLesson(a: StudentLessonContext, b: StudentLessonContext): number {
  return a.moduleOrder - b.moduleOrder
    || a.lessonOrder - b.lessonOrder
    || compareText(a.lessonTitle, b.lessonTitle)
    || compareText(a.lessonID, b.lessonID)
}

function compareText(a: string, b: string): number {
  return a < b ? -1 : a > b ? 1 : 0
}
