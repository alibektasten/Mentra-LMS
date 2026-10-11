import { apiClient } from './client'
import { CourseDTO, CreateCourseInput, UpdateCourseInput } from '@/types/course'
import { PaginatedResponse } from '@/types/api'

export async function listCourses(params?: { page?: number; page_size?: number }): Promise<PaginatedResponse<CourseDTO>> {
  const res = await apiClient.get<PaginatedResponse<CourseDTO>>('/courses', { params })
  return res.data ?? { data: [], meta: { page: params?.page ?? 1, page_size: params?.page_size ?? 10, total: 0 } }
}

export async function getCourse(id: string): Promise<CourseDTO | null> {
  const res = await apiClient.get<CourseDTO>(`/courses/${id}`)
  return res.data ?? null
}

export async function createCourse(input: CreateCourseInput): Promise<CourseDTO> {
  const res = await apiClient.post<CourseDTO>('/courses', input)
  return res.data
}

export async function updateCourse(id: string, input: UpdateCourseInput): Promise<CourseDTO> {
	const res = await apiClient.put<CourseDTO>(`/courses/${id}`, input)
  return res.data
}

export async function deleteCourse(id: string): Promise<void> {
  await apiClient.delete(`/courses/${id}`)
}
