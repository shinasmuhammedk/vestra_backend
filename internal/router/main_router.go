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
	paymentController *controller.PaymentController,
	addressController *controller.AddressController,
	jwtManager *jwt.JWTManager,
	pgRepo repo.IPgSQLRepository,
	cartController *controller.CartController,
	wishlistController *controller.WishlistController,
	orderController *controller.OrderController,
) {

	// ================= AUTH ROUTES (PUBLIC) =================
	authGroup := app.Group("/auth")
	authGroup.Post("/signup", auth.Signup)
	authGroup.Post("/verify-otp", auth.VerifyOTP)
	authGroup.Post("/login", auth.Login)
	authGroup.Post("/forgot-password", auth.ForgotPassword)
	authGroup.Post("/reset-password", auth.ResetPassword)

	app.Post("/refresh", auth.RefreshToken)

	// ================= PUBLIC PRODUCT ROUTES =================
	app.Get("/products", productController.GetAllProducts)
	app.Get("/products/search", productController.SearchProducts)
	app.Get("/products/:id", productController.GetProductByID)

	// ================= USER ROUTES (PROTECTED) =================
	userGroup := app.Group("/user", middleware.AuthMiddleware(jwtManager))

	// Profile
	userGroup.Get("/profile", auth.GetProfile)
	userGroup.Put("/profile", auth.UpdateProfile)

	// Payments
	userGroup.Post("/payment", paymentController.CreatePayment)
	userGroup.Post("/payment/verify", paymentController.VerifyPayment)
	userGroup.Get("/payment", paymentController.GetUserPayments)
	userGroup.Get("/payment/:id", paymentController.GetUserPaymentByID)
	userGroup.Put("/payment/:id/cancel", paymentController.CancelPayment)

	// Cart
	cartGroup := userGroup.Group("/cart")
	cartGroup.Post("/", cartController.AddToCart)
	cartGroup.Get("/", cartController.GetCart)
	cartGroup.Put("/:id", cartController.UpdateCartItem)
	cartGroup.Delete("/:id", cartController.RemoveCartItem)

	// Wishlist
	wishlistGroup := userGroup.Group("/wishlist")
	wishlistGroup.Post("/", wishlistController.AddToWishlist)
	wishlistGroup.Get("/", wishlistController.GetWishlist)
	wishlistGroup.Delete("/:product_id", wishlistController.RemoveFromWishlist)

	// Orders
	orderGroup := userGroup.Group("/orders")
	orderGroup.Get("/", orderController.GetUserOrders)
	orderGroup.Post("/", orderController.PlaceOrder)
	orderGroup.Get("/:id", orderController.GetOrderDetails)
	orderGroup.Put("/:id/status", orderController.UpdateOrderStatusUser)
	orderGroup.Put("/:id/cancel", orderController.CancelOrder)
	orderGroup.Delete("/:id", orderController.DeleteOrder)

	// Address
	addressGroup := userGroup.Group("/address")
	addressGroup.Post("/", addressController.CreateAddress)
	addressGroup.Get("/", addressController.GetAddresses)
	addressGroup.Put("/:id", addressController.UpdateAddress)
	addressGroup.Delete("/:id", addressController.DeleteAddress)

	// ================= ADMIN ROUTES (PROTECTED) =================
	adminGroup := app.Group("/admin", middleware.AdminAuthMiddleware(jwtManager, pgRepo))

	// Users
	adminGroup.Put("/users/:id/block", auth.ToggleUserBlock)

	// Products
	adminGroup.Post("/products", productController.CreateProduct)
	adminGroup.Patch("/products/:id", productController.UpdateProduct)
	adminGroup.Delete("/products/:id", productController.DeleteProduct)

	// Orders
	adminGroup.Get("/orders", orderController.GetAllOrders)
	adminGroup.Put("/order/:id", orderController.UpdateOrderStatusAdmin)

	// Payments
	adminGroup.Get("/payments", paymentController.GetAllPayments)
	adminGroup.Get("/payments/:id", paymentController.GetPaymentByIDAdmin)
	adminGroup.Put("/payments/:id/status", paymentController.UpdatePaymentStatus) // ✅ update payment status
}
