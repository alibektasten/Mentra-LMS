package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequireOptionalProgressCourseAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orgID, studentID, courseID := uuid.New(), uuid.New(), uuid.New()
	newRouter := func(courses *mocks.MockCourseRepository, groups *mocks.MockGroupRepository) *gin.Engine {
		router := gin.New()
		router.Use(ErrorHandler())
		router.Use(func(c *gin.Context) {
			c.Set(string(ContextKeyOrgID), orgID.String())
			c.Set(string(ContextKeyUserID), studentID.String())
			c.Set(string(ContextKeyRole), string(entities.RoleStudent))
			c.Next()
		})
		router.GET("/progress", RequireOptionalProgressCourseAccess(new(mocks.MockCourseTeacherRepository), groups, courses), func(c *gin.Context) { c.Status(http.StatusNoContent) })
		return router
	}
	request := func(router *gin.Engine, query string) int {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/progress"+query, nil))
		return response.Code
	}

	t.Run("legacy unfiltered progress remains available", func(t *testing.T) {
		assert.Equal(t, http.StatusNoContent, request(newRouter(new(mocks.MockCourseRepository), new(mocks.MockGroupRepository)), ""))
	})
	t.Run("malformed and empty course IDs are rejected", func(t *testing.T) {
		router := newRouter(new(mocks.MockCourseRepository), new(mocks.MockGroupRepository))
		assert.Equal(t, http.StatusUnprocessableEntity, request(router, "?course_id=bad"))
		assert.Equal(t, http.StatusUnprocessableEntity, request(router, "?course_id="))
	})
	t.Run("enrolled student may read published course", func(t *testing.T) {
		courses, groups := new(mocks.MockCourseRepository), new(mocks.MockGroupRepository)
		course := &entities.Course{ID: courseID, OrgID: orgID, Status: entities.StatusPublished}
		courses.On("FindByID", mock.Anything, courseID, orgID).Return(course, nil).Twice()
		groups.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).Return(&entities.Group{ID: uuid.New()}, nil).Once()
		assert.Equal(t, http.StatusNoContent, request(newRouter(courses, groups), "?course_id="+courseID.String()))
		courses.AssertExpectations(t)
		groups.AssertExpectations(t)
	})
	t.Run("another tenant's course is not exposed", func(t *testing.T) {
		courses := new(mocks.MockCourseRepository)
		courses.On("FindByID", mock.Anything, courseID, orgID).Return(nil, apperrors.NotFoundError("course", courseID.String())).Once()
		assert.Equal(t, http.StatusNotFound, request(newRouter(courses, new(mocks.MockGroupRepository)), "?course_id="+courseID.String()))
	})
	t.Run("unenrolled and draft courses are denied", func(t *testing.T) {
		courses, groups := new(mocks.MockCourseRepository), new(mocks.MockGroupRepository)
		courses.On("FindByID", mock.Anything, courseID, orgID).Return(&entities.Course{ID: courseID, OrgID: orgID, Status: entities.StatusPublished}, nil).Once()
		groups.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).Return(nil, apperrors.NotFoundError("group", studentID.String())).Once()
		assert.Equal(t, http.StatusForbidden, request(newRouter(courses, groups), "?course_id="+courseID.String()))

		courses, groups = new(mocks.MockCourseRepository), new(mocks.MockGroupRepository)
		courses.On("FindByID", mock.Anything, courseID, orgID).Return(&entities.Course{ID: courseID, OrgID: orgID, Status: entities.StatusDraft}, nil).Twice()
		groups.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).Return(&entities.Group{ID: uuid.New()}, nil).Once()
		assert.Equal(t, http.StatusForbidden, request(newRouter(courses, groups), "?course_id="+courseID.String()))
	})
	t.Run("repository failure is not treated as an empty course", func(t *testing.T) {
		courses := new(mocks.MockCourseRepository)
		courses.On("FindByID", mock.Anything, courseID, orgID).Return(nil, apperrors.InternalError(errors.New("db unavailable").Error())).Once()
		assert.Equal(t, http.StatusInternalServerError, request(newRouter(courses, new(mocks.MockGroupRepository)), "?course_id="+courseID.String()))
	})
}
