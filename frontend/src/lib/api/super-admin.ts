import { apiClient } from './client'
import { AdminUserDTO, InviteOrgAdminInput, OrgDTO, SystemStatsDTO } from '@/types/super-admin'
import { PaginatedResponse } from '@/types/api'

export async function getStats(): Promise<SystemStatsDTO> {
  const res = await apiClient.get<SystemStatsDTO>('/super-admin/stats')
  return res.data
}

export async function listOrgs(params?: { page?: number; page_size?: number }): Promise<PaginatedResponse<OrgDTO>> {
  const res = await apiClient.get('/super-admin/orgs', { params })
  return res.data
}

export async function createOrg(input: { name: string; slug: string }): Promise<OrgDTO> {
  const res = await apiClient.post<OrgDTO>('/super-admin/orgs', input)
  return res.data
}

export async function deleteOrg(id: string): Promise<void> {
  await apiClient.delete(`/super-admin/orgs/${id}`)
}

export async function listAllUsers(params?: { page?: number; page_size?: number }): Promise<PaginatedResponse<AdminUserDTO>> {
  const res = await apiClient.get('/super-admin/users', { params })
  return res.data
}

export async function inviteOrgAdmin(input: InviteOrgAdminInput): Promise<AdminUserDTO> {
  const res = await apiClient.post<AdminUserDTO>('/super-admin/orgs/invite-admin', input)
  return res.data
}
