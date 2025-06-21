package database

import (
	"context"
	"fmt"
	"time"

	"gotemplate/config"
	"gotemplate/internal/models" // Import your models package here!
	"gotemplate/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres" // GORM PostgreSQL driver
	"gorm.io/gorm"            // GORM main package
)

// NewPostgresDB establishes a new PostgreSQL database connection using GORM
// and performs auto-migration.
// It now returns *gorm.DB directly.
func NewPostgresDB(cfg *config.DatabaseConfig) (*gorm.DB, error) { // Changed return type
	// Construct the DSN (Data Source Name) for GORM
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	// Open connection with GORM
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// You can add GORM configurations here, e.g., Logger, NamingStrategy
		// Logger: logger.NewGormLogger(), // If you create a custom GORM logger
	})
	if err != nil {
		logger.Error("Failed to connect to database using GORM", zap.Error(err),
			zap.String("host", cfg.Host),
			zap.String("port", cfg.Port),
			zap.String("db_name", cfg.DBName))
		return nil, fmt.Errorf("failed to connect to database using GORM: %w", err)
	}

	// Get the underlying sql.DB to set connection pool settings and ping
	sqlDB, err := gormDB.DB()
	if err != nil {
		logger.Error("Failed to get underlying sql.DB from GORM", zap.Error(err))
		return nil, fmt.Errorf("failed to get underlying sql.DB from GORM: %w", err)
	}

	// Set connection pool settings (optional, but recommended)
	sqlDB.SetMaxIdleConns(10)                 // Maximum number of idle connections in the pool
	sqlDB.SetMaxOpenConns(100)                // Maximum number of open connections to the database
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Maximum amount of time a connection may be reused

	// Ping the database to verify the connection.
	pingCtx, cancelPing := context.WithTimeout(context.Background(), 3*time.Second) // Ping timeout
	defer cancelPing()

	err = sqlDB.PingContext(pingCtx)
	if err != nil {
		sqlDB.Close() // Close the underlying connection if ping fails
		logger.Error("Failed to ping database after GORM connection", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database after GORM connection: %w", err)
	}

	logger.Info("Successfully connected to PostgreSQL database with GORM",
		zap.String("host", cfg.Host),
		zap.String("db_name", cfg.DBName))

	// Perform auto-migration
	// Pass all your model structs here
	err = gormDB.AutoMigrate(
		&models.User{},
		&models.Product{}, // Make sure to uncomment or add all your GORM models here!
	)
	if err != nil {
		logger.Error("Failed to perform GORM auto-migration", zap.Error(err))
		return nil, fmt.Errorf("failed to perform GORM auto-migration: %w", err)
	}

	logger.Info("GORM auto-migration completed successfully")

	// Return *gorm.DB directly
	return gormDB, nil
}

// Close function is now a standalone helper or could be a method on a struct
// that holds the *gorm.DB if you still want to encapsulate it.
// For now, let's keep it as a simple function that takes *gorm.DB.
func CloseDB(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("Failed to get underlying sql.DB for closing", zap.Error(err))
			return
		}
		if sqlDB != nil {
			err := sqlDB.Close()
			if err != nil {
				logger.Error("Failed to close GORM database connection pool", zap.Error(err))
			} else {
				logger.Info("PostgreSQL database connection pool closed by GORM")
			}
		}
	}
}
