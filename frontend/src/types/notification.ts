export interface NotificationDTO {
  id: string
  type: 'announcement' | 'lesson' | 'assignment' | 'exam' | 'grade'
  title: string
  body: string
  course_id: string
  course_title: string
  due_at?: string | null
  created_at: string
}
