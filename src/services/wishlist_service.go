package services

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/google/uuid"
)

type WishlistService struct {
	repo repo.IPgSQLRepository
}

func NewWishlistService(repo repo.IPgSQLRepository) *WishlistService {
	return &WishlistService{repo: repo}
}

// AddToWishlist adds a product to the user's wishlist
func (s *WishlistService) AddToWishlist(userID, productID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid user ID",
		)
	}

	pID, err := uuid.Parse(productID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid product ID",
		)
	}

	// Check if already exists
	var existing model.Wishlist
	err = s.repo.FindOneWhere(&existing, "user_id = ? AND product_id = ?", uID, pID)
	if err == nil {
		return apperror.New(
			constant.CONFLICT,
			"",
			"Product already in wishlist",
		)
	}

	wishlist := model.Wishlist{
		UserID:    uID,
		ProductID: pID,
	}

	if err := s.repo.Insert(&wishlist); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to add product to wishlist",
		)
	}

	return nil
}

// GetWishlist retrieves all wishlist items for a user with product details
func (s *WishlistService) GetWishlist(userID string) ([]model.Wishlist, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid user ID",
		)
	}

	var wishlist []model.Wishlist

	// Use repository method with preload to fetch product details
	err = s.repo.FindWhereWithPreload(&wishlist, "user_id = ?", []interface{}{uID}, "Product")
	if err != nil {
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to fetch wishlist",
		)
	}

	return wishlist, nil
}

// RemoveFromWishlist removes a product from the user's wishlist
func (s *WishlistService) RemoveFromWishlist(userID, productID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid user ID",
		)
	}

	pID, err := uuid.Parse(productID)
	if err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid product ID",
		)
	}

	result := s.repo.Exec(
		"DELETE FROM wishlists WHERE user_id = ? AND product_id = ?",
		uID, pID,
	)

	if result.Error != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to remove wishlist item",
		)
	}

	if result.RowsAffected == 0 {
		return apperror.New(
			constant.NOTFOUND,
			"",
			"Wishlist item not found",
		)
	}

	return nil
}
