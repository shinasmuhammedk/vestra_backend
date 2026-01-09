package http

import (
	"vestra-ecommerce-backend/internal/delivery/http/handlers"
	"vestra-ecommerce-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	adminHandler *handlers.AdminHandler,
) {

	// =====================
	// AUTH ROUTES (Public)
	// =====================
	auth := r.Group("/auth")
	{
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/verify-otp", authHandler.VerifyOTP)
		auth.POST("/login", authHandler.Login)

		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)

		auth.POST("/refresh-token", authHandler.RefreshToken)
	}

	// =====================
	// USER ROUTES (Protected)
	// =====================
	user := r.Group("/user")
	user.Use(middleware.AuthMiddleware()) // 🔐 ACCESS TOKEN REQUIRED
	{
		user.GET("/profile", userHandler.GetProfile)
		user.PUT("/profile", userHandler.UpdateProfile)
	}

	// ====================
	// ADMIN ROUTES
	// ====================
	admin := r.Group("/admin")
	admin.Use(
		middleware.AuthMiddleware(),
		middleware.AdminOnlyMiddleware(),
	)
	{
		admin.GET("/users", adminHandler.GetAllUsers)
		admin.PUT("/users/:id/block", adminHandler.BlockUser)
        admin.DELETE("/users/:id", adminHandler.DeleteUser)
	}

}
