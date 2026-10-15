export type LessonType = 'video' | 'text' | 'quiz' | 'pdf' | 'link'

export interface LessonDTO {
  id: string
  module_id: string
  title: string
  type: LessonType
  content: string
  video_url?: string
  link_url?: string
  file_url?: string
  order: number
  created_at: string
  updated_at: string
}

export interface CreateLessonInput {
  title: string
  type: LessonType
  content: string
  video_url?: string
  link_url?: string
  file_url?: string
  upload_token?: string
  order?: number
}

export interface UpdateLessonInput {
  title?: string
  type?: LessonType
  content?: string
  video_url?: string
  link_url?: string
  file_url?: string
  upload_token?: string
  order?: number
}
