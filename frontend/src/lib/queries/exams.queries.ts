import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as examsApi from '@/lib/api/exams'
import { CreateExamInput, UpdateExamInput, GradeExamFileInput, GrantExtraAttemptInput } from '@/types/exam'

export const examKeys = {
  all: ['exams'] as const,
  byCourse: (courseID: string) => [...examKeys.all, 'course', courseID] as const,
  detail: (id: string) => [...examKeys.all, 'detail', id] as const,
  myAttempts: (examID: string) => [...examKeys.all, 'my-attempts', examID] as const,
  attempts: (examID: string) => [...examKeys.all, 'attempts', examID] as const,
}

export function useExams(courseID: string) {
  return useQuery({
    queryKey: examKeys.byCourse(courseID),
    queryFn: () => examsApi.listExams(courseID),
    enabled: !!courseID,
    retry: false,
  })
}

export function useExam(id: string) {
  return useQuery({
    queryKey: examKeys.detail(id),
    queryFn: () => examsApi.getExam(id),
    enabled: !!id,
    retry: false,
  })
}

export function useCreateExam(courseID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateExamInput) => examsApi.createExam(courseID, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: examKeys.byCourse(courseID) })
    },
  })
}

export function useUpdateExam(courseID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateExamInput }) => examsApi.updateExam(id, data),
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: examKeys.byCourse(courseID) })
      qc.invalidateQueries({ queryKey: examKeys.detail(id) })
    },
  })
}

export function useDeleteExam(courseID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => examsApi.deleteExam(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: examKeys.byCourse(courseID) })
    },
  })
}

export function useStartExamAttempt(examID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => examsApi.startExamAttempt(examID),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: examKeys.myAttempts(examID) })
    },
  })
}

export function useSubmitExamAttempt(examID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ attemptID, formData }: { attemptID: string; formData: FormData }) =>
      examsApi.submitExamAttempt(attemptID, formData),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: examKeys.myAttempts(examID) })
    },
  })
}

export function useSaveExamDraft() {
  return useMutation({
    mutationFn: ({ attemptID, mcqAnswers }: { attemptID: string; mcqAnswers: import('@/types/exam').ExamMCQAnswerInput[] }) =>
      examsApi.saveExamDraft(attemptID, mcqAnswers),
  })
}

export function useMyAttempts(examID: string) {
  return useQuery({
    queryKey: examKeys.myAttempts(examID),
    queryFn: () => examsApi.getMyAttempts(examID),
    enabled: !!examID,
    retry: false,
  })
}

export function useListAttempts(examID: string) {
  return useQuery({
    queryKey: examKeys.attempts(examID),
    queryFn: () => examsApi.listAttempts(examID),
    enabled: !!examID,
    retry: false,
  })
}

export function useGradeExamFile(examID: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ attemptID, data }: { attemptID: string; data: GradeExamFileInput }) =>
      examsApi.gradeExamFile(attemptID, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: examKeys.attempts(examID) })
    },
  })
}

export function useGrantExtraAttempt(examID: string) {
  return useMutation({
    mutationFn: (data: GrantExtraAttemptInput) => examsApi.grantExtraAttempt(examID, data),
  })
}
