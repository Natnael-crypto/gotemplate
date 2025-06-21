package auth

import (
	"fmt"
	"gotemplate/config"
	"gotemplate/pkg/logger"
	"time"

	"github.com/golang-jwt/jwt/v5" // JWT library
	"go.uber.org/zap"
)

// Claims defines the JWT custom claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secretKey     string
	expiresInHour time.Duration
}

// NewJWTManager creates a new JWTManager instance
func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey:     cfg.SecretKey,
		expiresInHour: cfg.ExpiresInHour,
	}
}

// GenerateToken generates a new JWT token for a given user ID
func (jm *JWTManager) GenerateToken(userID string) (string, error) {
	// Define the expiration time for the token
	expirationTime := time.Now().Add(jm.expiresInHour) // Use configured expiration

	// Create the JWT claims, including the user ID and standard claims
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // Token expiration time
			IssuedAt:  jwt.NewNumericDate(time.Now()),     // Token issuance time
			NotBefore: jwt.NewNumericDate(time.Now()),     // Token not valid before this time
			Issuer:    "*",                                // Token issuer
			Subject:   userID,                             // Token subject (typically the user ID)
		},
	}

	// Create the token with the specified signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(jm.secretKey))
	if err != nil {
		logger.Error("Failed to sign JWT token", zap.Error(err), zap.String("userID", userID))
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	logger.Info("JWT token generated successfully", zap.String("userID", userID))
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims if valid
func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token with the custom claims type and a key function
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return []byte(jm.secretKey), nil
	})

	if err != nil {
		logger.Error("Failed to parse JWT token", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if the token is valid
	if !token.Valid {
		logger.Warn("Invalid JWT token")
		return nil, fmt.Errorf("token is invalid")
	}

	// Assert the claims to our custom Claims type
	claims, ok := token.Claims.(*Claims)
	if !ok {
		logger.Error("Failed to get claims from JWT token")
		return nil, fmt.Errorf("invalid token claims")
	}

	logger.Debug("JWT token validated successfully", zap.String("userID", claims.UserID))
	return claims, nil
}
