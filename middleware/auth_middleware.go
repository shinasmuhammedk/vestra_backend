package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/jwt"
	"vestra-ecommerce/utils/response"
)

// AuthMiddleware protects routes using JWT access token
func AuthMiddleware(jwtManager *jwt.JWTManager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		// 1️⃣ Get Authorization header
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Authorization header missing",
				"AUTH_HEADER_MISSING",
				nil,
			)
		}

		// 2️⃣ Validate format: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Invalid authorization header format",
				"INVALID_AUTH_HEADER",
				nil,
			)
		}

		// 3️⃣ Validate token
		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Invalid or expired token",
				"TOKEN_INVALID",
				nil,
			)
		}

		// 4️⃣ Extract user_id from claims
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Invalid token claims",
				"TOKEN_INVALID",
				nil,
			)
		}

		// 5️⃣ Store user_id in context
		ctx.Locals("user_id", userID)

		return ctx.Next()
	}
}
