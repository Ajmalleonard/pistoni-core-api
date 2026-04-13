package repo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"pistoni/core/internal/modules/providers/dto"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type Repository interface {
	RegisterDriver(userID string, in dto.RegisterDriverDTO) (model.DriverProfile, error)
	FindDrivers(q dto.FindProviderQuery) ([]map[string]any, error)
	FindDriverByID(id string) (map[string]any, error)
	GetMyDriverProfile(userID string) (map[string]any, error)
	UpdateMyDriverAvailability(userID string, isAvailable bool) (map[string]any, error)
	UpdateDriverVerification(id string, in dto.UpdateVerificationDTO) (map[string]any, error)
	GetDriverStatus(id string) (map[string]any, error)
	GetMyDriverStatus(userID string) (map[string]any, error)
	UpdateMyDriverStatus(userID string, in dto.UpdateDriverStatusDTO) (map[string]any, error)
	PollDriverStatusUpdates(q dto.DriverStatusPollQuery) (map[string]any, error)
	ListDriverReviews(driverID string) (map[string]any, error)
	CreateDriverReview(driverID, authorID string, in dto.CreateDriverReviewDTO) (map[string]any, error)
	GetDriverAssignmentCounters(driverID string) (map[string]any, error)
	UpdateDriverAssignmentCounters(driverID string, in dto.DriverAssignmentDeltaDTO) (map[string]any, error)

	RegisterBodyGuard(userID string, in dto.RegisterBodyGuardDTO) (model.BodyGuardProfile, error)
	FindBodyGuards(q dto.FindProviderQuery) ([]map[string]any, error)
	FindBodyGuardByID(id string) (map[string]any, error)
	GetMyBodyGuardProfile(userID string) (map[string]any, error)
	GetMyBodyGuardQueue(userID string) (map[string]any, error)
	UpdateMyBodyGuardAvailability(userID string, isAvailable bool) (map[string]any, error)
	UpdateBodyGuardVerification(id string, in dto.UpdateVerificationDTO) (map[string]any, error)
	GetBodyGuardStatus(id string) (map[string]any, error)
	GetMyBodyGuardStatus(userID string) (map[string]any, error)
	UpdateMyBodyGuardStatus(userID string, in dto.UpdateBodyGuardStatusDTO) (map[string]any, error)
	JoinBodyGuardQueue(userID string, in dto.JoinBodyGuardQueueDTO) (map[string]any, error)
	ListBodyGuardQueue(status string) ([]map[string]any, error)
	ApproveBodyGuardQueue(id, reviewerID string) (map[string]any, error)
	RejectBodyGuardQueue(id, reviewerID, reason string) (map[string]any, error)
}

type MemoryRepository struct {
	store *storage.MemoryStore
	users usersrepo.Repository
}

func NewMemoryRepository(store *storage.MemoryStore, users usersrepo.Repository) *MemoryRepository {
	return &MemoryRepository{store: store, users: users}
}

func (r *MemoryRepository) RegisterDriver(userID string, in dto.RegisterDriverDTO) (model.DriverProfile, error) {
	u, err := r.users.FindByID(userID)
	if err != nil {
		return model.DriverProfile{}, apperr.New(404, "User not found")
	}
	if u.Role == model.UserRoleClient {
		_, _ = r.users.Update(userID, map[string]any{"role": string(model.UserRoleDriver)})
	}

	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	if _, exists := r.store.DriverProfileByUser[userID]; exists {
		return model.DriverProfile{}, apperr.New(409, "Driver profile already exists for this user")
	}
	now := time.Now()
	r.store.DriverStatusVersion++
	p := model.DriverProfile{ID: uuid.NewString(), UserID: userID, PistoniIDNumber: fmt.Sprintf("PST-DRV-%d", time.Now().UnixNano()%1_000_000_000),
		DriversLicenseNumber: in.DriversLicenseNumber, DailyRate: in.DailyRate, VerificationStatus: model.VerificationPending, IsAvailable: true,
		Status: model.DriverStatusOffline, StatusVersion: r.store.DriverStatusVersion, StatusUpdatedAt: now, CreatedAt: now, UpdatedAt: now}
	if in.Gender != "" {
		g := model.Gender(in.Gender)
		p.Gender = &g
	}
	if in.Address != "" {
		p.Address = &in.Address
	}
	if in.NIN != "" {
		p.NIN = &in.NIN
	}
	if in.DriversLicenseClass != "" {
		p.DriversLicenseClass = &in.DriversLicenseClass
	}
	if in.IssuingAuthority != "" {
		p.IssuingAuthority = &in.IssuingAuthority
	}
	if in.YearsOfExperience != nil {
		p.YearsOfExperience = in.YearsOfExperience
	}
	if in.EmergencyContactName != "" {
		p.EmergencyContactName = &in.EmergencyContactName
	}
	if in.EmergencyContactPhone != "" {
		p.EmergencyContactPhone = &in.EmergencyContactPhone
	}
	if in.Notes != "" {
		p.Notes = &in.Notes
	}
	r.store.DriverProfiles[p.ID] = p
	r.store.DriverProfileByUser[userID] = p.ID
	return p, nil
}

func (r *MemoryRepository) FindDrivers(q dto.FindProviderQuery) ([]map[string]any, error) {
	r.store.Mu.RLock()
	items := []model.DriverProfile{}
	for _, p := range r.store.DriverProfiles {
		if q.IsAvailable != nil && p.IsAvailable != *q.IsAvailable {
			continue
		}
		status := q.VerificationStatus
		if status == "" {
			status = string(model.VerificationApproved)
		}
		if string(p.VerificationStatus) != status {
			continue
		}
		items = append(items, p)
	}
	r.store.Mu.RUnlock()
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return r.driverViews(items, false), nil
}

func (r *MemoryRepository) FindDriverByID(id string) (map[string]any, error) {
	r.store.Mu.RLock()
	p, ok := r.store.DriverProfiles[id]
	r.store.Mu.RUnlock()
	if !ok {
		return nil, apperr.New(404, "Driver not found")
	}
	return r.driverView(p, false), nil
}

func (r *MemoryRepository) GetMyDriverProfile(userID string) (map[string]any, error) {
	r.store.Mu.RLock()
	id, ok := r.store.DriverProfileByUser[userID]
	if !ok {
		r.store.Mu.RUnlock()
		return nil, apperr.New(404, "Driver profile not found")
	}
	p := r.store.DriverProfiles[id]
	r.store.Mu.RUnlock()
	return r.driverView(p, true), nil
}

func (r *MemoryRepository) UpdateMyDriverAvailability(userID string, isAvailable bool) (map[string]any, error) {
	r.store.Mu.Lock()
	id, ok := r.store.DriverProfileByUser[userID]
	if !ok {
		r.store.Mu.Unlock()
		return nil, apperr.New(404, "Driver profile not found")
	}
	p := r.store.DriverProfiles[id]
	p.IsAvailable = isAvailable
	p.UpdatedAt = time.Now()
	r.store.DriverProfiles[id] = p
	r.store.Mu.Unlock()
	return r.driverView(p, true), nil
}

func (r *MemoryRepository) UpdateDriverVerification(id string, in dto.UpdateVerificationDTO) (map[string]any, error) {
	r.store.Mu.Lock()
	p, ok := r.store.DriverProfiles[id]
	if !ok {
		r.store.Mu.Unlock()
		return nil, apperr.New(404, "Driver not found")
	}
	p.VerificationStatus = model.ProviderVerificationStatus(in.VerificationStatus)
	if in.BackgroundCheckCompleted != nil {
		p.BackgroundCheckCompleted = *in.BackgroundCheckCompleted
	}
	if in.IsAvailable != nil {
		p.IsAvailable = *in.IsAvailable
	}
	if in.Notes != "" {
		p.Notes = &in.Notes
	}
	p.UpdatedAt = time.Now()
	r.store.DriverProfiles[id] = p
	r.store.Mu.Unlock()
	return r.driverView(p, true), nil
}

func (r *MemoryRepository) GetDriverStatus(id string) (map[string]any, error) {
	r.store.Mu.RLock()
	p, ok := r.store.DriverProfiles[id]
	r.store.Mu.RUnlock()
	if !ok {
		return nil, apperr.New(404, "Driver not found")
	}
	return map[string]any{"driverId": p.ID, "status": p.Status, "isAvailable": p.IsAvailable, "updatedAt": p.StatusUpdatedAt, "version": p.StatusVersion, "lastSeen": p.LastSeen}, nil
}

func (r *MemoryRepository) GetMyDriverStatus(userID string) (map[string]any, error) {
	r.store.Mu.RLock()
	id, ok := r.store.DriverProfileByUser[userID]
	if !ok {
		r.store.Mu.RUnlock()
		return nil, apperr.New(404, "Driver profile not found")
	}
	p := r.store.DriverProfiles[id]
	r.store.Mu.RUnlock()
	return map[string]any{"driverId": p.ID, "status": p.Status, "isAvailable": p.IsAvailable, "updatedAt": p.StatusUpdatedAt, "version": p.StatusVersion, "lastSeen": p.LastSeen}, nil
}

func (r *MemoryRepository) UpdateMyDriverStatus(userID string, in dto.UpdateDriverStatusDTO) (map[string]any, error) {
	r.store.Mu.Lock()
	id, ok := r.store.DriverProfileByUser[userID]
	if !ok {
		r.store.Mu.Unlock()
		return nil, apperr.New(404, "Driver profile not found")
	}
	p := r.store.DriverProfiles[id]
	p.Status = model.DriverStatus(in.Status)
	if in.IsAvailable != nil {
		p.IsAvailable = *in.IsAvailable
	}
	now := time.Now()
	p.LastSeen = &now
	p.UpdatedAt = now
	r.store.DriverStatusVersion++
	p.StatusVersion = r.store.DriverStatusVersion
	p.StatusUpdatedAt = now
	r.store.DriverProfiles[id] = p
	r.store.Mu.Unlock()
	return map[string]any{"driverId": p.ID, "status": p.Status, "isAvailable": p.IsAvailable, "updatedAt": p.StatusUpdatedAt, "version": p.StatusVersion, "lastSeen": p.LastSeen}, nil
}

func (r *MemoryRepository) PollDriverStatusUpdates(q dto.DriverStatusPollQuery) (map[string]any, error) {
	r.store.Mu.RLock()
	updates := make([]map[string]any, 0)
	since := time.UnixMilli(q.SinceUnixMs)
	latest := r.store.DriverStatusVersion
	for _, p := range r.store.DriverProfiles {
		if q.SinceVersion > 0 && p.StatusVersion <= q.SinceVersion {
			continue
		}
		if q.SinceUnixMs > 0 && p.StatusUpdatedAt.Before(since) {
			continue
		}
		updates = append(updates, map[string]any{"driverId": p.ID, "status": p.Status, "isAvailable": p.IsAvailable, "updatedAt": p.StatusUpdatedAt, "version": p.StatusVersion, "lastSeen": p.LastSeen})
	}
	r.store.Mu.RUnlock()
	sort.Slice(updates, func(i, j int) bool {
		return updates[i]["version"].(int64) < updates[j]["version"].(int64)
	})
	if q.Take <= 0 {
		q.Take = 100
	}
	if len(updates) > q.Take {
		updates = updates[:q.Take]
	}
	return map[string]any{"updates": updates, "latestVersion": latest, "nowUnixMs": time.Now().UnixMilli()}, nil
}

func (r *MemoryRepository) ListDriverReviews(driverID string) (map[string]any, error) {
	r.store.Mu.RLock()
	if _, ok := r.store.DriverProfiles[driverID]; !ok {
		r.store.Mu.RUnlock()
		return nil, apperr.New(404, "Driver not found")
	}
	ids := r.store.DriverReviewsByID[driverID]
	reviews := make([]model.DriverReview, 0, len(ids))
	for _, id := range ids {
		reviews = append(reviews, r.store.DriverReviews[id])
	}
	r.store.Mu.RUnlock()
	sort.Slice(reviews, func(i, j int) bool { return reviews[i].CreatedAt.After(reviews[j].CreatedAt) })
	count, avg := r.driverReviewAggregate(driverID)
	return map[string]any{"driverId": driverID, "reviews": reviews, "aggregate": map[string]any{"count": count, "avgRating": avg}}, nil
}

func (r *MemoryRepository) CreateDriverReview(driverID, authorID string, in dto.CreateDriverReviewDTO) (map[string]any, error) {
	r.store.Mu.Lock()
	if _, ok := r.store.DriverProfiles[driverID]; !ok {
		r.store.Mu.Unlock()
		return nil, apperr.New(404, "Driver not found")
	}
	now := time.Now()
	review := model.DriverReview{ID: uuid.NewString(), DriverID: driverID, AuthorID: authorID, Rating: in.Rating, CreatedAt: now}
	if in.Comment != "" {
		review.Comment = &in.Comment
	}
	r.store.DriverReviews[review.ID] = review
	r.store.DriverReviewsByID[driverID] = append(r.store.DriverReviewsByID[driverID], review.ID)
	r.store.Mu.Unlock()
	count, avg := r.driverReviewAggregate(driverID)
	return map[string]any{"review": review, "aggregate": map[string]any{"count": count, "avgRating": avg}}, nil
}

func (r *MemoryRepository) GetDriverAssignmentCounters(driverID string) (map[string]any, error) {
	r.store.Mu.RLock()
	p, ok := r.store.DriverProfiles[driverID]
	r.store.Mu.RUnlock()
	if !ok {
		return nil, apperr.New(404, "Driver not found")
	}
	return map[string]any{"driverId": p.ID, "routesCount": p.RoutesCount, "assignmentCount": p.AssignmentCount}, nil
}

func (r *MemoryRepository) UpdateDriverAssignmentCounters(driverID string, in dto.DriverAssignmentDeltaDTO) (map[string]any, error) {
	r.store.Mu.Lock()
	p, ok := r.store.DriverProfiles[driverID]
	if !ok {
		r.store.Mu.Unlock()
		return nil, apperr.New(404, "Driver not found")
	}
	p.RoutesCount += in.RouteDelta
	if p.RoutesCount < 0 {
		p.RoutesCount = 0
	}
	p.AssignmentCount += in.AssignmentDelta
	if p.AssignmentCount < 0 {
		p.AssignmentCount = 0
	}
	p.UpdatedAt = time.Now()
	r.store.DriverProfiles[driverID] = p
	r.store.Mu.Unlock()
	return map[string]any{"driverId": p.ID, "routesCount": p.RoutesCount, "assignmentCount": p.AssignmentCount}, nil
}

func (r *MemoryRepository) RegisterBodyGuard(userID string, in dto.RegisterBodyGuardDTO) (model.BodyGuardProfile, error) {
	u, err := r.users.FindByID(userID)
	if err != nil {
		return model.BodyGuardProfile{}, apperr.New(404, "User not found")
	}
	if u.Role == model.UserRoleClient {
		_, _ = r.users.Update(userID, map[string]any{"role": string(model.UserRoleBodyGuard)})
	}

	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	if _, exists := r.store.BodyGuardByUser[userID]; exists {
		return model.BodyGuardProfile{}, apperr.New(409, "Body guard profile already exists for this user")
	}
	now := time.Now()
	r.store.BodyGuardStatusVersion++
	p := model.BodyGuardProfile{ID: uuid.NewString(), UserID: userID, PistoniIDNumber: fmt.Sprintf("PST-BDG-%d", time.Now().UnixNano()%1_000_000_000),
		Gender: model.Gender(in.Gender), NIN: in.NIN, GovernmentIDNumber: in.GovernmentIDNumber, DailyRate: in.DailyRate,
		VerificationStatus: model.VerificationPending, IsAvailable: true, Status: model.BodyGuardStatusOffline, StatusVersion: r.store.BodyGuardStatusVersion,
		StatusUpdatedAt: now, CreatedAt: now, UpdatedAt: now}
	if in.Address != "" {
		p.Address = &in.Address
	}
	if in.SecurityLicenseNumber != "" {
		p.SecurityLicenseNumber = &in.SecurityLicenseNumber
	}
	if in.IssuingAuthority != "" {
		p.IssuingAuthority = &in.IssuingAuthority
	}
	if in.YearsOfExperience != nil {
		p.YearsOfExperience = in.YearsOfExperience
	}
	if in.EmergencyContactName != "" {
		p.EmergencyContactName = &in.EmergencyContactName
	}
	if in.EmergencyContactPhone != "" {
		p.EmergencyContactPhone = &in.EmergencyContactPhone
	}
	if in.Notes != "" {
		p.Notes = &in.Notes
	}
	r.store.BodyGuardProfiles[p.ID] = p
	r.store.BodyGuardByUser[userID] = p.ID
	return p, nil
}

func (r *MemoryRepository) FindBodyGuards(q dto.FindProviderQuery) ([]map[string]any, error) {
	if dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); dbURL != "" {
		if rows, err := findBodyGuardsFromDB(dbURL, q); err == nil {
			return rows, nil
		} else {
			log.Printf("[providers] db read failed for body guards, falling back to memory: %v", err)
		}
	}

	r.store.Mu.RLock()
	items := []model.BodyGuardProfile{}
	for _, p := range r.store.BodyGuardProfiles {
		if q.IsAvailable != nil && p.IsAvailable != *q.IsAvailable {
			continue
		}
		status := q.VerificationStatus
		if status == "" {
			status = string(model.VerificationApproved)
		}
		if string(p.VerificationStatus) != status {
			continue
		}
		if q.Gender != "" && string(p.Gender) != q.Gender {
			continue
		}
		items = append(items, p)
	}
	r.store.Mu.RUnlock()
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return r.bodyGuardViews(items, false), nil
}

func findBodyGuardsFromDB(dbURL string, q dto.FindProviderQuery) ([]map[string]any, error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, err
	}

	status := q.VerificationStatus
	if status == "" {
		status = string(model.VerificationApproved)
	}

	where := []string{"bg.\"verificationStatus\" = $1"}
	args := []any{status}
	idx := 2

	if q.IsAvailable != nil {
		where = append(where, fmt.Sprintf("bg.\"isAvailable\" = $%d", idx))
		args = append(args, *q.IsAvailable)
		idx++
	}
	if q.Gender != "" {
		where = append(where, fmt.Sprintf("bg.gender = $%d", idx))
		args = append(args, q.Gender)
		idx++
	}

	query := `
		SELECT
			bg.id,
			bg."pistoniIdNumber",
			bg.gender,
			bg."yearsOfExperience",
			bg."dailyRate",
			bg."isAvailable",
			bg."verificationStatus",
			u.id,
			u.name,
			u."phoneNumber",
			u."profilePicture",
			COALESCE(u.status, 'OFFLINE')
		FROM "BodyGuardProfile" bg
		JOIN "User" u ON u.id = bg."userId"
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY bg."createdAt" DESC
	`

	rs, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	out := make([]map[string]any, 0)
	for rs.Next() {
		var (
			id, pistoniID, gender, verificationStatus string
			userID, userName, statusStr               string
			yearsOfExperience                         sql.NullInt32
			dailyRate                                 float64
			isAvailable                               bool
			phoneNumber, profilePicture               sql.NullString
		)
		if err := rs.Scan(
			&id,
			&pistoniID,
			&gender,
			&yearsOfExperience,
			&dailyRate,
			&isAvailable,
			&verificationStatus,
			&userID,
			&userName,
			&phoneNumber,
			&profilePicture,
			&statusStr,
		); err != nil {
			return nil, err
		}

		var years *int
		if yearsOfExperience.Valid {
			v := int(yearsOfExperience.Int32)
			years = &v
		}
		var phonePtr *string
		if phoneNumber.Valid {
			v := phoneNumber.String
			phonePtr = &v
		}
		var picPtr *string
		if profilePicture.Valid {
			v := profilePicture.String
			picPtr = &v
		}

		out = append(out, map[string]any{
			"id":                 id,
			"pistoniIdNumber":    pistoniID,
			"gender":             gender,
			"yearsOfExperience":  years,
			"dailyRate":          dailyRate,
			"isAvailable":        isAvailable,
			"status":             statusStr,
			"verificationStatus": verificationStatus,
			"user": map[string]any{
				"id":             userID,
				"name":           userName,
				"phoneNumber":    phonePtr,
				"profilePicture": picPtr,
			},
		})
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *MemoryRepository) FindBodyGuardByID(id string) (map[string]any, error) {
	r.store.Mu.RLock()
	p, ok := r.store.BodyGuardProfiles[id]
	r.store.Mu.RUnlock()
	if !ok {
		return nil, apperr.New(404, "Body guard not found")
	}
	return r.bodyGuardView(p, false), nil
}

func (r *MemoryRepository) GetMyBodyGuardProfile(userID string) (map[string]any, error) {
	r.store.Mu.RLock()
	id, ok := r.store.BodyGuardByUser[userID]
	if !ok {
		r.store.Mu.RUnlock()
		return nil, apperr.New(404, "Body guard profile not found")
	}
	p := r.store.BodyGuardProfiles[id]
	r.store.Mu.RUnlock()
	return r.bodyGuardView(p, true), nil
}

func (r *MemoryRepository) GetMyBodyGuardQueue(userID string) (map[string]any, error) {
	r.store.Mu.RLock()
	id, ok := r.store.BodyGuardQueueByUser[userID]
	if !ok {
		r.store.Mu.RUnlock()
		return nil, apperr.New(404, "Body guard application not found")
	}
	item := r.store.BodyGuardQueue[id]
	r.store.Mu.RUnlock()
	return map[string]any{"queueItem": item}, nil
}

func (r *MemoryRepository) UpdateMyBodyGuardAvailability(userID string, isAvailable bool) (map[string]any, error) {
	r.store.Mu.Lock()
	id, ok := r.store.BodyGuardByUser[userID]
	if !ok {
		r.store.Mu.Unlock()
		return nil, apperr.New(404, "Body guard profile not found")
	}
	p := r.store.BodyGuardProfiles[id]
	p.IsAvailable = isAvailable
	p.UpdatedAt = time.Now()
	r.store.BodyGuardProfiles[id] = p
	r.store.Mu.Unlock()
	return r.bodyGuardView(p, true), nil
}

func (r *MemoryRepository) UpdateBodyGuardVerification(id string, in dto.UpdateVerificationDTO) (map[string]any, error) {
	r.store.Mu.Lock()
	p, ok := r.store.BodyGuardProfiles[id]
	if !ok {
		r.store.Mu.Unlock()
		return nil, apperr.New(404, "Body guard not found")
	}
	p.VerificationStatus = model.ProviderVerificationStatus(in.VerificationStatus)
	if in.BackgroundCheckCompleted != nil {
		p.BackgroundCheckCompleted = *in.BackgroundCheckCompleted
	}
	if in.IsAvailable != nil {
		p.IsAvailable = *in.IsAvailable
	}
	if in.Notes != "" {
		p.Notes = &in.Notes
	}
	p.UpdatedAt = time.Now()
	r.store.BodyGuardProfiles[id] = p
	r.store.Mu.Unlock()
	return r.bodyGuardView(p, true), nil
}

func (r *MemoryRepository) GetBodyGuardStatus(id string) (map[string]any, error) {
	r.store.Mu.RLock()
	p, ok := r.store.BodyGuardProfiles[id]
	r.store.Mu.RUnlock()
	if !ok {
		return nil, apperr.New(404, "Body guard not found")
	}
	return map[string]any{"bodyGuardId": p.ID, "status": p.Status, "isAvailable": p.IsAvailable, "updatedAt": p.StatusUpdatedAt, "version": p.StatusVersion, "lastSeen": p.LastSeen}, nil
}

func (r *MemoryRepository) GetMyBodyGuardStatus(userID string) (map[string]any, error) {
	r.store.Mu.RLock()
	id, ok := r.store.BodyGuardByUser[userID]
	if !ok {
		r.store.Mu.RUnlock()
		return nil, apperr.New(404, "Body guard profile not found")
	}
	p := r.store.BodyGuardProfiles[id]
	r.store.Mu.RUnlock()
	return map[string]any{"bodyGuardId": p.ID, "status": p.Status, "isAvailable": p.IsAvailable, "updatedAt": p.StatusUpdatedAt, "version": p.StatusVersion, "lastSeen": p.LastSeen}, nil
}

func (r *MemoryRepository) UpdateMyBodyGuardStatus(userID string, in dto.UpdateBodyGuardStatusDTO) (map[string]any, error) {
	r.store.Mu.Lock()
	id, ok := r.store.BodyGuardByUser[userID]
	if !ok {
		r.store.Mu.Unlock()
		return nil, apperr.New(404, "Body guard profile not found")
	}
	p := r.store.BodyGuardProfiles[id]
	p.Status = model.BodyGuardStatus(in.Status)
	if in.IsAvailable != nil {
		p.IsAvailable = *in.IsAvailable
	}
	now := time.Now()
	p.LastSeen = &now
	p.UpdatedAt = now
	r.store.BodyGuardStatusVersion++
	p.StatusVersion = r.store.BodyGuardStatusVersion
	p.StatusUpdatedAt = now
	r.store.BodyGuardProfiles[id] = p
	r.store.Mu.Unlock()
	return map[string]any{"bodyGuardId": p.ID, "status": p.Status, "isAvailable": p.IsAvailable, "updatedAt": p.StatusUpdatedAt, "version": p.StatusVersion, "lastSeen": p.LastSeen}, nil
}

func (r *MemoryRepository) JoinBodyGuardQueue(userID string, in dto.JoinBodyGuardQueueDTO) (map[string]any, error) {
	if _, err := r.users.FindByID(userID); err != nil {
		return nil, apperr.New(404, "User not found")
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	if _, exists := r.store.BodyGuardByUser[userID]; exists {
		return nil, apperr.New(409, "Body guard profile already exists")
	}
	if queuedID, exists := r.store.BodyGuardQueueByUser[userID]; exists {
		item := r.store.BodyGuardQueue[queuedID]
		return map[string]any{"queueItem": item}, nil
	}
	item := model.BodyGuardRegistrationQueueItem{
		ID:                 uuid.NewString(),
		UserID:             userID,
		Gender:             model.Gender(in.Gender),
		NIN:                in.NIN,
		GovernmentIDNumber: in.GovernmentIDNumber,
		DailyRate:          in.DailyRate,
		Status:             model.BodyGuardQueuePending,
		SubmittedAt:        time.Now(),
	}
	r.store.BodyGuardQueue[item.ID] = item
	r.store.BodyGuardQueueByUser[userID] = item.ID
	return map[string]any{"queueItem": item}, nil
}

func (r *MemoryRepository) ListBodyGuardQueue(status string) ([]map[string]any, error) {
	r.store.Mu.RLock()
	items := make([]model.BodyGuardRegistrationQueueItem, 0, len(r.store.BodyGuardQueue))
	for _, item := range r.store.BodyGuardQueue {
		if status != "" && string(item.Status) != status {
			continue
		}
		items = append(items, item)
	}
	r.store.Mu.RUnlock()
	sort.Slice(items, func(i, j int) bool { return items[i].SubmittedAt.After(items[j].SubmittedAt) })
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{"id": item.ID, "userId": item.UserID, "gender": item.Gender, "nin": item.NIN, "governmentIdNumber": item.GovernmentIDNumber,
			"dailyRate": item.DailyRate, "status": item.Status, "reason": item.Reason, "submittedAt": item.SubmittedAt, "reviewedAt": item.ReviewedAt, "reviewedBy": item.ReviewedBy})
	}
	return out, nil
}

func (r *MemoryRepository) ApproveBodyGuardQueue(id, reviewerID string) (map[string]any, error) {
	r.store.Mu.Lock()
	item, ok := r.store.BodyGuardQueue[id]
	if !ok {
		r.store.Mu.Unlock()
		return nil, apperr.New(404, "Queue item not found")
	}
	if item.Status != model.BodyGuardQueuePending {
		r.store.Mu.Unlock()
		return nil, apperr.New(409, "Queue item already reviewed")
	}
	now := time.Now()
	item.Status = model.BodyGuardQueueApproved
	item.ReviewedAt = &now
	item.ReviewedBy = &reviewerID
	r.store.BodyGuardQueue[id] = item

	r.store.BodyGuardStatusVersion++
	p := model.BodyGuardProfile{ID: uuid.NewString(), UserID: item.UserID, PistoniIDNumber: fmt.Sprintf("PST-BDG-%d", time.Now().UnixNano()%1_000_000_000),
		Gender: item.Gender, NIN: item.NIN, GovernmentIDNumber: item.GovernmentIDNumber, DailyRate: item.DailyRate,
		VerificationStatus: model.VerificationPending, IsAvailable: true, Status: model.BodyGuardStatusOffline, StatusVersion: r.store.BodyGuardStatusVersion,
		StatusUpdatedAt: now, CreatedAt: now, UpdatedAt: now}
	r.store.BodyGuardProfiles[p.ID] = p
	r.store.BodyGuardByUser[p.UserID] = p.ID
	r.store.Mu.Unlock()
	_, _ = r.users.Update(item.UserID, map[string]any{"role": string(model.UserRoleBodyGuard)})
	return map[string]any{"queueItem": item, "bodyGuardProfileId": p.ID}, nil
}

func (r *MemoryRepository) RejectBodyGuardQueue(id, reviewerID, reason string) (map[string]any, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	item, ok := r.store.BodyGuardQueue[id]
	if !ok {
		return nil, apperr.New(404, "Queue item not found")
	}
	if item.Status != model.BodyGuardQueuePending {
		return nil, apperr.New(409, "Queue item already reviewed")
	}
	now := time.Now()
	item.Status = model.BodyGuardQueueRejected
	if reason != "" {
		item.Reason = &reason
	}
	item.ReviewedAt = &now
	item.ReviewedBy = &reviewerID
	r.store.BodyGuardQueue[id] = item
	return map[string]any{"queueItem": item}, nil
}

func (r *MemoryRepository) driverViews(items []model.DriverProfile, private bool) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, p := range items {
		out = append(out, r.driverView(p, private))
	}
	return out
}

func (r *MemoryRepository) bodyGuardViews(items []model.BodyGuardProfile, private bool) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, p := range items {
		out = append(out, r.bodyGuardView(p, private))
	}
	return out
}

func (r *MemoryRepository) driverView(p model.DriverProfile, private bool) map[string]any {
	u, _ := r.users.FindByID(p.UserID)
	count, avg := r.driverReviewAggregate(p.ID)
	pub := map[string]any{"id": u.ID, "name": u.Name, "phoneNumber": u.PhoneNumber, "profilePicture": u.ProfilePicture}
	if !private {
		return map[string]any{"id": p.ID, "pistoniIdNumber": p.PistoniIDNumber, "gender": p.Gender, "yearsOfExperience": p.YearsOfExperience,
			"dailyRate": p.DailyRate, "isAvailable": p.IsAvailable, "status": p.Status, "verificationStatus": p.VerificationStatus,
			"rating": map[string]any{"count": count, "avgRating": avg}, "routesCount": p.RoutesCount, "assignmentCount": p.AssignmentCount, "user": pub}
	}
	priv := map[string]any{"id": u.ID, "email": u.Email, "name": u.Name, "phoneNumber": u.PhoneNumber, "profilePicture": u.ProfilePicture, "role": u.Role, "isVerified": u.IsVerified}
	return map[string]any{"id": p.ID, "pistoniIdNumber": p.PistoniIDNumber, "gender": p.Gender, "dateOfBirth": p.DateOfBirth, "address": p.Address,
		"nin": p.NIN, "driversLicenseNumber": p.DriversLicenseNumber, "driversLicenseClass": p.DriversLicenseClass, "driversLicenseExpiry": p.DriversLicenseExpiry,
		"issuingAuthority": p.IssuingAuthority, "yearsOfExperience": p.YearsOfExperience, "emergencyContactName": p.EmergencyContactName,
		"emergencyContactPhone": p.EmergencyContactPhone, "dailyRate": p.DailyRate, "verificationStatus": p.VerificationStatus,
		"isAvailable": p.IsAvailable, "status": p.Status, "statusVersion": p.StatusVersion, "statusUpdatedAt": p.StatusUpdatedAt, "lastSeen": p.LastSeen,
		"routesCount": p.RoutesCount, "assignmentCount": p.AssignmentCount,
		"rating":                   map[string]any{"count": count, "avgRating": avg},
		"backgroundCheckCompleted": p.BackgroundCheckCompleted, "notes": p.Notes, "createdAt": p.CreatedAt, "updatedAt": p.UpdatedAt,
		"user": priv}
}

func (r *MemoryRepository) bodyGuardView(p model.BodyGuardProfile, private bool) map[string]any {
	u, _ := r.users.FindByID(p.UserID)
	pub := map[string]any{"id": u.ID, "name": u.Name, "phoneNumber": u.PhoneNumber, "profilePicture": u.ProfilePicture}
	if !private {
		return map[string]any{"id": p.ID, "pistoniIdNumber": p.PistoniIDNumber, "gender": p.Gender, "yearsOfExperience": p.YearsOfExperience,
			"dailyRate": p.DailyRate, "isAvailable": p.IsAvailable, "status": p.Status, "verificationStatus": p.VerificationStatus, "user": pub}
	}
	priv := map[string]any{"id": u.ID, "email": u.Email, "name": u.Name, "phoneNumber": u.PhoneNumber, "profilePicture": u.ProfilePicture, "role": u.Role, "isVerified": u.IsVerified}
	return map[string]any{"id": p.ID, "pistoniIdNumber": p.PistoniIDNumber, "gender": p.Gender, "dateOfBirth": p.DateOfBirth, "address": p.Address,
		"nin": p.NIN, "governmentIdNumber": p.GovernmentIDNumber, "securityLicenseNumber": p.SecurityLicenseNumber, "securityLicenseExpiry": p.SecurityLicenseExpiry,
		"issuingAuthority": p.IssuingAuthority, "yearsOfExperience": p.YearsOfExperience, "emergencyContactName": p.EmergencyContactName,
		"emergencyContactPhone": p.EmergencyContactPhone, "dailyRate": p.DailyRate, "verificationStatus": p.VerificationStatus,
		"isAvailable": p.IsAvailable, "status": p.Status, "statusVersion": p.StatusVersion, "statusUpdatedAt": p.StatusUpdatedAt, "lastSeen": p.LastSeen,
		"backgroundCheckCompleted": p.BackgroundCheckCompleted, "notes": p.Notes, "createdAt": p.CreatedAt, "updatedAt": p.UpdatedAt,
		"user": priv}
}

func (r *MemoryRepository) driverReviewAggregate(driverID string) (int, float64) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	ids := r.store.DriverReviewsByID[driverID]
	if len(ids) == 0 {
		return 0, 0
	}
	total := 0
	for _, id := range ids {
		total += r.store.DriverReviews[id].Rating
	}
	return len(ids), float64(total) / float64(len(ids))
}
