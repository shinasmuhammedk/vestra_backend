package services

import (
	"github.com/google/uuid"

	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/utils/apperror"
)

type CartService struct {
	repo repo.IPgSQLRepository
}

func NewCartService(repo repo.IPgSQLRepository) *CartService {
	return &CartService{repo: repo}
}

func (s *CartService) AddToCart(
	userID string,
	productID string,
	size string,
	quantity int,
) error {

	// ---------- Validate UUIDs ----------
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"invalid user id",
		)
	}

	pID, err := uuid.Parse(productID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"invalid product id",
		)
	}

	// ---------- Get or Create Cart ----------
	var cart model.Cart
	err = s.repo.FindOneWhere(&cart, "user_id = ?", uID)
	if err != nil {
		cart = model.Cart{UserID: uID}
		if err := s.repo.Insert(&cart); err != nil {
			return apperror.ErrInternal
		}
	}

	// ---------- Check if item exists ----------
	var item model.CartItem
	err = s.repo.FindOneWhere(
		&item,
		"cart_id = ? AND product_id = ? AND size = ?",
		cart.ID,
		pID,
		size,
	)

	if err == nil {
		// Increase quantity
		return s.repo.UpdateByFields(
			&model.CartItem{},
			item.ID,
			map[string]interface{}{
				"quantity": item.Quantity + quantity,
			},
		)
	}

	// ---------- Add new item ----------
	cartItem := model.CartItem{
		CartID:    cart.ID,
		ProductID: pID,
		Size:      size,
		Quantity:  quantity,
	}

	if err := s.repo.Insert(&cartItem); err != nil {
		return apperror.ErrInternal
	}

	return nil
}

func (s *CartService) GetUserCart(userID string) (*model.Cart, error) {

	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"invalid user id",
		)
	}

	var cart model.Cart
	err = s.repo.Raw(
		"SELECT * FROM carts WHERE user_id = ?",
		uID,
	).Preload("Items").First(&cart).Error

	if err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"cart not found",
		)
	}

	return &cart, nil
}

func (s *CartService) UpdateCartItem(
	userID string,
	itemID string,
	size *string,
	quantity *int,
) error {

	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.ErrUnauthorized
	}

	iID, err := uuid.Parse(itemID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"invalid cart item id",
		)
	}

	// 1️⃣ Get cart for user
	var cart model.Cart
	if err := s.repo.FindOneWhere(&cart, "user_id = ?", uID); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"",
			"cart not found",
		)
	}

	// 2️⃣ Get item & verify ownership
	var item model.CartItem
	if err := s.repo.FindOneWhere(
		&item,
		"id = ? AND cart_id = ?",
		iID,
		cart.ID,
	); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"",
			"cart item not found",
		)
	}

	// 3️⃣ Build update fields
	updates := map[string]interface{}{}

	if size != nil {
		updates["size"] = *size
	}

	if quantity != nil {
		if *quantity <= 0 {
			return apperror.New(
				constant.BADREQUEST,
				"",
				"quantity must be greater than zero",
			)
		}
		updates["quantity"] = *quantity
	}

	if len(updates) == 0 {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"no fields to update",
		)
	}

	return s.repo.UpdateByFields(&model.CartItem{}, item.ID, updates)
}

// RemoveCartItem deletes a cart item by its ID
func (s *CartService) RemoveCartItem(cartItemID string) error {
	itemUUID, err := uuid.Parse(cartItemID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid cart item ID",
		)
	}

	var item model.CartItem
	err = s.repo.FindById(&item, itemUUID)
	if err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"",
			"Cart item not found",
		)
	}

	if err := s.repo.Delete(&item, item.ID); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			err.Error(),
		)
	}

	return nil
}
