package controller

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/services"
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
            400, 
            "Invalid request body", 
            "", 
            nil,
        )
	}

	payment, err := pc.service.CreatePayment(userID, req)
	if err != nil {
		return response.Error(
            c,
             500, 
             "Failed to create payment",
              "",
               err.Error(),
            )
	}

	return response.Success(
        c, 
        201, 
        "Payment created successfully",
         "", 
         payment,
        )
}




// POST /user/payment/verify
func (pc *PaymentController) VerifyPayment(c *fiber.Ctx) error {
    var req services.VerifyPaymentRequest

    if err := c.BodyParser(&req); err != nil {
        return response.Error(c, 400, "Invalid request body", "", nil)
    }

    payment, err := pc.service.VerifyPayment(req.PaymentID, req.TransactionID, req.Status)
    if err != nil {
        return response.Error(c, 400, err.Error(), "", nil)
    }

    return response.Success(c, 200, "Payment verified successfully", "", payment)
}



// GET /user/payment
func (pc *PaymentController) GetUserPayments(c *fiber.Ctx) error {
	// Get logged-in user ID from context
	userID := c.Locals("user_id").(string)

	payments, err := pc.service.GetPaymentsByUser(userID)
	if err != nil {
		return response.Error(c, 500, "Failed to fetch payments", "", err.Error())
	}

	return response.Success(c, 200, "Payments fetched successfully", "", payments)
}



// GET /user/payment/:id
func (pc *PaymentController) GetUserPaymentByID(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	paymentID := c.Params("id")

	payment, err := pc.service.GetPaymentByID(userID, paymentID)
	if err != nil {
		return response.Error(c, 404, err.Error(), "", nil)
	}

	return response.Success(
		c,
		200,
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
		return response.Error(c, 400, err.Error(), "", nil)
	}

	return response.Success(
		c,
		200,
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
			404,
			err.Error(),
			"",
			nil,
		)
	}

	return response.Success(
		c,
		200,
		"Payment fetched successfully",
		"",
		payment,
	)
}
