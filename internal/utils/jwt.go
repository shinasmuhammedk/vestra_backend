package utils

import (
    "errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// var accessSecret = []byte("access-secret")
// var refreshSecret = []byte("refresh-secret")

// 🔑 Access Token
func GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(AccessTokenSecret)
}

// 🔄 Refresh Token
func GenerateRefreshToken(userID uuid.UUID) (string, time.Time, error) {
	expiry := time.Now().Add(7 * 24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     expiry.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(RefreshTokenSecret)

	return tokenStr, expiry, err
}

// ✅ THIS IS RefreshClaims
type RefreshClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&RefreshClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return RefreshTokenSecret, nil
		},
	)

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
