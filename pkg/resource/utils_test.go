package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// UtilsTestResource is a simplified implementation of Resource interface for testing
type UtilsTestResource struct {
	EditableFieldsValue []string
	FieldsValue         []Field
}

func (r *UtilsTestResource) GetEditableFields() []string {
	return r.EditableFieldsValue
}

func (r *UtilsTestResource) GetFields() []Field {
	return r.FieldsValue
}

func (r *UtilsTestResource) GetName() string                                  { return "" }
func (r *UtilsTestResource) GetLabel() string                                 { return "" }
func (r *UtilsTestResource) GetIcon() string                                  { return "" }
func (r *UtilsTestResource) GetModel() interface{}                            { return nil }
func (r *UtilsTestResource) GetOperations() []Operation                       { return nil }
func (r *UtilsTestResource) HasOperation(op Operation) bool                   { return false }
func (r *UtilsTestResource) GetDefaultSort() *Sort                            { return nil }
func (r *UtilsTestResource) GetFilters() []Filter                             { return nil }
func (r *UtilsTestResource) GetMiddlewares() []interface{}                    { return nil }
func (r *UtilsTestResource) GetRelations() []Relation                         { return nil }
func (r *UtilsTestResource) HasRelation(name string) bool                     { return false }
func (r *UtilsTestResource) GetRelation(name string) *Relation                { return nil }
func (r *UtilsTestResource) GetIDFieldName() string                           { return "ID" }
func (r *UtilsTestResource) GetField(name string) *Field                      { return nil }
func (r *UtilsTestResource) GetSearchable() []string                          { return nil }
func (r *UtilsTestResource) GetFilterableFields() []string                    { return nil }
func (r *UtilsTestResource) GetSortableFields() []string                      { return nil }
func (r *UtilsTestResource) GetTableFields() []string                         { return nil }
func (r *UtilsTestResource) GetFormFields() []string                          { return nil }
func (r *UtilsTestResource) GetRequiredFields() []string                      { return nil }
func (r *UtilsTestResource) GetPermissions() map[string][]string              { return nil }
func (r *UtilsTestResource) HasPermission(operation string, role string) bool { return false }
func (r *UtilsTestResource) GetFormLayout() *FormLayout                       { return nil }
func (r *UtilsTestResource) CreateSliceOfType() interface{}                   { return nil }
func (r *UtilsTestResource) CreateInstanceOfType() interface{}                { return nil }
func (r *UtilsTestResource) SetID(obj interface{}, id interface{}) error      { return nil }
func (r *UtilsTestResource) SetCustomID(obj interface{}, field string, id interface{}) error {
	return nil
}

func TestGetEditableFieldsMap(t *testing.T) {
	// Resource with explicitly defined editable fields
	res := &UtilsTestResource{
		EditableFieldsValue: []string{"Name", "Email"},
	}

	// Get editable fields map
	editableFields := getEditableFieldsMap(res)

	// Check that the map contains the expected fields
	assert.True(t, editableFields["name"])
	assert.True(t, editableFields["email"])
	assert.False(t, editableFields["id"])
	assert.False(t, editableFields["createdat"])

	// Resource without explicitly defined editable fields
	resWithoutEditable := &UtilsTestResource{
		EditableFieldsValue: []string{},
		FieldsValue: []Field{
			{Name: "ID", ReadOnly: true},
			{Name: "Name", ReadOnly: false},
			{Name: "Email", ReadOnly: false},
			{Name: "CreatedAt", ReadOnly: true},
		},
	}

	// Get editable fields map
	editableFieldsFromReadOnly := getEditableFieldsMap(resWithoutEditable)

	// Check that the map contains the expected fields
	assert.False(t, editableFieldsFromReadOnly["id"])
	assert.True(t, editableFieldsFromReadOnly["name"])
	assert.True(t, editableFieldsFromReadOnly["email"])
	assert.False(t, editableFieldsFromReadOnly["createdat"])
}

func TestFilterMapData(t *testing.T) {
	// Create test data
	data := map[string]interface{}{
		"id":         1,
		"name":       "Test Name",
		"email":      "test@example.com",
		"created_at": "2023-01-01",
	}

	// Create editable fields map
	editableFields := map[string]bool{
		"name":  true,
		"email": true,
	}

	// Filter the map
	result := filterMapData(data, editableFields)

	// Check that only editable fields are included
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "Test Name", result["name"])
	assert.Equal(t, "test@example.com", result["email"])
	assert.NotContains(t, result, "id")
	assert.NotContains(t, result, "created_at")
}

func TestFilterStructData(t *testing.T) {
	// Create test struct
	type TestStruct struct {
		ID        int
		Name      string
		Email     string
		CreatedAt string
	}

	data := TestStruct{
		ID:        1,
		Name:      "Test Name",
		Email:     "test@example.com",
		CreatedAt: "2023-01-01",
	}

	// Create editable fields map
	editableFields := map[string]bool{
		"name":  true,
		"email": true,
	}

	// Filter the struct
	result := filterStructData(data, editableFields)

	// Check that only editable fields are included
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "Test Name", result["Name"])
	assert.Equal(t, "test@example.com", result["Email"])
	assert.NotContains(t, result, "ID")
	assert.NotContains(t, result, "CreatedAt")

	// Test with non-struct input
	nonStructResult := filterStructData("string", editableFields)
	assert.Nil(t, nonStructResult)
}

func TestFilterOutReadOnlyFields(t *testing.T) {
	type TestModel struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}

	// Create a test resource with editable fields
	testResource := &UtilsTestResource{
		EditableFieldsValue: []string{"Name", "Email"},
	}

	// Test struct data
	data := TestModel{
		ID:        1,
		Name:      "Test Name",
		Email:     "test@example.com",
		CreatedAt: "2023-01-01",
	}

	// Filter out read-only fields
	result := FilterOutReadOnlyFields(data, testResource)

	// Verify the result
	mapResult, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(mapResult))
	assert.Equal(t, "Test Name", mapResult["Name"])
	assert.Equal(t, "test@example.com", mapResult["Email"])
	assert.NotContains(t, mapResult, "ID")
	assert.NotContains(t, mapResult, "CreatedAt")

	// Test map data
	mapData := map[string]interface{}{
		"id":         1,
		"name":       "Test Name",
		"email":      "test@example.com",
		"created_at": "2023-01-01",
	}

	// Filter out read-only fields from map
	mapDataResult := FilterOutReadOnlyFields(mapData, testResource)
	resultMap, ok := mapDataResult.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(resultMap))
	assert.Equal(t, "Test Name", resultMap["name"])
	assert.Equal(t, "test@example.com", resultMap["email"])
	assert.NotContains(t, resultMap, "id")
	assert.NotContains(t, resultMap, "created_at")

	// Test nil input
	var nilData *TestModel
	nilResult := FilterOutReadOnlyFields(nilData, testResource)
	assert.Nil(t, nilResult)

	// Test non-struct, non-map input
	strData := "test string"
	strResult := FilterOutReadOnlyFields(strData, testResource)
	assert.Equal(t, strData, strResult)
}

func TestValidateNestedJson(t *testing.T) {
	// Create test config
	config := &JsonConfig{
		Properties: []JsonProperty{
			{
				Path: "name",
				Type: "string",
				Validation: &JsonValidation{
					Required:  true,
					MinLength: 3,
					MaxLength: 50,
				},
			},
			{
				Path: "age",
				Type: "number",
				Validation: &JsonValidation{
					Required: true,
					Min:      18,
					Max:      100,
				},
			},
		},
	}

	// Valid data
	validData := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}
	valid, errors := ValidateNestedJson(validData, config)
	assert.True(t, valid)
	assert.Empty(t, errors)

	// Invalid data - missing field
	invalidData1 := map[string]interface{}{
		"name": "John",
	}
	valid, errors = ValidateNestedJson(invalidData1, config)
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	// Invalid data - string too short
	invalidData2 := map[string]interface{}{
		"name": "Jo",
		"age":  30,
	}
	valid, errors = ValidateNestedJson(invalidData2, config)
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	// Invalid data - number too low
	invalidData3 := map[string]interface{}{
		"name": "John",
		"age":  17,
	}
	valid, errors = ValidateNestedJson(invalidData3, config)
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	// Nil data
	valid, errors = ValidateNestedJson(nil, config)
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	// Nil config
	valid, errors = ValidateNestedJson(validData, nil)
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	// String JSON data
	stringData := `{"name":"John Doe","age":30}`
	valid, errors = ValidateNestedJson(stringData, config)
	assert.True(t, valid)
	assert.Empty(t, errors)

	// Struct data
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	structData := Person{Name: "John Doe", Age: 30}
	valid, errors = ValidateNestedJson(structData, config)
	assert.True(t, valid)
	assert.Empty(t, errors)

	// Invalid string JSON
	invalidJSON := `{"name":"John","age":30`
	valid, errors = ValidateNestedJson(invalidJSON, config)
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	// Empty config properties
	emptyConfig := &JsonConfig{
		Properties: []JsonProperty{},
	}
	valid, errors = ValidateNestedJson(validData, emptyConfig)
	assert.True(t, valid)
	assert.Empty(t, errors)
}

func TestValidateNestedJson_ComplexObject(t *testing.T) {
	// Test with nested objects and arrays
	config := &JsonConfig{
		Properties: []JsonProperty{
			{
				Path: "person",
				Type: "object",
				Properties: []JsonProperty{
					{
						Path: "name",
						Type: "string",
						Validation: &JsonValidation{
							Required:  true,
							MinLength: 3,
						},
					},
					{
						Path: "address",
						Type: "object",
						Properties: []JsonProperty{
							{
								Path: "city",
								Type: "string",
								Validation: &JsonValidation{
									Required: true,
								},
							},
						},
						Validation: &JsonValidation{
							Required: true,
						},
					},
				},
				Validation: &JsonValidation{
					Required: true,
				},
			},
			{
				Path: "tags",
				Type: "array",
				Validation: &JsonValidation{
					Required:  true,
					MinLength: 1,
					MaxLength: 5,
				},
			},
		},
	}

	// Valid data
	validData := map[string]interface{}{
		"person": map[string]interface{}{
			"name": "John Doe",
			"address": map[string]interface{}{
				"city": "New York",
			},
		},
		"tags": []interface{}{"tag1", "tag2"},
	}
	valid, errors := ValidateNestedJson(validData, config)
	assert.True(t, valid)
	assert.Empty(t, errors)

	// Invalid data - missing required field in nested object
	invalidData := map[string]interface{}{
		"person": map[string]interface{}{
			"name":    "John Doe",
			"address": map[string]interface{}{
				// Missing required field "city"
			},
		},
		"tags": []interface{}{"tag1", "tag2"},
	}
	valid, errors = ValidateNestedJson(invalidData, config)
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	// Invalid data - array with too many items
	invalidData2 := map[string]interface{}{
		"person": map[string]interface{}{
			"name": "John Doe",
			"address": map[string]interface{}{
				"city": "New York",
			},
		},
		"tags": []interface{}{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6"},
	}
	valid, errors = ValidateNestedJson(invalidData2, config)
	assert.False(t, valid)
	assert.NotEmpty(t, errors)
}

func TestGetValueByPath(t *testing.T) {
	// Create test data
	data := map[string]interface{}{
		"name": "John",
		"age":  30,
		"address": map[string]interface{}{
			"street": "123 Main St",
			"city":   "New York",
			"zip":    "10001",
		},
		"tags": []interface{}{"tag1", "tag2", "tag3"},
		"scores": map[string]interface{}{
			"math":    []interface{}{90, 85, 95},
			"science": []interface{}{80, 75, 85},
		},
	}

	// Test root level path
	value, err := getValueByPath(data, "name")
	assert.NoError(t, err)
	assert.Equal(t, "John", value)

	// Test nested path
	value, err = getValueByPath(data, "address.city")
	assert.NoError(t, err)
	assert.Equal(t, "New York", value)

	// Test array access
	value, err = getValueByPath(data, "tags[1]")
	assert.NoError(t, err)
	assert.Equal(t, "tag2", value)

	// Test nested array access
	value, err = getValueByPath(data, "scores.math[2]")
	assert.NoError(t, err)
	assert.Equal(t, float64(95), value)

	// Test non-existent path
	_, err = getValueByPath(data, "non.existent.path")
	assert.Error(t, err)

	// Test non-existent array index
	_, err = getValueByPath(data, "tags[10]")
	assert.Error(t, err)

	// Test non-array with array access
	_, err = getValueByPath(data, "name[0]")
	assert.Error(t, err)

	// Test empty path
	value, err = getValueByPath(data, "")
	assert.NoError(t, err)
	assert.Equal(t, data, value)

	// Test nil data
	_, err = getValueByPath(nil, "name")
	assert.Error(t, err)
}

func TestProcessArrayAccess(t *testing.T) {
	// Test data
	data := map[string]interface{}{
		"items": []interface{}{"first", "second", "third"},
	}

	// Valid array access
	result, err := processArrayAccess(data, "items[1]", []string{})
	assert.NoError(t, err)
	assert.Equal(t, "second", result)

	// Invalid format - missing bracket
	_, err = processArrayAccess(data, "items", []string{})
	assert.Error(t, err)

	// Invalid format - no index
	_, err = processArrayAccess(data, "items[]", []string{})
	assert.Error(t, err)

	// Non-existing field
	_, err = processArrayAccess(data, "nonexistent[0]", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Invalid index
	_, err = processArrayAccess(data, "items[x]", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid array index")

	// Index out of bounds
	_, err = processArrayAccess(data, "items[10]", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of bounds")

	// Not an array
	nonArrayData := map[string]interface{}{
		"name": "test",
	}
	_, err = processArrayAccess(nonArrayData, "name[0]", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not an array")
}

func TestProcessFieldAccess(t *testing.T) {
	// Test data
	data := map[string]interface{}{
		"name":   "John",
		"age":    30,
		"active": true,
	}

	// Valid field access
	result, err := processFieldAccess(data, "name", []string{})
	assert.NoError(t, err)
	assert.Equal(t, "John", result)

	// Non-existing field
	_, err = processFieldAccess(data, "nonexistent", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Non-map data
	_, err = processFieldAccess("string", "name", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected object")

	// Test int to float64 conversion
	intData := map[string]interface{}{
		"count": 5, // int value
	}
	result, err = processFieldAccess(intData, "count", []string{})
	assert.NoError(t, err)
	assert.Equal(t, float64(5), result)
}

func TestConvertToJsonMap(t *testing.T) {
	// Test map input
	mapData := map[string]interface{}{
		"name": "John",
		"age":  30,
	}
	result, errors := convertToJsonMap(mapData)
	assert.Empty(t, errors)
	assert.Equal(t, mapData, result)

	// Test valid JSON string
	jsonStr := `{"name":"John","age":30}`
	result, errors = convertToJsonMap(jsonStr)
	assert.Empty(t, errors)
	assert.Equal(t, "John", result["name"])
	assert.Equal(t, float64(30), result["age"])

	// Test invalid JSON string
	invalidJson := `{"name":"John","age":30`
	result, errors = convertToJsonMap(invalidJson)
	assert.NotEmpty(t, errors)
	assert.Nil(t, result)

	// Test struct input
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	structData := Person{Name: "John", Age: 30}
	result, errors = convertToJsonMap(structData)
	assert.Empty(t, errors)
	assert.Equal(t, "John", result["name"])
	assert.Equal(t, float64(30), result["age"])
}

func TestValidateNumericProperty(t *testing.T) {
	// Test numeric property validation
	prop := &JsonProperty{
		Path: "age",
		Type: "number",
		Validation: &JsonValidation{
			Min: 18,
			Max: 100,
		},
	}

	// Valid value
	errors := validateNumericProperty(prop, float64(30))
	assert.Empty(t, errors)

	// Value too low
	errors = validateNumericProperty(prop, float64(10))
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "less than minimum")

	// Value too high
	errors = validateNumericProperty(prop, float64(150))
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "greater than maximum")

	// Invalid type
	errors = validateNumericProperty(prop, "not a number")
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "should be a number")

	// Test string that can be converted to number
	errors = validateNumericProperty(prop, "30")
	assert.Empty(t, errors)

	// Test integer value
	errors = validateNumericProperty(prop, 30)
	assert.Empty(t, errors)
}

func TestValidateStringProperty(t *testing.T) {
	// Test string property validation
	prop := &JsonProperty{
		Path: "name",
		Type: "string",
		Validation: &JsonValidation{
			MinLength: 3,
			MaxLength: 50,
			Pattern:   "^[a-zA-Z ]+$",
		},
	}

	// Valid value
	errors := validateStringProperty(prop, "John Doe")
	assert.Empty(t, errors)

	// String too short
	errors = validateStringProperty(prop, "Jo")
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "less than minimum length")

	// String too long
	longString := "This is a very long string that exceeds the maximum length limit"
	errors = validateStringProperty(prop, longString)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "greater than maximum length")

	// Pattern mismatch
	errors = validateStringProperty(prop, "John123")
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "does not match pattern")

	// Invalid type
	errors = validateStringProperty(prop, 123)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "should be a string")
}

func TestValidateObjectProperty(t *testing.T) {
	// Test object property validation
	prop := &JsonProperty{
		Path: "address",
		Type: "object",
		Properties: []JsonProperty{
			{
				Path: "city",
				Type: "string",
				Validation: &JsonValidation{
					Required: true,
				},
			},
		},
	}

	// Valid object
	validObj := map[string]interface{}{
		"city": "New York",
	}
	errors := validateObjectProperty(prop, validObj)
	assert.Empty(t, errors)

	// Missing required field
	invalidObj := map[string]interface{}{
		"street": "123 Main St",
	}
	errors = validateObjectProperty(prop, invalidObj)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "not found")

	// Not an object
	errors = validateObjectProperty(prop, "not an object")
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "should be an object")

	// Struct that can be converted to object
	type Address struct {
		City string `json:"city"`
	}
	structObj := Address{City: "New York"}
	errors = validateObjectProperty(prop, structObj)
	assert.Empty(t, errors)
}

func TestValidateArrayProperty(t *testing.T) {
	// Test array property validation
	prop := &JsonProperty{
		Path: "tags",
		Type: "array",
		Validation: &JsonValidation{
			MinLength: 1,
			MaxLength: 5,
		},
	}

	// Valid array
	validArray := []interface{}{"tag1", "tag2"}
	errors := validateArrayProperty(prop, validArray)
	assert.Empty(t, errors)

	// Array too short
	shortArray := []interface{}{}
	errors = validateArrayProperty(prop, shortArray)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "less than minimum")

	// Array too long
	longArray := []interface{}{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6"}
	errors = validateArrayProperty(prop, longArray)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "greater than maximum")

	// Not an array
	errors = validateArrayProperty(prop, "not an array")
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "should be an array")

	// Array-like struct
	type Tags struct {
		Items []string `json:"items"`
	}
	structArray := Tags{Items: []string{"tag1", "tag2"}}
	errors = validateArrayProperty(prop, structArray)
	assert.NotEmpty(t, errors) // This should fail since it's not directly an array
}

func TestDetectStringComponent(t *testing.T) {
	// Test password field detection
	passwordField := &Field{
		Name: "password",
		Type: "string",
	}
	component := detectStringComponent(passwordField)
	assert.Equal(t, "Password", component)

	// Test select field detection
	selectField := &Field{
		Name: "category",
		Type: "string",
		Options: []Option{
			{Value: "a", Label: "A"},
			{Value: "b", Label: "B"},
		},
	}
	component = detectStringComponent(selectField)
	assert.Equal(t, "Select", component)

	// Test rich text field detection
	richTextField := &Field{
		Name:     "description",
		Type:     "string",
		RichText: &RichTextConfig{},
	}
	component = detectStringComponent(richTextField)
	assert.Equal(t, "TextArea", component)

	// Test regular string field
	regularField := &Field{
		Name: "username",
		Type: "string",
	}
	component = detectStringComponent(regularField)
	assert.Equal(t, "Input", component)
}

func TestDetectDateTimeComponent(t *testing.T) {
	// Test time field
	component := detectDateTimeComponent("time")
	assert.Equal(t, "TimePicker", component)

	// Test date field
	component = detectDateTimeComponent("date")
	assert.Equal(t, "DatePicker", component)

	// Test datetime field
	component = detectDateTimeComponent("datetime")
	assert.Equal(t, "DatePicker", component)

	// Test unknown field
	component = detectDateTimeComponent("unknown")
	assert.Equal(t, "DatePicker", component)
}

func TestDetectFileComponent(t *testing.T) {
	// Test image file
	imageField := &Field{
		Name: "avatar",
		Type: "file",
		File: &FileConfig{
			IsImage: true,
		},
	}
	component := detectFileComponent(imageField)
	assert.Equal(t, "Upload.Image", component)

	// Test regular file
	fileField := &Field{
		Name: "document",
		Type: "file",
		File: &FileConfig{
			IsImage: false,
		},
	}
	component = detectFileComponent(fileField)
	assert.Equal(t, "Upload", component)

	// Test without file config
	noConfigField := &Field{
		Name: "attachment",
		Type: "file",
	}
	component = detectFileComponent(noConfigField)
	assert.Equal(t, "Upload", component)
}

func TestDetectSelectComponent(t *testing.T) {
	// Test select field
	component := detectSelectComponent("select")
	assert.Equal(t, "Select", component)

	// Test multiselect field
	component = detectSelectComponent("multiselect")
	assert.Equal(t, "Select", component)
}

func TestValidateProperty(t *testing.T) {
	// Create test data
	dataMap := map[string]interface{}{
		"name": "John",
		"age":  30,
		"address": map[string]interface{}{
			"city":    "New York",
			"zipcode": "10001",
		},
		"tags": []string{"tag1", "tag2"},
	}

	// Test case: property without validation
	propNoValidation := &JsonProperty{
		Path: "name",
		Type: "string",
	}
	errors := validateProperty(dataMap, propNoValidation)
	assert.Empty(t, errors)

	// Test case: required property present
	propRequired := &JsonProperty{
		Path: "name",
		Type: "string",
		Validation: &JsonValidation{
			Required: true,
		},
	}
	errors = validateProperty(dataMap, propRequired)
	assert.Empty(t, errors)

	// Test case: required property missing
	propMissing := &JsonProperty{
		Path: "missing.field",
		Type: "string",
		Validation: &JsonValidation{
			Required: true,
		},
	}
	errors = validateProperty(dataMap, propMissing)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "not found")

	// Test case: required property is nil
	propNilValue := &JsonProperty{
		Path: "nilValue",
		Type: "string",
		Validation: &JsonValidation{
			Required: true,
		},
	}
	dataMapWithNil := map[string]interface{}{
		"nilValue": nil,
	}
	errors = validateProperty(dataMapWithNil, propNilValue)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors[0], "required but has nil value")

	// Test case: non-required nil value
	propOptionalNil := &JsonProperty{
		Path: "nilValue",
		Type: "string",
		Validation: &JsonValidation{
			Required: false,
		},
	}
	errors = validateProperty(dataMapWithNil, propOptionalNil)
	assert.Empty(t, errors)

	// Test different property types

	// Numeric property
	propNumeric := &JsonProperty{
		Path: "age",
		Type: "number",
		Validation: &JsonValidation{
			Min: 18,
			Max: 100,
		},
	}
	errors = validateProperty(dataMap, propNumeric)
	assert.Empty(t, errors)

	// String property
	propString := &JsonProperty{
		Path: "name",
		Type: "string",
		Validation: &JsonValidation{
			MinLength: 3,
		},
	}
	errors = validateProperty(dataMap, propString)
	assert.Empty(t, errors)

	// Object property
	propObject := &JsonProperty{
		Path: "address",
		Type: "object",
		Properties: []JsonProperty{
			{
				Path: "city",
				Type: "string",
				Validation: &JsonValidation{
					Required: true,
				},
			},
		},
	}
	errors = validateProperty(dataMap, propObject)
	assert.Empty(t, errors)

	// Array property
	propArray := &JsonProperty{
		Path: "tags",
		Type: "array",
		Validation: &JsonValidation{
			MinLength: 1,
		},
	}
	errors = validateProperty(dataMap, propArray)
	assert.Empty(t, errors)
}

func TestConvertToNumber(t *testing.T) {
	// Test float64 input
	num, err := convertToNumber(float64(123.45))
	assert.NoError(t, err)
	assert.Equal(t, float64(123.45), num)

	// Test int input
	num, err = convertToNumber(42)
	assert.NoError(t, err)
	assert.Equal(t, float64(42), num)

	// Test int64 input
	num, err = convertToNumber(int64(42))
	assert.NoError(t, err)
	assert.Equal(t, float64(42), num)

	// Test float32 input
	num, err = convertToNumber(float32(42.5))
	assert.NoError(t, err)
	assert.Equal(t, float64(42.5), num)

	// Test string input with valid number
	num, err = convertToNumber("789.01")
	assert.NoError(t, err)
	assert.Equal(t, float64(789.01), num)

	// Test string input with integer
	num, err = convertToNumber("42")
	assert.NoError(t, err)
	assert.Equal(t, float64(42), num)

	// Test string input with invalid number
	_, err = convertToNumber("not-a-number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "should be a number")

	// Test boolean input
	_, err = convertToNumber(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "should be a number")

	// Test nil input
	_, err = convertToNumber(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "should be a number")

	// Test with unsupported numeric types
	_, err = convertToNumber(int32(42))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "should be a number")

	_, err = convertToNumber(uint(42))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "should be a number")

	_, err = convertToNumber(uint64(42))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "should be a number")
}

func TestDetectTextComponent(t *testing.T) {
	// Test with rich text config
	richTextField := &Field{
		Name:     "description",
		Type:     "text",
		RichText: &RichTextConfig{},
	}
	component := detectTextComponent(richTextField)
	assert.Equal(t, "TextArea", component)

	// Test without rich text config
	regularTextField := &Field{
		Name: "content",
		Type: "text",
	}
	component = detectTextComponent(regularTextField)
	assert.Equal(t, "TextArea", component)
}

func TestIsSlice(t *testing.T) {
	// Test with slice
	slice := []int{1, 2, 3}
	assert.True(t, IsSlice(slice))

	// Test with pointer to slice
	slicePtr := &slice
	assert.True(t, IsSlice(slicePtr))

	// Test with non-slice
	nonSlice := "string"
	assert.False(t, IsSlice(nonSlice))

	// Test with pointer to non-slice
	nonSlicePtr := &nonSlice
	assert.False(t, IsSlice(nonSlicePtr))

	// Test with nil
	assert.False(t, IsSlice(nil))

	// Test with empty slice
	emptySlice := []string{}
	assert.True(t, IsSlice(emptySlice))

	// Test with map
	mapValue := map[string]int{"one": 1}
	assert.False(t, IsSlice(mapValue))

	// Test with struct
	type TestStruct struct{}
	structValue := TestStruct{}
	assert.False(t, IsSlice(structValue))
}

func TestSetFieldValue(t *testing.T) {
	// Test struct
	type TestStruct struct {
		Name       string
		Age        int
		IsActive   bool
		Score      float64
		unexported string // unexported field
	}

	// Create test instance
	test := TestStruct{
		Name:       "Initial",
		Age:        30,
		IsActive:   false,
		Score:      75.5,
		unexported: "hidden",
	}

	// Test setting string field
	err := SetFieldValue(&test, "Name", "Updated")
	assert.NoError(t, err)
	assert.Equal(t, "Updated", test.Name)

	// Test setting int field
	err = SetFieldValue(&test, "Age", 40)
	assert.NoError(t, err)
	assert.Equal(t, 40, test.Age)

	// Test setting bool field
	err = SetFieldValue(&test, "IsActive", true)
	assert.NoError(t, err)
	assert.True(t, test.IsActive)

	// Test setting float field
	err = SetFieldValue(&test, "Score", 90.5)
	assert.NoError(t, err)
	assert.Equal(t, 90.5, test.Score)

	// Test with non-existent field
	err = SetFieldValue(&test, "NonExistent", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test with unexported field
	err = SetFieldValue(&test, "unexported", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be set")

	// Test with nil object
	err = SetFieldValue(nil, "Name", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Test with type conversion
	err = SetFieldValue(&test, "Age", int64(50))
	assert.NoError(t, err)
	assert.Equal(t, 50, test.Age)

	// Test with incompatible type
	err = SetFieldValue(&test, "Age", "not a number")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be converted")
}

func TestGetFieldValue(t *testing.T) {
	// Test struct
	type TestStruct struct {
		Name      string
		Age       int
		IsActive  bool
		Score     float64
		NestedPtr *TestStruct
	}

	// Create test instance
	nested := TestStruct{
		Name: "Nested",
	}

	test := TestStruct{
		Name:      "Test",
		Age:       30,
		IsActive:  true,
		Score:     75.5,
		NestedPtr: &nested,
	}

	// Test getting string field
	value, err := GetFieldValue(test, "Name")
	assert.NoError(t, err)
	assert.Equal(t, "Test", value)

	// Test getting int field
	value, err = GetFieldValue(test, "Age")
	assert.NoError(t, err)
	assert.Equal(t, 30, value)

	// Test getting bool field
	value, err = GetFieldValue(test, "IsActive")
	assert.NoError(t, err)
	assert.Equal(t, true, value)

	// Test getting float field
	value, err = GetFieldValue(test, "Score")
	assert.NoError(t, err)
	assert.Equal(t, 75.5, value)

	// Test with pointer
	value, err = GetFieldValue(&test, "Name")
	assert.NoError(t, err)
	assert.Equal(t, "Test", value)

	// Test with pointer to nested struct
	value, err = GetFieldValue(test, "NestedPtr")
	assert.NoError(t, err)
	assert.Equal(t, &nested, value)

	// Test with non-existent field
	_, err = GetFieldValue(test, "NonExistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test with nil object
	_, err = GetFieldValue(nil, "Name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Test with nil struct pointer
	var nilPtr *TestStruct
	_, err = GetFieldValue(nilPtr, "Name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not valid")
}
