package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	domainrepos "github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/infrastructure/database"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMExamAttemptRepository struct {
	db *gorm.DB
}

func NewGORMExamAttemptRepository(db *gorm.DB) *GORMExamAttemptRepository {
	return &GORMExamAttemptRepository{db: db}
}

func (r *GORMExamAttemptRepository) Create(ctx context.Context, attempt *entities.ExamAttempt) error {
	model, err := toExamAttemptModel(attempt)
	if err != nil {
		return apperrors.InternalError(err.Error())
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// StartOrResume is the database boundary for attempt admission. PostgreSQL's
// transaction-scoped advisory lock serialises starts for one (exam, student)
// pair, while the partial unique index added by migrations is the durable
// backstop against more than one in-progress attempt.
func (r *GORMExamAttemptRepository) StartOrResume(ctx context.Context, candidate *entities.ExamAttempt, allowedAttempts int, allowNewAttempt bool, now time.Time, expectedExamUpdatedAt *time.Time) (*entities.ExamAttempt, error) {
	var result *entities.ExamAttempt
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockExamConfiguration(tx, candidate.ExamID); err != nil {
			return apperrors.InternalError(err.Error())
		}
		if tx.Dialector.Name() == "postgres" {
			lockKey := fmt.Sprintf("%s:%s", candidate.ExamID, candidate.StudentID)
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
				return apperrors.InternalError(err.Error())
			}
		}
		if expectedExamUpdatedAt != nil {
			var currentExam database.ExamModel
			err := tx.Select("updated_at").First(&currentExam, "id = ? AND org_id = ?", candidate.ExamID.String(), candidate.OrgID.String()).Error
			if err == gorm.ErrRecordNotFound {
				return apperrors.NotFoundError("exam", candidate.ExamID.String())
			}
			if err != nil {
				return apperrors.InternalError(err.Error())
			}
			if !currentExam.UpdatedAt.Equal(*expectedExamUpdatedAt) {
				return domainrepos.ErrExamConfigurationChanged
			}
		}

		// The same one-minute server grace period is used by Start, draft save,
		// and submit. An attempt is only retired after that shared deadline.
		if err := tx.Model(&database.ExamAttemptModel{}).
			Where("exam_id = ? AND student_id = ? AND status = ? AND expires_at < ?", candidate.ExamID.String(), candidate.StudentID.String(), "in_progress", now.Add(-time.Minute)).
			Updates(map[string]interface{}{"status": "expired", "updated_at": now}).Error; err != nil {
			return apperrors.InternalError(err.Error())
		}

		var active database.ExamAttemptModel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("exam_id = ? AND student_id = ? AND status = ?", candidate.ExamID.String(), candidate.StudentID.String(), "in_progress").
			First(&active).Error
		if err == nil {
			var convertErr error
			result, convertErr = toExamAttemptEntity(&active)
			return convertErr
		}
		if err != gorm.ErrRecordNotFound {
			return apperrors.InternalError(err.Error())
		}
		if !allowNewAttempt {
			return apperrors.ValidationError("exam is no longer available for new attempts")
		}

		var count int64
		if err := tx.Model(&database.ExamAttemptModel{}).
			Where("exam_id = ? AND student_id = ?", candidate.ExamID.String(), candidate.StudentID.String()).
			Count(&count).Error; err != nil {
			return apperrors.InternalError(err.Error())
		}
		if int(count) >= allowedAttempts {
			return apperrors.ValidationError("maximum attempts reached")
		}

		model, err := toExamAttemptModel(candidate)
		if err != nil {
			return apperrors.InternalError(err.Error())
		}
		if err := tx.Create(model).Error; err != nil {
			return apperrors.InternalError(err.Error())
		}
		result = candidate
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *GORMExamAttemptRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.ExamAttempt, error) {
	var model database.ExamAttemptModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("exam attempt", id.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toExamAttemptEntity(&model)
}

func (r *GORMExamAttemptRepository) FindByExamAndStudent(ctx context.Context, examID, studentID uuid.UUID) ([]*entities.ExamAttempt, error) {
	var models []database.ExamAttemptModel
	err := r.db.WithContext(ctx).
		Where("exam_id = ? AND student_id = ?", examID.String(), studentID.String()).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]*entities.ExamAttempt, len(models))
	for i := range models {
		e, err := toExamAttemptEntity(&models[i])
		if err != nil {
			return nil, err
		}
		result[i] = e
	}
	return result, nil
}

func (r *GORMExamAttemptRepository) FindByExam(ctx context.Context, examID uuid.UUID) ([]*entities.ExamAttempt, error) {
	var models []database.ExamAttemptModel
	err := r.db.WithContext(ctx).
		Where("exam_id = ?", examID.String()).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]*entities.ExamAttempt, len(models))
	for i := range models {
		e, err := toExamAttemptEntity(&models[i])
		if err != nil {
			return nil, err
		}
		result[i] = e
	}
	return result, nil
}

func (r *GORMExamAttemptRepository) FindActiveAttempt(ctx context.Context, examID, studentID uuid.UUID) (*entities.ExamAttempt, error) {
	var model database.ExamAttemptModel
	err := r.db.WithContext(ctx).
		First(&model, "exam_id = ? AND student_id = ? AND status = ?", examID.String(), studentID.String(), "in_progress").Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toExamAttemptEntity(&model)
}

func (r *GORMExamAttemptRepository) HasAttempts(ctx context.Context, examID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&database.ExamAttemptModel{}).Where("exam_id = ?", examID.String()).Count(&count).Error; err != nil {
		return false, apperrors.InternalError(err.Error())
	}
	return count > 0, nil
}

// SaveDraftIfInProgress only changes answer data. In particular, a stale
// autosave cannot resurrect or overwrite a submitted attempt.
func (r *GORMExamAttemptRepository) SaveDraftIfInProgress(ctx context.Context, attemptID, studentID uuid.UUID, answers []entities.ExamMCQAnswer, updatedAt time.Time) (bool, error) {
	encoded, err := json.Marshal(answers)
	if err != nil {
		return false, apperrors.InternalError(err.Error())
	}
	result := r.db.WithContext(ctx).Model(&database.ExamAttemptModel{}).
		Where("id = ? AND student_id = ? AND status = ? AND expires_at >= ?", attemptID.String(), studentID.String(), "in_progress", updatedAt.Add(-time.Minute)).
		Updates(map[string]interface{}{"mcq_answers": string(encoded), "updated_at": updatedAt})
	if result.Error != nil {
		return false, apperrors.InternalError(result.Error.Error())
	}
	return result.RowsAffected == 1, nil
}

// SubmitIfInProgress is a compare-and-set transition. Exactly one concurrent
// submit can change the row; later callers re-read the submitted record and
// receive the same result idempotently.
func (r *GORMExamAttemptRepository) SubmitIfInProgress(ctx context.Context, attempt *entities.ExamAttempt) (bool, error) {
	model, err := toExamAttemptModel(attempt)
	if err != nil {
		return false, apperrors.InternalError(err.Error())
	}
	result := r.db.WithContext(ctx).Model(&database.ExamAttemptModel{}).
		Where("id = ? AND student_id = ? AND status = ? AND expires_at >= ?", attempt.ID.String(), attempt.StudentID.String(), "in_progress", attempt.UpdatedAt.Add(-time.Minute)).
		Updates(map[string]interface{}{
			"status":        model.Status,
			"submitted_at":  model.SubmittedAt,
			"mcq_answers":   model.MCQAnswers,
			"mcq_score":     model.MCQScore,
			"mcq_max_score": model.MCQMaxScore,
			"file_path":     model.FilePath,
			"file_points":   model.FilePoints,
			"total_score":   model.TotalScore,
			"updated_at":    model.UpdatedAt,
		})
	if result.Error != nil {
		return false, apperrors.InternalError(result.Error.Error())
	}
	return result.RowsAffected == 1, nil
}

func (r *GORMExamAttemptRepository) ExpireIfInProgress(ctx context.Context, attemptID, studentID uuid.UUID, updatedAt time.Time) (bool, error) {
	result := r.db.WithContext(ctx).Model(&database.ExamAttemptModel{}).
		Where("id = ? AND student_id = ? AND status = ? AND expires_at < ?", attemptID.String(), studentID.String(), "in_progress", updatedAt.Add(-time.Minute)).
		Updates(map[string]interface{}{"status": "expired", "updated_at": updatedAt})
	if result.Error != nil {
		return false, apperrors.InternalError(result.Error.Error())
	}
	return result.RowsAffected == 1, nil
}

func (r *GORMExamAttemptRepository) Update(ctx context.Context, attempt *entities.ExamAttempt) error {
	model, err := toExamAttemptModel(attempt)
	if err != nil {
		return apperrors.InternalError(err.Error())
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *GORMExamAttemptRepository) CountByExamAndStudent(ctx context.Context, examID, studentID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&database.ExamAttemptModel{}).
		Where("exam_id = ? AND student_id = ?", examID.String(), studentID.String()).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.InternalError(err.Error())
	}
	return int(count), nil
}

func toExamAttemptModel(a *entities.ExamAttempt) (*database.ExamAttemptModel, error) {
	answersJSON, err := json.Marshal(a.MCQAnswers)
	if err != nil {
		return nil, err
	}
	var gradedBy *string
	if a.GradedBy != nil {
		s := a.GradedBy.String()
		gradedBy = &s
	}
	return &database.ExamAttemptModel{
		ID:           a.ID.String(),
		ExamID:       a.ExamID.String(),
		StudentID:    a.StudentID.String(),
		OrgID:        a.OrgID.String(),
		Status:       a.Status,
		StartedAt:    a.StartedAt,
		ExpiresAt:    a.ExpiresAt,
		SubmittedAt:  a.SubmittedAt,
		MCQAnswers:   string(answersJSON),
		MCQScore:     a.MCQScore,
		MCQMaxScore:  a.MCQMaxScore,
		FilePath:     a.FilePath,
		FileFeedback: a.FileFeedback,
		FileScore:    a.FileScore,
		FilePoints:   a.FilePoints,
		TotalScore:   a.TotalScore,
		GradedBy:     gradedBy,
		GradedAt:     a.GradedAt,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}, nil
}

func toExamAttemptEntity(m *database.ExamAttemptModel) (*entities.ExamAttempt, error) {
	id, _ := uuid.Parse(m.ID)
	examID, _ := uuid.Parse(m.ExamID)
	studentID, _ := uuid.Parse(m.StudentID)
	orgID, _ := uuid.Parse(m.OrgID)

	var answers []entities.ExamMCQAnswer
	if m.MCQAnswers != "" {
		if err := json.Unmarshal([]byte(m.MCQAnswers), &answers); err != nil {
			return nil, apperrors.InternalError("failed to parse mcq answers: " + err.Error())
		}
	}

	var gradedBy *uuid.UUID
	if m.GradedBy != nil {
		parsed, err := uuid.Parse(*m.GradedBy)
		if err == nil {
			gradedBy = &parsed
		}
	}

	return &entities.ExamAttempt{
		ID:           id,
		ExamID:       examID,
		StudentID:    studentID,
		OrgID:        orgID,
		Status:       m.Status,
		StartedAt:    m.StartedAt,
		ExpiresAt:    m.ExpiresAt,
		SubmittedAt:  m.SubmittedAt,
		MCQAnswers:   answers,
		MCQScore:     m.MCQScore,
		MCQMaxScore:  m.MCQMaxScore,
		FilePath:     m.FilePath,
		FileFeedback: m.FileFeedback,
		FileScore:    m.FileScore,
		FilePoints:   m.FilePoints,
		TotalScore:   m.TotalScore,
		GradedBy:     gradedBy,
		GradedAt:     m.GradedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}
