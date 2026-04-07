package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/rentals/repo"
	"pistoni/core/internal/modules/rentals/service"
	"pistoni/core/internal/platform/storage"
)

func TestRentalsListEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(service.New(repo.NewMemoryRepository(storage.NewMemoryStore())))
	r := gin.New()
	h.Register(r.Group("/rentals"))
	req := httptest.NewRequest(http.MethodGet, "/rentals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}
