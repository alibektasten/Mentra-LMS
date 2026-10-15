import type { Page } from '@playwright/test'

/** Stable role labels in the three supported interface languages. */
export const labels = {
  signIn: /^(sign in|войти|кіру)$/i,
  content: /^(content|содержание|контент|мазмұны)$/i,
  myGrades: /^(my grades|мои оценки|бағаларым)$/i,
  deadlines: /^(deadlines|сроки|дедлайны|мерзімдер)$/i,
  details: /^(details|детали|сипаттамасы)$/i,
  contentBuilder: /^(content builder|конструктор контента|конструктор материалов|контент құрастырушы)$/i,
  grades: /^(grades|оценки|бағалар)$/i,
  addModule: /^(add module|добавить модуль|модуль қосу)$/i,
} as const

/**
 * One clean, predictable browser state per test. It avoids dependencies on a
 * prior test's language, theme, auth token, or persisted React Query cache.
 */
export async function clearClientState(page: Page): Promise<void> {
  // Clear once, on the test origin. An addInitScript runs on *every* later
  // navigation and silently erases the session created by the login form.
  await page.goto('/login')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })
}
