package mocks

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockCourseGradeRepository struct {
	mock.Mock
}

func (m *MockCourseGradeRepository) Find(ctx context.Context, courseID, studentID, orgID uuid.UUID) (*entities.CourseGrade, error) {
	args := m.Called(ctx, courseID, studentID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.CourseGrade), args.Error(1)
}

func (m *MockCourseGradeRepository) UpdateOverride(ctx context.Context, courseID, studentID, orgID, actorID uuid.UUID, overrideGrade *float64, overrideNote *string, clearOverride bool, updatedAt time.Time) (*entities.CourseGrade, error) {
	args := m.Called(ctx, courseID, studentID, orgID, actorID, overrideGrade, overrideNote, clearOverride, updatedAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.CourseGrade), args.Error(1)
}

func (m *MockCourseGradeRepository) SetPublication(ctx context.Context, grades []*entities.CourseGrade) error {
	args := m.Called(ctx, grades)
	return args.Error(0)
}
