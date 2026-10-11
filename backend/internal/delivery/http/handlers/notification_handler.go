package handlers

import (
	"net/http"

	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct{ notificationUC *usecases.NotificationUseCase }

func NewNotificationHandler(notificationUC *usecases.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{notificationUC: notificationUC}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, err := uuid.Parse(middleware.GetUserID(c))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid user id"))
		return
	}
	orgID, err := uuid.Parse(middleware.GetOrgID(c))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid organization id"))
		return
	}
	result, err := h.notificationUC.List(c.Request.Context(), userID, orgID, middleware.GetRole(c))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
