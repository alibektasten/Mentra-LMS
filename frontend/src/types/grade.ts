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
  percentage: number;
  assignments_percent: number | null;
  midterm_percent: number | null;
  endterm_percent: number | null;
  final_exam_percent: number | null;
  midterm_register: number | null;
  endterm_register: number | null;
  final_grade: number;
  is_final_grade_ready: boolean;
}

export interface DeadlineItemDTO {
  item_id: string;
  item_type: 'assignment' | 'quiz';
  title: string;
  due_date: string;
  submitted: boolean;
}
