package service

import (
	"testing"
	"time"

	"pistoni/core/internal/modules/chat/dto"
	"pistoni/core/internal/modules/chat/repo"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

func seedChatTargeting(t *testing.T) (*Service, string, string, string, string, string, string, string) {
	t.Helper()
	st := storage.NewMemoryStore()
	users := usersrepo.NewMemoryRepository(st)
	mustCreateUser := func(id string, role model.UserRole) {
		_, err := users.Create(model.User{ID: id, Email: id + "@test.com", Name: id, Role: role, AuthProvider: model.AuthProviderEmail, IsVerified: true})
		if err != nil {
			t.Fatal(err)
		}
	}
	mustCreateUser("client1", model.UserRoleClient)
	mustCreateUser("driver1", model.UserRoleDriver)
	mustCreateUser("bg1", model.UserRoleBodyGuard)
	mustCreateUser("bg2", model.UserRoleBodyGuard)
	mustCreateUser("other", model.UserRoleClient)

	driverID := "driver1"
	st.Rentals["trip1"] = model.Rental{ID: "trip1", UserID: "client1", DriverID: &driverID, BodyGuardIDs: []string{"bg1", "bg2"}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	room, err := repo.NewMemoryRepository(st, users).CreateGroupChat("client1", "ops", []string{"driver1", "bg1", "bg2"})
	if err != nil {
		t.Fatal(err)
	}
	s := New(repo.NewMemoryRepository(st, users))
	return s, room.ID, "trip1", "client1", "driver1", "bg1", "bg2", "other"
}

func TestSuggestionsEndpointData(t *testing.T) {
	s, _, tripID, _, _, _, _, _ := seedChatTargeting(t)
	res, err := s.GetCommandSuggestions(tripID, "/")
	if err != nil {
		t.Fatal(err)
	}
	out := res.([]dto.CommandSuggestion)
	if len(out) < 4 {
		t.Fatalf("expected slash options, got %d", len(out))
	}
}

func TestTargetedMessageToDriverWorks(t *testing.T) {
	s, roomID, tripID, _, driverID, _, _, _ := seedChatTargeting(t)
	res, err := s.CreateTargetedMessage(driverID, dto.CreateTargetedMessageDTO{TripID: tripID, ChatRoomID: roomID, Content: "ETA +5", TargetType: "driver"})
	if err != nil {
		t.Fatal(err)
	}
	msg := res.(model.Message)
	if msg.TargetType == nil || *msg.TargetType != "driver" {
		t.Fatalf("expected driver target, got %+v", msg.TargetType)
	}
}

func TestTargetedBodyguardByNumberWorks(t *testing.T) {
	s, roomID, tripID, _, driverID, bg1, _, _ := seedChatTargeting(t)
	idx := 1
	res, err := s.CreateTargetedMessage(driverID, dto.CreateTargetedMessageDTO{TripID: tripID, ChatRoomID: roomID, Content: "Check perimeter", TargetType: "bodyguard", BodyguardIndex: &idx})
	if err != nil {
		t.Fatal(err)
	}
	msg := res.(model.Message)
	if len(msg.TargetIDs) != 1 || msg.TargetIDs[0] != bg1 {
		t.Fatalf("expected target %s got %+v", bg1, msg.TargetIDs)
	}
}

func TestTargetedGroupWorks(t *testing.T) {
	s, roomID, tripID, clientID, _, bg1, bg2, _ := seedChatTargeting(t)
	res, err := s.CreateTargetedReport(clientID, dto.CreateTargetedReportDTO{TripID: tripID, ChatRoomID: roomID, Content: "Security check", TargetType: "bodyguard-group", ReportCategory: "security"})
	if err != nil {
		t.Fatal(err)
	}
	msg := res.(model.Message)
	if len(msg.TargetIDs) != 2 {
		t.Fatalf("expected 2 bodyguards got %+v", msg.TargetIDs)
	}
	if !(contains(msg.TargetIDs, bg1) && contains(msg.TargetIDs, bg2)) {
		t.Fatalf("missing expected bodyguards: %+v", msg.TargetIDs)
	}
}

func TestUnauthorizedTripAccessRejected(t *testing.T) {
	s, roomID, tripID, _, _, _, _, other := seedChatTargeting(t)
	_, err := s.CreateTargetedReport(other, dto.CreateTargetedReportDTO{TripID: tripID, ChatRoomID: roomID, Content: "x", TargetType: "driver", ReportCategory: "incident"})
	if err == nil {
		t.Fatal("expected forbidden")
	}
}

func contains(v []string, id string) bool {
	for _, x := range v {
		if x == id {
			return true
		}
	}
	return false
}
