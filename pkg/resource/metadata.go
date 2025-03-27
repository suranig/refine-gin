package resource

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

	// Field validators
	Validators []ValidatorMetadata `json:"validators,omitempty"`

	// JSON configuration (for object/jsonb fields)
	Json *JsonConfigMetadata `json:"json,omitempty"`
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
}

// JsonPropertyMetadata represents metadata for a JSON property
type JsonPropertyMetadata struct {
	// Property path (e.g. "config.oauth.client_id")
	Path string `json:"path,omitempty"`

	// Property label for display
	Label string `json:"label,omitempty"`

	// Property type (string, number, boolean, object, array)
	Type string `json:"type,omitempty"`

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

		result = append(result, fieldMeta)
	}

	return result
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
		Path:  prop.Path,
		Label: prop.Label,
		Type:  prop.Type,
	}

	// Add validation metadata if present
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
