package controller

import (
	"vestra-ecommerce/src/services"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/gofiber/fiber/v2"
)

type CartController struct {
	service *services.CartService
}

func NewCartController(service *services.CartService) *CartController {
	return &CartController{service: service}
}

type AddToCartRequest struct {
	ProductID string `json:"product_id"`
	Size      string `json:"size"`
	Quantity  int    `json:"quantity"`
}
func (cc *CartController) AddToCart(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	err := cc.service.AddToCart(
		userID,
		req.ProductID,
		req.Size,
		req.Quantity,
	)

	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return c.Status(appErr.Status).JSON(fiber.Map{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"message": "product added to cart",
	})
}



func (cc *CartController) GetCart(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	cart, err := cc.service.GetUserCart(userID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return c.Status(appErr.Status).JSON(fiber.Map{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.JSON(cart)
}




type UpdateCartItemRequest struct {
	Size     *string `json:"size"`
	Quantity *int    `json:"quantity"`
}

func (cc *CartController) UpdateCartItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	itemID := c.Params("id")

	if itemID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "cart item id is required",
		})
	}

	var req UpdateCartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	err := cc.service.UpdateCartItem(
		userID,
		itemID,
		req.Size,
		req.Quantity,
	)

	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return c.Status(appErr.Status).JSON(fiber.Map{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"message": "cart item updated successfully",
	})
}



// DELETE /cart/:id
func (cc *CartController) RemoveCartItem(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "cart item ID is required",
		})
	}

	if err := cc.service.RemoveCartItem(id); err != nil {
		appErr, ok := err.(*apperror.AppError)
		if ok {
			return c.Status(appErr.Status).JSON(fiber.Map{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "cart item removed successfully",
	})
}