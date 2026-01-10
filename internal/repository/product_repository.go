package repository

import (
	"context"

	"vestra-ecommerce-backend/internal/domain"
)

// ProductRepository defines product data access methods
type ProductRepository interface {

	// Create saves a new product
	Create(ctx context.Context, product *domain.Product) error

	// GetByID returns a product by its ID
	GetByID(ctx context.Context, productID string) (*domain.Product, error)

	// List returns all active products
	List(ctx context.Context) ([]*domain.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *domain.Product) error

	// Delete soft-deletes a product
	Delete(ctx context.Context, productID string) error
}
