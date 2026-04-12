package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/account/repo"
	"pistoni/core/internal/modules/account/service"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func TestAccountSettingsEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	user, err := users.Create(model.User{
		Email:        "client@pistoni.io",
		Name:         "Client User",
		Role:         model.UserRoleClient,
		AuthProvider: model.AuthProviderEmail,
		IsVerified:   true,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	h := New(service.New(repo.NewMemoryRepository(store, users), nil))
	r := gin.New()
	h.Register(r.Group("/account"), func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	})

	req := httptest.NewRequest(http.MethodPatch, "/account/settings", bytes.NewBufferString(`{"language":"fr","pushNotifications":false}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/account/overview", nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", getW.Code, getW.Body.String())
	}
}
