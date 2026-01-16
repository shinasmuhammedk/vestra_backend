package controller

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/services"

	"github.com/gofiber/fiber/v2"
)

type ProductController struct {
	service *services.ProductService
}

func NewProductController(service *services.ProductService) *ProductController {
	return &ProductController{service: service}
}

type CreateProductRequest struct {
	Name         string `json:"name"`
	Price        int    `json:"price"`
	ImageURL     string `json:"image_url"`
	League       string `json:"league"`
	KitType      string `json:"kit_type"`
	Year         int    `json:"year"`
	IsTopSelling bool   `json:"is_top_selling"`
	Sizes        []struct {
		Size     string `json:"size"`
		Quantity int    `json:"quantity"`
	} `json:"sizes"`
}

func (pc *ProductController) CreateProduct(c *fiber.Ctx) error {
	var req CreateProductRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// basic validation
	if req.Name == "" || req.Price <= 0 || len(req.Sizes) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name, price and sizes are required",
		})
	}

	product := model.Product{
		Name:         req.Name,
		Price:        req.Price,
		ImageURL:     req.ImageURL,
		League:       req.League,
		KitType:      req.KitType,
		Year:         req.Year,
		IsTopSelling: req.IsTopSelling,
		IsActive:     true,
	}

	for _, s := range req.Sizes {
		product.Sizes = append(product.Sizes, model.ProductSize{
			Size:     s.Size,
			Quantity: s.Quantity,
		})
	}

	if err := pc.service.CreateProduct(&product); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}



func (pc *ProductController) GetAllProducts(c *fiber.Ctx) error {
	products, err := pc.service.GetAllProducts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(products)
}


func (pc *ProductController) GetProductByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product ID is required",
		})
	}

	product, err := pc.service.GetProductByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "product not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(product)
}
