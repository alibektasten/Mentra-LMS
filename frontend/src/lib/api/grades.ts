import { apiClient } from './client';
import { StudentGradeDTO, DeadlineItemDTO, GradebookDTO, GradebookPublicationResultDTO, UpdateCourseGradeInput } from '@/types/grade';

export async function getMyGrades(courseID: string): Promise<StudentGradeDTO> {
  const res = await apiClient.get<StudentGradeDTO>(`/courses/${courseID}/my-grades`);
  return { ...res.data, items: res.data.items ?? [] };
}

export async function getUpcomingDeadlines(courseID: string): Promise<DeadlineItemDTO[]> {
  const res = await apiClient.get<DeadlineItemDTO[]>(`/courses/${courseID}/deadlines`);
  return res.data;
}

export async function getGradebook(courseID: string): Promise<GradebookDTO> {
  const res = await apiClient.get<GradebookDTO>(`/courses/${courseID}/gradebook`);
  return { ...res.data, students: res.data.students ?? [] };
}

export async function updateCourseGrade(courseID: string, studentID: string, input: UpdateCourseGradeInput): Promise<StudentGradeDTO> {
  const res = await apiClient.patch<StudentGradeDTO>(`/courses/${courseID}/gradebook/students/${studentID}`, input);
  return res.data;
}

export async function setGradebookPublication(courseID: string, published: boolean, studentIDs?: string[]): Promise<GradebookPublicationResultDTO> {
  const res = await apiClient.put<GradebookPublicationResultDTO>(`/courses/${courseID}/gradebook/publication`, {
    published,
    student_ids: studentIDs,
  });
  return res.data;
}
