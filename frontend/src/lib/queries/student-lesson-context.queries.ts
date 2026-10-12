import { useMemo } from 'react'
import { useQueries } from '@tanstack/react-query'
import { CourseDTO } from '@/types/course'
import { listModules } from '@/lib/api/modules'
import { listLessons } from '@/lib/api/lessons'
import { moduleKeys } from '@/lib/queries/modules.queries'
import { lessonKeys } from '@/lib/queries/lessons.queries'

export interface StudentLessonContext {
  lessonID: string
  lessonTitle: string
  moduleTitle: string
  courseID: string
  courseTitle: string
}

/** Builds a lookup for progress records so activity can name the exact lesson and course. */
export function useStudentLessonContexts(courses: CourseDTO[]) {
  const publishedCourses = useMemo(() => courses.filter((course) => course.status === 'published'), [courses])
  const moduleQueries = useQueries({
    queries: publishedCourses.map((course) => ({
      queryKey: moduleKeys.list(course.id),
      queryFn: () => listModules(course.id),
      staleTime: 60_000,
    })),
  })

  const modules = useMemo(() => moduleQueries.flatMap((query, index) => {
    const course = publishedCourses[index]
    if (!course) return []
    return (query.data ?? []).map((module) => ({ module, course }))
  }), [moduleQueries, publishedCourses])

  const lessonQueries = useQueries({
    queries: modules.map(({ module }) => ({
      queryKey: lessonKeys.list(module.id),
      queryFn: () => listLessons(module.id),
      staleTime: 60_000,
    })),
  })

  return useMemo(() => {
    const contexts = new Map<string, StudentLessonContext>()
    lessonQueries.forEach((query, index) => {
      const parent = modules[index]
      if (!parent) return
      ;(query.data ?? []).forEach((lesson) => contexts.set(lesson.id, {
        lessonID: lesson.id,
        lessonTitle: lesson.title,
        moduleTitle: parent.module.title,
        courseID: parent.course.id,
        courseTitle: parent.course.title,
      }))
    })
    return contexts
  }, [lessonQueries, modules])
}
