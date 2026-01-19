package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wishlist struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID;references:ID" json:"product"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (w *Wishlist) BeforeCreate(tx *gorm.DB) (err error) {
	w.ID = uuid.New()
	return
}
