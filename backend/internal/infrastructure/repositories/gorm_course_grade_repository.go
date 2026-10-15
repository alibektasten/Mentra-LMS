package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/database"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMCourseGradeRepository struct {
	db *gorm.DB
}

func NewGORMCourseGradeRepository(db *gorm.DB) *GORMCourseGradeRepository {
	return &GORMCourseGradeRepository{db: db}
}

// Find returns nil, nil when a teacher has not yet changed or published the
// student's result. That is a normal state, not a missing-resource error.
func (r *GORMCourseGradeRepository) Find(ctx context.Context, courseID, studentID, orgID uuid.UUID) (*entities.CourseGrade, error) {
	var model database.CourseGradeModel
	err := r.db.WithContext(ctx).First(&model,
		"course_id = ? AND student_id = ? AND org_id = ?",
		courseID.String(), studentID.String(), orgID.String(),
	).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toCourseGradeEntity(&model), nil
}

// UpdateOverride serialises an override transition for one student/course
// record. It reads the previous value while holding the row lock, updates only
// override columns (never publication columns), then writes the audit entry in
// the same transaction. The initial insert uses the course/student unique key
// so concurrent first overrides converge on one locked row.
func (r *GORMCourseGradeRepository) UpdateOverride(ctx context.Context, courseID, studentID, orgID, actorID uuid.UUID, overrideGrade *float64, overrideNote *string, clearOverride bool, updatedAt time.Time) (*entities.CourseGrade, error) {
	if overrideGrade != nil && (*overrideGrade < 0 || *overrideGrade > 100) {
		return nil, apperrors.ValidationError("override grade must be between 0 and 100")
	}
	var normalizedNote *string
	if overrideNote != nil {
		note := strings.TrimSpace(*overrideNote)
		if len(note) > 2000 {
			return nil, apperrors.ValidationError("override note must be 2000 characters or fewer")
		}
		normalizedNote = &note
	}

	var result *entities.CourseGrade
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model database.CourseGradeModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&model,
			"course_id = ? AND student_id = ? AND org_id = ?",
			courseID.String(), studentID.String(), orgID.String(),
		).Error
		if err == gorm.ErrRecordNotFound {
			if clearOverride {
				return nil // clearing a missing adjustment is idempotent and writes no audit row
			}
			if overrideGrade == nil {
				return apperrors.ValidationError("an override grade is required for a new adjustment")
			}
			seed := database.CourseGradeModel{
				ID: uuid.NewString(), CourseID: courseID.String(), StudentID: studentID.String(), OrgID: orgID.String(),
				CreatedAt: updatedAt, UpdatedAt: updatedAt,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "course_id"}, {Name: "student_id"}},
				DoNothing: true,
			}).Create(&seed).Error; err != nil {
				return err
			}
			err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&model,
				"course_id = ? AND student_id = ? AND org_id = ?",
				courseID.String(), studentID.String(), orgID.String(),
			).Error
		}
		if err != nil {
			return err
		}

		previous := copyCourseGradeFloat(model.OverrideGrade)
		action := "override_updated"
		reason := ""
		nextGrade := copyCourseGradeFloat(model.OverrideGrade)
		nextNote := model.OverrideNote
		if clearOverride {
			if model.OverrideGrade == nil {
				result = toCourseGradeEntity(&model)
				return nil
			}
			action = "override_cleared"
			reason = "Manual override removed"
			nextGrade = nil
			nextNote = ""
		} else {
			if normalizedNote == nil || *normalizedNote == "" {
				return apperrors.ValidationError("a reason is required for a manual adjustment")
			}
			if overrideGrade != nil {
				nextGrade = copyCourseGradeFloat(overrideGrade)
			}
			if normalizedNote != nil {
				nextNote = *normalizedNote
			}
			if nextGrade == nil {
				return apperrors.ValidationError("an override grade is required for a manual adjustment")
			}
			reason = nextNote
			if previous == nil {
				action = "override_set"
			}
		}

		if err := tx.Model(&database.CourseGradeModel{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
			"override_grade": nextGrade,
			"override_note":  nextNote,
			"updated_at":     updatedAt,
		}).Error; err != nil {
			return err
		}
		model.OverrideGrade = nextGrade
		model.OverrideNote = nextNote
		model.UpdatedAt = updatedAt
		courseGradeID, err := uuid.Parse(model.ID)
		if err != nil {
			return err
		}
		change := &entities.CourseGradeChange{
			ID: uuid.New(), CourseGradeID: courseGradeID, CourseID: courseID, StudentID: studentID, OrgID: orgID,
			ChangedBy: actorID, Action: action, PreviousOverride: previous, NewOverride: copyCourseGradeFloat(nextGrade), Reason: reason, CreatedAt: updatedAt,
		}
		if err := tx.Create(toCourseGradeChangeModel(change)).Error; err != nil {
			return err
		}
		result = toCourseGradeEntity(&model)
		return nil
	})
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, apperrors.InternalError(err.Error())
	}
	return result, nil
}

func copyCourseGradeFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

// SetPublication applies a selected class publication as an all-or-nothing
// transaction. The caller has already validated enrolment and constructed the
// complete selected set.
func (r *GORMCourseGradeRepository) SetPublication(ctx context.Context, grades []*entities.CourseGrade) error {
	if len(grades) == 0 {
		return nil
	}
	models := make([]database.CourseGradeModel, len(grades))
	for i, grade := range grades {
		models[i] = *toCourseGradeModel(grade)
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "course_id"}, {Name: "student_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"published_at":           models[0].PublishedAt,
				"published_by":           models[0].PublishedBy,
				"publication_changed_at": models[0].PublicationChangedAt,
				"publication_changed_by": models[0].PublicationChangedBy,
				"updated_at":             models[0].UpdatedAt,
			}),
		}).Create(&models).Error
	})
	if err != nil {
		return apperrors.InternalError(err.Error())
	}
	return nil
}

func toCourseGradeModel(grade *entities.CourseGrade) *database.CourseGradeModel {
	model := &database.CourseGradeModel{
		ID:                   grade.ID.String(),
		CourseID:             grade.CourseID.String(),
		StudentID:            grade.StudentID.String(),
		OrgID:                grade.OrgID.String(),
		OverrideGrade:        grade.OverrideGrade,
		OverrideNote:         grade.OverrideNote,
		PublishedAt:          grade.PublishedAt,
		PublicationChangedAt: grade.PublicationChangedAt,
		CreatedAt:            grade.CreatedAt,
		UpdatedAt:            grade.UpdatedAt,
	}
	if grade.PublishedBy != nil {
		value := grade.PublishedBy.String()
		model.PublishedBy = &value
	}
	if grade.PublicationChangedBy != nil {
		value := grade.PublicationChangedBy.String()
		model.PublicationChangedBy = &value
	}
	return model
}

func toCourseGradeEntity(model *database.CourseGradeModel) *entities.CourseGrade {
	id, _ := uuid.Parse(model.ID)
	courseID, _ := uuid.Parse(model.CourseID)
	studentID, _ := uuid.Parse(model.StudentID)
	orgID, _ := uuid.Parse(model.OrgID)
	grade := &entities.CourseGrade{
		ID:                   id,
		CourseID:             courseID,
		StudentID:            studentID,
		OrgID:                orgID,
		OverrideGrade:        model.OverrideGrade,
		OverrideNote:         model.OverrideNote,
		PublishedAt:          model.PublishedAt,
		PublicationChangedAt: model.PublicationChangedAt,
		CreatedAt:            model.CreatedAt,
		UpdatedAt:            model.UpdatedAt,
	}
	if model.PublishedBy != nil {
		if value, err := uuid.Parse(*model.PublishedBy); err == nil {
			grade.PublishedBy = &value
		}
	}
	if model.PublicationChangedBy != nil {
		if value, err := uuid.Parse(*model.PublicationChangedBy); err == nil {
			grade.PublicationChangedBy = &value
		}
	}
	return grade
}

func toCourseGradeChangeModel(change *entities.CourseGradeChange) *database.CourseGradeChangeModel {
	return &database.CourseGradeChangeModel{
		ID:               change.ID.String(),
		CourseGradeID:    change.CourseGradeID.String(),
		CourseID:         change.CourseID.String(),
		StudentID:        change.StudentID.String(),
		OrgID:            change.OrgID.String(),
		ChangedBy:        change.ChangedBy.String(),
		Action:           change.Action,
		PreviousOverride: change.PreviousOverride,
		NewOverride:      change.NewOverride,
		Reason:           change.Reason,
		CreatedAt:        change.CreatedAt,
	}
}
