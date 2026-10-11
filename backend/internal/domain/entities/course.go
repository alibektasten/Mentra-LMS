package entities

import (
	"time"

	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
)

type CourseStatus string

const (
	StatusDraft     CourseStatus = "draft"
	StatusPublished CourseStatus = "published"
	StatusArchived  CourseStatus = "archived"
)

type Course struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	Title       string
	Description string
	Status      CourseStatus
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c *Course) Publish() error {
	if c.Status == StatusPublished {
		return apperrors.ConflictError("course is already published")
	}
	c.Status = StatusPublished
	return nil
}
