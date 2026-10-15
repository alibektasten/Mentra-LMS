export interface ProgressDTO {
  id: string
  user_id: string
  lesson_id: string
  org_id: string
  score?: number
  completed_at?: string | null
  created_at: string
}

export interface InsightsDTO {
  insights: string
  total_lessons: number
  completed_lessons: number
  average_score: number
}

export interface CourseProgressSummary {
  course_id: string
  total_lessons: number
  completed_lessons: number
  percentage: number
  average_score?: number
}
