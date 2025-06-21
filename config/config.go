package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper" // Import viper for configuration management
)

// Config holds all application configurations
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig holds server-related configurations
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Debug        bool
}

// DatabaseConfig holds database-related configurations
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT-related configurations
type JWTConfig struct {
	SecretKey     string
	ExpiresInHour time.Duration // Token expiration time in hours
}

// LoadConfig loads configuration from environment variables or a config file
func LoadConfig() (*Config, error) {
	// Set the file name (without extension) for the config file
	viper.SetConfigName("config")
	// Set the type of the config file
	viper.SetConfigType("yaml") // or "json", "toml", etc.
	// Add paths where viper should look for the config file
	viper.AddConfigPath("./config") // Look in the 'config' directory
	viper.AddConfigPath(".")        // Look in the current directory

	// Read in the config file
	if err := viper.ReadInConfig(); err != nil {
		// If config file is not found, try to load from environment variables
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Config file not found, loading from environment variables...")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Set default values (important for clarity and robustness)
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.readTimeout", "10s")
	viper.SetDefault("server.writeTimeout", "10s")
	viper.SetDefault("server.debug", false)

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "user")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "yourdb")
	viper.SetDefault("database.sslmode", "disable")

	viper.SetDefault("jwt.expiresInHour", "24h") // Default JWT expiration is 24 hours

	// Bind environment variables to config keys (e.g., APP_SERVER_PORT maps to server.port)
	// This allows overriding config file settings with env vars
	// viper.SetEnvPrefix("APP") // Prefix for environment variables (e.g., APP_SERVER_PORT)
	viper.AutomaticEnv()      // Automatically bind environment variables

	// Manual binding for specific environment variables if needed
	_ = viper.BindEnv("SERVER_PORT", "APP_SERVER_PORT")
	_ = viper.BindEnv("DATABASE_HOST", "APP_DATABASE_HOST")
	_ = viper.BindEnv("DATABASE_PORT", "APP_DATABASE_PORT")
	_ = viper.BindEnv("DATABASE_USER", "APP_DATABASE_USER")
	_ = viper.BindEnv("DATABASE_PASSWORD", "APP_DATABASE_PASSWORD")
	_ = viper.BindEnv("DATABASE_DBNAME", "APP_DATABASE_DBNAME")
	_ = viper.BindEnv("DATABASE_SSLMODE", "APP_DATABASE_SSLMODE")
	_ = viper.BindEnv("JWT_SECRET_KEY", "APP_JWT_SECRET_KEY")

	// Check for mandatory JWT_SECRET_KEY
	if viper.GetString("jwt.secretKey") == "" {
		// If not set via config file or env, check for APP_JWT_SECRET_KEY
		if os.Getenv("APP_JWT_SECRET_KEY") == "" {
			return nil, fmt.Errorf("JWT_SECRET_KEY or APP_JWT_SECRET_KEY is not set in config or environment variables")
		}
	}

	var cfg Config
	// Unmarshal the loaded configurations into the Config struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
