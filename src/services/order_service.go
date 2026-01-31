// service/order_service.go
package services

import (
	// "fmt" // Uncomment if you add stock checks back
	"errors"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlaceOrderRequest struct {
	Type      string `json:"type" validate:"required,oneof=cart direct"` // "cart" or "direct"
	ProductID string `json:"product_id"`                                 // for direct orders
	Quantity  int    `json:"quantity"`                                   // for direct orders
	Size      string `json:"size"`                                       // for direct orders
	AddressID string `json:"address_id" validate:"omitempty,uuid"`
}

type OrderService struct {
	repo repo.IPgSQLRepository
}

func NewOrderService(r repo.IPgSQLRepository) *OrderService {
	return &OrderService{repo: r}
}

func (s *OrderService) PlaceOrder(userID string, req PlaceOrderRequest) (*model.Order, error) {
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

	// Defer rollback in case of panic or error
	defer func() {
		if r := recover(); r != nil {
			s.repo.Rollback(tx)
		}
	}()

	var orderItems []model.OrderItem
	var total int

	switch req.Type {
	case "direct":
		items, calcTotal, err := s.processDirectOrder(tx, req)
		if err != nil {
			s.repo.Rollback(tx)
			return nil, err
		}
		orderItems = items
		total = calcTotal
	case "cart":
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
		UserID: uID,
		// AddressID: addressID, // Uncomment if your Order model has this field
		Total:  total,
		Status: constant.PLACED,
	}

	if err := tx.Create(&order).Error; err != nil {
		s.repo.Rollback(tx)
		return nil, apperror.New(constant.INTERNALSERVERERROR, "ORDER_CREATE_ERROR", "Failed to create order")
	}

	// Create Order Items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID

		if err := tx.Create(&orderItems[i]).Error; err != nil {
			s.repo.Rollback(tx)
			return nil, apperror.New(constant.INTERNALSERVERERROR, "ORDER_ITEM_ERROR", "Failed to create order items")
		}

		// TODO: Add stock management here once you add Stock field to Product model
		// Uncomment below lines after adding `Stock int` to model.Product:
		/*
			if err := tx.Model(&model.Product{}).
				Where("id = ?", orderItems[i].ProductID).
				UpdateColumn("stock", gorm.Expr("stock - ?", orderItems[i].Quantity)).Error; err != nil {
				s.repo.Rollback(tx)
				return nil, apperror.New(constant.INTERNALSERVERERROR, "STOCK_UPDATE_ERROR", "Failed to update inventory")
			}
		*/
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

	// Fetch complete order with relations
	var fullOrder model.Order
	if err := s.repo.FindByIdWithPreload(&fullOrder, order.ID, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "FETCH_ERROR", "Order created but failed to fetch details")
	}

	return &fullOrder, nil
}

func (s *OrderService) processDirectOrder(tx *gorm.DB, req PlaceOrderRequest) ([]model.OrderItem, int, error) {
	if req.ProductID == "" || req.Quantity <= 0 {
		return nil, 0, apperror.New(constant.BADREQUEST, "INVALID_PRODUCT", "Product ID and valid quantity required")
	}

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

	// TODO: Add stock validation here once Stock field is added to Product model
	// Uncomment below lines after adding `Stock int` to model.Product:
	/*
		if product.Stock < req.Quantity {
			return nil, 0, apperror.New(constant.BADREQUEST, "INSUFFICIENT_STOCK",
				fmt.Sprintf("Only %d items available in stock", product.Stock))
		}
	*/

	item := model.OrderItem{
		ProductID: productID,
		Size:      req.Size,
		Quantity:  req.Quantity,
		Price:     product.Price,
	}

	total := product.Price * req.Quantity

	return []model.OrderItem{item}, total, nil
}

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
		// TODO: Add stock validation here once Stock field is added to Product model
		// Uncomment below lines after adding `Stock int` to model.Product:
		/*
			if item.Product.Stock < item.Quantity {
				return nil, 0, apperror.New(constant.BADREQUEST, "INSUFFICIENT_STOCK",
					"Insufficient stock for product: "+item.Product.Name)
			}
		*/

		orderItems = append(orderItems, model.OrderItem{
			ProductID: item.ProductID,
			Size:      item.Size,
			Quantity:  item.Quantity,
			Price:     item.Product.Price,
		})

		total += item.Product.Price * item.Quantity
	}

	return orderItems, total, nil
}

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

	// Load relations
	if err := s.repo.FindByIdWithPreload(&order, order.ID, "Items.Product"); err != nil {
		return nil, apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to load order details")
	}

	return &order, nil
}

// service/order_service.go - Add these methods

func (s *OrderService) UpdateOrderStatusUser(userID string, orderID string, status string) error {
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
			return apperror.New(constant.NOTFOUND, "ORDER_NOT_FOUND", "Order not found or does not belong to user")
		}
		return apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch order")
	}

	// Users can only cancel orders, not change to other statuses
	if status != constant.CANCELLED {
		return apperror.New(constant.BADREQUEST, "INVALID_STATUS", "Users can only cancel orders")
	}

	// Can only cancel if order is still in PLACED status
	if order.Status != constant.PLACED {
		return apperror.New(constant.BADREQUEST, "INVALID_OPERATION", "Cannot cancel order that is already processed or shipped")
	}

	if err := s.repo.UpdateByFields(&order, order.ID, map[string]interface{}{
		"status": status,
	}); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "UPDATE_ERROR", "Failed to update order status")
	}

	return nil
}

func (s *OrderService) CancelOrder(userID string, orderID string) error {
	// Essentially the same as updating status to cancelled
	return s.UpdateOrderStatusUser(userID, orderID, constant.CANCELLED)
}

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
			return apperror.New(constant.NOTFOUND, "ORDER_NOT_FOUND", "Order not found or does not belong to user")
		}
		return apperror.New(constant.INTERNALSERVERERROR, "DB_ERROR", "Failed to fetch order")
	}

	// Only allow deletion of cancelled orders or placed orders (not shipped/delivered)
	if order.Status != constant.PLACED && order.Status != constant.CANCELLED {
		return apperror.New(constant.BADREQUEST, "INVALID_OPERATION", "Cannot delete order that is being processed or already shipped")
	}

	// Delete order items first (foreign key constraint)
	if err := s.repo.Exec("DELETE FROM order_items WHERE order_id = ?", oID).Error; err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "DELETE_ERROR", "Failed to delete order items")
	}

	// Delete order
	if err := s.repo.Delete(&model.Order{}, oID); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "DELETE_ERROR", "Failed to delete order")
	}

	return nil
}

// service/order_service.go - Add these admin methods

func (s *OrderService) GetAllOrders(statusFilter, userIDFilter string) ([]model.Order, error) {
	var orders []model.Order
	var err error

	// Build query based on filters
	if statusFilter != "" && userIDFilter != "" {
		// Both filters provided
		uID, err := uuid.Parse(userIDFilter)
		if err != nil {
			return nil, apperror.New(constant.BADREQUEST, "INVALID_USER_ID", "Invalid user ID filter")
		}
		err = s.repo.FindWhereWithPreload(&orders, "status = ? AND user_id = ?", []interface{}{statusFilter, uID}, "Items.Product", "User")
	} else if statusFilter != "" {
		// Only status filter
		err = s.repo.FindWhereWithPreload(&orders, "status = ?", []interface{}{statusFilter}, "Items.Product", "User")
	} else if userIDFilter != "" {
		// Only user filter
		uID, err := uuid.Parse(userIDFilter)
		if err != nil {
			return nil, apperror.New(constant.BADREQUEST, "INVALID_USER_ID", "Invalid user ID filter")
		}
		err = s.repo.FindWhereWithPreload(&orders, "user_id = ?", []interface{}{uID}, "Items.Product", "User")
	} else {
		// No filters - get all
		err = s.repo.FindAllWithPreload(&orders, "Items.Product", "User")
	}

	if err != nil {
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

	// Validate status transition (optional business logic)
	validTransitions := map[string][]string{
		// constant.PLACED:     {constant.PROCESSING, constant.CANCELLED},
		// constant.PROCESSING: {constant.SHIPPED, constant.CANCELLED},
		constant.SHIPPED:   {constant.DELIVERED},
		constant.DELIVERED: {},
		constant.CANCELLED: {},
	}

	allowedStatuses, exists := validTransitions[order.Status]
	if !exists {
		return apperror.New(constant.BADREQUEST, "INVALID_STATUS", "Current order status is invalid")
	}

	// Check if the new status is allowed from current status
	isValid := false
	for _, allowed := range allowedStatuses {
		if allowed == status {
			isValid = true
			break
		}
	}

	// Admin can force status update even if not in valid transitions,
	// but we warn about it. Remove this check if admin should have full freedom.
	if !isValid && order.Status != status {
		// If you want strict validation, uncomment below:
		// return apperror.New(constant.BADREQUEST, "INVALID_TRANSITION",
		//     fmt.Sprintf("Cannot transition from %s to %s", order.Status, status))
	}

	if err := s.repo.UpdateByFields(&order, order.ID, map[string]interface{}{
		"status": status,
	}); err != nil {
		return apperror.New(constant.INTERNALSERVERERROR, "UPDATE_ERROR", "Failed to update order status")
	}

	return nil
}
