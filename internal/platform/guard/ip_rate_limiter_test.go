package guard

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestIPRateLimiterRouteLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewIPRateLimiter()

	r := gin.New()
	r.Use(limiter.Middleware(100, 2, time.Minute))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 first request, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req2.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 second request, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req3.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 third request, got %d", w3.Code)
	}
}

func TestIPRateLimiterDifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewIPRateLimiter()

	r := gin.New()
	r.Use(limiter.Middleware(100, 1, time.Minute))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 for ip1, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/x", nil)
	req2.RemoteAddr = "10.0.0.2:1234"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for ip2, got %d", w2.Code)
	}
}
