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

// GenerateResourceMetadata generates resource metadata from a resource
func GenerateResourceMetadata(res Resource) ResourceMetadata {
	metadata := ResourceMetadata{
		Name:        res.GetName(),
		Label:       res.GetLabel(),
		Icon:        res.GetIcon(),
		Operations:  res.GetOperations(),
		IDFieldName: res.GetIDFieldName(),
		DefaultSort: res.GetDefaultSort(),
		Filters:     res.GetFilters(),
		Searchable:  res.GetSearchable(),
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

	for _, field := range fields {
		fieldMeta := FieldMetadata{
			Name:       field.Name,
			Type:       field.Type,
			Filterable: field.Filterable,
			Sortable:   field.Sortable,
			Searchable: field.Searchable,
			Required:   field.Required,
			Unique:     field.Unique,
		}

		// Add validators metadata
		if len(field.Validators) > 0 {
			fieldMeta.Validators = GenerateValidatorsMetadata(field.Validators)
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
