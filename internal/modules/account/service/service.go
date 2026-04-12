package service

import (
	"context"

	"pistoni/core/internal/modules/account/dto"
	"pistoni/core/internal/modules/account/repo"
	"pistoni/core/internal/platform/model"
	platformredis "pistoni/core/internal/platform/redis"
)

type Service struct {
	repo  repo.Repository
	cache *platformredis.JSONCache
}

func New(r repo.Repository, cache *platformredis.JSONCache) *Service {
	return &Service{repo: r, cache: cache}
}

func (s *Service) GetOverview(userID string) (model.AccountOverview, error) {
	return remember(s.cache, accountCacheKey(userID, "overview"), func() (model.AccountOverview, error) {
		return s.repo.GetOverview(userID)
	})
}

func (s *Service) GetSettings(userID string) (model.UserSettings, error) {
	return remember(s.cache, accountCacheKey(userID, "settings"), func() (model.UserSettings, error) {
		return s.repo.GetSettings(userID)
	})
}

func (s *Service) UpdateSettings(userID string, in dto.UpdateSettingsDTO) (model.UserSettings, error) {
	out, err := s.repo.UpdateSettings(userID, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) GetSafety(userID string) (model.SafetyProfile, error) {
	return remember(s.cache, accountCacheKey(userID, "safety"), func() (model.SafetyProfile, error) {
		return s.repo.GetSafety(userID)
	})
}

func (s *Service) UpdateSafety(userID string, in dto.UpdateSafetyDTO) (model.SafetyProfile, error) {
	out, err := s.repo.UpdateSafety(userID, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) GetPrivacy(userID string) (model.PrivacySettings, error) {
	return remember(s.cache, accountCacheKey(userID, "privacy"), func() (model.PrivacySettings, error) {
		return s.repo.GetPrivacy(userID)
	})
}

func (s *Service) UpdatePrivacy(userID string, in dto.UpdatePrivacyDTO) (model.PrivacySettings, error) {
	out, err := s.repo.UpdatePrivacy(userID, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) GetSecurity(userID string) (model.SecuritySettings, error) {
	return remember(s.cache, accountCacheKey(userID, "security"), func() (model.SecuritySettings, error) {
		return s.repo.GetSecurity(userID)
	})
}

func (s *Service) UpdateSecurity(userID string, in dto.UpdateSecurityDTO) (model.SecuritySettings, error) {
	out, err := s.repo.UpdateSecurity(userID, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) ListTrustedContacts(userID string) ([]model.TrustedContact, error) {
	return remember(s.cache, accountCacheKey(userID, "trusted-contacts"), func() ([]model.TrustedContact, error) {
		return s.repo.ListTrustedContacts(userID)
	})
}

func (s *Service) CreateTrustedContact(userID string, in dto.CreateTrustedContactDTO) (model.TrustedContact, error) {
	out, err := s.repo.CreateTrustedContact(userID, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) UpdateTrustedContact(userID, id string, in dto.UpdateTrustedContactDTO) (model.TrustedContact, error) {
	out, err := s.repo.UpdateTrustedContact(userID, id, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) DeleteTrustedContact(userID, id string) (map[string]bool, error) {
	out, err := s.repo.DeleteTrustedContact(userID, id)
	s.invalidate(userID)
	return out, err
}

func (s *Service) ListFamilyMembers(userID string) ([]model.FamilyMember, error) {
	return remember(s.cache, accountCacheKey(userID, "family-members"), func() ([]model.FamilyMember, error) {
		return s.repo.ListFamilyMembers(userID)
	})
}

func (s *Service) CreateFamilyMember(userID string, in dto.CreateFamilyMemberDTO) (model.FamilyMember, error) {
	out, err := s.repo.CreateFamilyMember(userID, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) UpdateFamilyMember(userID, id string, in dto.UpdateFamilyMemberDTO) (model.FamilyMember, error) {
	out, err := s.repo.UpdateFamilyMember(userID, id, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) DeleteFamilyMember(userID, id string) (map[string]bool, error) {
	out, err := s.repo.DeleteFamilyMember(userID, id)
	s.invalidate(userID)
	return out, err
}

func (s *Service) ListTrustedDevices(userID string) ([]model.TrustedDevice, error) {
	return remember(s.cache, accountCacheKey(userID, "trusted-devices"), func() ([]model.TrustedDevice, error) {
		return s.repo.ListTrustedDevices(userID)
	})
}

func (s *Service) CreateTrustedDevice(userID string, in dto.CreateTrustedDeviceDTO) (model.TrustedDevice, error) {
	out, err := s.repo.CreateTrustedDevice(userID, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) UpdateTrustedDevice(userID, id string, in dto.UpdateTrustedDeviceDTO) (model.TrustedDevice, error) {
	out, err := s.repo.UpdateTrustedDevice(userID, id, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) DeleteTrustedDevice(userID, id string) (map[string]bool, error) {
	out, err := s.repo.DeleteTrustedDevice(userID, id)
	s.invalidate(userID)
	return out, err
}

func (s *Service) ListDocuments(userID string) ([]model.PersonalDocument, error) {
	return remember(s.cache, accountCacheKey(userID, "documents"), func() ([]model.PersonalDocument, error) {
		return s.repo.ListDocuments(userID)
	})
}

func (s *Service) CreateDocument(userID string, in dto.CreatePersonalDocumentDTO) (model.PersonalDocument, error) {
	out, err := s.repo.CreateDocument(userID, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) UpdateDocument(userID, id string, in dto.UpdatePersonalDocumentDTO) (model.PersonalDocument, error) {
	out, err := s.repo.UpdateDocument(userID, id, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) DeleteDocument(userID, id string) (map[string]bool, error) {
	out, err := s.repo.DeleteDocument(userID, id)
	s.invalidate(userID)
	return out, err
}

func (s *Service) ListCalendars(userID string) ([]model.CalendarConnection, error) {
	return remember(s.cache, accountCacheKey(userID, "calendars"), func() ([]model.CalendarConnection, error) {
		return s.repo.ListCalendars(userID)
	})
}

func (s *Service) CreateCalendar(userID string, in dto.CreateCalendarConnectionDTO) (model.CalendarConnection, error) {
	out, err := s.repo.CreateCalendar(userID, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) UpdateCalendar(userID, id string, in dto.UpdateCalendarConnectionDTO) (model.CalendarConnection, error) {
	out, err := s.repo.UpdateCalendar(userID, id, in)
	s.invalidate(userID)
	return out, err
}

func (s *Service) DeleteCalendar(userID, id string) (map[string]bool, error) {
	out, err := s.repo.DeleteCalendar(userID, id)
	s.invalidate(userID)
	return out, err
}

func (s *Service) DeleteAccount(userID string) (map[string]bool, error) {
	out, err := s.repo.DeleteAccount(userID)
	s.invalidate(userID)
	return out, err
}

func (s *Service) invalidate(userID string) {
	if s.cache == nil {
		return
	}
	ctx := context.Background()
	_ = s.cache.DeletePrefix(ctx, accountCacheKey(userID, ""))
	_ = s.cache.Delete(ctx, "auth:principal:"+userID)
}

func accountCacheKey(userID, suffix string) string {
	base := "account:" + userID + ":"
	if suffix == "" {
		return base
	}
	return base + suffix
}

func remember[T any](cache *platformredis.JSONCache, key string, loader func() (T, error)) (T, error) {
	var zero T
	if cache != nil {
		var cached T
		if ok, err := cache.Get(context.Background(), key, &cached); err == nil && ok {
			return cached, nil
		}
	}
	fresh, err := loader()
	if err != nil {
		return zero, err
	}
	if cache != nil {
		_ = cache.Set(context.Background(), key, fresh)
	}
	return fresh, nil
}
