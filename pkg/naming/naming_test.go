package naming

import (
	"reflect"
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"helloWorld", "hello_world"},
		{"HelloWorld", "hello_world"},
		{"hello_world", "hello_world"},
		{"Hello", "hello"},
		{"hello", "hello"},
		{"HTTPRequest", "http_request"},
		{"", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ToSnakeCase(test.input)
			if result != test.expected {
				t.Errorf("ToSnakeCase(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "helloWorld"},
		{"HelloWorld", "helloWorld"},
		{"hello_world_test", "helloWorldTest"},
		{"Hello", "hello"},
		{"hello", "hello"},
		{"HTTP_REQUEST", "httpRequest"},
		{"", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ToCamelCase(test.input)
			if result != test.expected {
				t.Errorf("ToCamelCase(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "HelloWorld"},
		{"HelloWorld", "HelloWorld"},
		{"hello_world_test", "HelloWorldTest"},
		{"Hello", "Hello"},
		{"hello", "Hello"},
		{"HTTP_REQUEST", "HttpRequest"},
		{"", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ToPascalCase(test.input)
			if result != test.expected {
				t.Errorf("ToPascalCase(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

func TestConvertKeys(t *testing.T) {
	// Test data
	data := map[string]interface{}{
		"userId":       1,
		"firstName":    "John",
		"last_name":    "Doe",
		"EmailAddress": "john.doe@example.com",
		"nestedObject": map[string]interface{}{
			"itemId":      123,
			"item_name":   "Test Item",
			"ItemDetails": "Some details",
		},
		"itemList": []interface{}{
			map[string]interface{}{
				"listItemId":     1,
				"list_item_name": "Item 1",
			},
			map[string]interface{}{
				"listItemId":     2,
				"list_item_name": "Item 2",
			},
		},
	}

	// Test snake_case conversion
	t.Run("SnakeCase", func(t *testing.T) {
		result := ConvertKeys(data, SnakeCase)

		// Check top-level keys
		assertKeyExists(t, result, "user_id")
		assertKeyExists(t, result, "first_name")
		assertKeyExists(t, result, "last_name")
		assertKeyExists(t, result, "email_address")

		// Check nested object
		nestedObj, ok := result["nested_object"].(map[string]interface{})
		if !ok {
			t.Fatal("nested_object is not a map")
		}
		assertKeyExists(t, nestedObj, "item_id")
		assertKeyExists(t, nestedObj, "item_name")
		assertKeyExists(t, nestedObj, "item_details")

		// Check array of objects
		itemList, ok := result["item_list"].([]interface{})
		if !ok {
			t.Fatal("item_list is not an array")
		}

		item0, ok := itemList[0].(map[string]interface{})
		if !ok {
			t.Fatal("itemList[0] is not a map")
		}
		assertKeyExists(t, item0, "list_item_id")
		assertKeyExists(t, item0, "list_item_name")
	})

	// Test camelCase conversion
	t.Run("CamelCase", func(t *testing.T) {
		result := ConvertKeys(data, CamelCase)

		// Check top-level keys
		assertKeyExists(t, result, "userId")
		assertKeyExists(t, result, "firstName")
		assertKeyExists(t, result, "lastName")
		assertKeyExists(t, result, "emailAddress")

		// Check nested object
		nestedObj, ok := result["nestedObject"].(map[string]interface{})
		if !ok {
			t.Fatal("nestedObject is not a map")
		}
		assertKeyExists(t, nestedObj, "itemId")
		assertKeyExists(t, nestedObj, "itemName")
		assertKeyExists(t, nestedObj, "itemDetails")
	})

	// Test PascalCase conversion
	t.Run("PascalCase", func(t *testing.T) {
		result := ConvertKeys(data, PascalCase)

		// Check top-level keys
		assertKeyExists(t, result, "UserId")
		assertKeyExists(t, result, "FirstName")
		assertKeyExists(t, result, "LastName")
		assertKeyExists(t, result, "EmailAddress")

		// Check nested object
		nestedObj, ok := result["NestedObject"].(map[string]interface{})
		if !ok {
			t.Fatal("NestedObject is not a map")
		}
		assertKeyExists(t, nestedObj, "ItemId")
		assertKeyExists(t, nestedObj, "ItemName")
		assertKeyExists(t, nestedObj, "ItemDetails")
	})
}

// Helper function to check if a key exists in a map
func assertKeyExists(t *testing.T, data map[string]interface{}, key string) {
	if _, ok := data[key]; !ok {
		t.Errorf("Key %q not found in map. Available keys: %v", key, reflect.ValueOf(data).MapKeys())
	}
}
