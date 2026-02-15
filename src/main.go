package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// Product matches the Product schema in api.yaml
type Product struct {
	ProductID    int    `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int    `json:"category_id"`
	Weight       int    `json:"weight"`
	SomeOtherID  int    `json:"some_other_id"`
}

// ErrorResponse matches the Error schema in api.yaml
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// In-memory store: hashmap for O(1) lookups
// sync.RWMutex for thread-safe concurrent access
var (
	products = make(map[int]Product)
	mu       sync.RWMutex
)

func main() {
	router := gin.Default()

	// Product endpoints per api.yaml
	router.GET("/products/:productId", getProduct)
	router.POST("/products/:productId/details", addProductDetails)

	// Health check (useful for ECS health checks)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.Run(":8080")
}

// getProduct handles GET /products/{productId}
// Returns 200 with product, 400 if bad ID, 404 if not found
func getProduct(c *gin.Context) {
	// Parse and validate productId
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid product ID",
			Details: "Product ID must be a positive integer",
		})
		return
	}

	// Lookup in map (read lock for concurrency safety)
	mu.RLock()
	product, exists := products[productID]
	mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Product not found",
			Details: "No product found with ID " + strconv.Itoa(productID),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// addProductDetails handles POST /products/{productId}/details
// Returns 204 on success, 400 if invalid input, 404 if path/body mismatch
func addProductDetails(c *gin.Context) {
	// Parse and validate productId from URL path
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil || productID < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid product ID in path",
			Details: "Product ID must be a positive integer",
		})
		return
	}

	// Bind JSON body
	var p Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Validate required fields and constraints
	if err := validateProduct(p); err != "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Validation failed",
			Details: err,
		})
		return
	}

	// Check that the path productId matches the body product_id
	if p.ProductID != productID {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "Product ID mismatch",
			Details: "Path product ID does not match body product_id",
		})
		return
	}

	// Store in memory (write lock)
	mu.Lock()
	products[productID] = p
	mu.Unlock()

	// 204 No Content on success
	c.Status(http.StatusNoContent)
}

// validateProduct checks all field constraints from the api.yaml schema
func validateProduct(p Product) string {
	if p.ProductID < 1 {
		return "product_id must be >= 1"
	}
	if len(p.SKU) == 0 || len(p.SKU) > 100 {
		return "sku must be between 1 and 100 characters"
	}
	if len(p.Manufacturer) == 0 || len(p.Manufacturer) > 200 {
		return "manufacturer must be between 1 and 200 characters"
	}
	if p.CategoryID < 1 {
		return "category_id must be >= 1"
	}
	if p.Weight < 0 {
		return "weight must be >= 0"
	}
	if p.SomeOtherID < 1 {
		return "some_other_id must be >= 1"
	}
	return ""
}
