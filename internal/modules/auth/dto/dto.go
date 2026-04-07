package dto

type LoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterDTO struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	Name           string `json:"name" binding:"required"`
	PhoneNumber    string `json:"phoneNumber"`
	Role           string `json:"role"`
	ProfilePicture string `json:"profilePicture"`
}

type RefreshDTO struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type CheckEmailDTO struct {
	Email string `json:"email" binding:"required,email"`
}

type ValidateEmailDTO struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type CompleteRegistrationDTO struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	Name        string `json:"name" binding:"required"`
	PhoneNumber string `json:"phoneNumber"`
	Role        string `json:"role"`
}

type UpdateProfileDTO struct {
	Name           *string `json:"name"`
	Email          *string `json:"email"`
	PhoneNumber    *string `json:"phoneNumber"`
	ProfilePicture *string `json:"profilePicture"`
}

type ChangePasswordDTO struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

type GoogleOAuthDTO struct {
	IDToken        string `json:"idToken" binding:"required"`
	AccessToken    string `json:"accessToken" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Name           string `json:"name" binding:"required"`
	ProfilePicture string `json:"profilePicture"`
}

type AppleOAuthDTO struct {
	IdentityToken     string `json:"identityToken" binding:"required"`
	AuthorizationCode string `json:"authorizationCode" binding:"required"`
	Email             string `json:"email"`
	UserIdentifier    string `json:"userIdentifier" binding:"required"`
	FullName          struct {
		GivenName  string `json:"givenName"`
		FamilyName string `json:"familyName"`
	} `json:"fullName"`
}
