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

type FileAttachmentUseCase struct {
	fileAttachmentRepo repositories.FileAttachmentRepository
}

func NewFileAttachmentUseCase(fileAttachmentRepo repositories.FileAttachmentRepository) *FileAttachmentUseCase {
	return &FileAttachmentUseCase{fileAttachmentRepo: fileAttachmentRepo}
}

func (uc *FileAttachmentUseCase) CreateAttachment(ctx context.Context, orgID, uploaderID uuid.UUID, req dto.CreateAttachmentRequest) (*dto.FileAttachmentDTO, error) {
	refID, err := uuid.Parse(req.RefID)
	if err != nil {
		return nil, apperrors.ValidationError("invalid ref_id")
	}

	f := &entities.FileAttachment{
		ID:           uuid.New(),
		OrgID:        orgID,
		UploaderID:   uploaderID,
		OriginalName: req.OriginalName,
		StoredPath:   req.StoredPath,
		MimeType:     req.MimeType,
		SizeBytes:    req.SizeBytes,
		RefType:      req.RefType,
		RefID:        refID,
		CreatedAt:    time.Now(),
	}
	if err := uc.fileAttachmentRepo.Create(ctx, f); err != nil {
		return nil, err
	}
	return toFileAttachmentDTO(f), nil
}

func (uc *FileAttachmentUseCase) ListByRef(ctx context.Context, refType string, refID uuid.UUID) ([]*dto.FileAttachmentDTO, error) {
	attachments, err := uc.fileAttachmentRepo.FindByRef(ctx, refType, refID)
	if err != nil {
		return nil, err
	}
	result := make([]*dto.FileAttachmentDTO, len(attachments))
	for i, a := range attachments {
		result[i] = toFileAttachmentDTO(a)
	}
	return result, nil
}

func (uc *FileAttachmentUseCase) GetAttachment(ctx context.Context, id, orgID uuid.UUID) (*dto.FileAttachmentDTO, error) {
	f, err := uc.fileAttachmentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if f.OrgID != orgID {
		return nil, apperrors.ForbiddenError("access denied")
	}
	return toFileAttachmentDTO(f), nil
}

func (uc *FileAttachmentUseCase) DeleteAttachment(ctx context.Context, id, orgID uuid.UUID) error {
	f, err := uc.fileAttachmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if f.OrgID != orgID {
		return apperrors.ForbiddenError("access denied")
	}
	return uc.fileAttachmentRepo.Delete(ctx, id)
}

func toFileAttachmentDTO(f *entities.FileAttachment) *dto.FileAttachmentDTO {
	return &dto.FileAttachmentDTO{
		ID:           f.ID.String(),
		RefType:      f.RefType,
		RefID:        f.RefID.String(),
		StoredPath:   f.StoredPath,
		OriginalName: f.OriginalName,
		MimeType:     f.MimeType,
		SizeBytes:    f.SizeBytes,
		CreatedAt:    f.CreatedAt,
	}
}
