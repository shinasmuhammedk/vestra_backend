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

// AdminAuthMiddleware protects admin-only routes
func AdminAuthMiddleware(jwtManager *jwt.JWTManager, repo repo.IPgSQLRepository) fiber.Handler {
	logging.Debug.Println("AdminAuthMiddleware initialized")
	
	return func(ctx *fiber.Ctx) error {
		logging.Debug.Println("AdminAuthMiddleware processing request")

		// Get Authorization header
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			logging.Error.Println("AdminAuthMiddleware - Authorization header missing")
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Authorization header missing",
				"AUTH_HEADER_MISSING",
				nil,
			)
		}
		logging.Debug.Printf("Admin auth header received: %s...", authHeader[:min(20, len(authHeader))])

		// Validate format: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logging.Error.Printf("AdminAuthMiddleware - Invalid authorization header format")
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Invalid authorization header format",
				"INVALID_AUTH_HEADER",
				nil,
			)
		}

		// Validate token
		token := parts[1]
		logging.Debug.Printf("Admin token validation (first 10 chars): %s...", token[:min(10, len(token))])
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			logging.Error.Printf("AdminAuthMiddleware - Token validation failed: %v", err)
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Invalid or expired token",
				"TOKEN_INVALID",
				nil,
			)
		}

		// Extract user_id from claims
		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			logging.Error.Println("AdminAuthMiddleware - Invalid token claims, user_id missing")
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Invalid token claims",
				"TOKEN_INVALID",
				nil,
			)
		}
		logging.Debug.Printf("Token validated, checking admin permissions for user: %s", userID)

		// Fetch user from DB
		var user model.User
		if err := repo.FindById(&user, userID); err != nil {
			logging.Error.Printf("AdminAuthMiddleware - User not found: %s", userID)
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"User not found",
				"USER_NOT_FOUND",
				nil,
			)
		}

		// Check if user is blocked
		if user.IsBlocked {
			logging.Error.Printf("AdminAuthMiddleware - User is blocked: %s", userID)
			return response.Error(
				ctx,
				constant.FORBIDDEN,
				"Your account has been blocked",
				"USER_BLOCKED",
				nil,
			)
		}

		// Check if user is admin
		if user.Role != "admin" {
			logging.Error.Printf("AdminAuthMiddleware - Non-admin user attempted admin access: %s, Role: %s", userID, user.Role)
			return response.Error(
				ctx,
				constant.FORBIDDEN,
				"Admin access required",
				"FORBIDDEN",
				nil,
			)
		}

		// Store user info in context for later use
		ctx.Locals("user_id", userID)
		ctx.Locals("role", user.Role)
		
		logging.Debug.Printf("AdminAuthMiddleware passed - Admin user: %s authenticated", userID)
		return ctx.Next()
	}
}

