package router

import (
	"vestra-ecommerce/middleware"
	"vestra-ecommerce/src/controller"
	"vestra-ecommerce/src/repo"
	"vestra-ecommerce/utils/jwt"
	"vestra-ecommerce/utils/logging"

	"github.com/gofiber/fiber/v2"
)

// Setup configures all application routes
func Setup(
	app *fiber.App,
	auth *controller.UserAuthController,
	productController *controller.ProductController,
	paymentController *controller.PaymentController,
	addressController *controller.AddressController,
	jwtManager *jwt.JWTManager,
	pgRepo repo.IPgSQLRepository,
	cartController *controller.CartController,
	wishlistController *controller.WishlistController,
	orderController *controller.OrderController,
) {
	logging.Debug.Println("Setting up application routes")

	// ================= AUTH ROUTES (PUBLIC) =================
	logging.Debug.Println("Configuring auth routes")
	authGroup := app.Group("/auth")
	authGroup.Post("/signup", auth.Signup)
	authGroup.Post("/verify-otp", auth.VerifyOTP)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/forgot-password", auth.ForgotPassword)
	authGroup.Post("/reset-password", auth.ResetPassword)
    authGroup.Post("/logout", auth.Logout)
	app.Post("/refresh", auth.RefreshToken)
	logging.Debug.Printf("Auth routes configured: 6 endpoints")

	// ================= PUBLIC PRODUCT ROUTES =================
	logging.Debug.Println("Configuring public product routes")
	app.Get("/products", productController.GetAllProducts)
	app.Get("/products/search", productController.SearchProducts)
	app.Get("/products/:id", productController.GetProductByID)
	logging.Debug.Printf("Public product routes configured: 3 endpoints")

	// ================= USER ROUTES (PROTECTED) =================
	logging.Debug.Println("Configuring user routes (protected)")
	userGroup := app.Group("/user", middleware.AuthMiddleware(jwtManager))

	// Profile
	logging.Debug.Println("Configuring profile routes")
	userGroup.Get("/profile", auth.GetProfile)
	userGroup.Put("/profile", auth.UpdateProfile)

	// Payments
	logging.Debug.Println("Configuring payment routes")
	userGroup.Post("/payment", paymentController.CreatePayment)
	userGroup.Post("/payment/verify", paymentController.VerifyPayment)
	userGroup.Get("/payment", paymentController.GetUserPayments)
	userGroup.Get("/payment/:id", paymentController.GetUserPaymentByID)
	userGroup.Put("/payment/:id/cancel", paymentController.CancelPayment)

	// Cart
	logging.Debug.Println("Configuring cart routes")
	cartGroup := userGroup.Group("/cart")
	cartGroup.Post("/", cartController.AddToCart)
	cartGroup.Get("/", cartController.GetCart)
	cartGroup.Put("/:id", cartController.UpdateCartItem)
	cartGroup.Delete("/:id", cartController.RemoveCartItem)

	// Wishlist
	logging.Debug.Println("Configuring wishlist routes")
	wishlistGroup := userGroup.Group("/wishlist")
	wishlistGroup.Post("/", wishlistController.AddToWishlist)
	wishlistGroup.Get("/", wishlistController.GetWishlist)
	wishlistGroup.Delete("/:product_id", wishlistController.RemoveFromWishlist)

	// Orders
	logging.Debug.Println("Configuring order routes")
	orderGroup := userGroup.Group("/orders")
	orderGroup.Get("/", orderController.GetUserOrders)
	orderGroup.Post("/", orderController.PlaceOrder)
	orderGroup.Get("/:id", orderController.GetOrderDetails)
	orderGroup.Put("/:id/status", orderController.UpdateOrderStatusUser)
	orderGroup.Put("/:id/cancel", orderController.CancelOrder)
	orderGroup.Delete("/:id", orderController.DeleteOrder)

	// Address
	logging.Debug.Println("Configuring address routes")
	addressGroup := userGroup.Group("/address")
	addressGroup.Post("/", addressController.CreateAddress)
	addressGroup.Get("/", addressController.GetAddresses)
	addressGroup.Put("/:id", addressController.UpdateAddress)
	addressGroup.Delete("/:id", addressController.DeleteAddress)

	logging.Debug.Printf("User routes configured: %d total endpoints", countRoutes(userGroup))

	// ================= ADMIN ROUTES (PROTECTED) =================
	logging.Debug.Println("Configuring admin routes (protected)")
	adminGroup := app.Group("/admin", middleware.AdminAuthMiddleware(jwtManager, pgRepo))

	// Users
	logging.Debug.Println("Configuring admin user routes")
	adminGroup.Get("/users", auth.GetAllUsers)
	adminGroup.Put("/users/:id/block", auth.ToggleUserBlock)
	adminGroup.Delete("/users/:id", auth.DeleteUserByID)

	// Products
	logging.Debug.Println("Configuring admin product routes")
	adminGroup.Post("/products", productController.CreateProduct)
	adminGroup.Patch("/products/:id", productController.UpdateProduct)
	adminGroup.Delete("/products/:id", productController.DeleteProduct)

	// Orders
	logging.Debug.Println("Configuring admin order routes")
	adminGroup.Get("/orders", orderController.GetAllOrders)
	adminGroup.Put("/order/:id", orderController.UpdateOrderStatusAdmin)
	adminGroup.Put("/order/:id/status", orderController.UpdateOrderStatusAdmin)

	// Payments
	logging.Debug.Println("Configuring admin payment routes")
	adminGroup.Get("/payments", paymentController.GetAllPayments)
	adminGroup.Get("/payments/:id", paymentController.GetPaymentByIDAdmin)
	adminGroup.Put("/payments/:id/status", paymentController.UpdatePaymentStatus)

	logging.Debug.Printf("Admin routes configured: %d total endpoints", countRoutes(adminGroup))
	logging.Debug.Println("All routes configured successfully")
}

// Helper function to count routes (for logging purposes)
func countRoutes(router fiber.Router) int {
	// This is a simplified count - in a real application,
	// you might want to implement more sophisticated route counting
	// Note: Fiber doesn't expose route count directly, so this is approximate
	return 0 // Placeholder - you could implement actual counting if needed
}