package services

import (
	"time"

	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/utils/apperror"
)

type PaymentService struct {
	repo repo.IPgSQLRepository
}

func NewPaymentService(repo repo.IPgSQLRepository) *PaymentService {
	logging.Debug.Println("PaymentService initialized")
	return &PaymentService{repo: repo}
}

type VerifyPaymentRequest struct {
	PaymentID     string `json:"payment_id" validate:"required"`
	TransactionID string `json:"transaction_id" validate:"required"`
	Status        string `json:"status" validate:"required"`
}

// CreatePayment creates new payment
func (s *PaymentService) CreatePayment(
	userID string,
	req model.PaymentRequest,
) (*model.Payment, error) {
	logging.Debug.Printf("Creating payment for user: %s, order: %s", userID, req.OrderID)

	payment := &model.Payment{
		UserID:        userID,
		OrderID:       req.OrderID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Status:        constant.PENDING,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Insert(payment); err != nil {
		logging.Error.Printf("Failed to create payment for order %s: %v", req.OrderID, err)
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to create payment",
		)
	}

	logging.Debug.Printf("Payment created: %s for order: %s", payment.ID, req.OrderID)
	return payment, nil
}

// VerifyPayment updates payment status
func (s *PaymentService) VerifyPayment(
	paymentID,
	transactionID,
	status string,
) (*model.Payment, error) {
	logging.Debug.Printf("Verifying payment: %s, status: %s", paymentID, status)

	var payment model.Payment

	if err := s.repo.FindById(&payment, paymentID); err != nil {
		logging.Error.Printf("Payment not found: %s", paymentID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	// Validate status
	switch status {
	case constant.PAID, constant.FAILED:
	default:
		logging.Error.Printf("Invalid payment status: %s", status)
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid payment status",
		)
	}

	updates := map[string]interface{}{
		"status":         status,
		"transaction_id": transactionID,
		"updated_at":     time.Now(),
	}

	if err := s.repo.UpdateByFields(&model.Payment{}, paymentID, updates); err != nil {
		logging.Error.Printf("Failed to verify payment %s: %v", paymentID, err)
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to verify payment",
		)
	}

	// Reload updated payment
	if err := s.repo.FindById(&payment, paymentID); err != nil {
		logging.Error.Printf("Failed to reload payment %s: %v", paymentID, err)
		return nil, apperror.ErrInternal
	}

	logging.Debug.Printf("Payment verified: %s, status: %s", paymentID, status)
	return &payment, nil
}

// GetPaymentsByUser retrieves user's payments
func (s *PaymentService) GetPaymentsByUser(userID string) ([]model.Payment, error) {
	logging.Debug.Printf("Getting payments for user: %s", userID)

	var payments []model.Payment

	if err := s.repo.FindAllWhere(&payments, "user_id = ?", userID); err != nil {
		logging.Error.Printf("Failed to get payments for user %s: %v", userID, err)
		return nil, apperror.ErrInternal
	}

	logging.Debug.Printf("Found %d payments for user: %s", len(payments), userID)
	return payments, nil
}

// GetPaymentByID gets payment for specific user
func (s *PaymentService) GetPaymentByID(
	userID,
	paymentID string,
) (*model.Payment, error) {
	logging.Debug.Printf("Getting payment %s for user: %s", paymentID, userID)

	var payment model.Payment

	if err := s.repo.FindOneWhere(
		&payment,
		"id = ? AND user_id = ?",
		paymentID,
		userID,
	); err != nil {
		logging.Error.Printf("Payment not found: %s for user: %s", paymentID, userID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	logging.Debug.Printf("Payment found: %s", paymentID)
	return &payment, nil
}

// CancelPayment cancels user's payment
func (s *PaymentService) CancelPayment(
	userID,
	paymentID string,
) (*model.Payment, error) {
	logging.Debug.Printf("Cancelling payment %s for user: %s", paymentID, userID)

	var payment model.Payment

	if err := s.repo.FindOneWhere(
		&payment,
		"id = ? AND user_id = ?",
		paymentID,
		userID,
	); err != nil {
		logging.Error.Printf("Payment to cancel not found: %s", paymentID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	if payment.Status != constant.PENDING {
		logging.Error.Printf("Cannot cancel non-pending payment: %s, status: %s", paymentID, payment.Status)
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Only pending payments can be cancelled",
		)
	}

	updates := map[string]interface{}{
		"status":     constant.CANCELLED,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateByFields(&model.Payment{}, paymentID, updates); err != nil {
		logging.Error.Printf("Failed to cancel payment %s: %v", paymentID, err)
		return nil, apperror.ErrInternal
	}

	// Reload
	if err := s.repo.FindById(&payment, paymentID); err != nil {
		logging.Error.Printf("Failed to reload cancelled payment %s: %v", paymentID, err)
		return nil, apperror.ErrInternal
	}

	logging.Debug.Printf("Payment cancelled: %s", paymentID)
	return &payment, nil
}

// GetAllPayments gets all payments (admin)
func (s *PaymentService) GetAllPayments() ([]model.Payment, error) {
	logging.Debug.Println("Getting all payments (admin)")

	var payments []model.Payment

	if err := s.repo.FindAll(&payments); err != nil {
		logging.Error.Printf("Failed to get all payments: %v", err)
		return nil, apperror.ErrInternal
	}

	logging.Debug.Printf("Found %d total payments", len(payments))
	return payments, nil
}

// GetPaymentByIDAdmin gets payment by ID (admin)
func (s *PaymentService) GetPaymentByIDAdmin(
	paymentID string,
) (*model.Payment, error) {
	logging.Debug.Printf("Admin getting payment: %s", paymentID)

	var payment model.Payment

	if err := s.repo.FindById(&payment, paymentID); err != nil {
		logging.Error.Printf("Payment not found (admin): %s", paymentID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	logging.Debug.Printf("Payment found (admin): %s", paymentID)
	return &payment, nil
}

// UpdatePaymentStatus updates payment status (admin)
func (s *PaymentService) UpdatePaymentStatus(
	paymentID string,
	status string,
) (*model.Payment, error) {
	logging.Debug.Printf("Admin updating payment %s to status: %s", paymentID, status)

	var payment model.Payment

	if err := s.repo.FindById(&payment, paymentID); err != nil {
		logging.Error.Printf("Payment not found for status update: %s", paymentID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	// Validate status
	switch status {
	case constant.PENDING, constant.PAID, constant.FAILED, constant.CANCELLED:
	default:
		logging.Error.Printf("Invalid status for update: %s", status)
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid payment status",
		)
	}

	if err := s.repo.UpdateByFields(
		&model.Payment{},
		paymentID,
		map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		},
	); err != nil {
		logging.Error.Printf("Failed to update payment status %s: %v", paymentID, err)
		return nil, apperror.ErrInternal
	}

	if err := s.repo.FindById(&payment, paymentID); err != nil {
		logging.Error.Printf("Failed to reload payment after status update %s: %v", paymentID, err)
		return nil, apperror.ErrInternal
	}

	logging.Debug.Printf("Payment status updated: %s to %s", paymentID, status)
	return &payment, nil
}