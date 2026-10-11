package handlers

import (
	"net/http"

	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CertificateHandler struct{ certificateUC *usecases.CertificateUseCase }

func NewCertificateHandler(certificateUC *usecases.CertificateUseCase) *CertificateHandler {
	return &CertificateHandler{certificateUC: certificateUC}
}

func (h *CertificateHandler) ListMine(c *gin.Context) {
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
	result, err := h.certificateUC.ListForStudent(c.Request.Context(), userID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *CertificateHandler) Verify(c *gin.Context) {
	result, err := h.certificateUC.Verify(c.Request.Context(), c.Param("code"))
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
