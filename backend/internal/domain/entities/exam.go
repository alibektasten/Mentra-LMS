package entities

import (
	"time"

	"github.com/google/uuid"
)

type Exam struct {
	ID              uuid.UUID
	CourseID        uuid.UUID
	OrgID           uuid.UUID
	AssessmentType  string
	Title           string
	Description     string
	DurationMinutes int
	MaxAttempts     int
	DueDate         *time.Time
	MCQEnabled      bool
	MCQPoints       int
	FileEnabled     bool
	FilePoints      int
	Questions       []ExamQuestion
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ExamQuestion struct {
	ID        uuid.UUID
	ExamID    uuid.UUID
	Question  string
	Position  int
	Answers   []ExamAnswer
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ExamAnswer struct {
	ID         uuid.UUID
	QuestionID uuid.UUID
	Answer     string
	IsCorrect  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
