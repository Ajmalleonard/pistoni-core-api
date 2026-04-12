package model

import "time"

type UserRole string

type AuthProvider string

type CarStatus string

type VehicleType string

type Gender string

type ProviderVerificationStatus string

type DriverStatus string

type BodyGuardStatus string

type BodyGuardQueueStatus string

type RentalStatus string

type NotificationType string

type PaymentMethodType string

type PaymentMethodStatus string

const (
	UserRoleAdmin     UserRole = "ADMIN"
	UserRoleCarOwner  UserRole = "CAR_OWNER"
	UserRoleDriver    UserRole = "DRIVER"
	UserRoleBodyGuard UserRole = "BODY_GUARD"
	UserRoleClient    UserRole = "CLIENT"
)

const (
	AuthProviderEmail  AuthProvider = "EMAIL"
	AuthProviderGoogle AuthProvider = "GOOGLE"
	AuthProviderApple  AuthProvider = "APPLE"
)

const (
	CarStatusPendingReview CarStatus = "PENDING_REVIEW"
	CarStatusApproved      CarStatus = "APPROVED"
	CarStatusRejected      CarStatus = "REJECTED"
	CarStatusMaintenance   CarStatus = "MAINTENANCE"
	CarStatusAvailable     CarStatus = "AVAILABLE"
	CarStatusRented        CarStatus = "RENTED"
)

const (
	VehicleTypeSUV         VehicleType = "SUV"
	VehicleTypeCoaster     VehicleType = "COASTER"
	VehicleTypeLongVehicle VehicleType = "LONG_VEHICLE"
)

const (
	RentalStatusPending   RentalStatus = "PENDING"
	RentalStatusConfirmed RentalStatus = "CONFIRMED"
	RentalStatusActive    RentalStatus = "ACTIVE"
	RentalStatusCompleted RentalStatus = "COMPLETED"
	RentalStatusCancelled RentalStatus = "CANCELLED"
)

const (
	NotificationTypeChat      NotificationType = "CHAT"
	NotificationTypeRental    NotificationType = "RENTAL"
	NotificationTypePayment   NotificationType = "PAYMENT"
	NotificationTypeSystem    NotificationType = "SYSTEM"
	NotificationTypePromotion NotificationType = "PROMOTION"
)

const (
	PaymentMethodTypeCard        PaymentMethodType = "CARD"
	PaymentMethodTypeMobileMoney PaymentMethodType = "MOBILE_MONEY"
)

const (
	PaymentMethodStatusActive   PaymentMethodStatus = "ACTIVE"
	PaymentMethodStatusInactive PaymentMethodStatus = "INACTIVE"
	PaymentMethodStatusDefault  PaymentMethodStatus = "DEFAULT"
)

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
)

const (
	VerificationPending   ProviderVerificationStatus = "PENDING_REVIEW"
	VerificationApproved  ProviderVerificationStatus = "APPROVED"
	VerificationRejected  ProviderVerificationStatus = "REJECTED"
	VerificationSuspended ProviderVerificationStatus = "SUSPENDED"
)

const (
	DriverStatusOffline     DriverStatus = "OFFLINE"
	DriverStatusOnline      DriverStatus = "ONLINE"
	DriverStatusBusy        DriverStatus = "BUSY"
	DriverStatusOnTrip      DriverStatus = "ON_TRIP"
	DriverStatusUnavailable DriverStatus = "UNAVAILABLE"
)

const (
	BodyGuardStatusOffline     BodyGuardStatus = "OFFLINE"
	BodyGuardStatusOnline      BodyGuardStatus = "ONLINE"
	BodyGuardStatusAssigned    BodyGuardStatus = "ASSIGNED"
	BodyGuardStatusUnavailable BodyGuardStatus = "UNAVAILABLE"
)

const (
	BodyGuardQueuePending  BodyGuardQueueStatus = "PENDING"
	BodyGuardQueueApproved BodyGuardQueueStatus = "APPROVED"
	BodyGuardQueueRejected BodyGuardQueueStatus = "REJECTED"
)

type User struct {
	ID                 string       `json:"id"`
	Email              string       `json:"email"`
	Password           *string      `json:"-"`
	Name               string       `json:"name"`
	PhoneNumber        *string      `json:"phoneNumber"`
	Role               UserRole     `json:"role"`
	ProfilePicture     *string      `json:"profilePicture"`
	IsVerified         bool         `json:"isVerified"`
	AuthProvider       AuthProvider `json:"authProvider"`
	VerificationToken  *string      `json:"-"`
	TokenExpiry        *time.Time   `json:"-"`
	VerificationCode   *string      `json:"-"`
	CodeType           *string      `json:"-"`
	CodeExpiry         *time.Time   `json:"-"`
	RefreshToken       *string      `json:"-"`
	RefreshTokenExpiry *time.Time   `json:"-"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
}

type PendingEmailVerification struct {
	Email      string
	Code       string
	CodeExpiry time.Time
	CreatedAt  time.Time
}

type Location struct {
	ID        string    `json:"id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	State     *string   `json:"state"`
	Country   string    `json:"country"`
	ZipCode   *string   `json:"zipCode"`
	CarID     string    `json:"carId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Car struct {
	ID               string      `json:"id"`
	Make             string      `json:"make"`
	Model            string      `json:"model"`
	Year             int         `json:"year"`
	Color            string      `json:"color"`
	LicensePlate     string      `json:"licensePlate"`
	Price            float64     `json:"price"`
	Images           []string    `json:"images"`
	Description      *string     `json:"description"`
	Status           CarStatus   `json:"status"`
	Features         []string    `json:"features"`
	SeatCount        int         `json:"seatCount"`
	TransmissionType string      `json:"transmissionType"`
	FuelType         string      `json:"fuelType"`
	VehicleType      VehicleType `json:"vehicleType"`
	OwnerID          string      `json:"ownerId"`
	CreatedAt        time.Time   `json:"createdAt"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

type CarView struct {
	Car
	Owner    *PublicUser `json:"owner,omitempty"`
	Location *Location   `json:"location,omitempty"`
	Reviews  any         `json:"reviews,omitempty"`
}

type PublicUser struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	PhoneNumber    *string `json:"phoneNumber"`
	ProfilePicture *string `json:"profilePicture"`
}

type PrivateUser struct {
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	Name           string   `json:"name"`
	PhoneNumber    *string  `json:"phoneNumber"`
	ProfilePicture *string  `json:"profilePicture"`
	Role           UserRole `json:"role"`
	IsVerified     bool     `json:"isVerified"`
}

type UserSettings struct {
	UserID                    string    `json:"userId"`
	Language                  string    `json:"language"`
	Timezone                  string    `json:"timezone"`
	EmailNotifications        bool      `json:"emailNotifications"`
	PushNotifications         bool      `json:"pushNotifications"`
	SMSNotifications          bool      `json:"smsNotifications"`
	TextSize                  string    `json:"textSize"`
	AccessibilityVoiceOver    bool      `json:"accessibilityVoiceOver"`
	AccessibilityHighContrast bool      `json:"accessibilityHighContrast"`
	AccessibilityReduceMotion bool      `json:"accessibilityReduceMotion"`
	UpdatedAt                 time.Time `json:"updatedAt"`
}

type SafetyProfile struct {
	UserID            string    `json:"userId"`
	HomeCity          *string   `json:"homeCity"`
	HomeAddress       *string   `json:"homeAddress"`
	EmergencyProtocol *string   `json:"emergencyProtocol"`
	EmergencyNotes    *string   `json:"emergencyNotes"`
	MedicalNotes      *string   `json:"medicalNotes"`
	ShareLiveLocation bool      `json:"shareLiveLocation"`
	SOSEnabled        bool      `json:"sosEnabled"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type PrivacySettings struct {
	UserID               string    `json:"userId"`
	ProfileVisibility    string    `json:"profileVisibility"`
	LocationSharing      string    `json:"locationSharing"`
	AllowMarketingEmails bool      `json:"allowMarketingEmails"`
	AllowAnalytics       bool      `json:"allowAnalytics"`
	AllowMessagePreviews bool      `json:"allowMessagePreviews"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type SecuritySettings struct {
	UserID             string    `json:"userId"`
	TwoFactorEnabled   bool      `json:"twoFactorEnabled"`
	TwoFactorMethod    *string   `json:"twoFactorMethod"`
	SessionAlerts      bool      `json:"sessionAlerts"`
	LoginNotifications bool      `json:"loginNotifications"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type TrustedContact struct {
	ID                string    `json:"id"`
	UserID            string    `json:"userId"`
	Name              string    `json:"name"`
	PhoneNumber       string    `json:"phoneNumber"`
	Relationship      string    `json:"relationship"`
	Email             *string   `json:"email"`
	ReceiveTripAlerts bool      `json:"receiveTripAlerts"`
	IsPrimary         bool      `json:"isPrimary"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type FamilyMember struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Name         string    `json:"name"`
	Relationship string    `json:"relationship"`
	PhoneNumber  *string   `json:"phoneNumber"`
	Email        *string   `json:"email"`
	AccessLevel  string    `json:"accessLevel"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type TrustedDevice struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	DeviceName string    `json:"deviceName"`
	Platform   string    `json:"platform"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	IsCurrent  bool      `json:"isCurrent"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type PersonalDocument struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"userId"`
	DocumentType       string     `json:"documentType"`
	Label              string     `json:"label"`
	DocumentNumber     *string    `json:"documentNumber"`
	IssuingCountry     *string    `json:"issuingCountry"`
	VerificationStatus string     `json:"verificationStatus"`
	FileURL            *string    `json:"fileUrl"`
	ExpiryDate         *time.Time `json:"expiryDate"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type CalendarConnection struct {
	ID                    string     `json:"id"`
	UserID                string     `json:"userId"`
	Provider              string     `json:"provider"`
	CalendarName          string     `json:"calendarName"`
	ExternalCalendarID    *string    `json:"externalCalendarId"`
	SyncEnabled           bool       `json:"syncEnabled"`
	ReminderMinutesBefore int        `json:"reminderMinutesBefore"`
	LastSyncedAt          *time.Time `json:"lastSyncedAt"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

type AccountOverview struct {
	User            User                 `json:"user"`
	Settings        UserSettings         `json:"settings"`
	Safety          SafetyProfile        `json:"safety"`
	Privacy         PrivacySettings      `json:"privacy"`
	Security        SecuritySettings     `json:"security"`
	TrustedContacts []TrustedContact     `json:"trustedContacts"`
	FamilyMembers   []FamilyMember       `json:"familyMembers"`
	TrustedDevices  []TrustedDevice      `json:"trustedDevices"`
	Documents       []PersonalDocument   `json:"documents"`
	Calendars       []CalendarConnection `json:"calendars"`
}

type DriverProfile struct {
	ID                       string                     `json:"id"`
	UserID                   string                     `json:"userId"`
	PistoniIDNumber          string                     `json:"pistoniIdNumber"`
	Gender                   *Gender                    `json:"gender"`
	DateOfBirth              *time.Time                 `json:"dateOfBirth"`
	Address                  *string                    `json:"address"`
	NIN                      *string                    `json:"nin"`
	DriversLicenseNumber     string                     `json:"driversLicenseNumber"`
	DriversLicenseClass      *string                    `json:"driversLicenseClass"`
	DriversLicenseExpiry     *time.Time                 `json:"driversLicenseExpiry"`
	IssuingAuthority         *string                    `json:"issuingAuthority"`
	YearsOfExperience        *int                       `json:"yearsOfExperience"`
	EmergencyContactName     *string                    `json:"emergencyContactName"`
	EmergencyContactPhone    *string                    `json:"emergencyContactPhone"`
	DailyRate                float64                    `json:"dailyRate"`
	VerificationStatus       ProviderVerificationStatus `json:"verificationStatus"`
	IsAvailable              bool                       `json:"isAvailable"`
	Status                   DriverStatus               `json:"status"`
	StatusVersion            int64                      `json:"statusVersion"`
	StatusUpdatedAt          time.Time                  `json:"statusUpdatedAt"`
	LastSeen                 *time.Time                 `json:"lastSeen"`
	RoutesCount              int                        `json:"routesCount"`
	AssignmentCount          int                        `json:"assignmentCount"`
	BackgroundCheckCompleted bool                       `json:"backgroundCheckCompleted"`
	Notes                    *string                    `json:"notes"`
	CreatedAt                time.Time                  `json:"createdAt"`
	UpdatedAt                time.Time                  `json:"updatedAt"`
}

type BodyGuardProfile struct {
	ID                       string                     `json:"id"`
	UserID                   string                     `json:"userId"`
	PistoniIDNumber          string                     `json:"pistoniIdNumber"`
	Gender                   Gender                     `json:"gender"`
	DateOfBirth              *time.Time                 `json:"dateOfBirth"`
	Address                  *string                    `json:"address"`
	NIN                      string                     `json:"nin"`
	GovernmentIDNumber       string                     `json:"governmentIdNumber"`
	SecurityLicenseNumber    *string                    `json:"securityLicenseNumber"`
	SecurityLicenseExpiry    *time.Time                 `json:"securityLicenseExpiry"`
	IssuingAuthority         *string                    `json:"issuingAuthority"`
	YearsOfExperience        *int                       `json:"yearsOfExperience"`
	EmergencyContactName     *string                    `json:"emergencyContactName"`
	EmergencyContactPhone    *string                    `json:"emergencyContactPhone"`
	DailyRate                float64                    `json:"dailyRate"`
	VerificationStatus       ProviderVerificationStatus `json:"verificationStatus"`
	IsAvailable              bool                       `json:"isAvailable"`
	Status                   BodyGuardStatus            `json:"status"`
	StatusVersion            int64                      `json:"statusVersion"`
	StatusUpdatedAt          time.Time                  `json:"statusUpdatedAt"`
	LastSeen                 *time.Time                 `json:"lastSeen"`
	BackgroundCheckCompleted bool                       `json:"backgroundCheckCompleted"`
	Notes                    *string                    `json:"notes"`
	CreatedAt                time.Time                  `json:"createdAt"`
	UpdatedAt                time.Time                  `json:"updatedAt"`
}

type DriverReview struct {
	ID        string    `json:"id"`
	DriverID  string    `json:"driverId"`
	AuthorID  string    `json:"authorId"`
	Rating    int       `json:"rating"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

type BodyGuardRegistrationQueueItem struct {
	ID                 string               `json:"id"`
	UserID             string               `json:"userId"`
	Gender             Gender               `json:"gender"`
	NIN                string               `json:"nin"`
	GovernmentIDNumber string               `json:"governmentIdNumber"`
	DailyRate          float64              `json:"dailyRate"`
	Status             BodyGuardQueueStatus `json:"status"`
	Reason             *string              `json:"reason"`
	SubmittedAt        time.Time            `json:"submittedAt"`
	ReviewedAt         *time.Time           `json:"reviewedAt"`
	ReviewedBy         *string              `json:"reviewedBy"`
}

type PaymentMethod struct {
	ID             string              `json:"id"`
	UserID         string              `json:"userId"`
	Type           PaymentMethodType   `json:"type"`
	Status         PaymentMethodStatus `json:"status"`
	IsDefault      bool                `json:"isDefault"`
	CardNumber     *string             `json:"cardNumber"`
	CardType       *string             `json:"cardType"`
	ExpiryMonth    *int                `json:"expiryMonth"`
	ExpiryYear     *int                `json:"expiryYear"`
	CardholderName *string             `json:"cardholderName"`
	MobileNumber   *string             `json:"mobileNumber"`
	Provider       *string             `json:"provider"`
	AccountName    *string             `json:"accountName"`
	EncryptedData  *string             `json:"encryptedData"`
	Token          *string             `json:"token"`
	CreatedAt      time.Time           `json:"createdAt"`
	UpdatedAt      time.Time           `json:"updatedAt"`
}

type Rental struct {
	ID                string       `json:"id"`
	StartDate         time.Time    `json:"startDate"`
	EndDate           time.Time    `json:"endDate"`
	TotalPrice        float64      `json:"totalPrice"`
	Status            RentalStatus `json:"status"`
	DistanceKM        *float64     `json:"distanceKm"`
	DriverRequired    bool         `json:"driverRequired"`
	SecurityRequired  bool         `json:"securityRequired"`
	SecurityNotes     *string      `json:"securityNotes"`
	UserID            string       `json:"userId"`
	CarID             string       `json:"carId"`
	PickupLocationID  string       `json:"pickupLocationId"`
	DropoffLocationID string       `json:"dropoffLocationId"`
	DriverID          *string      `json:"driverId"`
	BodyGuardIDs      []string     `json:"bodyGuardIds"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
}

type Notification struct {
	ID              string           `json:"id"`
	UserID          string           `json:"userId"`
	Title           string           `json:"title"`
	Body            string           `json:"body"`
	Type            NotificationType `json:"type"`
	RelatedEntityID *string          `json:"relatedEntityId"`
	IsRead          bool             `json:"isRead"`
	ImageURL        *string          `json:"imageUrl"`
	ActionURL       *string          `json:"actionUrl"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}

type ChatRoom struct {
	ID             string    `json:"id"`
	Name           *string   `json:"name"`
	Type           string    `json:"type"`
	ParticipantIDs []string  `json:"participantIds"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	LastMessageID  *string   `json:"lastMessageId"`
}

type Message struct {
	ID             string    `json:"id"`
	ChatRoomID     string    `json:"chatRoomId"`
	SenderID       string    `json:"senderId"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
	IsRead         bool      `json:"isRead"`
	IsDeleted      bool      `json:"isDeleted"`
	Edited         bool      `json:"edited"`
	TargetType     *string   `json:"targetType,omitempty"`
	TargetIDs      []string  `json:"targetIds,omitempty"`
	TargetLabel    *string   `json:"targetLabel,omitempty"`
	ReportCategory *string   `json:"reportCategory,omitempty"`
}
