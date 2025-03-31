package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Common errors
var (
	ErrInvalidType = errors.New("invalid type")
)

// IsSlice checks if the given value is a slice
func IsSlice(data interface{}) bool {
	v := reflect.ValueOf(data)

	// If it's a pointer, get the element
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return v.Kind() == reflect.Slice
}

// SetFieldValue sets a field value on an object using reflection
func SetFieldValue(obj interface{}, fieldName string, value interface{}) error {
	if obj == nil {
		return errors.New("object cannot be nil")
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Check if the object is valid for setting field
	if !v.IsValid() {
		return errors.New("object is not valid for setting field")
	}

	// Find the field
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found", fieldName)
	}

	// Check if the field is settable
	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}

	// Convert value to the field's type
	valueVal := reflect.ValueOf(value)
	if !valueVal.IsValid() {
		return errors.New("value is not valid")
	}

	// Try to convert the value to the field's type
	if valueVal.Type().ConvertibleTo(field.Type()) {
		field.Set(valueVal.Convert(field.Type()))
		return nil
	}

	return fmt.Errorf("value type %s cannot be converted to field type %s", valueVal.Type(), field.Type())
}

// GetFieldValue gets a field value from an object using reflection
func GetFieldValue(obj interface{}, fieldName string) (interface{}, error) {
	if obj == nil {
		return nil, errors.New("object cannot be nil")
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Check if the object is valid
	if !v.IsValid() {
		return nil, errors.New("object is not valid")
	}

	// Find the field
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s not found", fieldName)
	}

	// Return the field value
	return field.Interface(), nil
}

// FilterOutReadOnlyFields removes read-only fields from the update data
// This should be used before updating a resource to ensure read-only fields cannot be modified
func FilterOutReadOnlyFields(data interface{}, res Resource) interface{} {
	if data == nil {
		return nil
	}

	// Handle nil pointers
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr && dataValue.IsNil() {
		return nil
	}

	// Get a map of field names that are editable
	editableFields := make(map[string]bool)
	for _, fieldName := range res.GetEditableFields() {
		editableFields[strings.ToLower(fieldName)] = true
	}

	// If no editable fields are defined, filter by ReadOnly flag directly
	if len(editableFields) == 0 {
		fields := res.GetFields()
		for _, field := range fields {
			if !field.ReadOnly {
				editableFields[strings.ToLower(field.Name)] = true
			}
		}
	}

	// For map type, filter out read-only fields directly
	if mapData, ok := data.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		for key, value := range mapData {
			if editableFields[strings.ToLower(key)] {
				result[key] = value
			}
		}
		return result
	}

	// For struct type, create a new struct with only editable fields
	dataValue = reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	// Only handle struct types
	if dataValue.Kind() != reflect.Struct {
		return data // Return as is if not a struct
	}

	// Create a map to hold editable fields
	result := make(map[string]interface{})

	dataType := dataValue.Type()
	for i := 0; i < dataValue.NumField(); i++ {
		field := dataType.Field(i)
		fieldName := field.Name

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Check if field is editable
		if editableFields[strings.ToLower(fieldName)] {
			// Add field to result
			result[fieldName] = dataValue.Field(i).Interface()
		}
	}

	return result
}

// ValidateNestedJson validates a nested JSON structure against a JsonConfig
func ValidateNestedJson(data interface{}, config *JsonConfig) (bool, []string) {
	if data == nil || config == nil {
		return false, []string{"Invalid input: nil data or config"}
	}

	errors := make([]string, 0)

	// Convert data to map if needed
	var dataMap map[string]interface{}

	// Handle different input types
	switch v := data.(type) {
	case map[string]interface{}:
		dataMap = v
	case string:
		// Try to parse as JSON string
		if err := json.Unmarshal([]byte(v), &dataMap); err != nil {
			return false, []string{"Invalid JSON string: " + err.Error()}
		}
	default:
		// Try to convert using reflection
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return false, []string{"Cannot convert data to JSON: " + err.Error()}
		}

		if err := json.Unmarshal(jsonBytes, &dataMap); err != nil {
			return false, []string{"Cannot convert data to map: " + err.Error()}
		}
	}

	// No properties defined, consider valid
	if len(config.Properties) == 0 {
		return true, errors
	}

	// Validate each property
	for _, prop := range config.Properties {
		// Skip validation for properties without validation rules
		if prop.Validation == nil {
			continue
		}

		// Get the value using the property path
		value, err := getValueByPath(dataMap, prop.Path)
		if err != nil {
			// Only report error if the property is required
			if prop.Validation.Required {
				errors = append(errors, fmt.Sprintf("Required property '%s' not found: %s", prop.Path, err.Error()))
			}
			continue
		}

		// Validate required
		if prop.Validation.Required && value == nil {
			errors = append(errors, fmt.Sprintf("Property '%s' is required but has nil value", prop.Path))
			continue
		}

		// Skip further validation if value is nil
		if value == nil {
			continue
		}

		// Validate min/max for numbers
		if prop.Type == "number" || prop.Type == "integer" {
			// Convert to number
			var num float64
			switch v := value.(type) {
			case float64:
				num = v
			case float32:
				num = float64(v)
			case int:
				num = float64(v)
			case int64:
				num = float64(v)
			case string:
				parsed, err := strconv.ParseFloat(v, 64)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Property '%s' should be a number but got '%v'", prop.Path, v))
					continue
				}
				num = parsed
			default:
				errors = append(errors, fmt.Sprintf("Property '%s' should be a number but got '%T'", prop.Path, value))
				continue
			}

			// Validate min
			if prop.Validation.Min != 0 && num < prop.Validation.Min {
				errors = append(errors, fmt.Sprintf("Property '%s' value %v is less than minimum %v", prop.Path, num, prop.Validation.Min))
			}

			// Validate max
			if prop.Validation.Max != 0 && num > prop.Validation.Max {
				errors = append(errors, fmt.Sprintf("Property '%s' value %v is greater than maximum %v", prop.Path, num, prop.Validation.Max))
			}
		}

		// Validate min/max length for strings
		if prop.Type == "string" {
			str, ok := value.(string)
			if !ok {
				errors = append(errors, fmt.Sprintf("Property '%s' should be a string but got '%T'", prop.Path, value))
				continue
			}

			// Validate minLength
			if prop.Validation.MinLength > 0 && len(str) < prop.Validation.MinLength {
				errors = append(errors, fmt.Sprintf("Property '%s' length %d is less than minimum length %d", prop.Path, len(str), prop.Validation.MinLength))
			}

			// Validate maxLength
			if prop.Validation.MaxLength > 0 && len(str) > prop.Validation.MaxLength {
				errors = append(errors, fmt.Sprintf("Property '%s' length %d is greater than maximum length %d", prop.Path, len(str), prop.Validation.MaxLength))
			}

			// Validate pattern
			if prop.Validation.Pattern != "" {
				matched, err := regexp.MatchString(prop.Validation.Pattern, str)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Property '%s' has invalid pattern: %s", prop.Path, err.Error()))
				} else if !matched {
					errors = append(errors, fmt.Sprintf("Property '%s' value '%s' does not match pattern '%s'", prop.Path, str, prop.Validation.Pattern))
				}
			}
		}

		// Validate nested objects
		if prop.Type == "object" && len(prop.Properties) > 0 {
			nestedObj, ok := value.(map[string]interface{})
			if !ok {
				// Try to convert
				jsonBytes, err := json.Marshal(value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Property '%s' should be an object but got '%T'", prop.Path, value))
					continue
				}

				var nestedMap map[string]interface{}
				if err := json.Unmarshal(jsonBytes, &nestedMap); err != nil {
					errors = append(errors, fmt.Sprintf("Property '%s' should be an object but got '%T'", prop.Path, value))
					continue
				}
				nestedObj = nestedMap
			}

			// Create nested config
			nestedConfig := &JsonConfig{
				Properties: prop.Properties,
			}

			// Validate nested object
			valid, nestedErrors := ValidateNestedJson(nestedObj, nestedConfig)
			if !valid {
				// Prefix nested errors with property path
				for i, err := range nestedErrors {
					nestedErrors[i] = fmt.Sprintf("%s.%s", prop.Path, err)
				}
				errors = append(errors, nestedErrors...)
			}
		}

		// Validate arrays
		if prop.Type == "array" {
			arr, ok := value.([]interface{})
			if !ok {
				// Try to convert
				jsonBytes, err := json.Marshal(value)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Property '%s' should be an array but got '%T'", prop.Path, value))
					continue
				}

				var arrVal []interface{}
				if err := json.Unmarshal(jsonBytes, &arrVal); err != nil {
					errors = append(errors, fmt.Sprintf("Property '%s' should be an array but got '%T'", prop.Path, value))
					continue
				}
				arr = arrVal
			}

			// Validate min items
			if prop.Validation.MinLength > 0 && len(arr) < prop.Validation.MinLength {
				errors = append(errors, fmt.Sprintf("Property '%s' has %d items which is less than minimum %d", prop.Path, len(arr), prop.Validation.MinLength))
			}

			// Validate max items
			if prop.Validation.MaxLength > 0 && len(arr) > prop.Validation.MaxLength {
				errors = append(errors, fmt.Sprintf("Property '%s' has %d items which is greater than maximum %d", prop.Path, len(arr), prop.Validation.MaxLength))
			}
		}
	}

	return len(errors) == 0, errors
}

// getValueByPath retrieves a value from a nested map using a dot-separated path
func getValueByPath(data map[string]interface{}, path string) (interface{}, error) {
	if data == nil {
		return nil, fmt.Errorf("nil data")
	}

	// Handle empty path
	if path == "" {
		return data, nil
	}

	// Split path into parts
	parts := strings.Split(path, ".")

	// Navigate through the nested structure
	var current interface{} = data

	for i, part := range parts {
		// Handle array indexing if part contains [index]
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Extract field name and index
			openBracket := strings.Index(part, "[")
			closeBracket := strings.Index(part, "]")

			if openBracket > 0 && closeBracket > openBracket {
				fieldName := part[:openBracket]
				indexStr := part[openBracket+1 : closeBracket]

				// Get the array
				currentMap, ok := current.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("expected object at path '%s'", strings.Join(parts[:i], "."))
				}

				array, ok := currentMap[fieldName]
				if !ok {
					return nil, fmt.Errorf("field '%s' not found", fieldName)
				}

				// Parse index
				index, err := strconv.Atoi(indexStr)
				if err != nil {
					return nil, fmt.Errorf("invalid array index '%s'", indexStr)
				}

				// Get array item
				arrayItems, ok := array.([]interface{})
				if !ok {
					return nil, fmt.Errorf("field '%s' is not an array", fieldName)
				}

				if index < 0 || index >= len(arrayItems) {
					return nil, fmt.Errorf("array index %d out of bounds (0-%d)", index, len(arrayItems)-1)
				}

				current = arrayItems[index]

				// Convert to float64 for integer values in JSON
				if num, ok := current.(int); ok {
					current = float64(num)
				}

				continue
			}
		}

		// Regular field access
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected object at path '%s'", strings.Join(parts[:i], "."))
		}

		current, ok = currentMap[part]
		if !ok {
			return nil, fmt.Errorf("field '%s' not found", part)
		}

		// Convert to float64 for integer values in JSON
		if num, ok := current.(int); ok {
			current = float64(num)
		}
	}

	return current, nil
}

// MapValidationToAntDesignRules maps standard validation rules to Ant Design Form rules
func MapValidationToAntDesignRules(validation *Validation) []AntDesignRuleMetadata {
	if validation == nil {
		return nil
	}

	rules := make([]AntDesignRuleMetadata, 0)

	// Required rule
	if validation.Required {
		rules = append(rules, AntDesignRuleMetadata{
			Type:            "required",
			Message:         validation.Message,
			ValidateTrigger: "onBlur",
		})
	}

	// MinLength rule for strings
	if validation.MinLength > 0 {
		rules = append(rules, AntDesignRuleMetadata{
			Type:            "min",
			Value:           validation.MinLength,
			Message:         fmt.Sprintf("Minimum length is %d characters", validation.MinLength),
			ValidateTrigger: "onBlur",
		})
	}

	// MaxLength rule for strings
	if validation.MaxLength > 0 {
		rules = append(rules, AntDesignRuleMetadata{
			Type:            "max",
			Value:           validation.MaxLength,
			Message:         fmt.Sprintf("Maximum length is %d characters", validation.MaxLength),
			ValidateTrigger: "onBlur",
		})
	}

	// Pattern rule
	if validation.Pattern != "" {
		rules = append(rules, AntDesignRuleMetadata{
			Type:            "pattern",
			Pattern:         validation.Pattern,
			Message:         validation.Message,
			ValidateTrigger: "onBlur",
		})
	}

	// Min value rule for numbers
	if validation.Min != 0 {
		rules = append(rules, AntDesignRuleMetadata{
			Type:            "min",
			Value:           validation.Min,
			Message:         fmt.Sprintf("Minimum value is %v", validation.Min),
			ValidateTrigger: "onBlur",
		})
	}

	// Max value rule for numbers
	if validation.Max != 0 {
		rules = append(rules, AntDesignRuleMetadata{
			Type:            "max",
			Value:           validation.Max,
			Message:         fmt.Sprintf("Maximum value is %v", validation.Max),
			ValidateTrigger: "onBlur",
		})
	}

	// Custom validation rule
	if validation.Custom != "" {
		rules = append(rules, AntDesignRuleMetadata{
			Type:            "custom",
			Value:           validation.Custom,
			Message:         validation.Message,
			ValidateTrigger: "onBlur",
		})
	}

	// Conditional validation rule
	if validation.Conditional != nil {
		rules = append(rules, AntDesignRuleMetadata{
			Type: "conditional",
			Value: map[string]interface{}{
				"field":    validation.Conditional.Field,
				"operator": validation.Conditional.Operator,
				"value":    validation.Conditional.Value,
			},
			Message:         validation.Conditional.Message,
			ValidateTrigger: "onChange",
		})
	}

	// Async validation rule
	if validation.AsyncValidator != "" {
		rules = append(rules, AntDesignRuleMetadata{
			Type:            "async",
			Value:           validation.AsyncValidator,
			Message:         validation.Message,
			ValidateTrigger: "onBlur",
		})
	}

	return rules
}

// AutoDetectAntDesignComponent automatically detects the appropriate Ant Design component
// based on the field type and configuration
func AutoDetectAntDesignComponent(field *Field) string {
	if field == nil {
		return "Input" // Default
	}

	// Map field type to Ant Design component
	switch field.Type {
	case "string":
		// Check if it's an enum/select type
		if len(field.Options) > 0 || field.Select != nil {
			return "Select"
		}
		// Check if it's a password field
		if strings.Contains(strings.ToLower(field.Name), "password") {
			return "Password"
		}
		// Check if it's a rich text field
		if field.RichText != nil {
			return "TextArea"
		}
		// Default to Input
		return "Input"
	case "text":
		if field.RichText != nil {
			return "TextArea"
		}
		return "TextArea"
	case "number", "integer", "float", "double", "decimal":
		return "InputNumber"
	case "boolean":
		return "Switch"
	case "date":
		return "DatePicker"
	case "time":
		return "TimePicker"
	case "datetime":
		return "DatePicker" // With showTime prop
	case "json":
		return "JsonEditor" // Custom component
	case "file":
		if field.File != nil && field.File.IsImage {
			return "Upload.Image"
		}
		return "Upload"
	case "select":
		return "Select"
	case "multiselect":
		return "Select" // With mode="multiple" prop
	case "checkbox":
		return "Checkbox"
	case "radio":
		return "Radio"
	default:
		return "Input"
	}
}
