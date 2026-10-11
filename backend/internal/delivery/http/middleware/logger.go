package middleware

import (
	"time"

	"github.com/ailms/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Header("X-Request-ID", requestID)
		c.Next()
		log.Info("request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_id", GetUserID(c),
			"org_id", GetOrgID(c),
		)
	}
}
