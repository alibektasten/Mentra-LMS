package handlers

import (
	"errors"
	"net/http"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/infrastructure/storage"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AssignmentHandler struct {
	assignmentUC *usecases.AssignmentUseCase
	storage      *storage.LocalStorage
}

func NewAssignmentHandler(assignmentUC *usecases.AssignmentUseCase, storage *storage.LocalStorage) *AssignmentHandler {
	return &AssignmentHandler{assignmentUC: assignmentUC, storage: storage}
}

func (h *AssignmentHandler) Create(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("moduleID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid module id"))
		return
	}
	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	var req dto.CreateAssignmentRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.assignmentUC.Create(c.Request.Context(), courseID, moduleID, orgID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *AssignmentHandler) ListByModule(c *gin.Context) {
	moduleID, err := uuid.Parse(c.Param("moduleID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid module id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.assignmentUC.GetByModule(c.Request.Context(), moduleID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AssignmentHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid assignment id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.assignmentUC.GetByID(c.Request.Context(), id, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AssignmentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid assignment id"))
		return
	}
	var req dto.UpdateAssignmentRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	result, err := h.assignmentUC.Update(c.Request.Context(), id, orgID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AssignmentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid assignment id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if err := h.assignmentUC.Delete(c.Request.Context(), id, orgID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *AssignmentHandler) Submit(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, storage.MaxUploadBytes+1024*1024)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid assignment id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	studentID, _ := uuid.Parse(middleware.GetUserID(c))

	textContent := c.PostForm("text_content")
	linkURL := c.PostForm("link_url")

	var filePath string
	file, header, fileErr := c.Request.FormFile("file")
	if fileErr == nil {
		defer file.Close()
		filePath, err = h.storage.Save(file, header, orgID.String())
		if err != nil {
			if errors.Is(err, storage.ErrInvalidUpload) {
				c.Error(apperrors.ValidationError(err.Error()))
				return
			}
			c.Error(apperrors.InternalError("failed to save file: " + err.Error()))
			return
		}
	}

	result, err := h.assignmentUC.Submit(c.Request.Context(), id, studentID, orgID, textContent, linkURL, filePath)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AssignmentHandler) GetMySubmission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid assignment id"))
		return
	}
	studentID, _ := uuid.Parse(middleware.GetUserID(c))
	result, err := h.assignmentUC.GetMySubmission(c.Request.Context(), id, studentID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AssignmentHandler) ListSubmissions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid assignment id"))
		return
	}
	result, err := h.assignmentUC.ListSubmissions(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *AssignmentHandler) DeleteMySubmission(c *gin.Context) {
	assignmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid assignment id"))
		return
	}
	studentID, _ := uuid.Parse(middleware.GetUserID(c))
	if err := h.assignmentUC.DeleteMySubmission(c.Request.Context(), assignmentID, studentID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *AssignmentHandler) GradeSubmission(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid submission id"))
		return
	}
	var req dto.GradeSubmissionRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	graderID, _ := uuid.Parse(middleware.GetUserID(c))
	result, err := h.assignmentUC.GradeSubmission(c.Request.Context(), id, graderID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
