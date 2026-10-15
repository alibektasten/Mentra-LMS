import { apiClient } from './client'
import { QuizDTO, GenerateQuizInput } from '@/types/quiz'
import { InsightsDTO } from '@/types/progress'

export interface AssignmentFeedbackDTO {
  strengths: string[]
  gaps: string[]
  improvements: string[]
  overall: string
}

export interface FlashcardDTO {
  term: string
  definition: string
}

export interface GenerateTeachingPlanInput {
  topic: string
  audience: string
  duration_minutes: number
  language: 'English' | 'Russian' | 'Kazakh'
}

export async function generateQuiz(input: GenerateQuizInput): Promise<QuizDTO> {
  const res = await apiClient.post<QuizDTO>('/ai/generate-quiz', input)
  return res.data
}

export async function summarizeLesson(lessonId: string): Promise<{ summary: string }> {
  const res = await apiClient.post<{ summary: string }>('/ai/summarize', { lesson_id: lessonId })
  return res.data
}

export async function getAIInsights(): Promise<InsightsDTO> {
  const res = await apiClient.get<InsightsDTO>('/progress/insights')
  return res.data
}

export async function getAssignmentFeedback(submissionId: string): Promise<AssignmentFeedbackDTO> {
  const res = await apiClient.post<AssignmentFeedbackDTO>('/ai/assignment-feedback', { submission_id: submissionId })
  return res.data
}

export async function generateFlashcards(lessonId: string, numCards: number): Promise<FlashcardDTO[]> {
  const res = await apiClient.post<FlashcardDTO[]>('/ai/generate-flashcards', { lesson_id: lessonId, num_cards: numCards })
  return res.data
}

export async function generateTeachingPlan(input: GenerateTeachingPlanInput): Promise<{ plan: string }> {
  const res = await apiClient.post<{ plan: string }>('/ai/generate-teaching-plan', input)
  return res.data
}
