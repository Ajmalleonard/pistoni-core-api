package repo

import (
	"testing"
	"time"

	"pistoni/core/internal/modules/rentals/dto"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func TestCreateRouteRentalsCreatesOneRentalPerSelectedCar(t *testing.T) {
	st := storage.NewMemoryStore()
	st.Users["u1"] = model.User{ID: "u1"}
	st.Users["system"] = model.User{ID: "system", Role: model.UserRoleAdmin}
	st.Users["driver-user-1"] = model.User{ID: "driver-user-1", Role: model.UserRoleDriver}
	st.Users["guard-user-1"] = model.User{ID: "guard-user-1", Role: model.UserRoleBodyGuard}
	st.Cars["c1"] = model.Car{ID: "c1", Price: 100}
	st.Cars["c2"] = model.Car{ID: "c2", Price: 200}
	st.DriverProfiles["d1"] = model.DriverProfile{ID: "d1", UserID: "driver-user-1", IsAvailable: true}
	st.BodyGuardProfiles["bg1"] = model.BodyGuardProfile{ID: "bg1", UserID: "guard-user-1", IsAvailable: true}

	driverID := "d1"
	distance := 18.5
	repository := NewMemoryRepository(st)
	result, err := repository.CreateRouteRentals(dto.CreateRouteRentalsDTO{
		StartDate:         time.Now().UTC().Format(time.RFC3339),
		EndDate:           time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		TotalPrice:        600,
		UserID:            "u1",
		PickupLocationID:  "pickup-place",
		DropoffLocationID: "dropoff-place",
		DistanceKM:        &distance,
		BodyGuardIDs:      []string{"bg1"},
		Cars: []dto.RouteCarSelectionDTO{
			{CarID: "c1", DriverID: &driverID},
			{CarID: "c2"},
		},
	})
	if err != nil {
		t.Fatalf("CreateRouteRentals() error = %v", err)
	}
	rentals := result.Rentals
	if len(rentals) != 2 {
		t.Fatalf("expected 2 rentals got %d", len(rentals))
	}
	if rentals[0].TotalPrice != 300 || rentals[1].TotalPrice != 300 {
		t.Fatalf("expected price split across rentals, got %.2f and %.2f", rentals[0].TotalPrice, rentals[1].TotalPrice)
	}
	if rentals[0].DriverID == nil || *rentals[0].DriverID != "d1" {
		t.Fatalf("expected first rental driver d1 got %+v", rentals[0].DriverID)
	}
	for _, rental := range rentals {
		if !rental.SecurityRequired {
			t.Fatal("expected security to be required when bodyguards are selected")
		}
		if len(rental.BodyGuardIDs) != 1 || rental.BodyGuardIDs[0] != "bg1" {
			t.Fatalf("expected shared bodyguard bg1 got %+v", rental.BodyGuardIDs)
		}
	}
	if len(result.ChatRooms) != 5 {
		t.Fatalf("expected client/driver/bodyguard rooms across two rentals got %d", len(result.ChatRooms))
	}
	var hasClient, hasDriverOps, hasBGTeam bool
	for _, room := range result.ChatRooms {
		switch room.Type {
		case "client":
			hasClient = true
		case "driver_ops":
			hasDriverOps = true
		case "bg_team":
			hasBGTeam = true
		}
	}
	if !hasClient || !hasDriverOps || !hasBGTeam {
		t.Fatalf("expected VIP control plane room types, got %+v", result.ChatRooms)
	}
}

func TestCreateRouteRentalsRejectsMoreThanTenBodyGuards(t *testing.T) {
	st := storage.NewMemoryStore()
	st.Users["u1"] = model.User{ID: "u1"}
	st.Cars["c1"] = model.Car{ID: "c1", Price: 100}

	guards := make([]string, 11)
	for i := range guards {
		guards[i] = "bg" + string(rune('a'+i))
	}

	_, err := NewMemoryRepository(st).CreateRouteRentals(dto.CreateRouteRentalsDTO{
		StartDate:         time.Now().UTC().Format(time.RFC3339),
		EndDate:           time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		TotalPrice:        100,
		UserID:            "u1",
		PickupLocationID:  "pickup-place",
		DropoffLocationID: "dropoff-place",
		BodyGuardIDs:      guards,
		Cars:              []dto.RouteCarSelectionDTO{{CarID: "c1"}},
	})
	if err == nil {
		t.Fatal("expected max bodyguard validation error")
	}
}
