package main

import (
	"vestra-ecommerce-backend/internal/config"
	"vestra-ecommerce-backend/internal/delivery/http"
	"vestra-ecommerce-backend/internal/delivery/http/handlers"
	"vestra-ecommerce-backend/internal/infrastrusture/repository"
	"vestra-ecommerce-backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	db := config.ConnectDB()

	// Auto migrate tables
	db.AutoMigrate(&repository.User{}, &repository.OTPModel{})

	// Repositories
	userRepo := repository.NewUserRepoGorm(db)
	otpRepo := repository.NewOTPRepoGorm(db)

	// Use cases
	authUC := usecase.NewAuthUseCase(userRepo, otpRepo)
	userUC := usecase.NewUserUseCase(userRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authUC)
	userHandler := handlers.NewUserHandler(userUC)

	// Routes
	http.SetupRoutes(r, authHandler, userHandler)

	r.Run(":8080")
}
