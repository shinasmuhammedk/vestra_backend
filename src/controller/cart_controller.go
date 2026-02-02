package controller

import (
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/response"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/gofiber/fiber/v2"
)

type CartController struct {
	service *services.CartService
}

func NewCartController(service *services.CartService) *CartController {
	logging.Debug.Println("CartController initialized")
	return &CartController{service: service}
}

// Helper: Get user ID from context
func getUserID(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		logging.Error.Println("getUserID - User ID not found in context")
		return "", response.Error(
			c,
			constant.UNAUTHORIZED,
			"Unauthorized",
			"",
			nil,
		)
	}
	return userID, nil
}

// Request structs
type AddToCartRequest struct {
	ProductID string `json:"product_id"`
	Size      string `json:"size"`
	Quantity  int    `json:"quantity"`
}

type UpdateCartItemRequest struct {
	Size     *string `json:"size"`
	Quantity *int    `json:"quantity"`
}

// AddToCart adds product to user's cart
func (cc *CartController) AddToCart(c *fiber.Ctx) error {
	logging.Debug.Println("AddToCart endpoint called")

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	logging.Debug.Printf("User adding to cart: %s", userID)

	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("AddToCart - Invalid request body: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if req.ProductID == "" || req.Size == "" || req.Quantity <= 0 {
		logging.Error.Printf("AddToCart - Missing required fields: product=%s, size=%s, qty=%d", 
			req.ProductID, req.Size, req.Quantity)
		return response.Error(
			c,
			constant.BADREQUEST,
			"product_id, size and quantity are required",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Adding to cart - Product: %s, Size: %s, Qty: %d", 
		req.ProductID, req.Size, req.Quantity)

	if err := cc.service.AddToCart(userID, req.ProductID, req.Size, req.Quantity); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("AddToCart failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		logging.Error.Printf("AddToCart error: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to add product to cart",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Product added to cart successfully - User: %s", userID)
	return response.Success(
		c,
		constant.CREATED,
		"Product added to cart",
		"",
		nil,
	)
}

// GetCart retrieves user's cart
func (cc *CartController) GetCart(c *fiber.Ctx) error {
	logging.Debug.Println("GetCart endpoint called")

	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	logging.Debug.Printf("Fetching cart for user: %s", userID)

	cart, err := cc.service.GetUserCart(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("GetCart failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		logging.Error.Printf("GetCart error: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch cart",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Cart fetched - Items: %d, User: %s", len(cart.Items), userID)
	return response.Success(
		c,
		constant.SUCCESS,
		"Cart fetched successfully",
		"",
		cart,
	)
}

// UpdateCartItem updates cart item details
func (cc *CartController) UpdateCartItem(c *fiber.Ctx) error {
	logging.Debug.Println("UpdateCartItem endpoint called")

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	itemID := c.Params("id")
	if itemID == "" {
		logging.Error.Println("UpdateCartItem - Missing cart item ID")
		return response.Error(
			c,
			constant.BADREQUEST,
			"Cart item id is required",
			"",
			nil,
		)
	}
	logging.Debug.Printf("Updating cart item - User: %s, Item: %s", userID, itemID)

	var req UpdateCartItemRequest
	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("UpdateCartItem - Invalid request body: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if req.Size == nil && req.Quantity == nil {
		logging.Error.Println("UpdateCartItem - No fields to update")
		return response.Error(
			c,
			constant.BADREQUEST,
			"Nothing to update",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Update request - Size: %v, Quantity: %v", req.Size, req.Quantity)
	
	if err := cc.service.UpdateCartItem(userID, itemID, req.Size, req.Quantity); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("UpdateCartItem failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		logging.Error.Printf("UpdateCartItem error: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to update cart item",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Cart item updated successfully - Item: %s", itemID)
	return response.Success(
		c,
		constant.SUCCESS,
		"Cart item updated successfully",
		"",
		nil,
	)
}

// RemoveCartItem removes item from cart
func (cc *CartController) RemoveCartItem(c *fiber.Ctx) error {
	logging.Debug.Println("RemoveCartItem endpoint called")

	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	itemID := c.Params("id")
	if itemID == "" {
		logging.Error.Println("RemoveCartItem - Missing cart item ID")
		return response.Error(
			c,
			constant.BADREQUEST,
			"Cart item id is required",
			"",
			nil,
		)
	}
	logging.Debug.Printf("Removing cart item - User: %s, Item: %s", userID, itemID)

	if err := cc.service.RemoveCartItem(itemID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("RemoveCartItem failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		logging.Error.Printf("RemoveCartItem error: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to remove cart item",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Cart item removed successfully - Item: %s", itemID)
	return response.Success(
		c,
		constant.SUCCESS,
		"Cart item removed successfully",
		"",
		nil,
	)
}