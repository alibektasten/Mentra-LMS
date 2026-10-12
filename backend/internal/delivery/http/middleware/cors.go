package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS handles cross-origin requests. Allowed origins are read from the
// CORS_ALLOWED_ORIGINS environment variable (comma-separated). Falls back to
// allowing localhost:3000 in development.
func CORS() gin.HandlerFunc {
	allowedOrigins := parseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		if isAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Org-Slug")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}
	var out []string
	for _, o := range strings.Split(raw, ",") {
		if s := strings.TrimSpace(o); s != "" {
			out = append(out, s)
		}
	}
	// Local browsers do not treat localhost and 127.0.0.1 as the same origin.
	// Permit both loopback spellings when either is explicitly configured.
	if strings.Contains(raw, "http://localhost:3000") && !strings.Contains(raw, "http://127.0.0.1:3000") {
		out = append(out, "http://127.0.0.1:3000")
	}
	return out
}

func isAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}
