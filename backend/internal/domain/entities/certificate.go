package entities

import (
	"time"

	"github.com/google/uuid"
)

// Certificate is issued once per student and course after every lesson is
// completed. VerificationCode is safe to share publicly; internal UUIDs are
// never used as proof of completion.
type Certificate struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	CourseID         uuid.UUID
	OrgID            uuid.UUID
	VerificationCode string
	IssuedAt         time.Time
}
