package mocks

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockAttendanceRepository struct{ mock.Mock }

func (m *MockAttendanceRepository) ListByGroupAndDate(ctx context.Context, groupID uuid.UUID, date time.Time) ([]entities.AttendanceRecord, error) {
	args := m.Called(ctx, groupID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.AttendanceRecord), args.Error(1)
}

func (m *MockAttendanceRepository) UpsertMany(ctx context.Context, records []entities.AttendanceRecord) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}
