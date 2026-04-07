package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/chat/repo"
	"pistoni/core/internal/modules/chat/service"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func buildHandler() *Handler {
	st := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(st)
	_, _ = users.Create(model.User{ID: "test-user-123", Email: "client@test.com", Name: "Client", Role: model.UserRoleClient, AuthProvider: model.AuthProviderEmail, IsVerified: true})
	driver := "driver1"
	st.Rentals["trip1"] = model.Rental{ID: "trip1", UserID: "test-user-123", DriverID: &driver, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return New(service.New(repo.NewMemoryRepository(st, users)))
}

func TestChatRoomsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := buildHandler()
	r := gin.New()
	h.Register(r.Group("/chat"))
	req := httptest.NewRequest(http.MethodGet, "/chat/rooms", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestSlashSuggestionsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := buildHandler()
	r := gin.New()
	h.Register(r.Group("/chat"))
	req := httptest.NewRequest(http.MethodGet, "/chat/commands/suggestions?tripId=trip1&q=/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}
