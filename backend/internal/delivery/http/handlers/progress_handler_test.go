package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/mocks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProgressHandlerSelectsScopedOrLegacyCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID, orgID, courseID := uuid.New(), uuid.New(), uuid.New()
	progressRepo := new(mocks.MockLessonProgressRepository)
	uc := usecases.NewProgressUseCase(progressRepo, new(mocks.MockLessonRepository), new(mocks.MockAIService), new(mocks.MockModuleRepository), new(mocks.MockQuizRepository), new(mocks.MockQuizAttemptRepository))
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUserID), userID.String())
		c.Set(string(middleware.ContextKeyOrgID), orgID.String())
		c.Next()
	})
	router.GET("/progress", NewProgressHandler(uc).GetProgress)
	request := func(path string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		return response
	}

	progressRepo.On("FindByUser", mock.Anything, userID, orgID).Return([]entities.LessonProgress{{ID: uuid.New(), LessonID: uuid.New(), UserID: userID, OrgID: orgID}}, nil).Once()
	global := request("/progress")
	assert.Equal(t, http.StatusOK, global.Code)
	assert.Contains(t, global.Body.String(), "lesson_id")

	progressRepo.On("FindByUserAndCourse", mock.Anything, userID, courseID, orgID).Return([]entities.LessonProgress{}, nil).Once()
	scoped := request("/progress?course_id=" + courseID.String())
	assert.Equal(t, http.StatusOK, scoped.Code)
	assert.JSONEq(t, `{"data":[]}`, scoped.Body.String())
	assert.Equal(t, http.StatusUnprocessableEntity, request("/progress?course_id=bad").Code)

	progressRepo.On("FindByUserAndCourse", mock.Anything, userID, courseID, orgID).Return(nil, errors.New("database unavailable")).Once()
	assert.Equal(t, http.StatusInternalServerError, request("/progress?course_id="+courseID.String()).Code)
	progressRepo.AssertExpectations(t)
}
