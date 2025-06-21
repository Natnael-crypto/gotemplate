package handler

import (
	"gotemplate/internal/models"
	"gotemplate/internal/service"
	"gotemplate/pkg/logger"
	"net/http"
	"strconv" // Import for string to uint conversion

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserHandler defines the interface for user HTTP handlers
type UserHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	GetUser(c *gin.Context)
}

// userHandler implements UserHandler
type userHandler struct {
	userService service.UserService // Dependency on UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService service.UserService) UserHandler {
	return &userHandler{
		userService: userService,
	}
}

// Register handles user registration requests
func (h *userHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	// Bind JSON request body to the struct and validate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid register request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service layer to register the user
	user, err := h.userService.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to register user", zap.Error(err), zap.String("email", req.Email))
		// Handle specific errors for better client feedback
		if err.Error() == "user with this email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		}
		return
	}

	logger.Info("User registered successfully via API", zap.Uint("userID", user.ID)) // Use zap.Uint
	// Return a success response (excluding password)
	c.JSON(http.StatusCreated, gin.H{
		"message":  "User registered successfully",
		"id":       user.ID, // ID is now uint, will be marshaled as number
		"username": user.Username,
		"email":    user.Email,
	})
}

// Login handles user login requests
func (h *userHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	// Bind JSON request body to the struct and validate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid login request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service layer to log in the user and get a token
	res, err := h.userService.LoginUser(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to login user", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()}) // Return generic "invalid credentials"
		return
	}

	logger.Info("User logged in successfully via API", zap.String("email", req.Email))
	// Return the JWT token
	c.JSON(http.StatusOK, res)
}

// GetUser handles retrieving a user's profile
func (h *userHandler) GetUser(c *gin.Context) {
	// Retrieve userID from context, set by the AuthMiddleware
	// Now expecting a string from JWT parsing, convert to uint for service
	userIDFromContext, exists := c.Get("userID")
	if !exists {
		logger.Error("userID not found in context (AuthMiddleware issue)", zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		return
	}

	// AuthMiddleware likely puts string into context, so convert it to uint
	idStr, ok := userIDFromContext.(string)
	if !ok {
		logger.Error("userID in context is not a string, or unexpected type", zap.Any("userID", userIDFromContext), zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	idUint, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.Error("Failed to parse userID from context to uint", zap.Error(err), zap.String("userIDStr", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}
	userID := uint(idUint) // Convert to uint

	// Call the service layer to get the user profile
	user, err := h.userService.GetUserProfile(c.Request.Context(), userID) // Pass uint
	if err != nil {
		logger.Error("Failed to get user profile", zap.Error(err), zap.Uint("userID", userID)) // Use zap.Uint
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	logger.Info("User profile retrieved successfully via API", zap.Uint("userID", user.ID)) // Use zap.Uint
	// Return user profile (excluding password)
	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID, // ID is now uint
		"username":  user.Username,
		"email":     user.Email,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
	})
}
