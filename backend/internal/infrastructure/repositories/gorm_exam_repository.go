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

type GORMExamRepository struct {
	db *gorm.DB
}

func NewGORMExamRepository(db *gorm.DB) *GORMExamRepository {
	return &GORMExamRepository{db: db}
}

func (r *GORMExamRepository) Create(ctx context.Context, exam *entities.Exam) error {
	model := toExamModel(exam)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GORMExamRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*entities.Exam, error) {
	var model database.ExamModel
	err := r.db.WithContext(ctx).
		Preload("Questions.Answers").
		First(&model, "id = ? AND org_id = ?", id.String(), orgID.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("exam", id.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toExamEntity(&model), nil
}

func (r *GORMExamRepository) FindByCourse(ctx context.Context, courseID, orgID uuid.UUID) ([]*entities.Exam, error) {
	var models []database.ExamModel
	err := r.db.WithContext(ctx).
		Preload("Questions.Answers").
		Where("course_id = ? AND org_id = ?", courseID.String(), orgID.String()).
		Order("created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]*entities.Exam, len(models))
	for i := range models {
		result[i] = toExamEntity(&models[i])
	}
	return result, nil
}

func (r *GORMExamRepository) Update(ctx context.Context, exam *entities.Exam) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var questionIDs []string
		if err := tx.Model(&database.ExamQuestionModel{}).
			Where("exam_id = ?", exam.ID.String()).
			Pluck("id", &questionIDs).Error; err != nil {
			return apperrors.InternalError(err.Error())
		}

		if len(questionIDs) > 0 {
			if err := tx.Where("question_id IN ?", questionIDs).
				Delete(&database.ExamAnswerModel{}).Error; err != nil {
				return apperrors.InternalError(err.Error())
			}
		}

		if err := tx.Where("exam_id = ?", exam.ID.String()).
			Delete(&database.ExamQuestionModel{}).Error; err != nil {
			return apperrors.InternalError(err.Error())
		}

		exam.UpdatedAt = time.Now()
		examRecord := &database.ExamModel{
			ID:              exam.ID.String(),
			CourseID:        exam.CourseID.String(),
			OrgID:           exam.OrgID.String(),
			Title:           exam.Title,
			Description:     exam.Description,
			DurationMinutes: exam.DurationMinutes,
			MaxAttempts:     exam.MaxAttempts,
			DueDate:         exam.DueDate,
			MCQEnabled:      exam.MCQEnabled,
			MCQPoints:       exam.MCQPoints,
			FileEnabled:     exam.FileEnabled,
			FilePoints:      exam.FilePoints,
			CreatedAt:       exam.CreatedAt,
			UpdatedAt:       exam.UpdatedAt,
		}
		if err := tx.Save(examRecord).Error; err != nil {
			return apperrors.InternalError(err.Error())
		}

		for _, q := range exam.Questions {
			qModel := database.ExamQuestionModel{
				ID:        q.ID.String(),
				ExamID:    exam.ID.String(),
				Question:  q.Question,
				Position:  q.Position,
				CreatedAt: q.CreatedAt,
				UpdatedAt: q.UpdatedAt,
			}
			if err := tx.Create(&qModel).Error; err != nil {
				return apperrors.InternalError(err.Error())
			}
			for _, a := range q.Answers {
				aModel := database.ExamAnswerModel{
					ID:         a.ID.String(),
					QuestionID: q.ID.String(),
					Answer:     a.Answer,
					IsCorrect:  a.IsCorrect,
					CreatedAt:  a.CreatedAt,
					UpdatedAt:  a.UpdatedAt,
				}
				if err := tx.Create(&aModel).Error; err != nil {
					return apperrors.InternalError(err.Error())
				}
			}
		}
		return nil
	})
}

func (r *GORMExamRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var questionIDs []string
		if err := tx.Model(&database.ExamQuestionModel{}).
			Where("exam_id = ?", id.String()).
			Pluck("id", &questionIDs).Error; err != nil {
			return apperrors.InternalError(err.Error())
		}

		if len(questionIDs) > 0 {
			if err := tx.Where("question_id IN ?", questionIDs).
				Delete(&database.ExamAnswerModel{}).Error; err != nil {
				return apperrors.InternalError(err.Error())
			}
		}

		if err := tx.Where("exam_id = ?", id.String()).
			Delete(&database.ExamQuestionModel{}).Error; err != nil {
			return apperrors.InternalError(err.Error())
		}

		result := tx.Where("id = ? AND org_id = ?", id.String(), orgID.String()).
			Delete(&database.ExamModel{})
		if result.Error != nil {
			return apperrors.InternalError(result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return apperrors.NotFoundError("exam", id.String())
		}
		return nil
	})
}

func toExamModel(e *entities.Exam) *database.ExamModel {
	questions := make([]database.ExamQuestionModel, len(e.Questions))
	for i, q := range e.Questions {
		answers := make([]database.ExamAnswerModel, len(q.Answers))
		for j, a := range q.Answers {
			answers[j] = database.ExamAnswerModel{
				ID:         a.ID.String(),
				QuestionID: q.ID.String(),
				Answer:     a.Answer,
				IsCorrect:  a.IsCorrect,
				CreatedAt:  a.CreatedAt,
				UpdatedAt:  a.UpdatedAt,
			}
		}
		questions[i] = database.ExamQuestionModel{
			ID:        q.ID.String(),
			ExamID:    e.ID.String(),
			Question:  q.Question,
			Position:  q.Position,
			Answers:   answers,
			CreatedAt: q.CreatedAt,
			UpdatedAt: q.UpdatedAt,
		}
	}
	return &database.ExamModel{
		ID:              e.ID.String(),
		CourseID:        e.CourseID.String(),
		OrgID:           e.OrgID.String(),
		AssessmentType:  e.AssessmentType,
		Title:           e.Title,
		Description:     e.Description,
		DurationMinutes: e.DurationMinutes,
		MaxAttempts:     e.MaxAttempts,
		DueDate:         e.DueDate,
		MCQEnabled:      e.MCQEnabled,
		MCQPoints:       e.MCQPoints,
		FileEnabled:     e.FileEnabled,
		FilePoints:      e.FilePoints,
		Questions:       questions,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

func toExamEntity(m *database.ExamModel) *entities.Exam {
	id, _ := uuid.Parse(m.ID)
	courseID, _ := uuid.Parse(m.CourseID)
	orgID, _ := uuid.Parse(m.OrgID)

	questions := make([]entities.ExamQuestion, len(m.Questions))
	for i, qm := range m.Questions {
		qID, _ := uuid.Parse(qm.ID)
		examID, _ := uuid.Parse(qm.ExamID)
		answers := make([]entities.ExamAnswer, len(qm.Answers))
		for j, am := range qm.Answers {
			aID, _ := uuid.Parse(am.ID)
			qnID, _ := uuid.Parse(am.QuestionID)
			answers[j] = entities.ExamAnswer{
				ID:         aID,
				QuestionID: qnID,
				Answer:     am.Answer,
				IsCorrect:  am.IsCorrect,
				CreatedAt:  am.CreatedAt,
				UpdatedAt:  am.UpdatedAt,
			}
		}
		questions[i] = entities.ExamQuestion{
			ID:        qID,
			ExamID:    examID,
			Question:  qm.Question,
			Position:  qm.Position,
			Answers:   answers,
			CreatedAt: qm.CreatedAt,
			UpdatedAt: qm.UpdatedAt,
		}
	}
	return &entities.Exam{
		ID:              id,
		CourseID:        courseID,
		OrgID:           orgID,
		AssessmentType:  m.AssessmentType,
		Title:           m.Title,
		Description:     m.Description,
		DurationMinutes: m.DurationMinutes,
		MaxAttempts:     m.MaxAttempts,
		DueDate:         m.DueDate,
		MCQEnabled:      m.MCQEnabled,
		MCQPoints:       m.MCQPoints,
		FileEnabled:     m.FileEnabled,
		FilePoints:      m.FilePoints,
		Questions:       questions,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
