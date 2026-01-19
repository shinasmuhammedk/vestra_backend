package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string        `json:"name"`
	Price        int           `json:"price"`
	ImageURL     string        `json:"image_url"`
	League       string        `json:"league"`
	KitType      string        `json:"kit_type"`
	Year         int           `json:"year"`
	IsTopSelling bool          `json:"is_top_selling"`
	IsActive     bool          `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"` // <-- soft delete
	Sizes        []ProductSize  `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate auto-generates UUID
func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}
