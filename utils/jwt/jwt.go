package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"vestra-ecommerce/utils/logging"
)

// JWTManager manages JWT secrets and token expiry durations
type JWTManager struct {
	AccessSecret  string        // Secret key for access token
	RefreshSecret string        // Secret key for refresh token
	AccessTTL     time.Duration // Access token expiry duration
	RefreshTTL    time.Duration // Refresh token expiry duration
}

// NewJWTManager creates and initializes a JWTManager
func NewJWTManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	logging.Debug.Println("JWTManager initialized")
	return &JWTManager{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		AccessTTL:     accessTTL,
		RefreshTTL:    refreshTTL,
	}
}

// GenerateAccessToken generates a signed JWT access token
func (j *JWTManager) GenerateAccessToken(userID string, role string) (string, error) {
	logging.Debug.Println("Generating access token")

	claims := jwt.MapClaims{
		"user_id": userID,                                 // User identifier
		"role":    role,                                   // User role
		"exp":     time.Now().Add(j.AccessTTL).Unix(),      // Expiry time
		"iat":     time.Now().Unix(),                       // Issued at time
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.AccessSecret))
	if err != nil {
		logging.Error.Println("Failed to sign access token:", err)
		return "", err
	}

	return signedToken, nil
}

// GenerateRefreshToken generates a signed JWT refresh token
func (j *JWTManager) GenerateRefreshToken(userID string, role string) (string, error) {
	logging.Debug.Println("Generating refresh token")

	claims := jwt.MapClaims{
		"user_id": userID,                                 // User identifier
		"role":    role,                                   // User role
		"exp":     time.Now().Add(j.RefreshTTL).Unix(),     // Expiry time
		"iat":     time.Now().Unix(),                       // Issued at time
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.RefreshSecret))
	if err != nil {
		logging.Error.Println("Failed to sign refresh token:", err)
		return "", err
	}

	return signedToken, nil
}

// ValidateRefreshToken validates refresh token and returns claims
func (j *JWTManager) ValidateRefreshToken(tokenStr string) (map[string]interface{}, error) {
	logging.Debug.Println("Validating refresh token")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {

		// Ensure token uses HMAC signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logging.Error.Println("Invalid refresh token signing method")
			return nil, errors.New("invalid signing method")
		}

		return []byte(j.RefreshSecret), nil
	})

	if err != nil {
		logging.Error.Println("Refresh token parsing failed:", err)
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		logging.Error.Println("Invalid refresh token")
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ValidateAccessToken validates access token and returns claims
func (j *JWTManager) ValidateAccessToken(tokenStr string) (map[string]interface{}, error) {
	logging.Debug.Println("Validating access token")

	parsedToken, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {

		// Ensure token uses HMAC signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			logging.Error.Println("Invalid access token signing method")
			return nil, errors.New("unexpected signing method")
		}

		return []byte(j.AccessSecret), nil
	})

	if err != nil || !parsedToken.Valid {
		logging.Error.Println("Invalid access token")
		return nil, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		logging.Error.Println("Invalid access token claims")
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
