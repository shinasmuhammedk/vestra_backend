package services

import (
	"errors"

	"github.com/google/uuid"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
)

type OrderService struct {
	repo repo.IPgSQLRepository
}

func NewOrderService(repo repo.IPgSQLRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) PlaceOrder(userID string) (*model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	// 1️⃣ Get cart
	var cart model.Cart
	if err := s.repo.FindOneWhere(&cart, "user_id = ?", uID); err != nil {
		return nil, errors.New("cart not found")
	}

	// 2️⃣ Get cart items with Product
	var cartItems []model.CartItem
	if err := s.repo.FindWhereWithPreload(&cartItems, "cart_id = ?", []interface{}{cart.ID}, "Product"); err != nil {
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, errors.New("cart is empty")
	}

	// 3️⃣ Calculate total
	total := 0
	for _, item := range cartItems {
		total += item.Product.Price * item.Quantity
	}

	// 4️⃣ Create order
	order := model.Order{
		UserID: uID,
		Total:  total,
		Status: "PLACED",
	}

	if err := s.repo.Insert(&order); err != nil {
		return nil, err
	}

	// 5️⃣ Create order items
	for _, item := range cartItems {
		orderItem := model.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Size:      item.Size,
			Quantity:  item.Quantity,
			Price:     item.Product.Price,
		}
		if err := s.repo.Insert(&orderItem); err != nil {
			return nil, err
		}
	}

	// 6️⃣ Clear cart
	if err := s.repo.Exec("DELETE FROM cart_items WHERE cart_id = ?", cart.ID).Error; err != nil {
		return nil, err
	}

	// 7️⃣ Reload order with items and product info
	var fullOrder model.Order
	if err := s.repo.FindByIdWithPreload(&fullOrder, order.ID, "Items.Product"); err != nil {
		return nil, err
	}

	return &fullOrder, nil
}

func (s *OrderService) GetOrdersByUser(userID string) ([]model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	var orders []model.Order
	if err := s.repo.FindWhereWithPreload(&orders, "user_id = ?", []interface{}{uID}, "Items.Product"); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *OrderService) GetAllOrders() ([]model.Order, error) {
	var orders []model.Order
	// Use repo method that allows Preload
	if err := s.repo.FindWhereWithPreload(&orders, "1=1", []interface{}{}, "Items.Product"); err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) UpdateOrderStatus(orderID string, status string) (*model.Order, error) {
	// Validate UUID
	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, errors.New("invalid order id")
	}

	// Validate status
	validStatuses := map[string]bool{
		"PLACED":    true,
		"SHIPPED":   true,
		"DELIVERED": true,
		"CANCELLED": true,
	}
	if !validStatuses[status] {
		return nil, errors.New("invalid status")
	}

	// Fetch order
	var order model.Order
	if err := s.repo.FindById(&order, oID); err != nil {
		return nil, errors.New("order not found")
	}

	// Update status
	if err := s.repo.UpdateByFields(&order, oID, map[string]interface{}{"status": status}); err != nil {
		return nil, err
	}

	// Reload order with items & product
	var fullOrder model.Order
	if err := s.repo.FindByIdWithPreload(&fullOrder, oID, "Items.Product"); err != nil {
		return nil, err
	}

	return &fullOrder, nil
}

// UpdateOrderStatusByID updates an order's status.
// If isAdmin is false, user can only update their own orders (optional: e.g., cancel).
func (s *OrderService) UpdateOrderStatusByID(userID string, orderID string, status string, isAdmin bool) (*model.Order, error) {
	// Validate UUIDs
	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, errors.New("invalid order id")
	}

	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	// Validate status
	validStatuses := map[string]bool{
		"PLACED":    true,
		"SHIPPED":   true,
		"DELIVERED": true,
		"CANCELLED": true,
	}
	if !validStatuses[status] {
		return nil, errors.New("invalid status")
	}

	// Fetch order
	var order model.Order
	if err := s.repo.FindByIdWithPreload(&order, oID, "Items.Product"); err != nil {
		return nil, errors.New("order not found")
	}

	// Check ownership if not admin
	if !isAdmin && order.UserID != uID {
		return nil, errors.New("not authorized to update this order")
	}

	// Optional: Users can only cancel their own orders
	if !isAdmin && status != "CANCELLED" {
		return nil, errors.New("users can only cancel their own orders")
	}

	// Update status
	if err := s.repo.UpdateByFields(&order, oID, map[string]interface{}{"status": status}); err != nil {
		return nil, err
	}

	// Reload order with items
	if err := s.repo.FindByIdWithPreload(&order, oID, "Items.Product"); err != nil {
		return nil, err
	}

	return &order, nil
}



func (s *OrderService) GetOrderByID(userID string, orderID string) (*model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, errors.New("invalid order id")
	}

	var order model.Order
	if err := s.repo.FindByIdWithPreload(&order, oID, "Items.Product"); err != nil {
		return nil, errors.New("order not found")
	}

	// Ensure the order belongs to this user
	if order.UserID != uID {
		return nil, errors.New("unauthorized to view this order")
	}

	return &order, nil
}



func (s *OrderService) DeleteOrder(userID string, orderID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user id")
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return errors.New("invalid order id")
	}

	// Fetch order
	var order model.Order
	if err := s.repo.FindById(&order, oID); err != nil {
		return errors.New("order not found")
	}

	// Ownership check
	if order.UserID != uID {
		return errors.New("not authorized to delete this order")
	}

	// Status check
	if order.Status != "PLACED" && order.Status != "CANCELLED" {
		return errors.New("order cannot be deleted at this stage")
	}

	// Delete order items first (FK safety)
	if err := s.repo.Exec(
		"DELETE FROM order_items WHERE order_id = ?",
		oID,
	).Error; err != nil {
		return err
	}

	// Delete order
	if err := s.repo.Delete(&model.Order{}, oID); err != nil {
		return err
	}

	return nil
}




func (s *OrderService) CancelOrder(userID string, orderID string) (*model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, errors.New("invalid order id")
	}

	var order model.Order
	if err := s.repo.FindById(&order, oID); err != nil {
		return nil, errors.New("order not found")
	}

	// Ownership check
	if order.UserID != uID {
		return nil, errors.New("not authorized to cancel this order")
	}

	// Only PLACED orders can be cancelled
	if order.Status != "PLACED" {
		return nil, errors.New("only placed orders can be cancelled")
	}

	// Update status to CANCELLED
	if err := s.repo.UpdateByFields(&order, oID, map[string]interface{}{
		"status": "CANCELLED",
	}); err != nil {
		return nil, err
	}

	// Reload order with items & products
	if err := s.repo.FindByIdWithPreload(&order, oID, "Items.Product"); err != nil {
		return nil, err
	}

	return &order, nil
}
