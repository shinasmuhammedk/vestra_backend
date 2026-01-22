package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductSize struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index:idx_product_size,unique"`
	Size      string    `gorm:"not null;index:idx_product_size,unique"`
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ps *ProductSize) BeforeCreate(tx *gorm.DB) (err error) {
	ps.ID = uuid.New()
	return
}
