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

	// Parse pagination
	if page := c.DefaultQuery("page", "1"); page != "" {
		// Convert to int
		var pageInt int
		if _, err := fmt.Sscanf(page, "%d", &pageInt); err == nil && pageInt > 0 {
			opt.Page = pageInt
		}
	}

	if perPage := c.DefaultQuery("per_page", "10"); perPage != "" {
		// Convert to int
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

	// Parse sorting
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
