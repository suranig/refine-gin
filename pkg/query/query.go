package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// QueryOptions contains options for querying data
type QueryOptions struct {
	Resource          resource.Resource
	Page              int
	PerPage           int
	Sort              string
	Order             string
	Filters           map[string]interface{}
	Search            string
	DisablePagination bool
}

// Filter represents a filter condition
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// SortOption represents a sort option
type SortOption struct {
	Field string
	Order string
}

// PaginateOption represents pagination options
type PaginateOption struct {
	Page    int
	PerPage int
}

// ParseQueryOptions parses query options from the request
func ParseQueryOptions(c *gin.Context, res resource.Resource) QueryOptions {
	// Parse pagination
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

	// Parse sorting
	var sort string
	var order string

	if sortField := c.Query("sort"); sortField != "" {
		sort = sortField
		order = c.DefaultQuery("order", "asc")
	} else if defaultSort := res.GetDefaultSort(); defaultSort != nil {
		sort = defaultSort.Field
		order = defaultSort.Order
	}

	// Parse filters
	filters := make(map[string]interface{})
	for _, field := range res.GetFields() {
		if !field.Filterable {
			continue
		}

		value := c.Query(field.Name)
		if value == "" {
			continue
		}

		filters[field.Name] = value
	}

	// Parse search
	search := c.Query("q")

	return QueryOptions{
		Resource: res,
		Page:     page,
		PerPage:  perPage,
		Sort:     sort,
		Order:    order,
		Filters:  filters,
		Search:   search,
	}
}

// Apply stosuje opcje zapytania do zapytania GORM
func (o QueryOptions) Apply(query *gorm.DB) *gorm.DB {
	// Apply filters
	for field, value := range o.Filters {
		query = query.Where(field+" = ?", value)
	}

	// Apply search
	if o.Search != "" {
		query = ApplySearch(query, o.Resource, o.Search)
	}

	// Apply sorting
	if o.Sort != "" {
		query = query.Order(o.Sort + " " + o.Order)
	}

	// Apply pagination
	offset := (o.Page - 1) * o.PerPage
	query = query.Offset(offset).Limit(o.PerPage)

	return query
}

// ApplyWithPagination stosuje opcje zapytania do zapytania GORM i zwraca całkowitą liczbę rekordów
func (o QueryOptions) ApplyWithPagination(query *gorm.DB, dest interface{}) (int64, error) {
	var total int64

	countQuery := query

	// Apply filters to count query
	for field, value := range o.Filters {
		countQuery = countQuery.Where(field+" = ?", value)
	}

	// Apply search to count query
	if o.Search != "" {
		countQuery = ApplySearch(countQuery, o.Resource, o.Search)
	}

	// Count total records
	if err := countQuery.Count(&total).Error; err != nil {
		return 0, err
	}

	// Apply filters to main query
	for field, value := range o.Filters {
		query = query.Where(field+" = ?", value)
	}

	// Apply search to main query
	if o.Search != "" {
		query = ApplySearch(query, o.Resource, o.Search)
	}

	// Apply sorting
	if o.Sort != "" {
		query = query.Order(o.Sort + " " + o.Order)
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

func ApplySearch(query *gorm.DB, res resource.Resource, search string) *gorm.DB {
	if search == "" {
		return query
	}

	var searchableFields []string
	for _, field := range res.GetFields() {
		if field.Searchable {
			searchableFields = append(searchableFields, field.Name)
		}
	}

	if len(searchableFields) == 0 {
		return query
	}

	searchValue := "%" + search + "%"
	for i, field := range searchableFields {
		if i == 0 {
			query = query.Where(field+" LIKE ?", searchValue)
		} else {
			query = query.Or(field+" LIKE ?", searchValue)
		}
	}

	return query
}

// NewQueryOptions creates a new QueryOptions from a Gin context and resource
func NewQueryOptions(c *gin.Context, res resource.Resource) QueryOptions {
	return ParseQueryOptions(c, res)
}
