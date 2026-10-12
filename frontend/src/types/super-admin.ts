export interface OrgDTO {
  id: string
  name: string
  slug: string
  created_at: string
}

export interface AdminUserDTO {
  id: string
  email: string
  name: string
  role: string
  org_id: string
  org_name: string
  created_at: string
}

export interface SystemUserFilters {
  page?: number
  page_size?: number
  search?: string
  role?: string
  org_id?: string
  sort_by?: 'name' | 'email' | 'role' | 'org_name' | 'created_at'
  sort_order?: 'asc' | 'desc'
}

export interface SystemStatsDTO {
  total_orgs: number
  total_users: number
}

export interface InviteOrgAdminInput {
  email: string
  name: string
  password: string
  org_id: string
}
