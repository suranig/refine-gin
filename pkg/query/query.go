package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// QueryOptions contains options for querying data
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

// ParseQueryOptions parses query options from the request
func ParseQueryOptions(c *gin.Context, res resource.Resource) QueryOptions {
	// Parse pagination - support both standard and Refine.dev formats
	page := 1
	perPage := 10

	// Try Refine.dev 'current' parameter first
	if currentStr := c.Query("current"); currentStr != "" {
		if currentInt, err := strconv.Atoi(currentStr); err == nil && currentInt > 0 {
			page = currentInt
		}
	} else if pageStr := c.Query("page"); pageStr != "" {
		// Fall back to standard 'page' parameter
		if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	// Try Refine.dev 'pageSize' parameter first
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if pageSizeInt, err := strconv.Atoi(pageSizeStr); err == nil && pageSizeInt > 0 {
			perPage = pageSizeInt
		}
	} else if perPageStr := c.Query("per_page"); perPageStr != "" {
		// Fall back to standard 'per_page' parameter
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

