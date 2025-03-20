package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Default pagination values
const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// PaginateOption represents pagination options
type PaginateOption struct {
	Page    int
	PerPage int
}

// ApplyPaginate applies pagination to a GORM query
func ApplyPaginate(query *gorm.DB, paginate PaginateOption) *gorm.DB {
	offset := (paginate.Page - 1) * paginate.PerPage
	return query.Offset(offset).Limit(paginate.PerPage)
}

// ExtractPaginate extracts pagination options from HTTP query parameters
func ExtractPaginate(c *gin.Context) PaginateOption {
	// Get pagination parameters
	pageStr := c.DefaultQuery("page", strconv.Itoa(DefaultPage))
	perPageStr := c.DefaultQuery("per_page", strconv.Itoa(DefaultPageSize))

	// Handle Refine.dev format (current, pageSize)
	if current := c.DefaultQuery("current", ""); current != "" {
		pageStr = current
	}
	if pageSize := c.DefaultQuery("pageSize", ""); pageSize != "" {
		perPageStr = pageSize
	}

	// Also check Refine.js nested object format
	if c.Query("pagination[current]") != "" {
		pageStr = c.Query("pagination[current]")
	}
	if c.Query("pagination[pageSize]") != "" {
		perPageStr = c.Query("pagination[pageSize]")
	}

	// Convert to int
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = DefaultPage
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = DefaultPageSize
	}

	// Limit maximum page size
	if perPage > MaxPageSize {
		perPage = MaxPageSize
	}

	return PaginateOption{
		Page:    page,
		PerPage: perPage,
	}
}
