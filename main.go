package main

import (
	"vestra-ecommerce-backend/internal/config"
	httpRoutes "vestra-ecommerce-backend/internal/delivery/http"
	"vestra-ecommerce-backend/internal/delivery/http/handlers"
	"vestra-ecommerce-backend/internal/infrastrusture/repository"
	"vestra-ecommerce-backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	// ----------------------------------------
	// 1️⃣ Initialize Gin router
	// ----------------------------------------
	r := gin.Default()

	// ----------------------------------------
	// 2️⃣ Connect to database
	// ----------------------------------------
	db := config.ConnectDB()

	// ----------------------------------------
	// 3️⃣ Auto-migrate database tables
	// (Creates tables if they don't exist)
	// ----------------------------------------
	db.AutoMigrate(
		&repository.User{},     // users table
		&repository.OTPModel{}, // otp table
	)

	// ----------------------------------------
	// 4️⃣ Initialize repositories (DB layer)
	// ----------------------------------------
	userRepo := repository.NewUserRepoGorm(db)
	otpRepo := repository.NewOTPRepoGorm(db)

	// ----------------------------------------
	// 5️⃣ Initialize usecases (business logic)
	// ----------------------------------------
	authUC := usecase.NewAuthUseCase(userRepo, otpRepo)
	userUC := usecase.NewUserUseCase(userRepo)
	adminUC := usecase.NewAdminUsecase(userRepo)

	// ----------------------------------------
	// 6️⃣ Initialize handlers (HTTP layer)
	// ----------------------------------------
	authHandler := handlers.NewAuthHandler(authUC)
	userHandler := handlers.NewUserHandler(userUC)
	adminHandler := handlers.NewAdminHandler(adminUC)

	// ----------------------------------------
	// 7️⃣ Setup all routes
	// ----------------------------------------
	httpRoutes.SetupRoutes(
		r,
		authHandler,
		userHandler,
		adminHandler,
	)

	// ----------------------------------------
	// 8️⃣ Start HTTP server
	// ----------------------------------------
	r.Run(":8080") // http://localhost:8080
}
