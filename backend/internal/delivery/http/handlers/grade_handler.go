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

type GradeHandler struct {
	gradeUC *usecases.GradeUseCase
}

func NewGradeHandler(gradeUC *usecases.GradeUseCase) *GradeHandler {
	return &GradeHandler{gradeUC: gradeUC}
}

func (h *GradeHandler) GetMyGrades(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	studentID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.gradeUC.GetPublishedMyGrades(c.Request.Context(), courseID, studentID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *GradeHandler) GetUpcomingDeadlines(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	studentID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.gradeUC.GetUpcomingDeadlines(c.Request.Context(), courseID, studentID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *GradeHandler) GetGradebook(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.gradeUC.GetGradebook(c.Request.Context(), courseID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *GradeHandler) UpdateGradebookStudent(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	studentID, err := uuid.Parse(c.Param("studentID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid student id"))
		return
	}
	var req dto.UpdateCourseGradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.ValidationError("invalid gradebook update"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	actorID, _ := uuid.Parse(middleware.GetUserID(c))
	result, err := h.gradeUC.UpdateCourseGrade(c.Request.Context(), courseID, studentID, orgID, actorID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *GradeHandler) SetGradebookPublication(c *gin.Context) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	var req dto.SetGradebookPublicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.ValidationError("invalid publication state"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	actorID, _ := uuid.Parse(middleware.GetUserID(c))
	studentIDs := make([]uuid.UUID, 0, len(req.StudentIDs))
	for _, rawID := range req.StudentIDs {
		studentID, err := uuid.Parse(rawID)
		if err != nil {
			c.Error(apperrors.ValidationError("invalid student id"))
			return
		}
		studentIDs = append(studentIDs, studentID)
	}
	result, err := h.gradeUC.SetGradebookPublication(c.Request.Context(), courseID, orgID, actorID, req.Published, studentIDs)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
