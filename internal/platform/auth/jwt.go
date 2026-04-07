package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Email string `json:"email"`
	Sub   string `json:"sub"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret string
}

func NewJWTManager(secret string) *JWTManager {
	if secret == "" {
		secret = "your-secret-key"
	}
	return &JWTManager{secret: secret}
}

func (m *JWTManager) GenerateAccessToken(email, userID, role string) (string, error) {
	claims := Claims{
		Email: email,
		Sub:   userID,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.secret))
}

func (m *JWTManager) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}
