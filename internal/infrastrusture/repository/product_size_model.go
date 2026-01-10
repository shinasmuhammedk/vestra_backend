package repository

import (
	"time"

)

// type ProductSizeModel struct {
// 	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
// 	ProductID uuid.UUID `gorm:"type:uuid;index;not null"`
// 	Size      string    `gorm:"not null"`
// 	Quantity  int       `gorm:"not null"`

// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// }

type ProductSizeModel struct {
    ID        string `gorm:"primaryKey"`
    ProductID string `gorm:"type:uuid;not null;index:idx_product_size,unique"`
    Size      string `gorm:"not null;index:idx_product_size,unique"`
    Quantity  int
    CreatedAt time.Time
    UpdatedAt time.Time
}