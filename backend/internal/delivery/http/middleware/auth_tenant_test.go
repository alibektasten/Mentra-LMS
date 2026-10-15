package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const tenantTestSecret = "tenant-test-secret"

func TestTenant_RejectsHeaderForAnotherOrganization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgA, orgB, userID := uuid.New(), uuid.New(), uuid.New()
	orgRepo := new(mocks.MockOrganizationRepository)
	orgRepo.On("FindBySlug", mock.Anything, "org-b").Return(&entities.Organization{ID: orgB, Slug: "org-b"}, nil)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/protected", Auth(tenantTestSecret), Tenant(orgRepo), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+signedTenantToken(t, userID, orgA, entities.RoleAdmin))
	request.Header.Set("X-Org-Slug", "org-b")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestTenantWithMembership_RejectsRemovedMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, userID := uuid.New(), uuid.New()
	orgRepo := new(mocks.MockOrganizationRepository)
	membershipRepo := new(mocks.MockMembershipRepository)
	orgRepo.On("FindBySlug", mock.Anything, "org-a").Return(&entities.Organization{ID: orgID, Slug: "org-a"}, nil)
	membershipRepo.On("FindUserRole", mock.Anything, userID, orgID).Return(entities.Role(""), apperrors.NotFoundError("membership", userID.String()))

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/protected", Auth(tenantTestSecret), TenantWithMembership(orgRepo, membershipRepo), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+signedTenantToken(t, userID, orgID, entities.RoleTeacher))
	request.Header.Set("X-Org-Slug", "org-a")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestTenantWithMembership_RejectsStaleRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, userID := uuid.New(), uuid.New()
	orgRepo := new(mocks.MockOrganizationRepository)
	membershipRepo := new(mocks.MockMembershipRepository)
	orgRepo.On("FindBySlug", mock.Anything, "org-a").Return(&entities.Organization{ID: orgID, Slug: "org-a"}, nil)
	membershipRepo.On("FindUserRole", mock.Anything, userID, orgID).Return(entities.RoleStudent, nil)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/protected", Auth(tenantTestSecret), TenantWithMembership(orgRepo, membershipRepo), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+signedTenantToken(t, userID, orgID, entities.RoleAdmin))
	request.Header.Set("X-Org-Slug", "org-a")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusForbidden, response.Code)
}

func TestTenantWithMembership_AllowsCurrentMembership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, userID := uuid.New(), uuid.New()
	orgRepo := new(mocks.MockOrganizationRepository)
	membershipRepo := new(mocks.MockMembershipRepository)
	orgRepo.On("FindBySlug", mock.Anything, "org-a").Return(&entities.Organization{ID: orgID, Slug: "org-a"}, nil)
	membershipRepo.On("FindUserRole", mock.Anything, userID, orgID).Return(entities.RoleTeacher, nil)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/protected", Auth(tenantTestSecret), TenantWithMembership(orgRepo, membershipRepo), func(c *gin.Context) {
		assert.Equal(t, orgID.String(), GetOrgID(c))
		assert.Equal(t, string(entities.RoleTeacher), GetRole(c))
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+signedTenantToken(t, userID, orgID, entities.RoleTeacher))
	request.Header.Set("X-Org-Slug", "org-a")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestAuth_RejectsMissingOrMalformedIdentityClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	claims := jwt.MapClaims{"user_id": "not-a-uuid", "org_id": uuid.NewString(), "role": string(entities.RoleStudent), "exp": time.Now().Add(time.Minute).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(tenantTestSecret))
	assert.NoError(t, err)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/protected", Auth(tenantTestSecret), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+signed)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func signedTenantToken(t *testing.T, userID, orgID uuid.UUID, role entities.Role) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(), "org_id": orgID.String(), "role": string(role), "exp": time.Now().Add(time.Minute).Unix(),
	})
	signed, err := token.SignedString([]byte(tenantTestSecret))
	assert.NoError(t, err)
	return signed
}
