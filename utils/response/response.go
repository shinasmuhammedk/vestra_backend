package response

import "github.com/gofiber/fiber/v2"

// Success response
func Success(ctx *fiber.Ctx, status int, message string, code string, data interface{}) error {
	resp := APIResponse{
		StatusCode: status,
		// Success:    true,
		Message:    message,
	}

	if code != "" {
		resp.Code = code
	}
	if data != nil {
		resp.Data = data
	}

	return ctx.Status(status).JSON(resp)
}

// Error response
func Error(ctx *fiber.Ctx, status int, message string, code string, err interface{}) error {
	resp := APIResponse{
		StatusCode: status,
		// Success:    false,
		Message:    message,
	}

	if code != "" {
		resp.Code = code
	}
	if err != nil {
		resp.Error = err
	}

	return ctx.Status(status).JSON(resp)
}
