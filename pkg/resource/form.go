package resource

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// FormLayout defines the layout configuration for a form
type FormLayout struct {
	// Number of columns in the form grid
	Columns int `json:"columns"`

	// Space between grid cells
	Gutter int `json:"gutter"`

	// Form sections
	Sections []*FormSection `json:"sections,omitempty"`

	// Field layouts - used for positioning fields in the grid
	FieldLayouts []*FormFieldLayout `json:"fieldLayouts,omitempty"`
}

// FormSection defines a section in a form
type FormSection struct {
	// Section ID - must be unique within a form
	ID string `json:"id" validate:"required"`

	// Section title
	Title string `json:"title,omitempty"`

	// Icon for the section
	Icon string `json:"icon,omitempty"`

	// Whether the section is collapsible
	Collapsible bool `json:"collapsible,omitempty"`

	// Whether the section is collapsed by default
	DefaultCollapsed bool `json:"defaultCollapsed,omitempty"`

	// CSS class name for styling
	ClassName string `json:"className,omitempty"`

	// Visibility condition
	Condition *FormCondition `json:"condition,omitempty"`
}

// FormFieldLayout defines the position of a field in the form grid
type FormFieldLayout struct {
	// Field name/path
	Field string `json:"field" validate:"required"`

	// Section ID where this field should be placed
	SectionID string `json:"sectionId,omitempty"`

	// Column number (0-based)
	Column int `json:"column,omitempty"`

	// Row number (0-based)
	Row int `json:"row,omitempty"`

	// Column span
	ColSpan int `json:"colSpan,omitempty"`

	// Row span
	RowSpan int `json:"rowSpan,omitempty"`

	// CSS class name for styling
	ClassName string `json:"className,omitempty"`
}

// FormCondition defines a visibility condition
type FormCondition struct {
	// Field to check
	Field string `json:"field" validate:"required"`

	// Operator to apply
	Operator string `json:"operator" validate:"required,oneof=eq neq gt lt gte lte contains startsWith endsWith"`

	// Value to compare against
	Value interface{} `json:"value"`
}

// FormDependency defines a dependency between fields
type FormDependency struct {
	// Source field
	Field string `json:"field" validate:"required"`

	// Expected value for the field to show this field
	Value interface{} `json:"value"`

	// Whether the field should be hidden when condition is met
	HideWhenMatched bool `json:"hideWhenMatched,omitempty"`
}

// ValidateFormLayout validates a form layout
func ValidateFormLayout(layout *FormLayout) error {
	// Nil layout is valid
	if layout == nil {
		return nil
	}

	validate := validator.New()
	return validate.Struct(layout)
}

// ValidateFormDependencies validates form dependencies consistency
func ValidateFormDependencies(layout *FormLayout, fields []Field) error {
	// Check if all referenced fields exist
	if layout == nil {
		return nil
	}

	// Create a map of existing field names for quick lookup
	fieldMap := make(map[string]bool)
	for _, field := range fields {
		fieldMap[field.Name] = true
	}

	// Check field layouts
	for _, fieldLayout := range layout.FieldLayouts {
		if !fieldMap[fieldLayout.Field] {
			return fmt.Errorf("field '%s' referenced in layout does not exist", fieldLayout.Field)
		}

		// Check if section exists
		if fieldLayout.SectionID != "" {
			sectionExists := false
			for _, section := range layout.Sections {
				if section.ID == fieldLayout.SectionID {
					sectionExists = true
					break
				}
			}
			if !sectionExists {
				return fmt.Errorf("section '%s' referenced in field layout does not exist", fieldLayout.SectionID)
			}
		}
	}

	// Check conditions
	for _, section := range layout.Sections {
		if section.Condition != nil && !fieldMap[section.Condition.Field] {
			return fmt.Errorf("field '%s' referenced in section condition does not exist", section.Condition.Field)
		}
	}

	return nil
}

// GenerateDefaultFormLayout creates a default form layout based on fields
func GenerateDefaultFormLayout(fields []Field) *FormLayout {
	// Create a default section
	defaultSection := &FormSection{
		ID:    "default",
		Title: "General Information",
	}

	// Create field layouts
	fieldLayouts := make([]*FormFieldLayout, 0, len(fields))
	row := 0

	for _, field := range fields {
		// Skip ID fields and hidden fields
		if field.Name == "ID" || field.Name == "id" || field.Hidden {
			continue
		}

		fieldLayout := &FormFieldLayout{
			Field:     field.Name,
			SectionID: "default",
			Row:       row,
			Column:    0,
			ColSpan:   1,
		}

		fieldLayouts = append(fieldLayouts, fieldLayout)
		row++
	}

	return &FormLayout{
		Columns:      1,
		Gutter:       16,
		Sections:     []*FormSection{defaultSection},
		FieldLayouts: fieldLayouts,
	}
}
