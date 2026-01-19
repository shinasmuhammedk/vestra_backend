package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"

	"vestra-ecommerce/config"
	"vestra-ecommerce/internal/router"
	"vestra-ecommerce/migration"
	"vestra-ecommerce/src/controller"
	"vestra-ecommerce/src/repo"
	"vestra-ecommerce/src/services"
	database "vestra-ecommerce/utils/databases"
	"vestra-ecommerce/utils/email"
	"vestra-ecommerce/utils/jwt"
)

func main() {
	// -------------------- 1️⃣ Load Config --------------------
	cfg, err := config.LoadConfig("app.yaml")
	if err != nil {
		log.Fatal("❌ Config load failed:", err)
	}

	// -------------------- 2️⃣ Database --------------------
	db := database.GetInstancepostgres(cfg)

	// -------------------- 3️⃣ Repository --------------------
	repo.PgSQLInit()
	pgRepo := repo.GetPgSQLRepository()

	// -------------------- 4️⃣ Email --------------------
	email.Init(cfg.SMTP)

	// -------------------- 5️⃣ Migrations --------------------
	migration.Migrate()

	// -------------------- 6️⃣ Fiber App --------------------
	app := fiber.New(fiber.Config{
		Prefork: cfg.Server.Prefork,
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK 🚀")
	})

	// -------------------- 7️⃣ JWT Manager --------------------
	jwtManager := jwt.NewJWTManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		time.Minute*time.Duration(cfg.JWT.AccessTTLMinutes),
		time.Hour*time.Duration(cfg.JWT.RefreshTTLHours),
	)

	// -------------------- 8️⃣ Auth --------------------
	authService := services.NewUserAuthService(pgRepo, 5)
	authController := controller.NewUserAuthController(authService, jwtManager)

	// -------------------- 9️⃣ Products --------------------
	productService := services.NewProductService(pgRepo)
	productController := controller.NewProductController(productService)

	// -------------------- 🔟 Cart --------------------
	cartService := services.NewCartService(pgRepo)
	cartController := controller.NewCartController(cartService)

	// -------------------- Wishlist --------------------
	wishlistService := services.NewWishlistService(pgRepo)
	wishlistController := controller.NewWishlistController(wishlistService)

	orderService := services.NewOrderService(pgRepo)
	orderController := controller.NewOrderController(orderService)

	// -------------------- 1️⃣1️⃣ Routes --------------------
	router.Setup(
		app,
		authController,
		productController,
		jwtManager,
		pgRepo,
		cartController,
		wishlistController,
        orderController,
	)

	// -------------------- 1️⃣2️⃣ Graceful Shutdown --------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	port := cfg.Server.Port
	go func() {
		log.Printf("🚀 Server running on http://localhost:%d\n", port)
		if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
			log.Println("Server stopped:", err)
		}
	}()

	<-quit
	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Println("Server shutdown failed:", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	log.Println("✅ Server gracefully stopped")
}
