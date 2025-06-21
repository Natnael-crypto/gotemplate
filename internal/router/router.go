package router

import (
	"gotemplate/internal/handler"
	"gotemplate/pkg/auth"
	"gotemplate/pkg/middleware"

	"github.com/gin-gonic/gin" // Import Gin
)

// SetupRouter sets up all application routes and their handlers
func SetupRouter(
	userHandler handler.UserHandler,
	productHandler handler.ProductHandler,
	jwtManager *auth.JWTManager,
	debug bool,
) *gin.Engine {
	if !debug {
		gin.SetMode(gin.ReleaseMode) // Set Gin to release mode in production
	}

	router := gin.New() // Create a new Gin router

	// Global Middlewares
	router.Use(middleware.StructuredLogger()) // Custom structured logger middleware
	router.Use(gin.Recovery())                // Recovers from panics and writes a 500
	// router.Use(gin.Timeout(time.Second * 10)) // Set a global timeout for requests

	// Public routes (no authentication required)
	public := router.Group("/api/v1")
	{
		public.POST("/register", userHandler.Register) // User registration
		public.POST("/login", userHandler.Login)       // User login
	}

	// Authenticated routes (require JWT token)
	authenticated := router.Group("/api/v1")
	// Apply the authentication middleware to this group
	authenticated.Use(middleware.AuthMiddleware(jwtManager))
	{
		// User routes
		authenticated.GET("/user", userHandler.GetUser) // Get authenticated user's profile

		// Product routes
		authenticated.POST("/products", productHandler.AddProduct)          // Add a new product
		authenticated.GET("/products/:id", productHandler.GetProduct)       // Get a single product by ID
		authenticated.GET("/products", productHandler.GetProducts)          // Get all products for the authenticated user
		authenticated.PUT("/products/:id", productHandler.UpdateProduct)    // Update an existing product
		authenticated.DELETE("/products/:id", productHandler.DeleteProduct) // Delete a product
	}

	return router
}
