package handlers

import (
	"net/http"
	"path"
	"strings"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/infrastructure/storage"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LessonHandler struct {
	lessonUC *usecases.LessonUseCase
	storage  *storage.LocalStorage
}

func NewLessonHandler(lessonUC *usecases.LessonUseCase, store ...*storage.LocalStorage) *LessonHandler {
	h := &LessonHandler{lessonUC: lessonUC}
	if len(store) > 0 {
		h.storage = store[0]
	}
	return h
}

func (h *LessonHandler) authorizeNewUpload(c *gin.Context, path, token string) bool {
	if path == "" || !isLocalUploadReference(path) {
		return true
	}
	if h.storage == nil || h.storage.ValidateUploadClaim(middleware.GetOrgID(c), middleware.GetUserID(c), path, token) != nil {
		c.Error(apperrors.ForbiddenError("upload is not owned by this user or its proof has expired"))
		c.Abort()
		return false
	}
	return true
}

func isLocalUploadReference(value string) bool {
	// ResolvePrivatePath trims whitespace and cleans dot segments before
	// resolving the file. Classify the same effective path before deciding
	// whether an ownership claim is required; validation still uses the raw
	// value, so non-canonical aliases cannot reuse a canonical claim.
	cleaned := path.Clean(strings.TrimSpace(value))
	return strings.HasPrefix(cleaned, "uploads/") || strings.Contains(cleaned, "/uploads/")
}

func (h *LessonHandler) List(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("moduleID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid module id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	lessons, err := h.lessonUC.ListLessons(c.Request.Context(), moduleID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lessons})
}

func (h *LessonHandler) Create(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("moduleID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid module id"))
		return
	}
	var req dto.CreateLessonRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	if !h.authorizeNewUpload(c, req.FileURL, req.UploadToken) {
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	lesson, err := h.lessonUC.CreateLesson(c.Request.Context(), moduleID, orgID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": lesson})
}

func (h *LessonHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("lessonID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid lesson id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	lesson, err := h.lessonUC.GetLesson(c.Request.Context(), id, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lesson})
}

func (h *LessonHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("lessonID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid lesson id"))
		return
	}
	var req dto.UpdateLessonRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if req.FileURL != "" && isLocalUploadReference(req.FileURL) {
		current, err := h.lessonUC.GetLesson(c.Request.Context(), id, orgID)
		if err != nil {
			c.Error(err)
			return
		}
		if req.FileURL != current.FileURL && !h.authorizeNewUpload(c, req.FileURL, req.UploadToken) {
			return
		}
	}
	lesson, err := h.lessonUC.UpdateLesson(c.Request.Context(), id, orgID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lesson})
}

func (h *LessonHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("lessonID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid lesson id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if err := h.lessonUC.DeleteLesson(c.Request.Context(), id, orgID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
