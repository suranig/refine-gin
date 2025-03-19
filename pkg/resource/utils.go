package resource

import (
	"errors"
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
