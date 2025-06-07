package repository

import (
	"reflect"
	"testing"
)

func TestConvertToFieldType(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		target   reflect.Type
		expected interface{}
	}{
		{"string-to-int", "42", reflect.TypeOf(int(0)), int(42)},
		{"string-to-uint", "7", reflect.TypeOf(uint(0)), uint(7)},
		{"string-to-float", "3.14", reflect.TypeOf(float64(0)), float64(3.14)},
		{"string-to-bool", "true", reflect.TypeOf(true), true},
		{"same-type", 5, reflect.TypeOf(5), 5},
		// conversions from non-string values
		{"int-to-string", 42, reflect.TypeOf(""), "42"},
		{"uint-to-int", uint(7), reflect.TypeOf(int(0)), int(7)},
		{"float-to-string", 3.14, reflect.TypeOf(""), "3.14"},
		{"bool-to-string", true, reflect.TypeOf(""), "true"},
		{"int-to-float", 8, reflect.TypeOf(float64(0)), float64(8)},
	}

	for _, tt := range tests {
		got, err := convertToFieldType(tt.value, tt.target)
		if err != nil {
			t.Errorf("%s: unexpected error %v", tt.name, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, got)
		}
	}

	// unsupported conversion
	if _, err := convertToFieldType("bad", reflect.TypeOf(struct{}{})); err == nil {
		t.Errorf("expected error for unsupported conversion")
	}
}
