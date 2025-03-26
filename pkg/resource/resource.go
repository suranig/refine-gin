package resource

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/suranig/refine-gin/pkg/utils"
)

// Resource defines the interface for API resources
type Resource interface {
	GetName() string

	// Get display label for the resource
	GetLabel() string

	// Get icon for the resource
	GetIcon() string

	GetModel() interface{}

	GetFields() []Field

	GetOperations() []Operation

	HasOperation(op Operation) bool

	GetDefaultSort() *Sort

	GetFilters() []Filter

	GetMiddlewares() []interface{}

	GetRelations() []Relation
	HasRelation(name string) bool
	GetRelation(name string) *Relation

	// Zwraca nazwę pola identyfikatora (domyślnie "ID")
	GetIDFieldName() string

	// Returns a field by name
	GetField(name string) *Field

	// Returns searchable fields
	GetSearchable() []string
}

// ResourceConfig contains configuration for creating a resource
type ResourceConfig struct {
	Name        string
	Label       string
	Icon        string
	Model       interface{}
	Fields      []Field
	Operations  []Operation
	DefaultSort *Sort
	Filters     []Filter
	Middlewares []interface{}
	Relations   []Relation
	IDFieldName string // Nazwa pola identyfikatora (domyślnie "ID")
}

// DefaultResource implements the Resource interface
type DefaultResource struct {
	Name        string
	Label       string
	Icon        string
	Model       interface{}
	Fields      []Field
	Operations  []Operation
	DefaultSort *Sort
	Filters     []Filter
	Middlewares []interface{}
	Relations   []Relation
	IDFieldName string // Nazwa pola identyfikatora (domyślnie "ID")
}

func (r *DefaultResource) GetName() string {
	return r.Name
}

func (r *DefaultResource) GetLabel() string {
	if r.Label == "" {
		// If no label is provided, use capitalized name as default
		return strings.Title(r.Name)
	}
	return r.Label
}

func (r *DefaultResource) GetIcon() string {
	return r.Icon
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

func (r *DefaultResource) GetIDFieldName() string {
	if r.IDFieldName == "" {
		return "ID" // Domyślna nazwa pola identyfikatora
	}
	return r.IDFieldName
}

// NewResource creates a new resource from configuration
func NewResource(config ResourceConfig) Resource {
	// Extract fields from model if not provided
	fields := config.Fields
	if len(fields) == 0 {
		fields = GenerateFieldsFromModel(config.Model)
	}

	// Extract relations from model if not provided
	relations := config.Relations
	if len(relations) == 0 {
		relations = ExtractRelationsFromModel(config.Model)
	}

	// Ustaw domyślną nazwę pola identyfikatora, jeśli nie podano
	idFieldName := config.IDFieldName
	if idFieldName == "" {
		idFieldName = "ID"
	}

	// Set default label if not provided
	label := config.Label
	if label == "" {
		label = strings.Title(config.Name)
	}

	return &DefaultResource{
		Name:        config.Name,
		Label:       label,
		Icon:        config.Icon,
		Model:       config.Model,
		Fields:      fields,
		Operations:  config.Operations,
		DefaultSort: config.DefaultSort,
		Filters:     config.Filters,
		Middlewares: config.Middlewares,
		Relations:   relations,
		IDFieldName: idFieldName,
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

func (r *DefaultResource) GetRelations() []Relation {
	return r.Relations
}

func (r *DefaultResource) HasRelation(name string) bool {
	for _, relation := range r.Relations {
		if relation.Name == name {
			return true
		}
	}
	return false
}

func (r *DefaultResource) GetRelation(name string) *Relation {
	for _, relation := range r.Relations {
		if relation.Name == name {
			return &relation
		}
	}
	return nil
}

// CreateSliceOfType tworzy pustą tablicę elementów tego samego typu co model
func CreateSliceOfType(model interface{}) interface{} {
	modelType := reflect.TypeOf(model)

	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	sliceType := reflect.SliceOf(modelType)
	slice := reflect.MakeSlice(sliceType, 0, 0)

	slicePtr := reflect.New(sliceType)
	slicePtr.Elem().Set(slice)

	return slicePtr.Interface()
}

// CreateInstanceOfType tworzy nowy element tego samego typu co model
func CreateInstanceOfType(model interface{}) interface{} {
	modelType := reflect.TypeOf(model)

	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	instance := reflect.New(modelType)

	return instance.Interface()
}

// SetID sets the ID field of an object
func SetID(obj interface{}, id interface{}) error {
	return utils.SetID(obj, id, "ID")
}

// SetCustomID sets a custom ID field of an object
func SetCustomID(obj interface{}, id interface{}, idFieldName string) error {
	return utils.SetID(obj, id, idFieldName)
}

// GetField returns a field by name
func (r *DefaultResource) GetField(name string) *Field {
	for _, field := range r.Fields {
		if field.Name == name {
			return &field
		}
	}
	return nil
}

// GetSearchable returns a list of searchable field names
func (r *DefaultResource) GetSearchable() []string {
	var searchable []string
	for _, field := range r.Fields {
		if field.Searchable {
			searchable = append(searchable, field.Name)
		}
	}
	return searchable
}
