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

	// Field lists for different purposes
	FilterableFields []string
	SearchableFields []string
	SortableFields   []string
	TableFields      []string
	FormFields       []string
	RequiredFields   []string
	UniqueFields     []string
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

	// Field lists for different purposes
	FilterableFields []string
	SearchableFields []string
	SortableFields   []string
	TableFields      []string
	FormFields       []string
	RequiredFields   []string
	UniqueFields     []string
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

	// Set default label if not provided
	label := config.Label
	if label == "" {
		label = strings.Title(config.Name)
	}

	// Set default field lists if not provided
	filterableFields := config.FilterableFields
	if len(filterableFields) == 0 {
		// By default, all fields are filterable
		for _, f := range fields {
			filterableFields = append(filterableFields, f.Name)
		}
	}

	sortableFields := config.SortableFields
	if len(sortableFields) == 0 {
		// By default, all fields are sortable
		for _, f := range fields {
			sortableFields = append(sortableFields, f.Name)
		}
	}

	// Default table fields are all fields
	tableFields := config.TableFields
	if len(tableFields) == 0 {
		for _, f := range fields {
			tableFields = append(tableFields, f.Name)
		}
	}

	// Default form fields are all fields except ID
	formFields := config.FormFields
	if len(formFields) == 0 {
		for _, f := range fields {
			if f.Name != "ID" && f.Name != "id" {
				formFields = append(formFields, f.Name)
			}
		}
	}

	// Required fields from validation
	requiredFields := config.RequiredFields
	if len(requiredFields) == 0 {
		for _, f := range fields {
			if f.Validation != nil && f.Validation.Required {
				requiredFields = append(requiredFields, f.Name)
			}
		}
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
		IDFieldName: config.IDFieldName,

		// Field lists
		FilterableFields: filterableFields,
		SearchableFields: config.SearchableFields,
		SortableFields:   sortableFields,
		TableFields:      tableFields,
		FormFields:       formFields,
		RequiredFields:   requiredFields,
		UniqueFields:     config.UniqueFields,
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
			Name:  field.Name,
			Type:  field.Type.String(),
			Label: field.Name, // Default label is the field name
			List: &ListConfig{
				Width:    200, // Default width
				Ellipsis: true,
			},
			Form: &FormConfig{
				Placeholder: "Enter " + strings.ToLower(field.Name),
			},
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

		if strings.HasPrefix(part, "label=") {
			field.Label = part[6:]
			continue
		}

		if strings.HasPrefix(part, "placeholder=") {
			if field.Form == nil {
				field.Form = &FormConfig{}
			}
			field.Form.Placeholder = part[12:]
			continue
		}

		if strings.HasPrefix(part, "help=") {
			if field.Form == nil {
				field.Form = &FormConfig{}
			}
			field.Form.Help = part[5:]
			continue
		}

		if strings.HasPrefix(part, "tooltip=") {
			if field.Form == nil {
				field.Form = &FormConfig{}
			}
			field.Form.Tooltip = part[8:]
			continue
		}

		if strings.HasPrefix(part, "width=") {
			if width, err := strconv.Atoi(part[6:]); err == nil {
				if field.List == nil {
					field.List = &ListConfig{}
				}
				field.List.Width = width
			}
			continue
		}

		if strings.HasPrefix(part, "fixed=") {
			if field.List == nil {
				field.List = &ListConfig{}
			}
			field.List.Fixed = part[6:]
			continue
		}

		// Validation rules
		if strings.HasPrefix(part, "min=") {
			if value, err := strconv.Atoi(part[4:]); err == nil {
				if field.Validation == nil {
					field.Validation = &Validation{}
				}
				if field.Type == "string" {
					field.Validation.MinLength = value
				} else {
					field.Validation.Min = float64(value)
				}
			}
		} else if strings.HasPrefix(part, "max=") {
			if value, err := strconv.Atoi(part[4:]); err == nil {
				if field.Validation == nil {
					field.Validation = &Validation{}
				}
				if field.Type == "string" {
					field.Validation.MaxLength = value
				} else {
					field.Validation.Max = float64(value)
				}
			}
		} else if strings.HasPrefix(part, "pattern=") {
			if field.Validation == nil {
				field.Validation = &Validation{}
			}
			field.Validation.Pattern = part[8:]
		} else if part == "required" {
			if field.Validation == nil {
				field.Validation = &Validation{}
			}
			field.Validation.Required = true
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
	return r.SearchableFields
}
