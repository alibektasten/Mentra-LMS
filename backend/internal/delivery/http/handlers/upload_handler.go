package handlers

import (
	"errors"
	"net/http"

	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/infrastructure/storage"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	storage *storage.LocalStorage
}

func NewUploadHandler(storage *storage.LocalStorage) *UploadHandler {
	return &UploadHandler{storage: storage}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	// Limit the stream before parsing multipart data so oversized uploads cannot
	// consume memory or disk while Gin processes the form.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, storage.MaxUploadBytes+1024*1024)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.Error(apperrors.ValidationError("file is required"))
		return
	}
	defer file.Close()

	orgID := middleware.GetOrgID(c)
	storedPath, err := h.storage.Save(file, header, orgID)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidUpload) {
			c.Error(apperrors.ValidationError(err.Error()))
			return
		}
		c.Error(apperrors.InternalError("failed to save file: " + err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"path": storedPath,
		"name": header.Filename,
	}})
}
