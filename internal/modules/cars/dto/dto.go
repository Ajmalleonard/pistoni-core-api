package dto

type LocationCreate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
	City      string  `json:"city"`
	State     string  `json:"state"`
	Country   string  `json:"country"`
	ZipCode   string  `json:"zipCode"`
}

type LocationWrapper struct {
	Create *LocationCreate `json:"create"`
}

type CreateCarDTO struct {
	Make             string           `json:"make" binding:"required"`
	Model            string           `json:"model" binding:"required"`
	Year             int              `json:"year" binding:"required"`
	Color            string           `json:"color" binding:"required"`
	LicensePlate     string           `json:"licensePlate" binding:"required"`
	Price            float64          `json:"price" binding:"required"`
	Images           []string         `json:"images"`
	Description      string           `json:"description"`
	Features         []string         `json:"features"`
	SeatCount        int              `json:"seatCount" binding:"required"`
	TransmissionType string           `json:"transmissionType" binding:"required"`
	FuelType         string           `json:"fuelType" binding:"required"`
	VehicleType      string           `json:"vehicleType"`
	OwnerID          string           `json:"ownerId" binding:"required"`
	Location         *LocationWrapper `json:"location"`
}

type CreateBulkCarsDTO struct {
	Cars []CreateCarDTO `json:"cars" binding:"required"`
}
