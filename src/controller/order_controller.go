package controller

import (
	"github.com/gofiber/fiber/v2"
	"vestra-ecommerce/src/services"
)

type OrderController struct {
	service *services.OrderService
}

func NewOrderController(service *services.OrderService) *OrderController {
	return &OrderController{service: service}
}

func (oc *OrderController) PlaceOrder(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	order, err := oc.service.PlaceOrder(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

func (oc *OrderController) GetUserOrders(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	orders, err := oc.service.GetOrdersByUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(orders)
}




func (oc *OrderController) GetAllOrders(c *fiber.Ctx) error {
	orders, err := oc.service.GetAllOrders()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(orders)
}



type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

func (oc *OrderController) UpdateOrderStatus(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	order, err := oc.service.UpdateOrderStatus(orderID, req.Status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(order)
}



// type UpdateOrderStatusRequest struct {
// 	Status string `json:"status"`
// }

// Admin endpoint: PUT /admin/order/:id
func (oc *OrderController) UpdateOrderStatusAdmin(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "order id is required"})
	}

	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Admin can update any order
	order, err := oc.service.UpdateOrderStatusByID("", orderID, req.Status, true)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(order)
}

// User endpoint: PUT /order/:id/status
func (oc *OrderController) UpdateOrderStatusUser(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "order id is required"})
	}

	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID := c.Locals("user_id").(string)

	// User can only cancel their own orders
	order, err := oc.service.UpdateOrderStatusByID(userID, orderID, req.Status, false)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(order)
}



func (oc *OrderController) GetOrderDetails(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	userID := c.Locals("user_id").(string)

	order, err := oc.service.GetOrderByID(userID, orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(order)
}



func (oc *OrderController) DeleteOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	userID := c.Locals("user_id").(string)

	if err := oc.service.DeleteOrder(userID, orderID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "order deleted successfully",
	})
}




func (oc *OrderController) CancelOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	userID := c.Locals("user_id").(string)

	order, err := oc.service.CancelOrder(userID, orderID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(order)
}
