package query

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SortOrder defines sort direction
type SortOrder string

const (
	// Ascending order
	SortOrderAsc SortOrder = "asc"

	// Descending order
	SortOrderDesc SortOrder = "desc"
)

// SortOption represents a sorting option
type SortOption struct {
	Field string
	Order string // "asc" or "desc"
}

// ApplySort applies sorting to a GORM query
func ApplySort(query *gorm.DB, sort SortOption) *gorm.DB {
	if sort.Field != "" {
		orderStr := sort.Field
		if sort.Order == string(SortOrderDesc) {
			orderStr += " DESC"
		} else {
			orderStr += " ASC"
		}
		query = query.Order(orderStr)
	}
	return query
}

// ExtractSort extracts sorting options from HTTP query parameters
func ExtractSort(c *gin.Context, defaultSort *SortOption) SortOption {
	field := c.Query("sort")
	order := c.Query("order")

	// Handle Refine.js format with array notation
	if field == "" {
		// Check for Refine.dev format: sort[0][field]=name&sort[0][order]=asc
		indexField := c.Query("sort[0][field]")
		indexOrder := c.Query("sort[0][order]")

		if indexField != "" {
			field = indexField
			if indexOrder != "" {
				order = indexOrder
			}
		}
	}

	// Handle Refine.js format with object notation
	if field == "" && c.Query("sort[field]") != "" {
		field = c.Query("sort[field]")
		if c.Query("sort[order]") != "" {
			order = c.Query("sort[order]")
		}
	}

	// If no sorting provided, use default
	if field == "" && defaultSort != nil {
		return *defaultSort
	}

	// Validate order
	if order != string(SortOrderAsc) && order != string(SortOrderDesc) {
		order = string(SortOrderAsc) // default to ascending
	}

	return SortOption{
		Field: field,
		Order: order,
	}
}
