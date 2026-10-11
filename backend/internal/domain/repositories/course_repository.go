package repositories

import (
	"context"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/google/uuid"
)

type CourseRepository interface {
	Create(ctx context.Context, course *entities.Course) error
	FindByID(ctx context.Context, id, orgID uuid.UUID) (*entities.Course, error)
	FindByOrg(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]entities.Course, int64, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]entities.Course, int64, error)
	FindPublishedByIDs(ctx context.Context, ids []uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]entities.Course, int64, error)
	Update(ctx context.Context, course *entities.Course) error
	Delete(ctx context.Context, id, orgID uuid.UUID) error
}
