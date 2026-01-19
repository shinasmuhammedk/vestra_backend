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
	cartController *controller.CartController,
	wishlistController *controller.WishlistController,
	orderController *controller.OrderController,
) {
	// ---------------- Auth Routes (Public) ----------------
	authGroup := app.Group("/auth")
	authGroup.Post("/signup", auth.Signup)
	authGroup.Post("/verify-otp", auth.VerifyOTP)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/forgot-password", auth.ForgotPassword)
	authGroup.Post("/reset-password", auth.ResetPassword)

	app.Post("/refresh", auth.RefreshToken)

	// ---------------- Public Products ----------------
	app.Get("/products", productController.GetAllProducts)
	app.Get("/products/:id", productController.GetProductByID)

	// ----------------- User Routes (Protected) -----------------
	userGroup := app.Group("/user", middleware.AuthMiddleware(jwtManager))
	userGroup.Get("/profile", auth.GetProfile)
	userGroup.Put("/profile", auth.UpdateProfile)

	cartGroup := app.Group("/cart", middleware.AuthMiddleware(jwtManager))
	cartGroup.Post("/", cartController.AddToCart)
	cartGroup.Get("/", cartController.GetCart)
	cartGroup.Put("/:id", cartController.UpdateCartItem)
	cartGroup.Delete("/:id", cartController.RemoveCartItem)

	wishlistGroup := app.Group("/wishlist", middleware.AuthMiddleware(jwtManager))
	wishlistGroup.Post("/", wishlistController.AddToWishlist)
	wishlistGroup.Get("/", wishlistController.GetWishlist)
	wishlistGroup.Delete("/:product_id", wishlistController.RemoveFromWishlist)

	orderGroup := app.Group("/orders", middleware.AuthMiddleware(jwtManager))
	orderGroup.Get("/", orderController.GetUserOrders)
	orderGroup.Post("/", orderController.PlaceOrder)
	orderGroup.Put("/:id/status", orderController.UpdateOrderStatusUser) // User endpoint to update own order
	orderGroup.Get("/:id", orderController.GetOrderDetails)              // GET /orders/:id
	orderGroup.Delete("/:id", orderController.DeleteOrder)
	orderGroup.Put("/:id/cancel", orderController.CancelOrder) // ✅ NEW

	// ----------------- Admin Routes -----------------
	adminGroup := app.Group("/admin", middleware.AdminAuthMiddleware(jwtManager, pgRepo))
	adminGroup.Put("/users/:id/block", auth.ToggleUserBlock)
	adminGroup.Post("/products", productController.CreateProduct)
	adminGroup.Delete("/products/:id", productController.DeleteProduct)

	// Admin Order routes
	adminGroup.Get("/orders", orderController.GetAllOrders)
	adminGroup.Put("/order/:id", orderController.UpdateOrderStatusAdmin) // Admin updates any order
}
