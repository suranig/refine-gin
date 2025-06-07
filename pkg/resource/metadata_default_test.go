package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGenerateDefaultAntDesignConfig ensures default Ant Design config is generated correctly
func TestGenerateDefaultAntDesignConfig(t *testing.T) {
	// String field with placeholder
	strField := Field{
		Name: "username",
		Type: "string",
		Form: &FormConfig{Placeholder: "Enter username"},
	}
	strCfg := generateDefaultAntDesignConfig(&strField)
	assert.Equal(t, "Input", strCfg.ComponentType)
	assert.Equal(t, "Enter username", strCfg.Props["placeholder"])

	// Numeric field with min/max and read-only
	numField := Field{
		Name:       "age",
		Type:       "number",
		ReadOnly:   true,
		Validation: &Validation{Min: 18, Max: 65},
	}
	numCfg := generateDefaultAntDesignConfig(&numField)
	assert.Equal(t, "InputNumber", numCfg.ComponentType)
	assert.Equal(t, float64(18), numCfg.Props["min"])
	assert.Equal(t, float64(65), numCfg.Props["max"])
	assert.Equal(t, true, numCfg.Props["disabled"])

	// Boolean field
	boolField := Field{Name: "active", Type: "boolean"}
	boolCfg := generateDefaultAntDesignConfig(&boolField)
	assert.Equal(t, "Switch", boolCfg.ComponentType)
	assert.Equal(t, "checked", boolCfg.FormItemProps["valuePropName"])

	// Multiselect field with options and placeholder
	multiField := Field{
		Name:    "tags",
		Type:    "multiselect",
		Form:    &FormConfig{Placeholder: "Select tags"},
		Options: []Option{{Value: "a", Label: "A"}, {Value: "b", Label: "B"}},
	}
	multiCfg := generateDefaultAntDesignConfig(&multiField)
	assert.Equal(t, "Select", multiCfg.ComponentType)
	assert.Equal(t, "multiple", multiCfg.Props["mode"])
	assert.Equal(t, "Select tags", multiCfg.Props["placeholder"])

	opts, ok := multiCfg.Props["options"].([]map[string]interface{})
	assert.True(t, ok)
	assert.Len(t, opts, 2)
	assert.Equal(t, "a", opts[0]["value"])
	assert.Equal(t, "A", opts[0]["label"])
	assert.Equal(t, "b", opts[1]["value"])
	assert.Equal(t, "B", opts[1]["label"])
}
