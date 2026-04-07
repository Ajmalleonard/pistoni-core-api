package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	usersrepo "pistoni/core/internal/modules/users/repo"
	userssvc "pistoni/core/internal/modules/users/service"
	"pistoni/core/internal/platform/storage"
)

func TestCreateUserEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(userssvc.New(usersrepo.NewMemoryRepository(storage.NewMemoryStore())))
	r := gin.New()
	h.Register(r.Group("/users"))

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"email":"u@x.com","name":"U","role":"CLIENT","authProvider":"EMAIL"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}
