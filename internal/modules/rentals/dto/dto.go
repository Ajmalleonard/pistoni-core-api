package dto

type CreateRentalDTO struct {
	StartDate         string   `json:"startDate" binding:"required"`
	EndDate           string   `json:"endDate" binding:"required"`
	TotalPrice        float64  `json:"totalPrice" binding:"required"`
	Status            string   `json:"status"`
	DistanceKM        *float64 `json:"distanceKm"`
	UserID            string   `json:"userId" binding:"required"`
	CarID             string   `json:"carId" binding:"required"`
	PickupLocationID  string   `json:"pickupLocationId" binding:"required"`
	DropoffLocationID string   `json:"dropoffLocationId" binding:"required"`
	DriverRequired    *bool    `json:"driverRequired"`
	DriverID          *string  `json:"driverId"`
	SecurityRequired  *bool    `json:"securityRequired"`
	SecurityNotes     *string  `json:"securityNotes"`
	BodyGuardIDs      []string `json:"bodyGuardIds"`
}

type RouteCarSelectionDTO struct {
	CarID    string  `json:"carId" binding:"required"`
	DriverID *string `json:"driverId"`
}

type CreateRouteRentalsDTO struct {
	StartDate         string                 `json:"startDate" binding:"required"`
	EndDate           string                 `json:"endDate" binding:"required"`
	TotalPrice        float64                `json:"totalPrice" binding:"required"`
	Status            string                 `json:"status"`
	DistanceKM        *float64               `json:"distanceKm"`
	UserID            string                 `json:"userId" binding:"required"`
	PickupLocationID  string                 `json:"pickupLocationId" binding:"required"`
	DropoffLocationID string                 `json:"dropoffLocationId" binding:"required"`
	DriverRequired    *bool                  `json:"driverRequired"`
	SecurityRequired  *bool                  `json:"securityRequired"`
	SecurityNotes     *string                `json:"securityNotes"`
	BodyGuardIDs      []string               `json:"bodyGuardIds"`
	Cars              []RouteCarSelectionDTO `json:"cars" binding:"required"`
}

type CalculateRentalPriceDTO struct {
	StartDate    string   `json:"startDate" binding:"required"`
	EndDate      string   `json:"endDate" binding:"required"`
	CarID        string   `json:"carId" binding:"required"`
	DistanceKM   float64  `json:"distanceKm" binding:"required"`
	DriverID     *string  `json:"driverId"`
	BodyGuardIDs []string `json:"bodyGuardIds"`
}

type CalculateRoutePriceDTO struct {
	StartDate    string                 `json:"startDate" binding:"required"`
	EndDate      string                 `json:"endDate" binding:"required"`
	DistanceKM   float64                `json:"distanceKm" binding:"required"`
	Cars         []RouteCarSelectionDTO `json:"cars" binding:"required"`
	BodyGuardIDs []string               `json:"bodyGuardIds"`
}
