package middleware

import (
	"vestra-ecommerce/utils/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware configures CORS for the application
func CORSMiddleware() fiber.Handler {
    logging.Debug.Println("CORS middleware configured")
    
    return cors.New(cors.Config{
        AllowOrigins:     "http://localhost:5173,http://localhost:5174,http://localhost:5175,http://localhost:5176,http://localhost:5177,http://localhost:3000",
        AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With", // Fixed: removed Cookie, removed spaces
        ExposeHeaders:    "Content-Length,Content-Type", // Remove Authorization - not needed for cookies
        AllowCredentials: true,
        MaxAge:           86400, // Optional: cache preflight for 24 hours
    })
}