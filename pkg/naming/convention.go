package naming

import (
	"strings"
)

// NamingConvention represents the naming convention to use
type NamingConvention string

const (
	// CamelCase naming convention (e.g., "firstName", "lastName")
	CamelCase NamingConvention = "camelCase"
	
	// SnakeCase naming convention (e.g., "first_name", "last_name")
	SnakeCase NamingConvention = "snake_case"
	
	// PascalCase naming convention (e.g., "FirstName", "LastName")
	PascalCase NamingConvention = "PascalCase"
	
	// KebabCase naming convention (e.g., "first-name", "last-name")
	KebabCase NamingConvention = "kebab-case"
)

// RefineConfig contains configuration for Refine.dev integration
type RefineConfig struct {
	// Naming convention for JSON field names
	NamingConvention NamingConvention
	
	// ID field name (default: "id")
	IDFieldName string
	
	// Enable real-time features (future feature)
	EnableRealTime bool
	
	// Custom field mappings for specific fields
	FieldMappings map[string]string
}

// DefaultRefineConfig returns the default configuration for Refine.dev
func DefaultRefineConfig() *RefineConfig {
	return &RefineConfig{
		NamingConvention: CamelCase, // Default to camelCase for Refine.dev
		IDFieldName:      "id",      // Default ID field name
		EnableRealTime:   false,     // Disabled by default
		FieldMappings:    make(map[string]string),
	}
}

// WithNamingConvention sets the naming convention
func (c *RefineConfig) WithNamingConvention(convention NamingConvention) *RefineConfig {
	c.NamingConvention = convention
	return c
}

// WithIDFieldName sets the ID field name
func (c *RefineConfig) WithIDFieldName(fieldName string) *RefineConfig {
	c.IDFieldName = fieldName
	return c
}

// WithFieldMapping adds a custom field mapping
func (c *RefineConfig) WithFieldMapping(from, to string) *RefineConfig {
	if c.FieldMappings == nil {
		c.FieldMappings = make(map[string]string)
	}
	c.FieldMappings[from] = to
	return c
}

// ConvertFieldName converts a field name according to the naming convention
func (c *RefineConfig) ConvertFieldName(fieldName string) string {
	// Check for custom mapping first
	if mapped, exists := c.FieldMappings[fieldName]; exists {
		return mapped
	}

	switch c.NamingConvention {
	case CamelCase:
		return toCamelCase(fieldName)
	case SnakeCase:
		return toSnakeCase(fieldName)
	case PascalCase:
		return toPascalCase(fieldName)
	case KebabCase:
		return toKebabCase(fieldName)
	default:
		return fieldName
	}
}

// ConvertFieldNames converts multiple field names
func (c *RefineConfig) ConvertFieldNames(fieldNames []string) []string {
	result := make([]string, len(fieldNames))
	for i, fieldName := range fieldNames {
		result[i] = c.ConvertFieldName(fieldName)
	}
	return result
}

// toCamelCase converts a string to camelCase
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	// Handle common separators
	parts := splitBySeparators(s)
	if len(parts) == 0 {
		return s
	}

	result := strings.ToLower(parts[0])
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			result += strings.Title(strings.ToLower(parts[i]))
		}
	}

	return result
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	for i, char := range s {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(char)
	}

	return strings.ToLower(result.String())
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	if s == "" {
		return s
	}

	parts := splitBySeparators(s)
	if len(parts) == 0 {
		return strings.Title(s)
	}

	var result strings.Builder
	for _, part := range parts {
		if part != "" {
			result.WriteString(strings.Title(strings.ToLower(part)))
		}
	}

	return result.String()
}

// toKebabCase converts a string to kebab-case
func toKebabCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	for i, char := range s {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result.WriteByte('-')
		}
		result.WriteRune(char)
	}

	return strings.ToLower(result.String())
}

// splitBySeparators splits a string by common separators
func splitBySeparators(s string) []string {
	// Replace common separators with spaces
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, ".", " ")
	
	// Split by spaces and filter empty parts
	parts := strings.Fields(s)
	return parts
}

// IsValidNamingConvention checks if a naming convention is valid
func IsValidNamingConvention(convention string) bool {
	switch NamingConvention(convention) {
	case CamelCase, SnakeCase, PascalCase, KebabCase:
		return true
	default:
		return false
	}
}

// GetNamingConventionFromString converts a string to NamingConvention
func GetNamingConventionFromString(s string) NamingConvention {
	switch strings.ToLower(s) {
	case "camelcase", "camel_case":
		return CamelCase
	case "snakecase", "snake_case":
		return SnakeCase
	case "pascalcase", "pascal_case":
		return PascalCase
	case "kebabcase", "kebab_case", "kebab-case":
		return KebabCase
	default:
		return CamelCase // Default fallback
	}
}