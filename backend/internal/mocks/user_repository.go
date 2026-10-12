package mocks

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) ListAll(ctx context.Context, page, pageSize int) ([]entities.User, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]entities.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) ListForSystemAdmin(ctx context.Context, filters repositories.SystemUserFilters) ([]repositories.SystemUser, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]repositories.SystemUser), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]entities.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.User), args.Error(1)
}
