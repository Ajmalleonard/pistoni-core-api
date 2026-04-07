package repo

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"pistoni/core/internal/modules/rentals/dto"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type Query struct {
	Skip, Take     int
	UserID, Status string
}

type Repository interface {
	Create(in dto.CreateRentalDTO) (model.Rental, error)
	FindAll(q Query) ([]model.Rental, error)
	FindOne(id string) (model.Rental, error)
	Update(id string, patch map[string]any) (model.Rental, error)
	Remove(id string) (model.Rental, error)
	FindByUser(userID string) ([]model.Rental, error)
	CalculateRentalPrice(startDate, endDate time.Time, carID string, distanceKM float64, driverID *string, bodyGuardIDs []string) (map[string]any, error)
}

type MemoryRepository struct{ store *storage.MemoryStore }

func NewMemoryRepository(store *storage.MemoryStore) *MemoryRepository {
	return &MemoryRepository{store: store}
}

func (r *MemoryRepository) Create(in dto.CreateRentalDTO) (model.Rental, error) {
	sd, err := time.Parse(time.RFC3339, in.StartDate)
	if err != nil {
		return model.Rental{}, apperr.New(400, "Invalid startDate")
	}
	ed, err := time.Parse(time.RFC3339, in.EndDate)
	if err != nil {
		return model.Rental{}, apperr.New(400, "Invalid endDate")
	}
	status := model.RentalStatusPending
	if in.Status != "" {
		status = model.RentalStatus(in.Status)
	}
	driverRequired := in.DriverRequired != nil && *in.DriverRequired
	securityRequired := in.SecurityRequired != nil && *in.SecurityRequired
	if in.DriverID != nil {
		driverRequired = true
	}
	if len(in.BodyGuardIDs) > 0 {
		securityRequired = true
	}
	if driverRequired && in.DriverID == nil {
		return model.Rental{}, apperr.New(400, "A driver must be selected when driver service is required")
	}
	if securityRequired && len(in.BodyGuardIDs) == 0 {
		return model.Rental{}, apperr.New(400, "At least one body guard must be selected when security is required")
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	if _, ok := r.store.Users[in.UserID]; !ok {
		return model.Rental{}, apperr.New(404, "User not found")
	}
	car, ok := r.store.Cars[in.CarID]
	if !ok {
		return model.Rental{}, apperr.New(404, "Car not found")
	}
	if _, ok := r.store.LocationsByCarID[in.PickupLocationID]; !ok && in.PickupLocationID != car.ID { /* allow carId style */
	}
	if in.DriverID != nil {
		p, ok := r.store.DriverProfiles[*in.DriverID]
		if !ok || !p.IsAvailable {
			return model.Rental{}, apperr.New(400, "Driver not available")
		}
	}
	for _, id := range in.BodyGuardIDs {
		p, ok := r.store.BodyGuardProfiles[id]
		if !ok || !p.IsAvailable {
			return model.Rental{}, apperr.New(400, "One or more body guards not available")
		}
	}
	now := time.Now()
	rent := model.Rental{ID: uuid.NewString(), StartDate: sd, EndDate: ed, TotalPrice: in.TotalPrice, Status: status, DistanceKM: in.DistanceKM, DriverRequired: driverRequired, SecurityRequired: securityRequired, SecurityNotes: in.SecurityNotes, UserID: in.UserID, CarID: in.CarID, PickupLocationID: in.PickupLocationID, DropoffLocationID: in.DropoffLocationID, DriverID: in.DriverID, BodyGuardIDs: append([]string{}, in.BodyGuardIDs...), CreatedAt: now, UpdatedAt: now}
	r.store.Rentals[rent.ID] = rent
	return rent, nil
}
func (r *MemoryRepository) FindAll(q Query) ([]model.Rental, error) {
	r.store.Mu.RLock()
	out := []model.Rental{}
	for _, x := range r.store.Rentals {
		if q.UserID != "" && x.UserID != q.UserID {
			continue
		}
		if q.Status != "" && string(x.Status) != q.Status {
			continue
		}
		out = append(out, x)
	}
	r.store.Mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if q.Skip > len(out) {
		return []model.Rental{}, nil
	}
	out = out[q.Skip:]
	if q.Take > 0 && q.Take < len(out) {
		out = out[:q.Take]
	}
	return out, nil
}
func (r *MemoryRepository) FindOne(id string) (model.Rental, error) {
	r.store.Mu.RLock()
	x, ok := r.store.Rentals[id]
	r.store.Mu.RUnlock()
	if !ok {
		return model.Rental{}, apperr.New(404, "Record not found")
	}
	return x, nil
}
func (r *MemoryRepository) Update(id string, patch map[string]any) (model.Rental, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	x, ok := r.store.Rentals[id]
	if !ok {
		return model.Rental{}, apperr.New(404, "Record not found")
	}
	if v, ok := patch["status"].(string); ok {
		x.Status = model.RentalStatus(v)
	}
	x.UpdatedAt = time.Now()
	r.store.Rentals[id] = x
	return x, nil
}
func (r *MemoryRepository) Remove(id string) (model.Rental, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	x, ok := r.store.Rentals[id]
	if !ok {
		return model.Rental{}, apperr.New(404, "Record not found")
	}
	delete(r.store.Rentals, id)
	return x, nil
}
func (r *MemoryRepository) FindByUser(userID string) ([]model.Rental, error) {
	return r.FindAll(Query{UserID: userID})
}
func (r *MemoryRepository) CalculateRentalPrice(startDate, endDate time.Time, carID string, distanceKM float64, driverID *string, bodyGuardIDs []string) (map[string]any, error) {
	r.store.Mu.RLock()
	car, ok := r.store.Cars[carID]
	if !ok {
		r.store.Mu.RUnlock()
		return nil, apperr.New(404, "Car not found")
	}
	days := int(endDate.Sub(startDate).Hours() / 24)
	if endDate.Sub(startDate) > time.Duration(days)*24*time.Hour {
		days++
	}
	if days < 1 {
		days = 1
	}
	base := float64(days) * car.Price
	distance := base * 0.1 * (distanceKM / 100)
	driver := 0.0
	if driverID != nil {
		p, ok := r.store.DriverProfiles[*driverID]
		if !ok {
			r.store.Mu.RUnlock()
			return nil, apperr.New(404, "Driver not found")
		}
		driver = float64(days) * p.DailyRate
	}
	body := 0.0
	for _, id := range bodyGuardIDs {
		p, ok := r.store.BodyGuardProfiles[id]
		if !ok {
			r.store.Mu.RUnlock()
			return nil, apperr.New(404, "Body guard not found")
		}
		body += float64(days) * p.DailyRate
	}
	r.store.Mu.RUnlock()
	total := base + distance + driver + body
	return map[string]any{"basePrice": base, "distancePrice": distance, "driverPrice": driver, "bodyGuardPrice": body, "totalPrice": total, "durationInDays": days, "distanceKm": distanceKM}, nil
}
