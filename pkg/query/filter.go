package query

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Filter represents a filtering condition
type Filter struct {
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
	OperatorNotEqual FilterOperator = "ne"

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

	// Not in operator (NOT IN)
	OperatorNotIn FilterOperator = "nin"

	// Contains operator (LIKE %value%)
	OperatorContains FilterOperator = "contains"

	// Not contains operator (NOT LIKE %value%)
	OperatorNotContains FilterOperator = "ncontains"

	// Contains case-sensitive (LIKE BINARY %value%)
	OperatorContainsSensitive FilterOperator = "containss"

	// Not contains case-sensitive (NOT LIKE BINARY %value%)
	OperatorNotContainsSensitive FilterOperator = "ncontainss"

	// Starts with operator (LIKE value%)
	OperatorStartsWith FilterOperator = "startswith"

	// Not starts with operator (NOT LIKE value%)
	OperatorNotStartsWith FilterOperator = "nstartswith"

	// Ends with operator (LIKE %value)
	OperatorEndsWith FilterOperator = "endswith"

	// Not ends with operator (NOT LIKE %value)
	OperatorNotEndsWith FilterOperator = "nendswith"

	// Is null operator (IS NULL)
	OperatorIsNull FilterOperator = "isnull"

	// Is null operator for Refine (IS NULL)
	OperatorNull FilterOperator = "null"

	// Is not null operator (IS NOT NULL)
	OperatorNotNull FilterOperator = "nnull"

	// Between operator (BETWEEN)
	OperatorBetween FilterOperator = "between"

	// Not between operator (NOT BETWEEN)
	OperatorNotBetween FilterOperator = "nbetween"
)

// ResourceFilterConfig defines filter configuration for a resource
type ResourceFilterConfig struct {
	Fields       []string          // Allowed fields for filtering
	Operators    map[string]string // Mapping of operators to SQL
	DefaultField string            // Default field for search (q parameter)
}

// ApplyFilters applies filters to a GORM query
func ApplyFilters(query *gorm.DB, filters []Filter) *gorm.DB {
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
		case string(OperatorNotContains):
			query = query.Where(filter.Field+" NOT LIKE ?", "%"+filter.Value.(string)+"%")
		case string(OperatorContainsSensitive):
			query = query.Where(filter.Field+" LIKE BINARY ?", "%"+filter.Value.(string)+"%")
		case string(OperatorNotContainsSensitive):
			query = query.Where(filter.Field+" NOT LIKE BINARY ?", "%"+filter.Value.(string)+"%")
		case string(OperatorIn):
			query = query.Where(filter.Field+" IN ?", filter.Value)
		case string(OperatorNotIn):
			query = query.Where(filter.Field+" NOT IN ?", filter.Value)
		case string(OperatorStartsWith):
			query = query.Where(filter.Field+" LIKE ?", filter.Value.(string)+"%")
		case string(OperatorNotStartsWith):
			query = query.Where(filter.Field+" NOT LIKE ?", filter.Value.(string)+"%")
		case string(OperatorEndsWith):
			query = query.Where(filter.Field+" LIKE ?", "%"+filter.Value.(string))
		case string(OperatorNotEndsWith):
			query = query.Where(filter.Field+" NOT LIKE ?", "%"+filter.Value.(string))
		case string(OperatorIsNull), string(OperatorNull):
			query = query.Where(filter.Field + " IS NULL")
		case string(OperatorNotNull):
			query = query.Where(filter.Field + " IS NOT NULL")
		case string(OperatorBetween):
			if strVal, ok := filter.Value.(string); ok {
				parts := strings.Split(strVal, ",")
				if len(parts) == 2 {
					query = query.Where(filter.Field+" BETWEEN ? AND ?", parts[0], parts[1])
				}
			}
		case string(OperatorNotBetween):
			if strVal, ok := filter.Value.(string); ok {
				parts := strings.Split(strVal, ",")
				if len(parts) == 2 {
					query = query.Where(filter.Field+" NOT BETWEEN ? AND ?", parts[0], parts[1])
				}
			}
		}
	}
	return query
}

// ExtractFilters extracts filters from HTTP query parameters
func ExtractFilters(c *gin.Context, config ResourceFilterConfig) []Filter {
	var filters []Filter
	queryParams := c.Request.URL.Query()

	// Check each field if it's in query params
	for _, field := range config.Fields {
		// Standard filtering (field=value)
		if values, exists := queryParams[field]; exists && len(values) > 0 {
			// Default operator
			operator := string(OperatorEqual)

			filters = append(filters, Filter{
				Field:    field,
				Operator: operator,
				Value:    values[0],
			})
		}

		// Process Refine style fields (field_operator=value)
		for opStr := range config.Operators {
			paramName := field + "_" + opStr
			if values, exists := queryParams[paramName]; exists && len(values) > 0 {
				filters = append(filters, Filter{
					Field:    field,
					Operator: opStr,
					Value:    values[0],
				})
			}
		}
	}

	// Handle q parameter (search)
	if values, exists := queryParams["q"]; exists && len(values) > 0 && config.DefaultField != "" {
		filters = append(filters, Filter{
			Field:    config.DefaultField,
			Operator: string(OperatorLike),
			Value:    values[0],
		})
	}

	return filters
}

// ParseRefineFilters extracts filters from query parameters in Refine.dev format
func ParseRefineFilters(c *gin.Context, config ResourceFilterConfig) []Filter {
	var filters []Filter
	queryParams := c.Request.URL.Query()

	// Process all query parameters
	for key, values := range queryParams {
		if len(values) == 0 || key == "sort" || key == "order" || key == "current" || key == "pageSize" || key == "q" {
			continue
		}

		// Check if it's an operator key (ends with _operator)
		if strings.HasSuffix(key, "_operator") {
			continue
		}

		// Check Refine.dev format "filters[0][field]=name&filters[0][operator]=eq&filters[0][value]=John"
		if strings.HasPrefix(key, "filters[") && strings.Contains(key, "][") {
			parts := strings.Split(key, "][")
			if len(parts) >= 2 {
				// Extract index and field
				indexPart := strings.TrimPrefix(parts[0], "filters[")
				fieldPart := parts[1]

				// Check if this is a field, operator, or value
				if strings.HasSuffix(fieldPart, "field]") {
					field := values[0]

					// Find corresponding operator and value with the same index
					operatorKey := fmt.Sprintf("filters[%s][operator]", indexPart)
					valueKey := fmt.Sprintf("filters[%s][value]", indexPart)

					operator := queryParams.Get(operatorKey)
					value := queryParams.Get(valueKey)

					if field != "" && value != "" {
						// Check if field is allowed
						fieldAllowed := false
						for _, allowedField := range config.Fields {
							if field == allowedField {
								fieldAllowed = true
								break
							}
						}

						if fieldAllowed {
							// Use default operator if not specified
							if operator == "" {
								operator = string(OperatorEqual)
							}

							filters = append(filters, Filter{
								Field:    field,
								Operator: operator,
								Value:    value,
							})
						}
					}
				}
			}
			continue
		}

		// Check if it's a field with operator (field_operator=value)
		if strings.Contains(key, "_") {
			parts := strings.SplitN(key, "_", 2)
			if len(parts) == 2 {
				field := parts[0]
				operator := parts[1]

				// Check if field is allowed
				fieldAllowed := false
				for _, allowedField := range config.Fields {
					if field == allowedField {
						fieldAllowed = true
						break
					}
				}

				if fieldAllowed && len(values) > 0 {
					// Check if operator is allowed
					if _, ok := config.Operators[operator]; ok {
						filters = append(filters, Filter{
							Field:    field,
							Operator: operator,
							Value:    values[0],
						})
					}
				}
			}
			continue
		}

		// Standard field=value filtering
		field := key
		// Check if field is allowed
		fieldAllowed := false
		for _, allowedField := range config.Fields {
			if field == allowedField {
				fieldAllowed = true
				break
			}
		}

		if fieldAllowed && len(values) > 0 {
			filters = append(filters, Filter{
				Field:    field,
				Operator: string(OperatorEqual),
				Value:    values[0],
			})
		}
	}

	return filters
}
