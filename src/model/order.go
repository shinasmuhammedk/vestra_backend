package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID   `gorm:"type:uuid" json:"user_id"`
	Total     int         `json:"total"`
	Status    string      `json:"status"`
	Items     []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedAt time.Time   `json:"CreatedAt"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = uuid.New()
	return
}
