package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	"pistoni/core/internal/modules/auth/dto"
	authrepo "pistoni/core/internal/modules/auth/repo"
	usersrepo "pistoni/core/internal/modules/users/repo"
	"pistoni/core/internal/platform/apperr"
	"pistoni/core/internal/platform/auth"
	"pistoni/core/internal/platform/model"
)

type AuthResponse struct {
	User         model.User `json:"user"`
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
}

type EmailSender interface {
	SendCodeEmail(email, code, subject, intro string) error
}

type NoopEmailSender struct{}

func (NoopEmailSender) SendCodeEmail(email, code, subject, intro string) error { return nil }

type Principal struct{ User model.User }

func (p Principal) GetRole() string { return string(p.User.Role) }

type Service struct {
	users       usersrepo.Repository
	pendingRepo authrepo.PendingRepository
	jwt         *auth.JWTManager
	emailer     EmailSender
}

func New(users usersrepo.Repository, pending authrepo.PendingRepository, jwt *auth.JWTManager, emailer EmailSender) *Service {
	if emailer == nil {
		emailer = NoopEmailSender{}
	}
	return &Service{users: users, pendingRepo: pending, jwt: jwt, emailer: emailer}
}

func (s *Service) UserForAuth(userID string) (Principal, error) {
	u, err := s.users.FindByID(userID)
	if err != nil {
		return Principal{}, err
	}
	return Principal{User: u}, nil
}

func (s *Service) Login(in dto.LoginDTO) (AuthResponse, error) {
	u, err := s.users.FindByEmail(in.Email)
	if err != nil {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Invalid credentials")
	}
	if !u.IsVerified {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Please verify your email before logging in")
	}
	if u.Password == nil {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "This account uses OAuth authentication. Please use Google Sign In.")
	}
	if bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(in.Password)) != nil {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Invalid credentials")
	}
	return s.issueTokens(u)
}

func (s *Service) Register(in dto.RegisterDTO) (AuthResponse, error) {
	if _, err := s.users.FindByEmail(in.Email); err == nil {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Email already registered")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	role := model.UserRoleClient
	if in.Role != "" {
		role = model.UserRole(strings.ToUpper(in.Role))
	}
	u := model.User{Email: in.Email, Password: ptr(string(hash)), Name: in.Name, Role: role, AuthProvider: model.AuthProviderEmail, IsVerified: false}
	if in.PhoneNumber != "" {
		u.PhoneNumber = &in.PhoneNumber
	}
	if in.ProfilePicture != "" {
		u.ProfilePicture = &in.ProfilePicture
	}
	code := verificationCode()
	expiry := time.Now().Add(15 * time.Minute)
	u.VerificationCode, u.CodeType, u.CodeExpiry = &code, ptr("register"), &expiry
	created, err := s.users.Create(u)
	if err != nil {
		return AuthResponse{}, err
	}
	if err := s.emailer.SendCodeEmail(created.Email, code, "Verify Your Email - Pistoni", "Here is your registration verification code:"); err != nil {
		log.Printf("[auth-email] register send failed email=%s err=%v", created.Email, err)
	} else {
		log.Printf("[auth-email] register send ok email=%s", created.Email)
	}
	return s.issueTokens(created)
}

func (s *Service) Refresh(refreshToken string) (AuthResponse, error) {
	users, _ := s.users.FindAll(0, 0, nil)
	for _, u := range users {
		if u.RefreshToken != nil && *u.RefreshToken == refreshToken && u.RefreshTokenExpiry != nil && u.RefreshTokenExpiry.After(time.Now()) {
			return s.issueTokens(u)
		}
	}
	return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Invalid or expired refresh token")
}

func (s *Service) Logout(userID string) error {
	_, err := s.users.Update(userID, map[string]any{"refreshToken": (*string)(nil), "refreshTokenExpiry": (*time.Time)(nil)})
	return err
}

func (s *Service) RequestEmailVerification(email string) error {
	_ = s.pendingRepo.DeleteExpired(email, time.Now())
	if _, err := s.users.FindByEmail(email); err == nil {
		return apperr.New(http.StatusBadRequest, "Email already registered")
	}
	code := verificationCode()
	expiry := time.Now().Add(15 * time.Minute)
	_ = s.pendingRepo.Upsert(email, code, expiry)
	if err := s.emailer.SendCodeEmail(email, code, "Verify Your Email - Pistoni", "Here is your registration verification code:"); err != nil {
		log.Printf("[auth-email] check-email send failed email=%s err=%v", email, err)
		return err
	}
	log.Printf("[auth-email] check-email send ok email=%s", email)
	return nil
}

func (s *Service) VerifyEmailCode(email, code string) error {
	p, ok := s.pendingRepo.FindByEmail(email)
	if !ok || p.Code != code || p.CodeExpiry.Before(time.Now()) {
		return apperr.New(http.StatusBadRequest, "Invalid or expired verification code.")
	}
	return nil
}

func (s *Service) CompleteRegistration(in dto.CompleteRegistrationDTO) (AuthResponse, error) {
	p, ok := s.pendingRepo.FindByEmail(in.Email)
	if !ok || p.CodeExpiry.Before(time.Now()) {
		return AuthResponse{}, apperr.New(http.StatusBadRequest, "Email not verified or verification expired.")
	}
	if _, err := s.users.FindByEmail(in.Email); err == nil {
		return AuthResponse{}, apperr.New(http.StatusBadRequest, "Email already registered")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	role := model.UserRoleClient
	if in.Role != "" {
		role = model.UserRole(strings.ToUpper(in.Role))
	}
	u := model.User{Email: in.Email, Password: ptr(string(hash)), Name: in.Name, Role: role, AuthProvider: model.AuthProviderEmail, IsVerified: true}
	if in.PhoneNumber != "" {
		u.PhoneNumber = &in.PhoneNumber
	}
	created, err := s.users.Create(u)
	if err != nil {
		return AuthResponse{}, err
	}
	_ = s.pendingRepo.Delete(in.Email)
	return s.issueTokens(created)
}

func (s *Service) VerifyRegistrationCode(email, code string) error {
	u, err := s.users.FindByEmail(email)
	if err != nil || u.IsVerified || u.VerificationCode == nil || u.CodeExpiry == nil || *u.VerificationCode != code || u.CodeType == nil || *u.CodeType != "register" || u.CodeExpiry.Before(time.Now()) {
		return apperr.New(http.StatusBadRequest, "Invalid or expired verification code.")
	}
	_, err = s.users.Update(u.ID, map[string]any{"isVerified": true})
	return err
}

func (s *Service) VerifyEmailToken(token string) error {
	users, _ := s.users.FindAll(0, 0, nil)
	for _, u := range users {
		if u.VerificationToken != nil && *u.VerificationToken == token && u.TokenExpiry != nil && u.TokenExpiry.After(time.Now()) {
			_, err := s.users.Update(u.ID, map[string]any{"isVerified": true, "verificationToken": (*string)(nil)})
			return err
		}
	}
	return apperr.New(http.StatusBadRequest, "Invalid or expired verification token")
}

func (s *Service) UpdateProfile(userID string, in dto.UpdateProfileDTO) (model.User, error) {
	patch := map[string]any{}
	if in.Name != nil {
		patch["name"] = *in.Name
	}
	if in.Email != nil {
		patch["email"] = *in.Email
	}
	if in.PhoneNumber != nil {
		patch["phoneNumber"] = *in.PhoneNumber
	}
	if in.ProfilePicture != nil {
		patch["profilePicture"] = *in.ProfilePicture
	}
	return s.users.Update(userID, patch)
}

func (s *Service) ChangePassword(userID, currentPassword, newPassword string) error {
	u, err := s.users.FindByID(userID)
	if err != nil {
		return apperr.New(http.StatusUnauthorized, "User not found")
	}
	if u.Password == nil {
		return apperr.New(http.StatusUnauthorized, "Cannot change password for OAuth accounts")
	}
	if bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(currentPassword)) != nil {
		return apperr.New(http.StatusUnauthorized, "Current password is incorrect")
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	_, err = s.users.Update(userID, map[string]any{"password": ptr(string(hash))})
	return err
}

func (s *Service) GoogleOAuth(in dto.GoogleOAuthDTO) (AuthResponse, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Google OAuth not configured")
	}

	payload, err := idtoken.Validate(context.Background(), in.IDToken, clientID)
	if err != nil {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Invalid Google token")
	}

	emailClaim, _ := payload.Claims["email"].(string)
	if emailClaim == "" || !strings.EqualFold(emailClaim, in.Email) {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Token email does not match provided email")
	}

	u, err := s.users.FindByEmail(in.Email)
	if err != nil {
		u = model.User{Email: in.Email, Name: in.Name, AuthProvider: model.AuthProviderGoogle, IsVerified: true, Role: model.UserRoleClient}
		if in.ProfilePicture != "" {
			u.ProfilePicture = &in.ProfilePicture
		}
		u, err = s.users.Create(u)
		if err != nil {
			return AuthResponse{}, err
		}
	} else {
		patch := map[string]any{"authProvider": string(model.AuthProviderGoogle), "isVerified": true}
		if in.ProfilePicture != "" {
			patch["profilePicture"] = in.ProfilePicture
		}
		_, err = s.users.Update(u.ID, patch)
		if err != nil {
			return AuthResponse{}, err
		}
		u, _ = s.users.FindByID(u.ID)
	}
	return s.issueTokens(u)
}

func (s *Service) AppleOAuth(in dto.AppleOAuthDTO) (AuthResponse, error) {
	if in.IdentityToken == "" || in.UserIdentifier == "" {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Invalid Apple Sign-In data")
	}

	claims, err := parseAppleIdentityTokenClaims(in.IdentityToken)
	if err != nil {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Invalid Apple identity token")
	}

	if claims.Iss != "https://appleid.apple.com" {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Invalid Apple token issuer")
	}
	if claims.Sub == "" || claims.Sub != in.UserIdentifier {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Invalid Apple user identifier")
	}
	if appleClientID := os.Getenv("APPLE_CLIENT_ID"); appleClientID != "" && claims.Aud != "" && claims.Aud != appleClientID {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Invalid Apple token audience")
	}
	if claims.Exp <= time.Now().Unix() {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Expired Apple identity token")
	}

	email := strings.TrimSpace(in.Email)
	if email == "" {
		email = strings.TrimSpace(claims.Email)
	}

	// First lookup by stored Apple userIdentifier (kept in verificationToken for parity with old NestJS behavior).
	users, _ := s.users.FindAll(0, 0, nil)
	var existing *model.User
	for _, candidate := range users {
		if candidate.AuthProvider == model.AuthProviderApple && candidate.VerificationToken != nil && *candidate.VerificationToken == in.UserIdentifier {
			c := candidate
			existing = &c
			break
		}
	}

	if existing == nil {
		if email == "" {
			return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Unable to authenticate: email required for new Apple Sign-In users")
		}
		u, err := s.users.FindByEmail(email)
		if err == nil {
			existing = &u
		}
	}

	if existing != nil {
		patch := map[string]any{"authProvider": string(model.AuthProviderApple), "isVerified": true, "verificationToken": &in.UserIdentifier}
		_, err = s.users.Update(existing.ID, patch)
		if err != nil {
			return AuthResponse{}, err
		}
		u, _ := s.users.FindByID(existing.ID)
		return s.issueTokens(u)
	}

	if email == "" {
		return AuthResponse{}, apperr.New(http.StatusUnauthorized, "Email is required for new Apple Sign-In users")
	}
	name := strings.TrimSpace(in.FullName.GivenName + " " + in.FullName.FamilyName)
	if name == "" {
		name = strings.Split(email, "@")[0]
	}
	u := model.User{Email: email, Name: name, AuthProvider: model.AuthProviderApple, IsVerified: true, Role: model.UserRoleClient, VerificationToken: &in.UserIdentifier}
	u, err = s.users.Create(u)
	if err != nil {
		return AuthResponse{}, err
	}
	return s.issueTokens(u)
}

func (s *Service) issueTokens(u model.User) (AuthResponse, error) {
	access, err := s.jwt.GenerateAccessToken(u.Email, u.ID, string(u.Role))
	if err != nil {
		return AuthResponse{}, err
	}
	refresh := randomToken(40)
	expiry := time.Now().AddDate(1, 0, 0)
	updated, err := s.users.Update(u.ID, map[string]any{"refreshToken": &refresh, "refreshTokenExpiry": &expiry})
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{User: updated, AccessToken: access, RefreshToken: refresh}, nil
}

func verificationCode() string {
	return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type appleIdentityTokenClaims struct {
	Iss   string `json:"iss"`
	Aud   string `json:"aud"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

func parseAppleIdentityTokenClaims(identityToken string) (appleIdentityTokenClaims, error) {
	parts := strings.Split(identityToken, ".")
	if len(parts) != 3 {
		return appleIdentityTokenClaims{}, fmt.Errorf("invalid jwt format")
	}
	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return appleIdentityTokenClaims{}, err
	}
	var claims appleIdentityTokenClaims
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return appleIdentityTokenClaims{}, err
	}
	return claims, nil
}

func ptr[T any](v T) *T { return &v }
