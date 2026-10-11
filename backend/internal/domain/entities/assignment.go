package entities

import (
	"time"

	"github.com/google/uuid"
)

type Assignment struct {
	ID                  uuid.UUID
	OrgID               uuid.UUID
	CourseID            uuid.UUID
	ModuleID            uuid.UUID
	Title               string
	Description         string
	Rubric              string
	MaxPoints           int
	DueDate             *time.Time
	AllowLateSubmission bool
	Position            int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type AssignmentSubmission struct {
	ID           uuid.UUID
	AssignmentID uuid.UUID
	StudentID    uuid.UUID
	OrgID        uuid.UUID
	TextContent  string
	LinkURL      string
	FilePath     string
	Score        *int
	Feedback     string
	GradedBy     *uuid.UUID
	GradedAt     *time.Time
	SubmittedAt  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
