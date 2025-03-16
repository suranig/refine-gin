package resource

import (
	"fmt"
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

	GetRelations() []Relation
	HasRelation(name string) bool
	GetRelation(name string) *Relation
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
	Relations   []Relation
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
	Relations   []Relation
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

	// Extract relations from model if not provided
	relations := config.Relations
	if len(relations) == 0 {
		relations = ExtractRelationsFromModel(config.Model)
	}

	return &DefaultResource{
		Name:        config.Name,
		Model:       config.Model,
		Fields:      fields,
		Operations:  config.Operations,
		DefaultSort: config.DefaultSort,
		Filters:     config.Filters,
		Middlewares: config.Middlewares,
		Relations:   relations,
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

func SetID(obj interface{}, id interface{}) error {
	val := reflect.ValueOf(obj)

	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("object must be a pointer")
	}

	val = val.Elem()

	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return fmt.Errorf("ID field does not exist")
	}

	if !idField.CanSet() {
		return fmt.Errorf("ID field cannot be set")
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
			return fmt.Errorf("cannot convert ID to type %s", idField.Type())
		}
	} else {
		idField.Set(idValue)
	}

	return nil
}
