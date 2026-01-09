package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"vestra-ecommerce-backend/internal/utils"
)

// AdminOnlyMiddleware allows only admin users
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1️⃣ Read Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header missing",
			})
			c.Abort()
			return
		}

		// 2️⃣ Expect: Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// 3️⃣ Validate access token
		claims, err := utils.ValidateAccessToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// 4️⃣ Check admin role
		if claims.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "admin access only",
			})
			c.Abort()
			return
		}

		// 5️⃣ Store claims in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// ✅ Continue
		c.Next()
	}
}
