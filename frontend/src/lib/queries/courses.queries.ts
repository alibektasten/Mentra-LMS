import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as coursesApi from '@/lib/api/courses'
import { CreateCourseInput, UpdateCourseInput } from '@/types/course'
import { useAuthStore } from '@/lib/stores/auth.store'

export const courseKeys = {
  all: ['courses'] as const,
  lists: () => [...courseKeys.all, 'list'] as const,
  list: (scope: string, params?: Record<string, unknown>) => [...courseKeys.lists(), scope, params] as const,
  details: () => [...courseKeys.all, 'detail'] as const,
  detail: (scope: string, id: string) => [...courseKeys.details(), scope, id] as const,
}

export function useCourses(params?: { page?: number; page_size?: number }) {
  const scope = useAuthStore((state) => `${state.orgSlug ?? 'none'}:${state.user?.id ?? 'anonymous'}`)
  return useQuery({
    queryKey: courseKeys.list(scope, params),
    queryFn: () => coursesApi.listCourses(params),
    refetchOnMount: 'always',
  })
}

export function useCourse(id: string) {
  const scope = useAuthStore((state) => `${state.orgSlug ?? 'none'}:${state.user?.id ?? 'anonymous'}`)
  return useQuery({
    queryKey: courseKeys.detail(scope, id),
    queryFn: () => coursesApi.getCourse(id),
    enabled: !!id && scope !== 'none:anonymous',
    refetchOnMount: 'always',
  })
}

export function useCreateCourse() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateCourseInput) => coursesApi.createCourse(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: courseKeys.lists() }),
  })
}

export function useUpdateCourse() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateCourseInput }) =>
      coursesApi.updateCourse(id, input),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: courseKeys.details() })
      qc.invalidateQueries({ queryKey: courseKeys.lists() })
    },
  })
}

export function useDeleteCourse() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => coursesApi.deleteCourse(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: courseKeys.lists() }),
  })
}
