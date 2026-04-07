package dto

type FindProviderQuery struct {
	Skip               int    `form:"skip"`
	Take               int    `form:"take"`
	IsAvailable        *bool  `form:"isAvailable"`
	VerificationStatus string `form:"verificationStatus"`
	Gender             string `form:"gender"`
}

type DriverStatusPollQuery struct {
	SinceVersion int64 `form:"sinceVersion"`
	SinceUnixMs  int64 `form:"sinceUnixMs"`
	Take         int   `form:"take"`
}

type RegisterDriverDTO struct {
	Gender                string  `json:"gender"`
	DateOfBirth           string  `json:"dateOfBirth"`
	Address               string  `json:"address"`
	NIN                   string  `json:"nin"`
	DriversLicenseNumber  string  `json:"driversLicenseNumber" binding:"required"`
	DriversLicenseClass   string  `json:"driversLicenseClass"`
	DriversLicenseExpiry  string  `json:"driversLicenseExpiry"`
	IssuingAuthority      string  `json:"issuingAuthority"`
	YearsOfExperience     *int    `json:"yearsOfExperience"`
	EmergencyContactName  string  `json:"emergencyContactName"`
	EmergencyContactPhone string  `json:"emergencyContactPhone"`
	DailyRate             float64 `json:"dailyRate" binding:"required"`
	Notes                 string  `json:"notes"`
}

type CreateDriverReviewDTO struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

type UpdateDriverStatusDTO struct {
	Status      string `json:"status" binding:"required"`
	IsAvailable *bool  `json:"isAvailable"`
}

type DriverAssignmentDeltaDTO struct {
	RouteDelta      int `json:"routeDelta"`
	AssignmentDelta int `json:"assignmentDelta"`
}

type RegisterBodyGuardDTO struct {
	Gender                string  `json:"gender" binding:"required"`
	DateOfBirth           string  `json:"dateOfBirth"`
	Address               string  `json:"address"`
	NIN                   string  `json:"nin" binding:"required"`
	GovernmentIDNumber    string  `json:"governmentIdNumber" binding:"required"`
	SecurityLicenseNumber string  `json:"securityLicenseNumber"`
	SecurityLicenseExpiry string  `json:"securityLicenseExpiry"`
	IssuingAuthority      string  `json:"issuingAuthority"`
	YearsOfExperience     *int    `json:"yearsOfExperience"`
	EmergencyContactName  string  `json:"emergencyContactName"`
	EmergencyContactPhone string  `json:"emergencyContactPhone"`
	DailyRate             float64 `json:"dailyRate" binding:"required"`
	Notes                 string  `json:"notes"`
}

type JoinBodyGuardQueueDTO struct {
	Gender             string  `json:"gender" binding:"required"`
	NIN                string  `json:"nin" binding:"required"`
	GovernmentIDNumber string  `json:"governmentIdNumber" binding:"required"`
	DailyRate          float64 `json:"dailyRate" binding:"required"`
}

type ReviewBodyGuardQueueDTO struct {
	Reason string `json:"reason"`
}

type UpdateBodyGuardStatusDTO struct {
	Status      string `json:"status" binding:"required"`
	IsAvailable *bool  `json:"isAvailable"`
}

type UpdateAvailabilityDTO struct {
	IsAvailable bool `json:"isAvailable"`
}

type UpdateVerificationDTO struct {
	VerificationStatus       string `json:"verificationStatus" binding:"required"`
	BackgroundCheckCompleted *bool  `json:"backgroundCheckCompleted"`
	IsAvailable              *bool  `json:"isAvailable"`
	Notes                    string `json:"notes"`
}
