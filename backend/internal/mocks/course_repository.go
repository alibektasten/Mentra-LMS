package mocks

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockCourseRepository struct {
	mock.Mock
}

func (m *MockCourseRepository) Create(ctx context.Context, course *entities.Course) error {
	args := m.Called(ctx, course)
	return args.Error(0)
}

func (m *MockCourseRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*entities.Course, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Course), args.Error(1)
}

func (m *MockCourseRepository) FindByOrg(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]entities.Course, int64, error) {
	args := m.Called(ctx, orgID, page, pageSize)
	return args.Get(0).([]entities.Course), args.Get(1).(int64), args.Error(2)
}

func (m *MockCourseRepository) FindByIDs(ctx context.Context, ids []uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]entities.Course, int64, error) {
	args := m.Called(ctx, ids, orgID, page, pageSize)
	return args.Get(0).([]entities.Course), args.Get(1).(int64), args.Error(2)
}

func (m *MockCourseRepository) FindPublishedByIDs(ctx context.Context, ids []uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]entities.Course, int64, error) {
	args := m.Called(ctx, ids, orgID, page, pageSize)
	return args.Get(0).([]entities.Course), args.Get(1).(int64), args.Error(2)
}

func (m *MockCourseRepository) Update(ctx context.Context, course *entities.Course) error {
	args := m.Called(ctx, course)
	return args.Error(0)
}

func (m *MockCourseRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	args := m.Called(ctx, id, orgID)
	return args.Error(0)
}
