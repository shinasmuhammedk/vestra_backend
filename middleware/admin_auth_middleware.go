package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/jwt"
	"vestra-ecommerce/utils/response"
)

// AdminAuthMiddleware protects admin routes
func AdminAuthMiddleware(jwtManager *jwt.JWTManager, repo repo.IPgSQLRepository) fiber.Handler {
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

		// 5️⃣ Fetch user from DB
		var user model.User
		if err := repo.FindById(&user, userID); err != nil {
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"User not found",
				"USER_NOT_FOUND",
				nil,
			)
		}

		// 6️⃣ Check if role is admin
		if user.Role != "admin" {
			return response.Error(
				ctx,
				constant.FORBIDDEN,
				"Admin access required",
				"FORBIDDEN",
				nil,
			)
		}

		// 7️⃣ Store user info in context for later use
		ctx.Locals("user_id", userID)
		ctx.Locals("role", user.Role)

		return ctx.Next()
	}
}
