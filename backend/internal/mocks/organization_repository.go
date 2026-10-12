package mocks

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) Create(ctx context.Context, org *entities.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationRepository) Update(ctx context.Context, org *entities.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) FindBySlug(ctx context.Context, slug string) (*entities.Organization, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) ListAll(ctx context.Context, page, pageSize int) ([]entities.Organization, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]entities.Organization), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
