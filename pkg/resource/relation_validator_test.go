package resource

import (
	"testing"
)

// RelationTestModel is a test model with relations
type RelationTestModel struct {
	ID       uint
	Name     string
	Category *uint   `json:"category_id"`
	Tags     []uint  `json:"tag_ids"`
	Author   *string `json:"author_id"`
}

// TestRelationValidatorResource implements the Resource interface for testing relation validation
type TestRelationValidatorResource struct {
	DefaultResource
}

func setupTestResource() Resource {
	return &TestRelationValidatorResource{
		DefaultResource: DefaultResource{
			Name:  "test-resource",
			Model: RelationTestModel{},
			Relations: []Relation{
				{
					Name:     "category",
					Type:     RelationTypeManyToOne,
					Resource: "categories",
					Field:    "Category",
				},
				{
					Name:     "tags",
					Type:     RelationTypeOneToMany,
					Resource: "tags",
					Field:    "Tags",
				},
				{
					Name:     "author",
					Type:     RelationTypeManyToOne,
					Resource: "authors",
					Field:    "Author",
				},
			},
		},
	}
}

func setupTestRegistry() {
	// Wyczyść rejestr globalny
	GlobalResourceRegistry = NewResourceRegistry()

	// Zarejestruj powiązane zasoby
	GlobalResourceRegistry.Register(&DefaultResource{
		Name: "categories",
		Model: struct {
			ID   uint
			Name string
		}{},
	})

	GlobalResourceRegistry.Register(&DefaultResource{
		Name: "tags",
		Model: struct {
			ID   uint
			Name string
		}{},
	})
}

func TestValidateRelations(t *testing.T) {
	// Setup test resources and registry
	setupTestRegistry()

	// Test case 1: Valid relations (without DB check)
	model := RelationTestModel{
		ID:       1,
		Name:     "Test",
		Category: func() *uint { id := uint(1); return &id }(),
		Tags:     []uint{1, 2, 3},
		Author:   func() *string { id := "author1"; return &id }(),
	}

	// Rejestrowanie typu modelu do testów
	GlobalResourceRegistry.Register(&DefaultResource{
		Name:  "relation_test_model",
		Model: RelationTestModel{},
		Relations: []Relation{
			{
				Name:     "category",
				Type:     RelationTypeManyToOne,
				Resource: "categories",
				Field:    "Category",
				Required: false,
			},
			{
				Name:     "tags",
				Type:     RelationTypeOneToMany,
				Resource: "tags",
				Field:    "Tags",
				Required: false,
			},
		},
	})

	err := ValidateRelations(nil, model)
	if err != nil {
		t.Errorf("ValidateRelations() returned an error for valid relations: %v", err)
	}

	// Test case 2: Invalid relation format (non-pointer for Category)
	invalidModel := RelationTestModel{
		ID:       1,
		Name:     "Test",
		Category: nil,
		Tags:     []uint{999}, // Valid format but non-existent ID
		Author:   func() *string { id := "author1"; return &id }(),
	}

	err = ValidateRelations(nil, invalidModel)
	if err != nil {
		// This should not error since we're doing basic validation without checking existence
		t.Errorf("ValidateRelations() should not return an error for basic validation: %v", err)
	}

	// Test case 3: Model with nil values for optional relations
	nilModel := RelationTestModel{
		ID:       1,
		Name:     "Test",
		Category: nil,
		Tags:     nil,
		Author:   nil,
	}

	err = ValidateRelations(nil, nilModel)
	if err != nil {
		t.Errorf("ValidateRelations() returned an error for nil relations: %v", err)
	}
}

func TestRelationValidatorValidate(t *testing.T) {
	// Setup test
	relation := Relation{
		Name:     "category",
		Type:     RelationTypeManyToOne,
		Resource: "categories",
		Field:    "Category",
	}

	relatedResource := &DefaultResource{
		Name: "categories",
		Model: struct {
			ID   uint
			Name string
		}{},
	}

	// Create validator manually
	validator := RelationValidator{
		Relation:        relation,
		RelatedResource: relatedResource,
	}

	// Valid value
	categoryID := uint(1)

	// Test validation
	err := validator.Validate(&categoryID)
	if err != nil {
		t.Errorf("RelationValidator.Validate() returned an error for valid relation: %v", err)
	}

	// Test validation for nil value (should be valid as it's optional)
	err = validator.Validate(nil)
	if err != nil {
		t.Errorf("RelationValidator.Validate() returned an error for nil relation: %v", err)
	}

	// Test with required=true
	validatorRequired := RelationValidator{
		Relation:        relation,
		RelatedResource: relatedResource,
		Required:        true,
	}

	err = validatorRequired.Validate(nil)
	if err == nil {
		t.Errorf("RelationValidator.Validate() did not return an error for nil required relation")
	}

	// Test multiple relation
	multiRelation := Relation{
		Name:     "tags",
		Type:     RelationTypeOneToMany,
		Resource: "tags",
		Field:    "Tags",
	}

	multiValidator := RelationValidator{
		Relation:        multiRelation,
		RelatedResource: relatedResource,
	}

	// Valid array value
	tagIDs := []uint{1, 2, 3}

	err = multiValidator.Validate(tagIDs)
	if err != nil {
		t.Errorf("RelationValidator.Validate() returned an error for valid multi-relation: %v", err)
	}

	// Test with MinItems constraint
	minItemsValidator := RelationValidator{
		Relation:        multiRelation,
		RelatedResource: relatedResource,
		MinItems:        5,
	}

	err = minItemsValidator.Validate(tagIDs)
	if err == nil {
		t.Errorf("RelationValidator.Validate() did not return an error for multi-relation with too few items")
	}
}
