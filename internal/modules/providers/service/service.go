package service

import (
	"pistoni/core/internal/modules/providers/dto"
	"pistoni/core/internal/modules/providers/repo"
)

type Service struct{ repo repo.Repository }

func New(r repo.Repository) *Service { return &Service{repo: r} }

func (s *Service) RegisterDriver(userID string, in dto.RegisterDriverDTO) (any, error) {
	return s.repo.RegisterDriver(userID, in)
}
func (s *Service) FindDrivers(q dto.FindProviderQuery) (any, error) { return s.repo.FindDrivers(q) }
func (s *Service) GetMyDriverProfile(userID string) (any, error) {
	return s.repo.GetMyDriverProfile(userID)
}
func (s *Service) UpdateMyDriverAvailability(userID string, isAvailable bool) (any, error) {
	return s.repo.UpdateMyDriverAvailability(userID, isAvailable)
}
func (s *Service) UpdateDriverVerification(id string, in dto.UpdateVerificationDTO) (any, error) {
	return s.repo.UpdateDriverVerification(id, in)
}
func (s *Service) FindDriverByID(id string) (any, error)  { return s.repo.FindDriverByID(id) }
func (s *Service) GetDriverStatus(id string) (any, error) { return s.repo.GetDriverStatus(id) }
func (s *Service) GetMyDriverStatus(userID string) (any, error) {
	return s.repo.GetMyDriverStatus(userID)
}
func (s *Service) UpdateMyDriverStatus(userID string, in dto.UpdateDriverStatusDTO) (any, error) {
	return s.repo.UpdateMyDriverStatus(userID, in)
}
func (s *Service) PollDriverStatusUpdates(q dto.DriverStatusPollQuery) (any, error) {
	return s.repo.PollDriverStatusUpdates(q)
}
func (s *Service) ListDriverReviews(driverID string) (any, error) {
	return s.repo.ListDriverReviews(driverID)
}
func (s *Service) CreateDriverReview(driverID, authorID string, in dto.CreateDriverReviewDTO) (any, error) {
	return s.repo.CreateDriverReview(driverID, authorID, in)
}
func (s *Service) GetDriverAssignmentCounters(driverID string) (any, error) {
	return s.repo.GetDriverAssignmentCounters(driverID)
}
func (s *Service) UpdateDriverAssignmentCounters(driverID string, in dto.DriverAssignmentDeltaDTO) (any, error) {
	return s.repo.UpdateDriverAssignmentCounters(driverID, in)
}

func (s *Service) RegisterBodyGuard(userID string, in dto.RegisterBodyGuardDTO) (any, error) {
	return s.repo.RegisterBodyGuard(userID, in)
}
func (s *Service) FindBodyGuards(q dto.FindProviderQuery) (any, error) {
	return s.repo.FindBodyGuards(q)
}
func (s *Service) GetMyBodyGuardProfile(userID string) (any, error) {
	return s.repo.GetMyBodyGuardProfile(userID)
}
func (s *Service) UpdateMyBodyGuardAvailability(userID string, isAvailable bool) (any, error) {
	return s.repo.UpdateMyBodyGuardAvailability(userID, isAvailable)
}
func (s *Service) UpdateBodyGuardVerification(id string, in dto.UpdateVerificationDTO) (any, error) {
	return s.repo.UpdateBodyGuardVerification(id, in)
}
func (s *Service) FindBodyGuardByID(id string) (any, error)  { return s.repo.FindBodyGuardByID(id) }
func (s *Service) GetBodyGuardStatus(id string) (any, error) { return s.repo.GetBodyGuardStatus(id) }
func (s *Service) GetMyBodyGuardStatus(userID string) (any, error) {
	return s.repo.GetMyBodyGuardStatus(userID)
}
func (s *Service) UpdateMyBodyGuardStatus(userID string, in dto.UpdateBodyGuardStatusDTO) (any, error) {
	return s.repo.UpdateMyBodyGuardStatus(userID, in)
}
func (s *Service) JoinBodyGuardQueue(userID string, in dto.JoinBodyGuardQueueDTO) (any, error) {
	return s.repo.JoinBodyGuardQueue(userID, in)
}
func (s *Service) ListBodyGuardQueue(status string) (any, error) {
	return s.repo.ListBodyGuardQueue(status)
}
func (s *Service) ApproveBodyGuardQueue(id, reviewerID string) (any, error) {
	return s.repo.ApproveBodyGuardQueue(id, reviewerID)
}
func (s *Service) RejectBodyGuardQueue(id, reviewerID, reason string) (any, error) {
	return s.repo.RejectBodyGuardQueue(id, reviewerID, reason)
}
