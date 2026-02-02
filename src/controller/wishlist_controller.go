package controller

import (
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/response"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/gofiber/fiber/v2"
)

type WishlistController struct {
	service *services.WishlistService
}

func NewWishlistController(service *services.WishlistService) *WishlistController {
	logging.Debug.Println("WishlistController initialized")
	return &WishlistController{service: service}
}

type AddToWishlistRequest struct {
	ProductID string `json:"product_id"`
}

// AddToWishlist adds product to user's wishlist
func (wc *WishlistController) AddToWishlist(c *fiber.Ctx) error {
	logging.Debug.Println("AddToWishlist endpoint called")

	var req AddToWishlistRequest
	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("AddToWishlist - Invalid request body: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if req.ProductID == "" {
		logging.Error.Println("AddToWishlist - product_id is required")
		return response.Error(
			c,
			constant.BADREQUEST,
			"product_id is required",
			"",
			nil,
		)
	}

	userID := c.Locals("user_id").(string)
	logging.Debug.Printf("AddToWishlist - User: %s, Product: %s", userID, req.ProductID)

	if err := wc.service.AddToWishlist(userID, req.ProductID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("AddToWishlist failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(
				c,
				appErr.Status,
				appErr.Message,
				appErr.Code,
				nil,
			)
		}

		logging.Error.Printf("AddToWishlist failed: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to add product to wishlist",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Product added to wishlist - User: %s, Product: %s", userID, req.ProductID)
	return response.Success(
		c,
		constant.CREATED,
		"Product added to wishlist",
		"",
		nil,
	)
}

// GetWishlist retrieves user's wishlist
func (wc *WishlistController) GetWishlist(c *fiber.Ctx) error {
	logging.Debug.Println("GetWishlist endpoint called")

	userID := c.Locals("user_id").(string)
	logging.Debug.Printf("GetWishlist - User: %s", userID)

	items, err := wc.service.GetWishlist(userID)
	if err != nil {
		logging.Error.Printf("GetWishlist failed for user %s: %v", userID, err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch wishlist",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Wishlist retrieved successfully - User: %s, Items: %d", userID, len(items))
	return response.Success(
		c,
		constant.SUCCESS,
		"Wishlist fetched successfully",
		"",
		items,
	)
}

// RemoveFromWishlist removes product from user's wishlist
func (wc *WishlistController) RemoveFromWishlist(c *fiber.Ctx) error {
	logging.Debug.Println("RemoveFromWishlist endpoint called")

	productID := c.Params("product_id")
	if productID == "" {
		logging.Error.Println("RemoveFromWishlist - product_id is required")
		return response.Error(
			c,
			constant.BADREQUEST,
			"product_id is required",
			"",
			nil,
		)
	}

	userID := c.Locals("user_id").(string)
	logging.Debug.Printf("RemoveFromWishlist - User: %s, Product: %s", userID, productID)

	if err := wc.service.RemoveFromWishlist(userID, productID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("RemoveFromWishlist failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(
				c,
				appErr.Status,
				appErr.Message,
				appErr.Code,
				nil,
			)
		}

		logging.Error.Printf("RemoveFromWishlist failed: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to remove product from wishlist",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Product removed from wishlist - User: %s, Product: %s", userID, productID)
	return response.Success(
		c,
		constant.SUCCESS,
		"Product removed from wishlist",
		"",
		nil,
	)
}