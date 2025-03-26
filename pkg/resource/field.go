package resource

import (
	"fmt"
	"regexp"
	"strconv"
)

// Field represents a resource field
type Field struct {
	Name       string
	Type       string
	Label      string
	Validation *Validation
	Options    []Option
	Relation   *RelationConfig
	List       *ListConfig
	Form       *FormConfig
	Validators []Validator
}

// Validation defines field validation rules
type Validation struct {
	Required  bool
	Min       float64
	Max       float64
	MinLength int
	MaxLength int
	Pattern   string
	Message   string
}

// RelationConfig defines field relation configuration
type RelationConfig struct {
	Resource     string
	Type         string
	ValueField   string
	DisplayField string
	FetchMode    string
	Endpoint     string
	IDField      string
	Required     bool
	AllowNone    bool
	MinItems     int
	MaxItems     int
	Searchable   bool
	Async        bool
	Placeholder  string
}

// Option represents a field option for select/enum fields
type Option struct {
	Value interface{}
	Label string
}

// ListConfig defines field configuration for list view
type ListConfig struct {
	Width    int
	Fixed    string
	Ellipsis bool
}

// FormConfig defines field configuration for form view
type FormConfig struct {
	Placeholder string
	Help        string
	Tooltip     string
}

// Validator represents a field validator
type Validator interface {
	Validate(value interface{}) error
}

// StringValidator validates string values
type StringValidator struct {
	MinLength int
	MaxLength int
	Pattern   string
}

func (v StringValidator) Validate(value interface{}) error {
	// Convert value to string
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	// Check min length
	if v.MinLength > 0 && len(str) < v.MinLength {
		return fmt.Errorf("string length must be at least %d characters", v.MinLength)
	}

	// Check max length
	if v.MaxLength > 0 && len(str) > v.MaxLength {
		return fmt.Errorf("string length must not exceed %d characters", v.MaxLength)
	}

	// Check pattern
	if v.Pattern != "" {
		matched, err := regexp.MatchString(v.Pattern, str)
		if err != nil {
			return fmt.Errorf("invalid pattern: %v", err)
		}
		if !matched {
			return fmt.Errorf("string does not match pattern: %s", v.Pattern)
		}
	}

	return nil
}

// NumberValidator validates numeric values
type NumberValidator struct {
	Min float64
	Max float64
}

func (v NumberValidator) Validate(value interface{}) error {
	// Convert value to float64
	var num float64
	var err error

	switch val := value.(type) {
	case float64:
		num = val
	case float32:
		num = float64(val)
	case int:
		num = float64(val)
	case int8:
		num = float64(val)
	case int16:
		num = float64(val)
	case int32:
		num = float64(val)
	case int64:
		num = float64(val)
	case uint:
		num = float64(val)
	case uint8:
		num = float64(val)
	case uint16:
		num = float64(val)
	case uint32:
		num = float64(val)
	case uint64:
		num = float64(val)
	case string:
		num, err = strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("cannot convert string to number: %v", err)
		}
	default:
		return fmt.Errorf("value must be a number or a string that can be converted to a number")
	}

	// Check min value
	if v.Min != 0 && num < v.Min {
		return fmt.Errorf("number must be at least %v", v.Min)
	}

	// Check max value
	if v.Max != 0 && num > v.Max {
		return fmt.Errorf("number must not exceed %v", v.Max)
	}

	return nil
}

// Filter defines a filter configuration
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// Sort defines a sort configuration
type Sort struct {
	Field string
	Order string
}
