package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/jwt"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/response"
)

func AdminAuthMiddleware(jwtManager *jwt.JWTManager, repo repo.IPgSQLRepository) fiber.Handler {
	logging.Debug.Println("AdminAuthMiddleware initialized")

	return func(ctx *fiber.Ctx) error {
		logging.Debug.Println("AdminAuthMiddleware processing request")

		var token string

		// 1. Try cookie first
		cookieToken := ctx.Cookies("access_token")
		if cookieToken != "" {
			token = cookieToken
			logging.Debug.Println("AdminAuthMiddleware - Token read from cookie")
		} else {
			// 2. Fall back to Authorization header
			authHeader := ctx.Get("Authorization")
			if authHeader == "" {
				logging.Error.Println("AdminAuthMiddleware - No token found in cookie or Authorization header")
				return response.Error(ctx, constant.UNAUTHORIZED, "Unauthorized", "AUTH_MISSING", nil)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				logging.Error.Println("AdminAuthMiddleware - Invalid authorization header format")
				return response.Error(ctx, constant.UNAUTHORIZED, "Invalid authorization header format", "INVALID_AUTH_HEADER", nil)
			}

			token = parts[1]
			logging.Debug.Println("AdminAuthMiddleware - Token read from Authorization header")
		}

		// Validate token
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			logging.Error.Printf("AdminAuthMiddleware - Token validation failed: %v", err)
			return response.Error(ctx, constant.UNAUTHORIZED, "Invalid or expired token", "TOKEN_INVALID", nil)
		}

		// Extract user_id
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			logging.Error.Println("AdminAuthMiddleware - user_id missing from claims")
			return response.Error(ctx, constant.UNAUTHORIZED, "Invalid token claims", "TOKEN_INVALID", nil)
		}

		// Fetch user from DB
		var user model.User
		if err := repo.FindById(&user, userID); err != nil {
			logging.Error.Printf("AdminAuthMiddleware - User not found: %s", userID)
			return response.Error(ctx, constant.UNAUTHORIZED, "User not found", "USER_NOT_FOUND", nil)
		}

		// Check if blocked
		if user.IsBlocked {
			logging.Error.Printf("AdminAuthMiddleware - User is blocked: %s", userID)
			return response.Error(ctx, constant.FORBIDDEN, "Your account has been blocked", "USER_BLOCKED", nil)
		}

		// Check if admin
		if user.Role != "admin" {
			logging.Error.Printf("AdminAuthMiddleware - Non-admin access attempt: %s, Role: %s", userID, user.Role)
			return response.Error(ctx, constant.FORBIDDEN, "Admin access required", "FORBIDDEN", nil)
		}

		ctx.Locals("user_id", userID)
		ctx.Locals("role", user.Role)

		logging.Debug.Printf("AdminAuthMiddleware passed - Admin: %s authenticated", userID)
		return ctx.Next()
	}
}