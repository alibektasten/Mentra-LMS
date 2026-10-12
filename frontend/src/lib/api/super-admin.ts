import { apiClient } from './client'
import { AdminUserDTO, InviteOrgAdminInput, OrgDTO, SystemStatsDTO, SystemUserFilters } from '@/types/super-admin'
import { PaginatedResponse } from '@/types/api'

export async function getStats(): Promise<SystemStatsDTO> {
  const res = await apiClient.get<SystemStatsDTO>('/super-admin/stats')
  // Axios unwraps the API envelope in an interceptor. During a browser session
  // renewal a stale dev bundle can briefly yield an empty response; never pass
  // undefined to React Query, which treats it as a programming error.
  return res.data ?? { total_orgs: 0, total_users: 0 }
}

export async function listOrgs(params?: { page?: number; page_size?: number }): Promise<PaginatedResponse<OrgDTO>> {
  const res = await apiClient.get('/super-admin/orgs', { params })
  return res.data
}

export async function createOrg(input: { name: string; slug: string }): Promise<OrgDTO> {
  const res = await apiClient.post<OrgDTO>('/super-admin/orgs', input)
  return res.data
}

export async function updateOrg(id: string, input: { name: string; slug: string }): Promise<OrgDTO> {
  const res = await apiClient.put<OrgDTO>(`/super-admin/orgs/${id}`, input)
  return res.data
}

export async function deleteOrg(id: string): Promise<void> {
  await apiClient.delete(`/super-admin/orgs/${id}`)
}

export async function listAllUsers(params?: SystemUserFilters): Promise<PaginatedResponse<AdminUserDTO>> {
  const res = await apiClient.get('/super-admin/users', { params })
  return res.data
}

export async function updateSystemUserRole(membershipID: string, role: string): Promise<void> { await apiClient.put(`/super-admin/users/${membershipID}/role`, { role }) }

export async function inviteOrgAdmin(input: InviteOrgAdminInput): Promise<AdminUserDTO> {
  const res = await apiClient.post<AdminUserDTO>('/super-admin/orgs/invite-admin', input)
  return res.data
}
