package models

import (
	"gorm.io/gorm"
)

// Product represents a product in the system
type Product struct {
	gorm.Model // Embed gorm.Model for ID, CreatedAt, UpdatedAt, DeletedAt
	// ID        uint is provided by gorm.Model
	Name        string  `gorm:"not null"` // Name cannot be null
	Description string
	Price       float64 `gorm:"not null;check:price > 0"` // Price cannot be null and must be greater than 0
	UserID      uint    `gorm:"not null"`                 // Foreign key for User, GORM automatically infers `user_id` column
	User        User    // Belongs To relationship with User
	// CreatedAt time.Time is provided by gorm.Model
	// UpdatedAt time.Time is provided by gorm.Model
}

// AddProductRequest is the payload for adding a new product
type AddProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
}

// UpdateProductRequest is the payload for updating an existing product
type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price,omitempty" binding:"omitempty,gt=0"` // omitempty if not provided, gt=0 if provided
}