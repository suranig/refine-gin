package query

import (
	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// QueryOptions zawiera wszystkie opcje zapytania
type QueryOptions struct {
	Filters  []FilterOption
	Sort     SortOption
	Paginate PaginateOption
}

// NewQueryOptions tworzy nowe opcje zapytania z kontekstu Gin i zasobu
func NewQueryOptions(c *gin.Context, res resource.Resource) QueryOptions {
	// Przygotowanie konfiguracji filtrów
	var filterFields []string
	for _, field := range res.GetFields() {
		if field.Filterable {
			filterFields = append(filterFields, field.Name)
		}
	}

	// Znajdź domyślne pole wyszukiwania
	defaultField := ""
	for _, field := range res.GetFields() {
		if field.Searchable {
			defaultField = field.Name
			break
		}
	}

	// Przygotowanie konfiguracji filtrów
	filterConfig := ResourceFilterConfig{
		Fields: filterFields,
		Operators: map[string]string{
			"eq":         "=",
			"neq":        "!=",
			"gt":         ">",
			"gte":        ">=",
			"lt":         "<",
			"lte":        "<=",
			"like":       "LIKE",
			"contains":   "LIKE",
			"startswith": "LIKE",
			"endswith":   "LIKE",
			"null":       "IS NULL",
		},
		DefaultField: defaultField,
	}

	// Przygotowanie domyślnego sortowania
	var defaultSort *SortOption
	if res.GetDefaultSort() != nil {
		defaultSort = &SortOption{
			Field: res.GetDefaultSort().Field,
			Order: res.GetDefaultSort().Order,
		}
	}

	return QueryOptions{
		Filters:  ExtractFilters(c, filterConfig),
		Sort:     ExtractSort(c, defaultSort),
		Paginate: ExtractPaginate(c),
	}
}

// Apply stosuje wszystkie opcje zapytania do zapytania GORM
func (qo QueryOptions) Apply(query *gorm.DB) *gorm.DB {
	query = ApplyFilters(query, qo.Filters)
	query = ApplySort(query, qo.Sort)
	return query
}

// ApplyWithPagination stosuje opcje zapytania wraz z paginacją i zwraca wyniki
func (qo QueryOptions) ApplyWithPagination(query *gorm.DB, result interface{}) (int64, error) {
	var total int64

	// Zastosuj filtry i sortowanie
	query = ApplyFilters(query, qo.Filters)
	query = ApplySort(query, qo.Sort)

	// Pobierz całkowitą liczbę wyników
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}

	// Zastosuj paginację i pobierz wyniki
	query = ApplyPaginate(query, qo.Paginate)
	if err := query.Find(result).Error; err != nil {
		return 0, err
	}

	return total, nil
}
