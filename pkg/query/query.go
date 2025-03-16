package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// QueryOptions contains options for querying data
type QueryOptions struct {
	Resource resource.Resource
	Page     int
	PerPage  int
	Sort     *resource.Sort
	Filters  []Filter
	Search   string
}

// Filter represents a filter condition
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
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
	sort := res.GetDefaultSort()
	if sortField := c.Query("sort"); sortField != "" {
		sortOrder := c.DefaultQuery("order", "asc")
		sort = &resource.Sort{
			Field: sortField,
			Order: sortOrder,
		}
	}

	// Parse filters
	var filters []Filter
	for _, field := range res.GetFields() {
		if !field.Filterable {
			continue
		}

		value := c.Query(field.Name)
		if value == "" {
			continue
		}

		operator := c.DefaultQuery(field.Name+"_operator", "eq")
		filters = append(filters, Filter{
			Field:    field.Name,
			Operator: operator,
			Value:    value,
		})
	}

	// Parse search
	search := c.Query("q")

	return QueryOptions{
		Resource: res,
		Page:     page,
		PerPage:  perPage,
		Sort:     sort,
		Filters:  filters,
		Search:   search,
	}
}

// Apply stosuje opcje zapytania do zapytania GORM
func (o QueryOptions) Apply(query *gorm.DB) *gorm.DB {
	query = ApplyFilters(query, o.Filters)

	if o.Search != "" {
		query = ApplySearch(query, o.Resource, o.Search)
	}

	if o.Sort != nil {
		sortOption := SortOption{
			Field: o.Sort.Field,
			Order: o.Sort.Order,
		}
		query = ApplySort(query, sortOption)
	}

	paginateOption := PaginateOption{
		Page:    o.Page,
		PerPage: o.PerPage,
	}
	query = ApplyPaginate(query, paginateOption)

	return query
}

// ApplyWithPagination stosuje opcje zapytania do zapytania GORM i zwraca całkowitą liczbę rekordów
func (o QueryOptions) ApplyWithPagination(query *gorm.DB, dest interface{}) (int64, error) {
	var total int64

	countQuery := query

	countQuery = ApplyFilters(countQuery, o.Filters)

	if o.Search != "" {
		countQuery = ApplySearch(countQuery, o.Resource, o.Search)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return 0, err
	}

	query = ApplyFilters(query, o.Filters)

	if o.Search != "" {
		query = ApplySearch(query, o.Resource, o.Search)
	}

	if o.Sort != nil {
		sortOption := SortOption{
			Field: o.Sort.Field,
			Order: o.Sort.Order,
		}
		query = ApplySort(query, sortOption)
	}

	paginateOption := PaginateOption{
		Page:    o.Page,
		PerPage: o.PerPage,
	}
	query = ApplyPaginate(query, paginateOption)

	if err := query.Find(dest).Error; err != nil {
		return 0, err
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
