package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `gorm:"uuid:;primaryKey"`
	Name     string
	Email    string `gorm:"unique"`
	Password string
	Role     string

	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `gorm:"index"`
	IsVerified bool
	IsBlocked  bool `gorm:"default:false"`
}
