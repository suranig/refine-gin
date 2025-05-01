package query

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// Apply applies query options (filters, search, sorting, but not pagination) to a GORM query
func (o QueryOptions) Apply(tx *gorm.DB) *gorm.DB {
	// Apply filters
	for field, value := range o.Filters {
		// Make sure field exists in resource schema
		if f := o.Resource.GetField(field); f != nil {
			tx = tx.Where(fmt.Sprintf("%s = ?", field), value)
		}
	}

	// Apply advanced filters
	tx = applyAdvancedFilters(tx, o.AdvancedFilters, o.Resource)

	// Apply search if provided
	if o.Search != "" && o.Resource.GetSearchable() != nil {
		searchableFields := o.Resource.GetSearchable()
		if len(searchableFields) > 0 {
			var conditions []string
			var values []interface{}
			for _, field := range searchableFields {
				conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
				values = append(values, fmt.Sprintf("%%%s%%", o.Search))
			}
			tx = tx.Where(strings.Join(conditions, " OR "), values...)
		}
	}

	// Apply sorting
	if o.Sort != "" {
		// Check if we have multiple sort fields (comma-separated)
		if strings.Contains(o.Sort, ",") {
			// The sort string already contains both field and order information
			tx = tx.Order(o.Sort)
		} else {
			// Single sort field - validate existence in resource schema
			if f := o.Resource.GetField(o.Sort); f != nil {
				tx = tx.Order(fmt.Sprintf("%s %s", o.Sort, o.Order))
			}
		}
	}

	return tx
}

// ApplyWithPagination applies all query options including pagination to a GORM query
// Returns the updated query and the total count of records before pagination
func (o QueryOptions) ApplyWithPagination(tx *gorm.DB, dest interface{}) (int64, error) {
	// Apply non-pagination filters
	tx = o.Apply(tx)

	// Get total count
	var count int64
	countQuery := tx
	if err := countQuery.Count(&count).Error; err != nil {
		return 0, err
	}

	// Apply pagination if not disabled
	if !o.DisablePagination && dest != nil {
		offset := (o.Page - 1) * o.PerPage
		tx = tx.Offset(offset).Limit(o.PerPage)

		// Find records with pagination
		if err := tx.Find(dest).Error; err != nil {
			return 0, err
		}
	}

	return count, nil
}

// applyAdvancedFilters applies advanced filters with operators to a GORM query
func applyAdvancedFilters(tx *gorm.DB, filters []Filter, res resource.Resource) *gorm.DB {
	for _, filter := range filters {
		// Make sure field exists in resource schema
		if f := res.GetField(filter.Field); f != nil {
			// Apply based on operator
			switch strings.ToLower(filter.Operator) {
			case "eq":
				tx = tx.Where(fmt.Sprintf("`%s` = ?", filter.Field), filter.Value)
			case "ne":
				tx = tx.Where(fmt.Sprintf("`%s` <> ?", filter.Field), filter.Value)
			case "lt":
				tx = tx.Where(fmt.Sprintf("`%s` < ?", filter.Field), filter.Value)
			case "gt":
				tx = tx.Where(fmt.Sprintf("`%s` > ?", filter.Field), filter.Value)
			case "lte":
				tx = tx.Where(fmt.Sprintf("`%s` <= ?", filter.Field), filter.Value)
			case "gte":
				tx = tx.Where(fmt.Sprintf("`%s` >= ?", filter.Field), filter.Value)
			case "contains":
				tx = tx.Where(fmt.Sprintf("`%s` LIKE ?", filter.Field), fmt.Sprintf("%%%v%%", filter.Value))
			case "containsi":
				tx = tx.Where(fmt.Sprintf("LOWER(`%s`) LIKE LOWER(?)", filter.Field), fmt.Sprintf("%%%v%%", filter.Value))
			case "startswith":
				tx = tx.Where(fmt.Sprintf("`%s` LIKE ?", filter.Field), fmt.Sprintf("%v%%", filter.Value))
			case "endswith":
				tx = tx.Where(fmt.Sprintf("`%s` LIKE ?", filter.Field), fmt.Sprintf("%%%v", filter.Value))
			case "null":
				value := filter.Value
				boolValue, ok := value.(bool)
				if ok && boolValue {
					tx = tx.Where(fmt.Sprintf("`%s` IS NULL", filter.Field))
				} else {
					tx = tx.Where(fmt.Sprintf("`%s` IS NOT NULL", filter.Field))
				}
			case "in":
				// Handle array values
				if reflect.TypeOf(filter.Value).Kind() == reflect.String {
					// If string, split by comma
					values := strings.Split(filter.Value.(string), ",")
					tx = tx.Where(fmt.Sprintf("`%s` IN ?", filter.Field), values)
				} else {
					// Already an array/slice
					tx = tx.Where(fmt.Sprintf("`%s` IN ?", filter.Field), filter.Value)
				}
			default:
				// Default to equality
				tx = tx.Where(fmt.Sprintf("`%s` = ?", filter.Field), filter.Value)
			}
		}
	}
	return tx
}

// ParseQueryOptions parses query parameters from gin context
func ParseQueryOptions(c *gin.Context, res resource.Resource) QueryOptions {
	return NewQueryOptions(c, res)
}

// ParsePaginationResponse creates a pagination response
func ParsePaginationResponse(opts QueryOptions, total int64) map[string]interface{} {
	return map[string]interface{}{
		"page":      opts.Page,
		"per_page":  opts.PerPage,
		"total":     total,
		"last_page": int(total+int64(opts.PerPage)-1) / opts.PerPage,
	}
}

// ToResult converts paginated data to a response format
func ToResult(data interface{}, meta map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": data,
		"meta": meta,
	}
}

// ParseId parses an ID from a string
func ParseId(id string) (uint, error) {
	// Convert the ID to int
	var idInt int
	if _, err := fmt.Sscanf(id, "%d", &idInt); err != nil {
		return 0, err
	}
	return uint(idInt), nil
}

// ParseQueryParam parses a query parameter with type conversion
func ParseQueryParam(value string, fieldType string) (interface{}, error) {
	switch fieldType {
	case "int":
		return strconv.Atoi(value)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return strconv.ParseBool(value)
	default:
		return value, nil
	}
}
