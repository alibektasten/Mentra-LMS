import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import * as gradesApi from '@/lib/api/grades';

export const gradeKeys = {
  all: ['grades'] as const,
  myGrades: (courseID: string) => [...gradeKeys.all, 'my-grades', courseID] as const,
  deadlines: (courseID: string) => [...gradeKeys.all, 'deadlines', courseID] as const,
  gradebook: (courseID: string) => [...gradeKeys.all, 'gradebook', courseID] as const,
};

export function useMyGrades(courseID: string) {
  return useQuery({
    queryKey: gradeKeys.myGrades(courseID),
    queryFn: () => gradesApi.getMyGrades(courseID),
    enabled: !!courseID,
  });
}

export function useUpcomingDeadlines(courseID: string) {
  return useQuery({
    queryKey: gradeKeys.deadlines(courseID),
    queryFn: () => gradesApi.getUpcomingDeadlines(courseID),
    enabled: !!courseID,
  });
}

export function useGradebook(courseID: string) {
  return useQuery({
    queryKey: gradeKeys.gradebook(courseID),
    queryFn: () => gradesApi.getGradebook(courseID),
    enabled: !!courseID,
  });
}

export function useUpdateCourseGrade(courseID: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ studentID, input }: { studentID: string; input: import('@/types/grade').UpdateCourseGradeInput }) =>
      gradesApi.updateCourseGrade(courseID, studentID, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: gradeKeys.gradebook(courseID) });
      queryClient.invalidateQueries({ queryKey: gradeKeys.all });
    },
  });
}

export function useSetGradebookPublication(courseID: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: { published: boolean; studentIDs?: string[] }) => gradesApi.setGradebookPublication(courseID, input.published, input.studentIDs),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: gradeKeys.gradebook(courseID) });
      queryClient.invalidateQueries({ queryKey: gradeKeys.all });
    },
  });
}
