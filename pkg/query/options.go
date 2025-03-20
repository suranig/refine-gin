package query

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
)

// Filter represents a query filter with operator support
type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, lt, gt, lte, gte, contains, etc.
	Value    interface{} `json:"value"`
}

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

	// Advanced filters (with operators)
	AdvancedFilters []Filter

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

	// Default sort
	if defaultSort := res.GetDefaultSort(); defaultSort != nil {
		opt.Sort = defaultSort.Field
		opt.Order = defaultSort.Order
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
	// Get filterable fields from resource
	for _, field := range res.GetFields() {
		if field.Filterable {
			if value := c.DefaultQuery(field.Name, ""); value != "" {
				opt.Filters[field.Name] = value
			}
		}
	}

	// Initialize advanced filters
	opt.AdvancedFilters = make([]Filter, 0)

	// Parse Refine.dev advanced filters
	// Format 1: filter[field][operator]=value
	for key, values := range c.Request.URL.Query() {
		// Check if it's a filter parameter
		if strings.HasPrefix(key, "filter[") && strings.Contains(key, "][") && strings.HasSuffix(key, "]") {
			// Extract field and operator from the key
			key = strings.TrimPrefix(key, "filter[")
			key = strings.TrimSuffix(key, "]")
			parts := strings.Split(key, "][")

			if len(parts) == 2 && len(values) > 0 {
				field := parts[0]
				operator := parts[1]
				value := values[0]

				// Create advanced filter
				opt.AdvancedFilters = append(opt.AdvancedFilters, Filter{
					Field:    field,
					Operator: operator,
					Value:    value,
				})
			}
		}
	}

	// Format 2: filters[field]=value&operators[field]=operator
	operators := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if strings.HasPrefix(key, "operators[") && strings.HasSuffix(key, "]") && len(values) > 0 {
			field := strings.TrimPrefix(key, "operators[")
			field = strings.TrimSuffix(field, "]")
			operators[field] = values[0]
		}
	}

	for key, values := range c.Request.URL.Query() {
		if strings.HasPrefix(key, "filters[") && strings.HasSuffix(key, "]") && len(values) > 0 {
			field := strings.TrimPrefix(key, "filters[")
			field = strings.TrimSuffix(field, "]")

			// Get the operator, default to "eq" if not specified
			operator := "eq"
			if op, exists := operators[field]; exists {
				operator = op
			}

			// Create advanced filter
			opt.AdvancedFilters = append(opt.AdvancedFilters, Filter{
				Field:    field,
				Operator: operator,
				Value:    values[0],
			})
		}
	}

	// Parse sorting - support both standard (sort, order) and Refine.dev formats
	if sort := c.DefaultQuery("sort", ""); sort != "" {
		// Check if this is a multiple sort fields request (comma-separated)
		if strings.Contains(sort, ",") {
			sortFields := strings.Split(sort, ",")
			sortOrders := strings.Split(c.DefaultQuery("order", "asc"), ",")

			// Ensure we have enough orders for all fields
			for len(sortOrders) < len(sortFields) {
				sortOrders = append(sortOrders, "asc") // Default to asc
			}

			// Build sort expressions
			sortExpressions := make([]string, len(sortFields))
			for i, field := range sortFields {
				field = strings.TrimSpace(field)
				order := strings.TrimSpace(sortOrders[i])
				if order != "asc" && order != "desc" {
					order = "asc" // Default to asc for invalid values
				}
				sortExpressions[i] = field + " " + order
			}

			opt.Sort = strings.Join(sortExpressions, ", ")
		} else {
			opt.Sort = sort
			if order := c.DefaultQuery("order", "asc"); order == "desc" {
				opt.Order = "desc"
			}
		}
	} else if defaultSort := res.GetDefaultSort(); defaultSort != nil {
		opt.Sort = defaultSort.Field
		opt.Order = defaultSort.Order
	}

	return opt
}
