export interface ExamAnswerDTO {
  id: string
  answer: string
  is_correct: boolean
}

export interface ExamQuestionDTO {
  id: string
  question: string
  position: number
  answers: ExamAnswerDTO[]
}

export interface ExamDTO {
  id: string
  course_id: string
  org_id: string
  title: string
  description: string
  duration_minutes: number
  max_attempts: number
  total_points: number
  due_date: string | null
  mcq_enabled: boolean
  mcq_points: number
  file_enabled: boolean
  file_points: number
  questions: ExamQuestionDTO[]
  created_at: string
  updated_at: string
}

export interface ExamListItemDTO {
  id: string
  course_id: string
  org_id: string
  title: string
  description: string
  duration_minutes: number
  max_attempts: number
  total_points: number
  due_date: string | null
  mcq_enabled: boolean
  mcq_points: number
  file_enabled: boolean
  file_points: number
  created_at: string
  updated_at: string
}

export interface ExamMCQAnswerInput {
  question_id: string
  answer_id: string
}

export interface ExamAttemptDTO {
  id: string
  exam_id: string
  student_id: string
  status: 'in_progress' | 'submitted' | 'expired'
  started_at: string
  expires_at: string
  submitted_at: string | null
  mcq_answers: ExamMCQAnswerInput[]
  mcq_score: number | null
  mcq_max_score: number
  file_path: string
  file_score: number | null
  file_points: number
  file_feedback: string
  total_score: number | null
  graded_at: string | null
}

export interface StartAttemptResponse {
  attempt_id: string
  exam_id: string
  started_at: string
  expires_at: string
  saved_answers: ExamMCQAnswerInput[]
  exam: ExamDTO
}

export interface CreateExamAnswerInput {
  answer: string
  is_correct: boolean
}

export interface CreateExamQuestionInput {
  question: string
  position: number
  answers: CreateExamAnswerInput[]
}

export interface CreateExamInput {
  title: string
  description?: string
  duration_minutes: number
  max_attempts?: number
  due_date?: string | null
  mcq_enabled: boolean
  mcq_points?: number
  file_enabled: boolean
  file_points?: number
  questions?: CreateExamQuestionInput[]
}

export interface UpdateExamInput {
  title?: string
  description?: string
  duration_minutes?: number
  max_attempts?: number
  due_date?: string | null
  mcq_enabled?: boolean
  mcq_points?: number
  file_enabled?: boolean
  file_points?: number
  questions?: CreateExamQuestionInput[]
}

export interface GradeExamFileInput {
  score: number
  feedback?: string
}

export interface GrantExtraAttemptInput {
  student_id: string
  extra_count: number
}
