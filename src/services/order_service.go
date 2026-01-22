package services

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/google/uuid"
)

type OrderService struct {
	repo repo.IPgSQLRepository
}

func NewOrderService(repo repo.IPgSQLRepository) *OrderService {
	return &OrderService{repo: repo}
}

/* =======================
   PLACE ORDER
   ======================= */

func (s *OrderService) PlaceOrder(userID string) (*model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid user ID",
		)
	}

	var cart model.Cart
	if err := s.repo.FindOneWhere(&cart, "user_id = ?", uID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Cart not found",
		)
	}

	var cartItems []model.CartItem
	if err := s.repo.FindWhereWithPreload(&cartItems, "cart_id = ?", []interface{}{cart.ID}, "Product"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to fetch cart items",
		)
	}

	if len(cartItems) == 0 {
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Cart is empty",
		)
	}

	total := 0
	for _, item := range cartItems {
		total += item.Product.Price * item.Quantity
	}

	order := model.Order{
		UserID: uID,
		Total:  total,
		Status: constant.PLACED,
	}

	if err := s.repo.Insert(&order); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to create order",
		)
	}

	for _, item := range cartItems {
		orderItem := model.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Size:      item.Size,
			Quantity:  item.Quantity,
			Price:     item.Product.Price,
		}
		if err := s.repo.Insert(&orderItem); err != nil {
			return nil, apperror.New(
				constant.INTERNALSERVERERROR,
				"",
				"Failed to create order items",
			)
		}
	}

	if err := s.repo.Exec("DELETE FROM cart_items WHERE cart_id = ?", cart.ID).Error; err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to clear cart",
		)
	}

	var fullOrder model.Order
	if err := s.repo.FindByIdWithPreload(&fullOrder, order.ID, "Items.Product"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to fetch created order",
		)
	}

	return &fullOrder, nil
}

/* =======================
   GET ORDERS
   ======================= */

func (s *OrderService) GetOrdersByUser(userID string) ([]model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid user ID",
		)
	}

	var orders []model.Order
	if err := s.repo.FindWhereWithPreload(&orders, "user_id = ?", []interface{}{uID}, "Items.Product"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to fetch user orders",
		)
	}

	return orders, nil
}

func (s *OrderService) GetAllOrders() ([]model.Order, error) {
	var orders []model.Order
	if err := s.repo.FindWhereWithPreload(&orders, "1=1", []interface{}{}, "Items.Product"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to fetch all orders",
		)
	}
	return orders, nil
}

/* =======================
   GET ORDER BY ID
   ======================= */

func (s *OrderService) GetOrderByID(userID, orderID string) (*model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid user ID",
		)
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid order ID",
		)
	}

	var order model.Order
	if err := s.repo.FindByIdWithPreload(&order, oID, "Items.Product"); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Order not found",
		)
	}

	if order.UserID != uID {
		return nil, apperror.New(
            constant.UNAUTHORIZED, 
            "", 
            "Not authorized to view this order",
        )
	}

	return &order, nil
}

/* =======================
   UPDATE ORDER STATUS
   ======================= */

func (s *OrderService) UpdateOrderStatus(orderID, status string) (*model.Order, error) {
	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, apperror.New(
            constant.BADREQUEST, 
            "", 
            "Invalid order ID",
        )
	}

	validStatuses := map[string]bool{constant.PLACED: true, constant.SHIPPED: true, constant.DELIVERED: true, constant.CANCELLED: true}
	if !validStatuses[status] {
		return nil, apperror.New(
            constant.BADREQUEST,
             "",
              "Invalid order status",
            )
	}

	var order model.Order
	if err := s.repo.FindById(&order, oID); err != nil {
		return nil, apperror.New(
            constant.NOTFOUND,
             "", 
             "Order not found",
            )
	}

	if err := s.repo.UpdateByFields(&order, oID, map[string]interface{}{"status": status}); err != nil {
		return nil, apperror.New(
            constant.INTERNALSERVERERROR,
             "",
              "Failed to update order status",
            )
	}

	if err := s.repo.FindByIdWithPreload(&order, oID, "Items.Product"); err != nil {
		return nil, apperror.New(
            constant.INTERNALSERVERERROR, 
            "", 
            "Failed to reload order",
        )
	}

	return &order, nil
}

/* =======================
   CANCEL ORDER
   ======================= */

func (s *OrderService) CancelOrder(userID, orderID string) (*model.Order, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
            constant.BADREQUEST,
             "",
              "Invalid user ID",
            )
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, apperror.New(
            constant.BADREQUEST, 
            "", 
            "Invalid order ID",
        )
	}

	var order model.Order
	if err := s.repo.FindById(&order, oID); err != nil {
		return nil, apperror.New(
            constant.NOTFOUND,
             "", 
             "Order not found",
            )
	}

	if order.UserID != uID {
		return nil, apperror.New(
            constant.UNAUTHORIZED, 
            "", 
            "Not authorized to cancel this order",
        )
	}

	if order.Status != constant.PLACED {
		return nil, apperror.New(
            constant.BADREQUEST, 
            "",
             "Only placed orders can be cancelled",
            
            )
	}

	if err := s.repo.UpdateByFields(&order, oID, map[string]interface{}{"status": "CANCELLED"}); err != nil {
		return nil, apperror.New(
            constant.INTERNALSERVERERROR, 
            "", 
            "Failed to cancel order",
        )
	}

	if err := s.repo.FindByIdWithPreload(&order, oID, "Items.Product"); err != nil {
		return nil, apperror.New(
            constant.INTERNALSERVERERROR,
             "",
              "Failed to reload order",
            )
	}

	return &order, nil
}

/* =======================
   DELETE ORDER
   ======================= */

func (s *OrderService) DeleteOrder(userID, orderID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(
            constant.BADREQUEST, 
            "",
             "Invalid user ID",
            )
	}

	oID, err := uuid.Parse(orderID)
	if err != nil {
		return apperror.New(
            constant.BADREQUEST, 
            "", 
            "Invalid order ID",
        )
	}

	var order model.Order
	if err := s.repo.FindById(&order, oID); err != nil {
		return apperror.New(
            constant.NOTFOUND,
             "", 
             "Order not found",
            )
	}

	if order.UserID != uID {
		return apperror.New(
            constant.UNAUTHORIZED, 
            "", 
            "Not authorized to delete this order",
        )
	}

	if order.Status != constant.PLACED && order.Status != constant.CANCELLED {
		return apperror.New(
            constant.BADREQUEST,
             "",
              "Order cannot be deleted at this stage",
            )
	}

	if err := s.repo.Exec("DELETE FROM order_items WHERE order_id = ?", oID).Error; err != nil {
		return apperror.New(
            constant.INTERNALSERVERERROR,
             "",
              "Failed to delete order items",
            )
	}

	if err := s.repo.Delete(&model.Order{}, oID); err != nil {
		return apperror.New(
            constant.INTERNALSERVERERROR, 
            "", 
            "Failed to delete order",
        )
	}

	return nil
}



func (s *OrderService) UpdateOrderStatusByID(
	userID string,
	orderID string,
	status string,
	isAdmin bool,
) (*model.Order, error) {

	// Validate order ID
	oID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid order id",
		)
	}

	// Validate status
	validStatuses := map[string]bool{
		"PLACED":    true,
		"SHIPPED":   true,
		"DELIVERED": true,
		"CANCELLED": true,
	}

	if !validStatuses[status] {
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid order status",
		)
	}

	// Fetch order
	var order model.Order
	if err := s.repo.FindByIdWithPreload(&order, oID, "Items.Product"); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"Order not found",
		)
	}

	// USER restrictions
	if !isAdmin {
		uID, err := uuid.Parse(userID)
		if err != nil {
			return nil, apperror.New(
				constant.BADREQUEST,
				"",
				"Invalid user id",
			)
		}

		if order.UserID != uID {
			return nil, apperror.New(
				constant.FORBIDDEN,
				"",
				"You are not allowed to update this order",
			)
		}

		// Users can ONLY cancel
		if status != constant.CANCELLED {
			return nil, apperror.New(
				constant.FORBIDDEN,
				"",
				"Users can only cancel orders",
			)
		}
	}

	// Update status
	if err := s.repo.UpdateByFields(
		&model.Order{},
		oID,
		map[string]interface{}{"status": status},
	); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to update order status",
		)
	}

	// Reload updated order
	if err := s.repo.FindByIdWithPreload(&order, oID, "Items.Product"); err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to load updated order",
		)
	}

	return &order, nil
}
