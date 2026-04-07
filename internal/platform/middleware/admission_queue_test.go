package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAdmissionQueueAllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AdmissionQueue(1, 1, time.Second))
	r.POST("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}
