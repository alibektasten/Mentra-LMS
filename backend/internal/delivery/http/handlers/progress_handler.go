package handlers

import (
	"net/http"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProgressHandler struct {
	progressUC *usecases.ProgressUseCase
}

func NewProgressHandler(progressUC *usecases.ProgressUseCase) *ProgressHandler {
	return &ProgressHandler{progressUC: progressUC}
}

func (h *ProgressHandler) Complete(c *gin.Context) {
	lessonID, err := uuid.Parse(c.Param("lessonID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid lesson id"))
		return
	}
	var req dto.CompleteLessonRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if err := h.progressUC.CompleteLesson(c.Request.Context(), userID, lessonID, orgID, req.Score); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "completed"}})
}

func (h *ProgressHandler) GetProgress(c *gin.Context) {
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	var (
		progresses []dto.ProgressDTO
		err        error
	)
	if rawCourseID, present := c.GetQuery("course_id"); present {
		courseID, parseErr := uuid.Parse(rawCourseID)
		if parseErr != nil || courseID == uuid.Nil {
			c.Error(apperrors.ValidationError("invalid course id"))
			return
		}
		progresses, err = h.progressUC.GetStudentCourseProgress(c.Request.Context(), userID, courseID, orgID)
	} else {
		progresses, err = h.progressUC.GetStudentProgress(c.Request.Context(), userID, orgID)
	}
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": progresses})
}

func (h *ProgressHandler) GetInsights(c *gin.Context) {
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	insights, err := h.progressUC.GetProgressInsights(c.Request.Context(), userID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": insights})
}

func (h *ProgressHandler) GetCoursePacing(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	pacing, err := h.progressUC.GetCoursePacing(c.Request.Context(), courseID, userID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pacing})
}
