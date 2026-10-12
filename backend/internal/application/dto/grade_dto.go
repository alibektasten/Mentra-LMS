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
	StudentID     string         `json:"student_id"`
	Items         []GradeItemDTO `json:"items"`
	TotalEarned   int            `json:"total_earned"`
	TotalPossible int            `json:"total_possible"`
	Percentage    float64        `json:"percentage"`
}

type DeadlineItemDTO struct {
	ItemID    string    `json:"item_id"`
	ItemType  string    `json:"item_type"`
	Title     string    `json:"title"`
	DueDate   time.Time `json:"due_date"`
	Submitted bool      `json:"submitted"`
}
