package usecases

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAssignmentUseCase_Create_Success(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	courseID := uuid.New()
	moduleID := uuid.New()

	module := &entities.Module{ID: moduleID, OrgID: orgID}
	moduleRepo.On("FindByID", mock.Anything, moduleID, orgID).Return(module, nil)
	assignmentRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Assignment")).Return(nil)

	req := dto.CreateAssignmentRequest{Title: "HW1", Description: "Do it", MaxPoints: 100}
	result, err := uc.Create(context.Background(), courseID, moduleID, orgID, req)
	require.NoError(t, err)
	assert.Equal(t, "HW1", result.Title)
	assert.Equal(t, 100, result.MaxPoints)
}

func TestAssignmentUseCase_Create_ModuleNotFound(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	moduleID := uuid.New()
	moduleRepo.On("FindByID", mock.Anything, moduleID, orgID).Return(nil, apperrors.NotFoundError("module", moduleID.String()))

	_, err := uc.Create(context.Background(), uuid.New(), moduleID, orgID, dto.CreateAssignmentRequest{Title: "HW"})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.HTTPStatus)
}

func TestAssignmentUseCase_Submit_Success_NewSubmission(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	assignmentID := uuid.New()
	studentID := uuid.New()

	assignment := &entities.Assignment{
		ID:                  assignmentID,
		OrgID:               orgID,
		AllowLateSubmission: false,
	}

	assignmentRepo.On("FindByID", mock.Anything, assignmentID, orgID).Return(assignment, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, assignmentID, studentID).Return(nil, apperrors.NotFoundError("submission", ""))
	assignmentRepo.On("CreateSubmission", mock.Anything, mock.AnythingOfType("*entities.AssignmentSubmission")).Return(nil)

	result, err := uc.Submit(context.Background(), assignmentID, studentID, orgID, "My answer", "", "")
	require.NoError(t, err)
	assert.Equal(t, "My answer", result.TextContent)
	assert.Equal(t, studentID.String(), result.StudentID)
}

func TestAssignmentUseCase_Submit_UpdatesExistingSubmission(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	assignmentID := uuid.New()
	studentID := uuid.New()

	assignment := &entities.Assignment{ID: assignmentID, OrgID: orgID}
	existing := &entities.AssignmentSubmission{
		ID:           uuid.New(),
		AssignmentID: assignmentID,
		StudentID:    studentID,
		TextContent:  "Old answer",
	}

	assignmentRepo.On("FindByID", mock.Anything, assignmentID, orgID).Return(assignment, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, assignmentID, studentID).Return(existing, nil)
	assignmentRepo.On("UpdateSubmission", mock.Anything, mock.AnythingOfType("*entities.AssignmentSubmission")).Return(nil)

	result, err := uc.Submit(context.Background(), assignmentID, studentID, orgID, "New answer", "", "")
	require.NoError(t, err)
	assert.Equal(t, "New answer", result.TextContent)
	assignmentRepo.AssertNotCalled(t, "CreateSubmission", mock.Anything, mock.Anything)
}

func TestAssignmentUseCase_Submit_DeadlinePassed_LateNotAllowed(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	assignmentID := uuid.New()
	studentID := uuid.New()

	pastDue := time.Now().Add(-24 * time.Hour)
	assignment := &entities.Assignment{
		ID:                  assignmentID,
		OrgID:               orgID,
		DueDate:             &pastDue,
		AllowLateSubmission: false,
	}

	assignmentRepo.On("FindByID", mock.Anything, assignmentID, orgID).Return(assignment, nil)

	_, err := uc.Submit(context.Background(), assignmentID, studentID, orgID, "answer", "", "")
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}

func TestAssignmentUseCase_Submit_DeadlinePassed_LateAllowed_Succeeds(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	assignmentID := uuid.New()
	studentID := uuid.New()

	pastDue := time.Now().Add(-24 * time.Hour)
	assignment := &entities.Assignment{
		ID:                  assignmentID,
		OrgID:               orgID,
		DueDate:             &pastDue,
		AllowLateSubmission: true,
	}

	assignmentRepo.On("FindByID", mock.Anything, assignmentID, orgID).Return(assignment, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, assignmentID, studentID).Return(nil, apperrors.NotFoundError("submission", ""))
	assignmentRepo.On("CreateSubmission", mock.Anything, mock.AnythingOfType("*entities.AssignmentSubmission")).Return(nil)

	_, err := uc.Submit(context.Background(), assignmentID, studentID, orgID, "late answer", "", "")
	require.NoError(t, err)
}

func TestAssignmentUseCase_Submit_NoDueDate_Succeeds(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	orgID := uuid.New()
	assignmentID := uuid.New()
	studentID := uuid.New()

	assignment := &entities.Assignment{ID: assignmentID, OrgID: orgID, DueDate: nil}
	assignmentRepo.On("FindByID", mock.Anything, assignmentID, orgID).Return(assignment, nil)
	assignmentRepo.On("FindSubmission", mock.Anything, assignmentID, studentID).Return(nil, apperrors.NotFoundError("submission", ""))
	assignmentRepo.On("CreateSubmission", mock.Anything, mock.AnythingOfType("*entities.AssignmentSubmission")).Return(nil)

	_, err := uc.Submit(context.Background(), assignmentID, studentID, orgID, "answer", "", "")
	require.NoError(t, err)
}

func TestAssignmentUseCase_GradeSubmission_Success(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	moduleRepo := new(mocks.MockModuleRepository)
	uc := NewAssignmentUseCase(assignmentRepo, moduleRepo)

	graderID := uuid.New()
	submissionID := uuid.New()
	assignmentID := uuid.New()
	orgID := uuid.New()
	submission := &entities.AssignmentSubmission{
		ID:           submissionID,
		AssignmentID: assignmentID,
		OrgID:        orgID,
		TextContent:  "student answer",
	}

	assignmentRepo.On("FindSubmissionByID", mock.Anything, submissionID).Return(submission, nil)
	assignmentRepo.On("FindByID", mock.Anything, assignmentID, orgID).Return(&entities.Assignment{ID: assignmentID, OrgID: orgID, MaxPoints: 100}, nil)
	assignmentRepo.On("UpdateSubmission", mock.Anything, mock.AnythingOfType("*entities.AssignmentSubmission")).Return(nil)

	result, err := uc.GradeSubmission(context.Background(), submissionID, graderID, dto.GradeSubmissionRequest{Score: 85, Feedback: "Well done"})
	require.NoError(t, err)
	require.NotNil(t, result.Score)
	assert.Equal(t, 85, *result.Score)
	assert.Equal(t, "Well done", result.Feedback)
}

func TestAssignmentUseCase_GradeSubmission_RejectsScoreAboveMaximum(t *testing.T) {
	assignmentRepo := new(mocks.MockAssignmentRepository)
	uc := NewAssignmentUseCase(assignmentRepo, new(mocks.MockModuleRepository))
	assignmentID, orgID, submissionID := uuid.New(), uuid.New(), uuid.New()
	submission := &entities.AssignmentSubmission{ID: submissionID, AssignmentID: assignmentID, OrgID: orgID}

	assignmentRepo.On("FindSubmissionByID", mock.Anything, submissionID).Return(submission, nil)
	assignmentRepo.On("FindByID", mock.Anything, assignmentID, orgID).Return(&entities.Assignment{ID: assignmentID, OrgID: orgID, MaxPoints: 10}, nil)

	_, err := uc.GradeSubmission(context.Background(), submissionID, uuid.New(), dto.GradeSubmissionRequest{Score: 11})
	require.Error(t, err)
}
