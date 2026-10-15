import { apiClient } from './client'
import { LessonDTO, CreateLessonInput, UpdateLessonInput } from '@/types/lesson'

export async function listLessons(moduleId: string): Promise<LessonDTO[]> {
  const res = await apiClient.get<LessonDTO[]>(`/modules/${moduleId}/lessons`)
  return res.data
}

export async function getLesson(moduleId: string, lessonId: string): Promise<LessonDTO> {
  const res = await apiClient.get<LessonDTO>(`/modules/${moduleId}/lessons/${lessonId}`)
  return res.data
}

export async function createLesson(moduleId: string, input: CreateLessonInput): Promise<LessonDTO> {
  const res = await apiClient.post<LessonDTO>(`/modules/${moduleId}/lessons`, input)
  return res.data
}

export async function updateLesson(moduleId: string, lessonId: string, input: UpdateLessonInput): Promise<LessonDTO> {
  const res = await apiClient.put<LessonDTO>(`/modules/${moduleId}/lessons/${lessonId}`, input)
  return res.data
}

export async function deleteLesson(moduleId: string, lessonId: string): Promise<void> {
  await apiClient.delete(`/modules/${moduleId}/lessons/${lessonId}`)
}
