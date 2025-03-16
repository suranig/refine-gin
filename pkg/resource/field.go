package resource

// Field reprezentuje pole zasobu
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

// Validator reprezentuje walidator pola
type Validator interface {
	Validate(value interface{}) error
}

// StringValidator waliduje wartość string
type StringValidator struct {
	MinLength int
	MaxLength int
	Pattern   string
}

func (v StringValidator) Validate(value interface{}) error {
	// TODO: Implementacja walidacji
	return nil
}

// NumberValidator waliduje wartość liczbową
type NumberValidator struct {
	Min float64
	Max float64
}

func (v NumberValidator) Validate(value interface{}) error {
	// TODO: Implementacja walidacji
	return nil
}

// Filter reprezentuje filtr dla pola
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// Sort reprezentuje sortowanie po polu
type Sort struct {
	Field string
	Order string // "asc" lub "desc"
}
