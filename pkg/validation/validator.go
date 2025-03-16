package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Validation errors:\n")

	for _, err := range e {
		sb.WriteString(fmt.Sprintf("- %s\n", err.Error()))
	}

	return sb.String()
}

// Validator validates data
type Validator interface {
	Validate(data interface{}) ValidationErrors
}

// StructValidator validates struct data
type StructValidator struct {
	// Rules for validation
	Rules map[string][]Rule
}

// Rule represents a validation rule
type Rule interface {
	Validate(value interface{}) *ValidationError
}

// RequiredRule validates that a value is not empty
type RequiredRule struct {
	Message string
}

func (r RequiredRule) Validate(value interface{}) *ValidationError {
	if value == nil {
		return &ValidationError{Message: r.Message}
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		if v.String() == "" {
			return &ValidationError{Message: r.Message}
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if v.Len() == 0 {
			return &ValidationError{Message: r.Message}
		}
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return &ValidationError{Message: r.Message}
		}
	}

	return nil
}

// StringLengthRule validates string length
type StringLengthRule struct {
	Min     int
	Max     int
	Message string
}

func (r StringLengthRule) Validate(value interface{}) *ValidationError {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)

	if v.Kind() != reflect.String {
		return nil
	}

	length := len(v.String())

	if (r.Min > 0 && length < r.Min) || (r.Max > 0 && length > r.Max) {
		return &ValidationError{Message: r.Message}
	}

	return nil
}

// PatternRule validates string against a pattern
type PatternRule struct {
	Pattern string
	Message string
}

func (r PatternRule) Validate(value interface{}) *ValidationError {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)

	if v.Kind() != reflect.String {
		return nil
	}

	matched, err := regexp.MatchString(r.Pattern, v.String())
	if err != nil || !matched {
		return &ValidationError{Message: r.Message}
	}

	return nil
}

// Validate validates data against rules
func (v StructValidator) Validate(data interface{}) ValidationErrors {
	var errors ValidationErrors

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name
		fieldValue := val.Field(i).Interface()

		if rules, ok := v.Rules[fieldName]; ok {
			for _, rule := range rules {
				if err := rule.Validate(fieldValue); err != nil {
					err.Field = fieldName
					errors = append(errors, *err)
				}
			}
		}
	}

	return errors
}

// NewStructValidator creates a new struct validator
func NewStructValidator() *StructValidator {
	return &StructValidator{
		Rules: make(map[string][]Rule),
	}
}

// AddRule adds a rule for a field
func (v *StructValidator) AddRule(field string, rule Rule) *StructValidator {
	v.Rules[field] = append(v.Rules[field], rule)
	return v
}

// Required adds a required rule for a field
func (v *StructValidator) Required(field string, message string) *StructValidator {
	return v.AddRule(field, RequiredRule{Message: message})
}

// StringLength adds a string length rule for a field
func (v *StructValidator) StringLength(field string, min, max int, message string) *StructValidator {
	return v.AddRule(field, StringLengthRule{Min: min, Max: max, Message: message})
}

// Pattern adds a pattern rule for a field
func (v *StructValidator) Pattern(field string, pattern, message string) *StructValidator {
	return v.AddRule(field, PatternRule{Pattern: pattern, Message: message})
}
