package resource

// Field represents a resource field
type Field struct {
	Name       string
	Type       string
	Filterable bool
	Sortable   bool
	Searchable bool
	Required   bool
	Unique     bool
	Validators []Validator
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
	// TODO: Implement validation
	return nil
}

// NumberValidator validates numeric values
type NumberValidator struct {
	Min float64
	Max float64
}

func (v NumberValidator) Validate(value interface{}) error {
	// TODO: Implement validation
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
