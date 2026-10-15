import { randomUUID } from 'node:crypto'
import { test, expect, type APIRequestContext, type APIResponse } from '@playwright/test'
import { loginViaAPI } from './helpers/auth'
import { apiURL, assertIsolatedWriteTarget, isExtendedSuiteEnabled, isWriteSuiteEnabled, testCredentials, testOrg } from './helpers/environment'
import { labels } from './helpers/locale'

/**
 * Teacher course builder tests.
 *
 * Requires the isolated stack prepared by scripts/e2e/prepare-isolated.sh.
 */

const ORG = testOrg()
const { email: EMAIL, password: PASS } = testCredentials('teacher')

type APIEntity = { id: string }
type TeacherSession = { token: string }

async function responseData<T>(response: APIResponse, expectedStatus: number): Promise<T> {
  expect(response.status()).toBe(expectedStatus)
  return ((await response.json()) as { data: T }).data
}

async function teacherSession(request: APIRequestContext): Promise<TeacherSession> {
  const login = await request.post(`${apiURL()}/auth/login`, {
    headers: { 'X-Org-Slug': ORG },
    data: { email: EMAIL, password: PASS },
  })
  const data = await responseData<{ access_token: string }>(login, 200)
  return { token: data.access_token }
}

function teacherHeaders(session: TeacherSession) {
  return { Authorization: `Bearer ${session.token}`, 'X-Org-Slug': ORG }
}

// See the student suite for why detailed, language-specific UI checks are
// opt-in. The platform-smoke suite is the deployment gate.
test.skip(!isExtendedSuiteEnabled(), 'Run the explicit extended suite with E2E_RUN_EXTENDED=true.')

test.describe('Teacher — course builder', () => {
  test.beforeEach(async ({ page }) => {
    await loginViaAPI(page, ORG, EMAIL, PASS)
  })

  test('teacher dashboard loads', async ({ page }) => {
    await page.goto(`/${ORG}/teacher`)
    await expect(page).toHaveURL(new RegExp(`/${ORG}/teacher`))
    await expect(page.locator('h1, h2').first()).toBeVisible({ timeout: 10000 })
  })

  test('teacher courses list loads', async ({ page }) => {
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')
    // Either shows courses or empty state / create button
    const hasCourses = await page.locator('a[href*="/teacher/courses/"]').count() > 0
    const hasCreateBtn = await page.getByRole('button', { name: /new|create/i }).first().isVisible().catch(() => false)
    expect(hasCourses || hasCreateBtn).toBe(true)
  })

  test('course detail page shows Details / Content Builder / Grades tabs', async ({ page }) => {
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')

    const firstCourse = page.locator('a[href*="/teacher/courses/"]').first()
    test.skip(await firstCourse.count() === 0, 'No courses — seed the database first')

    await firstCourse.click()
    await page.waitForLoadState('networkidle')

    await expect(page.getByRole('tab', { name: labels.details })).toBeVisible({ timeout: 10000 })
    await expect(page.getByRole('tab', { name: labels.contentBuilder })).toBeVisible()
    await expect(page.getByRole('tab', { name: labels.grades })).toBeVisible()
  })

  test('content builder has Add Module button', async ({ page }) => {
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')

    const firstCourse = page.locator('a[href*="/teacher/courses/"]').first()
    test.skip(await firstCourse.count() === 0, 'No courses')

    await firstCourse.click()
    await page.getByRole('tab', { name: labels.contentBuilder }).click()
    await page.waitForTimeout(1000)

    await expect(page.getByRole('button', { name: labels.addModule })).toBeVisible({ timeout: 8000 })
  })

  test('can create a new module @write', async ({ page, request }) => {
    test.skip(!isWriteSuiteEnabled(), 'Run the explicit write suite with E2E_ALLOW_WRITES=true.')
    assertIsolatedWriteTarget()
    const marker = randomUUID().slice(0, 8)
    const session = await teacherSession(request)
    const headers = teacherHeaders(session)
    const course = await responseData<APIEntity>(await request.post(`${apiURL()}/courses`, {
      headers,
      data: { title: `E2E module add ${marker}`, description: 'Isolated browser fixture' },
    }), 201)

    await page.goto(`/${ORG}/teacher/courses/${course.id}`)
    await page.getByRole('tab', { name: labels.contentBuilder }).click()
    const moduleCreated = page.waitForResponse((response) => (
      response.url() === `${apiURL()}/courses/${course.id}/modules`
      && response.request().method() === 'POST'
      && response.status() === 201
    ))
    await page.getByRole('button', { name: labels.addModule }).click()
    const moduleResponse = await moduleCreated
    expect(moduleResponse.status()).toBe(201)
    const module = ((await moduleResponse.json()) as { data: { id: string; title: string } }).data
    const modules = await responseData<Array<{ id: string }>>(await request.get(`${apiURL()}/courses/${course.id}/modules`, { headers }), 200)
    expect(modules.map((item) => item.id)).toContain(module.id)
  })

  test('lesson edit dialog opens on pencil click', async ({ page }) => {
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')

    const firstCourse = page.locator('a[href*="/teacher/courses/"]').first()
    test.skip(await firstCourse.count() === 0, 'No courses')

    await firstCourse.click()
    await page.getByRole('tab', { name: labels.contentBuilder }).click()
    await page.waitForTimeout(1500)

    // Expand first module
    const firstModule = page.locator('.rounded-lg.border.bg-white.shadow-sm').first()
    test.skip(await firstModule.count() === 0, 'No modules')

    // Click the expand/toggle button (chevron)
    await firstModule.locator('button').nth(1).click()
    await page.waitForTimeout(1000)

    // Look for a lesson row
    const lessonRow = page.locator('.group.flex.items-center.gap-2').first()
    const hasLesson = await lessonRow.count() > 0
    test.skip(!hasLesson, 'No lessons in first module')

    // Hover to reveal action buttons
    await lessonRow.hover()
    await page.waitForTimeout(200)

    // Click the pencil/edit button (first action button after the type badge)
    const editBtn = lessonRow.locator('button').filter({ hasNot: page.locator('[data-destructive]') }).last()
    await editBtn.first().click()
    await page.waitForTimeout(500)

    // Dialog should be open
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 })
    await expect(page.getByRole('dialog').getByLabel('Title')).toBeVisible()
  })

  test('lesson edit dialog persists a unique lesson title @write', async ({ page, request }) => {
    test.skip(!isWriteSuiteEnabled(), 'Run the explicit write suite with E2E_ALLOW_WRITES=true.')
    assertIsolatedWriteTarget()
    const marker = randomUUID().slice(0, 8)
    const originalTitle = `E2E lesson ${marker}`
    const updatedTitle = `E2E lesson updated ${marker}`
    const session = await teacherSession(request)
    const headers = teacherHeaders(session)

    const course = await responseData<APIEntity>(await request.post(`${apiURL()}/courses`, {
      headers,
      data: { title: `E2E lesson edit ${marker}`, description: 'Isolated browser fixture' },
    }), 201)
    const module = await responseData<APIEntity>(await request.post(`${apiURL()}/courses/${course.id}/modules`, {
      headers,
      data: { title: `E2E module ${marker}`, order: 0 },
    }), 201)
    const lesson = await responseData<APIEntity>(await request.post(`${apiURL()}/modules/${module.id}/lessons`, {
      headers,
      data: { title: originalTitle, type: 'text', content: 'Isolated lesson content', order: 0 },
    }), 201)

    await page.goto(`/${ORG}/teacher/courses/${course.id}`)
    await page.getByRole('tab', { name: labels.contentBuilder }).click()
    await page.getByRole('button', { name: `Toggle module E2E module ${marker}` }).click()
    const editLesson = page.getByRole('button', { name: `Edit ${originalTitle}` })
    await expect(editLesson).toBeVisible()
    await editLesson.click()

    const dialog = page.getByRole('dialog', { name: 'Edit Lesson' })
    await expect(dialog).toBeVisible()
    await dialog.getByLabel('Title').fill(updatedTitle)
    const saved = page.waitForResponse((response) => (
      response.url() === `${apiURL()}/modules/${module.id}/lessons/${lesson.id}`
      && response.request().method() === 'PUT'
      && response.status() === 200
    ))
    await dialog.getByRole('button', { name: 'Save Changes' }).click()
    await saved
    await expect(dialog).toBeHidden()
    await expect(page.getByRole('button', { name: `Edit ${updatedTitle}` })).toBeVisible()

    const persisted = await responseData<{ title: string }>(await request.get(`${apiURL()}/modules/${module.id}/lessons/${lesson.id}`, { headers }), 200)
    expect(persisted.title).toBe(updatedTitle)
  })

  test('grades tab shows gradebook', async ({ page }) => {
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')

    const firstCourse = page.locator('a[href*="/teacher/courses/"]').first()
    test.skip(await firstCourse.count() === 0, 'No courses')

    await firstCourse.click()
    await page.getByRole('tab', { name: labels.grades }).click()
    await page.waitForTimeout(1000)

    // Either shows assignment data or empty state
    const hasContent = await page.locator('.rounded-lg.border').count() > 0
    const hasEmpty = await page
      .getByText(/no module|no graded|yet/i)
      .first()
      .isVisible()
      .catch(() => false)
    expect(hasContent || hasEmpty).toBe(true)
  })
})
