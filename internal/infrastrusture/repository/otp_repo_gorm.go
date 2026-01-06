package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"vestra-ecommerce-backend/internal/domain"
	repo "vestra-ecommerce-backend/internal/repository"
)

type OTPRepoGorm struct {
	db *gorm.DB
}

func NewOTPRepoGorm(db *gorm.DB) repo.OTPRepository {
	return &OTPRepoGorm{db: db}
}

func (r *OTPRepoGorm) Save(otp *domain.OTP) error {
	model := OTPModel{
		ID:        otp.ID,
		UserID:    otp.UserID,
		Code:      otp.Code,
		ExpiresAt: otp.ExpiresAt,
		Used:      otp.Used,
		CreatedAt: otp.CreatedAt,
	}
	return r.db.Create(&model).Error
}

func (r *OTPRepoGorm) FindValidOTP(userID uuid.UUID, code string) (*domain.OTP, error) {
	var model OTPModel

	err := r.db.
		Where(
			"user_id = ? AND code = ? AND used = false AND expires_at > ?",
			userID, code, time.Now(),
		).
		First(&model).Error

	if err != nil {
		return nil, err
	}

	return &domain.OTP{
		ID:        model.ID,
		UserID:    model.UserID,
		Code:      model.Code,
		ExpiresAt: model.ExpiresAt,
		Used:      model.Used,
		CreatedAt: model.CreatedAt,
	}, nil
}

func (r *OTPRepoGorm) MarkUsed(otpID uuid.UUID) error {
	return r.db.
		Model(&OTPModel{}).
		Where("id = ?", otpID).
		Update("used", true).
		Error
}
