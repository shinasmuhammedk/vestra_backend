package repository

import (
	"time"
)

// ProductModel represents the "products" table in the database
type ProductModel struct {
	ID           string             `gorm:"primaryKey;type:uuid"`
	Name         string             `gorm:"not null"`
	Price        int                `gorm:"not null"`
	ImageURL     string             `gorm:"column:image_url"`
	League       string             `gorm:"column:league"`
	KitType      string             `gorm:"column:kit_type"`
	Year         int                `gorm:"column:year"`
	IsTopSelling bool               `gorm:"column:is_top_selling;default:false"`
	IsActive     bool               `gorm:"column:is_active;default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// One-to-Many relationship with ProductSizeModel
	Sizes        []ProductSizeModel `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
}

// TableName explicitly tells GORM to use the "products" table
func (ProductModel) TableName() string {
	return "products"
}