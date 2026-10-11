package handlers

import (
	"net/http"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileAttachmentHandler struct {
	fileAttachmentUC *usecases.FileAttachmentUseCase
	courseTeachers   repositories.CourseTeacherRepository
	groups           repositories.GroupRepository
	courses          repositories.CourseRepository
	lessons          repositories.LessonRepository
	modules          repositories.ModuleRepository
	assignments      repositories.AssignmentRepository
}

func NewFileAttachmentHandler(
	fileAttachmentUC *usecases.FileAttachmentUseCase,
	courseTeachers repositories.CourseTeacherRepository,
	groups repositories.GroupRepository,
	courses repositories.CourseRepository,
	lessons repositories.LessonRepository,
	modules repositories.ModuleRepository,
	assignments repositories.AssignmentRepository,
) *FileAttachmentHandler {
	return &FileAttachmentHandler{
		fileAttachmentUC: fileAttachmentUC,
		courseTeachers:   courseTeachers,
		groups:           groups,
		courses:          courses,
		lessons:          lessons,
		modules:          modules,
		assignments:      assignments,
	}
}

func (h *FileAttachmentHandler) authorizeReference(c *gin.Context, refType string, refID uuid.UUID) bool {
	switch refType {
	case "lesson":
		return middleware.AuthorizeLessonAccess(c, h.courseTeachers, h.groups, h.courses, h.lessons, h.modules, refID)
	case "assignment":
		orgID, err := uuid.Parse(middleware.GetOrgID(c))
		if err != nil {
			c.Error(apperrors.UnauthorizedError("invalid organization identity"))
			c.Abort()
			return false
		}
		assignment, err := h.assignments.FindByID(c.Request.Context(), refID, orgID)
		if err != nil {
			c.Error(err)
			c.Abort()
			return false
		}
		return middleware.AuthorizeCourseAccess(c, h.courseTeachers, h.groups, h.courses, assignment.CourseID)
	default:
		c.Error(apperrors.ValidationError("unsupported attachment ref_type"))
		c.Abort()
		return false
	}
}

func (h *FileAttachmentHandler) Create(c *gin.Context) {
	var req dto.CreateAttachmentRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	refID, err := uuid.Parse(req.RefID)
	if err != nil {
		c.Error(apperrors.ValidationError("invalid ref_id"))
		return
	}
	if !h.authorizeReference(c, req.RefType, refID) {
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	uploaderID, _ := uuid.Parse(middleware.GetUserID(c))
	attachment, err := h.fileAttachmentUC.CreateAttachment(c.Request.Context(), orgID, uploaderID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": attachment})
}

func (h *FileAttachmentHandler) ListByRef(c *gin.Context) {
	refType := c.Query("ref_type")
	refIDStr := c.Query("ref_id")
	if refType == "" || refIDStr == "" {
		c.Error(apperrors.ValidationError("ref_type and ref_id query params are required"))
		return
	}
	refID, err := uuid.Parse(refIDStr)
	if err != nil {
		c.Error(apperrors.ValidationError("invalid ref_id"))
		return
	}
	if !h.authorizeReference(c, refType, refID) {
		return
	}
	attachments, err := h.fileAttachmentUC.ListByRef(c.Request.Context(), refType, refID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": attachments})
}

func (h *FileAttachmentHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid attachment id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	attachment, err := h.fileAttachmentUC.GetAttachment(c.Request.Context(), id, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	refID, err := uuid.Parse(attachment.RefID)
	if err != nil {
		c.Error(apperrors.ValidationError("invalid attachment reference"))
		return
	}
	if !h.authorizeReference(c, attachment.RefType, refID) {
		return
	}
	if err := h.fileAttachmentUC.DeleteAttachment(c.Request.Context(), id, orgID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
