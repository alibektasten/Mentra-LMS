export type CourseStatus = 'draft' | 'published' | 'archived'

export interface CourseDTO {
  id: string
  title: string
  description: string
  status: CourseStatus
  org_id: string
  teacher_id: string
  created_at: string
  updated_at: string
}

export interface CreateCourseInput {
  title: string
  description: string
  status?: CourseStatus
}

export interface UpdateCourseInput {
  title?: string
  description?: string
  status?: CourseStatus
}
