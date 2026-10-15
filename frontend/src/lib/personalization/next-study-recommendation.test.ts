import assert from 'node:assert/strict'
import test from 'node:test'
import { getNextStudyRecommendation } from './next-study-recommendation'
import type { CourseDTO } from '@/types/course'
import type { StudentGradeDTO } from '@/types/grade'
import type { StudentLessonContext } from '@/lib/queries/student-lesson-context.queries'

const course = (id: string, title: string): CourseDTO => ({
  id, title, description: '', status: 'published', org_id: 'org', teacher_id: 'teacher', created_at: '', updated_at: '',
})

const grades = (items: StudentGradeDTO['items']): StudentGradeDTO => ({
  student_id: 'student', items, total_earned: 0, total_possible: 0, assignments_percent: null,
  midterm_percent: null, endterm_percent: null, final_exam_percent: null, midterm_register: null,
  endterm_register: null, is_final_grade_ready: false, override_grade: null, override_note: '', is_published: false, published_at: null,
})

const item = (id: string, title: string, score: number, max = 100): StudentGradeDTO['items'][number] => ({
  item_id: id, item_type: 'assignment', title, score, max_points: max, status: 'graded', feedback: '', submitted_at: null, graded_at: null,
})

const lesson = (id: string, courseID: string, order: number): StudentLessonContext => ({
  lessonID: id, lessonTitle: `Lesson ${order}`, moduleTitle: 'Module', moduleOrder: 1, lessonOrder: order, courseID, courseTitle: 'Course',
})

test('recommends the first available lesson in the course with the lowest real source mark', () => {
  const first = course('course-1', 'Course one')
  const second = course('course-2', 'Course two')
  const contexts = new Map([
    ['done', lesson('done', first.id, 1)],
    ['next', lesson('next', first.id, 2)],
    ['other', lesson('other', second.id, 1)],
  ])
  const result = getNextStudyRecommendation(
    [first, second],
    { [first.id]: grades([item('a1', 'Arrays', 45)]), [second.id]: grades([item('a2', 'Functions', 65)]) },
    contexts,
    [{ id: 'p', user_id: 'student', org_id: 'org', lesson_id: 'done', completed_at: '2026-01-01', created_at: '2026-01-01' }],
  )
  assert.equal(result.status, 'recommendation')
  if (result.status === 'recommendation') {
    assert.equal(result.assessment.title, 'Arrays')
    assert.equal(result.lesson.lessonID, 'next')
    assert.equal(result.percentage, 45)
  }
})

test('does not invent a recommendation when there are no visible graded items', () => {
  const result = getNextStudyRecommendation([course('course-1', 'Course')], {}, new Map(), [])
  assert.deepEqual(result, { status: 'no-graded-work' })
})

test('ignores ungraded and zero-point records instead of treating them as weak evidence', () => {
  const current = course('course-1', 'Course')
  const ungraded = { ...item('a1', 'Draft', 0), status: 'submitted' as const, score: null }
  const zeroPoint = { ...item('a2', 'Invalid scale', 0), max_points: 0 }
  const result = getNextStudyRecommendation([current], { [current.id]: grades([ungraded, zeroPoint]) }, new Map(), [])
  assert.deepEqual(result, { status: 'no-graded-work' })
})

test('treats a real zero score as the weakest visible evidence', () => {
  const current = course('course-1', 'Course')
  const result = getNextStudyRecommendation(
    [current], { [current.id]: grades([item('a1', 'Arrays', 0)]) },
    new Map([['lesson', lesson('lesson', current.id, 1)]]), [],
  )
  assert.equal(result.status, 'recommendation')
  if (result.status === 'recommendation') assert.equal(result.percentage, 0)
})

test('excludes an unpublished course even when it has the lowest score', () => {
  const visible = course('course-1', 'Visible')
  const hidden = { ...course('course-2', 'Hidden'), status: 'draft' as const }
  const result = getNextStudyRecommendation(
    [hidden, visible],
    { [hidden.id]: grades([item('hidden', 'Draft assessment', 0)]), [visible.id]: grades([item('visible', 'Visible assessment', 80)]) },
    new Map(), [],
  )
  assert.deepEqual(result, { status: 'all-current-work-mastered' })
})

test('reports mastery when all visible graded items meet the threshold', () => {
  const current = course('course-1', 'Course')
  const result = getNextStudyRecommendation([current], { [current.id]: grades([item('a1', 'Arrays', 70)]) }, new Map(), [])
  assert.deepEqual(result, { status: 'all-current-work-mastered' })
})

test('keeps weak evidence but does not invent material when the course is complete', () => {
  const current = course('course-1', 'Course')
  const result = getNextStudyRecommendation(
    [current], { [current.id]: grades([item('a1', 'Arrays', 30)]) },
    new Map([['done', lesson('done', current.id, 1)]]),
    [{ id: 'p', user_id: 'student', org_id: 'org', lesson_id: 'done', completed_at: '2026-01-01', created_at: '2026-01-01' }],
  )
  assert.equal(result.status, 'no-next-material')
})

test('uses a stable title and item-id tie-breaker for equal low marks', () => {
  const alpha = course('course-a', 'Alpha')
  const beta = course('course-b', 'Beta')
  const contexts = new Map([['alpha-lesson', lesson('alpha-lesson', alpha.id, 1)], ['beta-lesson', lesson('beta-lesson', beta.id, 1)]])
  const result = getNextStudyRecommendation(
    [beta, alpha],
    { [beta.id]: grades([item('b', 'Same', 50)]), [alpha.id]: grades([item('a', 'Same', 50)]) },
    contexts,
    [],
  )
  assert.equal(result.status, 'recommendation')
  if (result.status === 'recommendation') assert.equal(result.course.id, alpha.id)
})
