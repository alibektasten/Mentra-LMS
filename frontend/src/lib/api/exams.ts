import { apiClient } from './client'
import {
  ExamDTO,
  ExamListItemDTO,
  ExamAttemptDTO,
  StartAttemptResponse,
  CreateExamInput,
  UpdateExamInput,
  GradeExamFileInput,
  GrantExtraAttemptInput,
  ExamMCQAnswerInput,
} from '@/types/exam'

export async function listExams(courseID: string): Promise<ExamListItemDTO[]> {
  const res = await apiClient.get<ExamListItemDTO[]>(`/courses/${courseID}/exams`)
  return res.data ?? []
}

export async function createExam(courseID: string, input: CreateExamInput): Promise<ExamDTO> {
  const res = await apiClient.post<ExamDTO>(`/courses/${courseID}/exams`, input)
  return res.data
}

export async function getExam(id: string): Promise<ExamDTO> {
  const res = await apiClient.get<ExamDTO>(`/exams/${id}`)
  return res.data
}

export async function updateExam(id: string, input: UpdateExamInput): Promise<ExamDTO> {
  const res = await apiClient.put<ExamDTO>(`/exams/${id}`, input)
  return res.data
}

export async function deleteExam(id: string): Promise<void> {
  await apiClient.delete(`/exams/${id}`)
}

export async function startExamAttempt(examID: string): Promise<StartAttemptResponse> {
  const res = await apiClient.post<StartAttemptResponse>(`/exams/${examID}/start`)
  return res.data
}

export async function submitExamAttempt(attemptID: string, formData: FormData): Promise<ExamAttemptDTO> {
  const res = await apiClient.post<ExamAttemptDTO>(`/exam-attempts/${attemptID}/submit`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data
}

export async function saveExamDraft(attemptID: string, mcqAnswers: ExamMCQAnswerInput[]): Promise<void> {
  await apiClient.put(`/exam-attempts/${attemptID}/draft`, { mcq_answers: mcqAnswers })
}

export async function getMyAttempts(examID: string): Promise<ExamAttemptDTO[]> {
  const res = await apiClient.get<ExamAttemptDTO[]>(`/exams/${examID}/my-attempts`)
  return res.data
}

export async function listAttempts(examID: string): Promise<ExamAttemptDTO[]> {
  const res = await apiClient.get<ExamAttemptDTO[]>(`/exams/${examID}/attempts`)
  return res.data
}

export async function gradeExamFile(attemptID: string, input: GradeExamFileInput): Promise<ExamAttemptDTO> {
  const res = await apiClient.put<ExamAttemptDTO>(`/exam-attempts/${attemptID}/grade`, input)
  return res.data
}

export async function grantExtraAttempt(examID: string, input: GrantExtraAttemptInput): Promise<void> {
  await apiClient.post(`/exams/${examID}/grant-attempt`, input)
}
