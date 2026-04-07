package service

import (
	"regexp"
	"strings"

	"pistoni/core/internal/modules/paymentmethods/dto"
	"pistoni/core/internal/modules/paymentmethods/repo"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
)

type Service struct{ repo repo.Repository }

func New(r repo.Repository) *Service { return &Service{repo: r} }

func (s *Service) GetPaymentMethods(userID string) ([]model.PaymentMethod, error) {
	return s.repo.GetPaymentMethods(userID)
}

func (s *Service) CreateCardPaymentMethod(userID string, in dto.CreateCardPaymentMethodDTO) (model.PaymentMethod, error) {
	if !isValidCardNumber(in.CardNumber) {
		return model.PaymentMethod{}, apperr.New(400, "Invalid card number")
	}
	if in.ExpiryMonth < 1 || in.ExpiryMonth > 12 {
		return model.PaymentMethod{}, apperr.New(400, "Invalid expiry month")
	}
	clean := strings.NewReplacer(" ", "", "-", "").Replace(in.CardNumber)
	return s.repo.CreateCardPaymentMethod(userID, in, clean[len(clean)-4:])
}

func (s *Service) CreateMobilePaymentMethod(userID string, in dto.CreateMobilePaymentMethodDTO) (model.PaymentMethod, error) {
	if !isValidMobileNumber(in.MobileNumber) {
		return model.PaymentMethod{}, apperr.New(400, "Invalid mobile number format")
	}
	parts := strings.Fields(in.AccountName)
	if len(parts) != 3 {
		return model.PaymentMethod{}, apperr.New(400, "Account name must contain exactly 3 names")
	}
	return s.repo.CreateMobilePaymentMethod(userID, in)
}

func (s *Service) UpdatePaymentMethod(userID, id string, in dto.UpdatePaymentMethodDTO) (model.PaymentMethod, error) {
	return s.repo.UpdatePaymentMethod(userID, id, in)
}
func (s *Service) DeletePaymentMethod(userID, id string) error {
	return s.repo.DeletePaymentMethod(userID, id)
}
func (s *Service) SetDefaultPaymentMethod(userID, id string) (model.PaymentMethod, error) {
	return s.repo.SetDefaultPaymentMethod(userID, id)
}

func isValidCardNumber(cardNumber string) bool {
	clean := strings.NewReplacer(" ", "", "-", "").Replace(cardNumber)
	if !regexp.MustCompile(`^\d{13,19}$`).MatchString(clean) {
		return false
	}
	sum := 0
	alt := false
	for i := len(clean) - 1; i >= 0; i-- {
		d := int(clean[i] - '0')
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return sum%10 == 0
}
func isValidMobileNumber(v string) bool {
	clean := strings.NewReplacer(" ", "", "+", "", "(", "", ")", "", "-", "").Replace(v)
	return regexp.MustCompile(`^(255|0)?[0-9]{9}$`).MatchString(clean)
}
