import { isAxiosError } from 'axios';
import { apiClient } from './client';
import { AssignmentDTO, SubmissionDTO, CreateAssignmentInput, UpdateAssignmentInput } from '@/types/assignment';

export async function listAssignments(courseID: string, moduleID: string): Promise<AssignmentDTO[]> {
  const res = await apiClient.get<AssignmentDTO[]>(`/courses/${courseID}/modules/${moduleID}/assignments`);
  return res.data ?? [];
}

export async function createAssignment(courseID: string, moduleID: string, input: CreateAssignmentInput): Promise<AssignmentDTO> {
  const res = await apiClient.post<AssignmentDTO>(`/courses/${courseID}/modules/${moduleID}/assignments`, input);
  return res.data;
}

export async function getAssignment(id: string): Promise<AssignmentDTO> {
  const res = await apiClient.get<AssignmentDTO>(`/assignments/${id}`);
  return res.data;
}

export async function updateAssignment(id: string, input: UpdateAssignmentInput): Promise<AssignmentDTO> {
  const res = await apiClient.put<AssignmentDTO>(`/assignments/${id}`, input);
  return res.data;
}

export async function deleteAssignment(id: string): Promise<void> {
  await apiClient.delete(`/assignments/${id}`);
}

export async function submitAssignment(id: string, formData: FormData): Promise<SubmissionDTO> {
  const res = await apiClient.post<SubmissionDTO>(`/assignments/${id}/submit`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return res.data;
}

export async function getMySubmission(assignmentID: string): Promise<SubmissionDTO | null> {
  try {
    const res = await apiClient.get<SubmissionDTO>(`/assignments/${assignmentID}/my-submission`);
    return res.data;
  } catch (error) {
    // The API represents an assignment with no submission yet as NOT_FOUND.
    // Authentication, authorization, network, and server failures must still surface.
    if (isAxiosError(error) && error.response?.status === 404) {
      return null;
    }

    throw error;
  }
}

export async function listSubmissions(assignmentID: string): Promise<SubmissionDTO[]> {
  const res = await apiClient.get<SubmissionDTO[]>(`/assignments/${assignmentID}/submissions`);
  return res.data;
}

export async function deleteMySubmission(assignmentID: string): Promise<void> {
  await apiClient.delete(`/assignments/${assignmentID}/my-submission`)
}

export async function gradeSubmission(submissionID: string, score: number, feedback: string): Promise<SubmissionDTO> {
  const res = await apiClient.put<SubmissionDTO>(`/submissions/${submissionID}/grade`, { score, feedback });
  return res.data;
}
