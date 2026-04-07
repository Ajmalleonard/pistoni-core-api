package service

import (
	"pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
)

type Service struct{ repo repo.Repository }

func New(r repo.Repository) *Service { return &Service{repo: r} }

func (s *Service) Create(data model.User) (model.User, error) { return s.repo.Create(data) }
func (s *Service) FindAll(skip, take int, role *model.UserRole) ([]model.User, error) {
	return s.repo.FindAll(skip, take, role)
}
func (s *Service) FindByRole(role model.UserRole) ([]model.User, error) {
	return s.repo.FindByRole(role)
}
func (s *Service) FindByEmail(email string) (model.User, error) { return s.repo.FindByEmail(email) }
func (s *Service) FindOne(id string) (model.User, error)        { return s.repo.FindByID(id) }
func (s *Service) Update(id string, patch map[string]any) (model.User, error) {
	return s.repo.Update(id, patch)
}
func (s *Service) Remove(id string) (model.User, error) { return s.repo.Delete(id) }
