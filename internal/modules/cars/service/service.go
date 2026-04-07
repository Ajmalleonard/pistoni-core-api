package service

import (
	"pistoni/core/internal/modules/cars/dto"
	"pistoni/core/internal/modules/cars/repo"
	"pistoni/core/internal/platform/model"
)

type Service struct{ repo repo.Repository }

func New(r repo.Repository) *Service { return &Service{repo: r} }

func (s *Service) Create(in dto.CreateCarDTO) (model.CarView, error) { return s.repo.Create(in) }
func (s *Service) CreateBulk(in []dto.CreateCarDTO) ([]model.CarView, error) {
	return s.repo.CreateBulk(in)
}
func (s *Service) FindAll(skip, take int, status string) ([]model.CarView, error) {
	q := repo.Query{Skip: skip, Take: take}
	if status != "" {
		st := model.CarStatus(status)
		q.Status = &st
	}
	return s.repo.FindAll(q)
}
func (s *Service) FindOne(id string) (model.CarView, error) { return s.repo.FindByID(id) }
func (s *Service) Update(id string, patch map[string]any) (model.CarView, error) {
	return s.repo.Update(id, patch)
}
func (s *Service) Remove(id string) (model.Car, error) { return s.repo.Delete(id) }
func (s *Service) FindByStatus(status string) ([]model.CarView, error) {
	return s.repo.FindByStatus(model.CarStatus(status))
}
func (s *Service) FindByOwner(ownerID string) ([]model.CarView, error) {
	return s.repo.FindByOwner(ownerID)
}
func (s *Service) FindByVehicleType(vt string) ([]model.CarView, error) {
	return s.repo.FindByVehicleType(vt)
}
func (s *Service) CarTypes() []string {
	return []string{string(model.VehicleTypeSUV), string(model.VehicleTypeCoaster), string(model.VehicleTypeLongVehicle)}
}
