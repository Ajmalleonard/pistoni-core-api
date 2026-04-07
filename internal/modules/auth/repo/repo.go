package repo

import (
	"time"

	"pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type PendingRepository interface {
	Upsert(email, code string, expiry time.Time) error
	FindByEmail(email string) (model.PendingEmailVerification, bool)
	Delete(email string) error
	DeleteExpired(email string, now time.Time) error
}

type MemoryPendingRepository struct{ store *storage.MemoryStore }

func NewMemoryPendingRepository(store *storage.MemoryStore) *MemoryPendingRepository {
	return &MemoryPendingRepository{store: store}
}

func (r *MemoryPendingRepository) Upsert(email, code string, expiry time.Time) error {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	r.store.PendingByEmail[email] = model.PendingEmailVerification{Email: email, Code: code, CodeExpiry: expiry, CreatedAt: time.Now()}
	return nil
}
func (r *MemoryPendingRepository) FindByEmail(email string) (model.PendingEmailVerification, bool) {
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	p, ok := r.store.PendingByEmail[email]
	return p, ok
}
func (r *MemoryPendingRepository) Delete(email string) error {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	delete(r.store.PendingByEmail, email)
	return nil
}
func (r *MemoryPendingRepository) DeleteExpired(email string, now time.Time) error {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	if p, ok := r.store.PendingByEmail[email]; ok && p.CodeExpiry.Before(now) {
		delete(r.store.PendingByEmail, email)
	}
	return nil
}

type UserRepository = repo.Repository
