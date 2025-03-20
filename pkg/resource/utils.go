package resource

import (
	"errors"
	"fmt"
	"reflect"
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
