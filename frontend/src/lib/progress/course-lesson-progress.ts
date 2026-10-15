import type { ProgressDTO } from '../../types/progress'

export function completedLessonIDs(progress: ProgressDTO[]): Set<string> {
  return new Set(progress.filter((entry) => Boolean(entry.completed_at)).map((entry) => entry.lesson_id))
}

export function lessonCompletion(lessonIDs: string[], completed: Set<string>): { completed: number; total: number } {
  return { completed: lessonIDs.filter((id) => completed.has(id)).length, total: lessonIDs.length }
}
