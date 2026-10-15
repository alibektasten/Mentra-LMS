import assert from 'node:assert/strict'
import test from 'node:test'
import { completedLessonIDs, lessonCompletion } from './course-lesson-progress.ts'

const entry = (lessonID, completedAt) => ({
  id: lessonID, user_id: 'student', org_id: 'org', lesson_id: lessonID,
  completed_at: completedAt, created_at: '2026-10-15T00:00:00Z',
})

test('fresh course shows 0 of its own lessons despite progress in course A', () => {
  const completed = completedLessonIDs([entry('course-a-lesson', '2026-10-15T00:00:00Z')])
  assert.deepEqual(lessonCompletion(['course-b-lesson'], completed), { completed: 0, total: 1 })
})

test('unfinished rows do not count and completing one lesson advances only that lesson', () => {
  const lessons = ['first', 'second']
  assert.deepEqual(lessonCompletion(lessons, completedLessonIDs([entry('first')])), { completed: 0, total: 2 })
  assert.deepEqual(lessonCompletion(lessons, completedLessonIDs([entry('first', '2026-10-15T00:00:00Z')])), { completed: 1, total: 2 })
})

test('assignments never enlarge the lesson denominator and duplicate rows count once', () => {
  const completed = completedLessonIDs([entry('lesson', '2026-10-15T00:00:00Z'), entry('lesson', '2026-10-15T00:00:00Z')])
  assert.deepEqual(lessonCompletion(['lesson'], completed), { completed: 1, total: 1 })
})
