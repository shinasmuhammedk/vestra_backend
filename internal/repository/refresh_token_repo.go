package repository

import (
	"time"
	"vestra-ecommerce-backend/internal/domain"

	"gorm.io/gorm"
)

type RefreshTokenRepo struct {
	db *gorm.DB
}

func NewRefreshTokenRepo(db *gorm.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Save(token domain.RefreshToken) error {
	return r.db.Create(&token).Error
}

func (r *RefreshTokenRepo) Find(token string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&rt).Error
	return &rt, err
}
