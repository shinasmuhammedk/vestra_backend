package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserModel maps to DB
type UserModel struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid"`
	Name       string
	Email      string `gorm:"uniqueIndex"`
	Password   string
	Role       string
	IsVerified bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}
