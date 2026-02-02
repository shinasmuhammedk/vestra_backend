package services

import (
	"errors"
	"fmt"
	"time"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlaceOrderRequest struct {
	Type      string `json:"type" validate:"required,oneof=cart direct"`
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Size      string `json:"size"`
	AddressID string `json:"address_id" validate:"required,uuid"`
}

type OrderService struct {
	repo repo.IPgSQLRepository
}

func NewOrderService(r repo.IPgSQLRepository) *OrderService {
	return &OrderService{repo: r}
}

// PlaceOrder handles both cart checkout and direct buy
func (s *OrderService) PlaceOrder(userID string, req PlaceOrderRequest) (*model.Order, error) {
	// DEBUG: Log what we received
	fmt.Printf("DEBUG PlaceOrder START: UserID=%s, Type=%s, ProductID=%s, Size=%s, Qty=%d\n", 
		userID, req.Type, req.ProductID, req.Size, req.Quantity)

	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(constant.BADREQUEST, "INVALID_USER_ID", "Invalid user ID format")
	}

	addressID, err := uuid.Parse(req.AddressID)
	if err != nil {
		return nil, apperror.New(constant.BADREQUEST, "INVALID_ADDRESS_ID", "Invalid address ID format")
	}

	// Verify address belongs to user
	var address model.UserAddress
	if err := s.repo.FindOneWhere(&address, "id = ? AND user_id = ?", addressID, uID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(constant.NOTFOUND, "ADDRESS_NOT_FOUND", "Address not found or does not belong to user")
		}
		return nil, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to verify address")
	}

	// Begin Transaction
	tx := s.repo.Begin()
	if tx.Error != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "TRANSACTION_ERROR", "Failed to start transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			s.repo.Rollback(tx)
		}
	}()

	var orderItems []model.OrderItem
	var total int

	switch req.Type {
	case "direct":
		// Validate direct order inputs
		if req.Size == "" {
			s.repo.Rollback(tx)
			return nil, apperror.New(constant.BADREQUEST, "SIZE_REQUIRED", "Size is required for direct orders")
		}
		if req.ProductID == "" {
			s.repo.Rollback(tx)
			return nil, apperror.New(constant.BADREQUEST, "PRODUCT_REQUIRED", "Product ID is required for direct orders")
		}
		if req.Quantity <= 0 {
			s.repo.Rollback(tx)
			return nil, apperror.New(constant.BADREQUEST, "QUANTITY_REQUIRED", "Quantity must be greater than 0")
		}

		fmt.Printf("DEBUG: Processing DIRECT order for Size=%s\n", req.Size)
		items, calcTotal, err := s.processDirectOrder(tx, req)
		if err != nil {
			s.repo.Rollback(tx)
			return nil, err
		}
		orderItems = items
		total = calcTotal

	case "cart":
		fmt.Printf("DEBUG: Processing CART order\n")
		items, calcTotal, err := s.processCartOrder(tx, uID)
		if err != nil {
			s.repo.Rollback(tx)
			return nil, err
		}
		orderItems = items
		total = calcTotal

	default:
		s.repo.Rollback(tx)
		return nil, apperror.New(constant.BADREQUEST, "INVALID_ORDER_TYPE", "Order type must be 'cart' or 'direct'")
	}

	// Create Order
	order := model.Order{
		ID:        uuid.New(),
		UserID:    uID,
		Total:     total,
		Status:    constant.PLACED,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Create(&order).Error; err != nil {
		s.repo.Rollback(tx)
		return nil, apperror.New(constant.INTERNALSERVERERROR, "ORDER_CREATE_ERROR", "Failed to create order")
	}

	// Create Order Items
	for i := range orderItems {
		orderItems[i].ID = uuid.New()
		orderItems[i].OrderID = order.ID
		orderItems[i].CreatedAt = time.Now()
		orderItems[i].UpdatedAt = time.Now()

		if err := tx.Create(&orderItems[i]).Error; err != nil {
			s.repo.Rollback(tx)
			return nil, apperror.New(constant.INTERNALSERVERERROR, "ORDER_ITEM_ERROR", "Failed to create order items")
		}
	}

	// Clear cart if cart order
	if req.Type == "cart" {
		if err := tx.Exec("DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM carts WHERE user_id = ?)", uID).Error; err != nil {
			s.repo.Rollback(tx)
			return nil, apperror.New(constant.INTERNALSERVERERROR, "CART_CLEAR_ERROR", "Failed to clear cart")
		}
	}

	// Commit transaction
	if err := s.repo.Commit(tx); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "COMMIT_ERROR", "Failed to finalize order")
	}

	fmt.Printf("DEBUG: Order placed successfully! OrderID=%s\n", order.ID)

	// Fetch complete order with relations
	var fullOrder model.Order
	if err := s.repo.FindByIdWithPreload(&fullOrder, order.ID, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "FETCH_ERROR", "Order created but failed to fetch details")
	}

	return &fullOrder, nil
}

// processDirectOrder handles Buy Now functionality with stock deduction
func (s *OrderService) processDirectOrder(tx *gorm.DB, req PlaceOrderRequest) ([]model.OrderItem, int, error) {
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, 0, apperror.New(constant.BADREQUEST, "INVALID_PRODUCT_ID", "Invalid product ID format")
	}

	var product model.Product
	if err := tx.First(&product, "id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperror.New(constant.NOTFOUND, "PRODUCT_NOT_FOUND", "Product not found")
		}
		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch product")
	}

	// CRITICAL: Check stock BEFORE any deduction
	var productSize model.ProductSize
	if err := tx.Where("product_id = ? AND size = ?", productID, req.Size).First(&productSize).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperror.New(constant.BADREQUEST, "SIZE_NOT_FOUND", 
				fmt.Sprintf("Size %s is not available for this product", req.Size))
		}
		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to check stock")
	}

	fmt.Printf("DEBUG processDirectOrder: Found size %s with quantity %d, ordering %d\n", 
		productSize.Size, productSize.Quantity, req.Quantity)

	// Validate sufficient stock
	if productSize.Quantity < req.Quantity {
		return nil, 0, apperror.New(constant.BADREQUEST, "INSUFFICIENT_STOCK",
			fmt.Sprintf("Only %d items available in size %s (you requested %d)", 
				productSize.Quantity, req.Size, req.Quantity))
	}

	// DEDUCT STOCK - Use the ID from the fetched record
	if err := tx.Model(&model.ProductSize{}).
		Where("id = ?", productSize.ID).
		UpdateColumn("quantity", gorm.Expr("quantity - ?", req.Quantity)).
		UpdateColumn("updated_at", time.Now()).Error; err != nil {
		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "STOCK_UPDATE_ERROR", "Failed to deduct inventory")
	}

	fmt.Printf("DEBUG: Deducted %d from size %s. New quantity: %d\n", 
		req.Quantity, req.Size, productSize.Quantity - req.Quantity)

	item := model.OrderItem{
		ID:        uuid.New(),
		ProductID: productID,
		Size:      req.Size,
		Quantity:  req.Quantity,
		Price:     product.Price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	total := product.Price * req.Quantity

	return []model.OrderItem{item}, total, nil
}

// processCartOrder handles cart checkout with stock deduction for each item
func (s *OrderService) processCartOrder(tx *gorm.DB, userID uuid.UUID) ([]model.OrderItem, int, error) {
	var cart model.Cart
	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, apperror.New(constant.NOTFOUND, "CART_NOT_FOUND", "Cart not found")
		}
		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch cart")
	}

	var cartItems []model.CartItem
	if err := tx.Preload("Product").Where("cart_id = ?", cart.ID).Find(&cartItems).Error; err != nil {
		return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch cart items")
	}

	if len(cartItems) == 0 {
		return nil, 0, apperror.New(constant.BADREQUEST, "EMPTY_CART", "Cart is empty")
	}

	var orderItems []model.OrderItem
	total := 0

	for _, item := range cartItems {
		// Check and deduct stock for each cart item
		var productSize model.ProductSize
		if err := tx.Where("product_id = ? AND size = ?", item.ProductID, item.Size).First(&productSize).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, apperror.New(constant.BADREQUEST, "SIZE_NOT_FOUND",
					fmt.Sprintf("Size %s is no longer available for %s", item.Size, item.Product.Name))
			}
			return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to check stock")
		}

		if productSize.Quantity < item.Quantity {
			return nil, 0, apperror.New(constant.BADREQUEST, "INSUFFICIENT_STOCK",
				fmt.Sprintf("Only %d items available in size %s for %s. Please reduce quantity or remove item.",
					productSize.Quantity, item.Size, item.Product.Name))
		}

		// Deduct stock
		if err := tx.Model(&model.ProductSize{}).
			Where("id = ?", productSize.ID).
			UpdateColumn("quantity", gorm.Expr("quantity - ?", item.Quantity)).
			UpdateColumn("updated_at", time.Now()).Error; err != nil {
			return nil, 0, apperror.New(constant.INTERNALSERVERERROR, "STOCK_UPDATE_ERROR",
				fmt.Sprintf("Failed to update stock for %s", item.Product.Name))
		}

		orderItems = append(orderItems, model.OrderItem{
			ID:        uuid.New(),
			ProductID: item.ProductID,
			Size:      item.Size,
			Quantity:  item.Quantity,
			Price:     item.Product.Price,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

		total += item.Product.Price * item.Quantity
	}

	return orderItems, total, nil
}

// GetUserOrders retrieves all orders for a user
func (s *OrderService) GetUserOrders(userID string) ([]model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(constant.BADREQUEST, "INVALID_USER_ID", "Invalid user ID")
	}

	var orders []model.Order
	if err := s.repo.FindWhereWithPreload(&orders, "user_id = ?", []interface{}{uID}, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch orders")
	}

	return orders, nil
}

// GetOrderDetails retrieves a specific order with items
func (s *OrderService) GetOrderDetails(userID string, orderID string) (*model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(constant.BADREQUEST, "INVALID_USER_ID", "Invalid user ID")
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, apperror.New(constant.BADREQUEST, "INVALID_ORDER_ID", "Invalid order ID")
	}

	var order model.Order
	if err := s.repo.FindOneWhere(&order, "id = ? AND user_id = ?", oID, uID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(constant.NOTFOUND, "ORDER_NOT_FOUND", "Order not found")
		}
		return nil, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch order")
	}

	if err := s.repo.FindByIdWithPreload(&order, order.ID, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to load order details")
	}

	return &order, nil
}

// CancelOrder cancels an order and restores stock
func (s *OrderService) CancelOrder(userID string, orderID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "INVALID_USER_ID", "Invalid user ID")
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "INVALID_ORDER_ID", "Invalid order ID")
	}

	tx := s.repo.Begin()
	if tx.Error != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "TRANSACTION_ERROR", "Failed to start transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			s.repo.Rollback(tx)
		}
	}()

	var order model.Order
	if err := tx.First(&order, "id = ? AND user_id = ?", oID, uID).Error; err != nil {
		s.repo.Rollback(tx)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(constant.NOTFOUND, "ORDER_NOT_FOUND", "Order not found")
		}
		return apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch order")
	}

	// Can only cancel if not already delivered/cancelled
	if order.Status == constant.DELIVERED || order.Status == constant.CANCELLED {
		s.repo.Rollback(tx)
		return apperror.New(constant.BADREQUEST, "INVALID_OPERATION", "Cannot cancel delivered or already cancelled orders")
	}

	// Restore stock for all items
	var orderItems []model.OrderItem
	if err := tx.Where("order_id = ?", order.ID).Find(&orderItems).Error; err != nil {
		s.repo.Rollback(tx)
		return apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch order items")
	}

	for _, item := range orderItems {
		if err := tx.Model(&model.ProductSize{}).
			Where("product_id = ? AND size = ?", item.ProductID, item.Size).
			UpdateColumn("quantity", gorm.Expr("quantity + ?", item.Quantity)).
			UpdateColumn("updated_at", time.Now()).Error; err != nil {
			s.repo.Rollback(tx)
			return apperror.New(constant.INTERNALSERVERERROR, "STOCK_RESTORE_ERROR", "Failed to restore inventory")
		}
	}

	// Update order status
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":     constant.CANCELLED,
		"updated_at": time.Now(),
	}).Error; err != nil {
		s.repo.Rollback(tx)
		return apperror.New(constant.INTERNALSERVERERROR, "UPDATE_ERROR", "Failed to cancel order")
	}

	if err := s.repo.Commit(tx); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "COMMIT_ERROR", "Failed to finalize cancellation")
	}

	return nil
}

// DeleteOrder deletes a cancelled order
func (s *OrderService) DeleteOrder(userID string, orderID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "INVALID_USER_ID", "Invalid user ID")
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "INVALID_ORDER_ID", "Invalid order ID")
	}

	var order model.Order
	if err := s.repo.FindOneWhere(&order, "id = ? AND user_id = ?", oID, uID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(constant.NOTFOUND, "ORDER_NOT_FOUND", "Order not found")
		}
		return apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch order")
	}

	if order.Status != constant.CANCELLED {
		return apperror.New(constant.BADREQUEST, "INVALID_OPERATION", "Can only delete cancelled orders")
	}

	if err := s.repo.Exec("DELETE FROM order_items WHERE order_id = ?", oID).Error; err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "DELETE_ERROR", "Failed to delete order items")
	}

	if err := s.repo.Delete(&model.Order{}, oID); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "DELETE_ERROR", "Failed to delete order")
	}

	return nil
}

// Admin Functions

func (s *OrderService) GetAllOrders(statusFilter, userIDFilter string) ([]model.Order, error) {
	var orders []model.Order
	var args []interface{}
	var query string

	if statusFilter != "" && userIDFilter != "" {
		uID, err := uuid.Parse(userIDFilter)
		if err != nil {
			return nil, apperror.New(constant.BADREQUEST, "INVALID_USER_ID", "Invalid user ID filter")
		}
		query = "status = ? AND user_id = ?"
		args = []interface{}{statusFilter, uID}
	} else if statusFilter != "" {
		query = "status = ?"
		args = []interface{}{statusFilter}
	} else if userIDFilter != "" {
		uID, err := uuid.Parse(userIDFilter)
		if err != nil {
			return nil, apperror.New(constant.BADREQUEST, "INVALID_USER_ID", "Invalid user ID filter")
		}
		query = "user_id = ?"
		args = []interface{}{uID}
	} else {
		if err := s.repo.FindAllWithPreload(&orders, "Items.Product"); err != nil {
			return nil, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch orders")
		}
		return orders, nil
	}

	if err := s.repo.FindWhereWithPreload(&orders, query, args, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch orders")
	}

	return orders, nil
}

func (s *OrderService) UpdateOrderStatusAdmin(orderID string, status string) error {
	oID, err := uuid.Parse(orderID)
	if err != nil {
		return apperror.New(constant.BADREQUEST, "INVALID_ORDER_ID", "Invalid order ID")
	}

	var order model.Order
	if err := s.repo.FindById(&order, oID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(constant.NOTFOUND, "ORDER_NOT_FOUND", "Order not found")
		}
		return apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch order")
	}

	// If cancelling, restore stock
	if status == constant.CANCELLED && order.Status != constant.CANCELLED {
		tx := s.repo.Begin()
		if tx.Error != nil {
			return apperror.New(constant.INTERNALSERVERERROR, "TRANSACTION_ERROR", "Failed to start transaction")
		}

		var orderItems []model.OrderItem
		if err := tx.Where("order_id = ?", order.ID).Find(&orderItems).Error; err != nil {
			s.repo.Rollback(tx)
			return apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch order items")
		}

		for _, item := range orderItems {
			if err := tx.Model(&model.ProductSize{}).
				Where("product_id = ? AND size = ?", item.ProductID, item.Size).
				UpdateColumn("quantity", gorm.Expr("quantity + ?", item.Quantity)).
				UpdateColumn("updated_at", time.Now()).Error; err != nil {
				s.repo.Rollback(tx)
				return apperror.New(constant.INTERNALSERVERERROR, "STOCK_RESTORE_ERROR", "Failed to restore inventory")
			}
		}

		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error; err != nil {
			s.repo.Rollback(tx)
			return apperror.New(constant.INTERNALSERVERERROR, "UPDATE_ERROR", "Failed to update order status")
		}

		if err := s.repo.Commit(tx); err != nil {
			return apperror.New(constant.INTERNALSERVERERROR, "COMMIT_ERROR", "Failed to finalize update")
		}
		return nil
	}

	// Normal status update
	if err := s.repo.UpdateByFields(&order, order.ID, map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "UPDATE_ERROR", "Failed to update order status")
	}

	return nil
}

func (s *OrderService) UpdateOrderStatusUser(userID string, orderID string, status string) error {
	if status != constant.CANCELLED {
		return apperror.New(constant.BADREQUEST, "UNAUTHORIZED", "Users can only cancel orders")
	}
	return s.CancelOrder(userID, orderID)
}