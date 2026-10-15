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

type GORMProgressRepository struct {
	db *gorm.DB
}

func NewGORMProgressRepository(db *gorm.DB) *GORMProgressRepository {
	return &GORMProgressRepository{db: db}
}

func (r *GORMProgressRepository) Create(ctx context.Context, progress *entities.LessonProgress) error {
	return r.db.WithContext(ctx).Create(toProgressModel(progress)).Error
}

func (r *GORMProgressRepository) FindByUserAndLesson(ctx context.Context, userID, lessonID, orgID uuid.UUID) (*entities.LessonProgress, error) {
	var model database.LessonProgressModel
	err := r.db.WithContext(ctx).First(&model, "user_id = ? AND lesson_id = ? AND org_id = ?",
		userID.String(), lessonID.String(), orgID.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("progress", userID.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toProgressEntity(&model), nil
}

func (r *GORMProgressRepository) FindByUser(ctx context.Context, userID, orgID uuid.UUID) ([]entities.LessonProgress, error) {
	var models []database.LessonProgressModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND org_id = ?", userID.String(), orgID.String()).
		Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]entities.LessonProgress, len(models))
	for i, m := range models {
		result[i] = *toProgressEntity(&m)
	}
	return result, nil
}

func (r *GORMProgressRepository) FindByUserAndCourse(ctx context.Context, userID, courseID, orgID uuid.UUID) ([]entities.LessonProgress, error) {
	var models []database.LessonProgressModel
	err := r.db.WithContext(ctx).
		Table("lesson_progress AS progress").
		Select("progress.*").
		Joins("JOIN lessons AS lesson ON lesson.id = progress.lesson_id AND lesson.org_id = progress.org_id").
		Joins("JOIN modules AS module ON module.id = lesson.module_id AND module.org_id = lesson.org_id").
		Joins("JOIN courses AS course ON course.id = module.course_id AND course.org_id = module.org_id").
		Where("progress.user_id = ? AND progress.org_id = ? AND course.id = ?", userID.String(), orgID.String(), courseID.String()).
		Order("progress.created_at DESC").Find(&models).Error
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]entities.LessonProgress, len(models))
	for i, m := range models {
		result[i] = *toProgressEntity(&m)
	}
	return result, nil
}

func (r *GORMProgressRepository) Update(ctx context.Context, progress *entities.LessonProgress) error {
	progress.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toProgressModel(progress)).Error
}

func toProgressModel(p *entities.LessonProgress) *database.LessonProgressModel {
	return &database.LessonProgressModel{
		ID:          p.ID.String(),
		UserID:      p.UserID.String(),
		LessonID:    p.LessonID.String(),
		OrgID:       p.OrgID.String(),
		CompletedAt: p.CompletedAt,
		Score:       p.Score,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toProgressEntity(m *database.LessonProgressModel) *entities.LessonProgress {
	id, _ := uuid.Parse(m.ID)
	userID, _ := uuid.Parse(m.UserID)
	lessonID, _ := uuid.Parse(m.LessonID)
	orgID, _ := uuid.Parse(m.OrgID)
	return &entities.LessonProgress{
		ID:          id,
		UserID:      userID,
		LessonID:    lessonID,
		OrgID:       orgID,
		CompletedAt: m.CompletedAt,
		Score:       m.Score,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
