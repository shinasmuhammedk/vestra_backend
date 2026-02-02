package services

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/utils/apperror"
	constant "vestra-ecommerce/utils/constants"
)

type AddressService struct {
	repo repo.IPgSQLRepository
}

func NewAddressService(repo repo.IPgSQLRepository) *AddressService {
	logging.Debug.Println("AddressService initialized")
	return &AddressService{repo: repo}
}

// CreateAddress creates a new address for user
func (s *AddressService) CreateAddress(address *model.UserAddress) error {
	logging.Debug.Printf("Creating address for user: %s", address.UserID)

	if address == nil {
		logging.Error.Println("Address data is nil")
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Address data is nil",
		)
	}

	if err := s.repo.Insert(address); err != nil {
		logging.Error.Printf("Failed to create address for user %s: %v", address.UserID, err)
		return err
	}

	logging.Debug.Printf("Address created: %s", address.ID)
	return nil
}

// GetUserAddresses retrieves all addresses for a user
func (s *AddressService) GetUserAddresses(userID string) ([]model.UserAddress, error) {
	logging.Debug.Printf("Getting addresses for user: %s", userID)

	var addresses []model.UserAddress
	err := s.repo.FindAllWhere(&addresses, "user_id = ?", userID)
	if err != nil {
		logging.Error.Printf("Failed to get addresses for user %s: %v", userID, err)
		return nil, err
	}

	logging.Debug.Printf("Found %d addresses for user: %s", len(addresses), userID)
	return addresses, nil
}

// UpdateAddress updates address details
func (s *AddressService) UpdateAddress(id string, fields map[string]interface{}) error {
	logging.Debug.Printf("Updating address: %s, fields: %v", id, fields)

	var address model.UserAddress
	if err := s.repo.FindById(&address, id); err != nil {
		logging.Error.Printf("Address not found: %s", id)
		return apperror.New(
			constant.NOTFOUND,
			"",
			"Address not found",
		)
	}

	if err := s.repo.UpdateByFields(&address, id, fields); err != nil {
		logging.Error.Printf("Failed to update address %s: %v", id, err)
		return err
	}

	logging.Debug.Printf("Address updated: %s", id)
	return nil
}

// DeleteAddress removes address by ID
func (s *AddressService) DeleteAddress(id string) error {
	logging.Debug.Printf("Deleting address: %s", id)

	var address model.UserAddress
	if err := s.repo.FindById(&address, id); err != nil {
		logging.Error.Printf("Address to delete not found: %s", id)
		return apperror.New(
			constant.NOTFOUND,
			"",
			"Address not found",
		)
	}

	if err := s.repo.Delete(&address, id); err != nil {
		logging.Error.Printf("Failed to delete address %s: %v", id, err)
		return err
	}

	logging.Debug.Printf("Address deleted: %s", id)
	return nil
}