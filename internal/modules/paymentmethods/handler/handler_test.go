package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"pistoni/core/internal/modules/paymentmethods/repo"
	"pistoni/core/internal/modules/paymentmethods/service"
	"pistoni/core/internal/platform/storage"
)

func TestPaymentMethodsListEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(service.New(repo.NewMemoryRepository(storage.NewMemoryStore())))
	r := gin.New()
	h.Register(r.Group("/payment-methods"))
	req := httptest.NewRequest(http.MethodGet, "/payment-methods", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestPaymentMethodsCreateCardEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(service.New(repo.NewMemoryRepository(storage.NewMemoryStore())))
	r := gin.New()
	h.Register(r.Group("/payment-methods"))
	body := []byte(`{"cardNumber":"4242424242424242","cardType":"VISA","expiryMonth":12,"expiryYear":2030,"cardholderName":"John Mark Doe"}`)
	req := httptest.NewRequest(http.MethodPost, "/payment-methods/card", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}
}
