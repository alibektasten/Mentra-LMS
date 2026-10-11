package handlers

import (
	"net/http"

	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeacherHandler struct{ teacherUC *usecases.TeacherUseCase }

func NewTeacherHandler(teacherUC *usecases.TeacherUseCase) *TeacherHandler {
	return &TeacherHandler{teacherUC: teacherUC}
}

func (h *TeacherHandler) Overview(c *gin.Context) {
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	data, err := h.teacherUC.GetOverview(c.Request.Context(), userID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *TeacherHandler) Students(c *gin.Context) {
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	data, err := h.teacherUC.ListStudents(c.Request.Context(), userID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func (h *TeacherHandler) ReviewQueue(c *gin.Context) {
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	data, err := h.teacherUC.ListReviewQueue(c.Request.Context(), userID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}
