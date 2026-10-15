package repositories

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
)

type AttendanceRepository interface {
	ListByGroupAndDate(ctx context.Context, groupID uuid.UUID, date time.Time) ([]entities.AttendanceRecord, error)
	UpsertMany(ctx context.Context, records []entities.AttendanceRecord) error
}
