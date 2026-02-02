package controller

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/response"

	"github.com/gofiber/fiber/v2"
)

type AddressController struct {
	service *services.AddressService
}

func NewAddressController(service *services.AddressService) *AddressController {
	logging.Debug.Println("AddressController initialized")
	return &AddressController{service: service}
}

// CreateAddress creates a new address for user
func (ac *AddressController) CreateAddress(c *fiber.Ctx) error {
	logging.Debug.Println("CreateAddress endpoint called")

	var req model.UserAddress
	if err := c.BodyParser(&req); err != nil {
		logging.Error.Printf("CreateAddress - Invalid request body: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	userID := c.Locals("user_id").(string)
	req.UserID = userID
	logging.Debug.Printf("Creating address for user: %s", userID)

	if err := ac.service.CreateAddress(&req); err != nil {
		logging.Error.Printf("CreateAddress failed: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to create address",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Address created successfully: %s", req.ID)
	return response.Success(
		c,
		constant.CREATED,
		"Address created successfully",
		"",
		req,
	)
}

// GetAddresses retrieves all addresses for user
func (ac *AddressController) GetAddresses(c *fiber.Ctx) error {
	logging.Debug.Println("GetAddresses endpoint called")

	userID := c.Locals("user_id").(string)
	logging.Debug.Printf("Fetching addresses for user: %s", userID)

	addresses, err := ac.service.GetUserAddresses(userID)
	if err != nil {
		logging.Error.Printf("GetAddresses failed: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch addresses",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Found %d addresses for user: %s", len(addresses), userID)
	return response.Success(
		c,
		constant.SUCCESS,
		"Addresses fetched successfully",
		"",
		addresses,
	)
}

// UpdateAddress updates user's address
func (ac *AddressController) UpdateAddress(c *fiber.Ctx) error {
	logging.Debug.Println("UpdateAddress endpoint called")

	id := c.Params("id")
	if id == "" {
		logging.Error.Println("UpdateAddress - Missing address ID")
		return response.Error(
			c,
			constant.BADREQUEST,
			"Address id is required",
			"",
			nil,
		)
	}
	logging.Debug.Printf("Updating address: %s", id)

	var fields map[string]interface{}
	if err := c.BodyParser(&fields); err != nil {
		logging.Error.Printf("UpdateAddress - Invalid request body: %v", err)
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Update fields: %v", fields)
	
	if err := ac.service.UpdateAddress(id, fields); err != nil {
		logging.Error.Printf("UpdateAddress failed: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to update address",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Address updated successfully: %s", id)
	return response.Success(
		c,
		constant.SUCCESS,
		"Address updated successfully",
		"",
		nil,
	)
}

// DeleteAddress removes user's address
func (ac *AddressController) DeleteAddress(c *fiber.Ctx) error {
	logging.Debug.Println("DeleteAddress endpoint called")

	id := c.Params("id")
	if id == "" {
		logging.Error.Println("DeleteAddress - Missing address ID")
		return response.Error(
			c,
			constant.BADREQUEST,
			"Address id is required",
			"",
			nil,
		)
	}
	logging.Debug.Printf("Deleting address: %s", id)

	if err := ac.service.DeleteAddress(id); err != nil {
		logging.Error.Printf("DeleteAddress failed: %v", err)
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to delete address",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Address deleted successfully: %s", id)
	return response.Success(
		c,
		constant.SUCCESS,
		"Address deleted successfully",
		"",
		nil,
	)
}