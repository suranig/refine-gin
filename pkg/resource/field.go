package resource

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Field represents a resource field
type Field struct {
	Name        string
	Type        string
	Label       string
	Validation  *Validation
	Options     []Option
	Relation    *RelationConfig
	List        *ListConfig
	Form        *FormConfig
	Validators  []Validator
	Json        *JsonConfig
	ReadOnly    bool                 // Indicates if the field is read-only (not editable)
	Hidden      bool                 // Indicates if the field should be hidden in UI
	File        *FileConfig          // Configuration for file/image fields
	RichText    *RichTextConfig      // Configuration for rich text fields
	Select      *SelectConfig        // Configuration for select fields
	Computed    *ComputedFieldConfig // Configuration for computed fields
	AntDesign   *AntDesignConfig     // Configuration specific to Ant Design
	Permissions map[string][]string  // Map of operations to roles with permission
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
	Required       bool                       `json:"required,omitempty"`
	Min            float64                    `json:"min,omitempty"`
	Max            float64                    `json:"max,omitempty"`
	MinLength      int                        `json:"minLength,omitempty"`
	MaxLength      int                        `json:"maxLength,omitempty"`
	Pattern        string                     `json:"pattern,omitempty"`
	Message        string                     `json:"message,omitempty"`
	Custom         string                     `json:"custom,omitempty"`         // Custom validation rule
	Conditional    *JsonConditionalValidation `json:"conditional,omitempty"`    // Validation rules that depend on other properties
	AsyncValidator string                     `json:"asyncValidator,omitempty"` // URL for asynchronous validation
}

// JsonConditionalValidation defines conditional validation for JSON properties
type JsonConditionalValidation struct {
	Path     string      `json:"path"`     // JSON path to the property this validation depends on
	Operator string      `json:"operator"` // Comparison operator (eq, neq, gt, lt, etc.)
	Value    interface{} `json:"value"`    // Value to compare against
	Message  string      `json:"message"`  // Custom message when condition fails
}

// JsonProperty defines a property in a JSON structure
type JsonProperty struct {
	// Property path (e.g. "config.oauth.client_id")
	Path string `json:"path"`

	// Property label for display
	Label string `json:"label,omitempty"`

	// Property type (string, number, boolean, object, array)
	Type string `json:"type"`

	// Whether the property is read-only (not editable)
	ReadOnly bool `json:"readOnly,omitempty"`

	// Whether the property should be hidden in UI
	Hidden bool `json:"hidden,omitempty"`

	// Additional validation for the property
	Validation *JsonValidation `json:"validation,omitempty"`

	// For object types, nested properties
	Properties []JsonProperty `json:"properties,omitempty"`

	// UI configuration
	Form *FormConfig `json:"form,omitempty"`
}

// Validation defines field validation rules
type Validation struct {
	Required       bool
	Min            float64
	Max            float64
	MinLength      int
	MaxLength      int
	Pattern        string
	Message        string
	Custom         string                 // Custom validation rule (e.g., JavaScript expression)
	Conditional    *ConditionalValidation // Validation rules that depend on other fields
	AsyncValidator string                 // URL for asynchronous validation
}

// ConditionalValidation defines validation rules that depend on other fields
type ConditionalValidation struct {
	Field    string      // The field this validation depends on
	Operator string      // Comparison operator (eq, neq, gt, lt, etc.)
	Value    interface{} // Value to compare against
	Message  string      // Custom message when condition fails
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
	Width       string
	DependentOn string
	Condition   string

	// Dependency configuration
	Dependent *FormDependency `json:"dependent,omitempty"`

	// Width percentage (25, 50, 75, 100)
	WidthPercent int `json:"widthPercent,omitempty" validate:"omitempty,oneof=25 50 75 100"`

	// Visibility condition using structured condition
	VisibilityCondition *FormCondition `json:"visibilityCondition,omitempty"`
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

// FileConfig defines configuration for file/image fields
type FileConfig struct {
	// Allowed MIME types (e.g. "image/jpeg", "application/pdf")
	AllowedTypes []string `json:"allowedTypes,omitempty"`

	// Maximum file size in bytes
	MaxSize int64 `json:"maxSize,omitempty"`

	// Storage path where files are saved
	StoragePath string `json:"storagePath,omitempty"`

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
	ThumbnailSizes []ThumbnailSize `json:"thumbnailSizes,omitempty"`
}

// ThumbnailSize defines a thumbnail configuration
type ThumbnailSize struct {
	// Name of the thumbnail (e.g. "small", "medium")
	Name string `json:"name"`

	// Width in pixels
	Width int `json:"width"`

	// Height in pixels
	Height int `json:"height"`

	// Whether to keep aspect ratio when resizing
	KeepAspectRatio bool `json:"keepAspectRatio,omitempty"`
}

// RichTextConfig defines configuration for rich text fields
type RichTextConfig struct {
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

// SelectConfig defines configuration for select fields
type SelectConfig struct {
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
	DependentOptions map[string][]Option `json:"dependentOptions,omitempty"`

	// Placeholder text
	Placeholder string `json:"placeholder,omitempty"`

	// Whether to allow clearing the selection
	Clearable bool `json:"clearable,omitempty"`

	// Display mode (dropdown, radio, checkboxes, tags)
	DisplayMode string `json:"displayMode,omitempty"`
}

// ComputedFieldConfig defines configuration for computed fields
type ComputedFieldConfig struct {
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

// AntDesignConfig defines configuration specific to Ant Design components
type AntDesignConfig struct {
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

// ConditionalValidator validates values based on conditions from other fields
type ConditionalValidator struct {
	Field       string                  // The field this validation depends on
	Operator    string                  // Comparison operator (eq, neq, gt, lt, etc.)
	Value       interface{}             // Value to compare against
	Message     string                  // Custom message when condition fails
	ValidateFn  func(interface{}) error // Validation function to apply conditionally
	ModelGetter func() interface{}      // Function to get the full model for accessing other fields
}

func (v ConditionalValidator) Validate(value interface{}) error {
	// If no model getter is provided, we can't perform conditional validation
	if v.ModelGetter == nil {
		return nil
	}

	// Get the model
	model := v.ModelGetter()
	if model == nil {
		return nil // No model to check conditions against
	}

	// Get the value of the dependent field
	fieldValue, err := GetFieldValue(model, v.Field)
	if err != nil {
		return nil // Can't get dependent field, skip validation
	}

	// Check the condition
	conditionMet, err := evaluateCondition(fmt.Sprintf("%v", fieldValue), v.Operator, v.Value)
	if err != nil {
		return nil // Can't evaluate condition, skip validation
	}

	// If condition is met, apply the validation
	if conditionMet && v.ValidateFn != nil {
		if err := v.ValidateFn(value); err != nil {
			if v.Message != "" {
				return fmt.Errorf(v.Message)
			}
			return err
		}
	}

	return nil
}

// evaluateCondition compares a field value against a condition
func evaluateCondition(fieldValue, operator string, compareValue interface{}) (bool, error) {
	// Convert fieldValue to comparable format
	var fieldValueFloat float64
	var compareValueFloat float64
	var err error

	// For numeric comparisons
	if operator == "gt" || operator == "lt" || operator == "gte" || operator == "lte" {
		fieldValueFloat, err = convertToFloat(fieldValue)
		if err != nil {
			return false, err
		}

		compareValueFloat, err = convertToFloat(compareValue)
		if err != nil {
			return false, err
		}
	}

	// Perform comparison based on operator
	switch operator {
	case "eq":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", compareValue), nil
	case "neq":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", compareValue), nil
	case "gt":
		return fieldValueFloat > compareValueFloat, nil
	case "lt":
		return fieldValueFloat < compareValueFloat, nil
	case "gte":
		return fieldValueFloat >= compareValueFloat, nil
	case "lte":
		return fieldValueFloat <= compareValueFloat, nil
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", compareValue)), nil
	case "startsWith":
		return strings.HasPrefix(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", compareValue)), nil
	case "endsWith":
		return strings.HasSuffix(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", compareValue)), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// convertToFloat converts an interface{} to float64 for numeric comparisons
func convertToFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// CustomValidator validates using a custom expression
type CustomValidator struct {
	Expression string      // Custom validation expression
	Message    string      // Error message
	Context    interface{} // Context for validation (e.g., the model)
}

func (v CustomValidator) Validate(value interface{}) error {
	// Note: In a real implementation, you would evaluate the expression
	// For now, we'll just return nil as this requires an expression engine
	// which is out of scope for this implementation

	// TODO: Implement expression evaluation logic
	// Example approaches:
	// 1. Use a JS interpreter like goja
	// 2. Use a rules engine
	// 3. Use template evaluation

	// For now, return success
	return nil
}

// AsyncValidator represents a validator that performs asynchronous validation
type AsyncValidator struct {
	URL     string // URL to send validation request to
	Message string // Error message on validation failure
}

func (v AsyncValidator) Validate(value interface{}) error {
	// Note: In a real implementation, you would make an HTTP request
	// to the validation endpoint and process the response.
	// For now, we'll just return nil as async validation happens client-side

	// Example implementation:
	// client := &http.Client{Timeout: 5 * time.Second}
	// valueJSON, _ := json.Marshal(value)
	// resp, err := client.Post(v.URL, "application/json", bytes.NewBuffer(valueJSON))
	// if err != nil {
	//     return err
	// }
	// defer resp.Body.Close()
	//
	// if resp.StatusCode != http.StatusOK {
	//     if v.Message != "" {
	//         return fmt.Errorf(v.Message)
	//     }
	//     return fmt.Errorf("validation failed with status: %d", resp.StatusCode)
	// }

	// For now, return success
	return nil
}
