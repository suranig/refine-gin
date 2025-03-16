package resource

import (
	"reflect"
)

// Resource reprezentuje zasób API
type Resource interface {
	// GetName zwraca nazwę zasobu
	GetName() string

	// GetModel zwraca model zasobu
	GetModel() interface{}

	// GetFields zwraca definicje pól zasobu
	GetFields() []Field

	// GetOperations zwraca dozwolone operacje na zasobie
	GetOperations() []Operation

	// HasOperation sprawdza, czy zasób obsługuje daną operację
	HasOperation(op Operation) bool

	// GetDefaultSort zwraca domyślne sortowanie
	GetDefaultSort() *Sort

	// GetFilters zwraca dozwolone filtry
	GetFilters() []Filter

	// GetMiddlewares zwraca middleware dla zasobu
	GetMiddlewares() []interface{} // gin.HandlerFunc
}

// ResourceConfig zawiera konfigurację zasobu
type ResourceConfig struct {
	Name        string
	Model       interface{}
	Fields      []Field
	Operations  []Operation
	DefaultSort *Sort
	Filters     []Filter
	Middlewares []interface{} // gin.HandlerFunc
}

// resource implementuje interfejs Resource
type resource struct {
	name        string
	model       interface{}
	fields      []Field
	operations  []Operation
	defaultSort *Sort
	filters     []Filter
	middlewares []interface{}
}

// NewResource tworzy nowy zasób
func NewResource(config ResourceConfig) Resource {
	// Walidacja konfiguracji
	if config.Name == "" {
		panic("Resource name cannot be empty")
	}

	if config.Model == nil {
		panic("Resource model cannot be nil")
	}

	// Jeśli nie podano pól, wygeneruj je automatycznie z modelu
	fields := config.Fields
	if len(fields) == 0 {
		fields = generateFieldsFromModel(config.Model)
	}

	return &resource{
		name:        config.Name,
		model:       config.Model,
		fields:      fields,
		operations:  config.Operations,
		defaultSort: config.DefaultSort,
		filters:     config.Filters,
		middlewares: config.Middlewares,
	}
}

func (r *resource) GetName() string {
	return r.name
}

func (r *resource) GetModel() interface{} {
	return r.model
}

func (r *resource) GetFields() []Field {
	return r.fields
}

func (r *resource) GetOperations() []Operation {
	return r.operations
}

func (r *resource) HasOperation(op Operation) bool {
	for _, operation := range r.operations {
		if operation == op {
			return true
		}
	}
	return false
}

func (r *resource) GetDefaultSort() *Sort {
	return r.defaultSort
}

func (r *resource) GetFilters() []Filter {
	return r.filters
}

func (r *resource) GetMiddlewares() []interface{} {
	return r.middlewares
}

// generateFieldsFromModel generuje definicje pól na podstawie modelu
func generateFieldsFromModel(model interface{}) []Field {
	var fields []Field

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// Pomijaj pola nieeksportowane
		if field.PkgPath != "" {
			continue
		}

		// Tworzenie definicji pola
		fieldDef := Field{
			Name:       field.Name,
			Type:       field.Type.String(),
			Filterable: true,  // Domyślnie wszystkie pola są filtrowalne
			Sortable:   true,  // Domyślnie wszystkie pola są sortowalne
			Searchable: false, // Domyślnie pola nie są przeszukiwalne
		}

		// Sprawdzanie tagów
		if tag, ok := field.Tag.Lookup("refine"); ok {
			parseFieldTag(&fieldDef, tag)
		}

		fields = append(fields, fieldDef)
	}

	return fields
}

// parseFieldTag parsuje tag pola
func parseFieldTag(field *Field, tag string) {
	// TODO: Implementacja parsowania tagów
}
