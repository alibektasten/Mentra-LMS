package usecases

import (
	"context"
	"net/http"
	"testing"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFileAttachmentUseCase_DeleteAttachment_Success(t *testing.T) {
	fileAttachmentRepo := new(mocks.MockFileAttachmentRepository)
	uc := NewFileAttachmentUseCase(fileAttachmentRepo)

	orgID := uuid.New()
	attachmentID := uuid.New()
	refID := uuid.New()

	attachment := &entities.FileAttachment{
		ID:    attachmentID,
		OrgID: orgID,
		RefID: refID,
	}

	fileAttachmentRepo.On("FindByID", mock.Anything, attachmentID).Return(attachment, nil)
	fileAttachmentRepo.On("Delete", mock.Anything, attachmentID).Return(nil)

	err := uc.DeleteAttachment(context.Background(), attachmentID, orgID)
	require.NoError(t, err)
	fileAttachmentRepo.AssertCalled(t, "Delete", mock.Anything, attachmentID)
}

func TestFileAttachmentUseCase_DeleteAttachment_WrongOrg_ReturnsForbidden(t *testing.T) {
	fileAttachmentRepo := new(mocks.MockFileAttachmentRepository)
	uc := NewFileAttachmentUseCase(fileAttachmentRepo)

	orgID := uuid.New()
	differentOrgID := uuid.New()
	attachmentID := uuid.New()

	attachment := &entities.FileAttachment{
		ID:    attachmentID,
		OrgID: orgID, // belongs to orgID
	}

	fileAttachmentRepo.On("FindByID", mock.Anything, attachmentID).Return(attachment, nil)

	// Caller is from differentOrgID — should be forbidden
	err := uc.DeleteAttachment(context.Background(), attachmentID, differentOrgID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusForbidden, appErr.HTTPStatus)
	fileAttachmentRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestFileAttachmentUseCase_CreateAttachment_Success(t *testing.T) {
	fileAttachmentRepo := new(mocks.MockFileAttachmentRepository)
	uc := NewFileAttachmentUseCase(fileAttachmentRepo)

	orgID := uuid.New()
	uploaderID := uuid.New()
	refID := uuid.New()

	fileAttachmentRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.FileAttachment")).Return(nil)

	req := dto.CreateAttachmentRequest{
		OriginalName: "document.pdf",
		StoredPath:   "uploads/document.pdf",
		MimeType:     "application/pdf",
		SizeBytes:    1024,
		RefType:      "lesson",
		RefID:        refID.String(),
	}

	result, err := uc.CreateAttachment(context.Background(), orgID, uploaderID, req)
	require.NoError(t, err)
	assert.Equal(t, "document.pdf", result.OriginalName)
	assert.Equal(t, "application/pdf", result.MimeType)
}

func TestFileAttachmentUseCase_CreateAttachment_InvalidRefID(t *testing.T) {
	fileAttachmentRepo := new(mocks.MockFileAttachmentRepository)
	uc := NewFileAttachmentUseCase(fileAttachmentRepo)

	req := dto.CreateAttachmentRequest{
		RefID: "not-a-uuid",
	}

	_, err := uc.CreateAttachment(context.Background(), uuid.New(), uuid.New(), req)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}

func TestFileAttachmentUseCase_GetAttachment_RejectsCrossOrgAccess(t *testing.T) {
	fileAttachmentRepo := new(mocks.MockFileAttachmentRepository)
	uc := NewFileAttachmentUseCase(fileAttachmentRepo)

	attachmentID := uuid.New()
	attachmentOrgID := uuid.New()
	requestOrgID := uuid.New()
	fileAttachmentRepo.On("FindByID", mock.Anything, attachmentID).
		Return(&entities.FileAttachment{ID: attachmentID, OrgID: attachmentOrgID}, nil).Once()

	_, err := uc.GetAttachment(context.Background(), attachmentID, requestOrgID)

	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusForbidden, appErr.HTTPStatus)
	fileAttachmentRepo.AssertExpectations(t)
}
