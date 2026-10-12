package repositories

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	ListAll(ctx context.Context, page, pageSize int) ([]entities.User, int64, error)
	ListForSystemAdmin(ctx context.Context, filters SystemUserFilters) ([]SystemUser, int64, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]entities.User, error)
}

// SystemUserFilters scopes the global user directory without exposing raw SQL
// fields to the HTTP layer.
type SystemUserFilters struct {
	Page      int
	PageSize  int
	Search    string
	Role      string
	OrgID     string
	SortBy    string
	SortOrder string
}

type SystemUser struct {
	ID        uuid.UUID
	Email     string
	Name      string
	Role      string
	OrgID     string
	OrgName   string
	CreatedAt time.Time
}
