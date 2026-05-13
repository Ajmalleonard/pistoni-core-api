package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"pistoni/core/internal/config"
	carsdto "pistoni/core/internal/modules/cars/dto"
	carrepo "pistoni/core/internal/modules/cars/repo"
	providersdto "pistoni/core/internal/modules/providers/dto"
	providersrepo "pistoni/core/internal/modules/providers/repo"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/model"
)

const productionSeedTimeout = 5 * time.Second

type productionSnapshot struct {
	Users      []map[string]any
	Cars       []map[string]any
	BodyGuards []map[string]any
}

func seedDevelopmentData(cfg config.Config, userRepo usersrepo.Repository, carRepo carrepo.Repository, providerRepo providersrepo.Repository) {
	seedPassword, err := hashSeedPassword()
	if err != nil {
		log.Printf("[seed] demo password hash failed: %v", err)
		return
	}

	var snapshot productionSnapshot
	if cfg.SeedFromProduction {
		var snapshotErr error
		snapshot, snapshotErr = fetchProductionSnapshot(cfg.SeedSourceURL)
		if snapshotErr != nil {
			log.Printf("[seed] production snapshot unavailable from %s: %v", cfg.SeedSourceURL, snapshotErr)
		}

		syncedUsers := syncProductionUsers(userRepo, snapshot.Users, seedPassword)
		if syncedUsers > 0 {
			log.Printf("[seed] synced %d users from %s", syncedUsers, cfg.SeedSourceURL)
		}
	} else {
		log.Printf("[seed] production sync disabled; using local fallback catalog")
	}

	fleetOwner, err := seedFleetOwner(userRepo, seedPassword)
	if err != nil {
		log.Printf("[seed] fleet owner seed failed, falling back to admin: %v", err)
		fleetOwner, err = userRepo.FindByEmail("admin@pistoni.local")
		if err != nil {
			log.Printf("[seed] admin lookup failed, demo data skipped: %v", err)
			return
		}
	}

	bodyguardsCreated := 0
	if len(snapshot.BodyGuards) > 0 {
		bodyguardsCreated = syncProductionBodyGuards(userRepo, providerRepo, snapshot.BodyGuards)
	}
	if bodyguardsCreated == 0 {
		bodyguardsCreated = seedDemoBodyGuards(userRepo, providerRepo, seedPassword)
	}

	carsCreated := 0
	if len(snapshot.Cars) > 0 {
		carsCreated = syncProductionCars(carRepo, userRepo, snapshot.Cars, fleetOwner.ID)
	}
	if carsCreated == 0 {
		carsCreated = seedDemoCars(carRepo, fleetOwner.ID)
	}

	log.Printf("[seed] core catalog ready source=%s owner=%s bodyguards=%d cars=%d", cfg.SeedSourceURL, fleetOwner.Email, bodyguardsCreated, carsCreated)
}

func seedFleetOwner(userRepo usersrepo.Repository, password string) (model.User, error) {
	if owners, err := userRepo.FindByRole(model.UserRoleCarOwner); err == nil && len(owners) > 0 {
		return owners[0], nil
	}

	email := "fleet.owner@pistoni.local"
	phone := "+255700900000"
	picture := "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=1200&q=80"
	if existing, err := userRepo.FindByEmail(email); err == nil {
		return existing, nil
	}
	return ensureSeedUser(userRepo, model.User{
		Email:          email,
		Password:       &password,
		Name:           "Fleet Owner",
		PhoneNumber:    &phone,
		ProfilePicture: &picture,
		Role:           model.UserRoleCarOwner,
		IsVerified:     true,
		AuthProvider:   model.AuthProviderEmail,
	})
}

func seedDemoBodyGuards(userRepo usersrepo.Repository, providerRepo providersrepo.Repository, password string) int {
	type seedPerson struct {
		First  string
		Gender model.Gender
	}
	type seedArea struct {
		Address string
		City    string
	}

	firstNames := []seedPerson{
		{First: "Amina", Gender: model.GenderFemale},
		{First: "Neema", Gender: model.GenderFemale},
		{First: "Zawadi", Gender: model.GenderFemale},
		{First: "Glory", Gender: model.GenderFemale},
		{First: "Anna", Gender: model.GenderFemale},
		{First: "Hassan", Gender: model.GenderMale},
		{First: "Juma", Gender: model.GenderMale},
		{First: "David", Gender: model.GenderMale},
		{First: "Emmanuel", Gender: model.GenderMale},
		{First: "Patrick", Gender: model.GenderMale},
	}
	lastNames := []string{"Suleiman", "Mollel", "Nyerere", "Mushi", "Joseph"}
	images := []string{
		"https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1438761681033-6461ffad8d80?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1488426862026-3ee34a7d66df?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1539571696357-5a69c17a67c6?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1544005313-94ddf0286df2?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1521119989659-a83eee488004?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1524504388940-b1c1722653e1?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?auto=format&fit=crop&w=1200&q=80",
	}
	areas := []seedArea{
		{Address: "Masaki, Dar es Salaam", City: "Dar es Salaam"},
		{Address: "Mikocheni, Dar es Salaam", City: "Dar es Salaam"},
		{Address: "Oysterbay, Dar es Salaam", City: "Dar es Salaam"},
		{Address: "Upanga, Dar es Salaam", City: "Dar es Salaam"},
		{Address: "Msasani, Dar es Salaam", City: "Dar es Salaam"},
	}

	approved := true
	available := true
	created := 0

	for i := 0; i < 50; i++ {
		person := firstNames[i%len(firstNames)]
		last := lastNames[(i/len(firstNames))%len(lastNames)]
		area := areas[i%len(areas)]
		name := fmt.Sprintf("%s %s", person.First, last)
		email := fmt.Sprintf("bodyguard%02d@pistoni.local", i+1)
		phone := fmt.Sprintf("+255710%05d", 10000+i)
		profilePicture := images[i%len(images)]
		years := 4 + (i % 12)
		bodyGuardUser, err := ensureSeedUser(userRepo, model.User{
			Email:          email,
			Password:       &password,
			Name:           name,
			PhoneNumber:    &phone,
			ProfilePicture: &profilePicture,
			Role:           model.UserRoleBodyGuard,
			IsVerified:     true,
			AuthProvider:   model.AuthProviderEmail,
		})
		if err != nil {
			log.Printf("[seed] bodyguard user %s: %v", email, err)
			continue
		}

		if profile, err := providerRepo.GetMyBodyGuardProfile(bodyGuardUser.ID); err == nil {
			if profileID, ok := profile["id"].(string); ok {
				if err := approveBodyGuardProfile(providerRepo, profileID, approved, available); err == nil {
					created++
				}
				continue
			}
		}

		dob := time.Date(1985+(i%12), time.Month((i%12)+1), (i%27)+1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
		expiry := time.Date(2031+(i%3), time.Month((i%12)+1), 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
		experience := years
		seedProfile, err := providerRepo.RegisterBodyGuard(bodyGuardUser.ID, providersdto.RegisterBodyGuardDTO{
			Gender:                string(person.Gender),
			DateOfBirth:           dob,
			Address:               area.Address,
			NIN:                   fmt.Sprintf("TZ-BG-NIN-%03d", i+1),
			GovernmentIDNumber:    fmt.Sprintf("TZ-BG-GOV-%03d", i+1),
			SecurityLicenseNumber: fmt.Sprintf("TZ-BG-SEC-%03d", i+1),
			SecurityLicenseExpiry: expiry,
			IssuingAuthority:      "Tanzania Private Security Authority",
			YearsOfExperience:     &experience,
			EmergencyContactName:  fmt.Sprintf("%s Family", last),
			EmergencyContactPhone: fmt.Sprintf("+255711%05d", 20000+i),
			DailyRate:             float64(180 + (i%10)*15),
			Notes:                 "Seeded demo bodyguard for booking flow",
		})
		if err != nil {
			log.Printf("[seed] bodyguard profile %s: %v", email, err)
			continue
		}
		if err := approveBodyGuardProfile(providerRepo, seedProfile.ID, approved, available); err != nil {
			log.Printf("[seed] bodyguard approval %s: %v", email, err)
			continue
		}
		created++
	}

	return created
}

func seedDemoCars(carRepo carrepo.Repository, ownerID string) int {
	type carBlueprint struct {
		Make             string
		Model            string
		VehicleType      model.VehicleType
		TransmissionType string
		FuelType         string
		SeatCount        int
		BasePrice        float64
		Features         []string
		Description      string
	}
	type seedLocation struct {
		Latitude  float64
		Longitude float64
		Address   string
		City      string
	}

	carBlueprints := []carBlueprint{
		{
			Make:             "Toyota",
			Model:            "Land Cruiser 300",
			VehicleType:      model.VehicleTypeSUV,
			TransmissionType: "Automatic",
			FuelType:         "Diesel",
			SeatCount:        7,
			BasePrice:        260000,
			Features:         []string{"4x4 traction", "Leather seats", "Climate control"},
			Description:      "Executive SUV for premium route travel.",
		},
		{
			Make:             "Land Rover",
			Model:            "Defender 110",
			VehicleType:      model.VehicleTypeSUV,
			TransmissionType: "Automatic",
			FuelType:         "Diesel",
			SeatCount:        7,
			BasePrice:        275000,
			Features:         []string{"Terrain response", "Panoramic roof", "Off-road package"},
			Description:      "Rugged premium SUV built for mixed terrain.",
		},
		{
			Make:             "Mercedes-Benz",
			Model:            "V-Class",
			VehicleType:      model.VehicleTypeLongVehicle,
			TransmissionType: "Automatic",
			FuelType:         "Petrol",
			SeatCount:        6,
			BasePrice:        320000,
			Features:         []string{"Captain seats", "Ambient lighting", "Premium sound"},
			Description:      "Luxury van for executive group transfers.",
		},
		{
			Make:             "Toyota",
			Model:            "Hiace Super Custom",
			VehicleType:      model.VehicleTypeCoaster,
			TransmissionType: "Automatic",
			FuelType:         "Diesel",
			SeatCount:        13,
			BasePrice:        180000,
			Features:         []string{"Rear AC", "Wide cabin", "Luggage space"},
			Description:      "Comfortable people mover for larger groups.",
		},
		{
			Make:             "Nissan",
			Model:            "Patrol",
			VehicleType:      model.VehicleTypeSUV,
			TransmissionType: "Automatic",
			FuelType:         "Petrol",
			SeatCount:        7,
			BasePrice:        255000,
			Features:         []string{"Adaptive cruise", "Premium leather", "4x4 drive"},
			Description:      "Bold SUV for secure and comfortable travel.",
		},
		{
			Make:             "Toyota",
			Model:            "Coaster Executive",
			VehicleType:      model.VehicleTypeCoaster,
			TransmissionType: "Automatic",
			FuelType:         "Diesel",
			SeatCount:        15,
			BasePrice:        200000,
			Features:         []string{"Executive seating", "PA system", "Wide aisle"},
			Description:      "Executive shuttle for team and event transfers.",
		},
		{
			Make:             "GMC",
			Model:            "Yukon Denali",
			VehicleType:      model.VehicleTypeSUV,
			TransmissionType: "Automatic",
			FuelType:         "Petrol",
			SeatCount:        7,
			BasePrice:        300000,
			Features:         []string{"V8 power", "Entertainment screen", "Heated seats"},
			Description:      "Full-size SUV with a strong executive feel.",
		},
		{
			Make:             "Toyota",
			Model:            "Alphard",
			VehicleType:      model.VehicleTypeLongVehicle,
			TransmissionType: "Automatic",
			FuelType:         "Petrol",
			SeatCount:        6,
			BasePrice:        290000,
			Features:         []string{"Lounge seating", "Power doors", "Soft suspension"},
			Description:      "Luxury people carrier for VIP routes.",
		},
		{
			Make:             "Cadillac",
			Model:            "Escalade",
			VehicleType:      model.VehicleTypeSUV,
			TransmissionType: "Automatic",
			FuelType:         "Petrol",
			SeatCount:        7,
			BasePrice:        340000,
			Features:         []string{"Massage seats", "Premium audio", "Heads-up display"},
			Description:      "High-end SUV for standout route experiences.",
		},
		{
			Make:             "Toyota",
			Model:            "Prado TX",
			VehicleType:      model.VehicleTypeSUV,
			TransmissionType: "Automatic",
			FuelType:         "Diesel",
			SeatCount:        7,
			BasePrice:        230000,
			Features:         []string{"Reliable 4x4", "Spacious boot", "Tour-ready trim"},
			Description:      "Reliable and familiar SUV for everyday premium trips.",
		},
	}

	colors := []string{"Midnight Black", "Pearl White", "Graphite Grey", "Forest Green", "Royal Blue"}
	images := []string{
		"https://images.unsplash.com/photo-1503376780353-7e6692767b70?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1511919884226-fd3cad34687c?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1494976388531-d1058494cdd8?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1542362567-b07e54358753?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1519641471654-76ce0107ad1b?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1553440569-bcc63803a83d?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1486262715619-67b85e0b08d3?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1525609004556-c46c7d6cf023?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1492144534655-ae79c964c9d7?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1503376780353-7e6692767b70?auto=format&fit=crop&w=1200&q=80",
	}
	locations := []seedLocation{
		{Latitude: -6.7762, Longitude: 39.2216, Address: "Masaki, Dar es Salaam", City: "Dar es Salaam"},
		{Latitude: -6.7701, Longitude: 39.2447, Address: "Mikocheni, Dar es Salaam", City: "Dar es Salaam"},
		{Latitude: -6.7585, Longitude: 39.2796, Address: "Oysterbay, Dar es Salaam", City: "Dar es Salaam"},
		{Latitude: -6.8253, Longitude: 39.2858, Address: "Upanga, Dar es Salaam", City: "Dar es Salaam"},
		{Latitude: -6.7512, Longitude: 39.2678, Address: "Msasani, Dar es Salaam", City: "Dar es Salaam"},
	}

	existingCars, err := carRepo.FindAll(carrepo.Query{})
	if err != nil {
		log.Printf("[seed] car lookup failed: %v", err)
		existingCars = []model.CarView{}
	}

	created := 0
	for i := 0; i < 50; i++ {
		blueprint := carBlueprints[i%len(carBlueprints)]
		color := colors[(i/len(carBlueprints))%len(colors)]
		location := locations[i%len(locations)]
		plate := fmt.Sprintf("PST-DEMO-%03d", i+1)
		if existing, ok := findSeedCarByPlate(existingCars, plate); ok {
			if _, err := carRepo.Update(existing.ID, map[string]any{
				"status":      string(model.CarStatusAvailable),
				"vehicleType": string(blueprint.VehicleType),
			}); err == nil {
				created++
			} else {
				log.Printf("[seed] car update %s: %v", plate, err)
			}
			continue
		}

		year := 2021 + (i % 4)
		car, err := carRepo.Create(carsdto.CreateCarDTO{
			Make:             blueprint.Make,
			Model:            blueprint.Model,
			Year:             year,
			Color:            color,
			LicensePlate:     plate,
			Price:            blueprint.BasePrice + float64((i%5)*5000),
			Images:           []string{images[i%len(images)], images[(i+3)%len(images)]},
			Description:      blueprint.Description,
			Features:         append([]string(nil), blueprint.Features...),
			SeatCount:        blueprint.SeatCount,
			TransmissionType: blueprint.TransmissionType,
			FuelType:         blueprint.FuelType,
			VehicleType:      string(blueprint.VehicleType),
			OwnerID:          ownerID,
			Location: &carsdto.LocationWrapper{
				Create: &carsdto.LocationCreate{
					Latitude:  location.Latitude,
					Longitude: location.Longitude,
					Address:   location.Address,
					City:      location.City,
					Country:   "Tanzania",
				},
			},
		})
		if err != nil {
			log.Printf("[seed] car create %s: %v", plate, err)
			continue
		}
		if _, err := carRepo.Update(car.ID, map[string]any{
			"status":      string(model.CarStatusAvailable),
			"vehicleType": string(blueprint.VehicleType),
		}); err != nil {
			log.Printf("[seed] car availability %s: %v", plate, err)
			continue
		}
		existingCars = append(existingCars, car)
		created++
	}

	return created
}

func ensureSeedUser(userRepo usersrepo.Repository, seed model.User) (model.User, error) {
	if existing, err := userRepo.FindByEmail(seed.Email); err == nil {
		patch := map[string]any{
			"name":         seed.Name,
			"role":         string(seed.Role),
			"isVerified":   seed.IsVerified,
			"authProvider": string(seed.AuthProvider),
		}
		if seed.PhoneNumber != nil && *seed.PhoneNumber != "" {
			patch["phoneNumber"] = *seed.PhoneNumber
		}
		if seed.ProfilePicture != nil && *seed.ProfilePicture != "" {
			patch["profilePicture"] = *seed.ProfilePicture
		}
		updated, updateErr := userRepo.Update(existing.ID, patch)
		if updateErr != nil {
			return existing, updateErr
		}
		return updated, nil
	}
	return userRepo.Create(seed)
}

func fetchProductionSnapshot(baseURL string) (productionSnapshot, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return productionSnapshot{}, nil
	}

	client := &http.Client{Timeout: productionSeedTimeout}

	users, usersErr := fetchJSONArray(client, baseURL+"/users")
	cars, carsErr := fetchJSONArray(client, baseURL+"/cars")
	bodyGuards, bodyGuardsErr := fetchJSONArray(client, baseURL+"/providers/body-guards")

	if usersErr != nil && carsErr != nil && bodyGuardsErr != nil {
		return productionSnapshot{}, firstNonNil(usersErr, carsErr, bodyGuardsErr)
	}

	snapshot := productionSnapshot{}
	if users != nil {
		snapshot.Users = users
	}
	if cars != nil {
		snapshot.Cars = cars
	}
	if bodyGuards != nil {
		snapshot.BodyGuards = bodyGuards
	}

	return snapshot, firstNonNil(usersErr, carsErr, bodyGuardsErr)
}

func fetchJSONArray(client *http.Client, endpoint string) ([]map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s returned %s: %s", endpoint, resp.Status, strings.TrimSpace(string(body)))
	}

	var out []map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func syncProductionUsers(userRepo usersrepo.Repository, items []map[string]any, password string) int {
	created := 0
	for _, item := range items {
		email := stringField(item, "email")
		if email == "" {
			continue
		}

		role := model.UserRole(stringField(item, "role"))
		if role == "" {
			role = model.UserRoleClient
		}
		authProvider := model.AuthProvider(stringField(item, "authProvider"))
		if authProvider == "" {
			authProvider = model.AuthProviderEmail
		}

		seed := model.User{
			ID:           stringField(item, "id"),
			Email:        email,
			Name:         stringField(item, "name"),
			Role:         role,
			IsVerified:   boolField(item, "isVerified"),
			AuthProvider: authProvider,
		}
		if v := stringPtrField(item, "phoneNumber"); v != nil {
			seed.PhoneNumber = v
		}
		if v := stringPtrField(item, "profilePicture"); v != nil {
			seed.ProfilePicture = v
		}
		seed.Password = &password

		if _, err := ensureSeedUser(userRepo, seed); err != nil {
			log.Printf("[seed] production user %s: %v", email, err)
			continue
		}
		created++
	}
	return created
}

func syncProductionBodyGuards(userRepo usersrepo.Repository, providerRepo providersrepo.Repository, items []map[string]any) int {
	created := 0
	for _, item := range items {
		userID := stringField(item, "userId")
		if userID == "" {
			if user := nestedMap(item, "user"); user != nil {
				userID = stringField(user, "id")
			}
		}
		if userID == "" {
			continue
		}

		user, err := userRepo.FindByID(userID)
		if err != nil {
			continue
		}

		if profile, err := providerRepo.GetMyBodyGuardProfile(userID); err == nil {
			if profileID := stringField(profile, "id"); profileID != "" {
				_ = syncBodyGuardVerification(providerRepo, profileID, item)
			}
			created++
			continue
		}

		dailyRate := floatField(item, "dailyRate")
		dto := providersdto.RegisterBodyGuardDTO{
			Gender:                stringField(item, "gender"),
			NIN:                   firstNonEmpty(stringField(item, "nin"), fmt.Sprintf("SYNC-BG-%s", shortIdentifier(userID))),
			GovernmentIDNumber:    firstNonEmpty(stringField(item, "governmentIdNumber"), fmt.Sprintf("SYNC-GOV-%s", shortIdentifier(userID))),
			SecurityLicenseNumber: stringField(item, "securityLicenseNumber"),
			SecurityLicenseExpiry: stringField(item, "securityLicenseExpiry"),
			IssuingAuthority:      stringField(item, "issuingAuthority"),
			DailyRate:             dailyRate,
			Notes:                 stringField(item, "notes"),
		}
		if dto.Gender == "" {
			dto.Gender = string(model.GenderMale)
		}
		if v := stringField(item, "address"); v != "" {
			dto.Address = v
		}
		if years := intField(item, "yearsOfExperience"); years > 0 {
			dto.YearsOfExperience = &years
		}
		if v := stringField(item, "emergencyContactName"); v != "" {
			dto.EmergencyContactName = v
		}
		if v := stringField(item, "emergencyContactPhone"); v != "" {
			dto.EmergencyContactPhone = v
		}

		if _, err := providerRepo.RegisterBodyGuard(user.ID, dto); err != nil {
			log.Printf("[seed] production bodyguard %s: %v", user.Email, err)
			continue
		}
		if profile, err := providerRepo.GetMyBodyGuardProfile(user.ID); err == nil {
			if profileID := stringField(profile, "id"); profileID != "" {
				_ = syncBodyGuardVerification(providerRepo, profileID, item)
			}
		}
		created++
	}
	return created
}

func syncProductionCars(carRepo carrepo.Repository, userRepo usersrepo.Repository, items []map[string]any, fallbackOwnerID string) int {
	created := 0
	existingCars, err := carRepo.FindAll(carrepo.Query{})
	if err != nil {
		existingCars = []model.CarView{}
	}

	for _, item := range items {
		plate := stringField(item, "licensePlate")
		if plate == "" {
			continue
		}
		if existing, ok := findSeedCarByPlate(existingCars, plate); ok {
			patch := map[string]any{}
			if status := stringField(item, "status"); status != "" {
				patch["status"] = status
			}
			if vehicleType := stringField(item, "vehicleType"); vehicleType != "" {
				patch["vehicleType"] = vehicleType
			}
			if len(patch) > 0 {
				if _, err := carRepo.Update(existing.ID, patch); err != nil {
					log.Printf("[seed] production car update %s: %v", plate, err)
					continue
				}
			}
			created++
			continue
		}

		ownerID := stringField(item, "ownerId")
		if ownerID == "" {
			if owner := nestedMap(item, "owner"); owner != nil {
				ownerID = stringField(owner, "id")
			}
		}
		if ownerID != "" {
			if _, err := userRepo.FindByID(ownerID); err != nil {
				ownerID = fallbackOwnerID
			}
		} else {
			ownerID = fallbackOwnerID
		}

		carDTO := carsdto.CreateCarDTO{
			Make:             stringField(item, "make"),
			Model:            stringField(item, "model"),
			Year:             intField(item, "year"),
			Color:            stringField(item, "color"),
			LicensePlate:     plate,
			Price:            floatField(item, "price"),
			Images:           stringSliceField(item, "images"),
			Description:      stringField(item, "description"),
			Features:         stringSliceField(item, "features"),
			SeatCount:        intField(item, "seatCount"),
			TransmissionType: stringField(item, "transmissionType"),
			FuelType:         stringField(item, "fuelType"),
			VehicleType:      stringField(item, "vehicleType"),
			OwnerID:          ownerID,
		}
		if carDTO.Year == 0 {
			carDTO.Year = time.Now().Year()
		}
		if carDTO.SeatCount == 0 {
			carDTO.SeatCount = 7
		}
		if location := nestedMap(item, "location"); location != nil {
			carDTO.Location = &carsdto.LocationWrapper{
				Create: &carsdto.LocationCreate{
					Latitude:  floatField(location, "latitude"),
					Longitude: floatField(location, "longitude"),
					Address:   stringField(location, "address"),
					City:      stringField(location, "city"),
					State:     stringField(location, "state"),
					Country:   firstNonEmpty(stringField(location, "country"), "Tanzania"),
					ZipCode:   stringField(location, "zipCode"),
				},
			}
		}

		car, err := carRepo.Create(carDTO)
		if err != nil {
			log.Printf("[seed] production car %s: %v", plate, err)
			continue
		}

		patch := map[string]any{}
		if status := stringField(item, "status"); status != "" {
			patch["status"] = status
		} else {
			patch["status"] = string(model.CarStatusAvailable)
		}
		if vehicleType := stringField(item, "vehicleType"); vehicleType != "" {
			patch["vehicleType"] = vehicleType
		}
		if _, err := carRepo.Update(car.ID, patch); err != nil {
			log.Printf("[seed] production car finalize %s: %v", plate, err)
			continue
		}

		existingCars = append(existingCars, car)
		created++
	}

	return created
}

func syncBodyGuardVerification(providerRepo providersrepo.Repository, profileID string, item map[string]any) error {
	patch := providersdto.UpdateVerificationDTO{
		VerificationStatus: stringField(item, "verificationStatus"),
		Notes:              stringField(item, "notes"),
	}
	if v := boolPtrField(item, "backgroundCheckCompleted"); v != nil {
		patch.BackgroundCheckCompleted = v
	}
	if v := boolPtrField(item, "isAvailable"); v != nil {
		patch.IsAvailable = v
	}
	if patch.VerificationStatus == "" {
		patch.VerificationStatus = string(model.VerificationApproved)
	}
	_, err := providerRepo.UpdateBodyGuardVerification(profileID, patch)
	return err
}

func nestedMap(item map[string]any, key string) map[string]any {
	raw, ok := item[key]
	if !ok || raw == nil {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return m
}

func stringField(item map[string]any, key string) string {
	raw, ok := item[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

func stringPtrField(item map[string]any, key string) *string {
	v := stringField(item, key)
	if v == "" {
		return nil
	}
	return &v
}

func boolField(item map[string]any, key string) bool {
	raw, ok := item[key]
	if !ok || raw == nil {
		return false
	}
	switch v := raw.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	default:
		return false
	}
}

func boolPtrField(item map[string]any, key string) *bool {
	raw, ok := item[key]
	if !ok || raw == nil {
		return nil
	}
	v := boolField(item, key)
	return &v
}

func intField(item map[string]any, key string) int {
	raw, ok := item[key]
	if !ok || raw == nil {
		return 0
	}
	switch v := raw.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		return 0
	}
}

func floatField(item map[string]any, key string) float64 {
	raw, ok := item[key]
	if !ok || raw == nil {
		return 0
	}
	switch v := raw.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		n, _ := v.Float64()
		return n
	case string:
		n, _ := strconv.ParseFloat(v, 64)
		return n
	default:
		return 0
	}
}

func stringSliceField(item map[string]any, key string) []string {
	raw, ok := item[key]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return append([]string(nil), v...)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, fmt.Sprint(item))
		}
		return out
	default:
		return nil
	}
}

func firstNonNil(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func shortIdentifier(value string) string {
	if len(value) <= 8 {
		return value
	}
	return value[:8]
}

func approveBodyGuardProfile(providerRepo providersrepo.Repository, profileID string, approved, available bool) error {
	_, err := providerRepo.UpdateBodyGuardVerification(profileID, providersdto.UpdateVerificationDTO{
		VerificationStatus:       string(model.VerificationApproved),
		BackgroundCheckCompleted: &approved,
		IsAvailable:              &available,
		Notes:                    "Seeded demo bodyguard for booking flow",
	})
	return err
}

func findSeedCarByPlate(cars []model.CarView, plate string) (model.CarView, bool) {
	for _, car := range cars {
		if car.LicensePlate == plate {
			return car, true
		}
	}
	return model.CarView{}, false
}

func hashSeedPassword() (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte("PistoniDemo123!"), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
