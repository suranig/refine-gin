package query

import (
	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// QueryOptions zawiera opcje zapytania
type QueryOptions struct {
	Filters  []FilterOption
	Sort     SortOption
	Paginate PaginateOption
	Resource resource.Resource
}

// NewQueryOptions tworzy nowe opcje zapytania na podstawie kontekstu Gin
func NewQueryOptions(c *gin.Context, res resource.Resource) QueryOptions {
	// Pobierz filtry
	var filters []FilterOption

	// Pobierz pola, które można filtrować
	var filterFields []string
	for _, field := range res.GetFields() {
		if field.Filterable {
			filterFields = append(filterFields, field.Name)
		}
	}

	// Pobierz domyślne pole wyszukiwania
	var defaultSearchField string
	for _, field := range res.GetFields() {
		if field.Searchable {
			defaultSearchField = field.Name
			break
		}
	}

	// Utwórz konfigurację filtrów
	filterConfig := ResourceFilterConfig{
		Fields:       filterFields,
		Operators:    map[string]string{}, // Można dodać mapowanie operatorów
		DefaultField: defaultSearchField,
	}

	// Ekstrahuj filtry z parametrów zapytania
	filters = ExtractFilters(c, filterConfig)

	// Pobierz sortowanie
	var defaultSort *SortOption
	if res.GetDefaultSort() != nil {
		defaultSort = &SortOption{
			Field: res.GetDefaultSort().Field,
			Order: res.GetDefaultSort().Order,
		}
	}
	sort := ExtractSort(c, defaultSort)

	// Pobierz paginację
	paginate := ExtractPaginate(c)

	return QueryOptions{
		Filters:  filters,
		Sort:     sort,
		Paginate: paginate,
		Resource: res,
	}
}

// Apply stosuje opcje zapytania do zapytania GORM
func (o QueryOptions) Apply(query *gorm.DB) *gorm.DB {
	// Zastosuj filtry
	query = ApplyFilters(query, o.Filters)

	// Zastosuj sortowanie
	query = ApplySort(query, o.Sort)

	// Zastosuj paginację
	query = ApplyPaginate(query, o.Paginate)

	return query
}

// ApplyWithPagination stosuje opcje zapytania do zapytania GORM i zwraca całkowitą liczbę rekordów
func (o QueryOptions) ApplyWithPagination(query *gorm.DB, dest interface{}) (int64, error) {
	var total int64

	// Klonuj zapytanie do liczenia
	countQuery := query

	// Zastosuj filtry do obu zapytań
	query = ApplyFilters(query, o.Filters)
	countQuery = ApplyFilters(countQuery, o.Filters)

	// Policz całkowitą liczbę rekordów
	if err := countQuery.Count(&total).Error; err != nil {
		return 0, err
	}

	// Zastosuj sortowanie
	query = ApplySort(query, o.Sort)

	// Zastosuj paginację
	query = ApplyPaginate(query, o.Paginate)

	// Wykonaj zapytanie
	if err := query.Find(dest).Error; err != nil {
		return 0, err
	}

	return total, nil
}
