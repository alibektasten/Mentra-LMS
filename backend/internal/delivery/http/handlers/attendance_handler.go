package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttendanceHandler struct{ attendanceUC *usecases.AttendanceUseCase }

func NewAttendanceHandler(attendanceUC *usecases.AttendanceUseCase) *AttendanceHandler {
	return &AttendanceHandler{attendanceUC: attendanceUC}
}

func (h *AttendanceHandler) GetSheet(c *gin.Context) {
	courseID, groupID, orgID, err := attendanceIDs(c)
	if err != nil {
		c.Error(err)
		return
	}
	date := time.Now()
	if value := strings.TrimSpace(c.Query("date")); value != "" {
		date, err = time.Parse("2006-01-02", value)
		if err != nil {
			c.Error(apperrors.ValidationError("date must use YYYY-MM-DD"))
			return
		}
	}
	result, err := h.attendanceUC.GetSheet(c.Request.Context(), courseID, groupID, orgID, date)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AttendanceHandler) SaveSheet(c *gin.Context) {
	courseID, groupID, orgID, err := attendanceIDs(c)
	if err != nil {
		c.Error(err)
		return
	}
	var req dto.SaveAttendanceRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	markerID, _ := uuid.Parse(middleware.GetUserID(c))
	result, err := h.attendanceUC.SaveSheet(c.Request.Context(), courseID, groupID, orgID, markerID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func attendanceIDs(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, apperrors.ValidationError("invalid course id")
	}
	groupID, err := uuid.Parse(c.Param("groupID"))
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, apperrors.ValidationError("invalid group id")
	}
	orgID, err := uuid.Parse(middleware.GetOrgID(c))
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, apperrors.ValidationError("invalid organization id")
	}
	return courseID, groupID, orgID, nil
}
