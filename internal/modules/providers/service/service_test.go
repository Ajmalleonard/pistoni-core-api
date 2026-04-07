package service

import (
	"testing"

	"pistoni/core/internal/modules/providers/dto"
	providersrepo "pistoni/core/internal/modules/providers/repo"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func TestRegisterBodyGuardAndFetch(t *testing.T) {
	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	u, _ := users.Create(model.User{Email: "bg@x.com", Name: "BG", Role: model.UserRoleClient, AuthProvider: model.AuthProviderEmail, IsVerified: true})
	svc := New(providersrepo.NewMemoryRepository(store, users))

	_, err := svc.RegisterBodyGuard(u.ID, dto.RegisterBodyGuardDTO{Gender: string(model.GenderMale), NIN: "NIN-1", GovernmentIDNumber: "GOV-1", DailyRate: 55})
	if err != nil {
		t.Fatalf("register body guard: %v", err)
	}

	my, err := svc.GetMyBodyGuardProfile(u.ID)
	if err != nil || my == nil {
		t.Fatalf("profile expected")
	}
}

func TestDriverStatusReviewAndCounters(t *testing.T) {
	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	driverUser, _ := users.Create(model.User{Email: "driver@x.com", Name: "DRV", Role: model.UserRoleClient, AuthProvider: model.AuthProviderEmail, IsVerified: true})
	reviewer, _ := users.Create(model.User{Email: "reviewer@x.com", Name: "REV", Role: model.UserRoleClient, AuthProvider: model.AuthProviderEmail, IsVerified: true})
	svc := New(providersrepo.NewMemoryRepository(store, users))

	createdAny, err := svc.RegisterDriver(driverUser.ID, dto.RegisterDriverDTO{DriversLicenseNumber: "DL-1", DailyRate: 80})
	if err != nil {
		t.Fatalf("register driver: %v", err)
	}
	created := createdAny.(model.DriverProfile)

	_, err = svc.UpdateMyDriverStatus(driverUser.ID, dto.UpdateDriverStatusDTO{Status: string(model.DriverStatusOnline)})
	if err != nil {
		t.Fatalf("update status: %v", err)
	}

	poll, err := svc.PollDriverStatusUpdates(dto.DriverStatusPollQuery{SinceVersion: 0})
	if err != nil {
		t.Fatalf("poll status: %v", err)
	}
	if len(poll.(map[string]any)["updates"].([]map[string]any)) == 0 {
		t.Fatalf("expected at least one status update")
	}

	_, err = svc.CreateDriverReview(created.ID, reviewer.ID, dto.CreateDriverReviewDTO{Rating: 5, Comment: "Great"})
	if err != nil {
		t.Fatalf("create review: %v", err)
	}
	reviews, err := svc.ListDriverReviews(created.ID)
	if err != nil {
		t.Fatalf("list reviews: %v", err)
	}
	agg := reviews.(map[string]any)["aggregate"].(map[string]any)
	if agg["count"].(int) != 1 {
		t.Fatalf("expected review count 1")
	}

	_, err = svc.UpdateDriverAssignmentCounters(created.ID, dto.DriverAssignmentDeltaDTO{RouteDelta: 2, AssignmentDelta: 1})
	if err != nil {
		t.Fatalf("update counters: %v", err)
	}
	counters, _ := svc.GetDriverAssignmentCounters(created.ID)
	cm := counters.(map[string]any)
	if cm["routesCount"].(int) != 2 || cm["assignmentCount"].(int) != 1 {
		t.Fatalf("unexpected counters: %+v", cm)
	}
}

func TestBodyGuardQueueApprove(t *testing.T) {
	store := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(store)
	admin, _ := users.Create(model.User{Email: "admin@x.com", Name: "Admin", Role: model.UserRoleAdmin, AuthProvider: model.AuthProviderEmail, IsVerified: true})
	candidate, _ := users.Create(model.User{Email: "candidate@x.com", Name: "Candidate", Role: model.UserRoleClient, AuthProvider: model.AuthProviderEmail, IsVerified: true})
	svc := New(providersrepo.NewMemoryRepository(store, users))

	joinAny, err := svc.JoinBodyGuardQueue(candidate.ID, dto.JoinBodyGuardQueueDTO{Gender: string(model.GenderMale), NIN: "NQ1", GovernmentIDNumber: "GQ1", DailyRate: 60})
	if err != nil {
		t.Fatalf("join queue: %v", err)
	}
	queueItem := joinAny.(map[string]any)["queueItem"].(model.BodyGuardRegistrationQueueItem)

	_, err = svc.ApproveBodyGuardQueue(queueItem.ID, admin.ID)
	if err != nil {
		t.Fatalf("approve queue: %v", err)
	}

	if _, err = svc.GetMyBodyGuardProfile(candidate.ID); err != nil {
		t.Fatalf("expected bodyguard profile after approval")
	}
}
