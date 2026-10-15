package entities

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceStatus string

const (
	AttendancePresent AttendanceStatus = "present"
	AttendanceLate    AttendanceStatus = "late"
	AttendanceAbsent  AttendanceStatus = "absent"
)

type AttendanceRecord struct {
	ID        uuid.UUID
	GroupID   uuid.UUID
	StudentID uuid.UUID
	OrgID     uuid.UUID
	Date      time.Time
	Status    AttendanceStatus
	MarkedBy  uuid.UUID
	MarkedAt  time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
