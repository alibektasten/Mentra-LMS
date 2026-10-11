import { expect, test } from '@playwright/test'
import { loginViaUI } from './helpers/auth'

const org = process.env.TEST_ORG ?? 'nu'
const apiURL = process.env.E2E_API_URL ?? process.env.TEST_API_URL ?? 'http://localhost:8080/api/v1'
const healthURL = `${apiURL.replace(/\/api\/v1\/?$/, '')}/health`

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
    await expect(page.getByRole('button', { name: /sign in/i })).toBeVisible()
  })

  test('student can sign in and reach their workspace', async ({ page }) => {
    await loginViaUI(
      page,
      org,
      process.env.TEST_STUDENT_EMAIL ?? 'sanzhar@nu.edu.kz',
      process.env.TEST_STUDENT_PASSWORD ?? 'password123',
    )
    await expect(page).toHaveURL(new RegExp(`/${org}/student(?:/|$)`))
    await expect(page.locator('h1, h2').first()).toBeVisible({ timeout: 10_000 })
  })

  test('teacher can sign in and reach their workspace', async ({ page }) => {
    await loginViaUI(
      page,
      org,
      process.env.TEST_TEACHER_EMAIL ?? 'baurzhan@nu.edu.kz',
      process.env.TEST_TEACHER_PASSWORD ?? 'password123',
    )
    await expect(page).toHaveURL(new RegExp(`/${org}/teacher(?:/|$)`))
    await expect(page.locator('h1, h2').first()).toBeVisible({ timeout: 10_000 })
  })
})
