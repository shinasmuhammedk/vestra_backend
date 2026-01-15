package response

import "github.com/gofiber/fiber/v2"

// Success response
func Success(
	ctx *fiber.Ctx,
	status int,
	message string,
	code string,
	data interface{},
) error {
	return ctx.Status(status).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Code:    code,
	})
}

// Error response
func Error(
	ctx *fiber.Ctx,
	status int,
	message string,
	code string,
	err interface{},
) error {
	return ctx.Status(status).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   err,
		Code:    code,
	})
}
