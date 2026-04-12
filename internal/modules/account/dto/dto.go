package dto

import "time"

type UpdateSettingsDTO struct {
	Language                  *string `json:"language"`
	Timezone                  *string `json:"timezone"`
	EmailNotifications        *bool   `json:"emailNotifications"`
	PushNotifications         *bool   `json:"pushNotifications"`
	SMSNotifications          *bool   `json:"smsNotifications"`
	TextSize                  *string `json:"textSize"`
	AccessibilityVoiceOver    *bool   `json:"accessibilityVoiceOver"`
	AccessibilityHighContrast *bool   `json:"accessibilityHighContrast"`
	AccessibilityReduceMotion *bool   `json:"accessibilityReduceMotion"`
}

type UpdateSafetyDTO struct {
	HomeCity          *string `json:"homeCity"`
	HomeAddress       *string `json:"homeAddress"`
	EmergencyProtocol *string `json:"emergencyProtocol"`
	EmergencyNotes    *string `json:"emergencyNotes"`
	MedicalNotes      *string `json:"medicalNotes"`
	ShareLiveLocation *bool   `json:"shareLiveLocation"`
	SOSEnabled        *bool   `json:"sosEnabled"`
}

type UpdatePrivacyDTO struct {
	ProfileVisibility    *string `json:"profileVisibility"`
	LocationSharing      *string `json:"locationSharing"`
	AllowMarketingEmails *bool   `json:"allowMarketingEmails"`
	AllowAnalytics       *bool   `json:"allowAnalytics"`
	AllowMessagePreviews *bool   `json:"allowMessagePreviews"`
}

type UpdateSecurityDTO struct {
	TwoFactorEnabled   *bool   `json:"twoFactorEnabled"`
	TwoFactorMethod    *string `json:"twoFactorMethod"`
	SessionAlerts      *bool   `json:"sessionAlerts"`
	LoginNotifications *bool   `json:"loginNotifications"`
}

type CreateTrustedContactDTO struct {
	Name              string `json:"name" binding:"required"`
	PhoneNumber       string `json:"phoneNumber" binding:"required"`
	Relationship      string `json:"relationship" binding:"required"`
	Email             string `json:"email"`
	ReceiveTripAlerts *bool  `json:"receiveTripAlerts"`
	IsPrimary         *bool  `json:"isPrimary"`
}

type UpdateTrustedContactDTO struct {
	Name              *string `json:"name"`
	PhoneNumber       *string `json:"phoneNumber"`
	Relationship      *string `json:"relationship"`
	Email             *string `json:"email"`
	ReceiveTripAlerts *bool   `json:"receiveTripAlerts"`
	IsPrimary         *bool   `json:"isPrimary"`
}

type CreateFamilyMemberDTO struct {
	Name         string `json:"name" binding:"required"`
	Relationship string `json:"relationship" binding:"required"`
	PhoneNumber  string `json:"phoneNumber"`
	Email        string `json:"email"`
	AccessLevel  string `json:"accessLevel"`
}

type UpdateFamilyMemberDTO struct {
	Name         *string `json:"name"`
	Relationship *string `json:"relationship"`
	PhoneNumber  *string `json:"phoneNumber"`
	Email        *string `json:"email"`
	AccessLevel  *string `json:"accessLevel"`
}

type CreateTrustedDeviceDTO struct {
	DeviceName string `json:"deviceName" binding:"required"`
	Platform   string `json:"platform" binding:"required"`
	IsCurrent  *bool  `json:"isCurrent"`
}

type UpdateTrustedDeviceDTO struct {
	DeviceName *string    `json:"deviceName"`
	Platform   *string    `json:"platform"`
	IsCurrent  *bool      `json:"isCurrent"`
	LastSeenAt *time.Time `json:"lastSeenAt"`
}

type CreatePersonalDocumentDTO struct {
	DocumentType       string     `json:"documentType" binding:"required"`
	Label              string     `json:"label" binding:"required"`
	DocumentNumber     string     `json:"documentNumber"`
	IssuingCountry     string     `json:"issuingCountry"`
	VerificationStatus string     `json:"verificationStatus"`
	FileURL            string     `json:"fileUrl"`
	ExpiryDate         *time.Time `json:"expiryDate"`
}

type UpdatePersonalDocumentDTO struct {
	DocumentType       *string    `json:"documentType"`
	Label              *string    `json:"label"`
	DocumentNumber     *string    `json:"documentNumber"`
	IssuingCountry     *string    `json:"issuingCountry"`
	VerificationStatus *string    `json:"verificationStatus"`
	FileURL            *string    `json:"fileUrl"`
	ExpiryDate         *time.Time `json:"expiryDate"`
}

type CreateCalendarConnectionDTO struct {
	Provider              string `json:"provider" binding:"required"`
	CalendarName          string `json:"calendarName" binding:"required"`
	ExternalCalendarID    string `json:"externalCalendarId"`
	SyncEnabled           *bool  `json:"syncEnabled"`
	ReminderMinutesBefore *int   `json:"reminderMinutesBefore"`
}

type UpdateCalendarConnectionDTO struct {
	Provider              *string    `json:"provider"`
	CalendarName          *string    `json:"calendarName"`
	ExternalCalendarID    *string    `json:"externalCalendarId"`
	SyncEnabled           *bool      `json:"syncEnabled"`
	ReminderMinutesBefore *int       `json:"reminderMinutesBefore"`
	LastSyncedAt          *time.Time `json:"lastSyncedAt"`
}
