package controller

import (
	"math"
	"strconv"
	"strings"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/response"

	"github.com/gofiber/fiber/v2"
)

type ProductController struct {
	service *services.ProductService
}

func NewProductController(service *services.ProductService) *ProductController {
	return &ProductController{service: service}
}

/* =======================
   REQUEST STRUCTS
   ======================= */

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

type UpdateProductRequest struct {
	Name         *string `json:"name"`
	Price        *int    `json:"price"`
	ImageURL     *string `json:"image_url"`
	League       *string `json:"league"`
	KitType      *string `json:"kit_type"`
	Year         *int    `json:"year"`
	IsTopSelling *bool   `json:"is_top_selling"`
	IsActive     *bool   `json:"is_active"`

	Sizes *[]struct {
		ID       *string `json:"id"` // existing size ID, optional
		Size     string  `json:"size"`
		Quantity int     `json:"quantity"`
	} `json:"sizes"`
}

/* =======================
   CREATE PRODUCT
   ======================= */

func (pc *ProductController) CreateProduct(c *fiber.Ctx) error {
	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		// return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if req.Name == "" || req.Price <= 0 || len(req.Sizes) == 0 {
		// return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 	"error": "name, price and sizes are required",
		// })
		return response.Error(
			c,
			constant.BADREQUEST,
			"name, price and sizes are required",
			"",
			nil,
		)
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
		// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to create product",
			"",
			err.Error(),
		)
	}

	// return c.Status(fiber.StatusCreated).JSON(product)
	return response.Success(
		c,
		constant.CREATED,
		"Product created successfully",
		"",
		product,
	)
}



func (pc *ProductController) GetAllProducts(c *fiber.Ctx) error {
    filter := services.ProductFilter{
        Category: c.Query("category"),
        Search:   c.Query("q"),
        Size:     c.Query("size"),      // xs, s, m, l, xl, 2xl
        KitType:  c.Query("kit_type"),  // home, away, third
        League:   strings.ToLower(c.Query("league")),    // laliga, premier, spl
    }

    // Parse numeric fields with validation
    page, _ := strconv.Atoi(c.Query("page", "1"))
    if page < 1 {
        page = 1
    }
    
    pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))
    if pageSize < 1 || pageSize > 100 {
        pageSize = 10 // Max 100 items per page
    }
    
    filter.MinPrice, _ = strconv.Atoi(c.Query("min_price"))
    filter.MaxPrice, _ = strconv.Atoi(c.Query("max_price"))

    products, total, err := pc.service.GetAllProducts(
        filter, 
        page, 
        pageSize, 
        c.Query("sort_by", "created_at"), // Changed default to created_at
        c.Query("sort_order", "desc"),    // Changed default to desc
    )

    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // Calculate pagination metadata
    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
    hasNext := page < totalPages
    hasPrev := page > 1

    return c.JSON(fiber.Map{
        "data": products,
        "pagination": fiber.Map{
            "current_page":  page,
            "page_size":     pageSize,
            "total_items":   total,
            "total_pages":   totalPages,
            "has_next":      hasNext,
            "has_previous":  hasPrev,
            "next_page":     page + 1,
            "previous_page": page - 1,
        },
    })
}


/* =======================
   GET PRODUCT BY ID
   ======================= */

func (pc *ProductController) GetProductByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		// return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "product id is required"})
		return response.Error(
			c,
			constant.BADREQUEST,
			"Product ID is required",
			"",
			nil,
		)
	}

	product, err := pc.service.GetProductByID(id)
	if err != nil {
		// return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "product not found"})
		return response.Error(
			c,
			constant.NOTFOUND,
			"Product not found",
			"",
			nil,
		)
	}

	// return c.Status(fiber.StatusOK).JSON(product)
	return response.Success(
		c,
		constant.SUCCESS,
		"Product fetched successfully",
		"",
		product,
	)
}

/* =======================
   DELETE PRODUCT
   ======================= */

func (pc *ProductController) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Product id is required",
			"",
			nil,
		)
	}

	if err := pc.service.DeleteProduct(id); err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to delete product",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Product deleted successfully",
		"",
		nil,
	)
}


/* =======================
   UPDATE PRODUCT
   ======================= */

func (pc *ProductController) UpdateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		// return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "product id is required"})
		return response.Error(
			c,
			constant.BADREQUEST,
			"Product ID is required",
			"",
			nil,
		)
	}

	var req UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		// return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	// Map controller sizes to service sizes
	var sizes *[]services.UpdateProductSizeInput
	if req.Sizes != nil {
		tmp := make([]services.UpdateProductSizeInput, 0, len(*req.Sizes))
		for _, s := range *req.Sizes {
			tmp = append(tmp, services.UpdateProductSizeInput{
				ID:       s.ID,
				Size:     s.Size,
				Quantity: s.Quantity,
			})
		}
		sizes = &tmp
	}

	input := services.UpdateProductInput{
		Name:         req.Name,
		Price:        req.Price,
		ImageURL:     req.ImageURL,
		League:       req.League,
		KitType:      req.KitType,
		Year:         req.Year,
		IsTopSelling: req.IsTopSelling,
		IsActive:     req.IsActive,
		Sizes:        sizes,
	}

	product, err := pc.service.UpdateProduct(id, &input)
	if err != nil {
		// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to update product",
			"",
			err.Error(),
		)
	}

	// return c.Status(fiber.StatusOK).JSON(product)
	return response.Success(
		c,
		constant.SUCCESS,
		"Product updated successfully",
		"",
		product,
	)
}

/* =======================
   SEARCH PRODUCTS
   ======================= */

func (pc *ProductController) SearchProducts(c *fiber.Ctx) error {
	query := c.Query("q")
	league := strings.ToLower(c.Query("league"))
	kitType := c.Query("kit_type")

	products, err := pc.service.SearchProducts(query, league, kitType, nil)
	if err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to search products",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Products fetched successfully",
		"",
		products,
	)
}



// GET /admin/payments
func (pc *PaymentController) GetAllPayments(c *fiber.Ctx) error {
	payments, err := pc.service.GetAllPayments()
	if err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch payments",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Payments fetched successfully",
		"",
		payments,
	)
}
