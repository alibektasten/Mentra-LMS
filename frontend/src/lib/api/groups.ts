import { apiClient } from './client'
import {
  GroupDTO, GroupScheduleDTO, GroupMemberDTO,
  CreateGroupInput, UpdateGroupInput, CreateGroupScheduleInput, AddMemberInput, AssignGroupInput, AttendanceSheetDTO, SaveAttendanceInput,
} from '@/types/group'

// Org-level group management
export async function listOrgGroups(): Promise<GroupDTO[]> {
  const res = await apiClient.get<GroupDTO[]>('/groups')
  return res.data ?? []
}

export async function createOrgGroup(input: CreateGroupInput): Promise<GroupDTO> {
  const res = await apiClient.post<GroupDTO>('/groups', input)
  return res.data
}

export async function updateOrgGroup(groupId: string, input: UpdateGroupInput): Promise<GroupDTO> {
  const res = await apiClient.put<GroupDTO>(`/groups/${groupId}`, input)
  return res.data
}

export async function deleteOrgGroup(groupId: string): Promise<void> {
  await apiClient.delete(`/groups/${groupId}`)
}

export async function assignGroupToCourse(groupId: string, input: AssignGroupInput): Promise<void> {
  await apiClient.post(`/groups/${groupId}/assign-course`, input)
}

export async function unassignGroupFromCourse(groupId: string): Promise<void> {
  await apiClient.delete(`/groups/${groupId}/assign-course`)
}

export async function listGroups(courseId: string): Promise<GroupDTO[]> {
  const res = await apiClient.get<GroupDTO[]>(`/courses/${courseId}/groups`)
  return res.data
}

export async function getGroup(courseId: string, groupId: string): Promise<GroupDTO> {
  const res = await apiClient.get<GroupDTO>(`/courses/${courseId}/groups/${groupId}`)
  return res.data
}

export async function createGroup(courseId: string, input: CreateGroupInput): Promise<GroupDTO> {
  const res = await apiClient.post<GroupDTO>(`/courses/${courseId}/groups`, input)
  return res.data
}

export async function updateGroup(courseId: string, groupId: string, input: UpdateGroupInput): Promise<GroupDTO> {
  const res = await apiClient.put<GroupDTO>(`/courses/${courseId}/groups/${groupId}`, input)
  return res.data
}

export async function deleteGroup(courseId: string, groupId: string): Promise<void> {
  await apiClient.delete(`/courses/${courseId}/groups/${groupId}`)
}

function membersBase(groupId: string, courseId?: string) {
  return courseId
    ? `/courses/${courseId}/groups/${groupId}/members`
    : `/groups/${groupId}/members`
}

function schedulesBase(groupId: string, courseId?: string) {
  return courseId
    ? `/courses/${courseId}/groups/${groupId}/schedules`
    : `/groups/${groupId}/schedules`
}

export async function listMembers(groupId: string, courseId?: string): Promise<GroupMemberDTO[]> {
  const res = await apiClient.get<GroupMemberDTO[]>(membersBase(groupId, courseId))
  return res.data ?? []
}

export async function addMember(groupId: string, input: AddMemberInput, courseId?: string): Promise<GroupMemberDTO> {
  const res = await apiClient.post<GroupMemberDTO>(membersBase(groupId, courseId), input)
  return res.data
}

export async function removeMember(groupId: string, studentId: string, courseId?: string): Promise<void> {
  await apiClient.delete(`${membersBase(groupId, courseId)}/${studentId}`)
}

export async function listSchedules(groupId: string, courseId?: string): Promise<GroupScheduleDTO[]> {
  const res = await apiClient.get<GroupScheduleDTO[]>(schedulesBase(groupId, courseId))
  return res.data ?? []
}

export async function addSchedule(groupId: string, input: CreateGroupScheduleInput, courseId?: string): Promise<GroupScheduleDTO> {
  const res = await apiClient.post<GroupScheduleDTO>(schedulesBase(groupId, courseId), input)
  return res.data
}

export async function deleteSchedule(groupId: string, scheduleId: string, courseId?: string): Promise<void> {
  await apiClient.delete(`${schedulesBase(groupId, courseId)}/${scheduleId}`)
}

export async function getMyGroup(courseId: string): Promise<GroupDTO> {
  const res = await apiClient.get<GroupDTO>(`/courses/${courseId}/my-group`)
  return res.data
}

export async function getAttendance(courseId: string, groupId: string, date: string): Promise<AttendanceSheetDTO> {
  const res = await apiClient.get<AttendanceSheetDTO>(`/courses/${courseId}/groups/${groupId}/attendance`, { params: { date } })
  return { ...res.data, students: res.data.students ?? [] }
}

export async function saveAttendance(courseId: string, groupId: string, input: SaveAttendanceInput): Promise<AttendanceSheetDTO> {
  const res = await apiClient.put<AttendanceSheetDTO>(`/courses/${courseId}/groups/${groupId}/attendance`, input)
  return res.data
}
