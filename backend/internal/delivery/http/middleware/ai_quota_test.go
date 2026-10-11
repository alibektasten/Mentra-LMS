package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDailyAIQuotaIsPerUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(ContextKeyUserID), c.GetHeader("X-Test-User"))
		c.Next()
	})
	router.Use(DailyAIQuota(1))
	router.GET("/ai", func(c *gin.Context) { c.Status(http.StatusOK) })

	request := func(user string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/ai", nil)
		r.Header.Set("X-Test-User", user)
		router.ServeHTTP(w, r)
		return w
	}

	first := request("student-a")
	if first.Code != http.StatusOK || first.Header().Get("X-AI-Quota-Remaining") != "0" {
		t.Fatalf("first request = status %d, remaining %q", first.Code, first.Header().Get("X-AI-Quota-Remaining"))
	}

	second := request("student-a")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request = %d, want %d", second.Code, http.StatusTooManyRequests)
	}

	otherUser := request("student-b")
	if otherUser.Code != http.StatusOK {
		t.Fatalf("other user request = %d, want %d", otherUser.Code, http.StatusOK)
	}
}
