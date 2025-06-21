package models

import (
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	gorm.Model // Embed gorm.Model for ID, CreatedAt, UpdatedAt, DeletedAt
	// ID        uint is provided by gorm.Model
	Username string `gorm:"unique;not null"` // Add unique and not null constraints
	Email    string `gorm:"unique;not null"` // Add unique and not null constraints
	Password string `gorm:"not null"`        // Store hashed password, not null
	// CreatedAt time.Time is provided by gorm.Model
	// UpdatedAt time.Time is provided by gorm.Model
}

// RegisterRequest is the payload for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest is the payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse contains the JWT token after successful login
type LoginResponse struct {
	Token string `json:"token"`
}
