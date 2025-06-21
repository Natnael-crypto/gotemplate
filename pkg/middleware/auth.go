package middleware

import (
	"gotemplate/pkg/auth"
	"gotemplate/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin" // Import Gin
	"go.uber.org/zap"          // Import zap for structured logging
)

// AuthMiddleware creates a middleware that authenticates requests using JWT
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header from the request
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Authorization header missing", zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort() // Abort the request chain
			return
		}

		// Check if the header starts with "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid Authorization header format", zap.String("authHeader", authHeader))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		// Extract the token string
		tokenString := parts[1]

		// Validate the token using the JWTManager
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			logger.Error("JWT token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// If the token is valid, set the UserID in the Gin context for later use
		c.Set("userID", claims.UserID)
		logger.Debug("User authenticated", zap.String("userID", claims.UserID))

		// Continue to the next handler in the chain
		c.Next()
	}
}
