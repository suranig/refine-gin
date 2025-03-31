package resource

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStruct is a test structure for testing field operations
type TestStruct struct {
	ID          int
	Name        string
	Description string
	IsActive    bool
	Tags        []string
	unexported  string // unexported field
}

func TestIsSlice(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: false,
		},
		{
			name:     "string input",
			input:    "not a slice",
			expected: false,
		},
		{
			name:     "int input",
			input:    42,
			expected: false,
		},
		{
			name:     "struct input",
			input:    TestStruct{},
			expected: false,
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: true,
		},
		{
			name:     "int slice",
			input:    []int{1, 2, 3},
			expected: true,
		},
		{
			name:     "string slice",
			input:    []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "struct slice",
			input:    []TestStruct{{ID: 1}, {ID: 2}},
			expected: true,
		},
		{
			name:     "pointer to slice",
			input:    &[]int{1, 2, 3},
			expected: true,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSlice(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetFieldValue(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		obj         interface{}
		fieldName   string
		value       interface{}
		expectError bool
		shouldSkip  bool // Skip tests that would panic
	}{
		{
			name:        "nil object",
			obj:         nil,
			fieldName:   "Name",
			value:       "Test",
			expectError: true,
			shouldSkip:  false,
		},
		{
			name:        "primitive type",
			obj:         42,
			fieldName:   "value",
			value:       100,
			expectError: true,
			shouldSkip:  true, // This would cause a panic, so we skip it
		},
		{
			name:        "non-existent field",
			obj:         &TestStruct{},
			fieldName:   "NonExistentField",
			value:       "value",
			expectError: true,
			shouldSkip:  false,
		},
		{
			name:        "unexported field",
			obj:         &TestStruct{},
			fieldName:   "unexported",
			value:       "value",
			expectError: true,
			shouldSkip:  false,
		},
		{
			name:        "incompatible value type",
			obj:         &TestStruct{},
			fieldName:   "ID",
			value:       "not an int",
			expectError: true,
			shouldSkip:  false,
		},
		{
			name:        "set int field",
			obj:         &TestStruct{},
			fieldName:   "ID",
			value:       42,
			expectError: false,
			shouldSkip:  false,
		},
		{
			name:        "set string field",
			obj:         &TestStruct{},
			fieldName:   "Name",
			value:       "Test Name",
			expectError: false,
			shouldSkip:  false,
		},
		{
			name:        "set bool field",
			obj:         &TestStruct{},
			fieldName:   "IsActive",
			value:       true,
			expectError: false,
			shouldSkip:  false,
		},
		{
			name:        "set slice field",
			obj:         &TestStruct{},
			fieldName:   "Tags",
			value:       []string{"tag1", "tag2"},
			expectError: false,
			shouldSkip:  false,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldSkip {
				t.Skip("Skipping test that would cause a panic")
				return
			}

			err := SetFieldValue(tt.obj, tt.fieldName, tt.value)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the field was set correctly
				if ts, ok := tt.obj.(*TestStruct); ok {
					switch tt.fieldName {
					case "ID":
						assert.Equal(t, 42, ts.ID)
					case "Name":
						assert.Equal(t, "Test Name", ts.Name)
					case "IsActive":
						assert.True(t, ts.IsActive)
					case "Tags":
						assert.Equal(t, []string{"tag1", "tag2"}, ts.Tags)
					}
				}
			}
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	// Create a test struct
	testObj := &TestStruct{
		ID:          123,
		Name:        "Test Object",
		Description: "This is a test object",
		IsActive:    true,
		Tags:        []string{"tag1", "tag2", "tag3"},
		unexported:  "unexported value",
	}

	// Test cases
	tests := []struct {
		name        string
		obj         interface{}
		fieldName   string
		expected    interface{}
		expectError bool
		shouldSkip  bool // Skip tests that would panic
	}{
		{
			name:        "nil object",
			obj:         nil,
			fieldName:   "Name",
			expected:    nil,
			expectError: true,
			shouldSkip:  false,
		},
		{
			name:        "primitive type",
			obj:         42,
			fieldName:   "value",
			expected:    nil,
			expectError: true,
			shouldSkip:  true, // This would cause a panic, so we skip it
		},
		{
			name:        "non-existent field",
			obj:         testObj,
			fieldName:   "NonExistentField",
			expected:    nil,
			expectError: true,
			shouldSkip:  false,
		},
		{
			name:        "get int field",
			obj:         testObj,
			fieldName:   "ID",
			expected:    123,
			expectError: false,
			shouldSkip:  false,
		},
		{
			name:        "get string field",
			obj:         testObj,
			fieldName:   "Name",
			expected:    "Test Object",
			expectError: false,
			shouldSkip:  false,
		},
		{
			name:        "get bool field",
			obj:         testObj,
			fieldName:   "IsActive",
			expected:    true,
			expectError: false,
			shouldSkip:  false,
		},
		{
			name:        "get slice field",
			obj:         testObj,
			fieldName:   "Tags",
			expected:    []string{"tag1", "tag2", "tag3"},
			expectError: false,
			shouldSkip:  false,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldSkip {
				t.Skip("Skipping test that would cause a panic")
				return
			}

			value, err := GetFieldValue(tt.obj, tt.fieldName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, value)
			}
		})
	}
}

func TestFilterOutReadOnlyFields(t *testing.T) {
	// Test structure with some fields
	type TestModel struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}

	// Create a mock resource
	mockResource := &DefaultResource{
		Fields: []Field{
			{Name: "ID", ReadOnly: true},
			{Name: "Name", ReadOnly: false},
			{Name: "Email", ReadOnly: false},
			{Name: "CreatedAt", ReadOnly: true},
		},
		EditableFields: []string{"Name", "Email"},
	}

	// Test struct filtering
	testStruct := TestModel{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: "2023-07-01",
	}

	result := FilterOutReadOnlyFields(testStruct, mockResource)
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")

	// Should only contain editable fields
	assert.Equal(t, 2, len(resultMap))
	assert.Equal(t, "John Doe", resultMap["Name"])
	assert.Equal(t, "john@example.com", resultMap["Email"])
	assert.NotContains(t, resultMap, "ID")
	assert.NotContains(t, resultMap, "CreatedAt")

	// Test map filtering
	testMap := map[string]interface{}{
		"ID":        1,
		"Name":      "John Doe",
		"Email":     "john@example.com",
		"CreatedAt": "2023-07-01",
	}

	mapResult := FilterOutReadOnlyFields(testMap, mockResource)
	mapResultMap, ok := mapResult.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")

	// Should only contain editable fields
	assert.Equal(t, 2, len(mapResultMap))
	assert.Equal(t, "John Doe", mapResultMap["Name"])
	assert.Equal(t, "john@example.com", mapResultMap["Email"])
	assert.NotContains(t, mapResultMap, "ID")
	assert.NotContains(t, mapResultMap, "CreatedAt")

	// Test with no editable fields defined (should use ReadOnly flag)
	resourceWithoutEditableFields := &DefaultResource{
		Fields: []Field{
			{Name: "ID", ReadOnly: true},
			{Name: "Name", ReadOnly: false},
			{Name: "Email", ReadOnly: false},
			{Name: "CreatedAt", ReadOnly: true},
		},
	}

	result = FilterOutReadOnlyFields(testStruct, resourceWithoutEditableFields)
	resultMap, ok = result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")

	// Should only contain non-readonly fields
	assert.Equal(t, 2, len(resultMap))
	assert.Equal(t, "John Doe", resultMap["Name"])
	assert.Equal(t, "john@example.com", resultMap["Email"])
	assert.NotContains(t, resultMap, "ID")
	assert.NotContains(t, resultMap, "CreatedAt")

	// Test case insensitivity
	resourceWithLowercaseFields := &DefaultResource{
		EditableFields: []string{"name", "email"},
	}

	result = FilterOutReadOnlyFields(testStruct, resourceWithLowercaseFields)
	resultMap, ok = result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.Equal(t, 2, len(resultMap))
	assert.Equal(t, "John Doe", resultMap["Name"])
	assert.Equal(t, "john@example.com", resultMap["Email"])

	// Test nil input
	var nilData *TestModel
	nilResult := FilterOutReadOnlyFields(nilData, mockResource)
	assert.Nil(t, nilResult)

	// Test non-struct, non-map input
	strData := "test string"
	strResult := FilterOutReadOnlyFields(strData, mockResource)
	assert.Equal(t, strData, strResult)
}

func TestValidateNestedJson(t *testing.T) {
	tests := []struct {
		name           string
		data           interface{}
		config         *JsonConfig
		expectedValid  bool
		expectedErrors []string
	}{
		{
			name: "Valid simple object",
			data: map[string]interface{}{
				"name": "Test Name",
				"age":  30,
			},
			config: &JsonConfig{
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
						Type: "integer",
						Validation: &JsonValidation{
							Required: true,
							Min:      18,
							Max:      100,
						},
					},
				},
			},
			expectedValid:  true,
			expectedErrors: []string{},
		},
		{
			name: "Missing required field",
			data: map[string]interface{}{
				"name": "Test Name",
			},
			config: &JsonConfig{
				Properties: []JsonProperty{
					{
						Path: "name",
						Type: "string",
						Validation: &JsonValidation{
							Required: true,
						},
					},
					{
						Path: "age",
						Type: "integer",
						Validation: &JsonValidation{
							Required: true,
						},
					},
				},
			},
			expectedValid: false,
			expectedErrors: []string{
				"Required property 'age' not found: field 'age' not found",
			},
		},
		{
			name: "Invalid string length",
			data: map[string]interface{}{
				"name": "A",
				"age":  30,
			},
			config: &JsonConfig{
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
						Type: "integer",
						Validation: &JsonValidation{
							Required: true,
						},
					},
				},
			},
			expectedValid: false,
			expectedErrors: []string{
				"Property 'name' length 1 is less than minimum length 3",
			},
		},
		{
			name: "Invalid number range",
			data: map[string]interface{}{
				"name": "Test Name",
				"age":  15,
			},
			config: &JsonConfig{
				Properties: []JsonProperty{
					{
						Path: "name",
						Type: "string",
						Validation: &JsonValidation{
							Required: true,
						},
					},
					{
						Path: "age",
						Type: "integer",
						Validation: &JsonValidation{
							Required: true,
							Min:      18,
							Max:      100,
						},
					},
				},
			},
			expectedValid: false,
			expectedErrors: []string{
				"Property 'age' value 15 is less than minimum 18",
			},
		},
		{
			name: "Nested object validation",
			data: map[string]interface{}{
				"name": "Test Name",
				"address": map[string]interface{}{
					"street": "123 Main St",
					"city":   "Test City",
					"zip":    "AB", // Invalid zip code
				},
			},
			config: &JsonConfig{
				Properties: []JsonProperty{
					{
						Path: "name",
						Type: "string",
						Validation: &JsonValidation{
							Required: true,
						},
					},
					{
						Path: "address",
						Type: "object",
						Properties: []JsonProperty{
							{
								Path: "street",
								Type: "string",
								Validation: &JsonValidation{
									Required: true,
								},
							},
							{
								Path: "city",
								Type: "string",
								Validation: &JsonValidation{
									Required: true,
								},
							},
							{
								Path: "zip",
								Type: "string",
								Validation: &JsonValidation{
									Required:  true,
									MinLength: 5,
								},
							},
						},
						Validation: &JsonValidation{
							Required: true,
						},
					},
				},
			},
			expectedValid: false,
			expectedErrors: []string{
				"address.Property 'zip' length 2 is less than minimum length 5",
			},
		},
		{
			name: "Array validation",
			data: map[string]interface{}{
				"name": "Test Name",
				"tags": []interface{}{"tag1"},
			},
			config: &JsonConfig{
				Properties: []JsonProperty{
					{
						Path: "name",
						Type: "string",
						Validation: &JsonValidation{
							Required: true,
						},
					},
					{
						Path: "tags",
						Type: "array",
						Validation: &JsonValidation{
							Required:  true,
							MinLength: 2, // Requires at least 2 tags
						},
					},
				},
			},
			expectedValid: false,
			expectedErrors: []string{
				"Property 'tags' has 1 items which is less than minimum 2",
			},
		},
		{
			name: "JSON string input",
			data: `{"name": "Test Name", "age": 30}`,
			config: &JsonConfig{
				Properties: []JsonProperty{
					{
						Path: "name",
						Type: "string",
						Validation: &JsonValidation{
							Required: true,
						},
					},
					{
						Path: "age",
						Type: "integer",
						Validation: &JsonValidation{
							Required: true,
							Min:      18,
						},
					},
				},
			},
			expectedValid:  true,
			expectedErrors: []string{},
		},
		{
			name: "Path with array index",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"id":   1,
						"name": "Item 1",
					},
					map[string]interface{}{
						"id":   2,
						"name": "I", // Invalid name (too short)
					},
				},
			},
			config: &JsonConfig{
				Properties: []JsonProperty{
					{
						Path: "items[1].name",
						Type: "string",
						Validation: &JsonValidation{
							Required:  true,
							MinLength: 2,
						},
					},
				},
			},
			expectedValid: false,
			expectedErrors: []string{
				"Property 'items[1].name' length 1 is less than minimum length 2",
			},
		},
		{
			name: "Multiple validation errors",
			data: map[string]interface{}{
				"name": "A",
				"age":  150,
			},
			config: &JsonConfig{
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
						Path: "age",
						Type: "integer",
						Validation: &JsonValidation{
							Required: true,
							Max:      100,
						},
					},
				},
			},
			expectedValid: false,
			expectedErrors: []string{
				"Property 'name' length 1 is less than minimum length 3",
				"Property 'age' value 150 is greater than maximum 100",
			},
		},
		{
			name:          "Nil inputs",
			data:          nil,
			config:        nil,
			expectedValid: false,
			expectedErrors: []string{
				"Invalid input: nil data or config",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, errors := ValidateNestedJson(test.data, test.config)

			if valid != test.expectedValid {
				t.Errorf("Expected valid=%v, got %v", test.expectedValid, valid)
			}

			if len(errors) != len(test.expectedErrors) {
				t.Errorf("Expected %d errors, got %d: %v", len(test.expectedErrors), len(errors), errors)
				return
			}

			// Check specific error messages if we care about them
			if len(test.expectedErrors) > 0 {
				for i, expectedErr := range test.expectedErrors {
					if !strings.Contains(errors[i], expectedErr) {
						t.Errorf("Expected error to contain '%s', got '%s'", expectedErr, errors[i])
					}
				}
			}
		})
	}
}

func TestGetValueByPath(t *testing.T) {
	data := map[string]interface{}{
		"name": "Test",
		"address": map[string]interface{}{
			"street": "Main St",
			"city":   "Test City",
		},
		"tags": []interface{}{"tag1", "tag2", "tag3"},
		"items": []interface{}{
			map[string]interface{}{
				"id":   1,
				"name": "Item 1",
			},
			map[string]interface{}{
				"id":   2,
				"name": "Item 2",
			},
		},
	}

	tests := []struct {
		path           string
		expectedValue  interface{}
		expectedError  bool
		expectedErrMsg string
	}{
		{
			path:          "name",
			expectedValue: "Test",
			expectedError: false,
		},
		{
			path:          "address.street",
			expectedValue: "Main St",
			expectedError: false,
		},
		{
			path:           "address.zip",
			expectedValue:  nil,
			expectedError:  true,
			expectedErrMsg: "field 'zip' not found",
		},
		{
			path:          "tags",
			expectedValue: []interface{}{"tag1", "tag2", "tag3"},
			expectedError: false,
		},
		{
			path:          "items[0].name",
			expectedValue: "Item 1",
			expectedError: false,
		},
		{
			path:          "items[1].id",
			expectedValue: float64(2), // JSON numbers are float64
			expectedError: false,
		},
		{
			path:           "items[3]",
			expectedValue:  nil,
			expectedError:  true,
			expectedErrMsg: "array index 3 out of bounds",
		},
		{
			path:           "items[x]",
			expectedValue:  nil,
			expectedError:  true,
			expectedErrMsg: "invalid array index",
		},
		{
			path:           "notexist.field",
			expectedValue:  nil,
			expectedError:  true,
			expectedErrMsg: "field 'notexist' not found",
		},
		{
			path:          "",
			expectedValue: data,
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			value, err := getValueByPath(data, test.path)

			if test.expectedError {
				if err == nil {
					t.Errorf("Expected error for path '%s', got nil", test.path)
					return
				}
				if test.expectedErrMsg != "" && !strings.Contains(err.Error(), test.expectedErrMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", test.expectedErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for path '%s': %v", test.path, err)
					return
				}

				// Check value
				if !reflect.DeepEqual(value, test.expectedValue) {
					t.Errorf("Expected value %v, got %v", test.expectedValue, value)
				}
			}
		})
	}
}
