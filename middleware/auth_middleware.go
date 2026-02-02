package middleware

import (
	"strings"

	"vestra-ecommerce/utils/logging"

	"github.com/gofiber/fiber/v2"

	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/jwt"
	"vestra-ecommerce/utils/response"
)

// AuthMiddleware protects routes using JWT access token
func AuthMiddleware(jwtManager *jwt.JWTManager) fiber.Handler {
	logging.Debug.Println("AuthMiddleware initialized")
	
	return func(ctx *fiber.Ctx) error {
		logging.Debug.Println("AuthMiddleware processing request")

		// Get Authorization header
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			logging.Error.Println("AuthMiddleware - Authorization header missing")
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Authorization header missing",
				"AUTH_HEADER_MISSING",
				nil,
			)
		}
		logging.Debug.Printf("Auth header received: %s", authHeader[:min(20, len(authHeader))] + "...")

		// Validate format: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logging.Error.Printf("AuthMiddleware - Invalid authorization header format: %s", authHeader)
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
		logging.Debug.Printf("Validating token (first 10 chars): %s...", token[:min(10, len(token))])
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			logging.Error.Printf("AuthMiddleware - Token validation failed: %v", err)
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
			logging.Error.Println("AuthMiddleware - Invalid token claims, user_id missing")
			return response.Error(
				ctx,
				constant.UNAUTHORIZED,
				"Invalid token claims",
				"TOKEN_INVALID",
				nil,
			)
		}
		
		// Extract role from claims (optional but useful for logging)
		role, _ := claims["role"].(string)
		logging.Debug.Printf("Token validated - UserID: %s, Role: %s", userID, role)

		// Store user_id in context for downstream handlers
		ctx.Locals("user_id", userID)
		
		// Also store role if needed (optional)
		if role != "" {
			ctx.Locals("role", role)
		}

		logging.Debug.Printf("AuthMiddleware passed - User: %s proceeding to handler", userID)
		return ctx.Next()
	}
}

