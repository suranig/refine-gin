package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

// GetFieldValue safely gets a field value from an object using reflection
func GetFieldValue(obj interface{}, fieldName string) (interface{}, error) {
	if obj == nil {
		return nil, fmt.Errorf("object cannot be nil")
	}

	// Handle map type
	if m, ok := obj.(map[string]interface{}); ok {
		if value, exists := m[fieldName]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("field %s not found in map", fieldName)
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.IsValid() {
		return nil, fmt.Errorf("object is not valid")
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s not found", fieldName)
	}

	return field.Interface(), nil
}

// SetFieldValue safely sets a field value on an object using reflection
func SetFieldValue(obj interface{}, fieldName string, value interface{}) error {
	if obj == nil {
		return fmt.Errorf("object cannot be nil")
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.IsValid() {
		return fmt.Errorf("object is not valid for setting field")
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found", fieldName)
	}

	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}

	valueVal := reflect.ValueOf(value)
	if !valueVal.IsValid() {
		return fmt.Errorf("value is not valid")
	}

	if valueVal.Type().ConvertibleTo(field.Type()) {
		field.Set(valueVal.Convert(field.Type()))
		return nil
	}

	return fmt.Errorf("value type %s cannot be converted to field type %s", valueVal.Type(), field.Type())
}

// SetID sets the ID field of an object using reflection
func SetID(obj interface{}, id interface{}, idFieldName string) error {
	if idFieldName == "" {
		idFieldName = "ID"
	}

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("object must be a pointer")
	}

	val = val.Elem()
	idField := val.FieldByName(idFieldName)
	if !idField.IsValid() {
		return fmt.Errorf("%s field does not exist", idFieldName)
	}

	if !idField.CanSet() {
		return fmt.Errorf("%s field cannot be set", idFieldName)
	}

	idValue := reflect.ValueOf(id)
	if idValue.Type() != idField.Type() {
		switch idField.Kind() {
		case reflect.String:
			idField.SetString(fmt.Sprintf("%v", id))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.ParseInt(fmt.Sprintf("%v", id), 10, 64)
			if err != nil {
				return err
			}
			idField.SetInt(intVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal, err := strconv.ParseUint(fmt.Sprintf("%v", id), 10, 64)
			if err != nil {
				return err
			}
			idField.SetUint(uintVal)
		default:
			return fmt.Errorf("cannot convert %s to type %s", idFieldName, idField.Type())
		}
	} else {
		idField.Set(idValue)
	}

	return nil
}

// IsSlice checks if the given value is a slice
func IsSlice(data interface{}) bool {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Kind() == reflect.Slice
}

// GetSliceField safely gets a slice field from an object
func GetSliceField(obj interface{}, fieldName string) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return reflect.Value{}, fmt.Errorf("field %s not found", fieldName)
	}

	if field.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf("field %s is not a slice", fieldName)
	}

	return field, nil
}

// CreateNewModelInstance creates a new instance of the given model
func CreateNewModelInstance(model interface{}) interface{} {
	if model == nil {
		return nil
	}

	modelType := reflect.TypeOf(model)

	// Handle pointer types
	if modelType.Kind() == reflect.Ptr {
		// Get the element type (*T -> T)
		elemType := modelType.Elem()
		// Create a new instance of that type
		newInstance := reflect.New(elemType).Interface()
		return newInstance
	}

	// Handle non-pointer types
	// Create a pointer to the model type
	newInstance := reflect.New(modelType).Interface()
	return newInstance
}

// CreateNewSliceOfModel creates a new slice for the given model type
func CreateNewSliceOfModel(model interface{}) interface{} {
	if model == nil {
		return nil
	}

	modelType := reflect.TypeOf(model)

	// Get the base type (handle pointer types)
	var elemType reflect.Type
	if modelType.Kind() == reflect.Ptr {
		// *T -> T
		elemType = modelType.Elem()
	} else {
		elemType = modelType
	}

	// Create a slice type of the element type
	sliceType := reflect.SliceOf(reflect.PtrTo(elemType))

	// Create a new pointer to a slice
	newSlice := reflect.New(sliceType).Interface()
	return newSlice
}
