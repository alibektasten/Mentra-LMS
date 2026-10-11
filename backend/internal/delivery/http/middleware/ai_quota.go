package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// DailyAIQuota prevents a single account from exhausting an organisation's AI
// budget. It is deliberately placed after Auth, so the quota is per account
// rather than per shared IP address. The in-memory counter is a safe default
// for one server; production multi-instance deployments should move it to
// Redis while keeping this route contract unchanged.
func DailyAIQuota(limit int) gin.HandlerFunc {
	if limit <= 0 {
		return func(c *gin.Context) { c.Next() }
	}

	type quotaEntry struct {
		day   string
		count int
	}
	var (
		entries = make(map[string]quotaEntry)
		mu      sync.Mutex
	)

	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "authentication required"}})
			return
		}

		today := time.Now().UTC().Format("2006-01-02")
		mu.Lock()
		entry := entries[userID]
		if entry.day != today {
			entry = quotaEntry{day: today}
		}
		if entry.count >= limit {
			mu.Unlock()
			c.Header("X-AI-Quota-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-AI-Quota-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": gin.H{
				"code":    "AI_DAILY_QUOTA_EXCEEDED",
				"message": "daily AI request limit reached; try again tomorrow",
			}})
			return
		}
		entry.count++
		entries[userID] = entry
		remaining := limit - entry.count
		mu.Unlock()

		c.Header("X-AI-Quota-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-AI-Quota-Remaining", fmt.Sprintf("%d", remaining))
		c.Next()
	}
}
