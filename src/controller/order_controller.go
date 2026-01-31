// controller/order_controller.go
package controller

import (
	"fmt"
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/response"
	"vestra-ecommerce/utils/utils/apperror"
	validator "vestra-ecommerce/utils/validation"

	"github.com/gofiber/fiber/v2"
)

type OrderController struct {
	services *services.OrderService
}

func NewOrderController(s *services.OrderService) *OrderController {
	return &OrderController{services: s}
}

func (oc *OrderController) PlaceOrder(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	var req services.PlaceOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, constant.BADREQUEST, "Invalid request body", "INVALID_BODY", err.Error())
	}
    
    // DEBUG: Check what you're receiving
    fmt.Printf("DEBUG - AddressID received: %q\n", req.AddressID)
    fmt.Printf("DEBUG - Type received: %q\n", req.Type)

	// Set default type if not provided
	if req.Type == "" {
		req.Type = "cart"
	}

	// Validate request using your custom validator function (not Struct method)
	if err := validator.Validate(req); err != nil {
		return response.Error(c, constant.BADREQUEST, "Validation failed", "VALIDATION_ERROR", err.Error())
	}

	order, err := oc.services.PlaceOrder(userID, req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to place order", "ORDER_ERROR", err.Error())
	}

	return response.Success(c, constant.CREATED, "Order placed successfully", "ORDER_CREATED", order)
}

func (oc *OrderController) GetMyOrders(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	orders, err := oc.services.GetUserOrders(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to fetch orders", "FETCH_ERROR", err.Error())
	}

	return response.Success(c, constant.SUCCESS, "Orders retrieved successfully", "ORDERS_FETCHED", orders)
}

func (oc *OrderController) GetOrderDetails(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	orderID := c.Params("id")

	order, err := oc.services.GetOrderDetails(userID, orderID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to fetch order details", "FETCH_ERROR", err.Error())
	}

	return response.Success(c, constant.SUCCESS, "Order details retrieved", "ORDER_FETCHED", order)
}


// controller/order_controller.go - Add these methods

func (oc *OrderController) GetUserOrders(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	
	orders, err := oc.services.GetUserOrders(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to fetch orders", "FETCH_ERROR", err.Error())
	}

	return response.Success(c, constant.SUCCESS, "Orders retrieved successfully", "ORDERS_FETCHED", orders)
}

func (oc *OrderController) UpdateOrderStatusUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	orderID := c.Params("id")

	var req struct {
		Status string `json:"status" validate:"required,oneof=CANCELLED"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, constant.BADREQUEST, "Invalid request body", "INVALID_BODY", err.Error())
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, constant.BADREQUEST, "Validation failed", "VALIDATION_ERROR", err.Error())
	}

	if err := oc.services.UpdateOrderStatusUser(userID, orderID, req.Status); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to update order status", "UPDATE_ERROR", err.Error())
	}

	return response.Success(c, constant.SUCCESS, "Order status updated successfully", "STATUS_UPDATED", nil)
}

func (oc *OrderController) CancelOrder(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	orderID := c.Params("id")

	if err := oc.services.CancelOrder(userID, orderID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to cancel order", "CANCEL_ERROR", err.Error())
	}

	return response.Success(c, constant.SUCCESS, "Order cancelled successfully", "ORDER_CANCELLED", nil)
}

func (oc *OrderController) DeleteOrder(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	orderID := c.Params("id")

	if err := oc.services.DeleteOrder(userID, orderID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to delete order", "DELETE_ERROR", err.Error())
	}

	return response.Success(c, constant.SUCCESS, "Order deleted successfully", "ORDER_DELETED", nil)
}


// controller/order_controller.go - Add these admin methods

func (oc *OrderController) GetAllOrders(c *fiber.Ctx) error {
	// Optional query params for filtering
	status := c.Query("status")
	userID := c.Query("user_id")
	
	orders, err := oc.services.GetAllOrders(status, userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to fetch orders", "FETCH_ERROR", err.Error())
	}

	return response.Success(c, constant.SUCCESS, "Orders retrieved successfully", "ORDERS_FETCHED", orders)
}

func (oc *OrderController) UpdateOrderStatusAdmin(c *fiber.Ctx) error {
	orderID := c.Params("id")

	var req struct {
		Status string `json:"status" validate:"required,oneof=PLACED PROCESSING SHIPPED DELIVERED CANCELLED"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, constant.BADREQUEST, "Invalid request body", "INVALID_BODY", err.Error())
	}

	if err := validator.Validate(req); err != nil {
		return response.Error(c, constant.BADREQUEST, "Validation failed", "VALIDATION_ERROR", err.Error())
	}

	if err := oc.services.UpdateOrderStatusAdmin(orderID, req.Status); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to update order status", "UPDATE_ERROR", err.Error())
	}

	return response.Success(c, constant.SUCCESS, "Order status updated successfully", "STATUS_UPDATED", nil)
}