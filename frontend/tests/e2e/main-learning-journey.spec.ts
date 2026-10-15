import { randomUUID } from 'node:crypto'
import { expect, test, type APIRequestContext, type APIResponse } from '@playwright/test'
import { loginViaUI } from './helpers/auth'
import { apiURL, assertIsolatedWriteTarget, isWriteSuiteEnabled, testCredentials, testOrg } from './helpers/environment'
import { labels } from './helpers/locale'

type Identity = { token: string; userID: string; org: string }
type Entity = { id: string }

function required(name: string): string {
  const value = process.env[name]?.trim()
  if (!value) throw new Error(`Missing ${name} for isolated learning-journey test`)
  return value
}

async function payload<T>(response: APIResponse, status: number): Promise<T> {
  expect(response.status()).toBe(status)
  return ((await response.json()) as { data: T }).data
}

async function signIn(request: APIRequestContext, org: string, email: string, password: string): Promise<Identity> {
  const login = await request.post(`${apiURL()}/auth/login`, {
    headers: { 'X-Org-Slug': org }, data: { email, password },
  })
  const result = await payload<{ access_token: string; user: Entity }>(login, 200)
  return { token: result.access_token, userID: result.user.id, org }
}

async function tenantRequest(request: APIRequestContext, identity: Identity, method: string, path: string, data?: Record<string, unknown>, orgOverride?: string): Promise<APIResponse> {
  return request.fetch(`${apiURL()}${path}`, {
    method,
    headers: { Authorization: `Bearer ${identity.token}`, 'X-Org-Slug': orgOverride ?? identity.org },
    ...(data ? { data } : {}),
  })
}

test.skip(!isWriteSuiteEnabled(), 'Run only the guarded isolated write suite with E2E_ALLOW_WRITES=true.')

test('teacher → student → teacher → student learning journey @write', async ({ page, request }) => {
  // This is intentionally first. Login, fixture creation and browser writes
  // are prohibited unless the running Compose stack proves its identity.
  assertIsolatedWriteTarget()

  const org = testOrg()
  const teacherCredentials = testCredentials('teacher')
  const adminCredentials = testCredentials('admin')
  const studentCredentials = testCredentials('student')
  const otherOrg = required('E2E_OTHER_ORG')
  const marker = randomUUID().slice(0, 8)
  const courseTitle = `E2E learning journey ${marker}`
  const moduleTitle = `Module ${marker}`
  const lessonTitle = `Lesson ${marker}`
  const nextLessonTitle = `Next lesson ${marker}`
  const assignmentTitle = `Assignment ${marker}`
  const comparisonCourseTitle = `E2E progress comparison ${marker}`
  const comparisonModuleTitle = `Comparison module ${marker}`
  const submissionText = `Submission ${marker}`
  const fileText = `Private submission file ${marker}`
  const feedback = `Reviewed ${marker}`

  const teacher = await signIn(request, org, teacherCredentials.email, teacherCredentials.password)
  const admin = await signIn(request, org, adminCredentials.email, adminCredentials.password)
  const student = await signIn(request, org, studentCredentials.email, studentCredentials.password)
  const otherStudent = await signIn(request, org, required('E2E_OTHER_STUDENT_EMAIL'), required('E2E_OTHER_STUDENT_PASSWORD'))
  const otherTeacher = await signIn(request, org, required('E2E_OTHER_TEACHER_EMAIL'), required('E2E_OTHER_TEACHER_PASSWORD'))
  const foreignStudent = await signIn(request, otherOrg, required('E2E_OTHER_ORG_STUDENT_EMAIL'), required('E2E_OTHER_ORG_STUDENT_PASSWORD'))

  // Teacher creates the course through the actual browser form.
  await loginViaUI(page, org, teacherCredentials.email, teacherCredentials.password)
  await page.goto(`/${org}/teacher/courses/new`)
  await page.locator('#title').fill(courseTitle)
  await page.locator('#description').fill(`Isolated course for ${marker}`)
  await page.locator('button[type="submit"]').click()
  await page.waitForURL(new RegExp(`/${org}/teacher/courses/[0-9a-f-]{36}$`))
  const courseID = new URL(page.url()).pathname.split('/').pop() ?? ''
  expect(courseID).toMatch(/^[0-9a-f-]{36}$/)
  expect((await payload<{ title: string; status: string }>(await tenantRequest(request, teacher, 'GET', `/courses/${courseID}`), 200)).title).toBe(courseTitle)

  // Teacher authors the module, lesson and assignment; read-after-write
  // assertions prove persistence rather than trusting a toast notification.
  const module = await payload<Entity>(await tenantRequest(request, teacher, 'POST', `/courses/${courseID}/modules`, { title: moduleTitle, position: 0 }), 201)
  const lesson = await payload<Entity>(await tenantRequest(request, teacher, 'POST', `/modules/${module.id}/lessons`, { title: lessonTitle, type: 'text', content: `Study material ${marker}`, position: 0 }), 201)
  const nextLesson = await payload<Entity>(await tenantRequest(request, teacher, 'POST', `/modules/${module.id}/lessons`, { title: nextLessonTitle, type: 'text', content: `Next study material ${marker}`, position: 1 }), 201)
  const assignment = await payload<Entity>(await tenantRequest(request, teacher, 'POST', `/courses/${courseID}/modules/${module.id}/assignments`, { title: assignmentTitle, description: `Answer ${marker}`, max_points: 100 }), 201)
  expect((await payload<{ title: string }>(await tenantRequest(request, teacher, 'GET', `/modules/${module.id}/lessons/${lesson.id}`), 200)).title).toBe(lessonTitle)
  expect((await payload<{ title: string }>(await tenantRequest(request, teacher, 'GET', `/assignments/${assignment.id}`), 200)).title).toBe(assignmentTitle)

  // This second course is intentionally empty from the student's perspective.
  // It proves the course-scoped progress view cannot leak A's completed lesson
  // into B, and that B's denominator includes only its own two lessons.
  const comparisonCourse = await payload<Entity>(await tenantRequest(request, teacher, 'POST', '/courses', {
    title: comparisonCourseTitle, description: `Progress comparison ${marker}`,
  }), 201)
  const comparisonModule = await payload<Entity>(await tenantRequest(request, teacher, 'POST', `/courses/${comparisonCourse.id}/modules`, {
    title: comparisonModuleTitle, position: 0,
  }), 201)
  await payload<Entity>(await tenantRequest(request, teacher, 'POST', `/modules/${comparisonModule.id}/lessons`, {
    title: `Comparison lesson one ${marker}`, type: 'text', content: 'First comparison lesson', position: 0,
  }), 201)
  await payload<Entity>(await tenantRequest(request, teacher, 'POST', `/modules/${comparisonModule.id}/lessons`, {
    title: `Comparison lesson two ${marker}`, type: 'text', content: 'Second comparison lesson', position: 1,
  }), 201)

  // A second teacher, another student, and another tenant have no access.
  expect((await tenantRequest(request, otherTeacher, 'GET', `/courses/${courseID}`)).status()).toBe(403)
  expect((await tenantRequest(request, otherStudent, 'GET', `/courses/${courseID}`)).status()).toBe(403)
  expect((await tenantRequest(request, foreignStudent, 'GET', `/courses/${courseID}`, undefined, org)).status()).toBe(403)
  expect((await tenantRequest(request, student, 'POST', '/courses', { title: 'forbidden' })).status()).toBe(403)

  // The admin enrolls exactly one student and publishes this new course.
  const group = await payload<Entity>(await tenantRequest(request, admin, 'POST', `/courses/${courseID}/groups`, { name: `Group ${marker}` }), 201)
  await payload<Entity>(await tenantRequest(request, admin, 'POST', `/groups/${group.id}/members`, { student_id: student.userID }), 201)
  expect((await tenantRequest(request, student, 'GET', `/courses/${courseID}`)).status()).toBe(403)
  await payload<{ status: string }>(await tenantRequest(request, admin, 'POST', `/courses/${courseID}/publish`), 200)
  const comparisonGroup = await payload<Entity>(await tenantRequest(request, admin, 'POST', `/courses/${comparisonCourse.id}/groups`, { name: `Comparison group ${marker}` }), 201)
  await payload<Entity>(await tenantRequest(request, admin, 'POST', `/groups/${comparisonGroup.id}/members`, { student_id: student.userID }), 201)
  await payload<{ status: string }>(await tenantRequest(request, admin, 'POST', `/courses/${comparisonCourse.id}/publish`), 200)
  expect((await tenantRequest(request, student, 'GET', `/courses/${courseID}`)).status()).toBe(200)
  expect((await tenantRequest(request, otherStudent, 'GET', `/assignments/${assignment.id}`)).status()).toBe(403)

  // The real student form must preserve the draft after a failed upload and
  // persist it after retry. API calls below only verify authorization/state.
  await loginViaUI(page, org, studentCredentials.email, studentCredentials.password)
  await page.goto(`/${org}/student/courses/${courseID}/lessons/${lesson.id}`)
  const completeLesson = page.waitForResponse((response) => (
    response.url() === `${apiURL()}/progress/lessons/${lesson.id}/complete`
    && response.request().method() === 'POST'
    && response.status() === 200
  ))
  await page.getByRole('button', { name: 'Mark complete' }).click()
  await completeLesson
  await expect(page.getByRole('button', { name: 'Completed' })).toBeDisabled()

  await page.goto(`/${org}/student/courses/${courseID}`)
  await expect(page.getByText('1 lesson completed')).toBeVisible()
  await expect(page.getByText('1/2', { exact: true })).toBeVisible()
  await page.goto(`/${org}/student/courses/${comparisonCourse.id}`)
  await expect(page.getByText('0 lessons completed')).toBeVisible()
  await expect(page.getByText('0/2', { exact: true })).toBeVisible()

  await page.goto(`/${org}/student/courses/${courseID}/assignments/${assignment.id}`)
  await expect(page.getByRole('heading', { name: assignmentTitle })).toBeVisible()
  await page.locator('#text_content').fill(submissionText)
  await page.locator('#submission-file').setInputFiles({ name: `submission-${marker}.txt`, mimeType: 'text/plain', buffer: Buffer.from(fileText) })
  await expect(page.getByText(`submission-${marker}.txt`)).toBeVisible()
  const submitURL = `${apiURL()}/assignments/${assignment.id}/submit`
  await page.route(submitURL, (route) => route.fulfill({
    status: 500,
    headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': new URL(page.url()).origin },
    body: JSON.stringify({ error: { message: 'isolated simulated submission failure' } }),
  }))
  const failedSubmit = page.waitForResponse((response) => response.url() === submitURL && response.status() === 500)
  await page.getByRole('button', { name: 'Submit', exact: true }).click()
  await failedSubmit
  await expect(page.locator('#text_content')).toHaveValue(submissionText)
  await expect(page.getByRole('alert').filter({ hasText: 'Could not submit your work' })).toBeVisible()
  await page.unroute(submitURL)
  const successfulSubmit = page.waitForResponse((response) => response.url() === submitURL && response.status() === 200)
  await page.getByRole('button', { name: 'Try submitting again' }).click()
  await successfulSubmit
  await expect(page.getByText(submissionText)).toBeVisible()
  const submitted = await payload<{ id: string; text_content: string; file_path: string; score: number | null }>(
    await tenantRequest(request, student, 'GET', `/assignments/${assignment.id}/my-submission`), 200,
  )
  expect(submitted.text_content).toBe(submissionText)
  expect(submitted.score).toBeNull()
  expect(submitted.file_path).toMatch(/^uploads\//)
  const ownFile = await tenantRequest(request, student, 'GET', `/files/submissions/${submitted.id}`)
  expect(ownFile.status()).toBe(200)
  expect(ownFile.headers()['cache-control']).toContain('private, no-store')
  expect(await ownFile.text()).toBe(fileText)
  expect((await tenantRequest(request, teacher, 'GET', `/files/submissions/${submitted.id}`)).status()).toBe(200)
  expect((await tenantRequest(request, otherStudent, 'GET', `/files/submissions/${submitted.id}`)).status()).toBe(403)
  expect((await tenantRequest(request, foreignStudent, 'GET', `/files/submissions/${submitted.id}`, undefined, org)).status()).toBe(403)
  expect((await request.get(`${apiURL()}/files/submissions/${submitted.id}`, { headers: { 'X-Org-Slug': org } })).status()).toBe(401)
  expect((await request.get(`${apiURL().replace(/\/api\/v1\/?$/, '')}/${submitted.file_path}`)).status()).toBe(404)
  expect((await payload<{ id: string }>(await tenantRequest(request, student, 'GET', `/assignments/${assignment.id}/my-submission`), 200)).id).toBe(submitted.id)
  expect((await tenantRequest(request, otherStudent, 'GET', `/assignments/${assignment.id}/my-submission`)).status()).toBe(403)
  expect((await tenantRequest(request, student, 'PUT', `/submissions/${submitted.id}/grade`, { score: 0 })).status()).toBe(403)
  expect((await tenantRequest(request, otherTeacher, 'PUT', `/submissions/${submitted.id}/grade`, { score: 0 })).status()).toBe(403)

  // Validation must reject impossible scores without mutating the submission.
  expect((await tenantRequest(request, teacher, 'PUT', `/submissions/${submitted.id}/grade`, { score: 101 })).status()).toBe(422)
  expect((await payload<{ score: number | null }>(await tenantRequest(request, student, 'GET', `/assignments/${assignment.id}/my-submission`), 200)).score).toBeNull()

  // The teacher grades in the actual SubmissionGrader, including its failed
  // save/retry path. Only forbidden and invalid grades above use direct API.
  await loginViaUI(page, org, teacherCredentials.email, teacherCredentials.password)
  await page.goto(`/${org}/teacher/courses/${courseID}/assignments/${assignment.id}`)
  await expect(page.getByText(submissionText)).not.toBeVisible()
  await page.getByRole('button', { name: 'Review', exact: true }).click()
  await expect(page.getByText(submissionText)).toBeVisible()
  const scoreInput = page.locator(`#submission-score-${submitted.id}`)
  const feedbackInput = page.locator(`#submission-feedback-${submitted.id}`)
  const gradeURL = `${apiURL()}/submissions/${submitted.id}/grade`
  let gradeRequests = 0
  const observeGradeRequest = (browserRequest: import('@playwright/test').Request) => {
    if (browserRequest.url() === gradeURL && browserRequest.method() === 'PUT') gradeRequests += 1
  }
  page.on('request', observeGradeRequest)
  await scoreInput.fill('1.5')
  await page.getByRole('button', { name: 'Save Grade' }).click()
  await expect(page.getByRole('alert').filter({ hasText: 'Score must be a whole number.' })).toBeVisible()
  await scoreInput.fill('1e2')
  await page.getByRole('button', { name: 'Save Grade' }).click()
  await expect(page.getByRole('alert').filter({ hasText: 'Score must be a whole number.' })).toBeVisible()
  expect(gradeRequests).toBe(0)
  page.off('request', observeGradeRequest)

  // The valid boundaries are sent unchanged through the real form. Keep the
  // final persisted score at zero for the downstream recommendation scenario.
  await scoreInput.fill('100')
  await feedbackInput.fill(`Maximum ${marker}`)
  const maximumGrade = page.waitForResponse((response) => response.url() === gradeURL && response.request().method() === 'PUT' && response.status() === 200)
  await page.getByRole('button', { name: 'Save Grade' }).click()
  await maximumGrade
  expect((await payload<{ score: number }>(await tenantRequest(request, student, 'GET', `/assignments/${assignment.id}/my-submission`), 200)).score).toBe(100)

  await scoreInput.fill('0')
  await feedbackInput.fill(feedback)
  await page.route(gradeURL, (route) => route.fulfill({
    status: 500,
    headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': new URL(page.url()).origin },
    body: JSON.stringify({ error: { message: 'isolated simulated grading failure' } }),
  }))
  const failedGrade = page.waitForResponse((response) => response.url() === gradeURL && response.status() === 500)
  await page.getByRole('button', { name: 'Update Grade' }).click()
  await failedGrade
  await expect(scoreInput).toHaveValue('0')
  await expect(feedbackInput).toHaveValue(feedback)
  await expect(page.getByRole('alert').filter({ hasText: 'Could not save this grade' })).toBeVisible()
  await page.unroute(gradeURL)
  const successfulGrade = page.waitForResponse((response) => response.url() === gradeURL && response.status() === 200)
  await page.getByRole('button', { name: 'Try saving again' }).click()
  await successfulGrade
  const graded = await payload<{ score: number; feedback: string }>(await tenantRequest(request, student, 'GET', `/assignments/${assignment.id}/my-submission`), 200)
  expect(graded.score).toBe(0)
  expect(graded.feedback).toBe(feedback)
  const grades = await payload<{ items: Array<{ item_id: string; status: string; score: number; feedback: string }>; final_grade: number | null }>(await tenantRequest(request, student, 'GET', `/courses/${courseID}/my-grades`), 200)
  expect(grades.items).toEqual(expect.arrayContaining([expect.objectContaining({ item_id: assignment.id, status: 'graded', score: 0, feedback })]))
  expect(grades.final_grade ?? null).toBeNull() // final result requires publication; omitted is also protected

  // Finally assert the student-facing browser renders the persisted mark.
  await loginViaUI(page, org, studentCredentials.email, studentCredentials.password)
  await page.goto(`/${org}/student/courses/${courseID}`)
  await page.getByRole('tab', { name: labels.grades }).click()
  await expect(page.getByText(assignmentTitle)).toBeVisible()
  await expect(page.getByText('0 / 100')).toBeVisible()
  await expect(page.getByText(feedback)).toBeVisible()

  // A real visible low source mark should identify available course material
  // on the student dashboard. This is a course-level heuristic, not AI.
  await page.goto(`/${org}/student`)
  await expect(page.getByText(assignmentTitle)).toBeVisible()
  await expect(page.getByRole('link', { name: new RegExp(nextLessonTitle) })).toHaveAttribute('href', `/${org}/student/courses/${courseID}/lessons/${nextLesson.id}`)
})

test('teacher course form keeps data after a server error and retries @write', async ({ page, request }) => {
  assertIsolatedWriteTarget()
  const org = testOrg()
  const teacherCredentials = testCredentials('teacher')
  const title = `E2E retry ${randomUUID().slice(0, 8)}`
  const description = 'Preserve this description across a failed request'
  await loginViaUI(page, org, teacherCredentials.email, teacherCredentials.password)
  await page.goto(`/${org}/teacher/courses/new`)
  await page.locator('#title').fill(title)
  await page.locator('#description').fill(description)

  await page.route(`${apiURL()}/courses`, async (route) => {
    if (route.request().method() !== 'POST') return route.continue()
    return route.fulfill({
      status: 500,
      headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': new URL(page.url()).origin },
      body: JSON.stringify({ error: { message: 'isolated simulated failure' } }),
    })
  })
  const failed = page.waitForResponse((response) => response.url() === `${apiURL()}/courses` && response.request().method() === 'POST' && response.status() === 500)
  await page.locator('button[type="submit"]').click()
  await failed
  await expect(page.locator('#title')).toHaveValue(title)
  await expect(page.locator('#description')).toHaveValue(description)
  await expect(page).toHaveURL(new RegExp(`/${org}/teacher/courses/new$`))

  await page.unroute(`${apiURL()}/courses`)
  await page.locator('button[type="submit"]').click()
  await page.waitForURL(new RegExp(`/${org}/teacher/courses/[0-9a-f-]{36}$`))
  const courseID = new URL(page.url()).pathname.split('/').pop() ?? ''
  const teacher = await signIn(request, org, teacherCredentials.email, teacherCredentials.password)
  const persisted = await payload<{ title: string; description: string }>(await tenantRequest(request, teacher, 'GET', `/courses/${courseID}`), 200)
  expect(persisted).toEqual(expect.objectContaining({ title, description }))
})

test('teacher authors a lesson and assignment on a narrow screen @write', async ({ page, request }) => {
  assertIsolatedWriteTarget()
  const org = testOrg()
  const credentials = testCredentials('teacher')
  const marker = randomUUID().slice(0, 8)
  const lessonTitle = `Browser lesson ${marker}`
  const assignmentTitle = `Browser assignment ${marker}`
  await page.setViewportSize({ width: 390, height: 844 })
  await loginViaUI(page, org, credentials.email, credentials.password)
  await page.goto(`/${org}/teacher/courses/new`)
  await page.locator('#title').fill(`Browser authoring ${marker}`)
  await page.locator('#description').fill(`Authoring course ${marker}`)
  await page.locator('button[type="submit"]').click()
  await page.waitForURL(new RegExp(`/${org}/teacher/courses/[0-9a-f-]{36}$`))
  const courseID = new URL(page.url()).pathname.split('/').pop() ?? ''
  const teacher = await signIn(request, org, credentials.email, credentials.password)

  await page.getByRole('tab', { name: labels.contentBuilder }).click()
  const addModule = page.getByRole('button', { name: labels.addModule })
  await expect(addModule).toBeVisible()
  const createdModule = page.waitForResponse((response) => response.url() === `${apiURL()}/courses/${courseID}/modules` && response.request().method() === 'POST' && response.status() === 201)
  await addModule.click()
  const moduleID = ((await (await createdModule).json()) as { data: Entity }).data.id
  const moduleRow = page.locator('.rounded-lg.border.bg-white.shadow-sm').filter({ hasText: 'New Module' }).first()
  await expect(moduleRow).toBeVisible()

  const createdLesson = page.waitForResponse((response) => response.url() === `${apiURL()}/modules/${moduleID}/lessons` && response.request().method() === 'POST' && response.status() === 201)
  await moduleRow.getByTitle('Add lesson').click()
  const lessonID = ((await (await createdLesson).json()) as { data: Entity }).data.id
  await moduleRow.getByRole('button', { name: 'Edit New Lesson' }).click()
  const lessonDialog = page.getByRole('dialog', { name: 'Edit Lesson' })
  await lessonDialog.getByLabel('Title').fill(lessonTitle)
  await lessonDialog.getByLabel('Content').fill(`Browser authored material ${marker}`)
  await lessonDialog.getByRole('button', { name: 'Save Changes' }).click()
  await expect(lessonDialog).toBeHidden()
  expect((await payload<{ title: string; content: string }>(await tenantRequest(request, teacher, 'GET', `/modules/${moduleID}/lessons/${lessonID}`), 200))).toEqual(expect.objectContaining({ title: lessonTitle, content: `Browser authored material ${marker}` }))

  const createdAssignment = page.waitForResponse((response) => response.url() === `${apiURL()}/courses/${courseID}/modules/${moduleID}/assignments` && response.request().method() === 'POST' && response.status() === 201)
  await moduleRow.getByTitle('Add assignment').click()
  const assignmentID = ((await (await createdAssignment).json()) as { data: Entity }).data.id
  await moduleRow.getByRole('button', { name: 'Edit New Assignment' }).click()
  const assignmentDialog = page.getByRole('dialog', { name: 'Edit Assignment' })
  await assignmentDialog.getByLabel('Title').fill(assignmentTitle)
  await assignmentDialog.getByLabel('Statement / Description').fill(`Browser authored task ${marker}`)
  await assignmentDialog.getByRole('button', { name: 'Save Changes' }).click()
  await expect(assignmentDialog).toBeHidden()
  expect((await payload<{ title: string; description: string }>(await tenantRequest(request, teacher, 'GET', `/assignments/${assignmentID}`), 200))).toEqual(expect.objectContaining({ title: assignmentTitle, description: `Browser authored task ${marker}` }))
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(391)
})
