// controller/order_controller.go
package controller

import (
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/response"
	"vestra-ecommerce/utils/utils/apperror"
	validator "vestra-ecommerce/utils/validation"

	"github.com/gofiber/fiber/v2"
)

type OrderController struct {
	services *services.OrderService
}

func NewOrderController(s *services.OrderService) *OrderController {
	logging.Debug.Println("OrderController initialized")
	return &OrderController{services: s}
}

// PlaceOrder places a new order
func (oc *OrderController) PlaceOrder(c *fiber.Ctx) error {
	logging.Debug.Println("PlaceOrder endpoint called")

	userID := c.Locals("user_id").(string)
	logging.Debug.Printf("User placing order: %s", userID)
	
	var req services.PlaceOrderRequest
	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("PlaceOrder - Invalid request body: %v", err)
		return response.Error(c, constant.BADREQUEST, "Invalid request body", "INVALID_BODY", err.Error())
	}
    
    logging.Debug.Printf("PlaceOrder request - AddressID: %q, Type: %q", req.AddressID, req.Type)

	// Set default type if not provided
	if req.Type == "" {
		req.Type = "cart"
		logging.Debug.Printf("Defaulting order type to: %s", req.Type)
	}

	// Validate request
	if err := validator.Validate(req); err != nil {
		logging.Error.Printf("PlaceOrder validation failed: %v", err)
		return response.Error(c, constant.BADREQUEST, "Validation failed", "VALIDATION_ERROR", err.Error())
	}

	logging.Debug.Printf("Processing order - Type: %s, AddressID: %s", req.Type, req.AddressID)
	order, err := oc.services.PlaceOrder(userID, req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("PlaceOrder failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("PlaceOrder error: %v", err)
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to place order", "ORDER_ERROR", err.Error())
	}

	logging.Debug.Printf("Order placed successfully: %s", order.ID)
	return response.Success(c, constant.CREATED, "Order placed successfully", "ORDER_CREATED", order)
}

// GetMyOrders retrieves user's orders
func (oc *OrderController) GetMyOrders(c *fiber.Ctx) error {
	logging.Debug.Println("GetMyOrders endpoint called")

	userID := c.Locals("user_id").(string)
	logging.Debug.Printf("Fetching orders for user: %s", userID)
	
	orders, err := oc.services.GetUserOrders(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("GetMyOrders failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("GetMyOrders error: %v", err)
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to fetch orders", "FETCH_ERROR", err.Error())
	}

	logging.Debug.Printf("Found %d orders for user: %s", len(orders), userID)
	return response.Success(c, constant.SUCCESS, "Orders retrieved successfully", "ORDERS_FETCHED", orders)
}

// GetOrderDetails retrieves specific order details
func (oc *OrderController) GetOrderDetails(c *fiber.Ctx) error {
	logging.Debug.Println("GetOrderDetails endpoint called")

	userID := c.Locals("user_id").(string)
	orderID := c.Params("id")
	logging.Debug.Printf("Fetching order %s for user: %s", orderID, userID)

	order, err := oc.services.GetOrderDetails(userID, orderID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("GetOrderDetails failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("GetOrderDetails error: %v", err)
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to fetch order details", "FETCH_ERROR", err.Error())
	}

	logging.Debug.Printf("Order details fetched: %s", order.ID)
	return response.Success(c, constant.SUCCESS, "Order details retrieved", "ORDER_FETCHED", order)
}

// GetUserOrders retrieves user's orders (same as GetMyOrders)
func (oc *OrderController) GetUserOrders(c *fiber.Ctx) error {
	logging.Debug.Println("GetUserOrders endpoint called")

	userID := c.Locals("user_id").(string)
	logging.Debug.Printf("Fetching user orders: %s", userID)
	
	orders, err := oc.services.GetUserOrders(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("GetUserOrders failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("GetUserOrders error: %v", err)
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to fetch orders", "FETCH_ERROR", err.Error())
	}

	logging.Debug.Printf("User orders retrieved - Count: %d", len(orders))
	return response.Success(c, constant.SUCCESS, "Orders retrieved successfully", "ORDERS_FETCHED", orders)
}

// UpdateOrderStatusUser updates order status (user actions)
func (oc *OrderController) UpdateOrderStatusUser(c *fiber.Ctx) error {
	logging.Debug.Println("UpdateOrderStatusUser endpoint called")

	userID := c.Locals("user_id").(string)
	orderID := c.Params("id")
	logging.Debug.Printf("User updating order status - User: %s, Order: %s", userID, orderID)

	var req struct {
		Status string `json:"status" validate:"required,oneof=CANCELLED"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("UpdateOrderStatusUser - Invalid request body: %v", err)
		return response.Error(c, constant.BADREQUEST, "Invalid request body", "INVALID_BODY", err.Error())
	}

	logging.Debug.Printf("User updating order %s status to: %s", orderID, req.Status)
	
	if err := validator.Validate(req); err != nil {
		logging.Error.Printf("UpdateOrderStatusUser validation failed: %v", err)
		return response.Error(c, constant.BADREQUEST, "Validation failed", "VALIDATION_ERROR", err.Error())
	}

	if err := oc.services.UpdateOrderStatusUser(userID, orderID, req.Status); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("UpdateOrderStatusUser failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("UpdateOrderStatusUser error: %v", err)
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to update order status", "UPDATE_ERROR", err.Error())
	}

	logging.Debug.Printf("User updated order status: %s -> %s", orderID, req.Status)
	return response.Success(c, constant.SUCCESS, "Order status updated successfully", "STATUS_UPDATED", nil)
}

// CancelOrder cancels an order
func (oc *OrderController) CancelOrder(c *fiber.Ctx) error {
	logging.Debug.Println("CancelOrder endpoint called")

	userID := c.Locals("user_id").(string)
	orderID := c.Params("id")
	logging.Debug.Printf("Cancelling order - User: %s, Order: %s", userID, orderID)

	if err := oc.services.CancelOrder(userID, orderID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("CancelOrder failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("CancelOrder error: %v", err)
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to cancel order", "CANCEL_ERROR", err.Error())
	}

	logging.Debug.Printf("Order cancelled: %s", orderID)
	return response.Success(c, constant.SUCCESS, "Order cancelled successfully", "ORDER_CANCELLED", nil)
}

// DeleteOrder deletes an order
func (oc *OrderController) DeleteOrder(c *fiber.Ctx) error {
	logging.Debug.Println("DeleteOrder endpoint called")

	userID := c.Locals("user_id").(string)
	orderID := c.Params("id")
	logging.Debug.Printf("Deleting order - User: %s, Order: %s", userID, orderID)

	if err := oc.services.DeleteOrder(userID, orderID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("DeleteOrder failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("DeleteOrder error: %v", err)
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to delete order", "DELETE_ERROR", err.Error())
	}

	logging.Debug.Printf("Order deleted: %s", orderID)
	return response.Success(c, constant.SUCCESS, "Order deleted successfully", "ORDER_DELETED", nil)
}

// GetAllOrders retrieves all orders (admin)
func (oc *OrderController) GetAllOrders(c *fiber.Ctx) error {
	logging.Debug.Println("GetAllOrders endpoint called (admin)")

	// Optional query params for filtering
	status := c.Query("status")
	userID := c.Query("user_id")
	logging.Debug.Printf("Admin fetching orders - Status filter: %s, User filter: %s", status, userID)
	
	orders, err := oc.services.GetAllOrders(status, userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("GetAllOrders failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("GetAllOrders error: %v", err)
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to fetch orders", "FETCH_ERROR", err.Error())
	}

	logging.Debug.Printf("Admin fetched %d orders", len(orders))
	return response.Success(c, constant.SUCCESS, "Orders retrieved successfully", "ORDERS_FETCHED", orders)
}

// UpdateOrderStatusAdmin updates order status (admin)
func (oc *OrderController) UpdateOrderStatusAdmin(c *fiber.Ctx) error {
	logging.Debug.Println("UpdateOrderStatusAdmin endpoint called")

	orderID := c.Params("id")
	logging.Debug.Printf("Admin updating order status: %s", orderID)

	var req struct {
		Status string `json:"status" validate:"required,oneof=PLACED PROCESSING SHIPPED DELIVERED CANCELLED"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("UpdateOrderStatusAdmin - Invalid request body: %v", err)
		return response.Error(c, constant.BADREQUEST, "Invalid request body", "INVALID_BODY", err.Error())
	}

	logging.Debug.Printf("Admin updating order %s to status: %s", orderID, req.Status)
	
	if err := validator.Validate(req); err != nil {
		logging.Error.Printf("UpdateOrderStatusAdmin validation failed: %v", err)
		return response.Error(c, constant.BADREQUEST, "Validation failed", "VALIDATION_ERROR", err.Error())
	}

	if err := oc.services.UpdateOrderStatusAdmin(orderID, req.Status); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("UpdateOrderStatusAdmin failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("UpdateOrderStatusAdmin error: %v", err)
		return response.Error(c, constant.INTERNALSERVERERROR, "Failed to update order status", "UPDATE_ERROR", err.Error())
	}

	logging.Debug.Printf("Admin updated order status: %s -> %s", orderID, req.Status)
	return response.Success(c, constant.SUCCESS, "Order status updated successfully", "STATUS_UPDATED", nil)
}