package controller

import (
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/response"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/gofiber/fiber/v2"
)

type CartController struct {
	service *services.CartService
}

func NewCartController(service *services.CartService) *CartController {
	return &CartController{service: service}
}

/* =======================
   ADD TO CART
   ======================= */

type AddToCartRequest struct {
	ProductID string `json:"product_id"`
	Size      string `json:"size"`
	Quantity  int    `json:"quantity"`
}

func (cc *CartController) AddToCart(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if req.ProductID == "" || req.Size == "" || req.Quantity <= 0 {
		return response.Error(
			c,
			constant.BADREQUEST,
			"product_id, size and quantity are required",
			"",
			nil,
		)
	}

	if err := cc.service.AddToCart(userID, req.ProductID, req.Size, req.Quantity); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to add product to cart",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.CREATED,
		"Product added to cart",
		"",
		nil,
	)
}

/* =======================
   GET CART
   ======================= */

func (cc *CartController) GetCart(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	cart, err := cc.service.GetUserCart(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch cart",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Cart fetched successfully",
		"",
		cart,
	)
}

/* =======================
   UPDATE CART ITEM
   ======================= */

type UpdateCartItemRequest struct {
	Size     *string `json:"size"`
	Quantity *int    `json:"quantity"`
}

func (cc *CartController) UpdateCartItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	itemID := c.Params("id")

	if itemID == "" {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Cart item id is required",
			"",
			nil,
		)
	}

	var req UpdateCartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if req.Size == nil && req.Quantity == nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Nothing to update",
			"",
			nil,
		)
	}

	if err := cc.service.UpdateCartItem(userID, itemID, req.Size, req.Quantity); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to update cart item",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Cart item updated successfully",
		"",
		nil,
	)
}

/* =======================
   REMOVE CART ITEM
   ======================= */

func (cc *CartController) RemoveCartItem(c *fiber.Ctx) error {
	itemID := c.Params("id")
	if itemID == "" {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Cart item id is required",
			"",
			nil,
		)
	}

	if err := cc.service.RemoveCartItem(itemID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(c, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to remove cart item",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Cart item removed successfully",
		"",
		nil,
	)
}
