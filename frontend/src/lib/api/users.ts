import { apiClient } from '@/lib/api/client'

export interface ProfileResponse {
  id: string
  email: string
  name: string
  avatar_url: string
  created_at: string
  updated_at: string
}

export interface UpdateProfileInput {
  name: string
  avatar_url?: string
}

export async function getMe(): Promise<ProfileResponse> {
  const res = await apiClient.get<ProfileResponse>('/me')
  return res.data
}

export async function updateMe(input: UpdateProfileInput): Promise<ProfileResponse> {
  const res = await apiClient.put<ProfileResponse>('/me', input)
  return res.data
}
