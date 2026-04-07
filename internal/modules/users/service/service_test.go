package service

import (
	"testing"

	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func TestCreateFindByRole(t *testing.T) {
	store := storage.NewMemoryStore()
	svc := New(usersrepo.NewMemoryRepository(store))
	_, _ = svc.Create(model.User{Email: "owner@x.com", Name: "Owner", Role: model.UserRoleCarOwner, AuthProvider: model.AuthProviderEmail, IsVerified: true})
	owners, err := svc.FindByRole(model.UserRoleCarOwner)
	if err != nil || len(owners) != 1 {
		t.Fatalf("expected one owner")
	}
}
