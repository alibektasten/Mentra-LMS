package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ailms/backend/internal/delivery/http/handlers"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/ailms/backend/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestRouterTenantChain(t *testing.T) {
	const secret = "router-security-test"
	orgA, orgB, userID := uuid.New(), uuid.New(), uuid.New()
	orgRepo := new(mocks.MockOrganizationRepository)
	members := new(mocks.MockMembershipRepository)
	orgRepo.On("FindBySlug", mock.Anything, "org-a").Return(&entities.Organization{ID: orgA}, nil)
	orgRepo.On("FindBySlug", mock.Anything, "org-b").Return(&entities.Organization{ID: orgB}, nil)
	members.On("FindUserRole", mock.Anything, userID, orgA).Return(entities.RoleStudent, nil).Once()
	members.On("FindUserRole", mock.Anything, userID, orgA).Return(entities.Role(""), apperrors.NotFoundError("membership", userID.String())).Once()
	members.On("FindUserRole", mock.Anything, userID, orgA).Return(entities.RoleTeacher, nil).Once()

	router := NewRouter(Dependencies{
		OrgRepo: orgRepo, MembershipRepo: members, JWTSecret: secret,
		UploadHandler: handlers.NewUploadHandler(nil),
		Logger:        logger.New(slog.LevelError, "text"),
	})
	request := func(slug string, role entities.Role, tokenOrg uuid.UUID) int {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID.String(), "org_id": tokenOrg.String(), "role": string(role), "exp": time.Now().Add(time.Hour).Unix(),
		})
		signed, err := token.SignedString([]byte(secret))
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", nil)
		req.Header.Set("Authorization", "Bearer "+signed)
		req.Header.Set("X-Org-Slug", slug)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response.Code
	}

	assert.Equal(t, http.StatusUnprocessableEntity, request("org-a", entities.RoleStudent, orgA), "current membership reaches upload handler")
	assert.Equal(t, http.StatusForbidden, request("org-b", entities.RoleStudent, orgA), "header cannot switch tenant")
	assert.Equal(t, http.StatusForbidden, request("org-a", entities.RoleStudent, orgA), "removed membership")
	assert.Equal(t, http.StatusForbidden, request("org-a", entities.RoleAdmin, orgA), "stale admin role")
	assert.Equal(t, http.StatusUnauthorized, request("org-a", entities.RoleSuperAdmin, uuid.Nil), "platform token cannot enter tenant routes")
	assert.Equal(t, http.StatusNotFound, func() int {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/uploads/anything.pdf", nil))
		return response.Code
	}(), "static uploads route must not exist")
}
