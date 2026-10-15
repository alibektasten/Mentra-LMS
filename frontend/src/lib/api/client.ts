import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/lib/stores/auth.store'
import { useReAuthStore } from '@/lib/stores/reauth.store'

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1'

// Base host for static file serving (strips /api/v1 suffix)
export const UPLOAD_BASE_URL = API_URL.replace(/\/api\/v\d+$/, '')

export const apiClient = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' },
})

type RetriableRequest = InternalAxiosRequestConfig & { _retry?: boolean; _skipReauth?: boolean; _sessionGeneration?: number }

let activeReauthentication: Promise<string> | null = null

function requestReauthentication(): Promise<string> {
  if (activeReauthentication) return activeReauthentication

  const { user } = useAuthStore.getState()
  if (!user?.email) return Promise.reject(new Error('Session expired. Sign in again.'))

  activeReauthentication = new Promise<string>((resolve, reject) => {
    useReAuthStore.getState().open(user.email, {
      onSuccess: resolve,
      onCancel: () => reject(new Error('Re-authentication cancelled')),
    })
  }).finally(() => {
    activeReauthentication = null
  })
	return activeReauthentication
}

// Request interceptor: attach auth + org slug
apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const { token, orgSlug, sessionGeneration } = useAuthStore.getState()
  const request = config as RetriableRequest
  request._sessionGeneration ??= sessionGeneration
  if (token) {
    config.headers['Authorization'] = `Bearer ${token}`
  }
  if (orgSlug) {
    config.headers['X-Org-Slug'] = orgSlug
  }
  return config
})

// Response interceptor: unwrap envelope + handle 401
apiClient.interceptors.response.use(
  (response) => {
    const request = response.config as RetriableRequest
    if (request._sessionGeneration !== undefined && request._sessionGeneration !== useAuthStore.getState().sessionGeneration) {
      return Promise.reject(new Error('Session changed while the request was running'))
    }
    // Unwrap { data: ... } envelope (handles both object and array payloads)
    if (response.data && 'data' in response.data) {
      // Keep paginated responses intact so callers can access .meta
      if ('meta' in response.data) {
        return response
      }
      response.data = response.data.data
    }
    return response
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as RetriableRequest | undefined
	if (originalRequest?._sessionGeneration !== undefined && originalRequest._sessionGeneration !== useAuthStore.getState().sessionGeneration) {
	  return Promise.reject(new Error('Session changed while the request was running'))
	}

    if (error.response?.status === 401 && originalRequest && !originalRequest._retry && !originalRequest._skipReauth) {
      originalRequest._retry = true
      const newToken = await requestReauthentication()
      if (originalRequest._sessionGeneration !== useAuthStore.getState().sessionGeneration) {
        return Promise.reject(new Error('Session changed during re-authentication'))
      }
      originalRequest.headers = originalRequest.headers ?? {}
      originalRequest.headers.Authorization = `Bearer ${newToken}`
      return apiClient.request(originalRequest)
    }

    return Promise.reject(error)
  }
)
