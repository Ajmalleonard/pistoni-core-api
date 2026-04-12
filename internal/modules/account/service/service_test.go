package service

import (
	"testing"

	"pistoni/core/internal/modules/account/dto"
	"pistoni/core/internal/modules/account/repo"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func TestAccountOverviewAndSettings(t *testing.T) {
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

	svc := New(repo.NewMemoryRepository(store, users), nil)
	settings, err := svc.UpdateSettings(user.ID, dto.UpdateSettingsDTO{
		Language:         ptr("sw"),
		Timezone:         ptr("Africa/Dar_es_Salaam"),
		SMSNotifications: boolPtr(true),
		TextSize:         ptr("large"),
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if settings.Language != "sw" || settings.Timezone != "Africa/Dar_es_Salaam" || !settings.SMSNotifications {
		t.Fatalf("unexpected settings: %+v", settings)
	}

	if _, err := svc.CreateTrustedContact(user.ID, dto.CreateTrustedContactDTO{
		Name:         "Asha",
		PhoneNumber:  "+255700000001",
		Relationship: "Sister",
	}); err != nil {
		t.Fatalf("create trusted contact: %v", err)
	}

	overview, err := svc.GetOverview(user.ID)
	if err != nil {
		t.Fatalf("get overview: %v", err)
	}
	if overview.Settings.Language != "sw" {
		t.Fatalf("expected language sw, got %s", overview.Settings.Language)
	}
	if len(overview.TrustedContacts) != 1 {
		t.Fatalf("expected 1 trusted contact, got %d", len(overview.TrustedContacts))
	}
}

func ptr(v string) *string { return &v }
func boolPtr(v bool) *bool { return &v }
