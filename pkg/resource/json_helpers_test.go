package resource

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetJsonPropertyType(t *testing.T) {
	type sample struct{ Field string }

	tests := []struct {
		name string
		typ  reflect.Type
		want string
	}{
		{"bool", reflect.TypeOf(true), "boolean"},
		{"int", reflect.TypeOf(42), "number"},
		{"float", reflect.TypeOf(3.14), "number"},
		{"string", reflect.TypeOf("foo"), "string"},
		{"struct", reflect.TypeOf(sample{}), "object"},
		{"map", reflect.TypeOf(map[string]int{}), "object"},
		{"slice", reflect.TypeOf([]string{}), "array"},
		{"time", reflect.TypeOf(time.Time{}), "string"},
		{"ptr to struct", reflect.TypeOf(&sample{}), "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getJsonPropertyType(tt.typ)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProcessJsonTag(t *testing.T) {
	t.Run("omitempty clears required", func(t *testing.T) {
		field := &Field{Name: "Email", Label: "Email", Validation: &Validation{Required: true}}
		ProcessJsonTag(field, "email,omitempty")
		assert.False(t, field.Validation.Required)
	})

	t.Run("json tag renames field", func(t *testing.T) {
		field := &Field{Name: "UserName", Label: "UserName"}
		ProcessJsonTag(field, "username")
		assert.Equal(t, "username", field.Name)
		assert.Equal(t, "UserName", field.Label)
	})

	t.Run("mixed tags with validation", func(t *testing.T) {
		field := &Field{Name: "EmailAddress", Label: "EmailAddress", Validation: &Validation{Required: true}}
		ProcessJsonTag(field, "email,omitempty")
		assert.Equal(t, "email", field.Name)
		assert.Equal(t, "EmailAddress", field.Label)
		assert.False(t, field.Validation.Required)
	})
}
