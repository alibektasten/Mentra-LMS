/**
 * Explicit contract for browser tests.
 *
 * The suite must never silently fall back to a developer's currently running
 * LMS instance.  An isolated stack is created by scripts/e2e/prepare-isolated.sh
 * and exports these values before Playwright is started.
 */
export type TestRole = 'student' | 'teacher' | 'admin' | 'super_admin'

type RoleCredentials = {
  email: string
  password: string
}

function required(name: string): string {
  const value = process.env[name]?.trim()
  if (!value) {
    throw new Error(`Missing ${name}. Run scripts/e2e/prepare-isolated.sh and source its output first.`)
  }
  return value
}

export function testOrg(): string {
  return required('E2E_TEST_ORG')
}

export function testCredentials(role: TestRole): RoleCredentials {
  const prefix = `E2E_${role.toUpperCase()}`
  return {
    email: required(`${prefix}_EMAIL`),
    password: required(`${prefix}_PASSWORD`),
  }
}

export function apiURL(): string {
  return required('E2E_API_URL')
}

export function assertIsolatedWriteTarget(): void {
  // Direct Playwright invocation must receive the same runtime proof as the
  // npm entrypoint, before a test performs any data-changing UI action.
  const { execFileSync } = require('node:child_process') as typeof import('node:child_process')
  const { resolve } = require('node:path') as typeof import('node:path')
  const script = resolve(__dirname, '../../../../scripts/e2e/verify-write-target.mjs')
  execFileSync(process.execPath, [script], { stdio: 'inherit', env: process.env })
}

export function isExtendedSuiteEnabled(): boolean {
  return process.env.E2E_RUN_EXTENDED === 'true'
}

export function isWriteSuiteEnabled(): boolean {
  return process.env.E2E_ALLOW_WRITES === 'true'
}
