package repo

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"pistoni/core/internal/modules/cars/dto"
	"pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type Query struct {
	Skip   int
	Take   int
	Status *model.CarStatus
}

type Repository interface {
	Create(in dto.CreateCarDTO) (model.CarView, error)
	CreateBulk(in []dto.CreateCarDTO) ([]model.CarView, error)
	FindAll(q Query) ([]model.CarView, error)
	FindByID(id string) (model.CarView, error)
	Update(id string, patch map[string]any) (model.CarView, error)
	Delete(id string) (model.Car, error)
	FindByStatus(status model.CarStatus) ([]model.CarView, error)
	FindByOwner(ownerID string) ([]model.CarView, error)
	FindByVehicleType(vt string) ([]model.CarView, error)
}

type MemoryRepository struct {
	store *storage.MemoryStore
	users repo.Repository
}

func NewMemoryRepository(store *storage.MemoryStore, users repo.Repository) *MemoryRepository {
	return &MemoryRepository{store: store, users: users}
}

func (r *MemoryRepository) Create(in dto.CreateCarDTO) (model.CarView, error) {
	r.store.Mu.Lock()
	if _, exists := r.store.CarByLicensePlate[in.LicensePlate]; exists {
		r.store.Mu.Unlock()
		return model.CarView{}, apperr.New(409, "Unique constraint violation on field(s): licensePlate")
	}
	car := model.Car{ID: uuid.NewString(), Make: in.Make, Model: in.Model, Year: in.Year, Color: in.Color, LicensePlate: in.LicensePlate,
		Price: in.Price, Images: in.Images, Features: in.Features, SeatCount: in.SeatCount, TransmissionType: in.TransmissionType,
		FuelType: in.FuelType, OwnerID: in.OwnerID, CreatedAt: time.Now(), UpdatedAt: time.Now(), Status: model.CarStatusPendingReview,
		VehicleType: model.VehicleTypeCoaster,
	}
	if in.Description != "" {
		car.Description = &in.Description
	}
	if in.VehicleType != "" {
		car.VehicleType = model.VehicleType(in.VehicleType)
	}
	r.store.Cars[car.ID] = car
	r.store.CarByLicensePlate[car.LicensePlate] = car.ID
	if in.Location != nil && in.Location.Create != nil {
		loc := in.Location.Create
		state, zip := nullable(loc.State), nullable(loc.ZipCode)
		r.store.LocationsByCarID[car.ID] = model.Location{ID: uuid.NewString(), CarID: car.ID, Latitude: loc.Latitude, Longitude: loc.Longitude,
			Address: loc.Address, City: loc.City, State: state, Country: loc.Country, ZipCode: zip, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	}
	r.store.Mu.Unlock()
	return r.hydrate(car), nil
}

func (r *MemoryRepository) CreateBulk(in []dto.CreateCarDTO) ([]model.CarView, error) {
	out := make([]model.CarView, 0, len(in))
	for _, c := range in {
		item, err := r.Create(c)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *MemoryRepository) FindAll(q Query) ([]model.CarView, error) {
	r.store.Mu.RLock()
	cars := []model.Car{}
	for _, c := range r.store.Cars {
		if q.Status != nil && c.Status != *q.Status {
			continue
		}
		cars = append(cars, c)
	}
	r.store.Mu.RUnlock()

	sort.Slice(cars, func(i, j int) bool { return cars[i].CreatedAt.Before(cars[j].CreatedAt) })
	if q.Skip > len(cars) {
		return []model.CarView{}, nil
	}
	cars = cars[q.Skip:]
	if q.Take > 0 && q.Take < len(cars) {
		cars = cars[:q.Take]
	}
	out := make([]model.CarView, 0, len(cars))
	for _, c := range cars {
		out = append(out, r.hydrate(c))
	}
	return out, nil
}

func (r *MemoryRepository) FindByID(id string) (model.CarView, error) {
	r.store.Mu.RLock()
	c, ok := r.store.Cars[id]
	r.store.Mu.RUnlock()
	if !ok {
		return model.CarView{}, apperr.New(404, "Record not found")
	}
	v := r.hydrate(c)
	v.Reviews = []any{}
	return v, nil
}

func (r *MemoryRepository) Update(id string, patch map[string]any) (model.CarView, error) {
	r.store.Mu.Lock()
	c, ok := r.store.Cars[id]
	if !ok {
		r.store.Mu.Unlock()
		return model.CarView{}, apperr.New(404, "Record not found")
	}
	if v, ok := patch["make"].(string); ok {
		c.Make = v
	}
	if v, ok := patch["model"].(string); ok {
		c.Model = v
	}
	if v, ok := patch["status"].(string); ok {
		c.Status = model.CarStatus(v)
	}
	if v, ok := patch["vehicleType"].(string); ok {
		c.VehicleType = model.VehicleType(v)
	}
	c.UpdatedAt = time.Now()
	r.store.Cars[id] = c
	r.store.Mu.Unlock()
	return r.hydrate(c), nil
}

func (r *MemoryRepository) Delete(id string) (model.Car, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	c, ok := r.store.Cars[id]
	if !ok {
		return model.Car{}, apperr.New(404, "Record not found")
	}
	delete(r.store.Cars, id)
	delete(r.store.CarByLicensePlate, c.LicensePlate)
	delete(r.store.LocationsByCarID, id)
	return c, nil
}

func (r *MemoryRepository) FindByStatus(status model.CarStatus) ([]model.CarView, error) {
	return r.FindAll(Query{Status: &status})
}
func (r *MemoryRepository) FindByOwner(ownerID string) ([]model.CarView, error) {
	all, _ := r.FindAll(Query{})
	out := []model.CarView{}
	for _, c := range all {
		if c.OwnerID == ownerID {
			c.Reviews = []any{}
			out = append(out, c)
		}
	}
	return out, nil
}
func (r *MemoryRepository) FindByVehicleType(vt string) ([]model.CarView, error) {
	all, _ := r.FindAll(Query{})
	out := []model.CarView{}
	for _, c := range all {
		if string(c.VehicleType) == vt && c.Status == model.CarStatusAvailable {
			c.Reviews = []any{}
			out = append(out, c)
		}
	}
	return out, nil
}

func (r *MemoryRepository) hydrate(c model.Car) model.CarView {
	v := model.CarView{Car: c}
	if owner, err := r.users.FindByID(c.OwnerID); err == nil {
		v.Owner = &model.PublicUser{ID: owner.ID, Name: owner.Name, PhoneNumber: owner.PhoneNumber, ProfilePicture: owner.ProfilePicture}
	}
	r.store.Mu.RLock()
	if loc, ok := r.store.LocationsByCarID[c.ID]; ok {
		l := loc
		v.Location = &l
	}
	r.store.Mu.RUnlock()
	return v
}

func nullable(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
