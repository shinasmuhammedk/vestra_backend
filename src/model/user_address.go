package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserAddress struct {
	ID        string    `gorm:"type:uuid;primary_key;" json:"id"`
	UserID    string    `gorm:"type:uuid;not null" json:"user_id"`
	Line1     string    `json:"line_1"`
	Line2     string    `json:"line_2"`
	City      string    `json:"city"`
	State     string    `json:"state"`
	Country   string    `json:"country"`
	ZipCode   string    `json:"zip_code"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *UserAddress) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.NewString()
	return
}
