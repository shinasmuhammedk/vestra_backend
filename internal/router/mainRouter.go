package router

import (
	"vestra-ecommerce/middleware"
	"vestra-ecommerce/src/controller"
	"vestra-ecommerce/utils/jwt"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, auth *controller.UserAuthController, jwtManager *jwt.JWTManager) {
	// ---------------- Auth Routes (Public) ----------------
	authGroup := app.Group("/auth")
	authGroup.Post("/signup", auth.Signup)
	authGroup.Post("/verify-otp", auth.VerifyOTP)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/forgot-password", auth.ForgotPassword)
	authGroup.Post("/reset-password", auth.ResetPassword)

	// Refresh token is usually outside /auth group
	app.Post("/refresh", auth.RefreshToken)

	// ---------------- User Routes (Protected) ----------------
	userGroup := app.Group("/user", middleware.AuthMiddleware(jwtManager))
	userGroup.Get("/profile", auth.GetProfile)
	userGroup.Put("/profile", auth.UpdateProfile)
    
    
    // ----------------- Admin routes -----------------
	adminGroup := app.Group("/admin", middleware.AuthMiddleware(jwtManager))
	adminGroup.Put("/users/:id/block", auth.ToggleUserBlock)


}
