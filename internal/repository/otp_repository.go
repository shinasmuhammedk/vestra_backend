package repository

import (
	"github.com/google/uuid"
	"vestra-ecommerce-backend/internal/domain"
)

type OTPRepository interface {
	Save(otp *domain.OTP) error
	FindValidOTP(userID uuid.UUID, code string) (*domain.OTP, error)
	MarkUsed(otpID uuid.UUID) error
}
