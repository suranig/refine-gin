package handler

import "testing"

func TestConvertToMap(t *testing.T) {
	type sample struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// struct value
	res := convertToMap(sample{Name: "John", Age: 30})
	if res["name"] != "John" || res["age"] != 30 {
		t.Errorf("unexpected map: %v", res)
	}

	// pointer to struct
	p := &sample{Name: "Jane", Age: 40}
	res = convertToMap(p)
	if res["name"] != "Jane" || res["age"] != 40 {
		t.Errorf("unexpected map for pointer: %v", res)
	}

	// nil pointer
	var nilPtr *sample
	if convertToMap(nilPtr) != nil {
		t.Errorf("expected nil for nil pointer")
	}

	// non-struct should return empty map
	if m := convertToMap(123); len(m) != 0 {
		t.Errorf("expected empty map")
	}
}

func TestGetMapKeys(t *testing.T) {
	m := map[string]interface{}{"a": 1, "b": 2, "c": 3}
	keys := getMapKeys(m)
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys")
	}
	seen := make(map[string]bool)
	for _, k := range keys {
		seen[k] = true
	}
	for k := range m {
		if !seen[k] {
			t.Errorf("missing key %s", k)
		}
	}
}
