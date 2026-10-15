package middleware

import (
	"strings"

	"github.com/ailms/backend/internal/domain/entities"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const (
	ContextKeyUserID contextKey = "user_id"
	ContextKeyOrgID  contextKey = "org_id"
	ContextKeyRole   contextKey = "role"
)

func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(apperrors.UnauthorizedError("missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Error(apperrors.UnauthorizedError("invalid authorization header format"))
			c.Abort()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.UnauthorizedError("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.Error(apperrors.UnauthorizedError("invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !validTokenClaims(claims) {
			c.Error(apperrors.UnauthorizedError("invalid token claims"))
			c.Abort()
			return
		}

		c.Set(string(ContextKeyUserID), claims["user_id"])
		c.Set(string(ContextKeyOrgID), claims["org_id"])
		c.Set(string(ContextKeyRole), claims["role"])
		c.Next()
	}
}

// validTokenClaims makes the identity put into the request context explicit.
// MapClaims values are untrusted input even after the token signature has been
// checked, so accepting empty or wrong-typed claims could turn into a tenant
// or role bypass farther down the handler chain.
func validTokenClaims(claims jwt.MapClaims) bool {
	if _, ok := claims["exp"].(float64); !ok {
		return false
	}
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return false
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil || parsedUserID == uuid.Nil {
		return false
	}

	orgID, ok := claims["org_id"].(string)
	if !ok || orgID == "" {
		return false
	}
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return false
	}

	role, ok := claims["role"].(string)
	if !ok || !isKnownRole(role) {
		return false
	}
	// A platform token must never be accepted by a tenant route, and tenant
	// tokens must always be bound to a concrete organization.
	if role == string(entities.RoleSuperAdmin) {
		return parsedOrgID == uuid.Nil
	}
	return parsedOrgID != uuid.Nil
}

func isKnownRole(role string) bool {
	switch entities.Role(role) {
	case entities.RoleSuperAdmin, entities.RoleAdmin, entities.RoleTeacher, entities.RoleStudent:
		return true
	default:
		return false
	}
}

func GetUserID(c *gin.Context) string {
	v, _ := c.Get(string(ContextKeyUserID))
	s, _ := v.(string)
	return s
}

func GetOrgID(c *gin.Context) string {
	v, _ := c.Get(string(ContextKeyOrgID))
	s, _ := v.(string)
	return s
}

func GetRole(c *gin.Context) string {
	v, _ := c.Get(string(ContextKeyRole))
	s, _ := v.(string)
	return s
}
