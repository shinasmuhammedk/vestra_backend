package services

import (
	"fmt"
	"strings"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	database "vestra-ecommerce/utils/databases"
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
    Category  string
    Search    string
    Size      string
    KitType   string // "home", "away", "third"
    League    string // "laliga", "premier", "spl", etc.
    MinPrice  int
    MaxPrice  int
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


func (s *ProductService) GetAllProducts(
    filter ProductFilter,
    page, pageSize int,
    sortBy, sortOrder string,
) ([]model.Product, int64, error) {

    var (
        ids        []string
        products   []model.Product
        totalCount int64
    )

    db := database.PgSQLDB.Table("products p")

    // -------- FILTERS --------
    if filter.Search != "" {
        search := "%" + filter.Search + "%"
        db = db.Where("(p.name ILIKE ? OR p.description ILIKE ?)", search, search)
    }

    // New Jersey-Specific Filters
    if filter.KitType != "" {
        db = db.Where("p.kit_type = ?", filter.KitType)
    }

    if filter.League != "" {
        db = db.Where("p.league = ?", filter.League)
    }

    if filter.MinPrice > 0 {
        db = db.Where("p.price >= ?", filter.MinPrice)
    }

    if filter.MaxPrice > 0 {
        db = db.Where("p.price <= ?", filter.MaxPrice)
    }

    // Size Filter (Using JOIN for better performance)
    if filter.Size != "" {
        db = db.Joins("JOIN product_sizes ps ON ps.product_id = p.id").
               Where("ps.size = ?", filter.Size)
    }

    // Category Join
    if filter.Category != "" {
        db = db.Joins("JOIN product_categories pc ON pc.product_id = p.id").
               Joins("JOIN categories c ON c.id = pc.category_id").
               Where("c.slug = ? OR c.name = ?", filter.Category, filter.Category)
    }

    // -------- COUNT & PAGINATION --------
    db.Select("COUNT(DISTINCT p.id)").Count(&totalCount)

    offset := (page - 1) * pageSize
    
    // Sort logic
    sortCol := "p.price" // Default
    if sortBy == "name" { sortCol = "p.name" }
    if sortBy == "created_at" { sortCol = "p.created_at" }

    // Step 1: Get IDs
    if err := db.Select("p.id").
        Group("p.id, " + sortCol).
        Order(sortCol + " " + sortOrder).
        Limit(pageSize).Offset(offset).
        Pluck("p.id", &ids).Error; err != nil {
        return nil, 0, err
    }

    if len(ids) == 0 { return []model.Product{}, totalCount, nil }

    // Step 2: Fetch full data with ordered UUIDs
    var quotedIds []string
    for _, id := range ids { quotedIds = append(quotedIds, fmt.Sprintf("'%s'", id)) }
    
    err := database.PgSQLDB.Model(&model.Product{}).
        Where("id IN ?", ids).
        Preload("Sizes").
        Order(fmt.Sprintf("array_position(ARRAY[%s]::uuid[], id)", strings.Join(quotedIds, ","))).
        Find(&products).Error

    return products, totalCount, err
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
