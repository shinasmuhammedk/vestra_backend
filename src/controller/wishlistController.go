package controller

import (
	"vestra-ecommerce/src/services"
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

// POST /wishlist
func (wc *WishlistController) AddToWishlist(c *fiber.Ctx) error {
	var req AddToWishlistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.ProductID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "product_id is required"})
	}

	userID := c.Locals("user_id").(string) // assuming JWT middleware sets this

	if err := wc.service.AddToWishlist(userID, req.ProductID); err != nil {
		appErr, ok := err.(*apperror.AppError)
		if ok {
			return c.Status(appErr.Status).JSON(fiber.Map{
				"code":    appErr.Code,
				"message": appErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "product added to wishlist"})
}



func (wc *WishlistController) GetWishlist(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	items, err := wc.service.GetWishlist(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(items)
}



func (wc *WishlistController) RemoveFromWishlist(c *fiber.Ctx) error {
	productID := c.Params("product_id")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product_id is required",
		})
	}

	userID := c.Locals("user_id").(string)

	if err := wc.service.RemoveFromWishlist(userID, productID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Product removed from wishlist",
	})
}
