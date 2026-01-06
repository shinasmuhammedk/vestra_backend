package domain

import (
	"time"

	"github.com/google/uuid"
)

type OTP struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index"` // ✅ FIXED
	Code      string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}
