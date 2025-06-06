package utils

import (
	"reflect"
	"testing"
)

type sampleStruct struct {
	Name string
	Age  int
	Tags []string
}

func TestGetFieldValue(t *testing.T) {
	obj := sampleStruct{Name: "John", Age: 30}
	m := map[string]interface{}{"key": 42}

	tests := []struct {
		name     string
		input    interface{}
		field    string
		expected interface{}
		wantErr  bool
	}{
		{"struct value", obj, "Name", "John", false},
		{"struct pointer", &obj, "Age", 30, false},
		{"map value", m, "key", 42, false},
		{"map missing", m, "other", nil, true},
		{"nil object", nil, "Name", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := GetFieldValue(tt.input, tt.field)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(val, tt.expected) {
					t.Fatalf("expected %v, got %v", tt.expected, val)
				}
			}
		})
	}
}

func TestSetFieldValue(t *testing.T) {
	obj := &sampleStruct{}

	if err := SetFieldValue(obj, "Name", "Alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Name != "Alice" {
		t.Fatalf("expected Name to be Alice, got %s", obj.Name)
	}

	if err := SetFieldValue(obj, "Age", 41); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Age != 41 {
		t.Fatalf("expected Age to be 41, got %d", obj.Age)
	}
	if err := SetFieldValue(obj, "Age", "42"); err == nil {
		t.Fatalf("expected type conversion error")
	}

	if err := SetFieldValue(sampleStruct{}, "Name", "Bob"); err == nil {
		t.Fatalf("expected error setting field on non-pointer")
	}
	if err := SetFieldValue(obj, "Unknown", 1); err == nil {
		t.Fatalf("expected error for unknown field")
	}
}

func TestIsSliceAndGetSliceField(t *testing.T) {
	obj := &sampleStruct{Tags: []string{"a", "b"}}

	if !IsSlice(obj.Tags) {
		t.Fatalf("expected slice to be detected")
	}
	if !IsSlice(&obj.Tags) {
		t.Fatalf("expected pointer to slice to be detected")
	}
	if IsSlice(5) {
		t.Fatalf("int should not be slice")
	}

	field, err := GetSliceField(obj, "Tags")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if field.Len() != 2 {
		t.Fatalf("expected field length 2, got %d", field.Len())
	}

	if _, err := GetSliceField(obj, "Missing"); err == nil {
		t.Fatalf("expected error for missing field")
	}
	type notSlice struct{ Field int }
	ns := &notSlice{Field: 1}
	if _, err := GetSliceField(ns, "Field"); err == nil {
		t.Fatalf("expected error for non-slice field")
	}
}

func TestCreateNewModelInstance(t *testing.T) {
	inst1 := CreateNewModelInstance(sampleStruct{})
	if _, ok := inst1.(*sampleStruct); !ok {
		t.Fatalf("expected *sampleStruct instance")
	}
	inst2 := CreateNewModelInstance(&sampleStruct{})
	if _, ok := inst2.(*sampleStruct); !ok {
		t.Fatalf("expected *sampleStruct instance for pointer input")
	}
	if CreateNewModelInstance(nil) != nil {
		t.Fatalf("expected nil for nil input")
	}
}

func TestCreateNewSliceOfModel(t *testing.T) {
	sl1 := CreateNewSliceOfModel(sampleStruct{})
	if _, ok := sl1.(*[]*sampleStruct); !ok {
		t.Fatalf("expected *[]*sampleStruct slice")
	}
	sl2 := CreateNewSliceOfModel(&sampleStruct{})
	if _, ok := sl2.(*[]*sampleStruct); !ok {
		t.Fatalf("expected *[]*sampleStruct for pointer input")
	}
	if CreateNewSliceOfModel(nil) != nil {
		t.Fatalf("expected nil for nil input")
	}
}
