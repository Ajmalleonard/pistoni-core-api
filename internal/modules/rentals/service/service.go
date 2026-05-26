package service

import (
	"time"

	"pistoni/core/internal/modules/rentals/dto"
	"pistoni/core/internal/modules/rentals/repo"
	"pistoni/core/internal/platform/model"
)

type Service struct{ repo repo.Repository }

func New(r repo.Repository) *Service                                   { return &Service{repo: r} }
func (s *Service) Create(in dto.CreateRentalDTO) (model.Rental, error) { return s.repo.Create(in) }
func (s *Service) CreateRouteRentals(in dto.CreateRouteRentalsDTO) (repo.RouteRentalsResult, error) {
	return s.repo.CreateRouteRentals(in)
}
func (s *Service) FindAll(skip, take int, userID, status string) ([]model.Rental, error) {
	return s.repo.FindAll(repo.Query{Skip: skip, Take: take, UserID: userID, Status: status})
}
func (s *Service) FindByUser(userID string) ([]model.Rental, error) { return s.repo.FindByUser(userID) }
func (s *Service) FindOne(id string) (model.Rental, error)          { return s.repo.FindOne(id) }
func (s *Service) Update(id string, patch map[string]any) (model.Rental, error) {
	return s.repo.Update(id, patch)
}
func (s *Service) Remove(id string) (model.Rental, error) { return s.repo.Remove(id) }
func (s *Service) CalculateRentalPrice(in dto.CalculateRentalPriceDTO) (map[string]any, error) {
	sd, err := time.Parse(time.RFC3339, in.StartDate)
	if err != nil {
		return nil, err
	}
	ed, err := time.Parse(time.RFC3339, in.EndDate)
	if err != nil {
		return nil, err
	}
	return s.repo.CalculateRentalPrice(sd, ed, in.CarID, in.DistanceKM, in.DriverID, in.BodyGuardIDs)
}

func (s *Service) CalculateRoutePrice(in dto.CalculateRoutePriceDTO) (map[string]any, error) {
	sd, err := time.Parse(time.RFC3339, in.StartDate)
	if err != nil {
		return nil, err
	}
	ed, err := time.Parse(time.RFC3339, in.EndDate)
	if err != nil {
		return nil, err
	}
	return s.repo.CalculateRoutePrice(sd, ed, in.Cars, in.DistanceKM, in.BodyGuardIDs)
}
