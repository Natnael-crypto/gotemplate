package service

import (
	"context"
	"fmt"
	"gotemplate/internal/models"
	"gotemplate/internal/repository"
	"gotemplate/pkg/logger"

	// Added for string to uint conversion
	// "github.com/google/uuid" // No longer needed for UUID generation
	"go.uber.org/zap"
)

// ProductService defines the interface for product-related business logic
type ProductService interface {
	// Changed userID and productID to uint
	AddProduct(ctx context.Context, userID uint, req *models.AddProductRequest) (*models.Product, error)
	GetProduct(ctx context.Context, productID uint) (*models.Product, error)
	GetProductsByOwner(ctx context.Context, userID uint) ([]*models.Product, error)
	UpdateProduct(ctx context.Context, productID uint, userID uint, req *models.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, productID uint, userID uint) error
}

// productService implements ProductService
type productService struct {
	productRepo repository.ProductRepository // Dependency on ProductRepository
}

// NewProductService creates a new ProductService instance
func NewProductService(productRepo repository.ProductRepository) ProductService {
	return &productService{
		productRepo: productRepo,
	}
}

// AddProduct adds a new product for a user
func (s *productService) AddProduct(ctx context.Context, userID uint, req *models.AddProductRequest) (*models.Product, error) {
	product := &models.Product{
		// ID, CreatedAt, UpdatedAt are handled by gorm.Model and the repository's raw SQL returning clause
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		UserID:      userID, // UserID is now uint
	}

	if err := s.productRepo.AddProduct(ctx, product); err != nil {
		logger.Error("Failed to add product in repository", zap.Error(err), zap.Uint("userID", userID)) // Changed userID to uint
		return nil, fmt.Errorf("failed to add product: %w", err)
	}

	logger.Info("Product added successfully", zap.Uint("productID", product.ID), zap.Uint("userID", userID)) // Changed userID and productID to uint
	return product, nil
}

// GetProduct retrieves a product by its ID
func (s *productService) GetProduct(ctx context.Context, productID uint) (*models.Product, error) { // Changed productID to uint
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		logger.Error("Failed to get product by ID in repository", zap.Error(err), zap.Uint("productID", productID)) // Changed productID to uint
		return nil, fmt.Errorf("product not found: %w", err)
	}
	logger.Debug("Product retrieved", zap.Uint("productID", productID)) // Changed productID to uint
	return product, nil
}

// GetProductsByOwner retrieves all products owned by a specific user
func (s *productService) GetProductsByOwner(ctx context.Context, userID uint) ([]*models.Product, error) { // Changed userID to uint
	products, err := s.productRepo.GetProductsByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get products by user ID in repository", zap.Error(err), zap.Uint("userID", userID)) // Changed userID to uint
		return nil, fmt.Errorf("failed to retrieve products: %w", err)
	}
	logger.Debug("Products retrieved by owner", zap.Uint("userID", userID), zap.Int("count", len(products))) // Changed userID to uint
	return products, nil
}

// UpdateProduct updates an existing product. Ensures the product belongs to the user.
func (s *productService) UpdateProduct(ctx context.Context, productID uint, userID uint, req *models.UpdateProductRequest) (*models.Product, error) { // Changed IDs to uint
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		logger.Error("Product not found for update", zap.Error(err), zap.Uint("productID", productID)) // Changed productID to uint
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Authorization check: ensure the current user owns the product
	// Both product.UserID and userID are now uint
	if product.UserID != userID {
		logger.Warn("Unauthorized attempt to update product", zap.Uint("productID", productID), zap.Uint("attemptingUserID", userID), zap.Uint("productOwnerID", product.UserID))
		return nil, fmt.Errorf("you are not authorized to update this product")
	}

	// Update fields if provided
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price != 0 { // Price is float64, check for non-zero
		product.Price = req.Price
	}
	// product.UpdatedAt is handled by the repository's raw SQL update (SET updated_at = ?)
	// It's still good practice to set it here if you need its value immediately for return,
	// but the DB update relies on the repository.
	// product.UpdatedAt = time.Now() // No longer strictly necessary here, but doesn't hurt

	if err := s.productRepo.UpdateProduct(ctx, product); err != nil {
		logger.Error("Failed to update product in repository", zap.Error(err), zap.Uint("productID", productID)) // Changed productID to uint
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	logger.Info("Product updated successfully", zap.Uint("productID", product.ID)) // Changed productID to uint
	return product, nil
}

// DeleteProduct deletes a product. Ensures the product belongs to the user.
func (s *productService) DeleteProduct(ctx context.Context, productID uint, userID uint) error { // Changed IDs to uint
	product, err := s.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		logger.Error("Product not found for deletion", zap.Error(err), zap.Uint("productID", productID)) // Changed productID to uint
		return fmt.Errorf("product not found: %w", err)
	}

	// Authorization check: ensure the current user owns the product
	if product.UserID != userID {
		logger.Warn("Unauthorized attempt to delete product", zap.Uint("productID", productID), zap.Uint("attemptingUserID", userID), zap.Uint("productOwnerID", product.UserID))
		return fmt.Errorf("you are not authorized to delete this product")
	}

	if err := s.productRepo.DeleteProduct(ctx, productID); err != nil {
		logger.Error("Failed to delete product in repository", zap.Error(err), zap.Uint("productID", productID)) // Changed productID to uint
		return fmt.Errorf("failed to delete product: %w", err)
	}

	logger.Info("Product deleted successfully", zap.Uint("productID", productID)) // Changed productID to uint
	return nil
}
