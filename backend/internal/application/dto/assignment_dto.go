package dto

import "time"

type CreateAssignmentRequest struct {
	Title               string     `json:"title" binding:"required"`
	Description         string     `json:"description"`
	Rubric              string     `json:"rubric"`
	MaxPoints           int        `json:"max_points" binding:"required,min=1"`
	DueDate             *time.Time `json:"due_date"`
	AllowLateSubmission bool       `json:"allow_late_submission"`
}

type UpdateAssignmentRequest struct {
	Title               *string    `json:"title"`
	Description         *string    `json:"description"`
	Rubric              *string    `json:"rubric"`
	MaxPoints           *int       `json:"max_points"`
	DueDate             *time.Time `json:"due_date"`
	AllowLateSubmission *bool      `json:"allow_late_submission"`
}

type AssignmentDTO struct {
	ID                  string     `json:"id"`
	CourseID            string     `json:"course_id"`
	ModuleID            string     `json:"module_id"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	Rubric              string     `json:"rubric"`
	MaxPoints           int        `json:"max_points"`
	DueDate             *time.Time `json:"due_date"`
	AllowLateSubmission bool       `json:"allow_late_submission"`
	Position            int        `json:"position"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type GradeSubmissionRequest struct {
	Score    int    `json:"score" binding:"gte=0"`
	Feedback string `json:"feedback"`
}

type SubmissionDTO struct {
	ID           string     `json:"id"`
	AssignmentID string     `json:"assignment_id"`
	StudentID    string     `json:"student_id"`
	TextContent  string     `json:"text_content"`
	LinkURL      string     `json:"link_url"`
	FilePath     string     `json:"file_path"`
	Score        *int       `json:"score"`
	Feedback     string     `json:"feedback"`
	GradedAt     *time.Time `json:"graded_at"`
	SubmittedAt  time.Time  `json:"submitted_at"`
}
