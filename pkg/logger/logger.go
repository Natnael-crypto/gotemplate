package logger

import (
	"log"

	"go.uber.org/zap"        // Import zap for structured logging
	"go.uber.org/zap/zapcore" // Import zapcore for logger configuration
)

var (
	ZapLogger *zap.Logger // Global ZapLogger instance
)

// InitLogger initializes the Zap logger
func InitLogger(debug bool) {
	var err error
	var config zap.Config

	if debug {
		// Development configuration: human-readable output, DPanicLevel and above
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add colors for development
	} else {
		// Production configuration: JSON format, InfoLevel and above
		config = zap.NewProductionConfig()
	}

	// Customize encoder settings
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // ISO8601 format for timestamps
	config.EncoderConfig.TimeKey = "timestamp"                   // Field name for timestamp
	config.EncoderConfig.StacktraceKey = "stacktrace"            // Field name for stacktrace

	ZapLogger, err = config.Build(zap.AddCallerSkip(1)) // Build the logger, skip 1 caller for correct file/line number
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err) // Use standard log if Zap fails
	}
	zap.ReplaceGlobals(ZapLogger) // Replace global Zap logger with our configured one
}

// For convenience, provide wrapper functions
func Debug(msg string, fields ...zap.Field) {
	ZapLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	ZapLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	ZapLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	ZapLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	ZapLogger.Fatal(msg, fields...)
}