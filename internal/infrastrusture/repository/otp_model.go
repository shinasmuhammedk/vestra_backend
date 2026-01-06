package repository

import (
	"time"

	"github.com/google/uuid"
)

type OTPModel struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Code      string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}
