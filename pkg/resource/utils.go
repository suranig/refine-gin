package resource

import (
	"errors"
	"fmt"
	"reflect"
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
