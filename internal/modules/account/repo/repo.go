package repo

import (
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"pistoni/core/internal/modules/account/dto"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type Repository interface {
	GetOverview(userID string) (model.AccountOverview, error)
	GetSettings(userID string) (model.UserSettings, error)
	UpdateSettings(userID string, in dto.UpdateSettingsDTO) (model.UserSettings, error)
	GetSafety(userID string) (model.SafetyProfile, error)
	UpdateSafety(userID string, in dto.UpdateSafetyDTO) (model.SafetyProfile, error)
	GetPrivacy(userID string) (model.PrivacySettings, error)
	UpdatePrivacy(userID string, in dto.UpdatePrivacyDTO) (model.PrivacySettings, error)
	GetSecurity(userID string) (model.SecuritySettings, error)
	UpdateSecurity(userID string, in dto.UpdateSecurityDTO) (model.SecuritySettings, error)
	ListTrustedContacts(userID string) ([]model.TrustedContact, error)
	CreateTrustedContact(userID string, in dto.CreateTrustedContactDTO) (model.TrustedContact, error)
	UpdateTrustedContact(userID, id string, in dto.UpdateTrustedContactDTO) (model.TrustedContact, error)
	DeleteTrustedContact(userID, id string) (map[string]bool, error)
	ListFamilyMembers(userID string) ([]model.FamilyMember, error)
	CreateFamilyMember(userID string, in dto.CreateFamilyMemberDTO) (model.FamilyMember, error)
	UpdateFamilyMember(userID, id string, in dto.UpdateFamilyMemberDTO) (model.FamilyMember, error)
	DeleteFamilyMember(userID, id string) (map[string]bool, error)
	ListTrustedDevices(userID string) ([]model.TrustedDevice, error)
	CreateTrustedDevice(userID string, in dto.CreateTrustedDeviceDTO) (model.TrustedDevice, error)
	UpdateTrustedDevice(userID, id string, in dto.UpdateTrustedDeviceDTO) (model.TrustedDevice, error)
	DeleteTrustedDevice(userID, id string) (map[string]bool, error)
	ListDocuments(userID string) ([]model.PersonalDocument, error)
	CreateDocument(userID string, in dto.CreatePersonalDocumentDTO) (model.PersonalDocument, error)
	UpdateDocument(userID, id string, in dto.UpdatePersonalDocumentDTO) (model.PersonalDocument, error)
	DeleteDocument(userID, id string) (map[string]bool, error)
	ListCalendars(userID string) ([]model.CalendarConnection, error)
	CreateCalendar(userID string, in dto.CreateCalendarConnectionDTO) (model.CalendarConnection, error)
	UpdateCalendar(userID, id string, in dto.UpdateCalendarConnectionDTO) (model.CalendarConnection, error)
	DeleteCalendar(userID, id string) (map[string]bool, error)
	DeleteAccount(userID string) (map[string]bool, error)
}

type MemoryRepository struct {
	store *storage.MemoryStore
	users usersrepo.Repository
}

func NewMemoryRepository(store *storage.MemoryStore, users usersrepo.Repository) *MemoryRepository {
	return &MemoryRepository{store: store, users: users}
}

func (r *MemoryRepository) GetOverview(userID string) (model.AccountOverview, error) {
	user, err := r.ensureUser(userID)
	if err != nil {
		return model.AccountOverview{}, err
	}

	settings, _ := r.GetSettings(userID)
	safety, _ := r.GetSafety(userID)
	privacy, _ := r.GetPrivacy(userID)
	security, _ := r.GetSecurity(userID)
	contacts, _ := r.ListTrustedContacts(userID)
	family, _ := r.ListFamilyMembers(userID)
	devices, _ := r.ListTrustedDevices(userID)
	documents, _ := r.ListDocuments(userID)
	calendars, _ := r.ListCalendars(userID)

	return model.AccountOverview{
		User:            user,
		Settings:        settings,
		Safety:          safety,
		Privacy:         privacy,
		Security:        security,
		TrustedContacts: contacts,
		FamilyMembers:   family,
		TrustedDevices:  devices,
		Documents:       documents,
		Calendars:       calendars,
	}, nil
}

func (r *MemoryRepository) GetSettings(userID string) (model.UserSettings, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.UserSettings{}, err
	}
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	if settings, ok := r.store.UserSettingsByUser[userID]; ok {
		return settings, nil
	}
	return defaultSettings(userID), nil
}

func (r *MemoryRepository) UpdateSettings(userID string, in dto.UpdateSettingsDTO) (model.UserSettings, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.UserSettings{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	settings := r.currentSettingsLocked(userID)
	if in.Language != nil && strings.TrimSpace(*in.Language) != "" {
		settings.Language = strings.TrimSpace(*in.Language)
	}
	if in.Timezone != nil && strings.TrimSpace(*in.Timezone) != "" {
		settings.Timezone = strings.TrimSpace(*in.Timezone)
	}
	if in.EmailNotifications != nil {
		settings.EmailNotifications = *in.EmailNotifications
	}
	if in.PushNotifications != nil {
		settings.PushNotifications = *in.PushNotifications
	}
	if in.SMSNotifications != nil {
		settings.SMSNotifications = *in.SMSNotifications
	}
	if in.TextSize != nil && strings.TrimSpace(*in.TextSize) != "" {
		settings.TextSize = strings.TrimSpace(*in.TextSize)
	}
	if in.AccessibilityVoiceOver != nil {
		settings.AccessibilityVoiceOver = *in.AccessibilityVoiceOver
	}
	if in.AccessibilityHighContrast != nil {
		settings.AccessibilityHighContrast = *in.AccessibilityHighContrast
	}
	if in.AccessibilityReduceMotion != nil {
		settings.AccessibilityReduceMotion = *in.AccessibilityReduceMotion
	}
	settings.UpdatedAt = time.Now()
	r.store.UserSettingsByUser[userID] = settings
	return settings, nil
}

func (r *MemoryRepository) GetSafety(userID string) (model.SafetyProfile, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.SafetyProfile{}, err
	}
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	if safety, ok := r.store.SafetyByUser[userID]; ok {
		return safety, nil
	}
	return defaultSafety(userID), nil
}

func (r *MemoryRepository) UpdateSafety(userID string, in dto.UpdateSafetyDTO) (model.SafetyProfile, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.SafetyProfile{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	safety := r.currentSafetyLocked(userID)
	if in.HomeCity != nil {
		safety.HomeCity = normalizeStringPtr(*in.HomeCity)
	}
	if in.HomeAddress != nil {
		safety.HomeAddress = normalizeStringPtr(*in.HomeAddress)
	}
	if in.EmergencyProtocol != nil {
		safety.EmergencyProtocol = normalizeStringPtr(*in.EmergencyProtocol)
	}
	if in.EmergencyNotes != nil {
		safety.EmergencyNotes = normalizeStringPtr(*in.EmergencyNotes)
	}
	if in.MedicalNotes != nil {
		safety.MedicalNotes = normalizeStringPtr(*in.MedicalNotes)
	}
	if in.ShareLiveLocation != nil {
		safety.ShareLiveLocation = *in.ShareLiveLocation
	}
	if in.SOSEnabled != nil {
		safety.SOSEnabled = *in.SOSEnabled
	}
	safety.UpdatedAt = time.Now()
	r.store.SafetyByUser[userID] = safety
	return safety, nil
}

func (r *MemoryRepository) GetPrivacy(userID string) (model.PrivacySettings, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.PrivacySettings{}, err
	}
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	if privacy, ok := r.store.PrivacyByUser[userID]; ok {
		return privacy, nil
	}
	return defaultPrivacy(userID), nil
}

func (r *MemoryRepository) UpdatePrivacy(userID string, in dto.UpdatePrivacyDTO) (model.PrivacySettings, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.PrivacySettings{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	privacy := r.currentPrivacyLocked(userID)
	if in.ProfileVisibility != nil && strings.TrimSpace(*in.ProfileVisibility) != "" {
		privacy.ProfileVisibility = strings.TrimSpace(*in.ProfileVisibility)
	}
	if in.LocationSharing != nil && strings.TrimSpace(*in.LocationSharing) != "" {
		privacy.LocationSharing = strings.TrimSpace(*in.LocationSharing)
	}
	if in.AllowMarketingEmails != nil {
		privacy.AllowMarketingEmails = *in.AllowMarketingEmails
	}
	if in.AllowAnalytics != nil {
		privacy.AllowAnalytics = *in.AllowAnalytics
	}
	if in.AllowMessagePreviews != nil {
		privacy.AllowMessagePreviews = *in.AllowMessagePreviews
	}
	privacy.UpdatedAt = time.Now()
	r.store.PrivacyByUser[userID] = privacy
	return privacy, nil
}

func (r *MemoryRepository) GetSecurity(userID string) (model.SecuritySettings, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.SecuritySettings{}, err
	}
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	if security, ok := r.store.SecurityByUser[userID]; ok {
		return security, nil
	}
	return defaultSecurity(userID), nil
}

func (r *MemoryRepository) UpdateSecurity(userID string, in dto.UpdateSecurityDTO) (model.SecuritySettings, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.SecuritySettings{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	security := r.currentSecurityLocked(userID)
	if in.TwoFactorEnabled != nil {
		security.TwoFactorEnabled = *in.TwoFactorEnabled
	}
	if in.TwoFactorMethod != nil {
		security.TwoFactorMethod = normalizeStringPtr(*in.TwoFactorMethod)
	}
	if in.SessionAlerts != nil {
		security.SessionAlerts = *in.SessionAlerts
	}
	if in.LoginNotifications != nil {
		security.LoginNotifications = *in.LoginNotifications
	}
	security.UpdatedAt = time.Now()
	r.store.SecurityByUser[userID] = security
	return security, nil
}

func (r *MemoryRepository) ListTrustedContacts(userID string) ([]model.TrustedContact, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return nil, err
	}
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	return collectOwned(r.store.TrustedContactByUser[userID], r.store.TrustedContacts, func(contact model.TrustedContact) time.Time {
		return contact.CreatedAt
	}), nil
}

func (r *MemoryRepository) CreateTrustedContact(userID string, in dto.CreateTrustedContactDTO) (model.TrustedContact, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.TrustedContact{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	now := time.Now()
	contact := model.TrustedContact{
		ID:                uuid.NewString(),
		UserID:            userID,
		Name:              strings.TrimSpace(in.Name),
		PhoneNumber:       strings.TrimSpace(in.PhoneNumber),
		Relationship:      strings.TrimSpace(in.Relationship),
		Email:             normalizeStringPtr(in.Email),
		ReceiveTripAlerts: boolOrDefault(in.ReceiveTripAlerts, true),
		IsPrimary:         boolOrDefault(in.IsPrimary, false),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if contact.IsPrimary {
		r.clearPrimaryContactLocked(userID)
	}
	r.store.TrustedContacts[contact.ID] = contact
	r.store.TrustedContactByUser[userID] = append(r.store.TrustedContactByUser[userID], contact.ID)
	return contact, nil
}

func (r *MemoryRepository) UpdateTrustedContact(userID, id string, in dto.UpdateTrustedContactDTO) (model.TrustedContact, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.TrustedContact{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	contact, ok := r.store.TrustedContacts[id]
	if !ok || contact.UserID != userID {
		return model.TrustedContact{}, apperr.New(404, "Record not found")
	}
	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		contact.Name = strings.TrimSpace(*in.Name)
	}
	if in.PhoneNumber != nil && strings.TrimSpace(*in.PhoneNumber) != "" {
		contact.PhoneNumber = strings.TrimSpace(*in.PhoneNumber)
	}
	if in.Relationship != nil && strings.TrimSpace(*in.Relationship) != "" {
		contact.Relationship = strings.TrimSpace(*in.Relationship)
	}
	if in.Email != nil {
		contact.Email = normalizeStringPtr(*in.Email)
	}
	if in.ReceiveTripAlerts != nil {
		contact.ReceiveTripAlerts = *in.ReceiveTripAlerts
	}
	if in.IsPrimary != nil {
		contact.IsPrimary = *in.IsPrimary
		if contact.IsPrimary {
			r.clearPrimaryContactLocked(userID)
			contact.IsPrimary = true
		}
	}
	contact.UpdatedAt = time.Now()
	r.store.TrustedContacts[id] = contact
	return contact, nil
}

func (r *MemoryRepository) DeleteTrustedContact(userID, id string) (map[string]bool, error) {
	return deleteOwned(&r.store.Mu, userID, id, r.store.TrustedContacts, r.store.TrustedContactByUser, func(contact model.TrustedContact) string {
		return contact.UserID
	})
}

func (r *MemoryRepository) ListFamilyMembers(userID string) ([]model.FamilyMember, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return nil, err
	}
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	return collectOwned(r.store.FamilyMemberByUser[userID], r.store.FamilyMembers, func(member model.FamilyMember) time.Time {
		return member.CreatedAt
	}), nil
}

func (r *MemoryRepository) CreateFamilyMember(userID string, in dto.CreateFamilyMemberDTO) (model.FamilyMember, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.FamilyMember{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	now := time.Now()
	member := model.FamilyMember{
		ID:           uuid.NewString(),
		UserID:       userID,
		Name:         strings.TrimSpace(in.Name),
		Relationship: strings.TrimSpace(in.Relationship),
		PhoneNumber:  normalizeStringPtr(in.PhoneNumber),
		Email:        normalizeStringPtr(in.Email),
		AccessLevel:  defaultString(strings.TrimSpace(in.AccessLevel), "view"),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	r.store.FamilyMembers[member.ID] = member
	r.store.FamilyMemberByUser[userID] = append(r.store.FamilyMemberByUser[userID], member.ID)
	return member, nil
}

func (r *MemoryRepository) UpdateFamilyMember(userID, id string, in dto.UpdateFamilyMemberDTO) (model.FamilyMember, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.FamilyMember{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	member, ok := r.store.FamilyMembers[id]
	if !ok || member.UserID != userID {
		return model.FamilyMember{}, apperr.New(404, "Record not found")
	}
	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		member.Name = strings.TrimSpace(*in.Name)
	}
	if in.Relationship != nil && strings.TrimSpace(*in.Relationship) != "" {
		member.Relationship = strings.TrimSpace(*in.Relationship)
	}
	if in.PhoneNumber != nil {
		member.PhoneNumber = normalizeStringPtr(*in.PhoneNumber)
	}
	if in.Email != nil {
		member.Email = normalizeStringPtr(*in.Email)
	}
	if in.AccessLevel != nil && strings.TrimSpace(*in.AccessLevel) != "" {
		member.AccessLevel = strings.TrimSpace(*in.AccessLevel)
	}
	member.UpdatedAt = time.Now()
	r.store.FamilyMembers[id] = member
	return member, nil
}

func (r *MemoryRepository) DeleteFamilyMember(userID, id string) (map[string]bool, error) {
	return deleteOwned(&r.store.Mu, userID, id, r.store.FamilyMembers, r.store.FamilyMemberByUser, func(member model.FamilyMember) string {
		return member.UserID
	})
}

func (r *MemoryRepository) ListTrustedDevices(userID string) ([]model.TrustedDevice, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return nil, err
	}
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	return collectOwned(r.store.TrustedDeviceByUser[userID], r.store.TrustedDevices, func(device model.TrustedDevice) time.Time {
		return device.CreatedAt
	}), nil
}

func (r *MemoryRepository) CreateTrustedDevice(userID string, in dto.CreateTrustedDeviceDTO) (model.TrustedDevice, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.TrustedDevice{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	now := time.Now()
	device := model.TrustedDevice{
		ID:         uuid.NewString(),
		UserID:     userID,
		DeviceName: strings.TrimSpace(in.DeviceName),
		Platform:   strings.TrimSpace(in.Platform),
		LastSeenAt: now,
		IsCurrent:  boolOrDefault(in.IsCurrent, false),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if device.IsCurrent {
		r.clearCurrentDeviceLocked(userID)
	}
	r.store.TrustedDevices[device.ID] = device
	r.store.TrustedDeviceByUser[userID] = append(r.store.TrustedDeviceByUser[userID], device.ID)
	return device, nil
}

func (r *MemoryRepository) UpdateTrustedDevice(userID, id string, in dto.UpdateTrustedDeviceDTO) (model.TrustedDevice, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.TrustedDevice{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	device, ok := r.store.TrustedDevices[id]
	if !ok || device.UserID != userID {
		return model.TrustedDevice{}, apperr.New(404, "Record not found")
	}
	if in.DeviceName != nil && strings.TrimSpace(*in.DeviceName) != "" {
		device.DeviceName = strings.TrimSpace(*in.DeviceName)
	}
	if in.Platform != nil && strings.TrimSpace(*in.Platform) != "" {
		device.Platform = strings.TrimSpace(*in.Platform)
	}
	if in.IsCurrent != nil {
		device.IsCurrent = *in.IsCurrent
		if device.IsCurrent {
			r.clearCurrentDeviceLocked(userID)
			device.IsCurrent = true
		}
	}
	if in.LastSeenAt != nil {
		device.LastSeenAt = *in.LastSeenAt
	} else {
		device.LastSeenAt = time.Now()
	}
	device.UpdatedAt = time.Now()
	r.store.TrustedDevices[id] = device
	return device, nil
}

func (r *MemoryRepository) DeleteTrustedDevice(userID, id string) (map[string]bool, error) {
	return deleteOwned(&r.store.Mu, userID, id, r.store.TrustedDevices, r.store.TrustedDeviceByUser, func(device model.TrustedDevice) string {
		return device.UserID
	})
}

func (r *MemoryRepository) ListDocuments(userID string) ([]model.PersonalDocument, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return nil, err
	}
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	return collectOwned(r.store.DocumentByUser[userID], r.store.Documents, func(document model.PersonalDocument) time.Time {
		return document.CreatedAt
	}), nil
}

func (r *MemoryRepository) CreateDocument(userID string, in dto.CreatePersonalDocumentDTO) (model.PersonalDocument, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.PersonalDocument{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	now := time.Now()
	document := model.PersonalDocument{
		ID:                 uuid.NewString(),
		UserID:             userID,
		DocumentType:       strings.TrimSpace(in.DocumentType),
		Label:              strings.TrimSpace(in.Label),
		DocumentNumber:     normalizeStringPtr(in.DocumentNumber),
		IssuingCountry:     normalizeStringPtr(in.IssuingCountry),
		VerificationStatus: defaultString(strings.TrimSpace(in.VerificationStatus), "PENDING"),
		FileURL:            normalizeStringPtr(in.FileURL),
		ExpiryDate:         in.ExpiryDate,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	r.store.Documents[document.ID] = document
	r.store.DocumentByUser[userID] = append(r.store.DocumentByUser[userID], document.ID)
	return document, nil
}

func (r *MemoryRepository) UpdateDocument(userID, id string, in dto.UpdatePersonalDocumentDTO) (model.PersonalDocument, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.PersonalDocument{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	document, ok := r.store.Documents[id]
	if !ok || document.UserID != userID {
		return model.PersonalDocument{}, apperr.New(404, "Record not found")
	}
	if in.DocumentType != nil && strings.TrimSpace(*in.DocumentType) != "" {
		document.DocumentType = strings.TrimSpace(*in.DocumentType)
	}
	if in.Label != nil && strings.TrimSpace(*in.Label) != "" {
		document.Label = strings.TrimSpace(*in.Label)
	}
	if in.DocumentNumber != nil {
		document.DocumentNumber = normalizeStringPtr(*in.DocumentNumber)
	}
	if in.IssuingCountry != nil {
		document.IssuingCountry = normalizeStringPtr(*in.IssuingCountry)
	}
	if in.VerificationStatus != nil && strings.TrimSpace(*in.VerificationStatus) != "" {
		document.VerificationStatus = strings.TrimSpace(*in.VerificationStatus)
	}
	if in.FileURL != nil {
		document.FileURL = normalizeStringPtr(*in.FileURL)
	}
	if in.ExpiryDate != nil {
		document.ExpiryDate = in.ExpiryDate
	}
	document.UpdatedAt = time.Now()
	r.store.Documents[id] = document
	return document, nil
}

func (r *MemoryRepository) DeleteDocument(userID, id string) (map[string]bool, error) {
	return deleteOwned(&r.store.Mu, userID, id, r.store.Documents, r.store.DocumentByUser, func(document model.PersonalDocument) string {
		return document.UserID
	})
}

func (r *MemoryRepository) ListCalendars(userID string) ([]model.CalendarConnection, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return nil, err
	}
	r.store.Mu.RLock()
	defer r.store.Mu.RUnlock()
	return collectOwned(r.store.CalendarByUser[userID], r.store.Calendars, func(calendar model.CalendarConnection) time.Time {
		return calendar.CreatedAt
	}), nil
}

func (r *MemoryRepository) CreateCalendar(userID string, in dto.CreateCalendarConnectionDTO) (model.CalendarConnection, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.CalendarConnection{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	now := time.Now()
	calendar := model.CalendarConnection{
		ID:                    uuid.NewString(),
		UserID:                userID,
		Provider:              strings.TrimSpace(in.Provider),
		CalendarName:          strings.TrimSpace(in.CalendarName),
		ExternalCalendarID:    normalizeStringPtr(in.ExternalCalendarID),
		SyncEnabled:           boolOrDefault(in.SyncEnabled, true),
		ReminderMinutesBefore: intOrDefault(in.ReminderMinutesBefore, 30),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	r.store.Calendars[calendar.ID] = calendar
	r.store.CalendarByUser[userID] = append(r.store.CalendarByUser[userID], calendar.ID)
	return calendar, nil
}

func (r *MemoryRepository) UpdateCalendar(userID, id string, in dto.UpdateCalendarConnectionDTO) (model.CalendarConnection, error) {
	if _, err := r.ensureUser(userID); err != nil {
		return model.CalendarConnection{}, err
	}
	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	calendar, ok := r.store.Calendars[id]
	if !ok || calendar.UserID != userID {
		return model.CalendarConnection{}, apperr.New(404, "Record not found")
	}
	if in.Provider != nil && strings.TrimSpace(*in.Provider) != "" {
		calendar.Provider = strings.TrimSpace(*in.Provider)
	}
	if in.CalendarName != nil && strings.TrimSpace(*in.CalendarName) != "" {
		calendar.CalendarName = strings.TrimSpace(*in.CalendarName)
	}
	if in.ExternalCalendarID != nil {
		calendar.ExternalCalendarID = normalizeStringPtr(*in.ExternalCalendarID)
	}
	if in.SyncEnabled != nil {
		calendar.SyncEnabled = *in.SyncEnabled
	}
	if in.ReminderMinutesBefore != nil {
		calendar.ReminderMinutesBefore = *in.ReminderMinutesBefore
	}
	if in.LastSyncedAt != nil {
		calendar.LastSyncedAt = in.LastSyncedAt
	}
	calendar.UpdatedAt = time.Now()
	r.store.Calendars[id] = calendar
	return calendar, nil
}

func (r *MemoryRepository) DeleteCalendar(userID, id string) (map[string]bool, error) {
	return deleteOwned(&r.store.Mu, userID, id, r.store.Calendars, r.store.CalendarByUser, func(calendar model.CalendarConnection) string {
		return calendar.UserID
	})
}

func (r *MemoryRepository) DeleteAccount(userID string) (map[string]bool, error) {
	user, err := r.ensureUser(userID)
	if err != nil {
		return nil, err
	}

	r.store.Mu.Lock()
	defer r.store.Mu.Unlock()

	delete(r.store.Users, userID)
	delete(r.store.UserByEmail, user.Email)
	delete(r.store.UserSettingsByUser, userID)
	delete(r.store.SafetyByUser, userID)
	delete(r.store.PrivacyByUser, userID)
	delete(r.store.SecurityByUser, userID)

	deleteOwnedSlice(r.store.TrustedContactByUser[userID], r.store.TrustedContacts)
	delete(r.store.TrustedContactByUser, userID)
	deleteOwnedSlice(r.store.FamilyMemberByUser[userID], r.store.FamilyMembers)
	delete(r.store.FamilyMemberByUser, userID)
	deleteOwnedSlice(r.store.TrustedDeviceByUser[userID], r.store.TrustedDevices)
	delete(r.store.TrustedDeviceByUser, userID)
	deleteOwnedSlice(r.store.DocumentByUser[userID], r.store.Documents)
	delete(r.store.DocumentByUser, userID)
	deleteOwnedSlice(r.store.CalendarByUser[userID], r.store.Calendars)
	delete(r.store.CalendarByUser, userID)

	for _, id := range r.store.PaymentMethodByUser[userID] {
		delete(r.store.PaymentMethods, id)
	}
	delete(r.store.PaymentMethodByUser, userID)
	for _, id := range r.store.NotificationByUser[userID] {
		delete(r.store.Notifications, id)
	}
	delete(r.store.NotificationByUser, userID)

	for carID, car := range r.store.Cars {
		if car.OwnerID == userID {
			delete(r.store.Cars, carID)
			delete(r.store.LocationsByCarID, carID)
		}
	}
	for rentalID, rental := range r.store.Rentals {
		if rental.UserID == userID {
			delete(r.store.Rentals, rentalID)
		}
	}

	return map[string]bool{"success": true}, nil
}

func (r *MemoryRepository) ensureUser(userID string) (model.User, error) {
	return r.users.FindByID(userID)
}

func (r *MemoryRepository) currentSettingsLocked(userID string) model.UserSettings {
	if settings, ok := r.store.UserSettingsByUser[userID]; ok {
		return settings
	}
	return defaultSettings(userID)
}

func (r *MemoryRepository) currentSafetyLocked(userID string) model.SafetyProfile {
	if safety, ok := r.store.SafetyByUser[userID]; ok {
		return safety
	}
	return defaultSafety(userID)
}

func (r *MemoryRepository) currentPrivacyLocked(userID string) model.PrivacySettings {
	if privacy, ok := r.store.PrivacyByUser[userID]; ok {
		return privacy
	}
	return defaultPrivacy(userID)
}

func (r *MemoryRepository) currentSecurityLocked(userID string) model.SecuritySettings {
	if security, ok := r.store.SecurityByUser[userID]; ok {
		return security
	}
	return defaultSecurity(userID)
}

func (r *MemoryRepository) clearPrimaryContactLocked(userID string) {
	for _, id := range r.store.TrustedContactByUser[userID] {
		contact := r.store.TrustedContacts[id]
		if contact.IsPrimary {
			contact.IsPrimary = false
			contact.UpdatedAt = time.Now()
			r.store.TrustedContacts[id] = contact
		}
	}
}

func (r *MemoryRepository) clearCurrentDeviceLocked(userID string) {
	for _, id := range r.store.TrustedDeviceByUser[userID] {
		device := r.store.TrustedDevices[id]
		if device.IsCurrent {
			device.IsCurrent = false
			device.UpdatedAt = time.Now()
			r.store.TrustedDevices[id] = device
		}
	}
}

func deleteOwned[T any](mu interface {
	Lock()
	Unlock()
}, userID, id string, values map[string]T, index map[string][]string, owner func(T) string) (map[string]bool, error) {
	mu.Lock()
	defer mu.Unlock()
	item, ok := values[id]
	if !ok || owner(item) != userID {
		return nil, apperr.New(404, "Record not found")
	}
	delete(values, id)
	index[userID] = filterIDs(index[userID], id)
	return map[string]bool{"success": true}, nil
}

func collectOwned[T any](ids []string, values map[string]T, createdAt func(T) time.Time) []T {
	out := make([]T, 0, len(ids))
	for _, id := range ids {
		if item, ok := values[id]; ok {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return createdAt(out[i]).Before(createdAt(out[j]))
	})
	return out
}

func deleteOwnedSlice[T any](ids []string, values map[string]T) {
	for _, id := range ids {
		delete(values, id)
	}
}

func filterIDs(ids []string, remove string) []string {
	out := ids[:0]
	for _, id := range ids {
		if id != remove {
			out = append(out, id)
		}
	}
	return out
}

func normalizeStringPtr(v string) *string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func boolOrDefault(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func intOrDefault(v *int, fallback int) int {
	if v == nil {
		return fallback
	}
	return *v
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func defaultSettings(userID string) model.UserSettings {
	return model.UserSettings{
		UserID:                    userID,
		Language:                  "en",
		Timezone:                  "UTC",
		EmailNotifications:        true,
		PushNotifications:         true,
		SMSNotifications:          false,
		TextSize:                  "medium",
		AccessibilityVoiceOver:    false,
		AccessibilityHighContrast: false,
		AccessibilityReduceMotion: false,
		UpdatedAt:                 time.Now(),
	}
}

func defaultSafety(userID string) model.SafetyProfile {
	return model.SafetyProfile{
		UserID:            userID,
		ShareLiveLocation: false,
		SOSEnabled:        true,
		UpdatedAt:         time.Now(),
	}
}

func defaultPrivacy(userID string) model.PrivacySettings {
	return model.PrivacySettings{
		UserID:               userID,
		ProfileVisibility:    "private",
		LocationSharing:      "whileInTrip",
		AllowMarketingEmails: false,
		AllowAnalytics:       true,
		AllowMessagePreviews: true,
		UpdatedAt:            time.Now(),
	}
}

func defaultSecurity(userID string) model.SecuritySettings {
	return model.SecuritySettings{
		UserID:             userID,
		TwoFactorEnabled:   false,
		SessionAlerts:      true,
		LoginNotifications: true,
		UpdatedAt:          time.Now(),
	}
}
