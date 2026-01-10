package usecase

import (
	"context"
	"errors"

	"vestra-ecommerce-backend/internal/domain"
	"vestra-ecommerce-backend/internal/repository"

)

// ProductUsecase contains business logic for products
type ProductUsecase struct {
	productRepo     repository.ProductRepository
	productSizeRepo repository.ProductSizeRepository
}

// NewProductUsecase creates a new product usecase
func NewProductUsecase(
	productRepo repository.ProductRepository,
	productSizeRepo repository.ProductSizeRepository,
) *ProductUsecase {
	return &ProductUsecase{
		productRepo:     productRepo,
		productSizeRepo: productSizeRepo,
	}
}

// CreateProduct handles product creation logic
func (u *ProductUsecase) CreateProduct(ctx context.Context, product *domain.Product) error {

	// 1️⃣ Validate basic fields
	if product.Name == "" {
		return errors.New("product name is required")
	}

	if product.Price <= 0 {
		return errors.New("price must be greater than zero")
	}

	if product.Year <= 0 {
		return errors.New("invalid product year")
	}

	// 2️⃣ Stock validation
	// if product.Stock < -1 {
	// 	return errors.New("invalid stock value")
	// }

	// If stock is limited, sizes must be provided
	// if product.Stock != -1 {
	// 	if len(product.Sizes) == 0 {
	// 		return errors.New("sizes are required for limited stock products")
	// 	}

	// 	for size, qty := range product.Sizes {
	// 		if size == "" {
	// 			return errors.New("size name cannot be empty")
	// 		}
	// 		if qty < 0 {
	// 			return errors.New("size quantity cannot be negative")
	// 		}
	// 	}
	// }

	// 3️⃣ Set default values
	product.IsActive = true

	// 4️⃣ Save product
	return u.productRepo.Create(ctx, product)
}

// GetProductByID returns a product by ID
func (u *ProductUsecase) GetProductByID(ctx context.Context, productID string) (*domain.Product, error) {
	if productID == "" {
		return nil, errors.New("product id is required")
	}

	return u.productRepo.GetByID(ctx, productID)
}

// ListProducts returns all active products
func (u *ProductUsecase) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	return u.productRepo.List(ctx)
}

// UpdateProduct handles product update logic
func (u *ProductUsecase) UpdateProduct(ctx context.Context, product *domain.Product) error {

	if product.ID == "" {
		return errors.New("product id is required")
	}

	if product.Price <= 0 {
		return errors.New("price must be greater than zero")
	}

	// if product.Stock < -1 {
	// 	return errors.New("invalid stock value")
	// }

	return u.productRepo.Update(ctx, product)
}

// DeleteProduct soft-deletes a product
func (u *ProductUsecase) DeleteProduct(ctx context.Context, productID string) error {
	if productID == "" {
		return errors.New("product id is required")
	}

	return u.productRepo.Delete(ctx, productID)
}

// func (u *ProductUsecase) AddProductSizes(
// 	ctx context.Context,
// 	productID string,
// 	sizes []struct {
// 		Size     string
// 		Quantity int
// 	},
// ) error {

// 	var productSizes []domain.ProductSize
// 	pid, err := uuid.Parse(productID)
// 	if err != nil {
// 		return err
// 	}
// 	for _, s := range sizes {
// 		productSizes = append(productSizes, domain.ProductSize{
// 			ProductID: pid,
// 			Size:      s.Size,
// 			Quantity:  s.Quantity,
// 		})
// 	}

// 	return u.productSizeRepo.Create(ctx, productSizes)
// }

func (u *ProductUsecase) AddProductSizes(
	ctx context.Context,
	sizes []domain.ProductSize,
) error {

	if len(sizes) == 0 {
		return errors.New("sizes cannot be empty")
	}

	for _, s := range sizes {
		if s.Size == "" {
			return errors.New("size name cannot be empty")
		}
		if s.Quantity < 0 {
			return errors.New("quantity cannot be negative")
		}
	}

	return u.productSizeRepo.Create(ctx, sizes)
}

