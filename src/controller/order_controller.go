package controller

import (
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/response"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/gofiber/fiber/v2"
)

type OrderController struct {
	service *services.OrderService
}

func NewOrderController(service *services.OrderService) *OrderController {
	return &OrderController{service: service}
}

/* =======================
   PLACE ORDER
   ======================= */

func (oc *OrderController) PlaceOrder(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	order, err := oc.service.PlaceOrder(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		return response.Error(
			c,
			constant.BADREQUEST,
			"Failed to place order",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.CREATED,
		"Order placed successfully",
		"",
		order,
	)
}

/* =======================
   GET USER ORDERS
   ======================= */

func (oc *OrderController) GetUserOrders(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	orders, err := oc.service.GetOrdersByUser(userID)
	if err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch orders",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Orders fetched successfully",
		"",
		orders,
	)
}

/* =======================
   GET ALL ORDERS (ADMIN)
   ======================= */

func (oc *OrderController) GetAllOrders(c *fiber.Ctx) error {
	orders, err := oc.service.GetAllOrders()
	if err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch orders",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"All orders fetched successfully",
		"",
		orders,
	)
}

/* =======================
   UPDATE ORDER STATUS
   ======================= */

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

func (oc *OrderController) UpdateOrderStatus(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return response.Error(
			c,
			constant.BADREQUEST,
			"order id is required",
			"",
			nil,
		)
	}

	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	order, err := oc.service.UpdateOrderStatus(orderID, req.Status)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(
			c,
			constant.BADREQUEST,
			"Failed to update order status",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Order status updated",
		"",
		order,
	)
}

/* =======================
   UPDATE ORDER STATUS (ADMIN)
   ======================= */

func (oc *OrderController) UpdateOrderStatusAdmin(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return response.Error(c, constant.BADREQUEST, "order id is required", "", nil)
	}

	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, constant.BADREQUEST, "Invalid request body", "", nil)
	}

	order, err := oc.service.UpdateOrderStatusByID("", orderID, req.Status, true)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(
			c,
			constant.BADREQUEST,
			"Failed to update order status",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Order status updated",
		"",
		order,
	)
}

/* =======================
   UPDATE ORDER STATUS (USER)
   ======================= */

func (oc *OrderController) UpdateOrderStatusUser(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return response.Error(c, constant.BADREQUEST, "order id is required", "", nil)
	}

	var req UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, constant.BADREQUEST, "Invalid request body", "", nil)
	}

	userID := c.Locals("user_id").(string)

	order, err := oc.service.UpdateOrderStatusByID(userID, orderID, req.Status, false)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(
			c,
			constant.BADREQUEST,
			"Failed to update order status",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Order status updated",
		"",
		order,
	)
}

/* =======================
   GET ORDER DETAILS
   ======================= */

func (oc *OrderController) GetOrderDetails(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return response.Error(c, constant.BADREQUEST, "order id is required", "", nil)
	}

	userID := c.Locals("user_id").(string)

	order, err := oc.service.GetOrderByID(userID, orderID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(
			c,
			constant.BADREQUEST,
			"Failed to fetch order",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Order fetched successfully",
		"",
		order,
	)
}

/* =======================
   DELETE ORDER
   ======================= */

func (oc *OrderController) DeleteOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return response.Error(c, constant.BADREQUEST, "order id is required", "", nil)
	}

	userID := c.Locals("user_id").(string)

	if err := oc.service.DeleteOrder(userID, orderID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(
			c,
			constant.BADREQUEST,
			"Failed to delete order",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Order deleted successfully",
		"",
		nil,
	)
}

/* =======================
   CANCEL ORDER
   ======================= */

func (oc *OrderController) CancelOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return response.Error(c, constant.BADREQUEST, "order id is required", "", nil)
	}

	userID := c.Locals("user_id").(string)

	order, err := oc.service.CancelOrder(userID, orderID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(
			c,
			constant.BADREQUEST,
			"Failed to cancel order",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Order cancelled successfully",
		"",
		order,
	)
}
