package utils

import (
	"strings"
)

// TypeCategory represents a general category of types
type TypeCategory string

const (
	TypeString   TypeCategory = "string"
	TypeInteger  TypeCategory = "integer"
	TypeNumber   TypeCategory = "number"
	TypeBoolean  TypeCategory = "boolean"
	TypeDateTime TypeCategory = "datetime"
	TypeArray    TypeCategory = "array"
	TypeObject   TypeCategory = "object"
)

// TypeMapping represents a mapping between Go types and other type systems
type TypeMapping struct {
	// The general category of the type
	Category TypeCategory
	// The format of the type (e.g., "int64", "float", "date-time")
	Format string
	// Whether the type is a primitive
	IsPrimitive bool
}

// GetTypeMapping returns the type mapping for a given Go type
func GetTypeMapping(goType string) TypeMapping {
	// Handle array types
	if strings.HasPrefix(goType, "[]") {
		return TypeMapping{
			Category:    TypeArray,
			Format:      strings.TrimPrefix(goType, "[]"),
			IsPrimitive: false,
		}
	}

	// Map Go types to general categories
	switch goType {
	case "string":
		return TypeMapping{
			Category:    TypeString,
			IsPrimitive: true,
		}
	case "int", "int8", "int16", "int32":
		return TypeMapping{
			Category:    TypeInteger,
			Format:      "int32",
			IsPrimitive: true,
		}
	case "int64":
		return TypeMapping{
			Category:    TypeInteger,
			Format:      "int64",
			IsPrimitive: true,
		}
	case "uint", "uint8", "uint16", "uint32":
		return TypeMapping{
			Category:    TypeInteger,
			Format:      "uint32",
			IsPrimitive: true,
		}
	case "uint64":
		return TypeMapping{
			Category:    TypeInteger,
			Format:      "uint64",
			IsPrimitive: true,
		}
	case "float32":
		return TypeMapping{
			Category:    TypeNumber,
			Format:      "float",
			IsPrimitive: true,
		}
	case "float64":
		return TypeMapping{
			Category:    TypeNumber,
			Format:      "double",
			IsPrimitive: true,
		}
	case "bool":
		return TypeMapping{
			Category:    TypeBoolean,
			IsPrimitive: true,
		}
	case "time.Time":
		return TypeMapping{
			Category:    TypeDateTime,
			Format:      "date-time",
			IsPrimitive: true,
		}
	default:
		return TypeMapping{
			Category:    TypeObject,
			Format:      goType,
			IsPrimitive: false,
		}
	}
}

// IsNumericType checks if the given Go type is numeric
func IsNumericType(goType string) bool {
	mapping := GetTypeMapping(goType)
	return mapping.Category == TypeInteger || mapping.Category == TypeNumber
}

// IsStringType checks if the given Go type is a string
func IsStringType(goType string) bool {
	mapping := GetTypeMapping(goType)
	return mapping.Category == TypeString
}

// IsArrayType checks if the given Go type is an array
func IsArrayType(goType string) bool {
	mapping := GetTypeMapping(goType)
	return mapping.Category == TypeArray
}

// GetArrayElementType returns the element type of an array type
func GetArrayElementType(goType string) string {
	if IsArrayType(goType) {
		return strings.TrimPrefix(goType, "[]")
	}
	return ""
}
