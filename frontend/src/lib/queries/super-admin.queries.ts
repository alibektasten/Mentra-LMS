import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as superAdminApi from '@/lib/api/super-admin'
import { InviteOrgAdminInput } from '@/types/super-admin'

export const superAdminKeys = {
  stats: ['super-admin', 'stats'] as const,
  orgs: (params?: Record<string, unknown>) => ['super-admin', 'orgs', params] as const,
  users: (params?: Record<string, unknown>) => ['super-admin', 'users', params] as const,
}

export function useSystemStats() {
  return useQuery({
    queryKey: superAdminKeys.stats,
    queryFn: () => superAdminApi.getStats(),
  })
}

export function useAdminOrgs(params?: { page?: number; page_size?: number }) {
  return useQuery({
    queryKey: superAdminKeys.orgs(params),
    queryFn: () => superAdminApi.listOrgs(params),
  })
}

export function useAllUsers(params?: { page?: number; page_size?: number }) {
  return useQuery({
    queryKey: superAdminKeys.users(params),
    queryFn: () => superAdminApi.listAllUsers(params),
  })
}

export function useDeleteOrg() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => superAdminApi.deleteOrg(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['super-admin', 'orgs'] })
      qc.invalidateQueries({ queryKey: superAdminKeys.stats })
    },
  })
}

export function useCreateOrg() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: superAdminApi.createOrg,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['super-admin', 'orgs'] })
      qc.invalidateQueries({ queryKey: superAdminKeys.stats })
    },
  })
}

export function useInviteOrgAdmin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: InviteOrgAdminInput) => superAdminApi.inviteOrgAdmin(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['super-admin', 'users'] }),
  })
}
