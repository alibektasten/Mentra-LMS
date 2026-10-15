import { expect, test } from '@playwright/test'
import { loginViaUI } from './helpers/auth'
import { apiURL, testCredentials, testOrg } from './helpers/environment'

const org = testOrg()
const healthURL = `${apiURL().replace(/\/api\/v1\/?$/, '')}/health`

test.describe('Platform smoke checks', () => {
  test('backend health endpoint reports ready', async ({ request }) => {
    const response = await request.get(healthURL)
    expect(response.ok()).toBeTruthy()
    await expect(response.json()).resolves.toEqual({ status: 'ok' })
  })

  test('organization login page is usable', async ({ page }) => {
    await page.goto(`/${org}/login`)
    await expect(page.locator('#email')).toBeVisible()
    await expect(page.locator('#password')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('student can sign in and reach their workspace', async ({ page }) => {
    const credentials = testCredentials('student')
    await loginViaUI(
      page,
      org,
      credentials.email,
      credentials.password,
    )
    await expect(page).toHaveURL(new RegExp(`/${org}/student(?:/|$)`))
    await expect(page.locator('h1, h2').first()).toBeVisible({ timeout: 10_000 })
  })

  test('teacher can sign in and reach their workspace', async ({ page }) => {
    const credentials = testCredentials('teacher')
    await loginViaUI(
      page,
      org,
      credentials.email,
      credentials.password,
    )
    await expect(page).toHaveURL(new RegExp(`/${org}/teacher(?:/|$)`))
    await expect(page.locator('h1, h2').first()).toBeVisible({ timeout: 10_000 })
  })

  test('organization administrator can sign in and reach their workspace', async ({ page }) => {
    const credentials = testCredentials('admin')
    await loginViaUI(page, org, credentials.email, credentials.password)
    await expect(page).toHaveURL(new RegExp(`/${org}/admin(?:/|$)`))
  })

  test('super administrator can sign in and reach the system console', async ({ page }) => {
    const credentials = testCredentials('super_admin')
    await loginViaUI(page, '', credentials.email, credentials.password)
    await expect(page).toHaveURL(/\/super-admin\/(?:dashboard|organizations|users|invite)(?:\/|$)/)
  })
})
