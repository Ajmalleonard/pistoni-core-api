package service

import (
	"testing"
	"time"

	"pistoni/core/internal/modules/rentals/dto"
	"pistoni/core/internal/modules/rentals/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func TestCalculateRentalPrice(t *testing.T) {
	st := storage.NewMemoryStore()
	st.Cars["c1"] = model.Car{ID: "c1", Price: 100}
	s := New(repo.NewMemoryRepository(st))
	in := dto.CalculateRentalPriceDTO{StartDate: time.Now().UTC().Format(time.RFC3339), EndDate: time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339), CarID: "c1", DistanceKM: 100}
	out, err := s.CalculateRentalPrice(in)
	if err != nil {
		t.Fatal(err)
	}
	if out["totalPrice"].(float64) <= 0 {
		t.Fatal("expected positive total")
	}
}
