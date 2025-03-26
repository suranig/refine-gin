package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockResource for testing metadata generation
type MetadataMockResource struct {
	mock.Mock
}

func (m *MetadataMockResource) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MetadataMockResource) GetLabel() string {
	args := m.Called()
	return args.String(0)
}

func (m *MetadataMockResource) GetIcon() string {
	args := m.Called()
	return args.String(0)
}

func (m *MetadataMockResource) GetModel() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MetadataMockResource) GetFields() []Field {
	args := m.Called()
	return args.Get(0).([]Field)
}

func (m *MetadataMockResource) GetOperations() []Operation {
	args := m.Called()
	return args.Get(0).([]Operation)
}

func (m *MetadataMockResource) HasOperation(op Operation) bool {
	args := m.Called(op)
	return args.Bool(0)
}

func (m *MetadataMockResource) GetDefaultSort() *Sort {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*Sort)
}

func (m *MetadataMockResource) GetFilters() []Filter {
	args := m.Called()
	return args.Get(0).([]Filter)
}

func (m *MetadataMockResource) GetMiddlewares() []interface{} {
	args := m.Called()
	return args.Get(0).([]interface{})
}

func (m *MetadataMockResource) GetRelations() []Relation {
	args := m.Called()
	return args.Get(0).([]Relation)
}

func (m *MetadataMockResource) HasRelation(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *MetadataMockResource) GetRelation(name string) *Relation {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	relation := args.Get(0).(Relation)
	return &relation
}

func (m *MetadataMockResource) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MetadataMockResource) GetField(name string) *Field {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	field := args.Get(0).(Field)
	return &field
}

func (m *MetadataMockResource) GetSearchable() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestGenerateResourceMetadata(t *testing.T) {
	// Create a mock resource
	mockResource := new(MetadataMockResource)

	// Setup expectations
	mockResource.On("GetName").Return("test_resource")
	mockResource.On("GetLabel").Return("Test Resource")
	mockResource.On("GetIcon").Return("test-icon")
	mockResource.On("GetOperations").Return([]Operation{OperationCreate, OperationRead, OperationList})
	mockResource.On("GetIDFieldName").Return("id")
	mockResource.On("GetDefaultSort").Return(&Sort{Field: "created_at", Order: "desc"})
	mockResource.On("GetFilters").Return([]Filter{
		{Field: "status", Operator: "eq", Value: "active"},
	})
	mockResource.On("GetSearchable").Return([]string{"name", "description"})

	// Setup fields with different validators
	mockResource.On("GetFields").Return([]Field{
		{
			Name:       "id",
			Type:       "int",
			Required:   true,
			Filterable: true,
			Sortable:   true,
		},
		{
			Name:       "name",
			Type:       "string",
			Required:   true,
			Searchable: true,
			Validators: []Validator{
				StringValidator{
					MinLength: 3,
					MaxLength: 50,
				},
			},
		},
		{
			Name: "age",
			Type: "int",
			Validators: []Validator{
				NumberValidator{
					Min: 18,
					Max: 120,
				},
			},
		},
	})

	// Setup relations
	mockResource.On("GetRelations").Return([]Relation{
		{
			Name:             "posts",
			Type:             RelationTypeOneToMany,
			Resource:         "post",
			Field:            "user_id",
			ReferenceField:   "id",
			IncludeByDefault: true,
		},
		{
			Name:           "profile",
			Type:           RelationTypeOneToOne,
			Resource:       "profile",
			Field:          "profile_id",
			ReferenceField: "id",
			DisplayField:   "bio",
			ValueField:     "id",
			Required:       true,
		},
	})

	// Call the function to test
	metadata := GenerateResourceMetadata(mockResource)

	// Check resource metadata
	assert.Equal(t, "test_resource", metadata.Name)
	assert.Equal(t, "Test Resource", metadata.Label)
	assert.Equal(t, "test-icon", metadata.Icon)
	assert.Equal(t, []Operation{OperationCreate, OperationRead, OperationList}, metadata.Operations)
	assert.Equal(t, "id", metadata.IDFieldName)
	assert.Equal(t, &Sort{Field: "created_at", Order: "desc"}, metadata.DefaultSort)
	assert.Equal(t, []Filter{{Field: "status", Operator: "eq", Value: "active"}}, metadata.Filters)
	assert.Equal(t, []string{"name", "description"}, metadata.Searchable)

	// Check fields metadata
	assert.Equal(t, 3, len(metadata.Fields))

	// Check first field
	assert.Equal(t, "id", metadata.Fields[0].Name)
	assert.Equal(t, "int", metadata.Fields[0].Type)
	assert.True(t, metadata.Fields[0].Required)
	assert.True(t, metadata.Fields[0].Filterable)
	assert.True(t, metadata.Fields[0].Sortable)

	// Check second field with string validator
	assert.Equal(t, "name", metadata.Fields[1].Name)
	assert.Equal(t, "string", metadata.Fields[1].Type)
	assert.True(t, metadata.Fields[1].Required)
	assert.True(t, metadata.Fields[1].Searchable)
	assert.Equal(t, 1, len(metadata.Fields[1].Validators))
	assert.Equal(t, "string", metadata.Fields[1].Validators[0].Type)

	// Check validator rules
	if minLength, ok := metadata.Fields[1].Validators[0].Rules["minLength"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var minLengthInt int
		switch v := minLength.(type) {
		case float64:
			minLengthInt = int(v)
		case int:
			minLengthInt = v
		}
		assert.Equal(t, 3, minLengthInt)
	} else {
		t.Error("minLength rule should exist")
	}

	if maxLength, ok := metadata.Fields[1].Validators[0].Rules["maxLength"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var maxLengthInt int
		switch v := maxLength.(type) {
		case float64:
			maxLengthInt = int(v)
		case int:
			maxLengthInt = v
		}
		assert.Equal(t, 50, maxLengthInt)
	} else {
		t.Error("maxLength rule should exist")
	}

	// Check third field with number validator
	assert.Equal(t, "age", metadata.Fields[2].Name)
	assert.Equal(t, "int", metadata.Fields[2].Type)
	assert.Equal(t, 1, len(metadata.Fields[2].Validators))
	assert.Equal(t, "number", metadata.Fields[2].Validators[0].Type)

	// Check validator rules
	if min, ok := metadata.Fields[2].Validators[0].Rules["min"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var minInt int
		switch v := min.(type) {
		case float64:
			minInt = int(v)
		case int:
			minInt = v
		}
		assert.Equal(t, 18, minInt)
	} else {
		t.Error("min rule should exist")
	}

	if max, ok := metadata.Fields[2].Validators[0].Rules["max"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var maxInt int
		switch v := max.(type) {
		case float64:
			maxInt = int(v)
		case int:
			maxInt = v
		}
		assert.Equal(t, 120, maxInt)
	} else {
		t.Error("max rule should exist")
	}

	// Check relations metadata
	assert.Equal(t, 2, len(metadata.Relations))

	// Check first relation (one-to-many)
	assert.Equal(t, "posts", metadata.Relations[0].Name)
	assert.Equal(t, RelationTypeOneToMany, metadata.Relations[0].Type)
	assert.Equal(t, "post", metadata.Relations[0].Resource)
	assert.Equal(t, "user_id", metadata.Relations[0].Field)
	assert.Equal(t, "id", metadata.Relations[0].ReferenceField)
	assert.True(t, metadata.Relations[0].IncludeByDefault)

	// Check second relation (one-to-one)
	assert.Equal(t, "profile", metadata.Relations[1].Name)
	assert.Equal(t, RelationTypeOneToOne, metadata.Relations[1].Type)
	assert.Equal(t, "profile", metadata.Relations[1].Resource)
	assert.Equal(t, "profile_id", metadata.Relations[1].Field)
	assert.Equal(t, "id", metadata.Relations[1].ReferenceField)
	assert.Equal(t, "bio", metadata.Relations[1].DisplayField)
	assert.Equal(t, "id", metadata.Relations[1].ValueField)
	assert.True(t, metadata.Relations[1].Required)

	// Verify all expectations were met
	mockResource.AssertExpectations(t)
}

func TestGenerateFieldsMetadata(t *testing.T) {
	// Test with empty fields
	emptyFields := []Field{}
	emptyMetadata := GenerateFieldsMetadata(emptyFields)
	assert.Empty(t, emptyMetadata)

	// Test with various fields
	fields := []Field{
		{
			Name:       "id",
			Type:       "int",
			Required:   true,
			Filterable: true,
			Sortable:   true,
		},
		{
			Name:       "name",
			Type:       "string",
			Searchable: true,
			Unique:     true,
		},
		{
			Name: "description",
			Type: "text",
		},
	}

	metadata := GenerateFieldsMetadata(fields)
	assert.Equal(t, 3, len(metadata))

	// Check first field
	assert.Equal(t, "id", metadata[0].Name)
	assert.Equal(t, "int", metadata[0].Type)
	assert.True(t, metadata[0].Required)
	assert.True(t, metadata[0].Filterable)
	assert.True(t, metadata[0].Sortable)

	// Check second field
	assert.Equal(t, "name", metadata[1].Name)
	assert.Equal(t, "string", metadata[1].Type)
	assert.True(t, metadata[1].Searchable)
	assert.True(t, metadata[1].Unique)

	// Check third field
	assert.Equal(t, "description", metadata[2].Name)
	assert.Equal(t, "text", metadata[2].Type)
}

func TestGenerateValidatorsMetadata(t *testing.T) {
	// Test with empty validators
	emptyValidators := []Validator{}
	emptyMetadata := GenerateValidatorsMetadata(emptyValidators)
	assert.Empty(t, emptyMetadata)

	// Test with string validator
	stringValidator := StringValidator{
		MinLength: 5,
		MaxLength: 100,
		Pattern:   "^[a-zA-Z0-9]+$",
	}

	// Test with number validator
	numberValidator := NumberValidator{
		Min: 10,
		Max: 1000,
	}

	validators := []Validator{stringValidator, numberValidator}
	metadata := GenerateValidatorsMetadata(validators)

	assert.Equal(t, 2, len(metadata))

	// Check string validator metadata
	assert.Equal(t, "string", metadata[0].Type)

	// Check validator rules
	if minLength, ok := metadata[0].Rules["minLength"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var minLengthInt int
		switch v := minLength.(type) {
		case float64:
			minLengthInt = int(v)
		case int:
			minLengthInt = v
		}
		assert.Equal(t, 5, minLengthInt)
	} else {
		t.Error("minLength rule should exist")
	}

	if maxLength, ok := metadata[0].Rules["maxLength"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var maxLengthInt int
		switch v := maxLength.(type) {
		case float64:
			maxLengthInt = int(v)
		case int:
			maxLengthInt = v
		}
		assert.Equal(t, 100, maxLengthInt)
	} else {
		t.Error("maxLength rule should exist")
	}

	assert.Equal(t, "^[a-zA-Z0-9]+$", metadata[0].Rules["pattern"])

	// Check number validator metadata
	assert.Equal(t, "number", metadata[1].Type)

	// Check validator rules
	if min, ok := metadata[1].Rules["min"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var minInt int
		switch v := min.(type) {
		case float64:
			minInt = int(v)
		case int:
			minInt = v
		}
		assert.Equal(t, 10, minInt)
	} else {
		t.Error("min rule should exist")
	}

	if max, ok := metadata[1].Rules["max"]; ok {
		// Convert to int if it's a float64, or leave as is if it's already an int
		var maxInt int
		switch v := max.(type) {
		case float64:
			maxInt = int(v)
		case int:
			maxInt = v
		}
		assert.Equal(t, 1000, maxInt)
	} else {
		t.Error("max rule should exist")
	}
}

func TestGenerateRelationsMetadata(t *testing.T) {
	// Test with empty relations
	emptyRelations := []Relation{}
	emptyMetadata := GenerateRelationsMetadata(emptyRelations)
	assert.Empty(t, emptyMetadata)

	// Test with various relations
	relations := []Relation{
		{
			Name:             "comments",
			Type:             RelationTypeOneToMany,
			Resource:         "comment",
			Field:            "post_id",
			ReferenceField:   "id",
			IncludeByDefault: true,
			MinItems:         1,
			MaxItems:         100,
		},
		{
			Name:           "author",
			Type:           RelationTypeManyToOne,
			Resource:       "user",
			Field:          "author_id",
			ReferenceField: "id",
			DisplayField:   "name",
			ValueField:     "id",
			Required:       true,
			Cascade:        true,
			OnDelete:       "CASCADE",
			OnUpdate:       "CASCADE",
		},
		{
			Name:         "tags",
			Type:         RelationTypeManyToMany,
			Resource:     "tag",
			PivotTable:   "post_tags",
			PivotFields:  map[string]string{"post_id": "id", "tag_id": "id"},
			DisplayField: "name",
			ValueField:   "id",
		},
	}

	metadata := GenerateRelationsMetadata(relations)
	assert.Equal(t, 3, len(metadata))

	// Check one-to-many relation
	assert.Equal(t, "comments", metadata[0].Name)
	assert.Equal(t, RelationTypeOneToMany, metadata[0].Type)
	assert.Equal(t, "comment", metadata[0].Resource)
	assert.Equal(t, "post_id", metadata[0].Field)
	assert.Equal(t, "id", metadata[0].ReferenceField)
	assert.True(t, metadata[0].IncludeByDefault)
	assert.Equal(t, 1, metadata[0].MinItems)
	assert.Equal(t, 100, metadata[0].MaxItems)

	// Check many-to-one relation
	assert.Equal(t, "author", metadata[1].Name)
	assert.Equal(t, RelationTypeManyToOne, metadata[1].Type)
	assert.Equal(t, "user", metadata[1].Resource)
	assert.Equal(t, "author_id", metadata[1].Field)
	assert.Equal(t, "id", metadata[1].ReferenceField)
	assert.Equal(t, "name", metadata[1].DisplayField)
	assert.Equal(t, "id", metadata[1].ValueField)
	assert.True(t, metadata[1].Required)
	assert.True(t, metadata[1].Cascade)
	assert.Equal(t, "CASCADE", metadata[1].OnDelete)
	assert.Equal(t, "CASCADE", metadata[1].OnUpdate)

	// Check many-to-many relation
	assert.Equal(t, "tags", metadata[2].Name)
	assert.Equal(t, RelationTypeManyToMany, metadata[2].Type)
	assert.Equal(t, "tag", metadata[2].Resource)
	assert.Equal(t, "post_tags", metadata[2].PivotTable)
	assert.Equal(t, map[string]string{"post_id": "id", "tag_id": "id"}, metadata[2].PivotFields)
	assert.Equal(t, "name", metadata[2].DisplayField)
	assert.Equal(t, "id", metadata[2].ValueField)
}
