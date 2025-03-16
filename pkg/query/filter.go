package query

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FilterOption represents a filtering option
type FilterOption struct {
	Field    string
	Operator string
	Value    interface{}
}

// FilterOperator defines available filter operators
type FilterOperator string

const (
	// Equal operator (=)
	OperatorEqual FilterOperator = "eq"

	// Not equal operator (!=)
	OperatorNotEqual FilterOperator = "neq"

	// Greater than operator (>)
	OperatorGreaterThan FilterOperator = "gt"

	// Greater than or equal operator (>=)
	OperatorGreaterThanEqual FilterOperator = "gte"

	// Less than operator (<)
	OperatorLessThan FilterOperator = "lt"

	// Less than or equal operator (<=)
	OperatorLessThanEqual FilterOperator = "lte"

	// Like operator (LIKE %value%)
	OperatorLike FilterOperator = "like"

	// In operator (IN)
	OperatorIn FilterOperator = "in"

	// Contains operator (LIKE %value%)
	OperatorContains FilterOperator = "contains"

	// Starts with operator (LIKE value%)
	OperatorStartsWith FilterOperator = "startswith"

	// Ends with operator (LIKE %value)
	OperatorEndsWith FilterOperator = "endswith"

	// Is null operator (IS NULL)
	OperatorIsNull FilterOperator = "null"
)

// ResourceFilterConfig defines filter configuration for a resource
type ResourceFilterConfig struct {
	Fields       []string          // Allowed fields for filtering
	Operators    map[string]string // Mapping of operators to SQL
	DefaultField string            // Default field for search (q parameter)
}

// ApplyFilters applies filters to a GORM query
func ApplyFilters(query *gorm.DB, filters []FilterOption) *gorm.DB {
	for _, filter := range filters {
		switch filter.Operator {
		case string(OperatorEqual), "":
			query = query.Where(filter.Field+" = ?", filter.Value)
		case string(OperatorNotEqual):
			query = query.Where(filter.Field+" != ?", filter.Value)
		case string(OperatorGreaterThan):
			query = query.Where(filter.Field+" > ?", filter.Value)
		case string(OperatorGreaterThanEqual):
			query = query.Where(filter.Field+" >= ?", filter.Value)
		case string(OperatorLessThan):
			query = query.Where(filter.Field+" < ?", filter.Value)
		case string(OperatorLessThanEqual):
			query = query.Where(filter.Field+" <= ?", filter.Value)
		case string(OperatorLike), string(OperatorContains):
			query = query.Where(filter.Field+" LIKE ?", "%"+filter.Value.(string)+"%")
		case string(OperatorIn):
			query = query.Where(filter.Field+" IN ?", filter.Value)
		case string(OperatorStartsWith):
			query = query.Where(filter.Field+" LIKE ?", filter.Value.(string)+"%")
		case string(OperatorEndsWith):
			query = query.Where(filter.Field+" LIKE ?", "%"+filter.Value.(string))
		case string(OperatorIsNull):
			if filter.Value.(bool) {
				query = query.Where(filter.Field + " IS NULL")
			} else {
				query = query.Where(filter.Field + " IS NOT NULL")
			}
		}
	}
	return query
}

// ExtractFilters extracts filters from HTTP query parameters
func ExtractFilters(c *gin.Context, config ResourceFilterConfig) []FilterOption {
	var filters []FilterOption

	// Check each field if it's in query params
	for _, field := range config.Fields {
		if value := c.Query(field); value != "" {
			// Check if operator is provided
			operator := c.Query(field + "_operator")
			if operator == "" {
				operator = string(OperatorEqual) // default operator
			}

			// Check if operator is allowed
			if _, ok := config.Operators[operator]; !ok {
				operator = string(OperatorEqual) // if not, use default
			}

			filters = append(filters, FilterOption{
				Field:    field,
				Operator: operator,
				Value:    value,
			})
		}
	}

	// Handle q parameter (search)
	if q := c.Query("q"); q != "" && config.DefaultField != "" {
		filters = append(filters, FilterOption{
			Field:    config.DefaultField,
			Operator: string(OperatorLike),
			Value:    q,
		})
	}

	return filters
}
