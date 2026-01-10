package handlers

import (
	"net/http"
	"vestra-ecommerce-backend/internal/domain"
	"vestra-ecommerce-backend/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	productUsecase *usecase.ProductUsecase
}

// NewProductHandler creates a new product handler
func NewProductHandler(productUsecase *usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{
		productUsecase: productUsecase,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req struct {
		Name         string         `json:"name"`
		Price        int            `json:"price"`
		ImageURL     string         `json:"image"`
		League       string         `json:"league"`
		KitType      string         `json:"kit"`
		Year         int            `json:"year"`
		IsTopSelling bool           `json:"topSelling"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	product := &domain.Product{
		ID:           uuid.New().String(), // Generate a new string ID
		Name:         req.Name,
		Price:        req.Price,
		ImageURL:     req.ImageURL,
		League:       req.League,
		KitType:      req.KitType,
		Year:         req.Year,
		IsTopSelling: req.IsTopSelling,
		IsActive:     true,
	}

	if err := h.productUsecase.CreateProduct(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "product created successfully",
		"product": product,
	})
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	productID := c.Param("id")

	product, err := h.productUsecase.GetProductByID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.productUsecase.ListProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")

	var req struct {
		Name         string `json:"name"`
		Price        int    `json:"price"`
		ImageURL     string `json:"image"`
		League       string `json:"league"`
		KitType      string `json:"kit"`
		Year         int    `json:"year"`
		IsTopSelling bool   `json:"topSelling"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	product := &domain.Product{
		ID:           productID,
		Name:         req.Name,
		Price:        req.Price,
		ImageURL:     req.ImageURL,
		League:       req.League,
		KitType:      req.KitType,
		Year:         req.Year,
		IsTopSelling: req.IsTopSelling,
	}

	if err := h.productUsecase.UpdateProduct(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product updated successfully"})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	if err := h.productUsecase.DeleteProduct(c.Request.Context(), productID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}

func (h *ProductHandler) AddProductSizes(c *gin.Context) {
    productIDStr := c.Param("id")

    // Validation: Ensure the string is a valid UUID format
    if _, err := uuid.Parse(productIDStr); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
        return
    }

    var req struct {
        Sizes map[string]int `json:"sizes"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
        return
    }

    var sizes []domain.ProductSize
    for size, qty := range req.Sizes {
        sizes = append(sizes, domain.ProductSize{
            ID:        uuid.New().String(), // Generate string UUID
            ProductID: productIDStr,        // Use string from URL
            Size:      size,
            Quantity:  qty,
        })
    }

    if err := h.productUsecase.AddProductSizes(c.Request.Context(), sizes); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "product sizes added successfully"})
}