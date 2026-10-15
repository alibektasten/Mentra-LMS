package middleware

import (
	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/domain/repositories"
	"github.com/ailms/backend/internal/domain/services"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Tenant(orgRepo repositories.OrganizationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !validateTenantHeader(c, orgRepo) {
			return
		}
		c.Next()
	}
}

// TenantWithMembership additionally verifies membership and role on every
// protected tenant request. Router construction must use this variant so a
// deleted membership or a role downgrade takes effect before JWT expiry.
func TenantWithMembership(orgRepo repositories.OrganizationRepository, memberships repositories.MembershipRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !validateTenantHeader(c, orgRepo) {
			return
		}
		userID, ok := requestTokenUserID(c)
		if !ok {
			return
		}
		orgID, ok := requestTokenOrgID(c)
		if !ok {
			return
		}
		role, err := memberships.FindUserRole(c.Request.Context(), userID, orgID)
		if err != nil {
			c.Error(apperrors.ForbiddenError("you are no longer a member of this organization"))
			c.Abort()
			return
		}
		if string(role) != GetRole(c) {
			c.Error(apperrors.ForbiddenError("your organization role has changed; sign in again"))
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(services.WithAIIdentity(c.Request.Context(), orgID, userID))
		c.Next()
	}
}

func validateTenantHeader(c *gin.Context, orgRepo repositories.OrganizationRepository) bool {
	slug := c.GetHeader("X-Org-Slug")
	if slug == "" {
		c.Error(apperrors.ValidationError("X-Org-Slug header is required"))
		c.Abort()
		return false
	}

	org, err := orgRepo.FindBySlug(c.Request.Context(), slug)
	if err != nil {
		c.Error(apperrors.NotFoundError("organization", slug))
		c.Abort()
		return false
	}

	tokenOrgID, ok := requestTokenOrgID(c)
	if !ok {
		return false
	}
	if tokenOrgID != org.ID {
		c.Error(apperrors.ForbiddenError("token is not valid for the requested organization"))
		c.Abort()
		return false
	}
	// Tenant routes deliberately do not accept platform tokens. Super-admin
	// APIs live on their own Auth + RequireRole route group.
	if GetRole(c) == string(entities.RoleSuperAdmin) {
		c.Error(apperrors.ForbiddenError("super admin tokens cannot access tenant routes"))
		c.Abort()
		return false
	}
	// Preserve the token-bound organization rather than replacing it with a
	// user-controlled request header.
	return true
}

func requestTokenOrgID(c *gin.Context) (uuid.UUID, bool) {
	orgID, err := uuid.Parse(GetOrgID(c))
	if err != nil || orgID == uuid.Nil {
		c.Error(apperrors.UnauthorizedError("invalid organization identity"))
		c.Abort()
		return uuid.Nil, false
	}
	return orgID, true
}

func requestTokenUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, err := uuid.Parse(GetUserID(c))
	if err != nil || userID == uuid.Nil {
		c.Error(apperrors.UnauthorizedError("invalid user identity"))
		c.Abort()
		return uuid.Nil, false
	}
	return userID, true
}
