package otp

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// HashOTP returns hashed OTP
func HashOTP(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Compare verifies OTP against hash and expiry
func VerifyOTP(hash, otp string, expiry time.Time) error {
	if time.Now().After(expiry) {
		return errors.New("otp expired")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(otp)); err != nil {
		return errors.New("invalid otp")
	}
	return nil
}
