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
	StudentID          string         `json:"student_id"`
	Items              []GradeItemDTO `json:"items"`
	TotalEarned        int            `json:"total_earned"`
	TotalPossible      int            `json:"total_possible"`
	Percentage         float64        `json:"percentage"`
	AssignmentsPercent *float64       `json:"assignments_percent"`
	MidtermPercent     *float64       `json:"midterm_percent"`
	EndtermPercent     *float64       `json:"endterm_percent"`
	FinalExamPercent   *float64       `json:"final_exam_percent"`
	MidtermRegister    *float64       `json:"midterm_register"`
	EndtermRegister    *float64       `json:"endterm_register"`
	FinalGrade         float64        `json:"final_grade"`
	IsFinalGradeReady  bool           `json:"is_final_grade_ready"`
}

type DeadlineItemDTO struct {
	ItemID    string    `json:"item_id"`
	ItemType  string    `json:"item_type"`
	Title     string    `json:"title"`
	DueDate   time.Time `json:"due_date"`
	Submitted bool      `json:"submitted"`
}
