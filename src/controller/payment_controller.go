package controller

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/response"

	"github.com/gofiber/fiber/v2"
)

type PaymentController struct {
	service *services.PaymentService
}

func NewPaymentController(service *services.PaymentService) *PaymentController {
	return &PaymentController{service: service}
}

// POST /user/payment
func (pc *PaymentController) CreatePayment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string) // get logged-in user

	var req model.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	payment, err := pc.service.CreatePayment(userID, req)
	if err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to create payment",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.CREATED,
		"Payment created successfully",
		"",
		payment,
	)
}

// POST /user/payment/verify
func (pc *PaymentController) VerifyPayment(c *fiber.Ctx) error {
	var req services.VerifyPaymentRequest

	if err := c.BodyParser(&req); err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	payment, err := pc.service.VerifyPayment(req.PaymentID, req.TransactionID, req.Status)
	if err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			err.Error(),
			"",
			nil,
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Payment verified successfully",
		"",
		payment,
	)
}

// GET /user/payment
func (pc *PaymentController) GetUserPayments(c *fiber.Ctx) error {
	// Get logged-in user ID from context
	userID := c.Locals("user_id").(string)

	payments, err := pc.service.GetPaymentsByUser(userID)
	if err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch payments",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Payments fetched successfully",
		"",
		payments,
	)
}

// GET /user/payment/:id
func (pc *PaymentController) GetUserPaymentByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	paymentID := c.Params("id")

	payment, err := pc.service.GetPaymentByID(userID, paymentID)
	if err != nil {
		return response.Error(
			c,
			constant.NOTFOUND,
			err.Error(),
			"",
			nil,
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Payment fetched successfully",
		"",
		payment,
	)
}

// PUT /user/payment/:id/cancel
func (pc *PaymentController) CancelPayment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	paymentID := c.Params("id")

	payment, err := pc.service.CancelPayment(userID, paymentID)
	if err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			err.Error(),
			"",
			nil,
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Payment cancelled successfully",
		"",
		payment,
	)
}

// GET /admin/payments/:id
func (pc *PaymentController) GetPaymentByIDAdmin(c *fiber.Ctx) error {
	paymentID := c.Params("id")

	payment, err := pc.service.GetPaymentByIDAdmin(paymentID)
	if err != nil {
		return response.Error(
			c,
			constant.NOTFOUND,
			err.Error(),
			"",
			nil,
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Payment fetched successfully",
		"",
		payment,
	)
}

// PUT /admin/payments/:id/status
func (pc *PaymentController) UpdatePaymentStatus(c *fiber.Ctx) error {
	paymentID := c.Params("id")

	var req struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	payment, err := pc.service.UpdatePaymentStatus(paymentID, req.Status)
	if err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			err.Error(),
			"",
			nil,
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Payment status updated successfully",
		"",
		payment,
	)
}
