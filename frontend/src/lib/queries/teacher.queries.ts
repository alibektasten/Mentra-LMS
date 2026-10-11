import { useQuery } from '@tanstack/react-query'
import * as teacherApi from '@/lib/api/teacher'
import { useAuthStore } from '@/lib/stores/auth.store'

export const teacherKeys = { all: ['teacher'] as const, overview: (scope: string) => ['teacher', 'overview', scope] as const, students: (scope: string) => ['teacher', 'students', scope] as const, reviewQueue: (scope: string) => ['teacher', 'review-queue', scope] as const }
function useTeacherScope() { return useAuthStore((state) => `${state.orgSlug ?? 'none'}:${state.user?.id ?? 'anonymous'}`) }
export function useTeacherOverview() { const scope = useTeacherScope(); return useQuery({ queryKey: teacherKeys.overview(scope), queryFn: teacherApi.getTeacherOverview, enabled: scope !== 'none:anonymous', refetchOnMount: 'always' }) }
export function useTeacherStudents() { const scope = useTeacherScope(); return useQuery({ queryKey: teacherKeys.students(scope), queryFn: teacherApi.listTeacherStudents, enabled: scope !== 'none:anonymous', refetchOnMount: 'always' }) }
export function useTeacherReviewQueue() { const scope = useTeacherScope(); return useQuery({ queryKey: teacherKeys.reviewQueue(scope), queryFn: teacherApi.listTeacherReviewQueue, enabled: scope !== 'none:anonymous', refetchOnMount: 'always' }) }
