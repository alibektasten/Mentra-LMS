import { apiClient } from './client'

export interface TeacherStudentDTO { user_id: string; name: string; email: string; courses: string[]; groups: string[] }
export interface TeacherOverviewDTO { total_courses: number; published_courses: number; draft_courses: number; total_groups: number; total_students: number }
export interface TeacherReviewItemDTO { submission_id: string; assignment_id: string; course_id: string; course_title: string; assignment_title: string; student_name: string; submitted_at: string; max_points: number }
export interface TeacherCalendarEventDTO {
  id: string
  item_id: string
  item_type: 'assignment' | 'exam' | 'class'
  title: string
  course_id: string
  course_title: string
  group_id?: string
  group_name?: string
  assessment_type?: string
  location?: string
  start_at: string
  end_at?: string
}

export async function getTeacherOverview(): Promise<TeacherOverviewDTO> { return (await apiClient.get<TeacherOverviewDTO>('/teacher/overview')).data }
export async function listTeacherStudents(): Promise<TeacherStudentDTO[]> { return (await apiClient.get<TeacherStudentDTO[]>('/teacher/students')).data ?? [] }
export async function listTeacherReviewQueue(): Promise<TeacherReviewItemDTO[]> { return (await apiClient.get<TeacherReviewItemDTO[]>('/teacher/review-queue')).data ?? [] }
export async function listTeacherCalendar(from: string, to: string): Promise<TeacherCalendarEventDTO[]> {
  return (await apiClient.get<TeacherCalendarEventDTO[]>('/teacher/calendar', { params: { from, to } })).data ?? []
}
