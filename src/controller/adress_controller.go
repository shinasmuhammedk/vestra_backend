package controller

import (
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/response"

	"github.com/gofiber/fiber/v2"
)

type AddressController struct {
	service *services.AddressService
}

func NewAddressController(service *services.AddressService) *AddressController {
	return &AddressController{service: service}
}

// POST /user/address
func (ac *AddressController) CreateAddress(c *fiber.Ctx) error {
	var req model.UserAddress
	if err := c.BodyParser(&req); err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	req.UserID = c.Locals("user_id").(string)

	if err := ac.service.CreateAddress(&req); err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to create address",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.CREATED,
		"Address created successfully",
		"",
		req,
	)
}

// GET /user/address
func (ac *AddressController) GetAddresses(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	addresses, err := ac.service.GetUserAddresses(userID)
	if err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to fetch addresses",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Addresses fetched successfully",
		"",
		addresses,
	)
}

// PUT /user/address/:id
func (ac *AddressController) UpdateAddress(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Address id is required",
			"",
			nil,
		)
	}

	var fields map[string]interface{}
	if err := c.BodyParser(&fields); err != nil {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if err := ac.service.UpdateAddress(id, fields); err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to update address",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Address updated successfully",
		"",
		nil,
	)
}

// DELETE /user/address/:id
func (ac *AddressController) DeleteAddress(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Error(
			c,
			constant.BADREQUEST,
			"Address id is required",
			"",
			nil,
		)
	}

	if err := ac.service.DeleteAddress(id); err != nil {
		return response.Error(
			c,
			constant.INTERNALSERVERERROR,
			"Failed to delete address",
			"",
			err.Error(),
		)
	}

	return response.Success(
		c,
		constant.SUCCESS,
		"Address deleted successfully",
		"",
		nil,
	)
}
