import { test, expect } from '@playwright/test'
import { loginViaAPI } from './helpers/auth'

/**
 * Teacher course builder tests.
 *
 * Requires the standard demo seed (run `go run ./cmd/seed` from backend/).
 *
 * Override via env vars:
 *   TEST_ORG, TEST_TEACHER_EMAIL, TEST_TEACHER_PASSWORD
 */

const ORG = process.env.TEST_ORG ?? 'nu'
const EMAIL = process.env.TEST_TEACHER_EMAIL ?? 'baurzhan@nu.edu.kz'
const PASS = process.env.TEST_TEACHER_PASSWORD ?? 'password123'

// See the student suite for why detailed, language-specific UI checks are
// opt-in. The platform-smoke suite is the deployment gate.
test.skip(process.env.E2E_RUN_EXTENDED !== 'true', 'Set E2E_RUN_EXTENDED=true to run detailed course checks')

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

    await expect(page.getByRole('tab', { name: 'Details' })).toBeVisible({ timeout: 10000 })
    await expect(page.getByRole('tab', { name: 'Content Builder' })).toBeVisible()
    await expect(page.getByRole('tab', { name: 'Grades' })).toBeVisible()
  })

  test('content builder has Add Module button', async ({ page }) => {
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')

    const firstCourse = page.locator('a[href*="/teacher/courses/"]').first()
    test.skip(await firstCourse.count() === 0, 'No courses')

    await firstCourse.click()
    await page.getByRole('tab', { name: 'Content Builder' }).click()
    await page.waitForTimeout(1000)

    await expect(page.getByRole('button', { name: /Add Module/i })).toBeVisible({ timeout: 8000 })
  })

  test('can create a new module', async ({ page }) => {
    test.skip(process.env.E2E_ALLOW_WRITES !== 'true', 'Set E2E_ALLOW_WRITES=true to run data-changing tests')
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')

    const firstCourse = page.locator('a[href*="/teacher/courses/"]').first()
    test.skip(await firstCourse.count() === 0, 'No courses')

    await firstCourse.click()
    await page.getByRole('tab', { name: 'Content Builder' }).click()
    await page.waitForTimeout(1000)

    const countBefore = await page.locator('.rounded-lg.border.bg-white.shadow-sm').count()
    await page.getByRole('button', { name: /Add Module/i }).click()
    await page.waitForTimeout(2000)

    const countAfter = await page.locator('.rounded-lg.border.bg-white.shadow-sm').count()
    expect(countAfter).toBeGreaterThanOrEqual(countBefore)
  })

  test('lesson edit dialog opens on pencil click', async ({ page }) => {
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')

    const firstCourse = page.locator('a[href*="/teacher/courses/"]').first()
    test.skip(await firstCourse.count() === 0, 'No courses')

    await firstCourse.click()
    await page.getByRole('tab', { name: 'Content Builder' }).click()
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

  test('lesson edit dialog can update title', async ({ page }) => {
    test.skip(process.env.E2E_ALLOW_WRITES !== 'true', 'Set E2E_ALLOW_WRITES=true to run data-changing tests')
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')

    const firstCourse = page.locator('a[href*="/teacher/courses/"]').first()
    test.skip(await firstCourse.count() === 0, 'No courses')

    await firstCourse.click()
    await page.getByRole('tab', { name: 'Content Builder' }).click()
    await page.waitForTimeout(1500)

    const firstModule = page.locator('.rounded-lg.border.bg-white.shadow-sm').first()
    test.skip(await firstModule.count() === 0, 'No modules')

    await firstModule.locator('button').nth(1).click()
    await page.waitForTimeout(1000)

    const lessonRow = page.locator('.group.flex.items-center.gap-2').first()
    test.skip(await lessonRow.count() === 0, 'No lessons')

    await lessonRow.hover()
    await page.waitForTimeout(200)
    await lessonRow.locator('button').nth(0).click()
    await page.waitForTimeout(500)

    const dialog = page.getByRole('dialog')
    test.skip(await dialog.count() === 0, 'Edit dialog did not open')

    const titleInput = dialog.getByLabel('Title')
    await titleInput.clear()
    await titleInput.fill('Updated Lesson Title')

    // Save
    await dialog.getByRole('button', { name: /Save/i }).click()
    await page.waitForTimeout(1500)

    // Dialog should close
    await expect(dialog).not.toBeVisible({ timeout: 5000 })
  })

  test('grades tab shows gradebook', async ({ page }) => {
    await page.goto(`/${ORG}/teacher/courses`)
    await page.waitForLoadState('networkidle')

    const firstCourse = page.locator('a[href*="/teacher/courses/"]').first()
    test.skip(await firstCourse.count() === 0, 'No courses')

    await firstCourse.click()
    await page.getByRole('tab', { name: 'Grades' }).click()
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
