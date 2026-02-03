package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"vestra-ecommerce/config"
	"vestra-ecommerce/internal/router"
	"vestra-ecommerce/middleware"
	"vestra-ecommerce/migration"
	"vestra-ecommerce/src/controller"
	"vestra-ecommerce/src/repo"
	"vestra-ecommerce/src/services"
	database "vestra-ecommerce/utils/databases"
	"vestra-ecommerce/utils/email"
	"vestra-ecommerce/utils/jwt"
	"vestra-ecommerce/utils/logging"
	validator "vestra-ecommerce/utils/validation"
)

func main() {

	// 1️⃣ INIT LOGGER — MUST BE FIRST
	logging.InitLogger()
	logging.Debug.Println("Logger initialized")

	// 2️⃣ Load application configuration
	cfg, err := config.LoadConfig("app.yaml")
	if err != nil {
		logging.Error.Fatalf("Failed to load config: %v", err)
	}

	// 3️⃣ Initialize Database
	db := database.GetInstancepostgres(cfg)
	logging.Debug.Println("Database connection initialized")

	// 4️⃣ Initialize Repository Layer
	repo.PgSQLInit()
	pgRepo := repo.GetPgSQLRepository()
	logging.Debug.Println("Repository layer initialized")

	// 5️⃣ Initialize Email Service
	email.Init(cfg.SMTP)
	logging.Debug.Println("Email service initialized")

	// 6️⃣ Initialize Validator
	validator.Init()
	logging.Debug.Println("Validator initialized")

	// 7️⃣ Run Database Migrations
	migration.Migrate()
	logging.Debug.Println("Database migration completed")

	// 8️⃣ Initialize Fiber App
	app := fiber.New(fiber.Config{
		Prefork: cfg.Server.Prefork,
	})

	// Global middlewares
	app.Use(middleware.CORSMiddleware())
	logging.Debug.Println("Fiber app initialized with CORS middleware")

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK 🚀")
	})

	// 9️⃣ Initialize JWT Manager
	jwtManager := jwt.NewJWTManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		time.Minute*time.Duration(cfg.JWT.AccessTTLMinutes),
		time.Hour*time.Duration(cfg.JWT.RefreshTTLHours),
	)
	logging.Debug.Println("JWT manager initialized")

	// 🔟 Initialize Services & Controllers

	authService := services.NewUserAuthService(pgRepo, 5)
	authController := controller.NewUserAuthController(authService, jwtManager)

	productService := services.NewProductService(pgRepo)
	productController := controller.NewProductController(productService)

	cartService := services.NewCartService(pgRepo)
	cartController := controller.NewCartController(cartService)

	wishlistService := services.NewWishlistService(pgRepo)
	wishlistController := controller.NewWishlistController(wishlistService)

	orderService := services.NewOrderService(pgRepo)
	orderController := controller.NewOrderController(orderService)

	addressService := services.NewAddressService(pgRepo)
	addressController := controller.NewAddressController(addressService)

	paymentService := services.NewPaymentService(pgRepo)
	paymentController := controller.NewPaymentController(paymentService)

	logging.Debug.Println("All services and controllers initialized")

	// 1️⃣1️⃣ Setup Routes
	router.Setup(
		app,
		authController,
		productController,
		paymentController,
		addressController,
		jwtManager,
		pgRepo,
		cartController,
		wishlistController,
		orderController,
	)

	logging.Debug.Println("Routes registered successfully")

	// 1️⃣2️⃣ Start Server
	port := cfg.Server.Port

	go func() {
		logging.Debug.Printf("Server starting on port %d\n", port)
		if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
			logging.Error.Printf("Server stopped with error: %v\n", err)
		}
	}()

	// 1️⃣3️⃣ Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Debug.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logging.Error.Printf("Server shutdown failed: %v\n", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
		logging.Debug.Println("Database connection closed")
	}

	logging.Debug.Println("Server gracefully stopped")
}
