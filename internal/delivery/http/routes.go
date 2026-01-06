package http

import (
	"vestra-ecommerce-backend/internal/delivery/http/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler) {
	auth := r.Group("/auth")
	{
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/verify-otp", authHandler.VerifyOTP)
	}
}
