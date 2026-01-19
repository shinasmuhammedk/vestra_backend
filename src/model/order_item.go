package model

import (

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid" json:"order_id"`
	ProductID uuid.UUID `gorm:"type:uuid" json:"product_id"`
	Size      string    `json:"size"`
	Quantity  int       `json:"quantity"`
	Price     int       `json:"price"`

	// 🔹 This tells GORM the relation
	Product Product `gorm:"foreignKey:ProductID" json:"product"`
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	oi.ID = uuid.New()
	return
}
