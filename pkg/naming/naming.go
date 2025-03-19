package naming

import (
	"strings"
	"unicode"
)

// NamingConvention represents a naming convention for JSON fields
type NamingConvention string

const (
	// SnakeCase represents snake_case convention (e.g. first_name)
	SnakeCase NamingConvention = "snake_case"

	// CamelCase represents camelCase convention (e.g. firstName)
	CamelCase NamingConvention = "camelCase"

	// PascalCase represents PascalCase convention (e.g. FirstName)
	PascalCase NamingConvention = "PascalCase"
)

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

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	result := ""
	nextUpper := false
	for i, r := range s {
		if r == '_' {
			nextUpper = true
		} else if nextUpper {
			result += string(unicode.ToUpper(r))
			nextUpper = false
		} else if i == 0 {
			result += string(unicode.ToLower(r))
		} else {
			result += string(r)
		}
	}
	return result
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	result := ""
	nextUpper := true
	for _, r := range s {
		if r == '_' {
			nextUpper = true
		} else if nextUpper {
			result += string(unicode.ToUpper(r))
			nextUpper = false
		} else {
			result += string(r)
		}
	}
	return result
}

// ConvertKeys converts all keys in a map to the specified naming convention
func ConvertKeys(data map[string]interface{}, convention NamingConvention) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range data {
		var newKey string
		switch convention {
		case SnakeCase:
			newKey = ToSnakeCase(k)
		case CamelCase:
			newKey = ToCamelCase(k)
		case PascalCase:
			newKey = ToPascalCase(k)
		default:
			newKey = k
		}

		// Convert nested maps recursively
		if nestedMap, ok := v.(map[string]interface{}); ok {
			result[newKey] = ConvertKeys(nestedMap, convention)
		} else if nestedSlice, ok := v.([]interface{}); ok {
			// Convert maps in slices recursively
			newSlice := make([]interface{}, len(nestedSlice))
			for i, item := range nestedSlice {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					newSlice[i] = ConvertKeys(nestedMap, convention)
				} else {
					newSlice[i] = item
				}
			}
			result[newKey] = newSlice
		} else {
			result[newKey] = v
		}
	}

	return result
}
