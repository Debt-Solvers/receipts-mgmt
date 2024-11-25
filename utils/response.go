package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// APIResponse is a struct for a standardized API response
type APIResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Pagination represents pagination information
type Pagination struct {
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

// SendResponse creates and sends a standardized JSON response
func SendResponse(context *gin.Context, status int, message string, data interface{}, errors interface{}) {
	context.JSON(status, APIResponse{
		Status:  status,
		Message: message,
		Data:    data,
		Errors:  errors,
	})
}

// CalculatePagination returns pagination details based on the total count, page, and limit
func CalculatePagination(totalCount, page, limit int) Pagination {
	// Prevent zero or negative values for limit
	if limit <= 0 {
		limit = 10 // Default limit if not specified
	}
	// Prevent zero or negative values for page
	if page <= 0 {
		page = 1 // Default page if not specified
	}

	// Calculate total pages
	totalPages := (totalCount + limit - 1) / limit

	// Create and return the pagination struct
	return Pagination{
		TotalCount: totalCount,
		Page:       page,
		PerPage:    limit,
		TotalPages: totalPages,
	}
}

// ParseQueryInt parses an integer query parameter from the context and returns a default value if parsing fails.
func ParseQueryInt(c *gin.Context, key string, defaultValue int) int {
	// Retrieve the query parameter as a string
	param := c.Query(key)
	if param == "" {
		// If the parameter is empty, return the default value
		return defaultValue
	}
	
	// Attempt to convert the parameter to an integer
	if value, err := strconv.Atoi(param); err == nil {
		return value
	}
	
	// If parsing fails, return the default value
	return defaultValue
}
