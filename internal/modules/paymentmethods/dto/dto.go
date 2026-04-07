package dto

type CreateCardPaymentMethodDTO struct {
	CardNumber     string `json:"cardNumber" binding:"required"`
	CardType       string `json:"cardType" binding:"required"`
	ExpiryMonth    int    `json:"expiryMonth" binding:"required"`
	ExpiryYear     int    `json:"expiryYear" binding:"required"`
	CardholderName string `json:"cardholderName" binding:"required"`
	IsDefault      bool   `json:"isDefault"`
}

type CreateMobilePaymentMethodDTO struct {
	MobileNumber string `json:"mobileNumber" binding:"required"`
	Provider     string `json:"provider" binding:"required"`
	AccountName  string `json:"accountName" binding:"required"`
	IsDefault    bool   `json:"isDefault"`
}

type UpdatePaymentMethodDTO struct {
	IsDefault *bool   `json:"isDefault"`
	Status    *string `json:"status"`
}
