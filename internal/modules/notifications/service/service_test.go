package service

import (
	"testing"

	"pistoni/core/internal/modules/notifications/dto"
	"pistoni/core/internal/modules/notifications/repo"
	"pistoni/core/internal/platform/storage"
)

func TestMarkAsReadNotFound(t *testing.T) {
	s := New(repo.NewMemoryRepository(storage.NewMemoryStore()))
	if _, err := s.MarkAsRead([]string{"missing"}, "u1"); err == nil {
		t.Fatal("expected error")
	}
	_, _ = s.CreateNotification(dto.CreateNotificationDTO{UserID: "u1", Title: "t", Body: "b", Type: "SYSTEM"})
}
