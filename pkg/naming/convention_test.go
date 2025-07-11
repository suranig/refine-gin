package naming

import (
	"reflect"
	"testing"
)

func TestNamingConventionConstants(t *testing.T) {
	tests := []struct {
		name string
		nc   NamingConvention
		want string
	}{
		{"CamelCase", CamelCase, "camelCase"},
		{"SnakeCase", SnakeCase, "snake_case"},
		{"PascalCase", PascalCase, "PascalCase"},
		{"KebabCase", KebabCase, "kebab-case"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.nc) != tt.want {
				t.Errorf("NamingConvention = %v, want %v", tt.nc, tt.want)
			}
		})
	}
}

func TestDefaultRefineConfig(t *testing.T) {
	config := DefaultRefineConfig()

	if config.NamingConvention != CamelCase {
		t.Errorf("DefaultRefineConfig().NamingConvention = %v, want %v", config.NamingConvention, CamelCase)
	}

	if config.IDFieldName != "id" {
		t.Errorf("DefaultRefineConfig().IDFieldName = %v, want %v", config.IDFieldName, "id")
	}

	if config.EnableRealTime != false {
		t.Errorf("DefaultRefineConfig().EnableRealTime = %v, want %v", config.EnableRealTime, false)
	}

	if config.FieldMappings == nil {
		t.Error("DefaultRefineConfig().FieldMappings should not be nil")
	}
}

func TestRefineConfig_WithNamingConvention(t *testing.T) {
	config := DefaultRefineConfig().WithNamingConvention(SnakeCase)

	if config.NamingConvention != SnakeCase {
		t.Errorf("WithNamingConvention() = %v, want %v", config.NamingConvention, SnakeCase)
	}
}

func TestRefineConfig_WithIDFieldName(t *testing.T) {
	config := DefaultRefineConfig().WithIDFieldName("uuid")

	if config.IDFieldName != "uuid" {
		t.Errorf("WithIDFieldName() = %v, want %v", config.IDFieldName, "uuid")
	}
}

func TestRefineConfig_WithFieldMapping(t *testing.T) {
	config := DefaultRefineConfig().WithFieldMapping("old_name", "newName")

	if config.FieldMappings["old_name"] != "newName" {
		t.Errorf("WithFieldMapping() = %v, want %v", config.FieldMappings["old_name"], "newName")
	}
}

func TestRefineConfig_ConvertFieldName(t *testing.T) {
	tests := []struct {
		name           string
		convention     NamingConvention
		input          string
		fieldMappings  map[string]string
		expectedOutput string
	}{
		// CamelCase tests
		{"camelCase basic", CamelCase, "first_name", nil, "firstName"},
		{"camelCase with underscore", CamelCase, "user_id", nil, "userId"},
		{"camelCase with hyphen", CamelCase, "last-name", nil, "lastName"},
		{"camelCase with dot", CamelCase, "email.address", nil, "emailAddress"},
		{"camelCase already camel", CamelCase, "firstName", nil, "firstName"},
		{"camelCase empty", CamelCase, "", nil, ""},

		// SnakeCase tests
		{"snakeCase basic", SnakeCase, "firstName", nil, "first_name"},
		{"snakeCase with underscore", SnakeCase, "user_id", nil, "user_id"},
		{"snakeCase with hyphen", SnakeCase, "last-name", nil, "last-name"},
		{"snakeCase with dot", SnakeCase, "email.address", nil, "email.address"},
		{"snakeCase empty", SnakeCase, "", nil, ""},

		// PascalCase tests
		{"pascalCase basic", PascalCase, "first_name", nil, "FirstName"},
		{"pascalCase with underscore", PascalCase, "user_id", nil, "UserId"},
		{"pascalCase with hyphen", PascalCase, "last-name", nil, "LastName"},
		{"pascalCase with dot", PascalCase, "email.address", nil, "EmailAddress"},
		{"pascalCase empty", PascalCase, "", nil, ""},

		// KebabCase tests
		{"kebabCase basic", KebabCase, "firstName", nil, "first-name"},
		{"kebabCase with underscore", KebabCase, "user_id", nil, "user_id"},
		{"kebabCase with hyphen", KebabCase, "last-name", nil, "last-name"},
		{"kebabCase with dot", KebabCase, "email.address", nil, "email.address"},
		{"kebabCase empty", KebabCase, "", nil, ""},

		// Custom field mapping tests
		{"custom mapping", CamelCase, "old_name", map[string]string{"old_name": "newName"}, "newName"},
		{"custom mapping override", SnakeCase, "firstName", map[string]string{"firstName": "first_name"}, "first_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultRefineConfig().WithNamingConvention(tt.convention)
			if tt.fieldMappings != nil {
				config.FieldMappings = tt.fieldMappings
			}

			result := config.ConvertFieldName(tt.input)
			if result != tt.expectedOutput {
				t.Errorf("ConvertFieldName() = %v, want %v", result, tt.expectedOutput)
			}
		})
	}
}

func TestRefineConfig_ConvertFieldNames(t *testing.T) {
	config := DefaultRefineConfig().WithNamingConvention(CamelCase)

	input := []string{"first_name", "last_name", "email_address"}
	expected := []string{"firstName", "lastName", "emailAddress"}

	result := config.ConvertFieldNames(input)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ConvertFieldNames() = %v, want %v", result, expected)
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"basic", "first_name", "firstName"},
		{"with hyphen", "last-name", "lastName"},
		{"with dot", "email.address", "emailAddress"},
		{"already camel", "firstName", "firstName"},
		{"mixed separators", "user_id-email", "userIdEmail"},
		{"empty", "", ""},
		{"single word", "name", "name"},
		{"uppercase", "FIRST_NAME", "firstName"},
		{"mixed case", "First_Name", "firstName"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"basic", "firstName", "first_name"},
		{"multiple words", "firstNameLastName", "first_name_last_name"},
		{"already snake", "first_name", "first_name"},
		{"with numbers", "user123Id", "user123_id"},
		{"empty", "", ""},
		{"single word", "name", "name"},
		{"all uppercase", "FIRSTNAME", "firstname"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"basic", "first_name", "FirstName"},
		{"with hyphen", "last-name", "LastName"},
		{"with dot", "email.address", "EmailAddress"},
		{"already pascal", "FirstName", "FirstName"},
		{"mixed separators", "user_id-email", "UserIdEmail"},
		{"empty", "", ""},
		{"single word", "name", "Name"},
		{"uppercase", "FIRST_NAME", "FirstName"},
		{"mixed case", "First_Name", "FirstName"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"basic", "firstName", "first-name"},
		{"multiple words", "firstNameLastName", "first-name-last-name"},
		{"already kebab", "first-name", "first-name"},
		{"with numbers", "user123Id", "user123-id"},
		{"empty", "", ""},
		{"single word", "name", "name"},
		{"all uppercase", "FIRSTNAME", "firstname"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toKebabCase(tt.input)
			if result != tt.expected {
				t.Errorf("toKebabCase() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSplitBySeparators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"underscore", "first_name", []string{"first", "name"}},
		{"hyphen", "last-name", []string{"last", "name"}},
		{"dot", "email.address", []string{"email", "address"}},
		{"mixed", "user_id-email", []string{"user", "id", "email"}},
		{"multiple spaces", "first  name", []string{"first", "name"}},
		{"empty", "", []string{}},
		{"single word", "name", []string{"name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitBySeparators(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("splitBySeparators() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidNamingConvention(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"camelCase", "camelCase", true},
		{"snake_case", "snake_case", true},
		{"PascalCase", "PascalCase", true},
		{"kebab-case", "kebab-case", true},
		{"invalid", "invalid", false},
		{"empty", "", false},
		{"case insensitive", "CAMELCASE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidNamingConvention(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidNamingConvention() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetNamingConventionFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected NamingConvention
	}{
		{"camelcase", "camelcase", CamelCase},
		{"camel_case", "camel_case", CamelCase},
		{"snakecase", "snakecase", SnakeCase},
		{"snake_case", "snake_case", SnakeCase},
		{"pascalcase", "pascalcase", PascalCase},
		{"pascal_case", "pascal_case", PascalCase},
		{"kebabcase", "kebabcase", KebabCase},
		{"kebab_case", "kebab_case", KebabCase},
		{"kebab-case", "kebab-case", KebabCase},
		{"invalid", "invalid", CamelCase}, // Default fallback
		{"empty", "", CamelCase},          // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNamingConventionFromString(tt.input)
			if result != tt.expected {
				t.Errorf("GetNamingConventionFromString() = %v, want %v", result, tt.expected)
			}
		})
	}
}