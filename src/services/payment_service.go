package services

import (
	"errors"
	"time"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
)

type PaymentService struct {
	repo repo.IPgSQLRepository
}

func NewPaymentService(repo repo.IPgSQLRepository) *PaymentService {
	return &PaymentService{repo: repo}
}



type VerifyPaymentRequest struct {
    PaymentID     string `json:"payment_id" validate:"required"`
    TransactionID string `json:"transaction_id" validate:"required"`
    Status        string `json:"status" validate:"required"` // e.g., "success" or "failed"
}



// CreatePayment inserts a new payment record
func (s *PaymentService) CreatePayment(userID string, req model.PaymentRequest) (*model.Payment, error) {
	payment := &model.Payment{
		UserID:        userID,
		OrderID:       req.OrderID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Status:        "pending", // initially pending
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Use your generic repo Insert method
	if err := s.repo.Insert(payment); err != nil {
		return nil, err
	}

	return payment, nil
}


// VerifyPayment updates the payment status
func (s *PaymentService) VerifyPayment(paymentID, transactionID, status string) (*model.Payment, error) {
    payment := &model.Payment{}

    // Find the payment by ID
    if err := s.repo.FindById(payment, paymentID); err != nil {
        return nil, errors.New("payment not found")
    }

    // Update the status and transaction ID
    update := map[string]interface{}{
        "status":         status,
        "transaction_id": transactionID,
    }

    if err := s.repo.UpdateByFields(payment, paymentID, update); err != nil {
        return nil, err
    }

    return payment, nil
}



// GetPaymentsByUser fetches all payments for a given user
func (s *PaymentService) GetPaymentsByUser(userID string) ([]model.Payment, error) {
	var payments []model.Payment

	// Use the generic repo method FindAllWhere
	if err := s.repo.FindAllWhere(&payments, "user_id = ?", userID); err != nil {
		return nil, err
	}

	return payments, nil
}



// GetPaymentByID fetches a single payment by ID for a user
func (s *PaymentService) GetPaymentByID(userID, paymentID string) (*model.Payment, error) {
	var payment model.Payment

	err := s.repo.FindOneWhere(
		&payment,
		"id = ? AND user_id = ?",
		paymentID,
		userID,
	)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	return &payment, nil
}



// CancelPayment cancels a pending payment
func (s *PaymentService) CancelPayment(userID, paymentID string) (*model.Payment, error) {
	var payment model.Payment

	// Ensure payment exists & belongs to user
	err := s.repo.FindOneWhere(
		&payment,
		"id = ? AND user_id = ?",
		paymentID,
		userID,
	)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	// Only pending payments can be cancelled
	if payment.Status != "pending" {
		return nil, errors.New("only pending payments can be cancelled")
	}

	update := map[string]interface{}{
		"status":     "cancelled",
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateByFields(&payment, paymentID, update); err != nil {
		return nil, err
	}

	// Reflect updated status in response
	payment.Status = "cancelled"
	payment.UpdatedAt = time.Now()

	return &payment, nil
}



// GetAllPayments fetches all payments (admin)
func (s *PaymentService) GetAllPayments() ([]model.Payment, error) {
	var payments []model.Payment

	if err := s.repo.FindAll(&payments); err != nil {
		return nil, err
	}

	return payments, nil
}



// GetPaymentByIDAdmin fetches a payment by ID (admin)
func (s *PaymentService) GetPaymentByIDAdmin(paymentID string) (*model.Payment, error) {
	var payment model.Payment

	if err := s.repo.FindById(&payment, paymentID); err != nil {
		return nil, errors.New("payment not found")
	}

	return &payment, nil
}
