package services

import (
	"github.com/google/uuid"

	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/utils/apperror"
)

type CartService struct {
	repo repo.IPgSQLRepository
}

func NewCartService(repo repo.IPgSQLRepository) *CartService {
	logging.Debug.Println("CartService initialized")
	return &CartService{repo: repo}
}

// AddToCart adds item to user's cart
func (s *CartService) AddToCart(
	userID string,
	productID string,
	size string,
	quantity int,
) error {
	logging.Debug.Printf("Adding to cart - user: %s, product: %s, size: %s, qty: %d",
		userID, productID, size, quantity)

	// Validate user UUID
	uID, err := uuid.Parse(userID)
	if err != nil {
		logging.Error.Printf("Invalid user ID: %s", userID)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"invalid user id",
		)
	}

	// Validate product UUID
	pID, err := uuid.Parse(productID)
	if err != nil {
		logging.Error.Printf("Invalid product ID: %s", productID)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"invalid product id",
		)
	}

	// Get or create cart
	var cart model.Cart
	err = s.repo.FindOneWhere(&cart, "user_id = ?", uID)
	if err != nil {
		logging.Debug.Printf("Creating new cart for user: %s", userID)
		cart = model.Cart{UserID: uID}
		if err := s.repo.Insert(&cart); err != nil {
			logging.Error.Printf("Failed to create cart for user %s: %v", userID, err)
			return apperror.ErrInternal
		}
		logging.Debug.Printf("New cart created: %s", cart.ID)
	} else {
		logging.Debug.Printf("Existing cart found: %s", cart.ID)
	}

	// Check if item exists
	var item model.CartItem
	err = s.repo.FindOneWhere(
		&item,
		"cart_id = ? AND product_id = ? AND size = ?",
		cart.ID,
		pID,
		size,
	)

	if err == nil {
		// Update quantity for existing item
		newQty := item.Quantity + quantity
		logging.Debug.Printf("Updating existing cart item %s quantity: %d -> %d",
			item.ID, item.Quantity, newQty)

		return s.repo.UpdateByFields(
			&model.CartItem{},
			item.ID,
			map[string]interface{}{
				"quantity": newQty,
			},
		)
	}

	// Add new item
	cartItem := model.CartItem{
		CartID:    cart.ID,
		ProductID: pID,
		Size:      size,
		Quantity:  quantity,
	}

	logging.Debug.Printf("Adding new cart item to cart: %s", cart.ID)
	if err := s.repo.Insert(&cartItem); err != nil {
		logging.Error.Printf("Failed to add cart item: %v", err)
		return apperror.ErrInternal
	}

	logging.Debug.Printf("Cart item added successfully: %s", cartItem.ID)
	return nil
}

// GetUserCart retrieves user's cart with items
func (s *CartService) GetUserCart(userID string) (*model.Cart, error) {
	logging.Debug.Printf("Getting cart for user: %s", userID)

	uID, err := uuid.Parse(userID)
	if err != nil {
		logging.Error.Printf("Invalid user ID: %s", userID)
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"invalid user id",
		)
	}

	var cart model.Cart

	err = s.repo.DB().
		Preload("Items.Product").
		Where("user_id = ?", uID).
		First(&cart).Error
	logging.Debug.Printf("CART DEBUG: %+v", cart)
    for _, item := range cart.Items {
	logging.Debug.Printf("ITEM: %+v", item)
	logging.Debug.Printf("PRODUCT: %+v", item.Product)
}


	if err != nil {
		logging.Debug.Printf("Cart not found for user: %s", userID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"cart not found",
		)
	}

	logging.Debug.Printf("Cart found with %d items for user: %s", len(cart.Items), userID)
	return &cart, nil
}

// UpdateCartItem updates cart item details
func (s *CartService) UpdateCartItem(
	userID string,
	itemID string,
	size *string,
	quantity *int,
) error {
	logging.Debug.Printf("Updating cart item %s for user: %s", itemID, userID)

	// Validate user UUID
	uID, err := uuid.Parse(userID)
	if err != nil {
		logging.Error.Printf("Invalid user ID: %s", userID)
		return apperror.ErrUnauthorized
	}

	// Validate item UUID
	iID, err := uuid.Parse(itemID)
	if err != nil {
		logging.Error.Printf("Invalid cart item ID: %s", itemID)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"invalid cart item id",
		)
	}

	// Get user's cart
	var cart model.Cart
	if err := s.repo.FindOneWhere(&cart, "user_id = ?", uID); err != nil {
		logging.Error.Printf("Cart not found for user: %s", userID)
		return apperror.New(
			constant.NOTFOUND,
			"",
			"cart not found",
		)
	}
	logging.Debug.Printf("Found cart: %s", cart.ID)

	// Get item and verify ownership
	var item model.CartItem
	if err := s.repo.FindOneWhere(
		&item,
		"id = ? AND cart_id = ?",
		iID,
		cart.ID,
	); err != nil {
		logging.Error.Printf("Cart item not found: %s in cart: %s", itemID, cart.ID)
		return apperror.New(
			constant.NOTFOUND,
			"",
			"cart item not found",
		)
	}
	logging.Debug.Printf("Found cart item: %s", item.ID)

	// Build update fields
	updates := map[string]interface{}{}

	if size != nil {
		updates["size"] = *size
		logging.Debug.Printf("Updating size to: %s", *size)
	}

	if quantity != nil {
		if *quantity <= 0 {
			logging.Error.Printf("Invalid quantity: %d", *quantity)
			return apperror.New(
				constant.BADREQUEST,
				"",
				"quantity must be greater than zero",
			)
		}
		updates["quantity"] = *quantity
		logging.Debug.Printf("Updating quantity to: %d", *quantity)
	}

	if len(updates) == 0 {
		logging.Error.Println("No fields to update")
		return apperror.New(
			constant.BADREQUEST,
			"",
			"no fields to update",
		)
	}

	logging.Debug.Printf("Updating cart item %s: %+v", item.ID, updates)
	return s.repo.UpdateByFields(&model.CartItem{}, item.ID, updates)
}

// RemoveCartItem removes item from cart
func (s *CartService) RemoveCartItem(cartItemID string) error {
	logging.Debug.Printf("Removing cart item: %s", cartItemID)

	itemUUID, err := uuid.Parse(cartItemID)
	if err != nil {
		logging.Error.Printf("Invalid cart item ID: %s", cartItemID)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid cart item ID",
		)
	}

	var item model.CartItem
	err = s.repo.FindById(&item, itemUUID)
	if err != nil {
		logging.Error.Printf("Cart item not found: %s", cartItemID)
		return apperror.New(
			constant.NOTFOUND,
			"",
			"Cart item not found",
		)
	}

	logging.Debug.Printf("Deleting cart item: %s", item.ID)
	if err := s.repo.Delete(&item, item.ID); err != nil {
		logging.Error.Printf("Failed to delete cart item %s: %v", item.ID, err)
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Cart item removed: %s", cartItemID)
	return nil
}
