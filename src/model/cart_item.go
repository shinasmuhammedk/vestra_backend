package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CartID    uuid.UUID `gorm:"type:uuid;index"`
	ProductID uuid.UUID `gorm:"type:uuid;index"`
	Size      string
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
	Product   Product `gorm:"foreignKey:ProductID"`
}

func (ci *CartItem) BeforeCreate(tx *gorm.DB) (err error) {
	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}
	return
}
