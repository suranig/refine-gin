package resource

import (
	"strings"
)

// ResourceMetadata represents the metadata for a resource
type ResourceMetadata struct {
	// Resource name (identifier)
	Name string `json:"name"`

	// Display name for the resource
	Label string `json:"label,omitempty"`

	// Icon for the resource
	Icon string `json:"icon,omitempty"`

	// Allowed operations for this resource
	Operations []Operation `json:"operations"`

	// All fields defined for this resource
	Fields []FieldMetadata `json:"fields"`

	// Relations for this resource
	Relations []RelationMetadata `json:"relations,omitempty"`

	// Default sort configuration
	DefaultSort *Sort `json:"defaultSort,omitempty"`

	// Available filters
	Filters []Filter `json:"filters,omitempty"`

	// Searchable fields
	Searchable []string `json:"searchable,omitempty"`

	// ID field name
	IDFieldName string `json:"idFieldName,omitempty"`

	// Additional field lists for UI
	FilterableFields []string `json:"filterableFields,omitempty"`
	SortableFields   []string `json:"sortableFields,omitempty"`
	TableFields      []string `json:"tableFields,omitempty"`
	FormFields       []string `json:"formFields,omitempty"`
	RequiredFields   []string `json:"requiredFields,omitempty"`
}

// FieldMetadata represents metadata for a resource field
type FieldMetadata struct {
	// Field name
	Name string `json:"name"`

	// Field type (string, number, boolean, etc.)
	Type string `json:"type"`

	// Field label for display
	Label string `json:"label,omitempty"`

	// Whether the field is filterable
	Filterable bool `json:"filterable"`

	// Whether the field is sortable
	Sortable bool `json:"sortable"`

	// Whether the field is searchable
	Searchable bool `json:"searchable"`

	// Whether the field is required
	Required bool `json:"required"`

	// Whether the field must be unique
	Unique bool `json:"unique"`

	// Whether the field is read-only (not editable)
	ReadOnly bool `json:"readOnly"`

	// Whether the field should be hidden in UI
	Hidden bool `json:"hidden"`

	// Field validators
	Validators []ValidatorMetadata `json:"validators,omitempty"`

	// JSON configuration (for object/jsonb fields)
	Json *JsonConfigMetadata `json:"json,omitempty"`

	// File configuration (for file/image fields)
	File *FileConfigMetadata `json:"file,omitempty"`

	// Rich text configuration
	RichText *RichTextConfigMetadata `json:"richText,omitempty"`

	// Select field configuration
	Select *SelectConfigMetadata `json:"select,omitempty"`

	// Computed field configuration
	Computed *ComputedFieldConfigMetadata `json:"computed,omitempty"`

	// Ant Design specific configuration
	AntDesign *AntDesignConfigMetadata `json:"antDesign,omitempty"`
}

// ValidatorMetadata represents metadata for a field validator
type ValidatorMetadata struct {
	// Validator type (string, number, etc.)
	Type string `json:"type"`

	// Validation rules
	Rules map[string]interface{} `json:"rules,omitempty"`

	// Error message
	Message string `json:"message,omitempty"`
}

// RelationMetadata represents metadata for a resource relation
type RelationMetadata struct {
	// Relation name
	Name string `json:"name"`

	// Relation type (one-to-one, one-to-many, many-to-one, many-to-many)
	Type RelationType `json:"type"`

	// Referenced resource name
	Resource string `json:"resource"`

	// Field in the current resource that holds the relation
	Field string `json:"field,omitempty"`

	// Field in the related resource that this relation refers to
	ReferenceField string `json:"referenceField,omitempty"`

	// Whether to include this relation in responses by default
	IncludeByDefault bool `json:"includeByDefault"`

	// Display field in the related resource
	DisplayField string `json:"displayField,omitempty"`

	// Value field in the related resource
	ValueField string `json:"valueField,omitempty"`

	// Whether the relation is required
	Required bool `json:"required,omitempty"`

	// Minimum number of related items (for to-many relations)
	MinItems int `json:"minItems,omitempty"`

	// Maximum number of related items (for to-many relations)
	MaxItems int `json:"maxItems,omitempty"`

	// Pivot table for many-to-many relations
	PivotTable string `json:"pivotTable,omitempty"`

	// Pivot fields for many-to-many relations
	PivotFields map[string]string `json:"pivotFields,omitempty"`

	// Cascade settings
	Cascade bool `json:"cascade,omitempty"`

	// On delete behavior
	OnDelete string `json:"onDelete,omitempty"`

	// On update behavior
	OnUpdate string `json:"onUpdate,omitempty"`
}

// JsonConfigMetadata represents metadata for JSON fields
type JsonConfigMetadata struct {
	// Schema for JSON field validation and UI
	Schema map[string]interface{} `json:"schema,omitempty"`

	// Properties defines nested fields in the JSON structure
	Properties []JsonPropertyMetadata `json:"properties,omitempty"`

	// Default expanded state in UI
	DefaultExpanded bool `json:"defaultExpanded,omitempty"`

	// JSON path prefix for filtering nested fields
	PathPrefix string `json:"pathPrefix,omitempty"`

	// Editor type (json, form, tree)
	EditorType string `json:"editorType,omitempty"`

	// Nested indicates if the JSON structure is nested (objects within objects)
	Nested bool `json:"nested,omitempty"`

	// RenderAs defines the UI presentation style ("tabs", "form", "tree", "grid")
	RenderAs string `json:"renderAs,omitempty"`

	// TabsConfig holds configuration for rendering in tabs format
	TabsConfig *JsonTabsConfigMetadata `json:"tabsConfig,omitempty"`

	// GridConfig holds configuration for grid layout
	GridConfig *JsonGridConfigMetadata `json:"gridConfig,omitempty"`

	// ObjectLabels provides display labels for nested objects by their path
	ObjectLabels map[string]string `json:"objectLabels,omitempty"`
}

// JsonPropertyMetadata represents metadata for a JSON property
type JsonPropertyMetadata struct {
	// Property path (e.g. "config.oauth.client_id")
	Path string `json:"path,omitempty"`

	// Property label for display
	Label string `json:"label,omitempty"`

	// Property type (string, number, boolean, object, array)
	Type string `json:"type,omitempty"`

	// Whether the property is read-only (not editable)
	ReadOnly bool `json:"readOnly,omitempty"`

	// Whether the property should be hidden in UI
	Hidden bool `json:"hidden,omitempty"`

	// Additional validation for the property
	Validation *ValidationMetadata `json:"validation,omitempty"`

	// For object types, nested properties
	Properties []JsonPropertyMetadata `json:"properties,omitempty"`

	// UI configuration
	Form *FormConfigMetadata `json:"form,omitempty"`
}

// ValidationMetadata represents metadata for validation rules
type ValidationMetadata struct {
	Required  bool    `json:"required,omitempty"`
	Min       float64 `json:"min,omitempty"`
	Max       float64 `json:"max,omitempty"`
	MinLength int     `json:"minLength,omitempty"`
	MaxLength int     `json:"maxLength,omitempty"`
	Pattern   string  `json:"pattern,omitempty"`
	Message   string  `json:"message,omitempty"`
}

// FormConfigMetadata represents metadata for form configuration
type FormConfigMetadata struct {
	Placeholder string `json:"placeholder,omitempty"`
	Help        string `json:"help,omitempty"`
	Tooltip     string `json:"tooltip,omitempty"`

	// Width in percentage of the form element
	Width string `json:"width,omitempty"`

	// Dependent field condition, field name that this field depends on
	DependentOn string `json:"dependentOn,omitempty"`

	// Display condition, a JS expression to determine if field should be displayed
	Condition string `json:"condition,omitempty"`
}

// JsonTabsConfigMetadata represents metadata for tabs-based JSON rendering
type JsonTabsConfigMetadata struct {
	// Tabs collection for the JSON structure
	Tabs []JsonTabMetadata `json:"tabs,omitempty"`

	// TabPosition specifies where tabs are placed ("top", "right", "bottom", "left")
	TabPosition string `json:"tabPosition,omitempty"`

	// DefaultActiveTab specifies the key of the default active tab
	DefaultActiveTab string `json:"defaultActiveTab,omitempty"`
}

// JsonTabMetadata represents metadata for a single tab in a tabs-based JSON editor
type JsonTabMetadata struct {
	// Key uniquely identifies the tab
	Key string `json:"key"`

	// Title is the display name of the tab
	Title string `json:"title,omitempty"`

	// Icon specifies an optional icon for the tab
	Icon string `json:"icon,omitempty"`

	// Fields lists the JSON property paths included in this tab
	Fields []string `json:"fields,omitempty"`
}

// JsonGridConfigMetadata represents metadata for grid layout of JSON fields
type JsonGridConfigMetadata struct {
	// Columns defines the number of columns in the grid
	Columns int `json:"columns,omitempty"`

	// Gutter defines the space between grid items
	Gutter int `json:"gutter,omitempty"`

	// FieldLayouts maps field paths to their layout configuration
	FieldLayouts map[string]*JsonFieldLayoutMetadata `json:"fieldLayouts,omitempty"`
}

// JsonFieldLayoutMetadata represents metadata for grid position and span of a field
type JsonFieldLayoutMetadata struct {
	// Column specifies the starting column (1-based)
	Column int `json:"column,omitempty"`

	// Row specifies the starting row (1-based)
	Row int `json:"row,omitempty"`

	// ColSpan specifies how many columns the field spans
	ColSpan int `json:"colSpan,omitempty"`

	// RowSpan specifies how many rows the field spans
	RowSpan int `json:"rowSpan,omitempty"`
}

// FileConfigMetadata represents metadata for file/image fields
type FileConfigMetadata struct {
	// Allowed MIME types (e.g. "image/jpeg", "application/pdf")
	AllowedTypes []string `json:"allowedTypes,omitempty"`

	// Maximum file size in bytes
	MaxSize int64 `json:"maxSize,omitempty"`

	// Base URL for accessing the files
	BaseURL string `json:"baseURL,omitempty"`

	// Whether this is an image field (enables preview and image-specific features)
	IsImage bool `json:"isImage,omitempty"`

	// Image-specific configuration (if isImage is true)
	MaxWidth  int `json:"maxWidth,omitempty"`
	MaxHeight int `json:"maxHeight,omitempty"`

	// Generate thumbnails (if isImage is true)
	GenerateThumbnails bool `json:"generateThumbnails,omitempty"`

	// Thumbnail sizes (if generateThumbnails is true)
	ThumbnailSizes []ThumbnailSizeMetadata `json:"thumbnailSizes,omitempty"`
}

// ThumbnailSizeMetadata defines metadata for a thumbnail configuration
type ThumbnailSizeMetadata struct {
	// Name of the thumbnail (e.g. "small", "medium")
	Name string `json:"name"`

	// Width in pixels
	Width int `json:"width"`

	// Height in pixels
	Height int `json:"height"`

	// Whether to keep aspect ratio when resizing
	KeepAspectRatio bool `json:"keepAspectRatio,omitempty"`
}

// RichTextConfigMetadata represents metadata for rich text fields
type RichTextConfigMetadata struct {
	// Toolbar configuration (available buttons/features)
	Toolbar []string `json:"toolbar,omitempty"`

	// Height of the editor in pixels or CSS value
	Height string `json:"height,omitempty"`

	// Placeholder text when editor is empty
	Placeholder string `json:"placeholder,omitempty"`

	// Whether to enable image uploads
	EnableImages bool `json:"enableImages,omitempty"`

	// Max content length in characters
	MaxLength int `json:"maxLength,omitempty"`

	// Whether to show character counter
	ShowCounter bool `json:"showCounter,omitempty"`

	// Content format (HTML, Markdown, etc.)
	Format string `json:"format,omitempty"`
}

// SelectConfigMetadata represents metadata for select fields
type SelectConfigMetadata struct {
	// Whether to allow multiple selections
	Multiple bool `json:"multiple,omitempty"`

	// Whether to allow searching among options
	Searchable bool `json:"searchable,omitempty"`

	// Whether to allow creating new options on the fly
	Creatable bool `json:"creatable,omitempty"`

	// URL to fetch options dynamically from an API
	OptionsURL string `json:"optionsURL,omitempty"`

	// Field to depend on (value of this field will change available options)
	DependsOn string `json:"dependsOn,omitempty"`

	// Mapping between DependsOn field values and available options
	// Key is the value of the DependsOn field, value is a list of available options
	DependentOptions map[string][]OptionMetadata `json:"dependentOptions,omitempty"`

	// Placeholder text
	Placeholder string `json:"placeholder,omitempty"`

	// Whether to allow clearing the selection
	Clearable bool `json:"clearable,omitempty"`

	// Display mode (dropdown, radio, checkboxes, tags)
	DisplayMode string `json:"displayMode,omitempty"`
}

// OptionMetadata represents metadata for a select option
type OptionMetadata struct {
	Value interface{} `json:"value"`
	Label string      `json:"label,omitempty"`
}

// ComputedFieldConfigMetadata represents metadata for computed fields
type ComputedFieldConfigMetadata struct {
	// Fields this computed field depends on
	DependsOn []string `json:"dependsOn,omitempty"`

	// Expression to compute the value (can be JS expression for frontend, or Go template for backend)
	Expression string `json:"expression,omitempty"`

	// Whether the computation happens on the client-side
	ClientSide bool `json:"clientSide,omitempty"`

	// Format for displaying the computed value (only applies to client-side)
	Format string `json:"format,omitempty"`

	// Whether the computed value should be persisted to the database
	Persist bool `json:"persist,omitempty"`

	// Order in which fields should be computed (if there are dependencies between computed fields)
	ComputeOrder int `json:"computeOrder,omitempty"`
}

// AntDesignConfigMetadata represents metadata for Ant Design components
type AntDesignConfigMetadata struct {
	// Component type to use (Input, Select, DatePicker, etc.)
	ComponentType string `json:"componentType,omitempty"`

	// Props to pass to the component
	Props map[string]interface{} `json:"props,omitempty"`

	// Rules for form validation in Ant Design format
	Rules []AntDesignRuleMetadata `json:"rules,omitempty"`

	// FormItemProps to pass to Form.Item component
	FormItemProps map[string]interface{} `json:"formItemProps,omitempty"`

	// Dependencies for field dependencies (array of field names)
	Dependencies []string `json:"dependencies,omitempty"`
}

// AntDesignRuleMetadata defines a validation rule metadata for Ant Design Form
type AntDesignRuleMetadata struct {
	// Rule type (required, max, min, etc.)
	Type string `json:"type,omitempty"`

	// Message to display when validation fails
	Message string `json:"message,omitempty"`

	// Value for the rule (e.g. minimum length for "min" rule)
	Value interface{} `json:"value,omitempty"`

	// Whether to validate on blur
	ValidateOnBlur bool `json:"validateOnBlur,omitempty"`

	// Whether to validate on change
	ValidateOnChange bool `json:"validateOnChange,omitempty"`

	// Pattern for regexp validation
	Pattern string `json:"pattern,omitempty"`

	// When to validate (onChange, onBlur, onSubmit)
	ValidateTrigger string `json:"validateTrigger,omitempty"`
}

// GenerateResourceMetadata generates resource metadata from a resource
func GenerateResourceMetadata(res Resource) ResourceMetadata {
	metadata := ResourceMetadata{
		Name:             res.GetName(),
		Label:            res.GetLabel(),
		Icon:             res.GetIcon(),
		Operations:       res.GetOperations(),
		IDFieldName:      res.GetIDFieldName(),
		DefaultSort:      res.GetDefaultSort(),
		Filters:          res.GetFilters(),
		Searchable:       res.GetSearchable(),
		FilterableFields: res.GetFilterableFields(),
		SortableFields:   res.GetSortableFields(),
		TableFields:      res.GetTableFields(),
		FormFields:       res.GetFormFields(),
		RequiredFields:   res.GetRequiredFields(),
	}

	// Generate field metadata
	metadata.Fields = GenerateFieldsMetadata(res.GetFields())

	// Generate relation metadata
	if rels := res.GetRelations(); len(rels) > 0 {
		metadata.Relations = GenerateRelationsMetadata(rels)
	}

	return metadata
}

// GenerateFieldsMetadata generates metadata for fields
func GenerateFieldsMetadata(fields []Field) []FieldMetadata {
	result := make([]FieldMetadata, 0, len(fields))

	// Domyślnie wszystkie pola są filtrowalne i sortowalne
	for _, field := range fields {
		// Sprawdź na podstawie konfiguracji pola
		isFilterable := true  // Domyślnie filtrowalne
		isSortable := true    // Domyślnie sortowalne
		isSearchable := false // Domyślnie nie przeszukiwalne
		isRequired := false   // Domyślnie niewymagane
		isUnique := false     // Domyślnie nieunikalne

		// Jeśli pole ma Validation, użyj go dla required
		if field.Validation != nil && field.Validation.Required {
			isRequired = true
		}

		fieldMeta := FieldMetadata{
			Name:       field.Name,
			Type:       field.Type,
			Filterable: isFilterable,
			Sortable:   isSortable,
			Searchable: isSearchable,
			Required:   isRequired,
			Unique:     isUnique,
			ReadOnly:   field.ReadOnly,
			Hidden:     field.Hidden,
		}

		// Dodaj label jeśli istnieje
		if field.Label != "" {
			fieldMeta.Label = field.Label
		}

		// Add validators metadata
		if len(field.Validators) > 0 {
			fieldMeta.Validators = GenerateValidatorsMetadata(field.Validators)
		}

		// Add JSON metadata if field is a JSON type
		if field.Json != nil {
			fieldMeta.Json = GenerateJsonConfigMetadata(field.Json)
		}

		// Add File metadata if field has file configuration
		if field.File != nil {
			fieldMeta.File = GenerateFileConfigMetadata(field.File)
		}

		// Add RichText metadata if field has rich text configuration
		if field.RichText != nil {
			fieldMeta.RichText = GenerateRichTextConfigMetadata(field.RichText)
		}

		// Add Select metadata if field has select configuration
		if field.Select != nil {
			fieldMeta.Select = GenerateSelectConfigMetadata(field.Select)
		}

		// Add Computed metadata if field has computed field configuration
		if field.Computed != nil {
			fieldMeta.Computed = GenerateComputedFieldConfigMetadata(field.Computed)
		}

		// Add Ant Design metadata if field has Ant Design configuration
		if field.AntDesign != nil {
			fieldMeta.AntDesign = GenerateAntDesignConfigMetadata(field.AntDesign)
		} else {
			// Automatically generate Ant Design configuration
			antDesignConfig := generateDefaultAntDesignConfig(&field)
			if antDesignConfig != nil {
				fieldMeta.AntDesign = antDesignConfig
			}
		}

		result = append(result, fieldMeta)
	}

	return result
}

// generateDefaultAntDesignConfig creates default Ant Design configuration for a field
func generateDefaultAntDesignConfig(field *Field) *AntDesignConfigMetadata {
	if field == nil {
		return nil
	}

	// Create basic config
	config := &AntDesignConfigMetadata{
		ComponentType: AutoDetectAntDesignComponent(field),
		Props:         make(map[string]interface{}),
		FormItemProps: make(map[string]interface{}),
	}

	// Map validation rules
	if field.Validation != nil {
		antDesignRules := MapValidationToAntDesignRules(field.Validation)
		if len(antDesignRules) > 0 {
			config.Rules = make([]AntDesignRuleMetadata, 0, len(antDesignRules))
			for _, rule := range antDesignRules {
				config.Rules = append(config.Rules, AntDesignRuleMetadata{
					Type:             rule.Type,
					Message:          rule.Message,
					Value:            rule.Value,
					ValidateOnBlur:   rule.ValidateOnBlur,
					ValidateOnChange: rule.ValidateOnChange,
					Pattern:          rule.Pattern,
					ValidateTrigger:  rule.ValidateTrigger,
				})
			}
		}
	}

	// Add field-specific props
	switch field.Type {
	case "string":
		if field.Form != nil && field.Form.Placeholder != "" {
			config.Props["placeholder"] = field.Form.Placeholder
		}

		// Detect password field
		if strings.Contains(strings.ToLower(field.Name), "password") {
			config.ComponentType = "Password"
		}

	case "number", "integer", "float", "double", "decimal":
		config.ComponentType = "InputNumber"
		if field.Validation != nil {
			if field.Validation.Min != 0 {
				config.Props["min"] = field.Validation.Min
			}
			if field.Validation.Max != 0 {
				config.Props["max"] = field.Validation.Max
			}
		}

	case "boolean":
		config.ComponentType = "Switch"
		config.FormItemProps["valuePropName"] = "checked"

	case "date":
		config.ComponentType = "DatePicker"
		if field.Form != nil && field.Form.Placeholder != "" {
			config.Props["placeholder"] = field.Form.Placeholder
		}

	case "datetime":
		config.ComponentType = "DatePicker"
		config.Props["showTime"] = true
		if field.Form != nil && field.Form.Placeholder != "" {
			config.Props["placeholder"] = field.Form.Placeholder
		}

	case "select", "multiselect":
		config.ComponentType = "Select"
		if field.Type == "multiselect" || (field.Select != nil && field.Select.Multiple) {
			config.Props["mode"] = "multiple"
		}

		// Add options if available
		if len(field.Options) > 0 {
			options := make([]map[string]interface{}, 0, len(field.Options))
			for _, opt := range field.Options {
				options = append(options, map[string]interface{}{
					"value": opt.Value,
					"label": opt.Label,
				})
			}
			config.Props["options"] = options
		}

		// Add placeholder if available
		if field.Form != nil && field.Form.Placeholder != "" {
			config.Props["placeholder"] = field.Form.Placeholder
		} else if field.Select != nil && field.Select.Placeholder != "" {
			config.Props["placeholder"] = field.Select.Placeholder
		}

	case "file":
		if field.File != nil && field.File.IsImage {
			config.ComponentType = "Upload.Image"
			config.Props["listType"] = "picture-card"
		} else {
			config.ComponentType = "Upload"
			config.Props["listType"] = "text"
		}

		// Add FormItemProps for Upload components
		config.FormItemProps["valuePropName"] = "fileList"
		config.FormItemProps["getValueFromEvent"] = "normFile"
	}

	// Add disabled state for read-only fields
	if field.ReadOnly {
		config.Props["disabled"] = true
	}

	// Return the configuration
	return config
}

// GenerateValidatorsMetadata generates metadata for validators
func GenerateValidatorsMetadata(validators []Validator) []ValidatorMetadata {
	result := make([]ValidatorMetadata, 0, len(validators))

	for _, validator := range validators {
		var validatorMeta ValidatorMetadata

		switch v := validator.(type) {
		case StringValidator:
			validatorMeta = ValidatorMetadata{
				Type:  "string",
				Rules: map[string]interface{}{},
			}

			if v.MinLength > 0 {
				validatorMeta.Rules["minLength"] = v.MinLength
			}

			if v.MaxLength > 0 {
				validatorMeta.Rules["maxLength"] = v.MaxLength
			}

			if v.Pattern != "" {
				validatorMeta.Rules["pattern"] = v.Pattern
			}

		case NumberValidator:
			validatorMeta = ValidatorMetadata{
				Type:  "number",
				Rules: map[string]interface{}{},
			}

			if v.Min != 0 {
				validatorMeta.Rules["min"] = v.Min
			}

			if v.Max != 0 {
				validatorMeta.Rules["max"] = v.Max
			}
		}

		result = append(result, validatorMeta)
	}

	return result
}

// GenerateRelationsMetadata generates metadata for relations
func GenerateRelationsMetadata(relations []Relation) []RelationMetadata {
	result := make([]RelationMetadata, 0, len(relations))

	for _, relation := range relations {
		relationMeta := RelationMetadata{
			Name:             relation.Name,
			Type:             relation.Type,
			Resource:         relation.Resource,
			Field:            relation.Field,
			ReferenceField:   relation.ReferenceField,
			IncludeByDefault: relation.IncludeByDefault,
			Required:         relation.Required,
			MinItems:         relation.MinItems,
			MaxItems:         relation.MaxItems,
			DisplayField:     relation.DisplayField,
			ValueField:       relation.ValueField,
			PivotTable:       relation.PivotTable,
			PivotFields:      relation.PivotFields,
			Cascade:          relation.Cascade,
			OnDelete:         relation.OnDelete,
			OnUpdate:         relation.OnUpdate,
		}

		result = append(result, relationMeta)
	}

	return result
}

// GenerateJsonConfigMetadata generates metadata for JSON configuration
func GenerateJsonConfigMetadata(config *JsonConfig) *JsonConfigMetadata {
	if config == nil {
		return nil
	}

	meta := &JsonConfigMetadata{
		Schema:          config.Schema,
		DefaultExpanded: config.DefaultExpanded,
		PathPrefix:      config.PathPrefix,
		EditorType:      config.EditorType,
		Nested:          config.Nested,
		RenderAs:        config.RenderAs,
		ObjectLabels:    config.ObjectLabels,
	}

	// Convert tabs configuration if present
	if config.TabsConfig != nil {
		tabsConfig := &JsonTabsConfigMetadata{
			TabPosition:      config.TabsConfig.TabPosition,
			DefaultActiveTab: config.TabsConfig.DefaultActiveTab,
		}

		// Convert tabs
		if len(config.TabsConfig.Tabs) > 0 {
			tabsConfig.Tabs = make([]JsonTabMetadata, 0, len(config.TabsConfig.Tabs))
			for _, tab := range config.TabsConfig.Tabs {
				tabsConfig.Tabs = append(tabsConfig.Tabs, JsonTabMetadata{
					Key:    tab.Key,
					Title:  tab.Title,
					Icon:   tab.Icon,
					Fields: tab.Fields,
				})
			}
		}

		meta.TabsConfig = tabsConfig
	}

	// Convert grid configuration if present
	if config.GridConfig != nil {
		gridConfig := &JsonGridConfigMetadata{
			Columns: config.GridConfig.Columns,
			Gutter:  config.GridConfig.Gutter,
		}

		// Convert field layouts
		if len(config.GridConfig.FieldLayouts) > 0 {
			gridConfig.FieldLayouts = make(map[string]*JsonFieldLayoutMetadata)
			for path, layout := range config.GridConfig.FieldLayouts {
				gridConfig.FieldLayouts[path] = &JsonFieldLayoutMetadata{
					Column:  layout.Column,
					Row:     layout.Row,
					ColSpan: layout.ColSpan,
					RowSpan: layout.RowSpan,
				}
			}
		}

		meta.GridConfig = gridConfig
	}

	// Generate properties metadata
	if len(config.Properties) > 0 {
		meta.Properties = make([]JsonPropertyMetadata, 0, len(config.Properties))

		for _, prop := range config.Properties {
			meta.Properties = append(meta.Properties, GenerateJsonPropertyMetadata(prop))
		}
	}

	return meta
}

// GenerateJsonPropertyMetadata generates metadata for a JSON property
func GenerateJsonPropertyMetadata(prop JsonProperty) JsonPropertyMetadata {
	propMeta := JsonPropertyMetadata{
		Path:     prop.Path,
		Label:    prop.Label,
		Type:     prop.Type,
		ReadOnly: prop.ReadOnly,
		Hidden:   prop.Hidden,
	}

	// Add validation if present
	if prop.Validation != nil {
		propMeta.Validation = &ValidationMetadata{
			Required:  prop.Validation.Required,
			Min:       prop.Validation.Min,
			Max:       prop.Validation.Max,
			MinLength: prop.Validation.MinLength,
			MaxLength: prop.Validation.MaxLength,
			Pattern:   prop.Validation.Pattern,
			Message:   prop.Validation.Message,
		}
	}

	// Add form metadata if present
	if prop.Form != nil {
		propMeta.Form = &FormConfigMetadata{
			Placeholder: prop.Form.Placeholder,
			Help:        prop.Form.Help,
			Tooltip:     prop.Form.Tooltip,
			Width:       prop.Form.Width,
			DependentOn: prop.Form.DependentOn,
			Condition:   prop.Form.Condition,
		}
	}

	// Generate nested properties metadata recursively
	if len(prop.Properties) > 0 {
		propMeta.Properties = make([]JsonPropertyMetadata, 0, len(prop.Properties))

		for _, nestedProp := range prop.Properties {
			propMeta.Properties = append(propMeta.Properties, GenerateJsonPropertyMetadata(nestedProp))
		}
	}

	return propMeta
}

// GenerateFileConfigMetadata generates metadata for file configuration
func GenerateFileConfigMetadata(config *FileConfig) *FileConfigMetadata {
	if config == nil {
		return nil
	}

	meta := &FileConfigMetadata{
		AllowedTypes:       config.AllowedTypes,
		MaxSize:            config.MaxSize,
		BaseURL:            config.BaseURL,
		IsImage:            config.IsImage,
		MaxWidth:           config.MaxWidth,
		MaxHeight:          config.MaxHeight,
		GenerateThumbnails: config.GenerateThumbnails,
	}

	// Convert thumbnail sizes
	if len(config.ThumbnailSizes) > 0 {
		meta.ThumbnailSizes = make([]ThumbnailSizeMetadata, 0, len(config.ThumbnailSizes))
		for _, size := range config.ThumbnailSizes {
			meta.ThumbnailSizes = append(meta.ThumbnailSizes, ThumbnailSizeMetadata{
				Name:            size.Name,
				Width:           size.Width,
				Height:          size.Height,
				KeepAspectRatio: size.KeepAspectRatio,
			})
		}
	}

	return meta
}

// GenerateRichTextConfigMetadata generates metadata for rich text configuration
func GenerateRichTextConfigMetadata(config *RichTextConfig) *RichTextConfigMetadata {
	if config == nil {
		return nil
	}

	return &RichTextConfigMetadata{
		Toolbar:      config.Toolbar,
		Height:       config.Height,
		Placeholder:  config.Placeholder,
		EnableImages: config.EnableImages,
		MaxLength:    config.MaxLength,
		ShowCounter:  config.ShowCounter,
		Format:       config.Format,
	}
}

// GenerateSelectConfigMetadata generates metadata for select configuration
func GenerateSelectConfigMetadata(config *SelectConfig) *SelectConfigMetadata {
	if config == nil {
		return nil
	}

	meta := &SelectConfigMetadata{
		Multiple:    config.Multiple,
		Searchable:  config.Searchable,
		Creatable:   config.Creatable,
		OptionsURL:  config.OptionsURL,
		DependsOn:   config.DependsOn,
		Placeholder: config.Placeholder,
		Clearable:   config.Clearable,
		DisplayMode: config.DisplayMode,
	}

	// Convert dependent options
	if len(config.DependentOptions) > 0 {
		meta.DependentOptions = make(map[string][]OptionMetadata)
		for key, options := range config.DependentOptions {
			optionsMeta := make([]OptionMetadata, 0, len(options))
			for _, opt := range options {
				optionsMeta = append(optionsMeta, OptionMetadata{
					Value: opt.Value,
					Label: opt.Label,
				})
			}
			meta.DependentOptions[key] = optionsMeta
		}
	}

	return meta
}

// GenerateComputedFieldConfigMetadata generates metadata for computed field configuration
func GenerateComputedFieldConfigMetadata(config *ComputedFieldConfig) *ComputedFieldConfigMetadata {
	if config == nil {
		return nil
	}

	return &ComputedFieldConfigMetadata{
		DependsOn:    config.DependsOn,
		Expression:   config.Expression,
		ClientSide:   config.ClientSide,
		Format:       config.Format,
		Persist:      config.Persist,
		ComputeOrder: config.ComputeOrder,
	}
}

// GenerateAntDesignConfigMetadata generates metadata for Ant Design configuration
func GenerateAntDesignConfigMetadata(config *AntDesignConfig) *AntDesignConfigMetadata {
	if config == nil {
		return nil
	}

	meta := &AntDesignConfigMetadata{
		ComponentType: config.ComponentType,
		Props:         config.Props,
		FormItemProps: config.FormItemProps,
		Dependencies:  config.Dependencies,
	}

	// Convert rules
	if len(config.Rules) > 0 {
		meta.Rules = make([]AntDesignRuleMetadata, 0, len(config.Rules))
		for _, rule := range config.Rules {
			meta.Rules = append(meta.Rules, AntDesignRuleMetadata{
				Type:             rule.Type,
				Message:          rule.Message,
				Value:            rule.Value,
				ValidateOnBlur:   rule.ValidateOnBlur,
				ValidateOnChange: rule.ValidateOnChange,
				Pattern:          rule.Pattern,
				ValidateTrigger:  rule.ValidateTrigger,
			})
		}
	}

	return meta
}
