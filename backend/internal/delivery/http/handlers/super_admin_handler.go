package handlers

import (
	"net/http"
	"strconv"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/domain/repositories"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SuperAdminHandler struct {
	superAdminUC *usecases.SuperAdminUseCase
	authUC       *usecases.AuthUseCase
}

func NewSuperAdminHandler(superAdminUC *usecases.SuperAdminUseCase, authUC *usecases.AuthUseCase) *SuperAdminHandler {
	return &SuperAdminHandler{superAdminUC: superAdminUC, authUC: authUC}
}

func (h *SuperAdminHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.authUC.SuperAdminLogin(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *SuperAdminHandler) GetStats(c *gin.Context) {
	stats, err := h.superAdminUC.GetStats(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func (h *SuperAdminHandler) ListOrgs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orgs, total, err := h.superAdminUC.ListOrgs(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": orgs,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

func (h *SuperAdminHandler) CreateOrg(c *gin.Context) {
	var req dto.CreateOrganizationRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	result, err := h.superAdminUC.CreateOrg(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": result})
}

func (h *SuperAdminHandler) UpdateOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid org id"))
		return
	}
	var req dto.UpdateOrganizationRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	result, err := h.superAdminUC.UpdateOrg(c.Request.Context(), id, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *SuperAdminHandler) DeleteOrg(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid org id"))
		return
	}

	if err := h.superAdminUC.DeleteOrg(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deleted": true}})
}

func (h *SuperAdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := repositories.SystemUserFilters{
		Page: page, PageSize: pageSize,
		Search: c.Query("search"), Role: c.Query("role"), OrgID: c.Query("org_id"),
		SortBy: c.DefaultQuery("sort_by", "created_at"), SortOrder: c.DefaultQuery("sort_order", "desc"),
	}
	users, total, err := h.superAdminUC.ListUsers(c.Request.Context(), filters)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"meta": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

func (h *SuperAdminHandler) UpdateUserRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("membershipID"))
	if err != nil {
		c.Error(apperrors.ValidationError("invalid membership id"))
		return
	}
	var req dto.UpdateSystemUserRoleRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}
	if err := h.superAdminUC.UpdateUserRole(c.Request.Context(), id, req); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"updated": true}})
}

func (h *SuperAdminHandler) InviteOrgAdmin(c *gin.Context) {
	var req dto.InviteOrgAdminRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	result, err := h.superAdminUC.InviteOrgAdmin(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}
