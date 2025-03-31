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

	// Returns field lists for UI
	GetFilterableFields() []string
	GetSortableFields() []string
	GetTableFields() []string
	GetFormFields() []string
	GetRequiredFields() []string
	GetEditableFields() []string

	// Permissions related methods
	GetPermissions() map[string][]string
	HasPermission(operation string, role string) bool

	// Returns form layout configuration
	GetFormLayout() *FormLayout
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
	IDFieldName string              // Nazwa pola identyfikatora (domyślnie "ID")
	Permissions map[string][]string // Map of operations to roles with permission

	// Field lists for different purposes
	FilterableFields []string
	SearchableFields []string
	SortableFields   []string
	TableFields      []string
	FormFields       []string
	RequiredFields   []string
	UniqueFields     []string
	EditableFields   []string // Fields that can be edited
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
	IDFieldName string              // Nazwa pola identyfikatora (domyślnie "ID")
	Permissions map[string][]string // Map of operations to roles with permission

	// Field lists for different purposes
	FilterableFields []string
	SearchableFields []string
	SortableFields   []string
	TableFields      []string
	FormFields       []string
	RequiredFields   []string
	UniqueFields     []string
	EditableFields   []string // Fields that can be edited

	// Form layout configuration
	FormLayout *FormLayout
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

	// Editable fields (all fields that are not marked as readonly)
	editableFields := config.EditableFields
	if len(editableFields) == 0 {
		for _, f := range fields {
			if !f.ReadOnly && f.Name != "ID" && f.Name != "id" {
				editableFields = append(editableFields, f.Name)
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
		Permissions: config.Permissions,

		// Field lists
		FilterableFields: filterableFields,
		SearchableFields: config.SearchableFields,
		SortableFields:   sortableFields,
		TableFields:      tableFields,
		FormFields:       formFields,
		RequiredFields:   requiredFields,
		UniqueFields:     config.UniqueFields,
		EditableFields:   editableFields,
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

		// Check for JSON or object field types
		if isJsonField(field.Type) {
			fieldDef.Type = "json"
			fieldDef.Json = &JsonConfig{
				DefaultExpanded: true,
				EditorType:      "form", // Default to form editor
			}

			// Extract JSON properties from struct tags if available
			jsonSchema, jsonProps := extractJsonSchemaAndProperties(field.Type)
			if jsonSchema != nil {
				fieldDef.Json.Schema = jsonSchema
			}
			if len(jsonProps) > 0 {
				fieldDef.Json.Properties = jsonProps
			}
		}

		if tag, ok := field.Tag.Lookup("refine"); ok {
			ParseFieldTag(&fieldDef, tag)
		}

		// Look for json tags to enhance the field definition
		if tag, ok := field.Tag.Lookup("json"); ok {
			ProcessJsonTag(&fieldDef, tag)
		}

		fields = append(fields, fieldDef)
	}

	return fields
}

// isJsonField checks if a field type is a JSON type (map, struct with json.RawMessage, etc.)
func isJsonField(t reflect.Type) bool {
	// Check direct json.RawMessage type
	if t.String() == "json.RawMessage" {
		return true
	}

	// Check for map with string keys
	if t.Kind() == reflect.Map && t.Key().Kind() == reflect.String {
		return true
	}

	// Check for struct with exported fields that have json tags
	if t.Kind() == reflect.Struct {
		hasJsonFields := false

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath == "" && field.Tag.Get("json") != "" {
				hasJsonFields = true
				break
			}
		}

		return hasJsonFields
	}

	return false
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

		// Handle readOnly and hidden tags
		if part == "readOnly" {
			field.ReadOnly = true
			continue
		}

		if part == "hidden" {
			field.Hidden = true
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

// extractJsonSchemaAndProperties extracts JSON schema and properties from a struct type
func extractJsonSchemaAndProperties(t reflect.Type) (map[string]interface{}, []JsonProperty) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Initialize schema
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}

	// If it's not a struct, return empty schema
	if t.Kind() != reflect.Struct {
		return schema, nil
	}

	var properties []JsonProperty

	// Extract properties from struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Get JSON field name from tag or use field name
		jsonName := field.Name
		if tag, ok := field.Tag.Lookup("json"); ok {
			parts := strings.Split(tag, ",")
			if parts[0] != "" && parts[0] != "-" {
				jsonName = parts[0]
			}
		}

		// Create property
		prop := JsonProperty{
			Path:  jsonName,
			Label: field.Name,
			Type:  getJsonPropertyType(field.Type),
		}

		// Add validation if available
		if validationTag, ok := field.Tag.Lookup("validate"); ok {
			prop.Validation = parseValidationTag(validationTag)
		}

		// For nested objects, recursively extract properties
		if isJsonField(field.Type) {
			_, nestedProps := extractJsonSchemaAndProperties(field.Type)

			// Set properties with prefixed paths
			for _, nestedProp := range nestedProps {
				nestedPath := jsonName + "." + nestedProp.Path
				nestedProp.Path = nestedPath
				properties = append(properties, nestedProp)
			}
		}

		properties = append(properties, prop)

		// Add to schema
		propSchema := map[string]interface{}{
			"type": prop.Type,
		}
		if prop.Validation != nil {
			if prop.Validation.Required {
				propSchema["required"] = true
			}
		}

		schema["properties"].(map[string]interface{})[jsonName] = propSchema
	}

	return schema, properties
}

// getJsonPropertyType returns the JSON type for a Go type
func getJsonPropertyType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Struct:
		if t.String() == "time.Time" {
			return "string" // Treat time as string
		}
		return "object"
	case reflect.Map:
		return "object"
	case reflect.Slice, reflect.Array:
		return "array"
	default:
		return "string" // Default to string
	}
}

// parseValidationTag parses validation tag into JsonValidation struct
func parseValidationTag(tag string) *JsonValidation {
	validation := &JsonValidation{}
	parts := strings.Split(tag, ",")

	for _, part := range parts {
		switch {
		case part == "required":
			validation.Required = true
		case strings.HasPrefix(part, "min="):
			value, err := strconv.Atoi(part[4:])
			if err == nil {
				validation.Min = float64(value)
			}
		case strings.HasPrefix(part, "max="):
			value, err := strconv.Atoi(part[4:])
			if err == nil {
				validation.Max = float64(value)
			}
		case strings.HasPrefix(part, "minlength="):
			value, err := strconv.Atoi(part[10:])
			if err == nil {
				validation.MinLength = value
			}
		case strings.HasPrefix(part, "maxlength="):
			value, err := strconv.Atoi(part[10:])
			if err == nil {
				validation.MaxLength = value
			}
		}
	}

	return validation
}

// ProcessJsonTag processes a json tag for field configuration
func ProcessJsonTag(field *Field, tag string) {
	// Check for omitempty or - (ignore field)
	if strings.Contains(tag, "omitempty") {
		// If field has validation, ensure it's not required
		if field.Validation != nil {
			field.Validation.Required = false
		}
	}

	// Extract the field name from json tag
	parts := strings.Split(tag, ",")
	if parts[0] != "" && parts[0] != "-" {
		// Store the original name as label if not already set
		if field.Label == field.Name {
			field.Label = field.Name
		}
		// Use json field name as the actual name
		field.Name = parts[0]
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

// Add implementations for DefaultResource
func (r *DefaultResource) GetFilterableFields() []string {
	return r.FilterableFields
}

func (r *DefaultResource) GetSortableFields() []string {
	return r.SortableFields
}

func (r *DefaultResource) GetTableFields() []string {
	return r.TableFields
}

func (r *DefaultResource) GetFormFields() []string {
	return r.FormFields
}

func (r *DefaultResource) GetRequiredFields() []string {
	return r.RequiredFields
}

func (r *DefaultResource) GetEditableFields() []string {
	return r.EditableFields
}

func (r *DefaultResource) GetPermissions() map[string][]string {
	return r.Permissions
}

func (r *DefaultResource) HasPermission(operation string, role string) bool {
	if r.Permissions == nil {
		return true // If no permissions are defined, allow access by default
	}

	roles, exists := r.Permissions[operation]
	if !exists {
		return true // If the operation doesn't have defined permissions, allow access
	}

	// Check if the role exists in the list of allowed roles
	for _, r := range roles {
		if r == role {
			return true
		}
	}

	return false
}

// GetFormLayout returns the form layout configuration
func (r *DefaultResource) GetFormLayout() *FormLayout {
	return r.FormLayout
}
