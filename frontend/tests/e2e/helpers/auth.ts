import { Page } from '@playwright/test'
import { apiURL } from './environment'
import { clearClientState } from './locale'

async function beginCleanSession(page: Page): Promise<void> {
  await page.context().clearCookies()
  await clearClientState(page)
}

/**
 * Log in via the UI login page.
 * Waits for the redirect to the role-based dashboard.
 */
export async function loginViaUI(
  page: Page,
  orgSlug: string,
  email: string,
  password: string,
): Promise<void> {
  await beginCleanSession(page)
  await page.goto(orgSlug ? `/${orgSlug}/login` : '/login')
  await page.waitForLoadState('networkidle')

  await page.locator('#email').fill(email)
  await page.locator('#password').fill(password)
  await page.locator('button[type="submit"]').click()

  // Wait for navigation away from the login page
  await page.waitForURL((url) => !url.pathname.endsWith('/login'), { timeout: 15000 })
}

/**
 * Faster alternative: call the backend API directly and inject
 * the session into localStorage so we skip the login UI.
 */
export async function loginViaAPI(
  page: Page,
  orgSlug: string,
  email: string,
  password: string,
): Promise<void> {
  await beginCleanSession(page)
  // First, navigate somewhere so we can set localStorage
  await page.goto(orgSlug ? `/${orgSlug}/login` : '/login')
  await page.waitForLoadState('domcontentloaded')

  const result = await page.evaluate(
    async ({ apiUrl, orgSlug, email, password }) => {
      const res = await fetch(`${apiUrl}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Org-Slug': orgSlug },
        body: JSON.stringify({ email, password }),
      })
      if (!res.ok) throw new Error(`Login failed: ${res.status}`)
      return res.json()
    },
    { apiUrl: apiURL(), orgSlug, email, password },
  )

  const raw = result.data
  const payload = JSON.parse(atob(raw.access_token.split('.')[1]))

  const nameParts: string[] = raw.user.name.split(' ')
  const session = {
    state: {
      token: raw.access_token,
      expiresAt: raw.expires_at,
      role: payload.role,
      orgSlug,
      user: {
        id: raw.user.id,
        email: raw.user.email,
        first_name: nameParts[0] ?? '',
        last_name: nameParts.slice(1).join(' ') ?? '',
        role: payload.role,
        org_id: payload.org_id ?? '',
      },
    },
    version: 0,
  }

  await page.evaluate((s) => {
    localStorage.setItem('mentra-auth', JSON.stringify(s))
    document.cookie = `mentra-role=${s.state.role}; path=/; max-age=${60 * 60 * 24}`
    document.cookie = `mentra-org=${s.state.orgSlug}; path=/; max-age=${60 * 60 * 24}`
  }, session)
}

// Helper to decode base64 in browser context
function atob(str: string): string {
  return Buffer.from(str, 'base64').toString('utf8')
}
