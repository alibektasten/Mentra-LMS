package usecases

import (
	"context"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
)

type UserUseCase struct {
	userRepo repositories.UserRepository
}

func NewUserUseCase(userRepo repositories.UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

func (uc *UserUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.ProfileResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.ProfileResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, apperrors.ValidationError("name is required")
	}

	user.Name = req.Name
	user.AvatarURL = req.AvatarURL

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, apperrors.InternalError("failed to update profile")
	}

	return &dto.ProfileResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
