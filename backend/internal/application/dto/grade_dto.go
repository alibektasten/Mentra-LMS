package dto

import "time"

type GradeItemDTO struct {
	ItemID      string     `json:"item_id"`
	ItemType    string     `json:"item_type"`
	Title       string     `json:"title"`
	MaxPoints   int        `json:"max_points"`
	Score       *int       `json:"score"`
	Status      string     `json:"status"`
	Feedback    string     `json:"feedback"`
	SubmittedAt *time.Time `json:"submitted_at"`
	GradedAt    *time.Time `json:"graded_at"`
}

type StudentGradeDTO struct {
	StudentID            string         `json:"student_id"`
	Items                []GradeItemDTO `json:"items"`
	TotalEarned          int            `json:"total_earned"`
	TotalPossible        int            `json:"total_possible"`
	Percentage           *float64       `json:"percentage,omitempty"`
	AssignmentsPercent   *float64       `json:"assignments_percent"`
	MidtermPercent       *float64       `json:"midterm_percent"`
	EndtermPercent       *float64       `json:"endterm_percent"`
	FinalExamPercent     *float64       `json:"final_exam_percent"`
	MidtermRegister      *float64       `json:"midterm_register"`
	EndtermRegister      *float64       `json:"endterm_register"`
	CalculatedFinalGrade *float64       `json:"calculated_final_grade,omitempty"`
	FinalGrade           *float64       `json:"final_grade,omitempty"`
	IsFinalGradeReady    bool           `json:"is_final_grade_ready"`
	OverrideGrade        *float64       `json:"override_grade,omitempty"`
	OverrideNote         string         `json:"override_note,omitempty"`
	IsPublished          bool           `json:"is_published"`
	PublishedAt          *time.Time     `json:"published_at"`
}

// GradebookStudentDTO is a row in a teacher's academic register.
type GradebookStudentDTO struct {
	StudentID    string           `json:"student_id"`
	StudentName  string           `json:"student_name"`
	StudentEmail string           `json:"student_email"`
	Grade        *StudentGradeDTO `json:"grade"`
}

type GradebookDTO struct {
	CourseID       string                `json:"course_id"`
	Students       []GradebookStudentDTO `json:"students"`
	PublishedCount int                   `json:"published_count"`
	GeneratedAt    time.Time             `json:"generated_at"`
}

type UpdateCourseGradeRequest struct {
	OverrideGrade *float64 `json:"override_grade"`
	OverrideNote  *string  `json:"override_note"`
	ClearOverride bool     `json:"clear_override"`
}

type SetGradebookPublicationRequest struct {
	Published  bool     `json:"published"`
	StudentIDs []string `json:"student_ids"`
}

type GradebookPublicationResultDTO struct {
	Published     bool       `json:"published"`
	AffectedCount int        `json:"affected_count"`
	PublishedAt   *time.Time `json:"published_at"`
}

type DeadlineItemDTO struct {
	ItemID    string    `json:"item_id"`
	ItemType  string    `json:"item_type"`
	Title     string    `json:"title"`
	DueDate   time.Time `json:"due_date"`
	Submitted bool      `json:"submitted"`
}
