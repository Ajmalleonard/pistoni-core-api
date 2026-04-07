package guard

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRetryBackoffEscalates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRetryBackoffLimiter()
	r := gin.New()
	r.Use(limiter.Middleware(60*time.Second, 10*time.Minute))
	r.POST("/providers/body-guards/join", func(c *gin.Context) { c.Status(http.StatusOK) })

	req1, _ := http.NewRequest(http.MethodPost, "/providers/body-guards/join", nil)
	req1.RemoteAddr = "10.1.1.1:4321"
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", w1.Code)
	}

	req2, _ := http.NewRequest(http.MethodPost, "/providers/body-guards/join", nil)
	req2.RemoteAddr = "10.1.1.1:4321"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d", w2.Code)
	}

	req3, _ := http.NewRequest(http.MethodPost, "/providers/body-guards/join", nil)
	req3.RemoteAddr = "10.1.1.1:4321"
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected third request 429, got %d", w3.Code)
	}
	if w3.Body.String() == w2.Body.String() {
		t.Fatalf("expected escalating penalty response body to differ")
	}
}
