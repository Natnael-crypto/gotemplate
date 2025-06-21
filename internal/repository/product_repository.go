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

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	AddProduct(ctx context.Context, product *models.Product) error
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
	GetProductsByUserID(ctx context.Context, userID uint) ([]*models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id uint) error
	// Add other product-related methods
}

// postgresProductRepository implements ProductRepository using GORM with raw SQL
type postgresProductRepository struct {
	db *gorm.DB
}

// NewPostgresProductRepository creates a new ProductRepository instance
func NewPostgresProductRepository(db *gorm.DB) ProductRepository {
	return &postgresProductRepository{db: db}
}

// AddProduct inserts a new product into the database using raw SQL
func (r *postgresProductRepository) AddProduct(ctx context.Context, product *models.Product) error {
	sqlQuery := `INSERT INTO products (name, description, price, user_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id`

	var newID uint
	result := r.db.WithContext(ctx).Raw(sqlQuery,
		product.Name,
		product.Description,
		product.Price,
		product.UserID,
		time.Now(), // Manually set timestamps
		time.Now(),
	).Scan(&newID)

	if result.Error != nil {
		logger.Error("Failed to add product to DB using raw SQL", zap.Error(result.Error), zap.String("productName", product.Name))
		return fmt.Errorf("failed to add product: %w", result.Error)
	}

	product.ID = newID // Set the ID on the product model

	logger.Info("Product added to DB successfully using raw SQL", zap.Uint("productID", product.ID))
	return nil
}

// GetProductByID retrieves a product by its ID using raw SQL
func (r *postgresProductRepository) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	product := &models.Product{}
	sqlQuery := `SELECT id, name, description, price, user_id, created_at, updated_at FROM products WHERE id = ?`

	result := r.db.WithContext(ctx).Raw(sqlQuery, id).Scan(product)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logger.Warn("Product not found by ID using raw SQL", zap.Uint("productID", id))
			return nil, fmt.Errorf("product not found with ID %d", id)
		}
		logger.Error("Failed to retrieve product by ID from DB using raw SQL", zap.Error(result.Error), zap.Uint("productID", id))
		return nil, fmt.Errorf("database error retrieving product by ID: %w", result.Error)
	}
	logger.Debug("Product retrieved by ID using raw SQL", zap.Uint("productID", product.ID))
	return product, nil
}

// GetProductsByUserID retrieves all products for a given user ID using raw SQL
func (r *postgresProductRepository) GetProductsByUserID(ctx context.Context, userID uint) ([]*models.Product, error) {
	var products []*models.Product
	sqlQuery := `SELECT id, name, description, price, user_id, created_at, updated_at FROM products WHERE user_id = ?`

	// Use Raw().Scan() to populate a slice of structs
	result := r.db.WithContext(ctx).Raw(sqlQuery, userID).Scan(&products)
	if result.Error != nil {
		logger.Error("Failed to get products by user ID from DB using raw SQL", zap.Error(result.Error), zap.Uint("userID", userID))
		return nil, fmt.Errorf("failed to get products by user ID: %w", result.Error)
	}
	logger.Debug("Products retrieved by user ID using raw SQL", zap.Uint("userID", userID), zap.Int("count", len(products)))
	return products, nil
}

// UpdateProduct updates an existing product in the database using raw SQL
func (r *postgresProductRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	sqlQuery := `UPDATE products SET name = ?, description = ?, price = ?, updated_at = ? WHERE id = ?`

	// Use Exec for UPDATE operations
	result := r.db.WithContext(ctx).Exec(sqlQuery,
		product.Name,
		product.Description,
		product.Price,
		time.Now(), // Manually update updated_at
		product.ID,
	)
	if result.Error != nil {
		logger.Error("Failed to update product in DB using raw SQL", zap.Error(result.Error), zap.Uint("productID", product.ID))
		return fmt.Errorf("failed to update product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("product with ID %d not found for update (raw SQL)", product.ID)
	}
	logger.Info("Product updated in DB successfully using raw SQL", zap.Uint("productID", product.ID))
	return nil
}

// DeleteProduct deletes a product from the database using raw SQL
func (r *postgresProductRepository) DeleteProduct(ctx context.Context, id uint) error {
	// For hard delete:
	sqlQuery := `DELETE FROM products WHERE id = ?`

	// For soft delete (if you want to use GORM's soft delete behavior with raw SQL,
	// you'd update the `deleted_at` column manually):
	// sqlQuery := `UPDATE products SET deleted_at = ? WHERE id = ?`

	result := r.db.WithContext(ctx).Exec(sqlQuery, id) // Pass id as a parameter

	if result.Error != nil {
		logger.Error("Failed to delete product from DB using raw SQL", zap.Error(result.Error), zap.Uint("productID", id))
		return fmt.Errorf("failed to delete product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("product with ID %d not found for deletion (raw SQL)", id)
	}
	logger.Info("Product deleted from DB successfully using raw SQL", zap.Uint("productID", id))
	return nil
}
