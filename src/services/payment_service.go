package services

import (
	"time"

	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/utils/apperror"
)

type PaymentService struct {
	repo repo.IPgSQLRepository
}

func NewPaymentService(repo repo.IPgSQLRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

/* =======================
   DTOs
   ======================= */

type VerifyPaymentRequest struct {
	PaymentID     string `json:"payment_id" validate:"required"`
	TransactionID string `json:"transaction_id" validate:"required"`
	Status        string `json:"status" validate:"required"`
}

/* =======================
   CREATE PAYMENT
   ======================= */

func (s *PaymentService) CreatePayment(
	userID string,
	req model.PaymentRequest,
) (*model.Payment, error) {

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
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to create payment",
		)
	}

	return payment, nil
}

/* =======================
   VERIFY PAYMENT
   ======================= */

func (s *PaymentService) VerifyPayment(
	paymentID,
	transactionID,
	status string,
) (*model.Payment, error) {

	var payment model.Payment

	if err := s.repo.FindById(&payment, paymentID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	// Only allow valid statuses
	switch status {
	case constant.PAID, constant.FAILED:
	default:
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
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to verify payment",
		)
	}

	// Reload
	if err := s.repo.FindById(&payment, paymentID); err != nil {
		return nil, apperror.ErrInternal
	}

	return &payment, nil
}

/* =======================
   USER PAYMENTS
   ======================= */

func (s *PaymentService) GetPaymentsByUser(userID string) ([]model.Payment, error) {
	var payments []model.Payment

	if err := s.repo.FindAllWhere(&payments, "user_id = ?", userID); err != nil {
		return nil, apperror.ErrInternal
	}

	return payments, nil
}

func (s *PaymentService) GetPaymentByID(
	userID,
	paymentID string,
) (*model.Payment, error) {

	var payment model.Payment

	if err := s.repo.FindOneWhere(
		&payment,
		"id = ? AND user_id = ?",
		paymentID,
		userID,
	); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	return &payment, nil
}

/* =======================
   CANCEL PAYMENT
   ======================= */

func (s *PaymentService) CancelPayment(
	userID,
	paymentID string,
) (*model.Payment, error) {

	var payment model.Payment

	if err := s.repo.FindOneWhere(
		&payment,
		"id = ? AND user_id = ?",
		paymentID,
		userID,
	); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	if payment.Status != constant.PENDING {
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
		return nil, apperror.ErrInternal
	}

	// Reload
	if err := s.repo.FindById(&payment, paymentID); err != nil {
		return nil, apperror.ErrInternal
	}

	return &payment, nil
}

/* =======================
   ADMIN
   ======================= */

func (s *PaymentService) GetAllPayments() ([]model.Payment, error) {
	var payments []model.Payment

	if err := s.repo.FindAll(&payments); err != nil {
		return nil, apperror.ErrInternal
	}

	return payments, nil
}

func (s *PaymentService) GetPaymentByIDAdmin(
	paymentID string,
) (*model.Payment, error) {

	var payment model.Payment

	if err := s.repo.FindById(&payment, paymentID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	return &payment, nil
}

/* =======================
   ADMIN UPDATE STATUS
   ======================= */

func (s *PaymentService) UpdatePaymentStatus(
	paymentID string,
	status string,
) (*model.Payment, error) {

	var payment model.Payment

	if err := s.repo.FindById(&payment, paymentID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Payment not found",
		)
	}

	switch status {
	case constant.PENDING, constant.PAID, constant.FAILED, constant.CANCELLED:
	default:
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
		return nil, apperror.ErrInternal
	}

	if err := s.repo.FindById(&payment, paymentID); err != nil {
		return nil, apperror.ErrInternal
	}

	return &payment, nil
}
