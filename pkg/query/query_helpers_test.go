package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseId(t *testing.T) {
	t.Run("numeric", func(t *testing.T) {
		id, err := ParseId("123")
		assert.NoError(t, err)
		assert.Equal(t, uint(123), id)
	})

	t.Run("non-numeric", func(t *testing.T) {
		id, err := ParseId("abc")
		assert.Error(t, err)
		assert.Equal(t, uint(0), id)
	})
}

func TestParseQueryParam(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		fieldType   string
		expected    interface{}
		expectError bool
	}{
		{name: "int valid", value: "42", fieldType: "int", expected: 42, expectError: false},
		{name: "int invalid", value: "abc", fieldType: "int", expectError: true},
		{name: "float valid", value: "3.14", fieldType: "float", expected: 3.14, expectError: false},
		{name: "float invalid", value: "abc", fieldType: "float", expectError: true},
		{name: "bool valid", value: "true", fieldType: "bool", expected: true, expectError: false},
		{name: "bool invalid", value: "notbool", fieldType: "bool", expectError: true},
		{name: "default", value: "hello", fieldType: "string", expected: "hello", expectError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseQueryParam(tt.value, tt.fieldType)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
