package services

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	"vestra-ecommerce/utils/utils/apperror"
	constant "vestra-ecommerce/utils/constants"
)

type AddressService struct {
	repo repo.IPgSQLRepository
}

func NewAddressService(repo repo.IPgSQLRepository) *AddressService {
	return &AddressService{repo: repo}
}

// Create Address
func (s *AddressService) CreateAddress(address *model.UserAddress) error {
	if address == nil {
		return apperror.New(
            constant.BADREQUEST, 
            "",
             "Address data is nil",
            )
	}
	return s.repo.Insert(address)
}

// Get all addresses for a user
func (s *AddressService) GetUserAddresses(userID string) ([]model.UserAddress, error) {
	var addresses []model.UserAddress
	err := s.repo.FindAllWhere(&addresses, "user_id = ?", userID)
	return addresses, err
}

// Update Address
func (s *AddressService) UpdateAddress(id string, fields map[string]interface{}) error {
	var address model.UserAddress
	if err := s.repo.FindById(&address, id); err != nil {
		return apperror.New(
            constant.NOTFOUND,
             "",
              "Address not found",
            )
	}
	return s.repo.UpdateByFields(&address, id, fields)
}

// Delete Address
func (s *AddressService) DeleteAddress(id string) error {
	var address model.UserAddress
	if err := s.repo.FindById(&address, id); err != nil {
		return apperror.New(
            constant.NOTFOUND, 
            "", 
            "Address not found",
        )
	}
	return s.repo.Delete(&address, id)
}
