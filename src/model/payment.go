package model

import "time"

type Payment struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        string    `json:"user_id"`
	OrderID       string    `json:"order_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`         // pending, completed, failed
	TransactionID string    `json:"transaction_id"` // <--- add this
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PaymentRequest struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"` // e.g., card, upi, paypal
}
