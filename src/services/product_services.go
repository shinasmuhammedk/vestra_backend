package services

import (
	"fmt"
	"strings"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	database "vestra-ecommerce/utils/databases"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/utils/apperror"
)

// ProductService handles product operations
type ProductService struct {
	repo repo.IPgSQLRepository
}

// NewProductService creates product service instance
func NewProductService(repo repo.IPgSQLRepository) *ProductService {
	logging.Debug.Println("ProductService initialized")
	return &ProductService{repo: repo}
}

// UpdateProductSizeInput for size updates
type UpdateProductSizeInput struct {
	ID       *string
	Size     string
	Quantity int
}

// UpdateProductInput for product updates
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

// ProductFilter for product filtering
type ProductFilter struct {
	Category  string
	Search    string
	Size      string
	KitType   string
	League    string
	MinPrice  int
	MaxPrice  int
}

// CreateProduct creates new product
func (s *ProductService) CreateProduct(product *model.Product) error {
	logging.Debug.Printf("Creating product: %s", product.Name)

	if product == nil {
		logging.Error.Println("CreateProduct: product data is nil")
		return apperror.New(constant.BADREQUEST, "", "Product data is nil")
	}

	if err := s.repo.Insert(product); err != nil {
		logging.Error.Printf("CreateProduct failed for %s: %v", product.Name, err)
		return apperror.New(constant.INTERNALSERVERERROR, "", "Failed to create product")
	}

	logging.Debug.Printf("Product created: %s", product.Name)
	return nil
}

// GetAllProducts retrieves products with filtering and pagination
func (s *ProductService) GetAllProducts(
	filter ProductFilter,
	page, pageSize int,
	sortBy, sortOrder string,
) ([]model.Product, int64, error) {
	logging.Debug.Printf("GetAllProducts: page=%d, pageSize=%d, filter=%+v", page, pageSize, filter)

	var (
		ids        []string
		products   []model.Product
		totalCount int64
	)

	// Build base query
	db := database.PgSQLDB.Table("products p")

	// Apply filters
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		db = db.Where("(p.name ILIKE ? OR p.description ILIKE ?)", search, search)
		logging.Debug.Printf("Applied search filter: %s", filter.Search)
	}

	if filter.KitType != "" {
		db = db.Where("p.kit_type = ?", filter.KitType)
	}

	if filter.League != "" {
		db = db.Where("p.league ILIKE ?", filter.League)
	}

	if filter.MinPrice > 0 {
		db = db.Where("p.price >= ?", filter.MinPrice)
	}

	if filter.MaxPrice > 0 {
		db = db.Where("p.price <= ?", filter.MaxPrice)
	}

	if filter.Size != "" {
		db = db.Joins("JOIN product_sizes ps ON ps.product_id = p.id").
			Where("ps.size = ?", filter.Size)
	}

	if filter.Category != "" {
		db = db.Joins("JOIN product_categories pc ON pc.product_id = p.id").
			Joins("JOIN categories c ON c.id = pc.category_id").
			Where("c.slug = ? OR c.name = ?", filter.Category, filter.Category)
	}

	// Get total count
	if err := database.PgSQLDB.Table("(?) as distinct_products", db.Select("DISTINCT p.id")).
		Count(&totalCount).Error; err != nil {
		logging.Error.Printf("Failed to count products: %v", err)
		return nil, 0, err
	}
	logging.Debug.Printf("Total products found: %d", totalCount)

	// Apply pagination
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Set sort column
	sortCol := "p.created_at"
	switch sortBy {
	case "price":
		sortCol = "p.price"
	case "name":
		sortCol = "p.name"
	case "year":
		sortCol = "p.year"
	}

	// Get paginated IDs
	if err := db.Select("DISTINCT p.id").
		Order(sortCol + " " + sortOrder).
		Limit(pageSize).
		Offset(offset).
		Pluck("p.id", &ids).Error; err != nil {
		logging.Error.Printf("Failed to get product IDs: %v", err)
		return nil, 0, err
	}

	if len(ids) == 0 {
		logging.Debug.Println("No products found with filters")
		return []model.Product{}, totalCount, nil
	}
	logging.Debug.Printf("Found %d products on page %d", len(ids), page)

	// Fetch full product data
	err := database.PgSQLDB.Model(&model.Product{}).
		Where("id IN ?", ids).
		Preload("Sizes").
		Preload("Categories").
		Order(fmt.Sprintf("array_position(ARRAY[%s]::uuid[], id)", strings.Join(quoteUUIDs(ids), ","))).
		Find(&products).Error

	if err != nil {
		logging.Error.Printf("Failed to load product details: %v", err)
	}

	return products, totalCount, err
}

// Helper to quote UUIDs for PostgreSQL
func quoteUUIDs(ids []string) []string {
	quoted := make([]string, len(ids))
	for i, id := range ids {
		quoted[i] = fmt.Sprintf("'%s'", id)
	}
	return quoted
}

// GetProductByID retrieves single product by ID
func (s *ProductService) GetProductByID(id string) (*model.Product, error) {
	logging.Debug.Printf("GetProductByID: %s", id)

	var product model.Product
	if err := s.repo.FindByIdWithPreload(&product, id, "Sizes"); err != nil {
		logging.Error.Printf("Product not found: %s - %v", id, err)
		return nil, apperror.New(constant.NOTFOUND, "", "Product not found")
	}

	logging.Debug.Printf("Product found: %s", product.Name)
	return &product, nil
}

// DeleteProduct removes product by ID
func (s *ProductService) DeleteProduct(id string) error {
	logging.Debug.Printf("Deleting product: %s", id)

	var product model.Product
	if err := s.repo.FindById(&product, id); err != nil {
		logging.Error.Printf("Product to delete not found: %s", id)
		return apperror.New(constant.NOTFOUND, "", "Product not found")
	}

	if err := s.repo.Delete(&product, id); err != nil {
		logging.Error.Printf("Failed to delete product %s: %v", id, err)
		return apperror.New(constant.INTERNALSERVERERROR, "", "Failed to delete product")
	}

	logging.Debug.Printf("Product deleted: %s", product.Name)
	return nil
}

// UpdateProduct updates product details
func (s *ProductService) UpdateProduct(id string, input *UpdateProductInput) (*model.Product, error) {
	logging.Debug.Printf("Updating product: %s", id)

	var product model.Product
	if err := s.repo.FindByIdWithPreload(&product, id, "Sizes"); err != nil {
		logging.Error.Printf("Product to update not found: %s", id)
		return nil, apperror.New(constant.NOTFOUND, "", "Product not found")
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

	// Apply product updates
	if len(updates) > 0 {
		if err := s.repo.UpdateByFields(&model.Product{}, id, updates); err != nil {
			logging.Error.Printf("Failed to update product %s: %v", id, err)
			return nil, apperror.New(constant.INTERNALSERVERERROR, "", "Failed to update product")
		}
		logging.Debug.Printf("Product fields updated: %+v", updates)
	}

	// Update sizes if provided
	if input.Sizes != nil {
		logging.Debug.Printf("Updating %d sizes for product %s", len(*input.Sizes), id)
		
		for _, sReq := range *input.Sizes {
			if sReq.ID != nil {
				// Update existing size
				fields := map[string]interface{}{
					"size":     sReq.Size,
					"quantity": sReq.Quantity,
				}
				if err := s.repo.UpdateByFields(&model.ProductSize{}, *sReq.ID, fields); err != nil {
					logging.Error.Printf("Failed to update size %s: %v", *sReq.ID, err)
					return nil, apperror.New(constant.INTERNALSERVERERROR, "", "Failed to update product size")
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
				logging.Error.Printf("Failed to add size %s: %v", sReq.Size, err)
				return nil, apperror.New(constant.INTERNALSERVERERROR, "", "Failed to add product size")
			}
		}
		logging.Debug.Println("Product sizes updated successfully")
	}

	// Reload updated product
	if err := s.repo.FindByIdWithPreload(&product, id, "Sizes"); err != nil {
		logging.Error.Printf("Failed to reload product %s: %v", id, err)
		return nil, apperror.New(constant.INTERNALSERVERERROR, "", "Failed to fetch updated product")
	}

	logging.Debug.Printf("Product updated successfully: %s", product.Name)
	return &product, nil
}

// SearchProducts searches products with criteria
func (s *ProductService) SearchProducts(
	query string,
	league string,
	kitType string,
	year *int,
) ([]model.Product, error) {
	logging.Debug.Printf("SearchProducts: query=%s, league=%s, kitType=%s", query, league, kitType)

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

	logging.Debug.Printf("Search query: %s, args: %v", dbQuery, args)

	if err := s.repo.FindWhereWithPreload(&products, dbQuery, args, "Sizes"); err != nil {
		logging.Error.Printf("SearchProducts failed: %v", err)
		return nil, apperror.New(constant.INTERNALSERVERERROR, "", "Failed to fetch products")
	}

	logging.Debug.Printf("Search found %d products", len(products))
	return products, nil
}