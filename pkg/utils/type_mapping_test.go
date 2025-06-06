package utils

import "testing"

func TestGetTypeMapping(t *testing.T) {
	tests := []struct {
		goType   string
		expected TypeMapping
	}{
		{"string", TypeMapping{Category: TypeString, IsPrimitive: true}},
		{"int64", TypeMapping{Category: TypeInteger, Format: "int64", IsPrimitive: true}},
		{"[]int", TypeMapping{Category: TypeArray, Format: "int", IsPrimitive: false}},
		{"Custom", TypeMapping{Category: TypeObject, Format: "Custom", IsPrimitive: false}},
	}

	for _, tt := range tests {
		got := GetTypeMapping(tt.goType)
		if got.Category != tt.expected.Category || got.Format != tt.expected.Format || got.IsPrimitive != tt.expected.IsPrimitive {
			t.Fatalf("GetTypeMapping(%q) = %+v, want %+v", tt.goType, got, tt.expected)
		}
	}
}

func TestHelpers(t *testing.T) {
	if !IsNumericType("float64") {
		t.Fatalf("expected float64 numeric")
	}
	if IsNumericType("string") {
		t.Fatalf("string should not be numeric")
	}
	if !IsStringType("string") {
		t.Fatalf("expected string type")
	}
	if IsStringType("int") {
		t.Fatalf("int is not string type")
	}
	if !IsArrayType("[]string") {
		t.Fatalf("expected array type")
	}
	if IsArrayType("string") {
		t.Fatalf("string is not array")
	}
	if GetArrayElementType("[]int") != "int" {
		t.Fatalf("unexpected element type for []int")
	}
	if GetArrayElementType("int") != "" {
		t.Fatalf("expected empty element type for non array")
	}
}
