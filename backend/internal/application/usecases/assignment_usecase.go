package usecases

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
)

type AssignmentUseCase struct {
	assignmentRepo repositories.AssignmentRepository
	moduleRepo     repositories.ModuleRepository
}

func NewAssignmentUseCase(assignmentRepo repositories.AssignmentRepository, moduleRepo repositories.ModuleRepository) *AssignmentUseCase {
	return &AssignmentUseCase{assignmentRepo: assignmentRepo, moduleRepo: moduleRepo}
}

func (uc *AssignmentUseCase) Create(ctx context.Context, courseID, moduleID, orgID uuid.UUID, req dto.CreateAssignmentRequest) (*dto.AssignmentDTO, error) {
	if _, err := uc.moduleRepo.FindByID(ctx, moduleID, orgID); err != nil {
		return nil, err
	}
	a := &entities.Assignment{
		ID:                  uuid.New(),
		OrgID:               orgID,
		CourseID:            courseID,
		ModuleID:            moduleID,
		Title:               req.Title,
		Description:         req.Description,
		Rubric:              req.Rubric,
		MaxPoints:           req.MaxPoints,
		DueDate:             req.DueDate,
		AllowLateSubmission: req.AllowLateSubmission,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	if err := uc.assignmentRepo.Create(ctx, a); err != nil {
		return nil, err
	}
	return toAssignmentDTO(a), nil
}

func (uc *AssignmentUseCase) GetByModule(ctx context.Context, moduleID, orgID uuid.UUID) ([]*dto.AssignmentDTO, error) {
	assignments, err := uc.assignmentRepo.FindByModule(ctx, moduleID, orgID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.AssignmentDTO, len(assignments))
	for i, a := range assignments {
		result[i] = toAssignmentDTO(a)
	}
	return result, nil
}

func (uc *AssignmentUseCase) GetByID(ctx context.Context, id, orgID uuid.UUID) (*dto.AssignmentDTO, error) {
	a, err := uc.assignmentRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	return toAssignmentDTO(a), nil
}

func (uc *AssignmentUseCase) Update(ctx context.Context, id, orgID uuid.UUID, req dto.UpdateAssignmentRequest) (*dto.AssignmentDTO, error) {
	a, err := uc.assignmentRepo.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Description != nil {
		a.Description = *req.Description
	}
	if req.Rubric != nil {
		a.Rubric = *req.Rubric
	}
	if req.MaxPoints != nil {
		a.MaxPoints = *req.MaxPoints
	}
	if req.DueDate != nil {
		a.DueDate = req.DueDate
	}
	if req.AllowLateSubmission != nil {
		a.AllowLateSubmission = *req.AllowLateSubmission
	}
	a.UpdatedAt = time.Now()
	if err := uc.assignmentRepo.Update(ctx, a); err != nil {
		return nil, err
	}
	return toAssignmentDTO(a), nil
}

func (uc *AssignmentUseCase) Delete(ctx context.Context, id, orgID uuid.UUID) error {
	return uc.assignmentRepo.Delete(ctx, id, orgID)
}

func (uc *AssignmentUseCase) Submit(ctx context.Context, assignmentID, studentID, orgID uuid.UUID, textContent, linkURL, filePath string) (*dto.SubmissionDTO, error) {
	assignment, err := uc.assignmentRepo.FindByID(ctx, assignmentID, orgID)
	if err != nil {
		return nil, err
	}

	// Check deadline
	if assignment.DueDate != nil && time.Now().After(*assignment.DueDate) && !assignment.AllowLateSubmission {
		return nil, apperrors.ValidationError("submission deadline has passed")
	}

	// Check if already submitted — update if exists
	existing, err := uc.assignmentRepo.FindSubmission(ctx, assignmentID, studentID)
	if err == nil && existing != nil {
		existing.TextContent = textContent
		existing.LinkURL = linkURL
		if filePath != "" {
			existing.FilePath = filePath
		}
		existing.UpdatedAt = time.Now()
		if err := uc.assignmentRepo.UpdateSubmission(ctx, existing); err != nil {
			return nil, err
		}
		return toSubmissionDTO(existing), nil
	}

	s := &entities.AssignmentSubmission{
		ID:           uuid.New(),
		AssignmentID: assignmentID,
		StudentID:    studentID,
		OrgID:        orgID,
		TextContent:  textContent,
		LinkURL:      linkURL,
		FilePath:     filePath,
		SubmittedAt:  time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.assignmentRepo.CreateSubmission(ctx, s); err != nil {
		return nil, err
	}
	return toSubmissionDTO(s), nil
}

func (uc *AssignmentUseCase) GetMySubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (*dto.SubmissionDTO, error) {
	s, err := uc.assignmentRepo.FindSubmission(ctx, assignmentID, studentID)
	if err != nil {
		return nil, err
	}
	return toSubmissionDTO(s), nil
}

func (uc *AssignmentUseCase) ListSubmissions(ctx context.Context, assignmentID uuid.UUID) ([]*dto.SubmissionDTO, error) {
	submissions, err := uc.assignmentRepo.FindSubmissionsByAssignment(ctx, assignmentID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.SubmissionDTO, len(submissions))
	for i, s := range submissions {
		result[i] = toSubmissionDTO(s)
	}
	return result, nil
}

func (uc *AssignmentUseCase) DeleteMySubmission(ctx context.Context, assignmentID, studentID uuid.UUID) error {
	s, err := uc.assignmentRepo.FindSubmission(ctx, assignmentID, studentID)
	if err != nil {
		return err
	}
	if s.Score != nil {
		return apperrors.ValidationError("cannot delete a graded submission")
	}
	return uc.assignmentRepo.DeleteSubmission(ctx, s.ID, studentID)
}

func (uc *AssignmentUseCase) GradeSubmission(ctx context.Context, submissionID, graderID uuid.UUID, req dto.GradeSubmissionRequest) (*dto.SubmissionDTO, error) {
	s, err := uc.assignmentRepo.FindSubmissionByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	assignment, err := uc.assignmentRepo.FindByID(ctx, s.AssignmentID, s.OrgID)
	if err != nil {
		return nil, err
	}
	if req.Score < 0 || req.Score > assignment.MaxPoints {
		return nil, apperrors.ValidationError("score must be between 0 and the assignment maximum")
	}
	s.Score = &req.Score
	s.Feedback = req.Feedback
	s.GradedBy = &graderID
	now := time.Now()
	s.GradedAt = &now
	s.UpdatedAt = now
	if err := uc.assignmentRepo.UpdateSubmission(ctx, s); err != nil {
		return nil, err
	}
	return toSubmissionDTO(s), nil
}

func toAssignmentDTO(a *entities.Assignment) *dto.AssignmentDTO {
	return &dto.AssignmentDTO{
		ID:                  a.ID.String(),
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
}

func toSubmissionDTO(s *entities.AssignmentSubmission) *dto.SubmissionDTO {
	return &dto.SubmissionDTO{
		ID:           s.ID.String(),
		AssignmentID: s.AssignmentID.String(),
		StudentID:    s.StudentID.String(),
		TextContent:  s.TextContent,
		LinkURL:      s.LinkURL,
		FilePath:     s.FilePath,
		Score:        s.Score,
		Feedback:     s.Feedback,
		GradedAt:     s.GradedAt,
		SubmittedAt:  s.SubmittedAt,
	}
}
