package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFileAttachmentHandler_ListByRef_RestrictsStudentToEnrolledCourse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	studentID := uuid.New()
	orgID := uuid.New()
	assignmentID := uuid.New()
	courseID := uuid.New()

	newRouter := func(enrolled bool) (*gin.Engine, *mocks.MockFileAttachmentRepository) {
		attachments := new(mocks.MockFileAttachmentRepository)
		courseTeachers := new(mocks.MockCourseTeacherRepository)
		groups := new(mocks.MockGroupRepository)
		courses := new(mocks.MockCourseRepository)
		lessons := new(mocks.MockLessonRepository)
		modules := new(mocks.MockModuleRepository)
		assignments := new(mocks.MockAssignmentRepository)

		assignments.On("FindByID", mock.Anything, assignmentID, orgID).
			Return(&entities.Assignment{ID: assignmentID, CourseID: courseID, OrgID: orgID}, nil).Once()
		if enrolled {
			groups.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).
				Return(&entities.Group{ID: uuid.New(), CourseID: &courseID, OrgID: orgID}, nil).Once()
			courses.On("FindByID", mock.Anything, courseID, orgID).
				Return(&entities.Course{ID: courseID, OrgID: orgID, Status: entities.StatusPublished}, nil).Once()
			attachments.On("FindByRef", mock.Anything, "assignment", assignmentID).
				Return([]*entities.FileAttachment{}, nil).Once()
		} else {
			groups.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).
				Return(nil, apperrors.NotFoundError("group", studentID.String())).Once()
		}

		handler := NewFileAttachmentHandler(
			usecases.NewFileAttachmentUseCase(attachments),
			courseTeachers, groups, courses, lessons, modules, assignments,
		)
		router := gin.New()
		router.Use(middleware.ErrorHandler())
		router.Use(func(c *gin.Context) {
			c.Set(string(middleware.ContextKeyUserID), studentID.String())
			c.Set(string(middleware.ContextKeyOrgID), orgID.String())
			c.Set(string(middleware.ContextKeyRole), string(entities.RoleStudent))
			c.Next()
		})
		router.GET("/attachments", handler.ListByRef)
		return router, attachments
	}

	t.Run("allows enrolled student", func(t *testing.T) {
		router, attachments := newRouter(true)
		request := httptest.NewRequest(http.MethodGet, "/attachments?ref_type=assignment&ref_id="+assignmentID.String(), nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		attachments.AssertExpectations(t)
	})

	t.Run("denies unenrolled student", func(t *testing.T) {
		router, attachments := newRouter(false)
		request := httptest.NewRequest(http.MethodGet, "/attachments?ref_type=assignment&ref_id="+assignmentID.String(), nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assert.Equal(t, http.StatusForbidden, response.Code)
		attachments.AssertNotCalled(t, "FindByRef", mock.Anything, mock.Anything, mock.Anything)
	})
}
