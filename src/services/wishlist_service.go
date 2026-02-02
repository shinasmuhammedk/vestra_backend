package services

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/utils/apperror"

	"github.com/google/uuid"
)

// WishlistService handles wishlist business logic
type WishlistService struct {
	repo repo.IPgSQLRepository // Repository dependency
}

// NewWishlistService creates a new wishlist service
func NewWishlistService(repo repo.IPgSQLRepository) *WishlistService {
	return &WishlistService{repo: repo}
}

// AddToWishlist adds a product to the user's wishlist
func (s *WishlistService) AddToWishlist(userID, productID string) error {
	logging.Debug.Println("AddToWishlist called")

	uID, err := uuid.Parse(userID) // Parse user ID
	if err != nil {
		logging.Error.Println("Invalid user ID:", err)
		return apperror.New(constant.BADREQUEST, "", "Invalid user ID")
	}

	pID, err := uuid.Parse(productID) // Parse product ID
	if err != nil {
		logging.Error.Println("Invalid product ID:", err)
		return apperror.New(constant.BADREQUEST, "", "Invalid product ID")
	}

	// Check if product already exists in wishlist
	var existing model.Wishlist
	err = s.repo.FindOneWhere(&existing, "user_id = ? AND product_id = ?", uID, pID)
	if err == nil {
		logging.Debug.Println("Product already exists in wishlist")
		return apperror.New(constant.CONFLICT, "", "Product already in wishlist")
	}

	wishlist := model.Wishlist{ // Create wishlist entity
		UserID:    uID,
		ProductID: pID,
	}

	if err := s.repo.Insert(&wishlist); err != nil {
		logging.Error.Println("Failed to insert wishlist item:", err)
		return apperror.New(constant.INTERNALSERVERERROR, "", "Failed to add product to wishlist")
	}

	logging.Debug.Println("Product added to wishlist successfully")
	return nil
}

// GetWishlist retrieves all wishlist items for a user
func (s *WishlistService) GetWishlist(userID string) ([]model.Wishlist, error) {
	logging.Debug.Println("GetWishlist called")

	uID, err := uuid.Parse(userID) // Parse user ID
	if err != nil {
		logging.Error.Println("Invalid user ID:", err)
		return nil, apperror.New(constant.BADREQUEST, "", "Invalid user ID")
	}

	var wishlist []model.Wishlist // Wishlist result slice

	// Fetch wishlist with product details
	err = s.repo.FindWhereWithPreload(&wishlist, "user_id = ?", []interface{}{uID}, "Product")
	if err != nil {
		logging.Error.Println("Failed to fetch wishlist:", err)
		return nil, apperror.New(constant.INTERNALSERVERERROR, "", "Failed to fetch wishlist")
	}

	logging.Debug.Println("Wishlist fetched successfully")
	return wishlist, nil
}

// RemoveFromWishlist removes a product from the user's wishlist
func (s *WishlistService) RemoveFromWishlist(userID, productID string) error {
	logging.Debug.Println("RemoveFromWishlist called")

	uID, err := uuid.Parse(userID) // Parse user ID
	if err != nil {
		logging.Error.Println("Invalid user ID:", err)
		return apperror.New(constant.BADREQUEST, "", "Invalid user ID")
	}

	pID, err := uuid.Parse(productID) // Parse product ID
	if err != nil {
		logging.Error.Println("Invalid product ID:", err)
		return apperror.New(constant.BADREQUEST, "", "Invalid product ID")
	}

	// Delete wishlist entry
	result := s.repo.Exec(
		"DELETE FROM wishlists WHERE user_id = ? AND product_id = ?",
		uID, pID,
	)

	if result.Error != nil {
		logging.Error.Println("Failed to delete wishlist item:", result.Error)
		return apperror.New(constant.INTERNALSERVERERROR, "", "Failed to remove wishlist item")
	}

	if result.RowsAffected == 0 {
		logging.Debug.Println("Wishlist item not found")
		return apperror.New(constant.NOTFOUND, "", "Wishlist item not found")
	}

	logging.Debug.Println("Wishlist item removed successfully")
	return nil
}
