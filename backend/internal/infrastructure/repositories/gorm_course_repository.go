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

type GORMCourseRepository struct {
	db *gorm.DB
}

func NewGORMCourseRepository(db *gorm.DB) *GORMCourseRepository {
	return &GORMCourseRepository{db: db}
}

func (r *GORMCourseRepository) Create(ctx context.Context, course *entities.Course) error {
	model := toCourseModel(course)
	return r.db.WithContext(ctx).Create(model).Error
}

// CreateWithAcademicSetup prevents a partially provisioned course if the
// teacher assignment or one of the required assessments cannot be stored.
func (r *GORMCourseRepository) CreateWithAcademicSetup(ctx context.Context, course *entities.Course, teacher *entities.CourseTeacher, exams []*entities.Exam) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(toCourseModel(course)).Error; err != nil {
			return err
		}
		if err := tx.Create(&database.CourseTeacherModel{
			ID:         teacher.ID.String(),
			CourseID:   teacher.CourseID.String(),
			TeacherID:  teacher.TeacherID.String(),
			OrgID:      teacher.OrgID.String(),
			AssignedAt: teacher.AssignedAt,
		}).Error; err != nil {
			return err
		}
		for _, exam := range exams {
			if err := tx.Create(toExamModel(exam)).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return apperrors.InternalError(err.Error())
	}
	return nil
}

func (r *GORMCourseRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*entities.Course, error) {
	var model database.CourseModel
	err := r.db.WithContext(ctx).First(&model, "id = ? AND org_id = ?", id.String(), orgID.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("course", id.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toCourseEntity(&model), nil
}

func (r *GORMCourseRepository) FindByOrg(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]entities.Course, int64, error) {
	var models []database.CourseModel
	var count int64
	err := r.db.WithContext(ctx).Model(&database.CourseModel{}).Where("org_id = ?", orgID.String()).Count(&count).Error
	if err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}
	err = r.db.WithContext(ctx).Where("org_id = ?", orgID.String()).
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}
	result := make([]entities.Course, len(models))
	for i, m := range models {
		result[i] = *toCourseEntity(&m)
	}
	return result, count, nil
}

func (r *GORMCourseRepository) FindByIDs(ctx context.Context, ids []uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]entities.Course, int64, error) {
	return r.findByIDs(ctx, ids, orgID, page, pageSize, "")
}

func (r *GORMCourseRepository) FindPublishedByIDs(ctx context.Context, ids []uuid.UUID, orgID uuid.UUID, page, pageSize int) ([]entities.Course, int64, error) {
	return r.findByIDs(ctx, ids, orgID, page, pageSize, entities.StatusPublished)
}

func (r *GORMCourseRepository) findByIDs(ctx context.Context, ids []uuid.UUID, orgID uuid.UUID, page, pageSize int, status entities.CourseStatus) ([]entities.Course, int64, error) {
	if len(ids) == 0 {
		return []entities.Course{}, 0, nil
	}
	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = id.String()
	}
	baseQuery := r.db.WithContext(ctx).Model(&database.CourseModel{}).
		Where("id IN ? AND org_id = ?", strIDs, orgID.String())
	if status != "" {
		baseQuery = baseQuery.Where("status = ?", string(status))
	}

	var models []database.CourseModel
	var count int64
	err := baseQuery.Count(&count).Error
	if err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}
	query := r.db.WithContext(ctx).
		Where("id IN ? AND org_id = ?", strIDs, orgID.String())
	if status != "" {
		query = query.Where("status = ?", string(status))
	}
	err = query.
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, 0, apperrors.InternalError(err.Error())
	}
	result := make([]entities.Course, len(models))
	for i, m := range models {
		result[i] = *toCourseEntity(&m)
	}
	return result, count, nil
}

func (r *GORMCourseRepository) Update(ctx context.Context, course *entities.Course) error {
	course.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toCourseModel(course)).Error
}

func (r *GORMCourseRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND org_id = ?", id.String(), orgID.String()).Delete(&database.CourseModel{})
	if result.Error != nil {
		return apperrors.InternalError(result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFoundError("course", id.String())
	}
	return nil
}

func toCourseModel(c *entities.Course) *database.CourseModel {
	return &database.CourseModel{
		ID:          c.ID.String(),
		OrgID:       c.OrgID.String(),
		Title:       c.Title,
		Description: c.Description,
		Status:      string(c.Status),
		CreatedBy:   c.CreatedBy.String(),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toCourseEntity(m *database.CourseModel) *entities.Course {
	id, _ := uuid.Parse(m.ID)
	orgID, _ := uuid.Parse(m.OrgID)
	createdBy, _ := uuid.Parse(m.CreatedBy)
	return &entities.Course{
		ID:          id,
		OrgID:       orgID,
		Title:       m.Title,
		Description: m.Description,
		Status:      entities.CourseStatus(m.Status),
		CreatedBy:   createdBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
