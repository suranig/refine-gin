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

	// Handle Refine.js format
	if field == "" {
		field = c.Query("sort[0][field]")
		order = c.Query("sort[0][order]")
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
