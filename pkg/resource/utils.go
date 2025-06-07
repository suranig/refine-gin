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

	// Get editable fields map
	editableFields := getEditableFieldsMap(res)

	// Process based on data type
	if mapData, ok := data.(map[string]interface{}); ok {
		return filterMapData(mapData, editableFields)
	}

	// If it's a struct, convert to map with only editable fields
	dataValue = reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	if dataValue.Kind() == reflect.Struct {
		return filterStructData(data, editableFields)
	}

	// Return as is for other types
	return data
}

// getEditableFieldsMap returns a map of field names that are editable
func getEditableFieldsMap(res Resource) map[string]bool {
	editableFields := make(map[string]bool)

	// First try to get explicitly defined editable fields
	definedFields := res.GetEditableFields()
	if len(definedFields) > 0 {
		for _, fieldName := range definedFields {
			editableFields[strings.ToLower(fieldName)] = true
		}
		return editableFields
	}

	// If no editable fields explicitly defined, filter by ReadOnly flag
	fields := res.GetFields()
	for _, field := range fields {
		if !field.ReadOnly {
			editableFields[strings.ToLower(field.Name)] = true
		}
	}

	return editableFields
}

// filterMapData filters a map to only include editable fields
func filterMapData(data map[string]interface{}, editableFields map[string]bool) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		if editableFields[strings.ToLower(key)] {
			result[key] = value
		}
	}
	return result
}

// filterStructData converts a struct to a map with only editable fields
func filterStructData(data interface{}, editableFields map[string]bool) map[string]interface{} {
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	// Only handle struct types
	if dataValue.Kind() != reflect.Struct {
		return nil // Return nil if not a struct
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

	// Convert data to map
	dataMap, conversionErrors := convertToJsonMap(data)
	if len(conversionErrors) > 0 {
		return false, conversionErrors
	}

	// No properties defined, consider valid
	if len(config.Properties) == 0 {
		return true, errors
	}

	// Validate each property
	for _, prop := range config.Properties {
		propErrors := validateProperty(dataMap, &prop)
		errors = append(errors, propErrors...)
	}

	return len(errors) == 0, errors
}

// convertToJsonMap converts various data types to a map[string]interface{}
func convertToJsonMap(data interface{}) (map[string]interface{}, []string) {
	var dataMap map[string]interface{}
	errors := make([]string, 0)

	// Handle different input types
	switch v := data.(type) {
	case map[string]interface{}:
		dataMap = v
	case string:
		// Try to parse as JSON string
		if err := json.Unmarshal([]byte(v), &dataMap); err != nil {
			return nil, []string{"Invalid JSON string: " + err.Error()}
		}
	default:
		// Try to convert using reflection
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, []string{"Cannot convert data to JSON: " + err.Error()}
		}

		if err := json.Unmarshal(jsonBytes, &dataMap); err != nil {
			return nil, []string{"Cannot convert data to map: " + err.Error()}
		}
	}

	return dataMap, errors
}

// validateProperty validates a single property according to its type and validation rules
func validateProperty(dataMap map[string]interface{}, prop *JsonProperty) []string {
	errors := make([]string, 0)

	// Skip validation for properties without validation rules
	if prop.Validation == nil {
		return errors
	}

	// Get the value using the property path
	value, err := getValueByPath(dataMap, prop.Path)
	if err != nil {
		// Only report error if the property is required
		if prop.Validation.Required {
			errors = append(errors, fmt.Sprintf("Required property '%s' not found: %s", prop.Path, err.Error()))
		}
		return errors
	}

	// Validate required
	if prop.Validation.Required && value == nil {
		errors = append(errors, fmt.Sprintf("Property '%s' is required but has nil value", prop.Path))
		return errors
	}

	// Skip further validation if value is nil
	if value == nil {
		return errors
	}

	// Validate based on property type
	switch prop.Type {
	case "number", "integer":
		errors = append(errors, validateNumericProperty(prop, value)...)
	case "string":
		errors = append(errors, validateStringProperty(prop, value)...)
	case "object":
		errors = append(errors, validateObjectProperty(prop, value)...)
	case "array":
		errors = append(errors, validateArrayProperty(prop, value)...)
	}

	return errors
}

// validateNumericProperty validates numeric properties
func validateNumericProperty(prop *JsonProperty, value interface{}) []string {
	errors := make([]string, 0)

	// Convert to number
	num, err := convertToNumber(value)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Property '%s' %s", prop.Path, err.Error()))
		return errors
	}

	// Validate min
	if prop.Validation.Min != 0 && num < prop.Validation.Min {
		errors = append(errors, fmt.Sprintf("Property '%s' value %v is less than minimum %v", prop.Path, num, prop.Validation.Min))
	}

	// Validate max
	if prop.Validation.Max != 0 && num > prop.Validation.Max {
		errors = append(errors, fmt.Sprintf("Property '%s' value %v is greater than maximum %v", prop.Path, num, prop.Validation.Max))
	}

	return errors
}

// convertToNumber converts a value to a float64
func convertToNumber(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("should be a number but got '%v'", v)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("should be a number but got '%T'", value)
	}
}

// validateStringProperty validates string properties
func validateStringProperty(prop *JsonProperty, value interface{}) []string {
	errors := make([]string, 0)

	str, ok := value.(string)
	if !ok {
		errors = append(errors, fmt.Sprintf("Property '%s' should be a string but got '%T'", prop.Path, value))
		return errors
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

	return errors
}

// validateObjectProperty validates object properties with nested validation
func validateObjectProperty(prop *JsonProperty, value interface{}) []string {
	errors := make([]string, 0)

	if len(prop.Properties) == 0 {
		return errors
	}

	nestedObj, ok := value.(map[string]interface{})
	if !ok {
		// Try to convert
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Property '%s' should be an object but got '%T'", prop.Path, value))
			return errors
		}

		var nestedMap map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &nestedMap); err != nil {
			errors = append(errors, fmt.Sprintf("Property '%s' should be an object but got '%T'", prop.Path, value))
			return errors
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

	return errors
}

// validateArrayProperty validates array properties
func validateArrayProperty(prop *JsonProperty, value interface{}) []string {
	errors := make([]string, 0)

	arr, ok := value.([]interface{})
	if !ok {
		// Try to convert
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Property '%s' should be an array but got '%T'", prop.Path, value))
			return errors
		}

		var arrVal []interface{}
		if err := json.Unmarshal(jsonBytes, &arrVal); err != nil {
			errors = append(errors, fmt.Sprintf("Property '%s' should be an array but got '%T'", prop.Path, value))
			return errors
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

	return errors
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
	var err error

	for i, part := range parts {
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			// Handle array access
			current, err = processArrayAccess(current, part, parts[:i])
			if err != nil {
				return nil, err
			}
		} else {
			// Regular field access
			current, err = processFieldAccess(current, part, parts[:i])
			if err != nil {
				return nil, err
			}
		}
	}

	return current, nil
}

// processArrayAccess handles array indexing in path expressions like "field[0]"
func processArrayAccess(current interface{}, part string, pathParts []string) (interface{}, error) {
	// Extract field name and index
	openBracket := strings.Index(part, "[")
	closeBracket := strings.Index(part, "]")

	if openBracket <= 0 || closeBracket <= openBracket {
		return nil, fmt.Errorf("invalid array access format in '%s'", part)
	}

	fieldName := part[:openBracket]
	indexStr := part[openBracket+1 : closeBracket]

	// Get the array
	currentMap, ok := current.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object at path '%s'", strings.Join(pathParts, "."))
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

	result := arrayItems[index]

	// Convert to float64 for integer values in JSON
	if num, ok := result.(int); ok {
		result = float64(num)
	}

	return result, nil
}

// processFieldAccess handles regular field access in path expressions
func processFieldAccess(current interface{}, part string, pathParts []string) (interface{}, error) {
	currentMap, ok := current.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected object at path '%s'", strings.Join(pathParts, "."))
	}

	value, ok := currentMap[part]
	if !ok {
		return nil, fmt.Errorf("field '%s' not found", part)
	}

	// Convert to float64 for integer values in JSON
	if num, ok := value.(int); ok {
		value = float64(num)
	}

	return value, nil
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

	// Use type-specific detection functions
	switch field.Type {
	case "string":
		return detectStringComponent(field)
	case "text":
		return detectTextComponent(field)
	case "number", "integer", "float", "double", "decimal":
		return "InputNumber"
	case "boolean":
		return "Switch"
	case "date", "time", "datetime":
		return detectDateTimeComponent(field.Type)
	case "json":
		return "JsonEditor" // Custom component
	case "file":
		return detectFileComponent(field)
	case "select", "multiselect":
		return detectSelectComponent(field.Type)
	case "checkbox":
		return "Checkbox"
	case "radio":
		return "Radio"
	default:
		return "Input"
	}
}

// detectStringComponent determines the appropriate component for string fields
func detectStringComponent(field *Field) string {
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
}

// detectTextComponent determines the appropriate component for text fields
func detectTextComponent(field *Field) string {
	if field.RichText != nil {
		return "TextArea"
	}
	return "TextArea"
}

// detectDateTimeComponent determines the appropriate component for date/time fields
func detectDateTimeComponent(fieldType string) string {
	switch fieldType {
	case "time":
		return "TimePicker"
	case "date", "datetime":
		return "DatePicker" // For datetime, we'll need to add showTime prop elsewhere
	default:
		return "DatePicker"
	}
}

// detectFileComponent determines the appropriate component for file fields
func detectFileComponent(field *Field) string {
	if field.File != nil && field.File.IsImage {
		return "Upload.Image"
	}
	return "Upload"
}

// detectSelectComponent determines the appropriate component for select fields
func detectSelectComponent(fieldType string) string {
	// Both "select" and "multiselect" use the Select component
	// For "multiselect", we'll need to add mode="multiple" prop elsewhere
	return "Select"
}
