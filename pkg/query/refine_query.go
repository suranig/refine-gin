package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// RefineQueryOptions contains options for querying data in Refine format
type RefineQueryOptions struct {
	Resource          resource.Resource
	Page              int
	PerPage           int
	Sort              string
	Order             string
	Search            string
	Filters           []Filter
	DisablePagination bool
}

// ParseRefineQueryOptions parses query options from the request in Refine format
func ParseRefineQueryOptions(c *gin.Context, res resource.Resource) RefineQueryOptions {
	// Parse pagination (Refine uses page and per_page)
	page := 1
	perPage := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	if perPageStr := c.Query("per_page"); perPageStr != "" {
		if perPageInt, err := strconv.Atoi(perPageStr); err == nil && perPageInt > 0 {
			perPage = perPageInt
		}
	}

	// Parse sorting (Refine uses sort and order)
	var sort string
	var order string

	if sortField := c.Query("sort"); sortField != "" {
		sort = sortField
		order = c.DefaultQuery("order", "asc")
	} else if defaultSort := res.GetDefaultSort(); defaultSort != nil {
		sort = defaultSort.Field
		order = defaultSort.Order
	}

	// Parse search (Refine uses q parameter)
	search := c.Query("q")

	// Parse filters in Refine format
	filters := ParseRefineFilters(c)

	return RefineQueryOptions{
		Resource: res,
		Page:     page,
		PerPage:  perPage,
		Sort:     sort,
		Order:    order,
		Search:   search,
		Filters:  filters,
	}
}

// ApplyWithPagination applies query options to a GORM query and returns total count
func (o RefineQueryOptions) ApplyWithPagination(query *gorm.DB, dest interface{}) (int64, error) {
	var total int64
	countQuery := query

	// Apply refine style filters to count query
	if len(o.Filters) > 0 {
		countQuery = ApplyRefineFilters(countQuery, o.Filters)
	}

	// Apply search to count query
	if o.Search != "" {
		countQuery = ApplySearch(countQuery, o.Resource, o.Search)
	}

	// Count total records
	if err := countQuery.Count(&total).Error; err != nil {
		return 0, err
	}

	// Apply refine style filters to main query
	if len(o.Filters) > 0 {
		query = ApplyRefineFilters(query, o.Filters)
	}

	// Apply search to main query
	if o.Search != "" {
		query = ApplySearch(query, o.Resource, o.Search)
	}

	// Apply sorting - convert camelCase to snake_case for database fields
	if o.Sort != "" {
		dbField := ToSnakeCase(o.Sort)
		query = query.Order(dbField + " " + o.Order)
	}

	// Only apply pagination if not disabled and we have a destination
	if dest != nil && !o.DisablePagination {
		// Apply pagination
		offset := (o.Page - 1) * o.PerPage
		query = query.Offset(offset).Limit(o.PerPage)

		// Execute query and scan results
		if err := query.Find(dest).Error; err != nil {
			return 0, err
		}
	}

	return total, nil
}

// NewRefineQueryOptions creates a new RefineQueryOptions from a Gin context and resource
func NewRefineQueryOptions(c *gin.Context, res resource.Resource) RefineQueryOptions {
	return ParseRefineQueryOptions(c, res)
}
