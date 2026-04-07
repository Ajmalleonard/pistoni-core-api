package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(NewRateLimiter().Middleware(2, 2, time.Minute))
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}
