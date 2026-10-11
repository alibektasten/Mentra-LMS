package middleware

import (
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

func TestRequireAssignedCourse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teacherID := uuid.New()
	orgID := uuid.New()
	courseID := uuid.New()

	newRouter := func(courseTeachers *mocks.MockCourseTeacherRepository, groups *mocks.MockGroupRepository, courses *mocks.MockCourseRepository, role entities.Role) *gin.Engine {
		router := gin.New()
		router.Use(ErrorHandler())
		router.Use(func(c *gin.Context) {
			c.Set(string(ContextKeyUserID), teacherID.String())
			c.Set(string(ContextKeyOrgID), orgID.String())
			c.Set(string(ContextKeyRole), string(role))
			c.Next()
		})
		router.GET("/courses/:id", RequireAssignedCourse(courseTeachers, groups, courses), func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
		router.GET("/courses", RequireAssignedCourse(courseTeachers, groups, courses), func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
		return router
	}

	t.Run("allows an assigned teacher", func(t *testing.T) {
		courseTeachers := new(mocks.MockCourseTeacherRepository)
		groups := new(mocks.MockGroupRepository)
		courses := new(mocks.MockCourseRepository)
		courseTeachers.On("Exists", mock.Anything, courseID, teacherID, orgID).Return(true, nil).Once()
		request := httptest.NewRequest(http.MethodGet, "/courses/"+courseID.String(), nil)
		response := httptest.NewRecorder()

		newRouter(courseTeachers, groups, courses, entities.RoleTeacher).ServeHTTP(response, request)

		assert.Equal(t, http.StatusNoContent, response.Code)
		courseTeachers.AssertExpectations(t)
	})

	t.Run("denies a teacher who is not assigned", func(t *testing.T) {
		courseTeachers := new(mocks.MockCourseTeacherRepository)
		groups := new(mocks.MockGroupRepository)
		courses := new(mocks.MockCourseRepository)
		courseTeachers.On("Exists", mock.Anything, courseID, teacherID, orgID).Return(false, nil).Once()
		request := httptest.NewRequest(http.MethodGet, "/courses/"+courseID.String(), nil)
		response := httptest.NewRecorder()

		newRouter(courseTeachers, groups, courses, entities.RoleTeacher).ServeHTTP(response, request)

		assert.Equal(t, http.StatusForbidden, response.Code)
		courseTeachers.AssertExpectations(t)
	})

	t.Run("allows an enrolled student in a published course", func(t *testing.T) {
		courseTeachers := new(mocks.MockCourseTeacherRepository)
		groups := new(mocks.MockGroupRepository)
		courses := new(mocks.MockCourseRepository)
		groups.On("GetStudentGroup", mock.Anything, courseID, teacherID, orgID).
			Return(&entities.Group{ID: uuid.New(), CourseID: &courseID}, nil).Once()
		courses.On("FindByID", mock.Anything, courseID, orgID).
			Return(&entities.Course{ID: courseID, OrgID: orgID, Status: entities.StatusPublished}, nil).Once()

		request := httptest.NewRequest(http.MethodGet, "/courses/"+courseID.String(), nil)
		response := httptest.NewRecorder()
		newRouter(courseTeachers, groups, courses, entities.RoleStudent).ServeHTTP(response, request)

		assert.Equal(t, http.StatusNoContent, response.Code)
		groups.AssertExpectations(t)
		courses.AssertExpectations(t)
	})

	t.Run("denies a student who is not enrolled", func(t *testing.T) {
		courseTeachers := new(mocks.MockCourseTeacherRepository)
		groups := new(mocks.MockGroupRepository)
		courses := new(mocks.MockCourseRepository)
		groups.On("GetStudentGroup", mock.Anything, courseID, teacherID, orgID).
			Return(nil, apperrors.NotFoundError("group", teacherID.String())).Once()

		request := httptest.NewRequest(http.MethodGet, "/courses/"+courseID.String(), nil)
		response := httptest.NewRecorder()
		newRouter(courseTeachers, groups, courses, entities.RoleStudent).ServeHTTP(response, request)

		assert.Equal(t, http.StatusForbidden, response.Code)
		groups.AssertExpectations(t)
		courses.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("denies a student access to a draft course", func(t *testing.T) {
		courseTeachers := new(mocks.MockCourseTeacherRepository)
		groups := new(mocks.MockGroupRepository)
		courses := new(mocks.MockCourseRepository)
		groups.On("GetStudentGroup", mock.Anything, courseID, teacherID, orgID).
			Return(&entities.Group{ID: uuid.New(), CourseID: &courseID}, nil).Once()
		courses.On("FindByID", mock.Anything, courseID, orgID).
			Return(&entities.Course{ID: courseID, OrgID: orgID, Status: entities.StatusDraft}, nil).Once()

		request := httptest.NewRequest(http.MethodGet, "/courses/"+courseID.String(), nil)
		response := httptest.NewRecorder()
		newRouter(courseTeachers, groups, courses, entities.RoleStudent).ServeHTTP(response, request)

		assert.Equal(t, http.StatusForbidden, response.Code)
		groups.AssertExpectations(t)
		courses.AssertExpectations(t)
	})

	t.Run("does not restrict administrators or course lists", func(t *testing.T) {
		courseTeachers := new(mocks.MockCourseTeacherRepository)
		groups := new(mocks.MockGroupRepository)
		courses := new(mocks.MockCourseRepository)
		adminRequest := httptest.NewRequest(http.MethodGet, "/courses/"+courseID.String(), nil)
		adminResponse := httptest.NewRecorder()
		newRouter(courseTeachers, groups, courses, entities.RoleAdmin).ServeHTTP(adminResponse, adminRequest)
		assert.Equal(t, http.StatusNoContent, adminResponse.Code)

		listRequest := httptest.NewRequest(http.MethodGet, "/courses", nil)
		listResponse := httptest.NewRecorder()
		newRouter(courseTeachers, groups, courses, entities.RoleTeacher).ServeHTTP(listResponse, listRequest)
		assert.Equal(t, http.StatusNoContent, listResponse.Code)
		courseTeachers.AssertNotCalled(t, "Exists", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestRequireAssignedModule_DeniesForeignCourseMaterial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	teacherID := uuid.New()
	orgID := uuid.New()
	moduleID := uuid.New()
	foreignCourseID := uuid.New()
	courseTeachers := new(mocks.MockCourseTeacherRepository)
	groups := new(mocks.MockGroupRepository)
	courses := new(mocks.MockCourseRepository)
	modules := new(mocks.MockModuleRepository)

	modules.On("FindByID", mock.Anything, moduleID, orgID).
		Return(&entities.Module{ID: moduleID, CourseID: foreignCourseID, OrgID: orgID}, nil).
		Once()
	courseTeachers.On("Exists", mock.Anything, foreignCourseID, teacherID, orgID).
		Return(false, nil).
		Once()

	router := gin.New()
	router.Use(ErrorHandler())
	router.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserID), teacherID.String())
		c.Set(string(ContextKeyOrgID), orgID.String())
		c.Set(string(ContextKeyRole), string(entities.RoleTeacher))
		c.Next()
	})
	router.GET("/modules/:moduleID/lessons",
		RequireAssignedModule(courseTeachers, groups, courses, modules, "moduleID"),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/modules/"+moduleID.String()+"/lessons", nil)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusForbidden, response.Code)
	modules.AssertExpectations(t)
	courseTeachers.AssertExpectations(t)
}

func TestRequireAssignedLesson_DeniesUnenrolledStudent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	studentID := uuid.New()
	orgID := uuid.New()
	lessonID := uuid.New()
	moduleID := uuid.New()
	courseID := uuid.New()
	courseTeachers := new(mocks.MockCourseTeacherRepository)
	groups := new(mocks.MockGroupRepository)
	courses := new(mocks.MockCourseRepository)
	lessons := new(mocks.MockLessonRepository)
	modules := new(mocks.MockModuleRepository)

	lessons.On("FindByID", mock.Anything, lessonID, orgID).
		Return(&entities.Lesson{ID: lessonID, ModuleID: moduleID, OrgID: orgID}, nil).Once()
	modules.On("FindByID", mock.Anything, moduleID, orgID).
		Return(&entities.Module{ID: moduleID, CourseID: courseID, OrgID: orgID}, nil).Once()
	groups.On("GetStudentGroup", mock.Anything, courseID, studentID, orgID).
		Return(nil, apperrors.NotFoundError("group", studentID.String())).Once()

	router := gin.New()
	router.Use(ErrorHandler())
	router.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserID), studentID.String())
		c.Set(string(ContextKeyOrgID), orgID.String())
		c.Set(string(ContextKeyRole), string(entities.RoleStudent))
		c.Next()
	})
	router.GET("/lessons/:lessonID/quiz",
		RequireAssignedLesson(courseTeachers, groups, courses, lessons, modules, "lessonID"),
		func(c *gin.Context) { c.Status(http.StatusNoContent) },
	)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/lessons/"+lessonID.String()+"/quiz", nil)
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusForbidden, response.Code)
	lessons.AssertExpectations(t)
	modules.AssertExpectations(t)
	groups.AssertExpectations(t)
	courses.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything, mock.Anything)
}
