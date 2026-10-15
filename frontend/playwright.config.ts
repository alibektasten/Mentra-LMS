import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: false,
  retries: 0,
  workers: 1,
  reporter: 'list',
  forbidOnly: Boolean(process.env.CI),
  outputDir: 'test-results',
  use: {
    // E2E_BASE_URL is deliberately required by helpers: no silent fallback to
    // a developer's browser session or a production deployment.
    baseURL: process.env.E2E_BASE_URL,
    locale: process.env.E2E_LOCALE ?? 'en-US',
    timezoneId: process.env.E2E_TIMEZONE ?? 'Asia/Almaty',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
