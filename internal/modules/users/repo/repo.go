package repo

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type Repository interface {
	Create(data model.User) (model.User, error)
	FindByID(id string) (model.User, error)
	FindByEmail(email string) (model.User, error)
	FindAll(skip, take int, role *model.UserRole) ([]model.User, error)
	Update(id string, patch map[string]any) (model.User, error)
	Delete(id string) (model.User, error)
	FindByRole(role model.UserRole) ([]model.User, error)
}

type MemoryRepository struct{ store *storage.MemoryStore }

func NewMemoryRepository(store *storage.MemoryStore) *MemoryRepository {
	return &MemoryRepository{store: store}
}

func (r *MemoryRepository) Create(data model.User) (model.User, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	if _, ok := r.store.UserByEmail[data.Email]; ok {
		return model.User{}, apperr.New(409, "Unique constraint violation on field(s): email")
	}
	now := time.Now()
	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	data.CreatedAt, data.UpdatedAt = now, now
	r.store.Users[data.ID] = data
	r.store.UserByEmail[data.Email] = data.ID
	return data, nil
}

func (r *MemoryRepository) FindByID(id string) (model.User, error) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	u, ok := r.store.Users[id]
	if !ok {
		return model.User{}, apperr.New(404, "Record not found")
	}
	return u, nil
}

func (r *MemoryRepository) FindByEmail(email string) (model.User, error) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	id, ok := r.store.UserByEmail[email]
	if !ok {
		return model.User{}, apperr.New(404, "Record not found")
	}
	return r.store.Users[id], nil
}

func (r *MemoryRepository) FindAll(skip, take int, role *model.UserRole) ([]model.User, error) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	out := make([]model.User, 0)
	for _, u := range r.store.Users {
		if role != nil && u.Role != *role {
			continue
		}
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	if skip > len(out) {
		return []model.User{}, nil
	}
	out = out[skip:]
	if take > 0 && take < len(out) {
		out = out[:take]
	}
	return out, nil
}

func (r *MemoryRepository) Update(id string, patch map[string]any) (model.User, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	u, ok := r.store.Users[id]
	if !ok {
		return model.User{}, apperr.New(404, "Record not found")
	}
	if v, ok := patch["email"].(string); ok {
		if existingID, exists := r.store.UserByEmail[v]; exists && existingID != id {
			return model.User{}, apperr.New(409, "Unique constraint violation on field(s): email")
		}
		delete(r.store.UserByEmail, u.Email)
		u.Email = v
		r.store.UserByEmail[v] = id
	}
	if v, ok := patch["name"].(string); ok {
		u.Name = v
	}
	if v, ok := patch["phoneNumber"].(string); ok {
		u.PhoneNumber = &v
	}
	if v, ok := patch["profilePicture"].(string); ok {
		u.ProfilePicture = &v
	}
	if v, ok := patch["role"].(string); ok {
		u.Role = model.UserRole(v)
	}
	if v, ok := patch["isVerified"].(bool); ok {
		u.IsVerified = v
	}
	if v, ok := patch["password"].(*string); ok {
		u.Password = v
	}
	if v, ok := patch["authProvider"].(string); ok {
		u.AuthProvider = model.AuthProvider(v)
	}
	if v, ok := patch["verificationToken"].(*string); ok {
		u.VerificationToken = v
	}
	if v, ok := patch["refreshToken"].(*string); ok {
		u.RefreshToken = v
	}
	if v, ok := patch["refreshTokenExpiry"].(*time.Time); ok {
		u.RefreshTokenExpiry = v
	}
	u.UpdatedAt = time.Now()
	r.store.Users[id] = u
	return u, nil
}

func (r *MemoryRepository) Delete(id string) (model.User, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	u, ok := r.store.Users[id]
	if !ok {
		return model.User{}, apperr.New(404, "Record not found")
	}
	delete(r.store.Users, id)
	delete(r.store.UserByEmail, u.Email)
	return u, nil
}

func (r *MemoryRepository) FindByRole(role model.UserRole) ([]model.User, error) {
	return r.FindAll(0, 0, &role)
}
