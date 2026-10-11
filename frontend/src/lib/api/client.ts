import axios, { AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/lib/stores/auth.store'
import { useReAuthStore } from '@/lib/stores/reauth.store'

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1'

// Base host for static file serving (strips /api/v1 suffix)
export const UPLOAD_BASE_URL = API_URL.replace(/\/api\/v\d+$/, '')

export const apiClient = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' },
})

// Queue of pending requests awaiting re-auth
let pendingRequests: Array<{
  resolve: (value: string) => void
  reject: (reason?: unknown) => void
}> = []

function processQueue(token: string | null, error?: unknown) {
  pendingRequests.forEach(({ resolve, reject }) => {
    if (token) resolve(token)
    else reject(error)
  })
  pendingRequests = []
}

// Request interceptor: attach auth + org slug
apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const { token, orgSlug } = useAuthStore.getState()
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
  async (error) => {
    const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean }

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      const { user } = useAuthStore.getState()
      const email = user?.email ?? ''

      return new Promise((resolve, reject) => {
        pendingRequests.push({ resolve, reject })

        useReAuthStore.getState().open(email, {
          onSuccess: (newToken: string) => {
            processQueue(newToken)
            if (originalRequest.headers) {
              (originalRequest.headers as Record<string, string>)['Authorization'] = `Bearer ${newToken}`
            }
            resolve(apiClient(originalRequest))
          },
          onCancel: () => {
            processQueue(null, new Error('Re-authentication cancelled'))
            reject(error)
          },
        })
      })
    }

    return Promise.reject(error)
  }
)
