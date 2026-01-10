package repository

import (
	"context"
	"errors"

	"vestra-ecommerce-backend/internal/domain"
	"vestra-ecommerce-backend/internal/repository"

	"gorm.io/gorm"
)

type productRepositoryGorm struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) repository.ProductRepository {
	return &productRepositoryGorm{db: db}
}

func (r *productRepositoryGorm) Create(ctx context.Context, product *domain.Product) error {
	model := ProductModel{
		ID:           product.ID,
		Name:         product.Name,
		Price:        product.Price,
		ImageURL:     product.ImageURL,
		League:       product.League,
		KitType:      product.KitType,
		Year:         product.Year,
		IsTopSelling: product.IsTopSelling,
		IsActive:     product.IsActive,
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	if len(product.Sizes) > 0 {
		return r.UpdateSizes(ctx, product.ID, product.Sizes)
	}

	return nil
}

func (r *productRepositoryGorm) GetByID(ctx context.Context, productID string) (*domain.Product, error) {
	var model ProductModel

	// Preload "Sizes" is essential to see size data
	err := r.db.WithContext(ctx).
		Preload("Sizes").
		Where("id = ?", productID). // Removed is_active filter for debugging
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return mapModelToDomain(&model), nil
}

func (r *productRepositoryGorm) List(ctx context.Context) ([]*domain.Product, error) {
	var models []ProductModel

	// Note: Removed 'is_active' filter to ensure you see your data first. 
	// You can add .Where("is_active = ?", true) back once data appears.
	err := r.db.WithContext(ctx).Find(&models).Error
	if err != nil {
		return nil, err
	}

	products := make([]*domain.Product, 0, len(models))
	for _, m := range models {
		products = append(products, mapModelToDomain(&m))
	}

	return products, nil
}

// Helper function to map Model -> Domain to avoid repetition
func mapModelToDomain(m *ProductModel) *domain.Product {
	var sizes []domain.ProductSize
	for _, s := range m.Sizes {
		sizes = append(sizes, domain.ProductSize{
			ID:        s.ID,
			ProductID: s.ProductID,
			Size:      s.Size,
			Quantity:  s.Quantity,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}

	return &domain.Product{
		ID:           m.ID,
		Name:         m.Name,
		Price:        m.Price,
		ImageURL:     m.ImageURL,
		League:       m.League,
		KitType:      m.KitType,
		Year:         m.Year,
		IsTopSelling: m.IsTopSelling,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		Sizes:        sizes,
	}
}

func (r *productRepositoryGorm) Update(ctx context.Context, product *domain.Product) error {
	return r.db.WithContext(ctx).
		Model(&ProductModel{}).
		Where("id = ?", product.ID).
		Updates(map[string]interface{}{
			"name":           product.Name,
			"price":          product.Price,
			"image_url":      product.ImageURL,
			"league":         product.League,
			"kit_type":       product.KitType,
			"year":           product.Year,
			"is_top_selling": product.IsTopSelling,
		}).Error
}

func (r *productRepositoryGorm) Delete(ctx context.Context, productID string) error {
	return r.db.WithContext(ctx).
		Model(&ProductModel{}).
		Where("id = ?", productID).
		Update("is_active", false).Error
}

func (r *productRepositoryGorm) UpdateSizes(ctx context.Context, productID string, sizes []domain.ProductSize) error {
	var product ProductModel
	if err := r.db.WithContext(ctx).Where("id = ?", productID).First(&product).Error; err != nil {
		return err
	}

	var sizeModels []ProductSizeModel
	for _, s := range sizes {
		sizeModels = append(sizeModels, ProductSizeModel{
			ID:        s.ID,
			ProductID: productID,
			Size:      s.Size,
			Quantity:  s.Quantity,
		})
	}

	return r.db.WithContext(ctx).Model(&product).Association("Sizes").Replace(sizeModels)
}