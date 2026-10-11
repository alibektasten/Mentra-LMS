package handlers

import (
	"net/http"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/domain/entities"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/ailms/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CourseHandler struct {
	courseUC *usecases.CourseUseCase
}

func NewCourseHandler(courseUC *usecases.CourseUseCase) *CourseHandler {
	return &CourseHandler{courseUC: courseUC}
}

func (h *CourseHandler) List(c *gin.Context) {
	orgID, err := uuid.Parse(middleware.GetOrgID(c))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid org id"))
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	role := middleware.GetRole(c)
	p := utils.ParsePagination(c)
	courses, total, err := h.courseUC.ListCoursesForRole(c.Request.Context(), userID, orgID, role, p.Page, p.PageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.PaginatedResponse[dto.CourseDTO]{
		Data: courses,
		Meta: utils.PaginationMeta{Page: p.Page, PageSize: p.PageSize, Total: total},
	})
}

func (h *CourseHandler) Create(c *gin.Context) {
	var req dto.CreateCourseRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	userID, _ := uuid.Parse(middleware.GetUserID(c))
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if req.Status == string(entities.StatusPublished) && middleware.GetRole(c) != string(entities.RoleAdmin) {
		c.Error(apperrors.ForbiddenError("only administrators can publish courses"))
		return
	}
	course, err := h.courseUC.CreateCourse(c.Request.Context(), req, userID, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": course})
}

func (h *CourseHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	course, err := h.courseUC.GetCourse(c.Request.Context(), id, orgID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": course})
}

func (h *CourseHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	var req dto.UpdateCourseRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if req.Status == string(entities.StatusPublished) && middleware.GetRole(c) != string(entities.RoleAdmin) {
		c.Error(apperrors.ForbiddenError("only administrators can publish courses"))
		return
	}
	course, err := h.courseUC.UpdateCourse(c.Request.Context(), id, orgID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": course})
}

func (h *CourseHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if err := h.courseUC.DeleteCourse(c.Request.Context(), id, orgID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *CourseHandler) Publish(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid course id"))
		return
	}
	orgID, _ := uuid.Parse(middleware.GetOrgID(c))
	if err := h.courseUC.PublishCourse(c.Request.Context(), id, orgID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "published"}})
}
