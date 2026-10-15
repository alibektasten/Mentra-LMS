import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as groupsApi from '@/lib/api/groups'
import {
  CreateGroupInput, UpdateGroupInput, CreateGroupScheduleInput, AddMemberInput, AssignGroupInput,
} from '@/types/group'

export const groupKeys = {
  all: ['groups'] as const,
  orgList: () => [...groupKeys.all, 'org-list'] as const,
  lists: () => [...groupKeys.all, 'list'] as const,
  list: (courseId: string) => [...groupKeys.lists(), courseId] as const,
  details: () => [...groupKeys.all, 'detail'] as const,
  detail: (courseId: string, groupId: string) => [...groupKeys.details(), courseId, groupId] as const,
  // courseId optional: omit it for standalone (org-level) member/schedule queries
  members: (groupId: string, courseId?: string) => ['groups', groupId, 'members', courseId ?? 'standalone'] as const,
  schedules: (groupId: string, courseId?: string) => ['groups', groupId, 'schedules', courseId ?? 'standalone'] as const,
  attendance: (courseId: string, groupId: string, date: string) => [...groupKeys.all, 'attendance', courseId, groupId, date] as const,
  myGroup: (courseId: string) => [...groupKeys.all, 'my-group', courseId] as const,
}

export function useOrgGroups() {
  return useQuery({
    queryKey: groupKeys.orgList(),
    queryFn: () => groupsApi.listOrgGroups(),
  })
}

export function useCreateOrgGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateGroupInput) => groupsApi.createOrgGroup(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: groupKeys.orgList() }),
  })
}

export function useUpdateOrgGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateGroupInput }) =>
      groupsApi.updateOrgGroup(id, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: groupKeys.orgList() })
      qc.invalidateQueries({ queryKey: groupKeys.lists() })
    },
  })
}

export function useDeleteOrgGroup() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (groupId: string) => groupsApi.deleteOrgGroup(groupId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: groupKeys.orgList() })
      qc.invalidateQueries({ queryKey: groupKeys.lists() })
    },
  })
}

export function useAssignGroupToCourse() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ groupId, input }: { groupId: string; input: AssignGroupInput }) =>
      groupsApi.assignGroupToCourse(groupId, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: groupKeys.orgList() })
      qc.invalidateQueries({ queryKey: groupKeys.lists() })
    },
  })
}

export function useUnassignGroupFromCourse() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (groupId: string) => groupsApi.unassignGroupFromCourse(groupId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: groupKeys.orgList() })
      qc.invalidateQueries({ queryKey: groupKeys.lists() })
    },
  })
}

export function useGroups(courseId: string) {
  return useQuery({
    queryKey: groupKeys.list(courseId),
    queryFn: () => groupsApi.listGroups(courseId),
    enabled: !!courseId,
  })
}

export function useGroup(courseId: string, groupId: string) {
  return useQuery({
    queryKey: groupKeys.detail(courseId, groupId),
    queryFn: () => groupsApi.getGroup(courseId, groupId),
    enabled: !!(courseId && groupId),
  })
}

export function useCreateGroup(courseId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateGroupInput) => groupsApi.createGroup(courseId, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: groupKeys.list(courseId) }),
  })
}

export function useUpdateGroup(courseId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateGroupInput }) =>
      groupsApi.updateGroup(courseId, id, input),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: groupKeys.list(courseId) })
      qc.invalidateQueries({ queryKey: groupKeys.detail(courseId, id) })
    },
  })
}

export function useDeleteGroup(courseId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (groupId: string) => groupsApi.deleteGroup(courseId, groupId),
    onSuccess: () => qc.invalidateQueries({ queryKey: groupKeys.list(courseId) }),
  })
}

export function useGroupMembers(groupId: string, courseId?: string) {
  return useQuery({
    queryKey: groupKeys.members(groupId, courseId),
    queryFn: () => groupsApi.listMembers(groupId, courseId),
    enabled: !!groupId,
  })
}

export function useAddMember(groupId: string, courseId?: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: AddMemberInput) => groupsApi.addMember(groupId, input, courseId),
    onSuccess: () => qc.invalidateQueries({ queryKey: groupKeys.members(groupId, courseId) }),
  })
}

export function useRemoveMember(groupId: string, courseId?: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (studentId: string) => groupsApi.removeMember(groupId, studentId, courseId),
    onSuccess: () => qc.invalidateQueries({ queryKey: groupKeys.members(groupId, courseId) }),
  })
}

export function useGroupSchedules(groupId: string, courseId?: string) {
  return useQuery({
    queryKey: groupKeys.schedules(groupId, courseId),
    queryFn: () => groupsApi.listSchedules(groupId, courseId),
    enabled: !!groupId,
  })
}

export function useAddSchedule(groupId: string, courseId?: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateGroupScheduleInput) => groupsApi.addSchedule(groupId, input, courseId),
    onSuccess: () => qc.invalidateQueries({ queryKey: groupKeys.schedules(groupId, courseId) }),
  })
}

export function useDeleteSchedule(groupId: string, courseId?: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (scheduleId: string) => groupsApi.deleteSchedule(groupId, scheduleId, courseId),
    onSuccess: () => qc.invalidateQueries({ queryKey: groupKeys.schedules(groupId, courseId) }),
  })
}

export function useMyGroup(courseId: string) {
  return useQuery({
    queryKey: groupKeys.myGroup(courseId),
    queryFn: () => groupsApi.getMyGroup(courseId),
    enabled: !!courseId,
    retry: false,
  })
}

export function useAttendance(courseId: string, groupId: string, date: string) {
  return useQuery({
    queryKey: groupKeys.attendance(courseId, groupId, date),
    queryFn: () => groupsApi.getAttendance(courseId, groupId, date),
    enabled: Boolean(courseId && groupId && date),
  })
}

export function useSaveAttendance(courseId: string, groupId: string, date: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: import('@/types/group').SaveAttendanceInput) => groupsApi.saveAttendance(courseId, groupId, input),
    onSuccess: () => qc.invalidateQueries({ queryKey: groupKeys.attendance(courseId, groupId, date) }),
  })
}
