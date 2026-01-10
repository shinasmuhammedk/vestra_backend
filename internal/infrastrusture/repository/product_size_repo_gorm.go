package repository

import (
	"context"
	"errors"

	"vestra-ecommerce-backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type productSizeRepositoryGorm struct {
	db *gorm.DB
}

func NewProductSizeRepository(db *gorm.DB) *productSizeRepositoryGorm {
	return &productSizeRepositoryGorm{db: db}
}

// CREATE
func (r *productSizeRepositoryGorm) Create(
	ctx context.Context,
	sizes []domain.ProductSize,
) error {
	var models []ProductSizeModel

	for _, s := range sizes {
		// If the ID is already set in the domain (e.g. by the handler), use it.
		// Otherwise, generate a new one.
		id := s.ID
		if id == "" {
			id = uuid.New().String() // FIX: Convert UUID object to string
		}

		models = append(models, ProductSizeModel{
			ID:        id,
			ProductID: s.ProductID,
			Size:      s.Size,
			Quantity:  s.Quantity,
		})
	}

	return r.db.WithContext(ctx).Create(&models).Error
}

// GET BY PRODUCT ID
// Changed productID type from uuid.UUID to string
func (r *productSizeRepositoryGorm) GetByProductID(
	ctx context.Context,
	productID string,
) ([]domain.ProductSize, error) {
	var models []ProductSizeModel

	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	var result []domain.ProductSize
	for _, m := range models {
		result = append(result, domain.ProductSize{
			ID:        m.ID,
			ProductID: m.ProductID,
			Size:      m.Size,
			Quantity:  m.Quantity,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		})
	}

	return result, nil
}

// REDUCE QUANTITY
// Changed productID type from uuid.UUID to string
func (r *productSizeRepositoryGorm) ReduceQuantity(
	ctx context.Context,
	productID string,
	size string,
	qty int,
) error {
	result := r.db.WithContext(ctx).
		Model(&ProductSizeModel{}).
		Where(
			"product_id = ? AND size = ? AND quantity >= ?",
			productID, size, qty,
		).
		Update("quantity", gorm.Expr("quantity - ?", qty))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("insufficient stock")
	}

	return nil
}

// GET TOTAL STOCK
// Changed productID type from uuid.UUID to string
func (r *productSizeRepositoryGorm) GetTotalStock(
	ctx context.Context,
	productID string,
) (int, error) {
	var total int

	err := r.db.WithContext(ctx).
		Model(&ProductSizeModel{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("product_id = ?", productID).
		Scan(&total).Error

	return total, err
}