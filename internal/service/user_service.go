package service

import (
	"context"
	"errors"
	"fmt"
	"gotemplate/internal/models"
	"gotemplate/internal/repository"
	"gotemplate/pkg/auth"
	"gotemplate/pkg/logger"

	// "github.com/google/uuid" // No longer needed for UUID generation if ID is uint
	"go.uber.org/zap"            // Import zap for structured logging
	"golang.org/x/crypto/bcrypt" // For password hashing
)

// UserService defines the interface for user-related business logic
type UserService interface {
	RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	LoginUser(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	GetUserProfile(ctx context.Context, userID uint) (*models.User, error) // Changed userID to uint
}

// userService implements UserService
type userService struct {
	userRepo   repository.UserRepository // Dependency on UserRepository
	jwtManager *auth.JWTManager          // Dependency on JWTManager
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// RegisterUser handles user registration
func (s *userService) RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Check if a user with the given email already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		logger.Warn("Attempted registration with existing email", zap.String("email", req.Email))
		return nil, errors.New("user with this email already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password during registration", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create a new user model
	// ID, CreatedAt, UpdatedAt are handled by gorm.Model and the repository's raw SQL returning clause
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	// Save the user to the database
	// The repository method is responsible for setting the user.ID after successful creation
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		logger.Error("Failed to create user in repository", zap.Error(err), zap.String("email", req.Email))
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	logger.Info("User registered successfully", zap.Uint("userID", user.ID), zap.String("email", user.Email))
	return user, nil
}

// LoginUser handles user login and token generation
func (s *userService) LoginUser(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Retrieve the user by email
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Warn("Login attempt with non-existent email", zap.String("email", req.Email), zap.Error(err))
		return nil, errors.New("invalid credentials") // Generic error for security
	}

	// Compare the provided password with the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.Warn("Login attempt with incorrect password", zap.String("email", req.Email))
		return nil, errors.New("invalid credentials") // Generic error for security
	}

	// Generate a JWT token
	// JWTManager typically expects string IDs, so convert uint to string here
	token, err := s.jwtManager.GenerateToken(fmt.Sprintf("%d", user.ID))
	if err != nil {
		logger.Error("Failed to generate JWT token during login", zap.Error(err), zap.Uint("userID", user.ID))
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	logger.Info("User logged in successfully", zap.Uint("userID", user.ID), zap.String("email", user.Email))
	return &models.LoginResponse{Token: token}, nil
}

// GetUserProfile retrieves a user's profile by their ID
func (s *userService) GetUserProfile(ctx context.Context, userID uint) (*models.User, error) { // Changed userID to uint
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user profile by ID", zap.Error(err), zap.Uint("userID", userID))
		return nil, fmt.Errorf("user not found: %w", err)
	}
	logger.Debug("User profile retrieved", zap.Uint("userID", userID))
	return user, nil
}
