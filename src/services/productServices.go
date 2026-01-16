package services

import (
	"errors"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"

)

type ProductService struct {
	repo repo.IPgSQLRepository
}

func NewProductService(repo repo.IPgSQLRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) CreateProduct(product *model.Product) error {
	if product == nil {
		return errors.New("product data is nil")
	}

	// Insert product with sizes (gorm handles relations)
	if err := s.repo.Insert(product); err != nil {
		return err
	}

	return nil
}


func (s *ProductService) GetAllProducts() ([]model.Product, error) {
	var products []model.Product
	// Preload Sizes to include them in the result
	if err := s.repo.Raw("SELECT * FROM products").Preload("Sizes").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}


func (s *ProductService) GetProductByID(id string) (*model.Product, error) {
	var product model.Product
	// Preload Sizes so we get the sizes as well
	if err := s.repo.Raw("SELECT * FROM products WHERE id = ?", id).Preload("Sizes").First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}


