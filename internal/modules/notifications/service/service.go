package service

import (
	"pistoni/core/internal/modules/notifications/dto"
	"pistoni/core/internal/modules/notifications/repo"
)

type Service struct{ repo repo.Repository }

func New(r repo.Repository) *Service { return &Service{repo: r} }
func (s *Service) CreateNotification(in dto.CreateNotificationDTO) (any, error) {
	return s.repo.CreateNotification(in)
}
func (s *Service) GetUserNotifications(userID string, page, limit int) (any, error) {
	return s.repo.GetUserNotifications(userID, page, limit)
}
func (s *Service) GetUnreadCount(userID string) (any, error) { return s.repo.GetUnreadCount(userID) }
func (s *Service) MarkAsRead(ids []string, userID string) (any, error) {
	return s.repo.MarkAsRead(ids, userID)
}
func (s *Service) MarkAllAsRead(userID string) (any, error) { return s.repo.MarkAllAsRead(userID) }
func (s *Service) DeleteNotification(id, userID string) (any, error) {
	return s.repo.DeleteNotification(id, userID)
}
func (s *Service) CreateSystemNotification(userID, title, body string) (any, error) {
	return s.repo.CreateNotification(dto.CreateNotificationDTO{UserID: userID, Title: title, Body: body, Type: "SYSTEM"})
}
