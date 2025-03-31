package resource

import (
	"fmt"
	"regexp"
	"strconv"
)

// Field represents a resource field
type Field struct {
	Name       string
	Type       string
	Label      string
	Validation *Validation
	Options    []Option
	Relation   *RelationConfig
	List       *ListConfig
	Form       *FormConfig
	Validators []Validator
	Json       *JsonConfig
	ReadOnly   bool // Indicates if the field is read-only (not editable)
	Hidden     bool // Indicates if the field should be hidden in UI
}

// JsonConfig defines configuration for JSON fields
type JsonConfig struct {
	// Schema for JSON field validation and UI
	Schema map[string]interface{} `json:"schema,omitempty"`

	// Properties defines nested fields in the JSON structure
	Properties []JsonProperty `json:"properties,omitempty"`

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
	TabsConfig *JsonTabsConfig `json:"tabsConfig,omitempty"`

	// GridConfig holds configuration for grid layout
	GridConfig *JsonGridConfig `json:"gridConfig,omitempty"`

	// ObjectLabels provides display labels for nested objects by their path
	ObjectLabels map[string]string `json:"objectLabels,omitempty"`
}

// JsonValidation defines validation rules for JSON properties
type JsonValidation struct {
	Required  bool    `json:"required,omitempty"`
	Min       float64 `json:"min,omitempty"`
	Max       float64 `json:"max,omitempty"`
	MinLength int     `json:"minLength,omitempty"`
	MaxLength int     `json:"maxLength,omitempty"`
	Pattern   string  `json:"pattern,omitempty"`
	Message   string  `json:"message,omitempty"`
}

// JsonProperty defines a property in a JSON structure
type JsonProperty struct {
	// Property path (e.g. "config.oauth.client_id")
	Path string `json:"path"`

	// Property label for display
	Label string `json:"label,omitempty"`

	// Property type (string, number, boolean, object, array)
	Type string `json:"type"`

	// Additional validation for the property
	Validation *JsonValidation `json:"validation,omitempty"`

	// For object types, nested properties
	Properties []JsonProperty `json:"properties,omitempty"`

	// UI configuration
	Form *FormConfig `json:"form,omitempty"`
}

// Validation defines field validation rules
type Validation struct {
	Required  bool
	Min       float64
	Max       float64
	MinLength int
	MaxLength int
	Pattern   string
	Message   string
}

// RelationConfig defines field relation configuration
type RelationConfig struct {
	Resource     string
	Type         string
	ValueField   string
	DisplayField string
	FetchMode    string
	Endpoint     string
	IDField      string
	Required     bool
	AllowNone    bool
	MinItems     int
	MaxItems     int
	Searchable   bool
	Async        bool
	Placeholder  string
}

// Option represents a field option for select/enum fields
type Option struct {
	Value interface{}
	Label string
}

// ListConfig defines field configuration for list view
type ListConfig struct {
	Width    int
	Fixed    string
	Ellipsis bool
}

// FormConfig defines field configuration for form view
type FormConfig struct {
	Placeholder string
	Help        string
	Tooltip     string
}

// Validator represents a field validator
type Validator interface {
	Validate(value interface{}) error
}

// StringValidator validates string values
type StringValidator struct {
	MinLength int
	MaxLength int
	Pattern   string
}

func (v StringValidator) Validate(value interface{}) error {
	// Convert value to string
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	// Check min length
	if v.MinLength > 0 && len(str) < v.MinLength {
		return fmt.Errorf("string length must be at least %d characters", v.MinLength)
	}

	// Check max length
	if v.MaxLength > 0 && len(str) > v.MaxLength {
		return fmt.Errorf("string length must not exceed %d characters", v.MaxLength)
	}

	// Check pattern
	if v.Pattern != "" {
		matched, err := regexp.MatchString(v.Pattern, str)
		if err != nil {
			return fmt.Errorf("invalid pattern: %v", err)
		}
		if !matched {
			return fmt.Errorf("string does not match pattern: %s", v.Pattern)
		}
	}

	return nil
}

// NumberValidator validates numeric values
type NumberValidator struct {
	Min float64
	Max float64
}

func (v NumberValidator) Validate(value interface{}) error {
	// Convert value to float64
	var num float64
	var err error

	switch val := value.(type) {
	case float64:
		num = val
	case float32:
		num = float64(val)
	case int:
		num = float64(val)
	case int8:
		num = float64(val)
	case int16:
		num = float64(val)
	case int32:
		num = float64(val)
	case int64:
		num = float64(val)
	case uint:
		num = float64(val)
	case uint8:
		num = float64(val)
	case uint16:
		num = float64(val)
	case uint32:
		num = float64(val)
	case uint64:
		num = float64(val)
	case string:
		num, err = strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("cannot convert string to number: %v", err)
		}
	default:
		return fmt.Errorf("value must be a number or a string that can be converted to a number")
	}

	// Check min value
	if v.Min != 0 && num < v.Min {
		return fmt.Errorf("number must be at least %v", v.Min)
	}

	// Check max value
	if v.Max != 0 && num > v.Max {
		return fmt.Errorf("number must not exceed %v", v.Max)
	}

	return nil
}

// Filter defines a filter configuration
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// Sort defines a sort configuration
type Sort struct {
	Field string
	Order string
}

// JsonTabsConfig defines configuration for rendering JSON in tabs
type JsonTabsConfig struct {
	// Tabs collection for the JSON structure
	Tabs []JsonTab `json:"tabs,omitempty"`

	// TabPosition specifies where tabs are placed ("top", "right", "bottom", "left")
	TabPosition string `json:"tabPosition,omitempty"`

	// DefaultActiveTab specifies the key of the default active tab
	DefaultActiveTab string `json:"defaultActiveTab,omitempty"`
}

// JsonTab defines a single tab in a tabs-based JSON editor
type JsonTab struct {
	// Key uniquely identifies the tab
	Key string `json:"key"`

	// Title is the display name of the tab
	Title string `json:"title,omitempty"`

	// Icon specifies an optional icon for the tab
	Icon string `json:"icon,omitempty"`

	// Fields lists the JSON property paths included in this tab
	Fields []string `json:"fields,omitempty"`
}

// JsonGridConfig defines configuration for grid layout of JSON fields
type JsonGridConfig struct {
	// Columns defines the number of columns in the grid
	Columns int `json:"columns,omitempty"`

	// Gutter defines the space between grid items
	Gutter int `json:"gutter,omitempty"`

	// FieldLayouts maps field paths to their layout configuration
	FieldLayouts map[string]*JsonFieldLayout `json:"fieldLayouts,omitempty"`
}

// JsonFieldLayout defines the grid position and span of a field
type JsonFieldLayout struct {
	// Column specifies the starting column (1-based)
	Column int `json:"column,omitempty"`

	// Row specifies the starting row (1-based)
	Row int `json:"row,omitempty"`

	// ColSpan specifies how many columns the field spans
	ColSpan int `json:"colSpan,omitempty"`

	// RowSpan specifies how many rows the field spans
	RowSpan int `json:"rowSpan,omitempty"`
}
