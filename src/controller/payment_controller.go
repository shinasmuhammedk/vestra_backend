package controller

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/response"

	"github.com/gofiber/fiber/v2"
)

type PaymentController struct {
	service *services.PaymentService
}

func NewPaymentController(service *services.PaymentService) *PaymentController {
	logging.Debug.Println("PaymentController initialized")
	return &PaymentController{service: service}
}

// CreatePayment creates a new payment for user
func (pc *PaymentController) CreatePayment(c *fiber.Ctx) error {
	logging.Debug.Println("CreatePayment endpoint called")

	userID := c.Locals("user_id").(string)
	logging.Debug.Printf("Creating payment for user: %s", userID)

	var req model.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("CreatePayment - Invalid request body: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Payment request - Order: %s, Amount: %d, Method: %s", 
		req.OrderID, req.Amount, req.PaymentMethod)

	payment, err := pc.service.CreatePayment(userID, req)
	if err != nil {
		logging.Error.Printf("CreatePayment failed: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to create payment",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Payment created: %s", payment.ID)
	return response.Success(
		c,
		constant.CREATED,
		"Payment created successfully",
		"",
		payment,
	)
}

// VerifyPayment verifies payment status
func (pc *PaymentController) VerifyPayment(c *fiber.Ctx) error {
	logging.Debug.Println("VerifyPayment endpoint called")

	var req services.VerifyPaymentRequest

	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("VerifyPayment - Invalid request body: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Verifying payment: %s, Status: %s", req.PaymentID, req.Status)

	payment, err := pc.service.VerifyPayment(req.PaymentID, req.TransactionID, req.Status)
	if err != nil {
		logging.Error.Printf("VerifyPayment failed: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			err.Error(),
			"",
			nil,
		)
	}

	logging.Debug.Printf("Payment verified: %s, New status: %s", payment.ID, payment.Status)
	return response.Success(
		c,
		constant.SUCCESS,
		"Payment verified successfully",
		"",
		payment,
	)
}

// GetUserPayments fetches all payments for logged-in user
func (pc *PaymentController) GetUserPayments(c *fiber.Ctx) error {
	logging.Debug.Println("GetUserPayments endpoint called")

	userID := c.Locals("user_id").(string)
	logging.Debug.Printf("Fetching payments for user: %s", userID)

	payments, err := pc.service.GetPaymentsByUser(userID)
	if err != nil {
		logging.Error.Printf("GetUserPayments failed: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch payments",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Found %d payments for user: %s", len(payments), userID)
	return response.Success(
		c,
		constant.SUCCESS,
		"Payments fetched successfully",
		"",
		payments,
	)
}

// GetUserPaymentByID fetches specific payment for user
func (pc *PaymentController) GetUserPaymentByID(c *fiber.Ctx) error {
	logging.Debug.Println("GetUserPaymentByID endpoint called")

	userID := c.Locals("user_id").(string)
	paymentID := c.Params("id")
	logging.Debug.Printf("Fetching payment %s for user: %s", paymentID, userID)

	payment, err := pc.service.GetPaymentByID(userID, paymentID)
	if err != nil {
		logging.Error.Printf("GetUserPaymentByID failed: %v", err)
		return response.Error(
			c,
			constant.NOTFOUND,
			err.Error(),
			"",
			nil,
		)
	}

	logging.Debug.Printf("Payment found: %s", payment.ID)
	return response.Success(
		c,
		constant.SUCCESS,
		"Payment fetched successfully",
		"",
		payment,
	)
}

// CancelPayment cancels user's payment
func (pc *PaymentController) CancelPayment(c *fiber.Ctx) error {
	logging.Debug.Println("CancelPayment endpoint called")

	userID := c.Locals("user_id").(string)
	paymentID := c.Params("id")
	logging.Debug.Printf("Cancelling payment %s for user: %s", paymentID, userID)

	payment, err := pc.service.CancelPayment(userID, paymentID)
	if err != nil {
		logging.Error.Printf("CancelPayment failed: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			err.Error(),
			"",
			nil,
		)
	}

	logging.Debug.Printf("Payment cancelled: %s", payment.ID)
	return response.Success(
		c,
		constant.SUCCESS,
		"Payment cancelled successfully",
		"",
		payment,
	)
}

// GetPaymentByIDAdmin fetches payment for admin
func (pc *PaymentController) GetPaymentByIDAdmin(c *fiber.Ctx) error {
	logging.Debug.Println("GetPaymentByIDAdmin endpoint called")

	paymentID := c.Params("id")
	logging.Debug.Printf("Admin fetching payment: %s", paymentID)

	payment, err := pc.service.GetPaymentByIDAdmin(paymentID)
	if err != nil {
		logging.Error.Printf("GetPaymentByIDAdmin failed: %v", err)
		return response.Error(
			c,
			constant.NOTFOUND,
			err.Error(),
			"",
			nil,
		)
	}

	logging.Debug.Printf("Payment fetched by admin: %s", payment.ID)
	return response.Success(
		c,
		constant.SUCCESS,
		"Payment fetched successfully",
		"",
		payment,
	)
}

// UpdatePaymentStatus updates payment status (admin only)
func (pc *PaymentController) UpdatePaymentStatus(c *fiber.Ctx) error {
	logging.Debug.Println("UpdatePaymentStatus endpoint called")

	paymentID := c.Params("id")
	logging.Debug.Printf("Updating status for payment: %s", paymentID)

	var req struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("UpdatePaymentStatus - Invalid request body: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Updating payment %s to status: %s", paymentID, req.Status)

	payment, err := pc.service.UpdatePaymentStatus(paymentID, req.Status)
	if err != nil {
		logging.Error.Printf("UpdatePaymentStatus failed: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			err.Error(),
			"",
			nil,
		)
	}

	logging.Debug.Printf("Payment status updated: %s -> %s", paymentID, payment.Status)
	return response.Success(
		c,
		constant.SUCCESS,
		"Payment status updated successfully",
		"",
		payment,
	)
}