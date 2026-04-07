package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	carsrepo "pistoni/core/internal/modules/cars/repo"
	carssvc "pistoni/core/internal/modules/cars/service"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func TestCreateCarEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	owner, _ := users.Create(model.User{Email: "o@o.com", Name: "Owner", Role: model.UserRoleCarOwner, AuthProvider: model.AuthProviderEmail, IsVerified: true})
	h := New(carssvc.New(carsrepo.NewMemoryRepository(store, users)))
	r := gin.New()
	h.Register(r.Group("/cars"))

	payload := `{"make":"Toyota","model":"Land Cruiser","year":2022,"color":"Black","licensePlate":"ABC123","price":100,"seatCount":7,"transmissionType":"AUTO","fuelType":"DIESEL","ownerId":"` + owner.ID + `"}`
	req := httptest.NewRequest(http.MethodPost, "/cars", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}
}
