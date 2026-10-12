export interface AssignmentDTO {
  id: string;
  course_id: string;
  module_id: string;
  title: string;
  description: string;
  rubric: string;
  max_points: number;
  due_date: string | null;
  allow_late_submission: boolean;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface SubmissionDTO {
  id: string;
  assignment_id: string;
  student_id: string;
  student_name: string;
  text_content: string;
  link_url: string;
  file_path: string;
  score: number | null;
  feedback: string;
  graded_at: string | null;
  submitted_at: string;
}

export interface CreateAssignmentInput {
  title: string;
  description?: string;
  rubric?: string;
  max_points: number;
  due_date?: string;
  allow_late_submission?: boolean;
}

export interface UpdateAssignmentInput {
  title?: string;
  description?: string;
  rubric?: string;
  max_points?: number;
  due_date?: string;
  allow_late_submission?: boolean;
}
