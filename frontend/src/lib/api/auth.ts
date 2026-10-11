import axios from 'axios'
import { Role, UserDTO } from '@/types/auth'
import { ApiEnvelope } from '@/types/api'

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1'

export interface LoginInput {
  email: string
  password: string
}

// Shape the backend actually returns
export interface BackendLoginResponse {
  access_token: string
  expires_at: string
  user: {
    id: string
    email: string
    name: string
  }
}

export interface LoginResponse {
  token: string
  expiresAt: string
  user: UserDTO
  role: Role
}

function decodeJwtPayload(token: string): Record<string, unknown> {
  try {
    return JSON.parse(atob(token.split('.')[1]))
  } catch {
    return {}
  }
}

export function toLoginResponse(raw: BackendLoginResponse): LoginResponse {
  const payload = decodeJwtPayload(raw.access_token)
  const role = (payload.role as Role) ?? 'student'
  const nameParts = raw.user.name.split(' ')
  return {
    token: raw.access_token,
    expiresAt: raw.expires_at,
    role,
    user: {
      id: raw.user.id,
      email: raw.user.email,
      first_name: nameParts[0] ?? '',
      last_name: nameParts.slice(1).join(' ') ?? '',
      role,
      org_id: (payload.org_id as string) ?? '',
    },
  }
}

export function parseGoogleOAuthPayload(value: string): LoginResponse {
  const base64 = value.replace(/-/g, '+').replace(/_/g, '/')
  const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, '=')
  return toLoginResponse(JSON.parse(atob(padded)) as BackendLoginResponse)
}

export async function getGoogleOAuthStatus(): Promise<boolean> {
  const res = await axios.get<ApiEnvelope<{ enabled: boolean }>>(`${API_URL}/auth/google/status`)
  return res.data.data?.enabled ?? false
}

export function getGoogleOAuthStartURL(orgSlug: string): string {
  return `${API_URL}/auth/google/start?org=${encodeURIComponent(orgSlug)}`
}

// Raw axios (no interceptor) for auth calls to avoid circular re-auth loops
export async function login(orgSlug: string, input: LoginInput): Promise<LoginResponse> {
  const res = await axios.post<ApiEnvelope<BackendLoginResponse>>(
    `${API_URL}/auth/login`,
    input,
    { headers: { 'X-Org-Slug': orgSlug, 'Content-Type': 'application/json' } }
  )
  return toLoginResponse(res.data.data)
}

export interface RegisterInput {
  email: string
  password: string
  first_name: string
  last_name: string
  role?: string
}

export async function superAdminLogin(input: LoginInput): Promise<LoginResponse> {
  const res = await axios.post<ApiEnvelope<BackendLoginResponse>>(
    `${API_URL}/super-admin/auth/login`,
    input,
    { headers: { 'Content-Type': 'application/json' } }
  )
  return toLoginResponse(res.data.data)
}

export async function register(orgSlug: string, input: RegisterInput): Promise<UserDTO> {
  const res = await axios.post<ApiEnvelope<UserDTO>>(
    `${API_URL}/auth/register`,
    input,
    { headers: { 'X-Org-Slug': orgSlug, 'Content-Type': 'application/json' } }
  )
  return res.data.data
}

export async function getMe(token: string, orgSlug: string): Promise<UserDTO> {
  const res = await axios.get<ApiEnvelope<UserDTO>>(
    `${API_URL}/auth/me`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
        'X-Org-Slug': orgSlug,
        'Content-Type': 'application/json',
      },
    }
  )
  return res.data.data
}
