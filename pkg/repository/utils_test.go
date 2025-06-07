package repository

import (
	"testing"
)

type testModel struct{ id string }

func (m *testModel) SetID(id interface{}) { m.id = id.(string) }

func TestTrySetID(t *testing.T) {
	m := &testModel{}
	ok := TrySetID(m, "123")
	if !ok || m.id != "123" {
		t.Errorf("TrySetID failed to set id")
	}

	var plain interface{} = struct{}{}
	if TrySetID(plain, "abc") {
		t.Errorf("TrySetID should return false for unsupported types")
	}
}
