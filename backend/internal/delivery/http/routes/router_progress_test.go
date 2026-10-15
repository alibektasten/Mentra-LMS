package routes

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ailms/backend/internal/application/usecases"
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
)

func TestRouterProgressCourseFilterAndAccess(t *testing.T) {
	const secret = "progress-route-test"
	orgID, studentID, courseID, foreignCourseID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	orgs := new(mocks.MockOrganizationRepository)
	members := new(mocks.MockMembershipRepository)
	courses := new(mocks.MockCourseRepository)
	groups := new(mocks.MockGroupRepository)
	progress := new(mocks.MockLessonProgressRepository)
	orgs.On("FindBySlug", mock.Anything, "nu").Return(&entities.Organization{ID: orgID}, nil)
	members.On("FindUserRole", mock.Anything, studentID, orgID).Return(entities.RoleStudent, nil)
	courses.On("FindByID", mock.Anything, courseID, orgID).Return(&entities.Course{ID: courseID, OrgID: orgID, Status: entities.StatusPublished}, nil)
	courses.On("FindByID", mock.Anything, foreignCourseID, orgID).Return(nil, apperrors.NotFoundError("course", foreignCourseID.String()))
	groups.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).Return(&entities.Group{ID: uuid.New()}, nil)
	progress.On("FindByUserAndCourse", mock.Anything, studentID, courseID, orgID).Return([]entities.LessonProgress{}, nil).Once()
	progress.On("FindByUser", mock.Anything, studentID, orgID).Return([]entities.LessonProgress{{ID: uuid.New(), UserID: studentID, OrgID: orgID, LessonID: uuid.New()}}, nil).Once()
	progressUC := usecases.NewProgressUseCase(progress, new(mocks.MockLessonRepository), new(mocks.MockAIService), new(mocks.MockModuleRepository), new(mocks.MockQuizRepository), new(mocks.MockQuizAttemptRepository))
	router := NewRouter(Dependencies{
		OrgRepo: orgs, MembershipRepo: members, CourseRepo: courses, GroupRepo: groups,
		CourseTeacherRepo: new(mocks.MockCourseTeacherRepository), ProgressHandler: handlers.NewProgressHandler(progressUC),
		JWTSecret: secret, Logger: logger.New(slog.LevelError, "text"),
	})
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": studentID.String(), "org_id": orgID.String(), "role": string(entities.RoleStudent), "exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	request := func(query string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/progress"+query, nil)
		req.Header.Set("Authorization", "Bearer "+signed)
		req.Header.Set("X-Org-Slug", "nu")
		router.ServeHTTP(response, req)
		return response
	}
	assert.JSONEq(t, `{"data":[]}`, request("?course_id="+courseID.String()).Body.String())
	assert.Equal(t, http.StatusOK, request("").Code, "legacy dashboard collection remains available")
	assert.Equal(t, http.StatusUnprocessableEntity, request("?course_id=invalid").Code)
	assert.Equal(t, http.StatusNotFound, request("?course_id="+foreignCourseID.String()).Code)
	progress.AssertExpectations(t)
}
