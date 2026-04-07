package storage

import (
	"sync"

	"pistoni/core/internal/platform/model"
)

type MemoryStore struct {
	Mu sync.RWMutex

	Users                  map[string]model.User
	UserByEmail            map[string]string
	PendingByEmail         map[string]model.PendingEmailVerification
	Cars                   map[string]model.Car
	CarByLicensePlate      map[string]string
	LocationsByCarID       map[string]model.Location
	DriverProfiles         map[string]model.DriverProfile
	DriverProfileByUser    map[string]string
	DriverReviews          map[string]model.DriverReview
	DriverReviewsByID      map[string][]string
	BodyGuardProfiles      map[string]model.BodyGuardProfile
	BodyGuardByUser        map[string]string
	BodyGuardQueue         map[string]model.BodyGuardRegistrationQueueItem
	BodyGuardQueueByUser   map[string]string
	PaymentMethods         map[string]model.PaymentMethod
	PaymentMethodByUser    map[string][]string
	Rentals                map[string]model.Rental
	Notifications          map[string]model.Notification
	NotificationByUser     map[string][]string
	ChatRooms              map[string]model.ChatRoom
	Messages               map[string]model.Message
	MessagesByRoom         map[string][]string
	DriverStatusVersion    int64
	BodyGuardStatusVersion int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Users:                map[string]model.User{},
		UserByEmail:          map[string]string{},
		PendingByEmail:       map[string]model.PendingEmailVerification{},
		Cars:                 map[string]model.Car{},
		CarByLicensePlate:    map[string]string{},
		LocationsByCarID:     map[string]model.Location{},
		DriverProfiles:       map[string]model.DriverProfile{},
		DriverProfileByUser:  map[string]string{},
		DriverReviews:        map[string]model.DriverReview{},
		DriverReviewsByID:    map[string][]string{},
		BodyGuardProfiles:    map[string]model.BodyGuardProfile{},
		BodyGuardByUser:      map[string]string{},
		BodyGuardQueue:       map[string]model.BodyGuardRegistrationQueueItem{},
		BodyGuardQueueByUser: map[string]string{},
		PaymentMethods:       map[string]model.PaymentMethod{},
		PaymentMethodByUser:  map[string][]string{},
		Rentals:              map[string]model.Rental{},
		Notifications:        map[string]model.Notification{},
		NotificationByUser:   map[string][]string{},
		ChatRooms:            map[string]model.ChatRoom{},
		Messages:             map[string]model.Message{},
		MessagesByRoom:       map[string][]string{},
	}
}
