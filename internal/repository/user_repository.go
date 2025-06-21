package repository

import (
	"context"
	"fmt"
	"gotemplate/internal/models"
	"gotemplate/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	// Add other user-related methods as needed
}

// postgresUserRepository implements UserRepository using GORM with raw SQL
type postgresUserRepository struct {
	db *gorm.DB
}

// NewPostgresUserRepository creates a new UserRepository instance
func NewPostgresUserRepository(db *gorm.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

// CreateUser inserts a new user into the database using raw SQL
func (r *postgresUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	// SQL query to insert a new user.
	// We're inserting Username, Email, Password.
	// GORM's gorm.Model fields (ID, CreatedAt, UpdatedAt) are typically handled by the DB itself
	// or GORM's hooks. For raw SQL, we'll let the DB assign ID and timestamps.
	// We might need to fetch them back if the application needs the generated ID immediately.
	sqlQuery := `INSERT INTO users (username, email, password, created_at, updated_at) VALUES (?, ?, ?, ?, ?) RETURNING id`

	// Use Exec for DML operations (INSERT, UPDATE, DELETE)
	// For RETURNING ID, we can use Scan, but Exec also returns RowsAffected.
	// To get the ID back, a separate query or a specific DB driver feature might be needed if RETURNING is not directly supported by Exec.
	// For PostgreSQL, RETURNING is common. GORM's Raw().Scan() is better for this.
	var newID uint
	result := r.db.WithContext(ctx).Raw(sqlQuery,
		user.Username,
		user.Email,
		user.Password,
		time.Now(), // Manually set timestamps for raw insert
		time.Now(),
	).Scan(&newID) // Scan the returned ID into newID

	if result.Error != nil {
		logger.Error("Failed to create user in DB using raw SQL", zap.Error(result.Error), zap.String("email", user.Email))
		return fmt.Errorf("failed to create user: %w", result.Error)
	}

	user.ID = newID // Set the ID on the user model

	logger.Info("User created in DB successfully using raw SQL", zap.Uint("userID", user.ID), zap.String("email", user.Email))
	return nil
}

// GetUserByEmail retrieves a user by their email address using raw SQL
func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	sqlQuery := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE email = ?`

	// Use Raw().Scan() for SELECT queries where you want to scan results into a struct
	result := r.db.WithContext(ctx).Raw(sqlQuery, email).Scan(user)
	fmt.Println(user)
	if result.Error != nil || user.ID == 0 {
		if result.Error == gorm.ErrRecordNotFound {
			logger.Warn("User not found by email using raw SQL", zap.String("email", email))
			return nil, fmt.Errorf("user not found with email %s", email)
		}
		logger.Error("Failed to retrieve user by email from DB using raw SQL", zap.Error(result.Error), zap.String("email", email))
		return nil, fmt.Errorf("database error retrieving user by email: %w", result.Error)
	}
	logger.Debug("User retrieved by email using raw SQL", zap.String("email", user.Email))
	return user, nil
}

// GetUserByID retrieves a user by their ID using raw SQL
func (r *postgresUserRepository) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	user := &models.User{}
	sqlQuery := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = ?`

	result := r.db.WithContext(ctx).Raw(sqlQuery, id).Scan(user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logger.Warn("User not found by ID using raw SQL", zap.Uint("userID", id))
			return nil, fmt.Errorf("user not found with ID %d", id)
		}
		logger.Error("Failed to retrieve user by ID from DB using raw SQL", zap.Error(result.Error), zap.Uint("userID", id))
		return nil, fmt.Errorf("database error retrieving user by ID: %w", result.Error)
	}
	logger.Debug("User retrieved by ID using raw SQL", zap.Uint("userID", user.ID))
	return user, nil
}