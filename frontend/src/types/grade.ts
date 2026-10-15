export interface GradeItemDTO {
  item_id: string;
  item_type: 'assignment' | 'quiz' | 'midterm' | 'endterm' | 'final';
  title: string;
  max_points: number;
  score: number | null;
  status: 'not_submitted' | 'submitted' | 'graded';
  feedback: string;
  submitted_at: string | null;
  graded_at: string | null;
}

export interface StudentGradeDTO {
  student_id: string;
  items: GradeItemDTO[];
  total_earned: number;
  total_possible: number;
  percentage?: number;
  assignments_percent: number | null;
  midterm_percent: number | null;
  endterm_percent: number | null;
  final_exam_percent: number | null;
  midterm_register: number | null;
  endterm_register: number | null;
  calculated_final_grade?: number;
  final_grade?: number;
  is_final_grade_ready: boolean;
  override_grade: number | null;
  override_note: string;
  is_published: boolean;
  published_at: string | null;
}

export interface GradebookStudentDTO {
  student_id: string;
  student_name: string;
  student_email: string;
  grade: StudentGradeDTO;
}

export interface GradebookDTO {
  course_id: string;
  students: GradebookStudentDTO[];
  published_count: number;
  generated_at: string;
}

export interface UpdateCourseGradeInput {
  override_grade?: number;
  override_note?: string;
  clear_override?: boolean;
}

export interface GradebookPublicationResultDTO {
  published: boolean;
  affected_count: number;
  published_at: string | null;
}

export interface DeadlineItemDTO {
  item_id: string;
  item_type: 'assignment' | 'quiz' | 'midterm' | 'endterm' | 'final';
  title: string;
  due_date: string;
  submitted: boolean;
}
