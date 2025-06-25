package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/lang-portal/backend_go/internal/models"
)

// Pagination represents common pagination response
type Pagination struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	TotalItems   int `json:"total_items"`
	ItemsPerPage int `json:"items_per_page"`
}

// PaginatedResponse wraps any response with pagination
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

// Handler wraps database connection for handlers
type Handler struct {
	db *models.DB
}

// NewHandler creates a new handler with database connection
func NewHandler(db *models.DB) *Handler {
	return &Handler{db: db}
}

// calculatePagination helps calculate pagination values
func calculatePagination(page, itemsPerPage, totalItems int) Pagination {
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage
	if totalPages == 0 {
		totalPages = 1
	}
	
	return Pagination{
		CurrentPage:  page,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		ItemsPerPage: itemsPerPage,
	}
}

// respondWithError sends a JSON error response
func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"error": message,
	})
} 