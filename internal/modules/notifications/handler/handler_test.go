package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/notifications/repo"
	"pistoni/core/internal/modules/notifications/service"
	"pistoni/core/internal/platform/storage"
)

func TestNotificationsUnreadCountEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(service.New(repo.NewMemoryRepository(storage.NewMemoryStore())))
	r := gin.New()
	h.Register(r.Group("/notifications"))
	req := httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}
