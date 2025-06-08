package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test models
type TestCategory struct {
	ID   uint `gorm:"primarykey"`
	Name string
}

type TestProduct struct {
	ID         uint
	Name       string
	CategoryID uint
	Category   TestCategory
	Tags       []TestTag
}

type TestTag struct {
	ID   uint
	Name string
}

func TestRelationValidator_Validate_Nil(t *testing.T) {
	// Test validator with nil value and not required
	validator := RelationValidator{
		Relation: Relation{
			Name: "TestRelation",
		},
		Required: false,
	}

	err := validator.Validate(nil)
	assert.NoError(t, err)

	// Test validator with nil value and required
	validator.Required = true
	err = validator.Validate(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")

	// Test with custom message
	validator.Message = "Custom error message"
	err = validator.Validate(nil)
	assert.Error(t, err)
	assert.Equal(t, "Custom error message", err.Error())
}

func TestRelationValidator_Validate_UnsupportedType(t *testing.T) {
	// Test validator with unsupported relation type
	validator := RelationValidator{
		Relation: Relation{
			Name: "TestRelation",
			Type: "UnsupportedType",
		},
	}

	err := validator.Validate("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported relation type")
}

func TestRelationsValidation(t *testing.T) {
	// Setup global registry
	oldRegistry := GlobalResourceRegistry
	GlobalResourceRegistry = NewResourceRegistry()
	defer func() { GlobalResourceRegistry = oldRegistry }()

	// Create and register a test resource
	productResource := NewResource(ResourceConfig{
		Name:  "products",
		Model: TestProduct{},
		Relations: []Relation{
			{
				Name:           "Category",
				Type:           RelationTypeManyToOne,
				Resource:       "categories",
				Field:          "Category",
				ReferenceField: "ID",
				Required:       true,
			},
		},
	})
	RegisterToRegistry(productResource)

	// Test with invalid model type
	err := ValidateRelations(nil, "not a model")
	assert.NoError(t, err) // Should return nil as no resource found

	// Test with nil DB (should skip validation)
	product := TestProduct{
		ID:         1,
		Name:       "Test Product",
		CategoryID: 2,
		Category:   TestCategory{ID: 2},
	}
	err = ValidateRelations(nil, &product)
	assert.NoError(t, err)
}
