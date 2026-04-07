package repo

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"pistoni/core/internal/modules/paymentmethods/dto"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type Repository interface {
	GetPaymentMethods(userID string) ([]model.PaymentMethod, error)
	CreateCardPaymentMethod(userID string, in dto.CreateCardPaymentMethodDTO, last4 string) (model.PaymentMethod, error)
	CreateMobilePaymentMethod(userID string, in dto.CreateMobilePaymentMethodDTO) (model.PaymentMethod, error)
	UpdatePaymentMethod(userID, id string, in dto.UpdatePaymentMethodDTO) (model.PaymentMethod, error)
	DeletePaymentMethod(userID, id string) error
	SetDefaultPaymentMethod(userID, id string) (model.PaymentMethod, error)
}

type MemoryRepository struct{ store *storage.MemoryStore }

func NewMemoryRepository(store *storage.MemoryStore) *MemoryRepository {
	return &MemoryRepository{store: store}
}

func (r *MemoryRepository) GetPaymentMethods(userID string) ([]model.PaymentMethod, error) {
	r.store.Mu.RLock()
	ids := append([]string{}, r.store.PaymentMethodByUser[userID]...)
	out := make([]model.PaymentMethod, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.store.PaymentMethods[id])
	}
	r.store.Mu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsDefault != out[j].IsDefault {
			return out[i].IsDefault
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}
func (r *MemoryRepository) CreateCardPaymentMethod(userID string, in dto.CreateCardPaymentMethodDTO, last4 string) (model.PaymentMethod, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	if in.IsDefault {
		r.unsetDefaultLocked(userID)
	}
	now := time.Now()
	ctype := in.CardType
	ch := in.CardholderName
	em := in.ExpiryMonth
	ey := in.ExpiryYear
	cn := last4
	m := model.PaymentMethod{ID: uuid.NewString(), UserID: userID, Type: model.PaymentMethodTypeCard, Status: model.PaymentMethodStatusActive, IsDefault: in.IsDefault, CardNumber: &cn, CardType: &ctype, ExpiryMonth: &em, ExpiryYear: &ey, CardholderName: &ch, CreatedAt: now, UpdatedAt: now}
	r.store.PaymentMethods[m.ID] = m
	r.store.PaymentMethodByUser[userID] = append(r.store.PaymentMethodByUser[userID], m.ID)
	return m, nil
}
func (r *MemoryRepository) CreateMobilePaymentMethod(userID string, in dto.CreateMobilePaymentMethodDTO) (model.PaymentMethod, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	if in.IsDefault {
		r.unsetDefaultLocked(userID)
	}
	now := time.Now()
	mn := in.MobileNumber
	p := in.Provider
	an := in.AccountName
	m := model.PaymentMethod{ID: uuid.NewString(), UserID: userID, Type: model.PaymentMethodTypeMobileMoney, Status: model.PaymentMethodStatusActive, IsDefault: in.IsDefault, MobileNumber: &mn, Provider: &p, AccountName: &an, CreatedAt: now, UpdatedAt: now}
	r.store.PaymentMethods[m.ID] = m
	r.store.PaymentMethodByUser[userID] = append(r.store.PaymentMethodByUser[userID], m.ID)
	return m, nil
}
func (r *MemoryRepository) UpdatePaymentMethod(userID, id string, in dto.UpdatePaymentMethodDTO) (model.PaymentMethod, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	m, ok := r.store.PaymentMethods[id]
	if !ok || m.UserID != userID {
		return model.PaymentMethod{}, apperr.New(404, "Payment method not found")
	}
	if in.IsDefault != nil {
		if *in.IsDefault {
			r.unsetDefaultLocked(userID)
		}
		m.IsDefault = *in.IsDefault
	}
	if in.Status != nil {
		m.Status = model.PaymentMethodStatus(*in.Status)
	}
	m.UpdatedAt = time.Now()
	r.store.PaymentMethods[id] = m
	return m, nil
}
func (r *MemoryRepository) DeletePaymentMethod(userID, id string) error {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	m, ok := r.store.PaymentMethods[id]
	if !ok || m.UserID != userID {
		return apperr.New(404, "Payment method not found")
	}
	if m.IsDefault {
		return apperr.New(400, "Cannot delete default payment method")
	}
	delete(r.store.PaymentMethods, id)
	ids := r.store.PaymentMethodByUser[userID]
	out := ids[:0]
	for _, mid := range ids {
		if mid != id {
			out = append(out, mid)
		}
	}
	r.store.PaymentMethodByUser[userID] = out
	return nil
}
func (r *MemoryRepository) SetDefaultPaymentMethod(userID, id string) (model.PaymentMethod, error) {
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()
	m, ok := r.store.PaymentMethods[id]
	if !ok || m.UserID != userID {
		return model.PaymentMethod{}, apperr.New(404, "Payment method not found")
	}
	r.unsetDefaultLocked(userID)
	m.IsDefault = true
	m.UpdatedAt = time.Now()
	r.store.PaymentMethods[id] = m
	return m, nil
}
func (r *MemoryRepository) unsetDefaultLocked(userID string) {
	for _, id := range r.store.PaymentMethodByUser[userID] {
		m := r.store.PaymentMethods[id]
		if m.IsDefault {
			m.IsDefault = false
			m.UpdatedAt = time.Now()
			r.store.PaymentMethods[id] = m
		}
	}
}
