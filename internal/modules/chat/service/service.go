package service

import (
	"pistoni/core/internal/modules/chat/dto"
	"pistoni/core/internal/modules/chat/repo"
)

type Service struct{ repo repo.Repository }

func New(r repo.Repository) *Service { return &Service{repo: r} }
func (s *Service) GetUserChatRooms(userID string) (any, error) {
	return s.repo.GetUserChatRooms(userID)
}
func (s *Service) GetChatRoom(userID, id string) (any, error) { return s.repo.GetChatRoom(userID, id) }
func (s *Service) GetMessages(chatRoomID string) (any, error) { return s.repo.GetMessages(chatRoomID) }
func (s *Service) CreateMessage(userID string, in dto.CreateMessageDTO) (any, error) {
	return s.repo.CreateMessage(userID, in)
}
func (s *Service) EditMessage(userID string, in dto.EditMessageDTO) (any, error) {
	return s.repo.EditMessage(userID, in)
}
func (s *Service) GetCommandSuggestions(tripID, q string) (any, error) {
	return s.repo.GetCommandSuggestions(tripID, q)
}
func (s *Service) GetCommandTargets(userID, tripID, targetType string) (any, error) {
	return s.repo.GetCommandTargets(userID, tripID, targetType)
}
func (s *Service) CreateTargetedMessage(userID string, in dto.CreateTargetedMessageDTO) (any, error) {
	return s.repo.CreateTargetedMessage(userID, in)
}
func (s *Service) CreateTargetedReport(userID string, in dto.CreateTargetedReportDTO) (any, error) {
	return s.repo.CreateTargetedReport(userID, in)
}
func (s *Service) DeleteMessage(userID string, in dto.DeleteMessageDTO) (any, error) {
	return s.repo.DeleteMessage(userID, in)
}
func (s *Service) CreatePrivateChat(userID, participantID string) (any, error) {
	return s.repo.CreatePrivateChat(userID, participantID)
}
func (s *Service) CreateGroupChat(creatorID, name string, participantIDs []string) (any, error) {
	return s.repo.CreateGroupChat(creatorID, name, participantIDs)
}
func (s *Service) GetUserByEmail(email string) (any, error) { return s.repo.GetUserByEmail(email) }
