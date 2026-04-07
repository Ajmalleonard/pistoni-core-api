package service

import (
	"testing"

	authdto "pistoni/core/internal/modules/auth/dto"
	authrepo "pistoni/core/internal/modules/auth/repo"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/auth"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func TestRegisterLoginRefresh(t *testing.T) {
	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	svc := New(users, authrepo.NewMemoryPendingRepository(store), auth.NewJWTManager("secret"), nil)

	reg, err := svc.Register(authdto.RegisterDTO{Email: "a@b.com", Password: "secret12", Name: "A"})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if reg.User.Email != "a@b.com" {
		t.Fatalf("unexpected user")
	}

	_, _ = users.Update(reg.User.ID, map[string]any{"isVerified": true})
	login, err := svc.Login(authdto.LoginDTO{Email: "a@b.com", Password: "secret12"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if login.AccessToken == "" || login.RefreshToken == "" {
		t.Fatalf("missing tokens")
	}

	ref, err := svc.Refresh(login.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if ref.RefreshToken == login.RefreshToken {
		t.Fatalf("refresh token should rotate")
	}
}

func TestCompleteRegistrationFlow(t *testing.T) {
	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	svc := New(users, authrepo.NewMemoryPendingRepository(store), auth.NewJWTManager("secret"), nil)

	if err := svc.RequestEmailVerification("new@example.com"); err != nil {
		t.Fatalf("request verification: %v", err)
	}
	pending, ok := store.PendingByEmail["new@example.com"]
	if !ok {
		t.Fatalf("pending not stored")
	}
	if err := svc.VerifyEmailCode("new@example.com", pending.Code); err != nil {
		t.Fatalf("verify code: %v", err)
	}

	out, err := svc.CompleteRegistration(authdto.CompleteRegistrationDTO{Email: "new@example.com", Password: "mypassword", Name: "New", Role: string(model.UserRoleCarOwner)})
	if err != nil {
		t.Fatalf("complete registration: %v", err)
	}
	if out.User.Role != model.UserRoleCarOwner {
		t.Fatalf("expected role car owner")
	}
}
