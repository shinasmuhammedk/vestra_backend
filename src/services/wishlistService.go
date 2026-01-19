package services

import (
	"errors"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/google/uuid"
)

type WishlistService struct {
	repo repo.IPgSQLRepository
}

func NewWishlistService(repo repo.IPgSQLRepository) *WishlistService {
	return &WishlistService{repo: repo}
}

func (s *WishlistService) AddToWishlist(userID, productID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperror.New(400, "INVALID_USER_ID", "Invalid user ID")
	}

	pID, err := uuid.Parse(productID)
	if err != nil {
		return apperror.New(400, "INVALID_PRODUCT_ID", "Invalid product ID")
	}

	// Check if already exists
	var existing model.Wishlist
	err = s.repo.FindOneWhere(&existing, "user_id = ? AND product_id = ?", uID, pID)
	if err == nil {
		return apperror.New(409, "ALREADY_EXISTS", "Product already in wishlist")
	}

	wishlist := model.Wishlist{
		UserID:    uID,
		ProductID: pID,
	}

	if err := s.repo.Insert(&wishlist); err != nil {
		return apperror.New(500, "INSERT_FAILED", err.Error())
	}

	return nil
}



func (s *WishlistService) GetWishlist(userID string) ([]model.Wishlist, error) {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	var wishlist []model.Wishlist

	// preload product details
	err = s.repo.
		Raw("").
		Preload("Product").
		Find(&wishlist, "user_id = ?", uID).Error

	if err != nil {
		return nil, err
	}

	return wishlist, nil
}




func (s *WishlistService) RemoveFromWishlist(userID string, productID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user id")
	}

	pID, err := uuid.Parse(productID)
	if err != nil {
		return errors.New("invalid product id")
	}

	result := s.repo.Exec(
		"DELETE FROM wishlists WHERE user_id = ? AND product_id = ?",
		uID, pID,
	)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("wishlist item not found")
	}

	return nil
}
