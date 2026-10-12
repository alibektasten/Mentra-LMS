package usecases

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newSuperAdminUC(userRepo *mocks.MockUserRepository, orgRepo *mocks.MockOrganizationRepository, memberRepo *mocks.MockMembershipRepository) *SuperAdminUseCase {
	uc := NewSuperAdminUseCase(userRepo, orgRepo, memberRepo)
	uc.bcryptCost = bcrypt.MinCost
	return uc
}

func TestSuperAdminUseCase_ListOrgs_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := newSuperAdminUC(userRepo, orgRepo, memberRepo)

	now := time.Now()
	orgs := []entities.Organization{
		{ID: uuid.New(), Name: "Org A", Slug: "org-a", CreatedAt: now},
		{ID: uuid.New(), Name: "Org B", Slug: "org-b", CreatedAt: now},
	}
	orgRepo.On("ListAll", mock.Anything, 1, 10).Return(orgs, int64(2), nil)

	result, total, err := uc.ListOrgs(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, "Org A", result[0].Name)
}

func TestSuperAdminUseCase_DeleteOrg_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := newSuperAdminUC(userRepo, orgRepo, memberRepo)

	orgID := uuid.New()
	memberRepo.On("FindByOrg", mock.Anything, orgID, 1, 1).Return([]entities.Membership{}, int64(0), nil)
	orgRepo.On("Delete", mock.Anything, orgID).Return(nil)

	err := uc.DeleteOrg(context.Background(), orgID)
	require.NoError(t, err)
}

func TestSuperAdminUseCase_DeleteOrg_BlocksOrganizationWithMembers(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := newSuperAdminUC(userRepo, orgRepo, memberRepo)

	orgID := uuid.New()
	memberRepo.On("FindByOrg", mock.Anything, orgID, 1, 1).Return([]entities.Membership{}, int64(1), nil)

	err := uc.DeleteOrg(context.Background(), orgID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
	orgRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestSuperAdminUseCase_ListUsers_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := newSuperAdminUC(userRepo, orgRepo, memberRepo)

	now := time.Now()
	users := []repositories.SystemUser{
		{ID: uuid.New(), Email: "u1@test.com", Name: "User 1", Role: "student", OrgName: "Test University", CreatedAt: now},
		{ID: uuid.New(), Email: "u2@test.com", Name: "User 2", Role: "teacher", OrgName: "Test University", CreatedAt: now},
	}
	filters := repositories.SystemUserFilters{Page: 1, PageSize: 20, SortBy: "created_at", SortOrder: "desc"}
	userRepo.On("ListForSystemAdmin", mock.Anything, filters).Return(users, int64(2), nil)

	result, total, err := uc.ListUsers(context.Background(), filters)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, "u1@test.com", result[0].Email)
	assert.Equal(t, "student", result[0].Role)
}

func TestSuperAdminUseCase_GetStats_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := newSuperAdminUC(userRepo, orgRepo, memberRepo)

	orgRepo.On("ListAll", mock.Anything, 1, 1).Return([]entities.Organization{}, int64(5), nil)
	userRepo.On("ListAll", mock.Anything, 1, 1).Return([]entities.User{}, int64(42), nil)

	result, err := uc.GetStats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(5), result.TotalOrgs)
	assert.Equal(t, int64(42), result.TotalUsers)
}

func TestSuperAdminUseCase_InviteOrgAdmin_Success(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := newSuperAdminUC(userRepo, orgRepo, memberRepo)

	orgID := uuid.New()
	org := &entities.Organization{ID: orgID, Name: "Test Org", Slug: "test-org"}

	orgRepo.On("FindByID", mock.Anything, orgID).Return(org, nil)
	userRepo.On("FindByEmail", mock.Anything, "newadmin@test.com").Return(nil, apperrors.NotFoundError("user", ""))
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)
	memberRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Membership")).Return(nil)

	req := dto.InviteOrgAdminRequest{
		Email:    "newadmin@test.com",
		Name:     "New Admin",
		Password: "password123",
		OrgID:    orgID.String(),
	}
	result, err := uc.InviteOrgAdmin(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "newadmin@test.com", result.Email)
	assert.Equal(t, "New Admin", result.Name)
}

func TestSuperAdminUseCase_InviteOrgAdmin_InvalidOrgID(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := newSuperAdminUC(userRepo, orgRepo, memberRepo)

	req := dto.InviteOrgAdminRequest{
		Email:    "admin@test.com",
		Name:     "Admin",
		Password: "password123",
		OrgID:    "not-a-uuid",
	}
	_, err := uc.InviteOrgAdmin(context.Background(), req)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusUnprocessableEntity, appErr.HTTPStatus)
}

func TestSuperAdminUseCase_InviteOrgAdmin_DuplicateEmail(t *testing.T) {
	userRepo := new(mocks.MockUserRepository)
	orgRepo := new(mocks.MockOrganizationRepository)
	memberRepo := new(mocks.MockMembershipRepository)
	uc := newSuperAdminUC(userRepo, orgRepo, memberRepo)

	orgID := uuid.New()
	org := &entities.Organization{ID: orgID, Name: "Test Org"}
	existingUser := makeUser("existing@test.com", "pass", entities.RoleAdmin)

	orgRepo.On("FindByID", mock.Anything, orgID).Return(org, nil)
	userRepo.On("FindByEmail", mock.Anything, "existing@test.com").Return(existingUser, nil)

	req := dto.InviteOrgAdminRequest{
		Email:    "existing@test.com",
		Name:     "Admin",
		Password: "password123",
		OrgID:    orgID.String(),
	}
	_, err := uc.InviteOrgAdmin(context.Background(), req)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.HTTPStatus)
}
