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

// ProductHandler defines the interface for product HTTP handlers
type ProductHandler interface {
	AddProduct(c *gin.Context)
	GetProduct(c *gin.Context)
	GetProducts(c *gin.Context)
	UpdateProduct(c *gin.Context)
	DeleteProduct(c *gin.Context)
}

// productHandler implements ProductHandler
type productHandler struct {
	productService service.ProductService // Dependency on ProductService
}

// NewProductHandler creates a new ProductHandler instance
func NewProductHandler(productService service.ProductService) ProductHandler {
	return &productHandler{
		productService: productService,
	}
}

// AddProduct handles adding a new product
func (h *productHandler) AddProduct(c *gin.Context) {
	// Get userID from context (set by AuthMiddleware)
	userIDFromContext, exists := c.Get("userID")
	
	if !exists {
		logger.Error("userID not found in context for AddProduct", zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		return
	}
	idStr, ok := userIDFromContext.(string) // Expecting string from JWT parsing
	if !ok {
		logger.Error("userID in context is not a string for AddProduct", zap.Any("userID", userIDFromContext), zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	userID, err := strconv.ParseUint(idStr, 10, 64) // Convert to uint
	if err != nil {
		logger.Error("Failed to parse userID from context to uint for AddProduct", zap.Error(err), zap.String("userIDStr", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req models.AddProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid AddProduct request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.AddProduct(c.Request.Context(), uint(userID), &req) // Pass uint
	if err != nil {
		logger.Error("Failed to add product", zap.Error(err), zap.Uint("userID", uint(userID))) // Use zap.Uint
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product"})
		return
	}

	logger.Info("Product added successfully via API", zap.Uint("productID", product.ID), zap.Uint("userID", uint(userID))) // Use zap.Uint
	c.JSON(http.StatusCreated, product)                                                                                    // Product will be marshaled correctly with uint ID
}

// GetProduct handles retrieving a single product by ID
func (h *productHandler) GetProduct(c *gin.Context) {
	productIDStr := c.Param("id") // Get product ID from URL parameter (string)
	if productIDStr == "" {
		logger.Warn("Product ID is missing in GetProduct request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	productID, err := strconv.ParseUint(productIDStr, 10, 64) // Convert to uint
	if err != nil {
		logger.Warn("Invalid product ID format in GetProduct request", zap.Error(err), zap.String("productIDStr", productIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	product, err := h.productService.GetProduct(c.Request.Context(), uint(productID)) // Pass uint
	if err != nil {
		logger.Error("Failed to get product", zap.Error(err), zap.Uint("productID", uint(productID))) // Use zap.Uint
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	logger.Info("Product retrieved successfully via API", zap.Uint("productID", product.ID)) // Use zap.Uint
	c.JSON(http.StatusOK, product)                                                           // Product will be marshaled correctly with uint ID
}

// GetProducts handles retrieving all products for the authenticated user
func (h *productHandler) GetProducts(c *gin.Context) {
	userIDFromContext, exists := c.Get("userID") // Get userID from context
	if !exists {
		logger.Error("userID not found in context for GetProducts", zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		return
	}
	idStr, ok := userIDFromContext.(string)
	if !ok {
		logger.Error("userID in context is not a string for GetProducts", zap.Any("userID", userIDFromContext), zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	userID, err := strconv.ParseUint(idStr, 10, 64) // Convert to uint
	if err != nil {
		logger.Error("Failed to parse userID from context to uint for GetProducts", zap.Error(err), zap.String("userIDStr", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	products, err := h.productService.GetProductsByOwner(c.Request.Context(), uint(userID)) // Pass uint
	if err != nil {
		logger.Error("Failed to get products for user", zap.Error(err), zap.Uint("userID", uint(userID))) // Use zap.Uint
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	logger.Info("Products retrieved successfully for user via API", zap.Uint("userID", uint(userID)), zap.Int("count", len(products))) // Use zap.Uint
	c.JSON(http.StatusOK, products)
}

// UpdateProduct handles updating an existing product
func (h *productHandler) UpdateProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	if productIDStr == "" {
		logger.Warn("Product ID is missing in UpdateProduct request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}
	productID, err := strconv.ParseUint(productIDStr, 10, 64) // Convert to uint
	if err != nil {
		logger.Warn("Invalid product ID format in UpdateProduct request", zap.Error(err), zap.String("productIDStr", productIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	userIDFromContext, exists := c.Get("userID") // Get userID from context
	if !exists {
		logger.Error("userID not found in context for UpdateProduct", zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		return
	}
	idStr, ok := userIDFromContext.(string)
	if !ok {
		logger.Error("userID in context is not a string for UpdateProduct", zap.Any("userID", userIDFromContext), zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	userID, err := strconv.ParseUint(idStr, 10, 64) // Convert to uint
	if err != nil {
		logger.Error("Failed to parse userID from context to uint for UpdateProduct", zap.Error(err), zap.String("userIDStr", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid UpdateProduct request payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.productService.UpdateProduct(c.Request.Context(), uint(productID), uint(userID), &req) // Pass uints
	if err != nil {
		logger.Error("Failed to update product", zap.Error(err), zap.Uint("productID", uint(productID)), zap.Uint("userID", uint(userID))) // Use zap.Uint
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "you are not authorized to update this product" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		}
		return
	}

	logger.Info("Product updated successfully via API", zap.Uint("productID", product.ID), zap.Uint("userID", uint(userID))) // Use zap.Uint
	c.JSON(http.StatusOK, product)
}

// DeleteProduct handles deleting a product
func (h *productHandler) DeleteProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	if productIDStr == "" {
		logger.Warn("Product ID is missing in DeleteProduct request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}
	productID, err := strconv.ParseUint(productIDStr, 10, 64) // Convert to uint
	if err != nil {
		logger.Warn("Invalid product ID format in DeleteProduct request", zap.Error(err), zap.String("productIDStr", productIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID format"})
		return
	}

	userIDFromContext, exists := c.Get("userID") // Get userID from context
	if !exists {
		logger.Error("userID not found in context for DeleteProduct", zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		return
	}
	idStr, ok := userIDFromContext.(string)
	if !ok {
		logger.Error("userID in context is not a string for DeleteProduct", zap.Any("userID", userIDFromContext), zap.String("path", c.Request.URL.Path))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	userID, err := strconv.ParseUint(idStr, 10, 64) // Convert to uint
	if err != nil {
		logger.Error("Failed to parse userID from context to uint for DeleteProduct", zap.Error(err), zap.String("userIDStr", idStr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	err = h.productService.DeleteProduct(c.Request.Context(), uint(productID), uint(userID)) // Pass uints
	if err != nil {
		logger.Error("Failed to delete product", zap.Error(err), zap.Uint("productID", uint(productID)), zap.Uint("userID", uint(userID))) // Use zap.Uint
		if err.Error() == "product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "you are not authorized to delete this product" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		}
		return
	}

	logger.Info("Product deleted successfully via API", zap.Uint("productID", uint(productID)), zap.Uint("userID", uint(userID))) // Use zap.Uint
	c.JSON(http.StatusNoContent, nil)                                                                                             // 204 No Content for successful deletion
}
