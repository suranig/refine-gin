package query

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
)

// QueryOptions holds query parameters
type QueryOptions struct {
	// Resource to query
	Resource resource.Resource

	// Pagination parameters
	Page    int
	PerPage int

	// Disable pagination for count operations
	DisablePagination bool

	// Search parameters
	Search string

	// Filters
	Filters map[string]interface{}

	// Sorting
	Sort  string
	Order string
}

// NewQueryOptions creates a new QueryOptions from a gin context
func NewQueryOptions(c *gin.Context, res resource.Resource) QueryOptions {
	// Default options
	opt := QueryOptions{
		Resource: res,
		Page:     1,
		PerPage:  10,
		Order:    "asc",
	}

	// Parse pagination - support both standard (page, per_page) and Refine.dev formats (current, pageSize)
	// Try Refine.dev 'current' parameter first
	if current := c.DefaultQuery("current", ""); current != "" {
		// Convert to int
		var currentInt int
		if _, err := fmt.Sscanf(current, "%d", &currentInt); err == nil && currentInt > 0 {
			opt.Page = currentInt
		}
	} else if page := c.DefaultQuery("page", "1"); page != "" {
		// Fall back to standard 'page' parameter
		var pageInt int
		if _, err := fmt.Sscanf(page, "%d", &pageInt); err == nil && pageInt > 0 {
			opt.Page = pageInt
		}
	}

	// Try Refine.dev 'pageSize' parameter first
	if pageSize := c.DefaultQuery("pageSize", ""); pageSize != "" {
		var pageSizeInt int
		if _, err := fmt.Sscanf(pageSize, "%d", &pageSizeInt); err == nil && pageSizeInt > 0 {
			opt.PerPage = pageSizeInt
		}
	} else if perPage := c.DefaultQuery("per_page", "10"); perPage != "" {
		// Fall back to standard 'per_page' parameter
		var perPageInt int
		if _, err := fmt.Sscanf(perPage, "%d", &perPageInt); err == nil && perPageInt > 0 {
			opt.PerPage = perPageInt
		}
	}

	// Parse search
	opt.Search = c.DefaultQuery("q", "")

	// Parse filters
	opt.Filters = make(map[string]interface{})
	for _, filter := range res.GetFilters() {
		if value := c.DefaultQuery(filter.Field, ""); value != "" {
			opt.Filters[filter.Field] = value
		}
	}

	// Parse sorting - support both standard (sort, order) and Refine.dev formats
	if sort := c.DefaultQuery("sort", ""); sort != "" {
		opt.Sort = sort
		if order := c.DefaultQuery("order", "asc"); order == "desc" {
			opt.Order = "desc"
		}
	} else if defaultSort := res.GetDefaultSort(); defaultSort != nil {
		opt.Sort = defaultSort.Field
		opt.Order = defaultSort.Order
	}

	return opt
}
