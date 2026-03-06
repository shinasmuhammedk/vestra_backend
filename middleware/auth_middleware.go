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
// Reads token from cookie first, then falls back to Authorization header
func AuthMiddleware(jwtManager *jwt.JWTManager) fiber.Handler {
	logging.Debug.Println("AuthMiddleware initialized")

	return func(ctx *fiber.Ctx) error {
		logging.Debug.Println("AuthMiddleware processing request")

		var token string

		// 1. Try reading from cookie first (HttpOnly cookie set at login)
		cookieToken := ctx.Cookies("access_token")
		if cookieToken != "" {
			token = cookieToken
			logging.Debug.Println("AuthMiddleware - Token read from cookie")
		} else {
			// 2. Fall back to Authorization header (for API clients, mobile apps etc.)
			authHeader := ctx.Get("Authorization")
			if authHeader == "" {
				logging.Error.Println("AuthMiddleware - No token found in cookie or Authorization header")
				return response.Error(
					ctx,
					constant.UNAUTHORIZED,
					"Unauthorized",
					"AUTH_MISSING",
					nil,
				)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				logging.Error.Printf("AuthMiddleware - Invalid authorization header format")
				return response.Error(
					ctx,
					constant.UNAUTHORIZED,
					"Invalid authorization header format",
					"INVALID_AUTH_HEADER",
					nil,
				)
			}

			token = parts[1]
			logging.Debug.Println("AuthMiddleware - Token read from Authorization header")
		}

		// Validate token
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

		// Extract role from claims
		role, _ := claims["role"].(string)
		logging.Debug.Printf("Token validated - UserID: %s, Role: %s", userID, role)

		// Store in context for downstream handlers
		ctx.Locals("user_id", userID)
		if role != "" {
			ctx.Locals("role", role)
		}

		logging.Debug.Printf("AuthMiddleware passed - User: %s proceeding to handler", userID)
		return ctx.Next()
	}
}