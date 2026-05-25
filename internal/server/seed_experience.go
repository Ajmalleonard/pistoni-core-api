package server

import (
	"fmt"
	"log"
	"sort"
	"time"

	accountdto "pistoni/core/internal/modules/account/dto"
	accountrepo "pistoni/core/internal/modules/account/repo"
	carrepo "pistoni/core/internal/modules/cars/repo"
	chatdto "pistoni/core/internal/modules/chat/dto"
	chatrepo "pistoni/core/internal/modules/chat/repo"
	notificationsdto "pistoni/core/internal/modules/notifications/dto"
	notificationsrepo "pistoni/core/internal/modules/notifications/repo"
	paymentmethodsdto "pistoni/core/internal/modules/paymentmethods/dto"
	paymentmethodsrepo "pistoni/core/internal/modules/paymentmethods/repo"
	providersdto "pistoni/core/internal/modules/providers/dto"
	providersrepo "pistoni/core/internal/modules/providers/repo"
	rentalsdto "pistoni/core/internal/modules/rentals/dto"
	rentalsrepo "pistoni/core/internal/modules/rentals/repo"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
	"pistoni/core/internal/platform/storage"
)

type seedDependencies struct {
	Store         *storage.MemoryStore
	Users         usersrepo.Repository
	Cars          carrepo.Repository
	Providers     providersrepo.Repository
	Accounts      accountrepo.Repository
	Payments      paymentmethodsrepo.Repository
	Rentals       rentalsrepo.Repository
	Chat          chatrepo.Repository
	Notifications notificationsrepo.Repository
}

type seedDriverActor struct {
	User    model.User
	Profile model.DriverProfile
}

type seedTripBundle struct {
	Rental     model.Rental
	Client     model.User
	Driver     *seedDriverActor
	BodyGuards []model.BodyGuardProfile
	ClientRoom model.ChatRoom
	OpsRoom    model.ChatRoom
	TripTitle  string
}

type seedTripPlan struct {
	Title          string
	ClientIndex    int
	CarType        model.VehicleType
	CarOccurrence  int
	DriverIndex    *int
	BodyGuardStart int
	BodyGuardCount int
	Status         model.RentalStatus
	StartOffset    time.Duration
	Duration       time.Duration
	DistanceKM     float64
	Pickup         string
	Dropoff        string
	SecurityNotes  string
	ClientLead     string
	DriverLead     string
	BodyGuardLead  string
	ClientNotice   string
	OperatorNotice string
}

func seedRealAppExperience(deps seedDependencies, admin model.User, seedPassword string) {
	clients := seedDemoClients(deps.Users, seedPassword)
	drivers := seedDemoDrivers(deps.Users, deps.Providers, seedPassword)
	if len(clients) == 0 || len(drivers) == 0 {
		log.Printf("[seed] real app experience skipped: clients=%d drivers=%d", len(clients), len(drivers))
		return
	}

	seedDemoAccountSurfaces(deps.Accounts, append([]model.User{admin}, clients...))
	seedDemoPaymentMethods(deps.Payments, append([]model.User{admin}, clients...))

	bodyguards := collectBodyGuardProfiles(deps.Store, 12)
	if len(bodyguards) == 0 {
		log.Printf("[seed] real app experience skipped: no body guards available")
		return
	}

	cars, err := deps.Cars.FindAll(carrepo.Query{})
	if err != nil {
		log.Printf("[seed] real app experience car lookup failed: %v", err)
		return
	}
	if len(cars) == 0 {
		log.Printf("[seed] real app experience skipped: no cars available")
		return
	}

	bundles := seedDemoTripFlows(deps, admin, clients, drivers, bodyguards, cars)
	seedTripNotifications(deps.Notifications, admin, bundles)

	log.Printf("[seed] real app experience ready clients=%d drivers=%d trips=%d rooms=%d notifications=%d payments=%d", len(clients), len(drivers), len(bundles), len(bundles)*2, len(bundles)*6, (len(clients)+1)*2)
}

func seedDemoClients(userRepo usersrepo.Repository, password string) []model.User {
	type demoPerson struct {
		Name           string
		Email          string
		Phone          string
		ProfilePicture string
	}

	people := []demoPerson{
		{Name: "Noor Hassan", Email: "noor.hassan@pistoni.local", Phone: "+255712400001", ProfilePicture: "https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=1200&q=80"},
		{Name: "Daniel Mrema", Email: "daniel.mrema@pistoni.local", Phone: "+255712400002", ProfilePicture: "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=1200&q=80"},
		{Name: "Amina Said", Email: "amina.said@pistoni.local", Phone: "+255712400003", ProfilePicture: "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?auto=format&fit=crop&w=1200&q=80"},
		{Name: "Kelvin Mwinyi", Email: "kelvin.mwinyi@pistoni.local", Phone: "+255712400004", ProfilePicture: "https://images.unsplash.com/photo-1521119989659-a83eee488004?auto=format&fit=crop&w=1200&q=80"},
	}

	clients := make([]model.User, 0, len(people))
	for _, person := range people {
		phone := person.Phone
		picture := person.ProfilePicture
		client, err := ensureSeedUser(userRepo, model.User{
			Email:          person.Email,
			Password:       &password,
			Name:           person.Name,
			PhoneNumber:    &phone,
			ProfilePicture: &picture,
			Role:           model.UserRoleClient,
			IsVerified:     true,
			AuthProvider:   model.AuthProviderEmail,
		})
		if err != nil {
			log.Printf("[seed] client %s: %v", person.Email, err)
			continue
		}
		clients = append(clients, client)
	}
	return clients
}

func seedDemoDrivers(userRepo usersrepo.Repository, providerRepo providersrepo.Repository, password string) []seedDriverActor {
	type demoDriver struct {
		Name           string
		Email          string
		Phone          string
		Picture        string
		Gender         model.Gender
		Rate           float64
		Years          int
		License        string
		Class          string
		Authority      string
		Address        string
		NationalID     string
		Notes          string
		Emergency      string
		EmergencyPhone string
	}

	drivers := []demoDriver{
		{Name: "Hassan Juma", Email: "hassan.juma@pistoni.local", Phone: "+255713400001", Picture: "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=1200&q=80", Gender: model.GenderMale, Rate: 220, Years: 9, License: "TZ-DRV-1001", Class: "D", Authority: "Land Transport Regulatory Authority", Address: "Masaki, Dar es Salaam", NationalID: "TZ-DR-NIN-1001", Notes: "Night route specialist", Emergency: "Juma Family", EmergencyPhone: "+255711500001"},
		{Name: "Fatuma Ally", Email: "fatuma.ally@pistoni.local", Phone: "+255713400002", Picture: "https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=1200&q=80", Gender: model.GenderFemale, Rate: 240, Years: 11, License: "TZ-DRV-1002", Class: "D", Authority: "Land Transport Regulatory Authority", Address: "Oysterbay, Dar es Salaam", NationalID: "TZ-DR-NIN-1002", Notes: "Airport transfer specialist", Emergency: "Ally Family", EmergencyPhone: "+255711500002"},
		{Name: "Joseph Mollel", Email: "joseph.mollel@pistoni.local", Phone: "+255713400003", Picture: "https://images.unsplash.com/photo-1519085360753-af0119f7cbe7?auto=format&fit=crop&w=1200&q=80", Gender: model.GenderMale, Rate: 205, Years: 7, License: "TZ-DRV-1003", Class: "C1", Authority: "Land Transport Regulatory Authority", Address: "Msasani, Dar es Salaam", NationalID: "TZ-DR-NIN-1003", Notes: "Executive convoy lead", Emergency: "Mollel Family", EmergencyPhone: "+255711500003"},
		{Name: "Neema Makori", Email: "neema.makori@pistoni.local", Phone: "+255713400004", Picture: "https://images.unsplash.com/photo-1488426862026-3ee34a7d66df?auto=format&fit=crop&w=1200&q=80", Gender: model.GenderFemale, Rate: 260, Years: 13, License: "TZ-DRV-1004", Class: "D", Authority: "Land Transport Regulatory Authority", Address: "Upanga, Dar es Salaam", NationalID: "TZ-DR-NIN-1004", Notes: "VIP client escort", Emergency: "Makori Family", EmergencyPhone: "+255711500004"},
	}

	seeded := make([]seedDriverActor, 0, len(drivers))
	for _, driver := range drivers {
		phone := driver.Phone
		picture := driver.Picture
		user, err := ensureSeedUser(userRepo, model.User{
			Email:          driver.Email,
			Password:       &password,
			Name:           driver.Name,
			PhoneNumber:    &phone,
			ProfilePicture: &picture,
			Role:           model.UserRoleClient,
			IsVerified:     true,
			AuthProvider:   model.AuthProviderEmail,
		})
		if err != nil {
			log.Printf("[seed] driver user %s: %v", driver.Email, err)
			continue
		}

		profile, err := providerRepo.RegisterDriver(user.ID, providersdto.RegisterDriverDTO{
			Gender:                string(driver.Gender),
			Address:               driver.Address,
			NIN:                   driver.NationalID,
			DriversLicenseNumber:  driver.License,
			DriversLicenseClass:   driver.Class,
			IssuingAuthority:      driver.Authority,
			YearsOfExperience:     &driver.Years,
			EmergencyContactName:  driver.Emergency,
			EmergencyContactPhone: driver.EmergencyPhone,
			DailyRate:             driver.Rate,
			Notes:                 driver.Notes,
		})
		if err != nil {
			if existing, existingErr := providerRepo.GetMyDriverProfile(user.ID); existingErr == nil {
				if profileID := stringField(existing, "id"); profileID != "" {
					profile = model.DriverProfile{
						ID:                 profileID,
						UserID:             user.ID,
						DailyRate:          driver.Rate,
						VerificationStatus: model.VerificationApproved,
						IsAvailable:        true,
						Status:             model.DriverStatusOnline,
					}
				}
			} else {
				log.Printf("[seed] driver profile %s: %v", driver.Email, err)
				continue
			}
		}

		approved := true
		available := true
		_ = approveDriverProfile(providerRepo, profile.ID, approved, available)
		_, _ = providerRepo.UpdateMyDriverStatus(user.ID, providersdto.UpdateDriverStatusDTO{
			Status:      string(model.DriverStatusOnline),
			IsAvailable: &available,
		})
		seeded = append(seeded, seedDriverActor{User: user, Profile: profile})
	}
	return seeded
}

func seedDemoAccountSurfaces(accountRepo accountrepo.Repository, users []model.User) {
	cities := []string{"Dar es Salaam", "Arusha", "Mwanza", "Zanzibar"}
	zones := []string{"Masaki", "Mikocheni", "Oysterbay", "Upanga"}
	protocols := []string{
		"Call operator, secure perimeter, route to safe house.",
		"Alert trip operator, keep live location on, move to nearest safe location.",
		"Notify family contact and switch to convoy mode.",
		"Use pinned emergency channel and standby with driver.",
	}
	for i, user := range users {
		language := []string{"en", "sw"}[i%2]
		timezone := "Africa/Dar_es_Salaam"
		emailNotifications := true
		pushNotifications := true
		smsNotifications := i%2 == 0
		textSize := []string{"medium", "large"}[i%2]
		voiceOver := false
		highContrast := i%3 == 0
		reduceMotion := i%2 == 1
		_, _ = accountRepo.UpdateSettings(user.ID, accountdto.UpdateSettingsDTO{
			Language:                  &language,
			Timezone:                  &timezone,
			EmailNotifications:        &emailNotifications,
			PushNotifications:         &pushNotifications,
			SMSNotifications:          &smsNotifications,
			TextSize:                  &textSize,
			AccessibilityVoiceOver:    &voiceOver,
			AccessibilityHighContrast: &highContrast,
			AccessibilityReduceMotion: &reduceMotion,
		})

		homeCity := cities[i%len(cities)]
		homeAddress := fmt.Sprintf("%d %s, %s", 10+i, zones[i%len(zones)], homeCity)
		protocol := protocols[i%len(protocols)]
		emergencyNotes := "Keep trip operator on first call."
		medicalNotes := "No known allergies."
		shareLiveLocation := true
		sosEnabled := true
		_, _ = accountRepo.UpdateSafety(user.ID, accountdto.UpdateSafetyDTO{
			HomeCity:          &homeCity,
			HomeAddress:       &homeAddress,
			EmergencyProtocol: &protocol,
			EmergencyNotes:    &emergencyNotes,
			MedicalNotes:      &medicalNotes,
			ShareLiveLocation: &shareLiveLocation,
			SOSEnabled:        &sosEnabled,
		})

		profileVisibility := []string{"private", "friends", "private", "private"}[i%4]
		locationSharing := []string{"whileInTrip", "whileInTrip", "always", "whileInTrip"}[i%4]
		allowMarketingEmails := false
		allowAnalytics := true
		allowMessagePreviews := true
		_, _ = accountRepo.UpdatePrivacy(user.ID, accountdto.UpdatePrivacyDTO{
			ProfileVisibility:    &profileVisibility,
			LocationSharing:      &locationSharing,
			AllowMarketingEmails: &allowMarketingEmails,
			AllowAnalytics:       &allowAnalytics,
			AllowMessagePreviews: &allowMessagePreviews,
		})

		twoFactorEnabled := i%2 == 0
		twoFactorMethod := []string{"authenticator", "sms"}[i%2]
		sessionAlerts := true
		loginNotifications := true
		_, _ = accountRepo.UpdateSecurity(user.ID, accountdto.UpdateSecurityDTO{
			TwoFactorEnabled:   &twoFactorEnabled,
			TwoFactorMethod:    &twoFactorMethod,
			SessionAlerts:      &sessionAlerts,
			LoginNotifications: &loginNotifications,
		})

		receiveTripAlerts := true
		primary := true
		_, _ = accountRepo.CreateTrustedContact(user.ID, accountdto.CreateTrustedContactDTO{
			Name:              fmt.Sprintf("%s Support", user.Name),
			PhoneNumber:       fmt.Sprintf("+255714500%03d", i+1),
			Relationship:      "trip contact",
			Email:             fmt.Sprintf("support-%d@pistoni.local", i+1),
			ReceiveTripAlerts: &receiveTripAlerts,
			IsPrimary:         &primary,
		})
		_, _ = accountRepo.CreateFamilyMember(user.ID, accountdto.CreateFamilyMemberDTO{
			Name:         fmt.Sprintf("%s Family", user.Name),
			Relationship: "family",
			PhoneNumber:  fmt.Sprintf("+255715600%03d", i+1),
			Email:        fmt.Sprintf("family-%d@pistoni.local", i+1),
			AccessLevel:  "view",
		})
		current := true
		_, _ = accountRepo.CreateTrustedDevice(user.ID, accountdto.CreateTrustedDeviceDTO{
			DeviceName: fmt.Sprintf("%s iPhone", user.Name),
			Platform:   "iOS",
			IsCurrent:  &current,
		})
		expiry := time.Now().AddDate(4, 0, 0)
		_, _ = accountRepo.CreateDocument(user.ID, accountdto.CreatePersonalDocumentDTO{
			DocumentType:       "national_id",
			Label:              "Identity document",
			DocumentNumber:     fmt.Sprintf("ID-%04d", i+1),
			IssuingCountry:     "Tanzania",
			VerificationStatus: "APPROVED",
			FileURL:            fmt.Sprintf("https://images.unsplash.com/photo-1516054575922-1a5f2f1f7e5f?auto=format&fit=crop&w=1200&q=80&sig=%d", i+1),
			ExpiryDate:         &expiry,
		})
		reminder := 30
		_, _ = accountRepo.CreateCalendar(user.ID, accountdto.CreateCalendarConnectionDTO{
			Provider:              "Google Calendar",
			CalendarName:          fmt.Sprintf("%s Travel Calendar", user.Name),
			ExternalCalendarID:    fmt.Sprintf("cal-%s", shortIdentifier(user.ID)),
			SyncEnabled:           &current,
			ReminderMinutesBefore: &reminder,
		})
	}
}

func seedDemoPaymentMethods(paymentRepo paymentmethodsrepo.Repository, users []model.User) {
	for i, user := range users {
		if i == 0 {
			_, _ = paymentRepo.CreateCardPaymentMethod(user.ID, paymentmethodsdto.CreateCardPaymentMethodDTO{
				CardNumber:     fmt.Sprintf("424242424242%04d", 1000+i),
				CardType:       "Visa",
				ExpiryMonth:    12,
				ExpiryYear:     time.Now().Year() + 4,
				CardholderName: user.Name,
				IsDefault:      true,
			}, fmt.Sprintf("%04d", 1000+i))
		} else {
			_, _ = paymentRepo.CreateCardPaymentMethod(user.ID, paymentmethodsdto.CreateCardPaymentMethodDTO{
				CardNumber:     fmt.Sprintf("555555555555%04d", 2000+i),
				CardType:       "Mastercard",
				ExpiryMonth:    11,
				ExpiryYear:     time.Now().Year() + 3,
				CardholderName: user.Name,
				IsDefault:      true,
			}, fmt.Sprintf("%04d", 2000+i))
		}
		_, _ = paymentRepo.CreateMobilePaymentMethod(user.ID, paymentmethodsdto.CreateMobilePaymentMethodDTO{
			MobileNumber: fmt.Sprintf("+255717700%03d", i+1),
			Provider:     []string{"M-Pesa", "Tigo Pesa", "Airtel Money"}[i%3],
			AccountName:  user.Name,
			IsDefault:    i != 0,
		})
	}
}

func seedDemoTripFlows(deps seedDependencies, admin model.User, clients []model.User, drivers []seedDriverActor, bodyguards []model.BodyGuardProfile, cars []model.CarView) []seedTripBundle {
	plans := []seedTripPlan{
		{
			Title:          "Airport arrival escort",
			ClientIndex:    0,
			CarType:        model.VehicleTypeSUV,
			CarOccurrence:  0,
			DriverIndex:    intPtr(0),
			BodyGuardStart: 0,
			BodyGuardCount: 2,
			Status:         model.RentalStatusConfirmed,
			StartOffset:    24 * time.Hour,
			Duration:       36 * time.Hour,
			DistanceKM:     37.5,
			Pickup:         "Masaki, Dar es Salaam",
			Dropoff:        "Julius Nyerere International Airport",
			SecurityNotes:  "VIP arrival from airport",
			ClientLead:     "Trip confirmed and crew scheduled.",
			DriverLead:     "Vehicle pre-check done and route is clean.",
			BodyGuardLead:  "Maintain a quiet perimeter and keep the pickup lane clear.",
			ClientNotice:   "Your airport escort is confirmed and the team is ready.",
			OperatorNotice: "Airport arrival escort is briefed for tomorrow.",
		},
		{
			Title:          "Board transfer convoy",
			ClientIndex:    1,
			CarType:        model.VehicleTypeCoaster,
			CarOccurrence:  0,
			DriverIndex:    intPtr(1),
			BodyGuardStart: 2,
			BodyGuardCount: 4,
			Status:         model.RentalStatusActive,
			StartOffset:    -4 * time.Hour,
			Duration:       18 * time.Hour,
			DistanceKM:     18,
			Pickup:         "Mlimani City, Dar es Salaam",
			Dropoff:        "Pangani, Dar es Salaam",
			SecurityNotes:  "Board members moving between venues",
			ClientLead:     "Client has live route visibility.",
			DriverLead:     "We are on route and bodyguards are stationed.",
			BodyGuardLead:  "Keep both exits clean and stay in contact with driver.",
			ClientNotice:   "Your convoy is already active and the operator is tracking it.",
			OperatorNotice: "Board convoy is active with security detail attached.",
		},
		{
			Title:          "Evening event escort",
			ClientIndex:    2,
			CarType:        model.VehicleTypeLongVehicle,
			CarOccurrence:  0,
			DriverIndex:    nil,
			BodyGuardStart: 6,
			BodyGuardCount: 3,
			Status:         model.RentalStatusPending,
			StartOffset:    48 * time.Hour,
			Duration:       24 * time.Hour,
			DistanceKM:     11,
			Pickup:         "Oysterbay, Dar es Salaam",
			Dropoff:        "Sea Cliff Hotel, Dar es Salaam",
			SecurityNotes:  "No driver selected yet, security hold only",
			ClientLead:     "Waiting for driver selection.",
			DriverLead:     "Driver selection is still open for this booking.",
			BodyGuardLead:  "Hold position until the operator confirms transport.",
			ClientNotice:   "Your booking is saved. Driver selection is still open.",
			OperatorNotice: "Driver selection is still open for the evening escort.",
		},
		{
			Title:          "Family resort return",
			ClientIndex:    3,
			CarType:        model.VehicleTypeSUV,
			CarOccurrence:  1,
			DriverIndex:    intPtr(2),
			BodyGuardStart: 9,
			BodyGuardCount: 1,
			Status:         model.RentalStatusCompleted,
			StartOffset:    -72 * time.Hour,
			Duration:       30 * time.Hour,
			DistanceKM:     61,
			Pickup:         "Upanga, Dar es Salaam",
			Dropoff:        "White Sands Resort, Dar es Salaam",
			SecurityNotes:  "Family return transfer",
			ClientLead:     "Trip closed and review request sent.",
			DriverLead:     "Trip completed and route counter updated.",
			BodyGuardLead:  "Escort closed and crew returned to standby.",
			ClientNotice:   "Your family return transfer is complete.",
			OperatorNotice: "Completed trip is archived and ready for review.",
		},
	}

	bundles := make([]seedTripBundle, 0, len(plans))
	for _, plan := range plans {
		if plan.ClientIndex >= len(clients) {
			continue
		}
		car, ok := pickCarByType(cars, plan.CarType, plan.CarOccurrence)
		if !ok {
			continue
		}

		client := clients[plan.ClientIndex]
		var driver *seedDriverActor
		if plan.DriverIndex != nil && *plan.DriverIndex >= 0 && *plan.DriverIndex < len(drivers) {
			driver = &drivers[*plan.DriverIndex]
		}
		bgSlice := seedBodyGuardSlice(bodyguards, plan.BodyGuardStart, plan.BodyGuardCount)
		bgIDs := make([]string, 0, len(bgSlice))
		for _, bg := range bgSlice {
			bgIDs = append(bgIDs, bg.ID)
		}

		start := time.Now().Add(plan.StartOffset)
		end := start.Add(plan.Duration)
		var driverID *string
		if driver != nil {
			id := driver.Profile.ID
			driverID = &id
		}
		distance := plan.DistanceKM
		priceSummary, err := deps.Rentals.CalculateRentalPrice(start, end, car.ID, distance, driverID, bgIDs)
		if err != nil {
			log.Printf("[seed] rental pricing %s: %v", plan.Title, err)
			continue
		}

		driverRequired := driverID != nil
		securityRequired := len(bgIDs) > 0
		createdTrip, err := deps.Rentals.Create(rentalsdto.CreateRentalDTO{
			StartDate:         start.Format(time.RFC3339),
			EndDate:           end.Format(time.RFC3339),
			TotalPrice:        priceSummary["totalPrice"].(float64),
			Status:            string(plan.Status),
			DistanceKM:        &distance,
			UserID:            client.ID,
			CarID:             car.ID,
			PickupLocationID:  plan.Pickup,
			DropoffLocationID: plan.Dropoff,
			DriverRequired:    &driverRequired,
			DriverID:          driverID,
			SecurityRequired:  &securityRequired,
			SecurityNotes:     &plan.SecurityNotes,
			BodyGuardIDs:      bgIDs,
		})
		if err != nil {
			log.Printf("[seed] rental create %s: %v", plan.Title, err)
			continue
		}

		if driver != nil {
			isAvailable := true
			status := string(model.DriverStatusOnline)
			if plan.Status == model.RentalStatusActive {
				status = string(model.DriverStatusOnTrip)
				isAvailable = false
			}
			_, _ = deps.Providers.UpdateMyDriverStatus(driver.User.ID, providersdto.UpdateDriverStatusDTO{Status: status, IsAvailable: &isAvailable})
			_, _ = deps.Providers.UpdateDriverAssignmentCounters(driver.Profile.ID, providersdto.DriverAssignmentDeltaDTO{RouteDelta: 1, AssignmentDelta: 1})
			if plan.Status == model.RentalStatusCompleted {
				comment := fmt.Sprintf("%s finished smoothly with a polished handoff.", plan.Title)
				_, _ = deps.Providers.CreateDriverReview(driver.Profile.ID, client.ID, providersdto.CreateDriverReviewDTO{Rating: 5, Comment: comment})
			}
		}

		for _, bg := range bgSlice {
			isAvailable := true
			status := string(model.BodyGuardStatusOnline)
			if plan.Status == model.RentalStatusActive {
				status = string(model.BodyGuardStatusAssigned)
				isAvailable = false
			}
			_, _ = deps.Providers.UpdateMyBodyGuardStatus(bg.UserID, providersdto.UpdateBodyGuardStatusDTO{Status: status, IsAvailable: &isAvailable})
		}

		clientRoomName := fmt.Sprintf("%s Client Room", plan.Title)
		clientRoom, err := deps.Chat.CreatePrivateChat(client.ID, admin.ID)
		if err != nil {
			log.Printf("[seed] client room %s: %v", plan.Title, err)
			continue
		}
		clientRoom.Name = &clientRoomName
		deps.Store.ChatRooms[clientRoom.ID] = clientRoom

		participants := []string{admin.ID}
		if driver != nil {
			participants = append(participants, driver.User.ID)
		}
		for _, bg := range bgSlice {
			participants = append(participants, bg.UserID)
		}
		participants = uniqueStrings(participants)
		opsRoomName := fmt.Sprintf("%s Ops Room", plan.Title)
		creatorID := admin.ID
		if driver != nil {
			creatorID = driver.User.ID
		}
		opsRoom, err := deps.Chat.CreateGroupChat(creatorID, opsRoomName, participants)
		if err != nil {
			log.Printf("[seed] ops room %s: %v", plan.Title, err)
			continue
		}

		_, _ = deps.Chat.CreateMessage(client.ID, chatdto.CreateMessageDTO{ChatRoomID: clientRoom.ID, Content: plan.ClientLead})
		_, _ = deps.Chat.CreateMessage(admin.ID, chatdto.CreateMessageDTO{ChatRoomID: clientRoom.ID, Content: plan.ClientNotice})
		_, _ = deps.Chat.CreateMessage(admin.ID, chatdto.CreateMessageDTO{ChatRoomID: opsRoom.ID, Content: plan.OperatorNotice})
		if driver != nil {
			_, _ = deps.Chat.CreateMessage(driver.User.ID, chatdto.CreateMessageDTO{ChatRoomID: opsRoom.ID, Content: plan.DriverLead})
			if len(bgIDs) > 0 {
				_, _ = deps.Chat.CreateTargetedMessage(driver.User.ID, chatdto.CreateTargetedMessageDTO{TripID: createdTrip.ID, ChatRoomID: opsRoom.ID, Content: plan.BodyGuardLead, TargetType: "bodyguard-group", TargetIDs: bgIDs})
			}
		} else if len(bgIDs) > 0 {
			_, _ = deps.Chat.CreateMessage(admin.ID, chatdto.CreateMessageDTO{ChatRoomID: opsRoom.ID, Content: plan.BodyGuardLead})
		}

		bundles = append(bundles, seedTripBundle{
			Rental:     createdTrip,
			Client:     client,
			Driver:     driver,
			BodyGuards: append([]model.BodyGuardProfile(nil), bgSlice...),
			ClientRoom: clientRoom,
			OpsRoom:    opsRoom,
			TripTitle:  plan.Title,
		})
	}

	return bundles
}

func seedTripNotifications(notificationRepo notificationsrepo.Repository, admin model.User, bundles []seedTripBundle) {
	if len(bundles) == 0 {
		return
	}
	for _, bundle := range bundles {
		clientImage := firstImage(bundle.Client.ProfilePicture, nil)
		imageURL := ""
		if clientImage != nil {
			imageURL = *clientImage
		}
		tripURL := fmt.Sprintf("/rentals/%s", bundle.Rental.ID)
		chatURL := fmt.Sprintf("/chat/rooms/%s", bundle.ClientRoom.ID)
		_, _ = notificationRepo.CreateNotification(notificationsdto.CreateNotificationDTO{
			UserID:          bundle.Client.ID,
			Title:           fmt.Sprintf("%s is ready", bundle.TripTitle),
			Body:            fmt.Sprintf("Your booking to %s has been scheduled and the crew is on standby.", bundle.Rental.DropoffLocationID),
			Type:            string(model.NotificationTypeRental),
			RelatedEntityID: bundle.Rental.ID,
			ImageURL:        imageURL,
			ActionURL:       tripURL,
		})
		_, _ = notificationRepo.CreateNotification(notificationsdto.CreateNotificationDTO{
			UserID:          bundle.Client.ID,
			Title:           "New trip message",
			Body:            "Trip operator has posted an update in your concierge room.",
			Type:            string(model.NotificationTypeChat),
			RelatedEntityID: bundle.ClientRoom.ID,
			ImageURL:        imageURL,
			ActionURL:       chatURL,
		})
		if bundle.Driver != nil {
			_, _ = notificationRepo.CreateNotification(notificationsdto.CreateNotificationDTO{
				UserID:          bundle.Driver.User.ID,
				Title:           "Driver assignment live",
				Body:            "Your ops room has the current trip brief and the route is active.",
				Type:            string(model.NotificationTypeChat),
				RelatedEntityID: bundle.OpsRoom.ID,
				ImageURL:        imageURL,
				ActionURL:       fmt.Sprintf("/chat/rooms/%s", bundle.OpsRoom.ID),
			})
		}
		for _, bg := range bundle.BodyGuards {
			_, _ = notificationRepo.CreateNotification(notificationsdto.CreateNotificationDTO{
				UserID:          bg.UserID,
				Title:           "Escort detail ready",
				Body:            fmt.Sprintf("%s is in the active security board and ready for the next handoff.", bundle.TripTitle),
				Type:            string(model.NotificationTypeChat),
				RelatedEntityID: bundle.OpsRoom.ID,
				ImageURL:        imageURL,
				ActionURL:       fmt.Sprintf("/chat/rooms/%s", bundle.OpsRoom.ID),
			})
		}
		_, _ = notificationRepo.CreateNotification(notificationsdto.CreateNotificationDTO{
			UserID:          admin.ID,
			Title:           fmt.Sprintf("%s updated", bundle.TripTitle),
			Body:            fmt.Sprintf("Trip %s is staged with %d escort profiles.", bundle.Rental.ID, len(bundle.BodyGuards)),
			Type:            string(model.NotificationTypeSystem),
			RelatedEntityID: bundle.Rental.ID,
			ImageURL:        imageURL,
			ActionURL:       tripURL,
		})
	}

	welcomeTargets := []model.User{admin}
	for _, bundle := range bundles {
		welcomeTargets = append(welcomeTargets, bundle.Client)
	}
	for i, user := range welcomeTargets {
		_, _ = notificationRepo.CreateNotification(notificationsdto.CreateNotificationDTO{
			UserID:          user.ID,
			Title:           "Pistoni demo ready",
			Body:            fmt.Sprintf("Your dashboard is seeded with live trip, crew, and chat data for flow %d.", i+1),
			Type:            string(model.NotificationTypePromotion),
			RelatedEntityID: "seed-demo",
			ActionURL:       "/chat/rooms",
		})
	}
}

func collectBodyGuardProfiles(store *storage.MemoryStore, limit int) []model.BodyGuardProfile {
	store.Mu.RLock()
	items := make([]model.BodyGuardProfile, 0, len(store.BodyGuardProfiles))
	for _, profile := range store.BodyGuardProfiles {
		items = append(items, profile)
	}
	store.Mu.RUnlock()
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func seedBodyGuardSlice(items []model.BodyGuardProfile, start, count int) []model.BodyGuardProfile {
	if start < 0 || start >= len(items) || count <= 0 {
		return nil
	}
	end := start + count
	if end > len(items) {
		end = len(items)
	}
	return append([]model.BodyGuardProfile(nil), items[start:end]...)
}

func pickCarByType(cars []model.CarView, vehicleType model.VehicleType, occurrence int) (model.CarView, bool) {
	seen := 0
	for _, car := range cars {
		if car.VehicleType != vehicleType || car.Status != model.CarStatusAvailable {
			continue
		}
		if seen == occurrence {
			return car, true
		}
		seen++
	}
	return model.CarView{}, false
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func intPtr(v int) *int { return &v }

func driverProfileOrNil(driver *seedDriverActor) *model.DriverProfile {
	if driver == nil {
		return nil
	}
	p := driver.Profile
	return &p
}

func firstImage(image *string, fallback *string) *string {
	if image != nil && *image != "" {
		return image
	}
	return fallback
}

func approveDriverProfile(providerRepo providersrepo.Repository, profileID string, approved, available bool) error {
	_, err := providerRepo.UpdateDriverVerification(profileID, providersdto.UpdateVerificationDTO{
		VerificationStatus:       string(model.VerificationApproved),
		BackgroundCheckCompleted: &approved,
		IsAvailable:              &available,
		Notes:                    "Seeded demo driver for booking flow",
	})
	return err
}
