package domain

import (
	"time"

)

type ProductSize struct {
	ID        string `gorm:"primaryKey"`
	ProductID string `gorm:"type:uuid;not null;index:idx_product_size,unique"`
	Size      string    `gorm:"not null;index:idx_product_size,unique"` // unique with ProductID
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}
