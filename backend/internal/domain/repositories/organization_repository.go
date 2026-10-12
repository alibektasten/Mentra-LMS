package repositories

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
)

type OrganizationRepository interface {
	Create(ctx context.Context, org *entities.Organization) error
	Update(ctx context.Context, org *entities.Organization) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error)
	FindBySlug(ctx context.Context, slug string) (*entities.Organization, error)
	ListAll(ctx context.Context, page, pageSize int) ([]entities.Organization, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
