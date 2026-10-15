import { expect, test } from '@playwright/test'
import { AxiosError, type AxiosAdapter, type AxiosResponse } from 'axios'
import { QueryClient } from '@tanstack/react-query'
import { apiClient } from '../../src/lib/api/client'
import { getGoogleOAuthStartURL } from '../../src/lib/api/auth'
import { registerSessionQueryClient } from '../../src/lib/query/session-cache'
import { useAuthStore } from '../../src/lib/stores/auth.store'
import { useReAuthStore } from '../../src/lib/stores/reauth.store'

const makeSession = (id: string, orgSlug = 'school') => ({
  token: `token-${id}`,
  user: { id, email: `${id}@example.test`, first_name: id, last_name: '', role: 'student' as const, org_id: orgSlug },
  role: 'student' as const,
  orgSlug,
  expiresAt: new Date(Date.now() + 60_000).toISOString(),
  authProvider: 'password' as const,
})

function response(config: Parameters<AxiosAdapter>[0], data: unknown, status = 200): AxiosResponse {
  return { config, data, status, statusText: status === 200 ? 'OK' : 'Unauthorized', headers: {} }
}

test.beforeEach(() => { useAuthStore.getState().clearSession(); useAuthStore.getState().setSession(makeSession('alice')) })
test.afterEach(() => { useReAuthStore.getState().reject(); useAuthStore.getState().clearSession() })

test('two simultaneous 401s share one re-auth and keep their own responses', async () => {
  const previousAdapter = apiClient.defaults.adapter
  let initialCalls = 0
  let retryCalls = 0
  apiClient.defaults.adapter = async (config) => {
    if (config.headers.Authorization === 'Bearer token-alice') {
      initialCalls++
      throw new AxiosError('unauthorized', 'ERR_BAD_REQUEST', config, undefined, response(config, {}, 401))
    }
    retryCalls++
    return response(config, { data: { endpoint: config.url } })
  }
  try {
    const first = apiClient.get('/one')
    const second = apiClient.get('/two')
    await expect.poll(() => useReAuthStore.getState().isOpen).toBe(true)
    expect(initialCalls).toBe(2)
    useAuthStore.getState().setSession({ ...makeSession('alice'), token: 'fresh-token' })
    useReAuthStore.getState().resolve('fresh-token')
    const [one, two] = await Promise.all([first, second])
    expect(one.data.endpoint).toBe('/one')
    expect(two.data.endpoint).toBe('/two')
    expect(retryCalls).toBe(2)
  } finally { apiClient.defaults.adapter = previousAdapter }
})

test('cancel rejects every waiting request and repeated 401 does not loop', async () => {
  const previousAdapter = apiClient.defaults.adapter
  let calls = 0
  apiClient.defaults.adapter = async (config) => {
    calls++
    throw new AxiosError('unauthorized', 'ERR_BAD_REQUEST', config, undefined, response(config, {}, 401))
  }
  try {
    const first = apiClient.get('/one').catch((error: Error) => error)
    const second = apiClient.get('/two').catch((error: Error) => error)
    await expect.poll(() => useReAuthStore.getState().isOpen).toBe(true)
    useReAuthStore.getState().reject()
    expect(await first).toBeInstanceOf(Error)
    expect(await second).toBeInstanceOf(Error)
    expect(calls).toBe(2)

    const repeated = apiClient.get('/again').catch((error: Error) => error)
    await expect.poll(() => useReAuthStore.getState().isOpen).toBe(true)
    useReAuthStore.getState().resolve('fresh-token')
    expect(await repeated).toBeInstanceOf(Error)
    expect(calls).toBe(4)
  } finally { apiClient.defaults.adapter = previousAdapter }
})

test('logout discards an in-flight response and a new user gets no private cached data', async () => {
  const previousAdapter = apiClient.defaults.adapter
  const queryClient = new QueryClient()
  const unregister = registerSessionQueryClient(queryClient)
  queryClient.setQueryData(['grades'], [{ score: 90 }])
  queryClient.setQueryData(['notifications'], [{ message: 'private' }])
  let finish!: (value: AxiosResponse) => void
  apiClient.defaults.adapter = (config) => new Promise((resolve) => { finish = (value) => resolve({ ...value, config }) })
  try {
    const pending = apiClient.get('/grades').catch((error: Error) => error)
    await expect.poll(() => typeof finish).toBe('function')
    useAuthStore.getState().clearSession()
    useAuthStore.getState().setSession(makeSession('bob'))
    expect(queryClient.getQueryData(['grades'])).toBeUndefined()
    expect(queryClient.getQueryData(['notifications'])).toBeUndefined()
    finish(response({} as Parameters<AxiosAdapter>[0], { data: [{ score: 90 }] }))
    expect(await pending).toBeInstanceOf(Error)
  } finally { unregister(); apiClient.defaults.adapter = previousAdapter }
})

test('Google re-auth URL retains the organization context', () => {
  expect(getGoogleOAuthStartURL('school-with-space')).toContain('org=school-with-space')
  expect(getGoogleOAuthStartURL('other school')).toContain('org=other%20school')
})
