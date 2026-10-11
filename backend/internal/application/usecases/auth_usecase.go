package usecases

import (
	"context"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	userRepo   repositories.UserRepository
	orgRepo    repositories.OrganizationRepository
	memberRepo repositories.MembershipRepository
	jwtSecret  string
	bcryptCost int
}

func NewAuthUseCase(
	userRepo repositories.UserRepository,
	orgRepo repositories.OrganizationRepository,
	memberRepo repositories.MembershipRepository,
	jwtSecret string,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:   userRepo,
		orgRepo:    orgRepo,
		memberRepo: memberRepo,
		jwtSecret:  jwtSecret,
		bcryptCost: bcrypt.DefaultCost,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.TokenResponse, error) {
	existing, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, apperrors.ConflictError("user")
	}

	existingOrg, _ := uc.orgRepo.FindBySlug(ctx, req.OrgSlug)
	if existingOrg != nil {
		return nil, apperrors.ConflictError("organization")
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

	org := &entities.Organization{
		ID:        uuid.New(),
		Name:      req.OrgName,
		Slug:      req.OrgSlug,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	membership := &entities.Membership{
		ID:        uuid.New(),
		UserID:    user.ID,
		OrgID:     org.ID,
		Role:      entities.RoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.memberRepo.Create(ctx, membership); err != nil {
		return nil, err
	}

	return uc.generateToken(user, org.ID, entities.RoleAdmin)
}

func (uc *AuthUseCase) Login(ctx context.Context, req dto.LoginRequest, orgSlug string) (*dto.TokenResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.UnauthorizedError("invalid credentials")
	}

	if !user.ValidatePassword(req.Password) {
		return nil, apperrors.UnauthorizedError("invalid credentials")
	}

	org, err := uc.orgRepo.FindBySlug(ctx, orgSlug)
	if err != nil {
		return nil, apperrors.NotFoundError("organization", orgSlug)
	}

	role, err := uc.memberRepo.FindUserRole(ctx, user.ID, org.ID)
	if err != nil {
		return nil, apperrors.ForbiddenError("user is not a member of this organization")
	}

	return uc.generateToken(user, org.ID, role)
}

// LoginWithOAuth turns a verified provider email into a tenant session. OAuth
// providers do not create memberships: an administrator must invite the user
// first, preserving organization boundaries.
func (uc *AuthUseCase) LoginWithOAuth(ctx context.Context, email, orgSlug string) (*dto.TokenResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.UnauthorizedError("this Google account is not registered")
	}
	org, err := uc.orgRepo.FindBySlug(ctx, orgSlug)
	if err != nil {
		return nil, apperrors.NotFoundError("organization", orgSlug)
	}
	role, err := uc.memberRepo.FindUserRole(ctx, user.ID, org.ID)
	if err != nil {
		return nil, apperrors.ForbiddenError("this Google account is not a member of this organization")
	}
	return uc.generateToken(user, org.ID, role)
}

func (uc *AuthUseCase) SuperAdminLogin(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.UnauthorizedError("invalid credentials")
	}

	if !user.ValidatePassword(req.Password) {
		return nil, apperrors.UnauthorizedError("invalid credentials")
	}

	if user.Role != string(entities.RoleSuperAdmin) {
		return nil, apperrors.ForbiddenError("not a super admin")
	}

	return uc.generateToken(user, uuid.Nil, entities.RoleSuperAdmin)
}

func (uc *AuthUseCase) generateToken(user *entities.User, orgID uuid.UUID, role entities.Role) (*dto.TokenResponse, error) {
	expiresAt := time.Now().Add(15 * time.Minute)
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"org_id":  orgID.String(),
		"role":    string(role),
		"exp":     expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return nil, apperrors.InternalError("failed to sign token")
	}
	return &dto.TokenResponse{
		AccessToken: signed,
		ExpiresAt:   expiresAt,
		User: dto.UserDTO{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
		},
	}, nil
}
