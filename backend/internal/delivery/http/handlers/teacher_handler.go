package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	apperrors "github.com/ailms/backend/pkg/errors"
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

// Calendar returns events from the teacher's assigned courses. Dates use the
// YYYY-MM-DD form and form an inclusive range, which makes month views simple
// for API clients to request.
func (h *TeacherHandler) Calendar(c *gin.Context) {
	from, to, err := calendarRange(c.Query("from"), c.Query("to"), time.Now())
	if err != nil {
		c.Error(err)
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	data, err := h.teacherUC.ListCalendar(c.Request.Context(), userID, orgID, from, to)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func calendarRange(fromValue, toValue string, now time.Time) (time.Time, time.Time, error) {
	const layout = "2006-01-02"
	fromValue, toValue = strings.TrimSpace(fromValue), strings.TrimSpace(toValue)
	if fromValue == "" && toValue == "" {
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return from, from.AddDate(0, 1, -1), nil
	}
	if fromValue == "" || toValue == "" {
		return time.Time{}, time.Time{}, apperrors.ValidationError("from and to must be supplied together")
	}
	from, err := time.Parse(layout, fromValue)
	if err != nil {
		return time.Time{}, time.Time{}, apperrors.ValidationError("from must use YYYY-MM-DD")
	}
	to, err := time.Parse(layout, toValue)
	if err != nil {
		return time.Time{}, time.Time{}, apperrors.ValidationError("to must use YYYY-MM-DD")
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, apperrors.ValidationError("to must not be before from")
	}
	if to.Sub(from) > 366*24*time.Hour {
		return time.Time{}, time.Time{}, apperrors.ValidationError("calendar range must not exceed 366 days")
	}
	return from, to, nil
}
