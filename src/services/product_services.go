package services

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/utils/apperror"
)

type ProductService struct {
	repo repo.IPgSQLRepository
}

func NewProductService(repo repo.IPgSQLRepository) *ProductService {
	return &ProductService{repo: repo}
}

/* =======================
   INPUT STRUCTS
   ======================= */

type UpdateProductSizeInput struct {
	ID       *string
	Size     string
	Quantity int
}

type UpdateProductInput struct {
	Name         *string
	Price        *int
	ImageURL     *string
	League       *string
	KitType      *string
	Year         *int
	IsTopSelling *bool
	IsActive     *bool
	Sizes        *[]UpdateProductSizeInput
}


type ProductFilter struct {
	Category string
	MinPrice int
	MaxPrice int
	Search   string
	Size     string
}


/* =======================
   CREATE PRODUCT
   ======================= */

func (s *ProductService) CreateProduct(product *model.Product) error {
	if product == nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Product data is nil",
		)
	}

	if err := s.repo.Insert(product); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to create product",
		)
	}

	return nil
}

/* =======================
   GET PRODUCTS
   ======================= */

func (s *ProductService) GetAllProducts(filter ProductFilter) ([]model.Product, error) {
	var products []model.Product

	query := "1 = 1"
	args := []interface{}{}

	if filter.Category != "" {
		query += " AND category = ?"
		args = append(args, filter.Category)
	}

	if filter.MinPrice > 0 {
		query += " AND price >= ?"
		args = append(args, filter.MinPrice)
	}

	if filter.MaxPrice > 0 {
		query += " AND price <= ?"
		args = append(args, filter.MaxPrice)
	}

	if filter.Search != "" {
		query += " AND (name ILIKE ? OR description ILIKE ?)"
		search := "%" + filter.Search + "%"
		args = append(args, search, search)
	}

	if filter.Size != "" {
		query += `
			AND id IN (
				SELECT product_id
				FROM product_sizes
				WHERE size = ?
			)
		`
		args = append(args, filter.Size)
	}

	err := s.repo.FindWhereWithPreload(&products, query, args, "Sizes")
	if err != nil {
		return nil, err
	}

	return products, nil
}


func (s *ProductService) GetProductByID(id string) (*model.Product, error) {
	var product model.Product
	if err := s.repo.FindByIdWithPreload(&product, id, "Sizes"); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Product not found",
		)
	}
	return &product, nil
}

/* =======================
   DELETE PRODUCT
   ======================= */

func (s *ProductService) DeleteProduct(id string) error {
	var product model.Product
	if err := s.repo.FindById(&product, id); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"",
			"Product not found",
		)
	}
	if err := s.repo.Delete(&product, id); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to delete product",
		)
	}
	return nil
}

/* =======================
   UPDATE PRODUCT
   ======================= */

func (s *ProductService) UpdateProduct(id string, input *UpdateProductInput) (*model.Product, error) {
	var product model.Product
	if err := s.repo.FindByIdWithPreload(&product, id, "Sizes"); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Product not found",
		)
	}

	updates := map[string]interface{}{}

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Price != nil {
		updates["price"] = *input.Price
	}
	if input.ImageURL != nil {
		updates["image_url"] = *input.ImageURL
	}
	if input.League != nil {
		updates["league"] = *input.League
	}
	if input.KitType != nil {
		updates["kit_type"] = *input.KitType
	}
	if input.Year != nil {
		updates["year"] = *input.Year
	}
	if input.IsTopSelling != nil {
		updates["is_top_selling"] = *input.IsTopSelling
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if len(updates) > 0 {
		if err := s.repo.UpdateByFields(&model.Product{}, id, updates); err != nil {
			return nil, apperror.New(
				constant.INTERNALSERVERERROR,
				"",
				"Failed to update product",
			)
		}
	}

	if input.Sizes != nil {
		for _, sReq := range *input.Sizes {
			if sReq.ID != nil {
				// Update existing size
				fields := map[string]interface{}{
					"size":     sReq.Size,
					"quantity": sReq.Quantity,
				}
				if err := s.repo.UpdateByFields(&model.ProductSize{}, *sReq.ID, fields); err != nil {
					return nil, apperror.New(
						constant.INTERNALSERVERERROR,
						"",
						"Failed to update product size",
					)
				}
				continue
			}

			// Add new size
			newSize := model.ProductSize{
				ProductID: product.ID,
				Size:      sReq.Size,
				Quantity:  sReq.Quantity,
			}
			if err := s.repo.Insert(&newSize); err != nil {
				return nil, apperror.New(
					constant.INTERNALSERVERERROR,
					"",
					"Failed to add product size",
				)
			}
		}
	}

	// Reload updated product
	if err := s.repo.FindByIdWithPreload(&product, id, "Sizes"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to fetch updated product",
		)
	}

	return &product, nil
}

/* =======================
   SEARCH PRODUCTS
   ======================= */
func (s *ProductService) SearchProducts(
	query string,
	league string,
	kitType string,
	year *int,
) ([]model.Product, error) {

	var products []model.Product

	dbQuery := "1 = 1"
	args := []interface{}{}

	if query != "" {
		dbQuery += " AND name ILIKE ?"
		args = append(args, "%"+query+"%")
	}

	if league != "" {
		dbQuery += " AND league = ?"
		args = append(args, league)
	}

	if kitType != "" {
		dbQuery += " AND kit_type = ?"
		args = append(args, kitType)
	}

	if year != nil {
		dbQuery += " AND year = ?"
		args = append(args, *year)
	}

	if err := s.repo.FindWhereWithPreload(&products, dbQuery, args, "Sizes"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to fetch products",
		)
	}

	return products, nil
}
