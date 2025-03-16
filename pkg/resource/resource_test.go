package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	ID        string `json:"id" refine:"filterable;sortable;searchable"`
	Name      string `json:"name" refine:"filterable;sortable"`
	Email     string `json:"email" refine:"filterable"`
	CreatedAt string `json:"created_at" refine:"filterable;sortable"`
}

func TestNewResource(t *testing.T) {
	// Create a new resource
	res := NewResource(ResourceConfig{
		Name:  "tests",
		Model: TestModel{},
		Operations: []Operation{
			OperationList,
			OperationCreate,
			OperationRead,
			OperationUpdate,
			OperationDelete,
		},
		DefaultSort: &Sort{
			Field: "created_at",
			Order: "desc",
		},
	})

	// Check resource properties
	assert.Equal(t, "tests", res.GetName())
	assert.Equal(t, TestModel{}, res.GetModel())
	assert.Len(t, res.GetFields(), 4) // ID, Name, Email, CreatedAt
	assert.Len(t, res.GetOperations(), 5)
	assert.True(t, res.HasOperation(OperationList))
	assert.True(t, res.HasOperation(OperationCreate))
	assert.True(t, res.HasOperation(OperationRead))
	assert.True(t, res.HasOperation(OperationUpdate))
	assert.True(t, res.HasOperation(OperationDelete))
	assert.NotNil(t, res.GetDefaultSort())
	assert.Equal(t, "created_at", res.GetDefaultSort().Field)
	assert.Equal(t, "desc", res.GetDefaultSort().Order)
}

func TestGenerateFieldsFromModel(t *testing.T) {
	// Generate fields from model
	fields := GenerateFieldsFromModel(TestModel{})

	// Check fields
	assert.Len(t, fields, 4)

	// Check ID field
	idField := fields[0]
	assert.Equal(t, "ID", idField.Name)
	assert.True(t, idField.Filterable)
	assert.True(t, idField.Sortable)
	assert.True(t, idField.Searchable)

	// Check Name field
	nameField := fields[1]
	assert.Equal(t, "Name", nameField.Name)
	assert.True(t, nameField.Filterable)
	assert.True(t, nameField.Sortable)
	assert.False(t, nameField.Searchable)

	// Check Email field
	emailField := fields[2]
	assert.Equal(t, "Email", emailField.Name)
	assert.True(t, emailField.Filterable)
	assert.True(t, emailField.Sortable)
	assert.False(t, emailField.Searchable)
}

func TestParseFieldTag(t *testing.T) {
	// Create a field
	field := Field{
		Name:       "test",
		Type:       "string",
		Filterable: false,
		Sortable:   false,
		Searchable: false,
		Required:   false,
		Unique:     false,
	}

	// Parse tag
	ParseFieldTag(&field, "filterable;sortable;searchable;required;unique;min=5;max=10;pattern=[a-z]+")

	// Check field properties
	assert.True(t, field.Filterable)
	assert.True(t, field.Sortable)
	assert.True(t, field.Searchable)
	assert.True(t, field.Required)
	assert.True(t, field.Unique)

	// Check validators
	assert.Len(t, field.Validators, 3)

	// Check string validator with min length
	stringValidator1 := field.Validators[0].(StringValidator)
	assert.Equal(t, 5, stringValidator1.MinLength)

	// Check string validator with max length
	stringValidator2 := field.Validators[1].(StringValidator)
	assert.Equal(t, 10, stringValidator2.MaxLength)

	// Check string validator with pattern
	stringValidator3 := field.Validators[2].(StringValidator)
	assert.Equal(t, "[a-z]+", stringValidator3.Pattern)
}
