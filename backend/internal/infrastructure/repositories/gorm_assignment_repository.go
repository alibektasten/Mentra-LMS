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

type GORMAssignmentRepository struct {
	db *gorm.DB
}

func NewGORMAssignmentRepository(db *gorm.DB) *GORMAssignmentRepository {
	return &GORMAssignmentRepository{db: db}
}

func (r *GORMAssignmentRepository) Create(ctx context.Context, a *entities.Assignment) error {
	return r.db.WithContext(ctx).Create(toAssignmentModel(a)).Error
}

func (r *GORMAssignmentRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (*entities.Assignment, error) {
	var model database.AssignmentModel
	err := r.db.WithContext(ctx).First(&model, "id = ? AND org_id = ?", id.String(), orgID.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("assignment", id.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toAssignmentEntity(&model), nil
}

func (r *GORMAssignmentRepository) FindByModule(ctx context.Context, moduleID, orgID uuid.UUID) ([]*entities.Assignment, error) {
	var models []database.AssignmentModel
	err := r.db.WithContext(ctx).
		Where("module_id = ? AND org_id = ?", moduleID.String(), orgID.String()).
		Order("position ASC").Find(&models).Error
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]*entities.Assignment, len(models))
	for i := range models {
		result[i] = toAssignmentEntity(&models[i])
	}
	return result, nil
}

func (r *GORMAssignmentRepository) FindByCourse(ctx context.Context, courseID, orgID uuid.UUID) ([]*entities.Assignment, error) {
	var models []database.AssignmentModel
	err := r.db.WithContext(ctx).
		Where("course_id = ? AND org_id = ?", courseID.String(), orgID.String()).
		Order("position ASC").Find(&models).Error
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]*entities.Assignment, len(models))
	for i := range models {
		result[i] = toAssignmentEntity(&models[i])
	}
	return result, nil
}

func (r *GORMAssignmentRepository) Update(ctx context.Context, a *entities.Assignment) error {
	a.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toAssignmentModel(a)).Error
}

func (r *GORMAssignmentRepository) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND org_id = ?", id.String(), orgID.String()).Delete(&database.AssignmentModel{})
	if result.Error != nil {
		return apperrors.InternalError(result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFoundError("assignment", id.String())
	}
	return nil
}

func (r *GORMAssignmentRepository) CreateSubmission(ctx context.Context, s *entities.AssignmentSubmission) error {
	return r.db.WithContext(ctx).Create(toSubmissionModel(s)).Error
}

func (r *GORMAssignmentRepository) FindSubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (*entities.AssignmentSubmission, error) {
	var model database.AssignmentSubmissionModel
	err := r.db.WithContext(ctx).First(&model, "assignment_id = ? AND student_id = ?", assignmentID.String(), studentID.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("submission", assignmentID.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toSubmissionEntity(&model), nil
}

func (r *GORMAssignmentRepository) FindSubmissionByID(ctx context.Context, id uuid.UUID) (*entities.AssignmentSubmission, error) {
	var model database.AssignmentSubmissionModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id.String()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.NotFoundError("submission", id.String())
	}
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	return toSubmissionEntity(&model), nil
}

func (r *GORMAssignmentRepository) FindSubmissionsByAssignment(ctx context.Context, assignmentID uuid.UUID) ([]*entities.AssignmentSubmission, error) {
	var models []database.AssignmentSubmissionModel
	err := r.db.WithContext(ctx).
		Where("assignment_id = ?", assignmentID.String()).
		Order("CASE WHEN score IS NULL THEN 0 ELSE 1 END").
		Order("submitted_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, apperrors.InternalError(err.Error())
	}
	result := make([]*entities.AssignmentSubmission, len(models))
	for i := range models {
		result[i] = toSubmissionEntity(&models[i])
	}
	return result, nil
}

func (r *GORMAssignmentRepository) UpdateSubmission(ctx context.Context, s *entities.AssignmentSubmission) error {
	s.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(toSubmissionModel(s)).Error
}

func (r *GORMAssignmentRepository) DeleteSubmission(ctx context.Context, id, studentID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND student_id = ?", id.String(), studentID.String()).
		Delete(&database.AssignmentSubmissionModel{})
	if result.Error != nil {
		return apperrors.InternalError(result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFoundError("submission", id.String())
	}
	return nil
}

func toAssignmentModel(a *entities.Assignment) *database.AssignmentModel {
	m := &database.AssignmentModel{
		ID:                  a.ID.String(),
		OrgID:               a.OrgID.String(),
		CourseID:            a.CourseID.String(),
		ModuleID:            a.ModuleID.String(),
		Title:               a.Title,
		Description:         a.Description,
		Rubric:              a.Rubric,
		MaxPoints:           a.MaxPoints,
		DueDate:             a.DueDate,
		AllowLateSubmission: a.AllowLateSubmission,
		Position:            a.Position,
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}
	return m
}

func toAssignmentEntity(m *database.AssignmentModel) *entities.Assignment {
	id, _ := uuid.Parse(m.ID)
	orgID, _ := uuid.Parse(m.OrgID)
	courseID, _ := uuid.Parse(m.CourseID)
	moduleID, _ := uuid.Parse(m.ModuleID)
	return &entities.Assignment{
		ID:                  id,
		OrgID:               orgID,
		CourseID:            courseID,
		ModuleID:            moduleID,
		Title:               m.Title,
		Description:         m.Description,
		Rubric:              m.Rubric,
		MaxPoints:           m.MaxPoints,
		DueDate:             m.DueDate,
		AllowLateSubmission: m.AllowLateSubmission,
		Position:            m.Position,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func toSubmissionModel(s *entities.AssignmentSubmission) *database.AssignmentSubmissionModel {
	m := &database.AssignmentSubmissionModel{
		ID:           s.ID.String(),
		AssignmentID: s.AssignmentID.String(),
		StudentID:    s.StudentID.String(),
		OrgID:        s.OrgID.String(),
		TextContent:  s.TextContent,
		LinkURL:      s.LinkURL,
		FilePath:     s.FilePath,
		Score:        s.Score,
		Feedback:     s.Feedback,
		GradedAt:     s.GradedAt,
		SubmittedAt:  s.SubmittedAt,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
	if s.GradedBy != nil {
		str := s.GradedBy.String()
		m.GradedBy = &str
	}
	return m
}

func toSubmissionEntity(m *database.AssignmentSubmissionModel) *entities.AssignmentSubmission {
	id, _ := uuid.Parse(m.ID)
	assignmentID, _ := uuid.Parse(m.AssignmentID)
	studentID, _ := uuid.Parse(m.StudentID)
	orgID, _ := uuid.Parse(m.OrgID)
	s := &entities.AssignmentSubmission{
		ID:           id,
		AssignmentID: assignmentID,
		StudentID:    studentID,
		OrgID:        orgID,
		TextContent:  m.TextContent,
		LinkURL:      m.LinkURL,
		FilePath:     m.FilePath,
		Score:        m.Score,
		Feedback:     m.Feedback,
		GradedAt:     m.GradedAt,
		SubmittedAt:  m.SubmittedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
	if m.GradedBy != nil {
		parsed, _ := uuid.Parse(*m.GradedBy)
		s.GradedBy = &parsed
	}
	return s
}
