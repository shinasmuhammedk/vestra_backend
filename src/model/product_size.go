package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductSize struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index:idx_product_size,unique" json:"product_id"`
	Size      string    `gorm:"not null;index:idx_product_size,unique" json:"size"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ps *ProductSize) BeforeCreate(tx *gorm.DB) (err error) {
	ps.ID = uuid.New()
	return
}
