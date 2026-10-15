export type Role = 'super_admin' | 'admin' | 'teacher' | 'student'
export type AuthProvider = 'password' | 'google'

export interface UserDTO {
  id: string
  email: string
  first_name: string
  last_name: string
  role: Role
  org_id: string
}

export interface TokenResponse {
  access_token: string
  token_type: string
  expires_in: number
}

export interface AuthSession {
  token: string
  user: UserDTO
  role: Role
  orgSlug: string
  expiresAt: string
  authProvider?: AuthProvider
}
