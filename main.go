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
	db.AutoMigrate(&repository.UserModel{}, &repository.OTPModel{})

	// Repositories
	userRepo := repository.NewUserRepoGorm(db)
	otpRepo := repository.NewOTPRepoGorm(db)

	// Use case
	authUC := usecase.NewAuthUseCase(userRepo, otpRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authUC)

	// Routes
	http.SetupRoutes(r, authHandler)

	r.Run(":8080")
}
