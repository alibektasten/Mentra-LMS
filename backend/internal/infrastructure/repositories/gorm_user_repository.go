package repositories

import (
	"context"
	"strings"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	domainrepos "github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/infrastructure/database"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GORMUserRepository struct {
	db *gorm.DB
}

func NewGORMUserRepository(db *gorm.DB) *GORMUserRepository {
	return &GORMUserRepository{db: db}
}

func (r *GORMUserRepository) Create(ctx context.Context, user *entities.User) error {
	model := &database.UserModel{
		ID:           user.ID.String(),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		AvatarURL:    user.AvatarURL,
		Role:         user.Role,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GORMUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var model database.UserModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("user", id.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toUserEntity(&model), nil
}

func (r *GORMUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	var model database.UserModel
	err := r.db.WithContext(ctx).First(&model, "email = ?", email).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("user", email)
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toUserEntity(&model), nil
}

func (r *GORMUserRepository) Update(ctx context.Context, user *entities.User) error {
	user.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(&database.UserModel{
		ID:           user.ID.String(),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		AvatarURL:    user.AvatarURL,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}).Error
}

func (r *GORMUserRepository) ListAll(ctx context.Context, page, pageSize int) ([]entities.User, int64, error) {
	var models []database.UserModel
	var count int64
	if err := r.db.WithContext(ctx).Model(&database.UserModel{}).Count(&count).Error; err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}
	if err := r.db.WithContext(ctx).Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}
	result := make([]entities.User, len(models))
	for i, m := range models {
		result[i] = *toUserEntity(&m)
	}
	return result, count, nil
}

func (r *GORMUserRepository) ListForSystemAdmin(ctx context.Context, filters domainrepos.SystemUserFilters) ([]domainrepos.SystemUser, int64, error) {
	query := r.db.WithContext(ctx).
		Table("users").
		Select(`COALESCE(memberships.id::text, '') AS membership_id, users.id, users.email, users.name, COALESCE(NULLIF(memberships.role, ''), NULLIF(users.role, ''), 'unassigned') AS role,
			COALESCE(memberships.org_id::text, '') AS org_id, COALESCE(organizations.name, 'System') AS org_name, users.created_at`).
		Joins("LEFT JOIN memberships ON memberships.user_id = users.id").
		Joins("LEFT JOIN organizations ON organizations.id = memberships.org_id")

	if search := strings.TrimSpace(strings.ToLower(filters.Search)); search != "" {
		like := "%" + search + "%"
		query = query.Where("LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ?", like, like)
	}
	if filters.Role != "" {
		query = query.Where("COALESCE(NULLIF(memberships.role, ''), NULLIF(users.role, ''), 'unassigned') = ?", filters.Role)
	}
	if filters.OrgID != "" {
		query = query.Where("memberships.org_id = ?", filters.OrgID)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}

	orderColumns := map[string]string{
		"name": "users.name", "email": "users.email", "role": "role", "org_name": "organizations.name", "created_at": "users.created_at",
	}
	orderBy, ok := orderColumns[filters.SortBy]
	if !ok {
		orderBy = "users.created_at"
	}
	direction := "DESC"
	if strings.EqualFold(filters.SortOrder, "asc") {
		direction = "ASC"
	}

	var rows []domainrepos.SystemUser
	if err := query.Order(orderBy + " " + direction).Order("users.id ASC").
		Offset((filters.Page - 1) * filters.PageSize).Limit(filters.PageSize).Scan(&rows).Error; err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}
	return rows, total, nil
}

func (r *GORMUserRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]entities.User, error) {
	if len(ids) == 0 {
		return []entities.User{}, nil
	}
	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = id.String()
	}
	var models []database.UserModel
	if err := r.db.WithContext(ctx).Where("id IN ?", strIDs).Find(&models).Error; err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]entities.User, len(models))
	for i, m := range models {
		result[i] = *toUserEntity(&m)
	}
	return result, nil
}

func toUserEntity(m *database.UserModel) *entities.User {
	id, _ := uuid.Parse(m.ID)
	return &entities.User{
		ID:           id,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Name:         m.Name,
		AvatarURL:    m.AvatarURL,
		Role:         m.Role,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
