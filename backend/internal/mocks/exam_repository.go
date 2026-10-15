package mocks

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockExamRepository struct {
	mock.Mock
}

func (m *MockExamRepository) Create(ctx context.Context, exam *entities.Exam) error {
	args := m.Called(ctx, exam)
	return args.Error(0)
}

func (m *MockExamRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*entities.Exam, error) {
	args := m.Called(ctx, id, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Exam), args.Error(1)
}

func (m *MockExamRepository) FindByCourse(ctx context.Context, courseID, orgID uuid.UUID) ([]*entities.Exam, error) {
	args := m.Called(ctx, courseID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Exam), args.Error(1)
}

func (m *MockExamRepository) Update(ctx context.Context, exam *entities.Exam) error {
	args := m.Called(ctx, exam)
	return args.Error(0)
}

func (m *MockExamRepository) UpdateWithAttemptGuard(ctx context.Context, exam *entities.Exam, rejectIfAttempts bool) (bool, error) {
	args := m.Called(ctx, exam, rejectIfAttempts)
	return args.Bool(0), args.Error(1)
}

func (m *MockExamRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	args := m.Called(ctx, id, orgID)
	return args.Error(0)
}
