package router

import (
	"vestra-ecommerce/middleware"
	"vestra-ecommerce/src/controller"
	"vestra-ecommerce/src/repo"
	"vestra-ecommerce/utils/jwt"

	"github.com/gofiber/fiber/v2"
)

func Setup(
	app *fiber.App,
	auth *controller.UserAuthController,
	productController *controller.ProductController,
	jwtManager *jwt.JWTManager,
	pgRepo repo.IPgSQLRepository,
) {
	// ---------------- Auth Routes (Public) ----------------
	authGroup := app.Group("/auth")
	authGroup.Post("/signup", auth.Signup)
	authGroup.Post("/verify-otp", auth.VerifyOTP)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/forgot-password", auth.ForgotPassword)
	authGroup.Post("/reset-password", auth.ResetPassword)

	// Refresh token is usually outside /auth group
	app.Post("/refresh", auth.RefreshToken)
	// Public Products route
	app.Get("/products", productController.GetAllProducts)
	// Public Products route
	app.Get("/products/:id", productController.GetProductByID)

	// ---------------- User Routes (Protected) ----------------
	userGroup := app.Group("/user", middleware.AuthMiddleware(jwtManager))
	userGroup.Get("/profile", auth.GetProfile)
	userGroup.Put("/profile", auth.UpdateProfile)

	// ----------------- Admin Routes -----------------
	adminGroup := app.Group("/admin", middleware.AdminAuthMiddleware(jwtManager, pgRepo))

	// User management
	adminGroup.Put("/users/:id/block", auth.ToggleUserBlock)

	// Product management
	adminGroup.Post("/products", productController.CreateProduct)

}
