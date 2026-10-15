import { expect, test, type Page } from '@playwright/test'
import { loginViaUI } from './helpers/auth'
import { apiURL, assertIsolatedWriteTarget, testCredentials, testOrg } from './helpers/environment'

const firstMarker = 'Private grade notification for first student'
const secondMarker = 'Private grade notification for second student'

function notification(title: string) {
  return {
    data: [{
      id: title === firstMarker ? '11111111-1111-4111-8111-111111111111' : '22222222-2222-4222-8222-222222222222',
      type: 'grade',
      title,
      body: 'Private result',
      course_id: '33333333-3333-4333-8333-333333333333',
      course_title: 'Private course',
      created_at: new Date().toISOString(),
    }],
  }
}

async function storedSession(page: Page): Promise<{ token?: string; email?: string } | null> {
  return page.evaluate(() => {
    const raw = localStorage.getItem('mentra-auth')
    if (!raw) return null
    const state = JSON.parse(raw).state
    return { token: state?.token, email: state?.user?.email }
  })
}

test('real logout isolates delayed private response from the next student', async ({ page }) => {
  assertIsolatedWriteTarget()
  const org = testOrg()
  const first = testCredentials('student')
  const second = {
    email: process.env.E2E_OTHER_STUDENT_EMAIL?.trim(),
    password: process.env.E2E_OTHER_STUDENT_PASSWORD?.trim(),
  }
  expect(second.email, 'E2E_OTHER_STUDENT_EMAIL is required').toBeTruthy()
  expect(second.password, 'E2E_OTHER_STUDENT_PASSWORD is required').toBeTruthy()
  expect(second.email).not.toBe(first.email)

  let firstToken: string | undefined
  let delayFirstResponse = false
  let releaseFirstResponse!: () => void
  let firstResponseStarted!: () => void
  const firstResponseGate = new Promise<void>((resolve) => { releaseFirstResponse = resolve })
  const firstRequestSeen = new Promise<void>((resolve) => { firstResponseStarted = resolve })
  const secondRequestTokens: string[] = []

  await page.route(`${apiURL()}/notifications`, async (route) => {
    const authorization = route.request().headers().authorization ?? ''
    if (!firstToken || authorization === `Bearer ${firstToken}`) {
      if (delayFirstResponse) {
        firstResponseStarted()
        await firstResponseGate
      }
      try { await route.fulfill({ json: notification(firstMarker) }) } catch { /* Logout can abort this request. */ }
      return
    }
    secondRequestTokens.push(authorization)
    await route.fulfill({ json: notification(secondMarker) })
  })

  await loginViaUI(page, org, first.email, first.password)
  const firstSession = await storedSession(page)
  firstToken = firstSession?.token
  expect(firstToken).toBeTruthy()
  await page.getByRole('button', { name: 'Notifications' }).click()
  await expect(page.locator('#notification-panel')).toContainText(firstMarker)
  await page.getByRole('button', { name: 'Notifications' }).click()

  delayFirstResponse = true
  await page.reload({ waitUntil: 'domcontentloaded' })
  await firstRequestSeen
  await page.locator('header button[data-state]').last().click()
  await page.getByRole('menuitem', { name: /^(Sign out|Выйти|Шығу)$/ }).click()
  await page.waitForURL('**/login')
  await page.goto(`/${org}/login`)
  await page.locator('#email').fill(second.email!)
  await page.locator('#password').fill(second.password!)
  await page.locator('button[type="submit"]').click()
  await page.waitForURL((url) => url.pathname.startsWith(`/${org}/student`) && !url.pathname.endsWith('/login'))
  await expect.poll(async () => (await storedSession(page))?.email).toBe(second.email)
  await expect.poll(() => secondRequestTokens.length).toBeGreaterThan(0)
  releaseFirstResponse()

  await page.getByRole('button', { name: 'Notifications' }).click()
  await expect(page.locator('#notification-panel')).toContainText(secondMarker)
  await expect(page.locator('#notification-panel')).not.toContainText(firstMarker)
  expect(secondRequestTokens.every((token) => token !== `Bearer ${firstToken}`)).toBe(true)
})

test('two concurrent 401s open one re-auth dialog and Sign out clears credentials', async ({ page }) => {
  assertIsolatedWriteTarget()
  const org = testOrg()
  const student = testCredentials('student')
  await loginViaUI(page, org, student.email, student.password)
  let unauthorizedReads = 0
  for (const endpoint of ['me', 'notifications']) {
    await page.route(`${apiURL()}/${endpoint}`, async (route) => {
      unauthorizedReads++
      await route.fulfill({ status: 401, json: { error: { message: 'Expired session' } } })
    })
  }

  await page.reload({ waitUntil: 'domcontentloaded' })
  await expect.poll(() => unauthorizedReads).toBeGreaterThanOrEqual(2)
  const dialog = page.getByRole('dialog')
  await expect(dialog).toHaveCount(1)
  await expect(dialog).toContainText('Session Expired')
  await expect(dialog.locator('#reauth-email')).toHaveValue(student.email)
  await dialog.getByRole('button', { name: 'Sign out' }).click()
  await page.waitForURL('**/login')
  await expect.poll(async () => (await storedSession(page))?.token).toBeFalsy()
  expect((await page.context().cookies()).filter((cookie) => ['mentra-role', 'mentra-org'].includes(cookie.name))).toHaveLength(0)
  expect(unauthorizedReads).toBe(2)
})
