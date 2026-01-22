package controller

import (
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/response"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/gofiber/fiber/v2"
)

type WishlistController struct {
	service *services.WishlistService
}

func NewWishlistController(service *services.WishlistService) *WishlistController {
	return &WishlistController{service: service}
}

type AddToWishlistRequest struct {
	ProductID string `json:"product_id"`
}

/* =======================
   ADD TO WISHLIST
   ======================= */

func (wc *WishlistController) AddToWishlist(c *fiber.Ctx) error {
	var req AddToWishlistRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if req.ProductID == "" {
		return response.Error(
			c,
			constant.BADREQUEST,
			"product_id is required",
			"",
			nil,
		)
	}

	userID := c.Locals("user_id").(string)

	if err := wc.service.AddToWishlist(userID, req.ProductID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(
				c,
				appErr.Status,
				appErr.Message,
				appErr.Code,
				nil,
			)
		}

		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to add product to wishlist",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.CREATED,
		"Product added to wishlist",
		"",
		nil,
	)
}

/* =======================
   GET WISHLIST
   ======================= */

func (wc *WishlistController) GetWishlist(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	items, err := wc.service.GetWishlist(userID)
	if err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch wishlist",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Wishlist fetched successfully",
		"",
		items,
	)
}

/* =======================
   REMOVE FROM WISHLIST
   ======================= */

func (wc *WishlistController) RemoveFromWishlist(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	if productID == "" {
		return response.Error(
			c,
			constant.BADREQUEST,
			"product_id is required",
			"",
			nil,
		)
	}

	userID := c.Locals("user_id").(string)

	if err := wc.service.RemoveFromWishlist(userID, productID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(
				c,
				appErr.Status,
				appErr.Message,
				appErr.Code,
				nil,
			)
		}

		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to remove product from wishlist",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Product removed from wishlist",
		"",
		nil,
	)
}
