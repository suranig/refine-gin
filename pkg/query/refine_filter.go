package query

import (
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RefineFilterOperator represents a filtering operator used in Refine.dev
type RefineFilterOperator string

// Available Refine filter operators
const (
	// RefineEQ represents the "equal" operator
	RefineEQ RefineFilterOperator = "eq"

	// RefineNE represents the "not equal" operator
	RefineNE RefineFilterOperator = "ne"

	// RefineGT represents the "greater than" operator
	RefineGT RefineFilterOperator = "gt"

	// RefineLT represents the "less than" operator
	RefineLT RefineFilterOperator = "lt"

	// RefineGTE represents the "greater than or equal" operator
	RefineGTE RefineFilterOperator = "gte"

	// RefineLTE represents the "less than or equal" operator
	RefineLTE RefineFilterOperator = "lte"

	// RefineIN represents the "in" operator
	RefineIN RefineFilterOperator = "in"

	// RefineNIN represents the "not in" operator
	RefineNIN RefineFilterOperator = "nin"

	// RefineCONTAINS represents the "contains" operator
	RefineCONTAINS RefineFilterOperator = "contains"

	// RefineNCONTAINS represents the "not contains" operator
	RefineNCONTAINS RefineFilterOperator = "ncontains"

	// RefineCONTAINSS represents the "contains (case sensitive)" operator
	RefineCONTAINSS RefineFilterOperator = "containss"

	// RefineNCONTAINSS represents the "not contains (case sensitive)" operator
	RefineNCONTAINSS RefineFilterOperator = "ncontainss"

	// RefineNULL represents the "is null" operator
	RefineNULL RefineFilterOperator = "null"

	// RefineNNULL represents the "is not null" operator
	RefineNNULL RefineFilterOperator = "nnull"

	// RefineBETWEEN represents the "between" operator
	RefineBETWEEN RefineFilterOperator = "between"

	// RefineNBETWEEN represents the "not between" operator
	RefineNBETWEEN RefineFilterOperator = "nbetween"

	// RefineSTARTS_WITH represents the "starts with" operator
	RefineSTARTS_WITH RefineFilterOperator = "startswith"

	// RefineNSTARTS_WITH represents the "not starts with" operator
	RefineNSTARTS_WITH RefineFilterOperator = "nstartswith"

	// RefineSTARTS_WITH_S represents the "starts with (case sensitive)" operator
	RefineSTARTS_WITH_S RefineFilterOperator = "startswith_s"

	// RefineNSTARTS_WITH_S represents the "not starts with (case sensitive)" operator
	RefineNSTARTS_WITH_S RefineFilterOperator = "nstartswith_s"

	// RefineENDS_WITH represents the "ends with" operator
	RefineENDS_WITH RefineFilterOperator = "endswith"

	// RefineNENDS_WITH represents the "not ends with" operator
	RefineNENDS_WITH RefineFilterOperator = "nendswith"

	// RefineENDS_WITH_S represents the "ends with (case sensitive)" operator
	RefineENDS_WITH_S RefineFilterOperator = "endswith_s"

	// RefineNENDS_WITH_S represents the "not ends with (case sensitive)" operator
	RefineNENDS_WITH_S RefineFilterOperator = "nendswith_s"

	// RefineOR represents the "or" operator
	RefineOR RefineFilterOperator = "or"
)

// ParseRefineFilters extracts filters from query parameters in Refine format
func ParseRefineFilters(c *gin.Context) []Filter {
	filters := []Filter{}
	query := c.Request.URL.Query()

	// Handle regular filters
	for key, values := range query {
		if len(values) == 0 || key == "sort" || key == "order" || key == "page" || key == "per_page" || key == "q" {
			continue
		}

		// Check if it's an operator key (ends with _operator)
		if strings.HasSuffix(key, "_operator") {
			continue
		}

		field := key
		operator := "eq" // Default operator is "eq"
		value := values[0]

		// Check if there's a specific operator for this field
		operatorKey := key + "_operator"
		if opValues, exists := query[operatorKey]; exists && len(opValues) > 0 {
			operator = opValues[0]
		}

		filters = append(filters, Filter{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
	}

	return filters
}

// ApplyRefineFilters applies all refine filters to a query
func ApplyRefineFilters(db *gorm.DB, filters []Filter) *gorm.DB {
	for _, filter := range filters {
		db = ApplyRefineFilter(db, filter)
	}
	return db
}

// ApplyRefineFilter applies refine-specific filter logic to a GORM query
func ApplyRefineFilter(db *gorm.DB, filter Filter) *gorm.DB {
	field := filter.Field
	value := filter.Value

	// Convert field name to snake_case for database query
	field = ToSnakeCase(field)

	switch RefineFilterOperator(filter.Operator) {
	case RefineEQ:
		return db.Where(field+" = ?", value)
	case RefineNE:
		return db.Where(field+" <> ?", value)
	case RefineGT:
		return db.Where(field+" > ?", value)
	case RefineLT:
		return db.Where(field+" < ?", value)
	case RefineGTE:
		return db.Where(field+" >= ?", value)
	case RefineLTE:
		return db.Where(field+" <= ?", value)
	case RefineIN:
		values := strings.Split(value.(string), ",")
		return db.Where(field+" IN ?", values)
	case RefineNIN:
		values := strings.Split(value.(string), ",")
		return db.Where(field+" NOT IN ?", values)
	case RefineCONTAINS:
		return db.Where(field+" ILIKE ?", "%"+value.(string)+"%")
	case RefineNCONTAINS:
		return db.Where(field+" NOT ILIKE ?", "%"+value.(string)+"%")
	case RefineCONTAINSS:
		return db.Where(field+" LIKE ?", "%"+value.(string)+"%")
	case RefineNCONTAINSS:
		return db.Where(field+" NOT LIKE ?", "%"+value.(string)+"%")
	case RefineNULL:
		return db.Where(field + " IS NULL")
	case RefineNNULL:
		return db.Where(field + " IS NOT NULL")
	case RefineSTARTS_WITH:
		return db.Where(field+" ILIKE ?", value.(string)+"%")
	case RefineNSTARTS_WITH:
		return db.Where(field+" NOT ILIKE ?", value.(string)+"%")
	case RefineSTARTS_WITH_S:
		return db.Where(field+" LIKE ?", value.(string)+"%")
	case RefineNSTARTS_WITH_S:
		return db.Where(field+" NOT LIKE ?", value.(string)+"%")
	case RefineENDS_WITH:
		return db.Where(field+" ILIKE ?", "%"+value.(string))
	case RefineNENDS_WITH:
		return db.Where(field+" NOT ILIKE ?", "%"+value.(string))
	case RefineENDS_WITH_S:
		return db.Where(field+" LIKE ?", "%"+value.(string))
	case RefineNENDS_WITH_S:
		return db.Where(field+" NOT LIKE ?", "%"+value.(string))
	default:
		return db.Where(field+" = ?", value)
	}
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ParseRefineLogicalOperators parses logical operators (OR) in Refine format
func ParseRefineLogicalOperators(c *gin.Context) []map[string]string {
	result := []map[string]string{}
	query := c.Request.URL.Query()

	// Check for _or parameters
	for key, values := range query {
		if !strings.HasPrefix(key, "_or") || len(values) == 0 {
			continue
		}

		// Parse the _or parameter index from the key
		// Example: _or.0.field=name&_or.0.operator=contains&_or.0.value=John
		parts := strings.Split(key, ".")
		if len(parts) != 3 {
			continue
		}

		index := parts[1]
		field := parts[2]

		// Check if we already have a map for this index
		found := false
		for i, m := range result {
			if _, ok := m["index"]; ok && m["index"] == index {
				result[i][field] = values[0]
				found = true
				break
			}
		}

		if !found {
			newMap := map[string]string{
				"index": index,
				field:   values[0],
			}
			result = append(result, newMap)
		}
	}

	return result
}

// ParseRefineRangeFilters parses range filters in Refine format
func ParseRefineRangeFilters(c *gin.Context) []Filter {
	filters := []Filter{}
	query := c.Request.URL.Query()

	// Check for range parameters
	for key, values := range query {
		// Range filters are expected in the format field_lte or field_gte
		if (!strings.HasSuffix(key, "_lte") && !strings.HasSuffix(key, "_gte")) || len(values) == 0 {
			continue
		}

		var field, operator string
		if strings.HasSuffix(key, "_lte") {
			field = strings.TrimSuffix(key, "_lte")
			operator = "lte"
		} else if strings.HasSuffix(key, "_gte") {
			field = strings.TrimSuffix(key, "_gte")
			operator = "gte"
		}

		filters = append(filters, Filter{
			Field:    field,
			Operator: operator,
			Value:    values[0],
		})
	}

	return filters
}
