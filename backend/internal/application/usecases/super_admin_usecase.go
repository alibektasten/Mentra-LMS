package usecases

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SuperAdminUseCase struct {
	userRepo   repositories.UserRepository
	orgRepo    repositories.OrganizationRepository
	memberRepo repositories.MembershipRepository
	bcryptCost int
}

func NewSuperAdminUseCase(
	userRepo repositories.UserRepository,
	orgRepo repositories.OrganizationRepository,
	memberRepo repositories.MembershipRepository,
) *SuperAdminUseCase {
	return &SuperAdminUseCase{
		userRepo:   userRepo,
		orgRepo:    orgRepo,
		memberRepo: memberRepo,
		bcryptCost: bcrypt.DefaultCost,
	}
}

func (uc *SuperAdminUseCase) ListOrgs(ctx context.Context, page, pageSize int) ([]dto.OrgDTO, int64, error) {
	orgs, total, err := uc.orgRepo.ListAll(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.OrgDTO, len(orgs))
	for i, o := range orgs {
		result[i] = dto.OrgDTO{
			ID:        o.ID.String(),
			Name:      o.Name,
			Slug:      o.Slug,
			CreatedAt: o.CreatedAt,
		}
	}
	return result, total, nil
}

func (uc *SuperAdminUseCase) DeleteOrg(ctx context.Context, id uuid.UUID) error {
	_, members, err := uc.memberRepo.FindByOrg(ctx, id, 1, 1)
	if err != nil {
		return err
	}
	if members > 0 {
		return apperrors.ValidationError("remove organization members before deleting an organization")
	}
	return uc.orgRepo.Delete(ctx, id)
}

func (uc *SuperAdminUseCase) CreateOrg(ctx context.Context, req dto.CreateOrganizationRequest) (*dto.OrgDTO, error) {
	if existing, _ := uc.orgRepo.FindBySlug(ctx, req.Slug); existing != nil {
		return nil, apperrors.ConflictError("organization slug")
	}
	now := time.Now()
	org := &entities.Organization{ID: uuid.New(), Name: req.Name, Slug: req.Slug, CreatedAt: now, UpdatedAt: now}
	if err := uc.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}
	return &dto.OrgDTO{ID: org.ID.String(), Name: org.Name, Slug: org.Slug, CreatedAt: org.CreatedAt}, nil
}

func (uc *SuperAdminUseCase) ListUsers(ctx context.Context, filters repositories.SystemUserFilters) ([]dto.AdminUserDTO, int64, error) {
	users, total, err := uc.userRepo.ListForSystemAdmin(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.AdminUserDTO, len(users))
	for i, u := range users {
		result[i] = dto.AdminUserDTO{
			ID:        u.ID.String(),
			Email:     u.Email,
			Name:      u.Name,
			Role:      u.Role,
			OrgID:     u.OrgID,
			OrgName:   u.OrgName,
			CreatedAt: u.CreatedAt,
		}
	}
	return result, total, nil
}

func (uc *SuperAdminUseCase) GetStats(ctx context.Context) (*dto.SystemStatsDTO, error) {
	_, totalOrgs, err := uc.orgRepo.ListAll(ctx, 1, 1)
	if err != nil {
		return nil, err
	}
	_, totalUsers, err := uc.userRepo.ListAll(ctx, 1, 1)
	if err != nil {
		return nil, err
	}
	return &dto.SystemStatsDTO{
		TotalOrgs:  totalOrgs,
		TotalUsers: totalUsers,
	}, nil
}

func (uc *SuperAdminUseCase) InviteOrgAdmin(ctx context.Context, req dto.InviteOrgAdminRequest) (*dto.AdminUserDTO, error) {
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return nil, apperrors.ValidationError("invalid org_id")
	}

	_, err = uc.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	existing, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, apperrors.ConflictError("user")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), uc.bcryptCost)
	if err != nil {
		return nil, apperrors.InternalError("failed to hash password")
	}

	user := &entities.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	membership := &entities.Membership{
		ID:        uuid.New(),
		UserID:    user.ID,
		OrgID:     orgID,
		Role:      entities.RoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.memberRepo.Create(ctx, membership); err != nil {
		return nil, err
	}

	return &dto.AdminUserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	}, nil
}
