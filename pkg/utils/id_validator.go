package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// IDValidationError represents an error during ID validation
type IDValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e IDValidationError) Error() string {
	return e.Message
}

// Common ID validation error codes
const (
	ErrCodeEmptyID        = "EMPTY_ID"
	ErrCodeInvalidIDType  = "INVALID_ID_TYPE"
	ErrCodeInvalidIDValue = "INVALID_ID_VALUE"
	ErrCodeDuplicateID    = "DUPLICATE_ID"
	ErrCodeIDNotFound     = "ID_NOT_FOUND"
)

// ValidateIDFormat validates a single ID value
func ValidateIDFormat(id interface{}) error {
	if id == nil {
		return IDValidationError{
			Code:    ErrCodeEmptyID,
			Message: "ID cannot be nil",
		}
	}

	switch v := id.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be empty string",
			}
		}
		// Check if string ID contains only valid characters
		if !isValidStringID(v) {
			return IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("Invalid ID format: %s", v),
			}
		}
	case int:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
		if v < 0 {
			return IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("ID cannot be negative: %v", v),
			}
		}
	case int8:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
		if v < 0 {
			return IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("ID cannot be negative: %v", v),
			}
		}
	case int16:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
		if v < 0 {
			return IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("ID cannot be negative: %v", v),
			}
		}
	case int32:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
		if v < 0 {
			return IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("ID cannot be negative: %v", v),
			}
		}
	case int64:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
		if v < 0 {
			return IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("ID cannot be negative: %v", v),
			}
		}
	case uint:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
	case uint8:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
	case uint16:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
	case uint32:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
	case uint64:
		if v == 0 {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
	case float32:
		// Convert to string to check if it's a valid number
		idStr := fmt.Sprintf("%v", v)
		if idStr == "0" || idStr == "0.0" {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
		if v < 0 {
			return IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("ID cannot be negative: %v", v),
			}
		}
	case float64:
		// Convert to string to check if it's a valid number
		idStr := fmt.Sprintf("%v", v)
		if idStr == "0" || idStr == "0.0" {
			return IDValidationError{
				Code:    ErrCodeEmptyID,
				Message: "ID cannot be zero",
			}
		}
		if v < 0 {
			return IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("ID cannot be negative: %v", v),
			}
		}
	default:
		return IDValidationError{
			Code:    ErrCodeInvalidIDType,
			Message: fmt.Sprintf("Unsupported ID type: %T", id),
		}
	}

	return nil
}

// ValidateIDSlice validates a slice of IDs
func ValidateIDSlice(ids []interface{}) error {
	if len(ids) == 0 {
		return IDValidationError{
			Code:    ErrCodeEmptyID,
			Message: "IDs array cannot be empty",
		}
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for i, id := range ids {
		// Validate individual ID
		if err := ValidateIDFormat(id); err != nil {
			return IDValidationError{
				Code:    err.(IDValidationError).Code,
				Message: fmt.Sprintf("Invalid ID at index %d: %s", i, err.Error()),
				Details: map[string]interface{}{
					"index": i,
					"id":    id,
				},
			}
		}

		// Convert ID to string for duplicate checking
		idStr := fmt.Sprintf("%v", id)
		if seen[idStr] {
			return IDValidationError{
				Code:    ErrCodeDuplicateID,
				Message: fmt.Sprintf("Duplicate ID found: %s", idStr),
				Details: map[string]interface{}{
					"duplicate_id": idStr,
				},
			}
		}
		seen[idStr] = true
	}

	return nil
}

// ParseAndValidateIDs parses and validates IDs from various formats
func ParseAndValidateIDs(idData interface{}) ([]interface{}, error) {
	if idData == nil {
		return nil, IDValidationError{
			Code:    ErrCodeEmptyID,
			Message: "IDs data cannot be nil",
		}
	}

	var ids []interface{}

	// Handle different ID formats
	switch v := idData.(type) {
	case []interface{}:
		ids = v
	case []string:
		// Convert string slice to interface slice
		for _, id := range v {
			ids = append(ids, id)
		}
	case []int:
		// Convert int slice to interface slice
		for _, id := range v {
			ids = append(ids, id)
		}
	case []int64:
		// Convert int64 slice to interface slice
		for _, id := range v {
			ids = append(ids, id)
		}
	case string:
		// Single string ID
		ids = []interface{}{v}
	case int, int64, uint, uint64:
		// Single numeric ID
		ids = []interface{}{v}
	default:
		// Try to convert using reflection
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Slice {
			for i := 0; i < val.Len(); i++ {
				ids = append(ids, val.Index(i).Interface())
			}
		} else {
			// Single value
			ids = []interface{}{v}
		}
	}

	// Validate the parsed IDs
	if err := ValidateIDSlice(ids); err != nil {
		return nil, err
	}

	return ids, nil
}

// ConvertIDToString converts an ID to string format
func ConvertIDToString(id interface{}) (string, error) {
	if err := ValidateIDFormat(id); err != nil {
		return "", err
	}

	switch v := id.(type) {
	case string:
		return v, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v), nil
	case float32, float64:
		// Convert float to string without decimal if it's a whole number
		idStr := fmt.Sprintf("%v", v)
		if strings.HasSuffix(idStr, ".0") {
			return strings.TrimSuffix(idStr, ".0"), nil
		}
		return idStr, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ConvertStringToID converts a string ID to the appropriate type
func ConvertStringToID(idStr string, targetType reflect.Type) (interface{}, error) {
	if strings.TrimSpace(idStr) == "" {
		return nil, IDValidationError{
			Code:    ErrCodeEmptyID,
			Message: "ID string cannot be empty",
		}
	}

	switch targetType.Kind() {
	case reflect.String:
		return idStr, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("Cannot convert '%s' to %s: %v", idStr, targetType.String(), err),
			}
		}
		return val, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("Cannot convert '%s' to %s: %v", idStr, targetType.String(), err),
			}
		}
		return val, nil
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(idStr, 64)
		if err != nil {
			return nil, IDValidationError{
				Code:    ErrCodeInvalidIDValue,
				Message: fmt.Sprintf("Cannot convert '%s' to %s: %v", idStr, targetType.String(), err),
			}
		}
		return val, nil
	default:
		return nil, IDValidationError{
			Code:    ErrCodeInvalidIDType,
			Message: fmt.Sprintf("Unsupported target type for ID conversion: %s", targetType.String()),
		}
	}
}

// isValidStringID checks if a string ID contains only valid characters
func isValidStringID(id string) bool {
	// Allow alphanumeric characters, hyphens, underscores, and dots
	// This is a basic validation - you can customize based on your requirements
	for _, char := range id {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.') {
			return false
		}
	}
	return true
}

// IsIDValidationError checks if an error is an IDValidationError
func IsIDValidationError(err error) bool {
	var idErr IDValidationError
	return errors.As(err, &idErr)
}

// GetIDValidationErrorCode returns the error code if it's an IDValidationError
func GetIDValidationErrorCode(err error) string {
	var idErr IDValidationError
	if errors.As(err, &idErr) {
		return idErr.Code
	}
	return ""
}