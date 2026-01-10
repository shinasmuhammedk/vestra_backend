package repository

import (
	"context"
	"vestra-ecommerce-backend/internal/domain"
)

type ProductSizeRepository interface {
	Create(
		ctx context.Context,
		sizes []domain.ProductSize,
	) error

	// Changed uuid.UUID to string
	GetByProductID(
		ctx context.Context,
		productID string,
	) ([]domain.ProductSize, error)

	// Changed uuid.UUID to string
	ReduceQuantity(
		ctx context.Context,
		productID string,
		size string,
		qty int,
	) error

	// Changed uuid.UUID to string
	GetTotalStock(
		ctx context.Context,
		productID string,
	) (int, error)
}