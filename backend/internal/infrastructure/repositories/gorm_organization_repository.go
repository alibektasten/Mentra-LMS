package repositories

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/database"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMOrganizationRepository struct {
	db *gorm.DB
}

func NewGORMOrganizationRepository(db *gorm.DB) *GORMOrganizationRepository {
	return &GORMOrganizationRepository{db: db}
}

func (r *GORMOrganizationRepository) Create(ctx context.Context, org *entities.Organization) error {
	model := &database.OrganizationModel{
		ID:        org.ID.String(),
		Name:      org.Name,
		Slug:      org.Slug,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GORMOrganizationRepository) Update(ctx context.Context, org *entities.Organization) error {
	result := r.db.WithContext(ctx).Model(&database.OrganizationModel{}).Where("id = ?", org.ID.String()).Updates(map[string]any{"name": org.Name, "slug": org.Slug, "updated_at": time.Now()})
	if result.Error != nil {
		return apperrors.InternalError(result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFoundError("organization", org.ID.String())
	}
	return nil
}

func (r *GORMOrganizationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	var model database.OrganizationModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("organization", id.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toOrgEntity(&model), nil
}

func (r *GORMOrganizationRepository) FindBySlug(ctx context.Context, slug string) (*entities.Organization, error) {
	var model database.OrganizationModel
	err := r.db.WithContext(ctx).First(&model, "slug = ?", slug).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("organization", slug)
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toOrgEntity(&model), nil
}

func (r *GORMOrganizationRepository) ListAll(ctx context.Context, page, pageSize int) ([]entities.Organization, int64, error) {
	var models []database.OrganizationModel
	var count int64
	if err := r.db.WithContext(ctx).Model(&database.OrganizationModel{}).Count(&count).Error; err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}
	if err := r.db.WithContext(ctx).Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}
	result := make([]entities.Organization, len(models))
	for i, m := range models {
		result[i] = *toOrgEntity(&m)
	}
	return result, count, nil
}

func (r *GORMOrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&database.OrganizationModel{}, "id = ?", id.String())
	if res.Error != nil {
		return apperrors.InternalError(res.Error.Error())
	}
	if res.RowsAffected == 0 {
		return apperrors.NotFoundError("organization", id.String())
	}
	return nil
}

func toOrgEntity(m *database.OrganizationModel) *entities.Organization {
	id, _ := uuid.Parse(m.ID)
	return &entities.Organization{
		ID:        id,
		Name:      m.Name,
		Slug:      m.Slug,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
