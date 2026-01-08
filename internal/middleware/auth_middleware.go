package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"vestra-ecommerce-backend/internal/utils"
)

// AuthMiddleware protects routes by validating the access token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1️⃣ Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			c.Abort()
			return
		}

		// 2️⃣ Split "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// 3️⃣ Validate token
		claims, err := utils.ValidateAccessToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// 4️⃣ Set claims in context for handlers to use
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// ✅ Continue to handler
		c.Next()
	}
}
