package service

import (
	"testing"

	"pistoni/core/internal/modules/cars/dto"
	carsrepo "pistoni/core/internal/modules/cars/repo"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func TestCreateAndFilterCars(t *testing.T) {
	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	owner, _ := users.Create(model.User{Email: "owner@x.com", Name: "Owner", Role: model.UserRoleCarOwner, AuthProvider: model.AuthProviderEmail, IsVerified: true})
	svc := New(carsrepo.NewMemoryRepository(store, users))

	_, err := svc.Create(dto.CreateCarDTO{Make: "Toyota", Model: "Prado", Year: 2022, Color: "Black", LicensePlate: "T123", Price: 120, SeatCount: 7, TransmissionType: "AUTO", FuelType: "DIESEL", OwnerID: owner.ID})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	_, _ = svc.Create(dto.CreateCarDTO{Make: "Nissan", Model: "X", Year: 2020, Color: "White", LicensePlate: "T124", Price: 100, SeatCount: 5, TransmissionType: "AUTO", FuelType: "PETROL", OwnerID: owner.ID, VehicleType: string(model.VehicleTypeSUV)})

	all, err := svc.FindAll(0, 0, "")
	if err != nil || len(all) != 2 {
		t.Fatalf("expected 2 cars, got %d, err=%v", len(all), err)
	}

	ownerCars, _ := svc.FindByOwner(owner.ID)
	if len(ownerCars) != 2 {
		t.Fatalf("owner cars mismatch")
	}
}
