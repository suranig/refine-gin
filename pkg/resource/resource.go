package resource

import (
	"reflect"
	"strconv"
	"strings"
)

// Resource defines the interface for API resources
type Resource interface {
	GetName() string

	GetModel() interface{}

	GetFields() []Field

	GetOperations() []Operation

	HasOperation(op Operation) bool

	GetDefaultSort() *Sort

	GetFilters() []Filter

	GetMiddlewares() []interface{}
}

// ResourceConfig contains configuration for creating a resource
type ResourceConfig struct {
	Name        string
	Model       interface{}
	Fields      []Field
	Operations  []Operation
	DefaultSort *Sort
	Filters     []Filter
	Middlewares []interface{}
}

// DefaultResource implements the Resource interface
type DefaultResource struct {
	Name        string
	Model       interface{}
	Fields      []Field
	Operations  []Operation
	DefaultSort *Sort
	Filters     []Filter
	Middlewares []interface{}
}

func (r *DefaultResource) GetName() string {
	return r.Name
}

func (r *DefaultResource) GetModel() interface{} {
	return r.Model
}

func (r *DefaultResource) GetFields() []Field {
	return r.Fields
}

func (r *DefaultResource) GetOperations() []Operation {
	return r.Operations
}

func (r *DefaultResource) HasOperation(op Operation) bool {
	for _, operation := range r.Operations {
		if operation == op {
			return true
		}
	}
	return false
}

func (r *DefaultResource) GetDefaultSort() *Sort {
	return r.DefaultSort
}

func (r *DefaultResource) GetFilters() []Filter {
	return r.Filters
}

func (r *DefaultResource) GetMiddlewares() []interface{} {
	return r.Middlewares
}

// NewResource creates a new resource from configuration
func NewResource(config ResourceConfig) Resource {
	// Extract fields from model if not provided
	fields := config.Fields
	if len(fields) == 0 {
		fields = GenerateFieldsFromModel(config.Model)
	}

	return &DefaultResource{
		Name:        config.Name,
		Model:       config.Model,
		Fields:      fields,
		Operations:  config.Operations,
		DefaultSort: config.DefaultSort,
		Filters:     config.Filters,
		Middlewares: config.Middlewares,
	}
}

// GenerateFieldsFromModel generates field definitions based on the model
func GenerateFieldsFromModel(model interface{}) []Field {
	var fields []Field

	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		if field.PkgPath != "" {
			continue
		}

		fieldDef := Field{
			Name:       field.Name,
			Type:       field.Type.String(),
			Filterable: true,  // By default, all fields are filterable
			Sortable:   true,  // By default, all fields are sortable
			Searchable: false, // By default, fields are not searchable
		}

		if tag, ok := field.Tag.Lookup("refine"); ok {
			ParseFieldTag(&fieldDef, tag)
		}

		fields = append(fields, fieldDef)
	}

	return fields
}

// ParseFieldTag parses the field tag and updates the field definition
func ParseFieldTag(field *Field, tag string) {
	parts := strings.Split(tag, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch part {
		case "filterable":
			field.Filterable = true
		case "!filterable":
			field.Filterable = false
		case "sortable":
			field.Sortable = true
		case "!sortable":
			field.Sortable = false
		case "searchable":
			field.Searchable = true
		case "!searchable":
			field.Searchable = false
		case "required":
			field.Required = true
		case "!required":
			field.Required = false
		case "unique":
			field.Unique = true
		case "!unique":
			field.Unique = false
		}

		if strings.HasPrefix(part, "min=") {
			if value, err := strconv.Atoi(part[4:]); err == nil {
				if field.Type == "string" {
					field.Validators = append(field.Validators, StringValidator{MinLength: value})
				} else if strings.HasPrefix(field.Type, "int") || strings.HasPrefix(field.Type, "float") {
					field.Validators = append(field.Validators, NumberValidator{Min: float64(value)})
				}
			}
		} else if strings.HasPrefix(part, "max=") {
			if value, err := strconv.Atoi(part[4:]); err == nil {
				if field.Type == "string" {
					field.Validators = append(field.Validators, StringValidator{MaxLength: value})
				} else if strings.HasPrefix(field.Type, "int") || strings.HasPrefix(field.Type, "float") {
					field.Validators = append(field.Validators, NumberValidator{Max: float64(value)})
				}
			}
		} else if strings.HasPrefix(part, "pattern=") {
			pattern := part[8:]
			field.Validators = append(field.Validators, StringValidator{Pattern: pattern})
		}
	}
}

func ExtractFieldsFromModel(model interface{}) []Field {
	return GenerateFieldsFromModel(model)
}
