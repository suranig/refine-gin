package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseId(t *testing.T) {
	tests := []struct {
		input       string
		expected    uint
		expectError bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"abc", 0, true},
		{"12a", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			id, err := ParseId(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, id)
			}
		})
	}
}

func TestParseQueryParam(t *testing.T) {
	tests := []struct {
		value       string
		typ         string
		expected    interface{}
		expectError bool
	}{
		{"42", "int", 42, false},
		{"notint", "int", 0, true},
		{"3.14", "float", 3.14, false},
		{"nofloat", "float", 0.0, true},
		{"true", "bool", true, false},
		{"maybe", "bool", false, true},
		{"some", "string", "some", false},
	}

	for _, tt := range tests {
		t.Run(tt.typ+"_"+tt.value, func(t *testing.T) {
			res, err := ParseQueryParam(tt.value, tt.typ)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}
