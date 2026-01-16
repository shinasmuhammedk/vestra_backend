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
	// -------------------- 1️⃣ Load config --------------------
	cfg, err := config.LoadConfig("app.yaml")
	if err != nil {
		log.Fatal("Config load failed:", err)
	}

	// -------------------- 2️⃣ Connect DB --------------------
	db := database.GetInstancepostgres(cfg)

	// -------------------- 3️⃣ Init repository --------------------
	repo.PgSQLInit()
	userRepo := repo.GetPgSQLRepository() // must implement IPgSQLRepository

	// -------------------- 4️⃣ Init email --------------------
	email.Init(cfg.SMTP)

	// -------------------- 5️⃣ Run migrations --------------------
	migration.Migrate()

	// -------------------- 6️⃣ Fiber app --------------------
	app := fiber.New(fiber.Config{
		Prefork: cfg.Server.Prefork, // use config value
	})

	// -------------------- 7️⃣ Health check --------------------
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Fiber + DB + SMTP connected 🚀")
	})

	// -------------------- 8️⃣ Initialize JWTManager --------------------
	jwtManager := jwt.NewJWTManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		time.Minute*time.Duration(cfg.JWT.AccessTTLMinutes),
		time.Hour*time.Duration(cfg.JWT.RefreshTTLHours),
	)

	// -------------------- 9️⃣ Initialize Auth Service & Controller --------------------
	authService := services.NewUserAuthService(userRepo, 5) // OTP expiry 5 min
	authController := controller.NewUserAuthController(authService, jwtManager)

	// -------------------- 🔟 Initialize Product Service & Controller --------------------
	productService := services.NewProductService(userRepo)
	productController := controller.NewProductController(productService)

	// -------------------- 1️⃣1️⃣ Register routes --------------------
	router.Setup(app, authController, productController, jwtManager, userRepo)

	// -------------------- 1️⃣2️⃣ Graceful shutdown --------------------
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	port := cfg.Server.Port
	go func() {
		log.Printf("🚀 Server started on http://localhost:%d\n", port)
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
