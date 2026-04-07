package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	authdto "pistoni/core/internal/modules/auth/dto"
	authrepo "pistoni/core/internal/modules/auth/repo"
	authsvc "pistoni/core/internal/modules/auth/service"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/auth"
	"pistoni/core/internal/platform/storage"
)

func TestLoginEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	svc := authsvc.New(users, authrepo.NewMemoryPendingRepository(store), auth.NewJWTManager("secret"), nil)
	_, _ = svc.Register(authdto.RegisterDTO{Email: "user@a.com", Password: "secret12", Name: "User"})
	u, _ := users.FindByEmail("user@a.com")
	_, _ = users.Update(u.ID, map[string]any{"isVerified": true})

	r := gin.New()
	New(svc).Register(r.Group("/auth"), func(c *gin.Context) { c.Next() })

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"user@a.com","password":"secret12"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}
}
