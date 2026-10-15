package mocks

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockLessonProgressRepository struct {
	mock.Mock
}

func (m *MockLessonProgressRepository) Create(ctx context.Context, progress *entities.LessonProgress) error {
	args := m.Called(ctx, progress)
	return args.Error(0)
}

func (m *MockLessonProgressRepository) FindByUserAndLesson(ctx context.Context, userID, lessonID, orgID uuid.UUID) (*entities.LessonProgress, error) {
	args := m.Called(ctx, userID, lessonID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.LessonProgress), args.Error(1)
}

func (m *MockLessonProgressRepository) FindByUser(ctx context.Context, userID, orgID uuid.UUID) ([]entities.LessonProgress, error) {
	args := m.Called(ctx, userID, orgID)
	return args.Get(0).([]entities.LessonProgress), args.Error(1)
}

func (m *MockLessonProgressRepository) FindByUserAndCourse(ctx context.Context, userID, courseID, orgID uuid.UUID) ([]entities.LessonProgress, error) {
	args := m.Called(ctx, userID, courseID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.LessonProgress), args.Error(1)
}

func (m *MockLessonProgressRepository) Update(ctx context.Context, progress *entities.LessonProgress) error {
	args := m.Called(ctx, progress)
	return args.Error(0)
}
