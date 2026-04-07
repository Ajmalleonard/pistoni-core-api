package service

import (
	"testing"

	"pistoni/core/internal/modules/paymentmethods/dto"
	"pistoni/core/internal/modules/paymentmethods/repo"
	"pistoni/core/internal/platform/storage"
)

func TestCreateCardValidation(t *testing.T) {
	s := New(repo.NewMemoryRepository(storage.NewMemoryStore()))
	if _, err := s.CreateCardPaymentMethod("u1", dto.CreateCardPaymentMethodDTO{CardNumber: "123", CardType: "Visa", ExpiryMonth: 1, ExpiryYear: 2030, CardholderName: "A B C"}); err == nil {
		t.Fatal("expected invalid card number")
	}
}
