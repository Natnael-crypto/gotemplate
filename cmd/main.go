package main

import (
	"context"
	"fmt"
	"gotemplate/config"
	"gotemplate/internal/handler"
	"gotemplate/internal/repository"
	"gotemplate/internal/router"
	"gotemplate/internal/service"
	"gotemplate/pkg/auth"
	"gotemplate/pkg/database"
	"gotemplate/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "github.com/joho/godotenv" // For loading .env files locally
	"go.uber.org/zap" // Import zap for structured logging
)

func main() {
	// Load environment variables from .env file (for local development)
	// In production, env vars are typically set directly.
	// if err := godotenv.Load(); err != nil {
	// 	logger.Warn(".env file not found, assuming environment variables are set externally.", zap.Error(err))
	// }

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// Using standard logger here because Zap might not be fully initialized yet
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize the logger based on debug mode from config
	logger.InitLogger(cfg.Server.Debug)
	defer func() {
		// Ensure all buffered logs are flushed before exiting
		if err := logger.ZapLogger.Sync(); err != nil {
			fmt.Printf("Error syncing logger: %v\n", err)
		}
	}()
	logger.Info("Application starting...", zap.Bool("debug_mode", cfg.Server.Debug))

	// Initialize database connection
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	fmt.Println("Database auto-migration completed successfully!")
	defer database.CloseDB(db)

	// --- Dependency Injection ---
	// Instantiate Repositories
	userRepo := repository.NewPostgresUserRepository(db)
	productRepo := repository.NewPostgresProductRepository(db)

	// Instantiate JWT Manager
	jwtManager := auth.NewJWTManager(&cfg.JWT)

	// Instantiate Services with their respective repositories and managers
	userService := service.NewUserService(userRepo, jwtManager)
	productService := service.NewProductService(productRepo)

	// Instantiate Handlers with their respective services
	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService)

	// Setup Gin Router with all handlers and middleware
	r := router.SetupRouter(userHandler, productHandler, jwtManager, cfg.Server.Debug)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,   // Server address (e.g., ":8080")
		Handler:      r,                       // Gin router as the handler
		ReadTimeout:  cfg.Server.ReadTimeout,  // Timeout for reading request body
		WriteTimeout: cfg.Server.WriteTimeout, // Timeout for writing response body
		IdleTimeout:  time.Minute * 2,         // Timeout for idle connections
	}

	// Start the server in a goroutine so it doesn't block the main thread
	go func() {
		logger.Info("Server listening", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to listen", zap.Error(err))
		}
	}()

	// --- Graceful Shutdown ---
	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	// Trap OS signals: Interrupt (Ctrl+C) and Terminate
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received

	logger.Info("Shutting down server...")

	// Create a context with a timeout for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 5-second shutdown timeout
	defer cancel()

	// Shutdown the HTTP server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
