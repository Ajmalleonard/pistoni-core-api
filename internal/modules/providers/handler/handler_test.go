package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	providersrepo "pistoni/core/internal/modules/providers/repo"
	providerssvc "pistoni/core/internal/modules/providers/service"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/storage"
)

func TestListBodyGuardsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	h := New(providerssvc.New(providersrepo.NewMemoryRepository(store, users)))
	r := gin.New()
	h.Register(r.Group("/providers"), func(c *gin.Context) { c.Next() }, func(c *gin.Context) { c.Next() }, func(c *gin.Context) { c.Next() }, func(c *gin.Context) { c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/providers/body-guards", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}
}
