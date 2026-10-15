package entities

import (
	"time"

	"github.com/google/uuid"
)

// CourseGrade stores teacher-controlled state for the final result. Source
// assessment scores remain in submissions and exam attempts; an override is an
// explicit, auditable adjustment to the calculated result.
type CourseGrade struct {
	ID                   uuid.UUID
	CourseID             uuid.UUID
	StudentID            uuid.UUID
	OrgID                uuid.UUID
	OverrideGrade        *float64
	OverrideNote         string
	PublishedAt          *time.Time
	PublishedBy          *uuid.UUID
	PublicationChangedAt *time.Time
	PublicationChangedBy *uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// CourseGradeChange is an append-only audit record for manual changes to a
// student's final result. Calculated assignment and examination scores remain
// in their source records and are never replaced by an override.
type CourseGradeChange struct {
	ID               uuid.UUID
	CourseGradeID    uuid.UUID
	CourseID         uuid.UUID
	StudentID        uuid.UUID
	OrgID            uuid.UUID
	ChangedBy        uuid.UUID
	Action           string
	PreviousOverride *float64
	NewOverride      *float64
	Reason           string
	CreatedAt        time.Time
}
