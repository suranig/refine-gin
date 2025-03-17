package resource

import (
	"reflect"
	"testing"
	"time"

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
	// Define a test model with various field types and tags
	type TestModel struct {
		ID        string    `json:"id" refine:"!filterable;!sortable"`
		Name      string    `json:"name" refine:"sortable;!filterable"`
		Email     string    `json:"email" refine:"filterable;!sortable"`
		Age       int       `json:"age" refine:"sortable;filterable"`
		IsActive  bool      `json:"is_active" refine:"!sortable;!filterable"`
		CreatedAt time.Time `json:"created_at" refine:"sortable;!filterable"`
		UpdatedAt time.Time `json:"-" refine:"-"` // Should be ignored
		Password  string    `json:"-"`            // Should be ignored due to json tag
		Notes     string    // No tags, should use default behavior
	}

	// Generate fields from the model
	fields := GenerateFieldsFromModel(TestModel{})

	// Verify the number of fields (all fields are included regardless of json tags)
	assert.Len(t, fields, 9)

	// Find specific fields by name
	var idField, nameField, emailField, ageField, isActiveField, createdAtField Field
	var notesField Field

	for _, field := range fields {
		switch field.Name {
		case "ID":
			idField = field
		case "Name":
			nameField = field
		case "Email":
			emailField = field
		case "Age":
			ageField = field
		case "IsActive":
			isActiveField = field
		case "CreatedAt":
			createdAtField = field
		case "Notes":
			notesField = field
		}
	}

	// Check specific fields
	assert.Equal(t, "ID", idField.Name)
	assert.False(t, idField.Filterable)
	assert.False(t, idField.Sortable)

	assert.Equal(t, "Name", nameField.Name)
	assert.True(t, nameField.Sortable)
	assert.False(t, nameField.Filterable)

	assert.Equal(t, "Email", emailField.Name)
	assert.False(t, emailField.Sortable)
	assert.True(t, emailField.Filterable)

	assert.Equal(t, "Age", ageField.Name)
	assert.True(t, ageField.Sortable)
	assert.True(t, ageField.Filterable)

	assert.Equal(t, "IsActive", isActiveField.Name)
	assert.False(t, isActiveField.Sortable)
	assert.False(t, isActiveField.Filterable)

	assert.Equal(t, "CreatedAt", createdAtField.Name)
	assert.True(t, createdAtField.Sortable)
	assert.False(t, createdAtField.Filterable)

	assert.Equal(t, "Notes", notesField.Name)
	assert.True(t, notesField.Sortable)
	assert.True(t, notesField.Filterable)

	// Test with a pointer to the model
	fieldsFromPtr := GenerateFieldsFromModel(&TestModel{})
	assert.Len(t, fieldsFromPtr, 9)
}

func TestParseFieldTag(t *testing.T) {
	// Test all possible tag options

	// Test basic flags
	field := Field{
		Name:       "test",
		Type:       "string",
		Filterable: false,
		Sortable:   false,
		Searchable: false,
		Required:   false,
		Unique:     false,
	}

	ParseFieldTag(&field, "filterable;sortable;searchable;required;unique")

	assert.True(t, field.Filterable)
	assert.True(t, field.Sortable)
	assert.True(t, field.Searchable)
	assert.True(t, field.Required)
	assert.True(t, field.Unique)

	// Test negative flags
	field = Field{
		Name:       "test",
		Type:       "string",
		Filterable: true,
		Sortable:   true,
		Searchable: true,
		Required:   true,
		Unique:     true,
	}

	ParseFieldTag(&field, "!filterable;!sortable;!searchable;!required;!unique")

	assert.False(t, field.Filterable)
	assert.False(t, field.Sortable)
	assert.False(t, field.Searchable)
	assert.False(t, field.Required)
	assert.False(t, field.Unique)

	// Test string validators
	field = Field{
		Name: "test",
		Type: "string",
	}

	ParseFieldTag(&field, "min=5;max=10;pattern=[a-z]+")

	assert.Len(t, field.Validators, 3)

	// Check string validator with min length
	stringValidator1, ok := field.Validators[0].(StringValidator)
	assert.True(t, ok)
	assert.Equal(t, 5, stringValidator1.MinLength)

	// Check string validator with max length
	stringValidator2, ok := field.Validators[1].(StringValidator)
	assert.True(t, ok)
	assert.Equal(t, 10, stringValidator2.MaxLength)

	// Check string validator with pattern
	stringValidator3, ok := field.Validators[2].(StringValidator)
	assert.True(t, ok)
	assert.Equal(t, "[a-z]+", stringValidator3.Pattern)

	// Test number validators for int type
	field = Field{
		Name: "test",
		Type: "int",
	}

	ParseFieldTag(&field, "min=5;max=10")

	assert.Len(t, field.Validators, 2)

	// Check number validator with min value
	numberValidator1, ok := field.Validators[0].(NumberValidator)
	assert.True(t, ok)
	assert.Equal(t, float64(5), numberValidator1.Min)

	// Check number validator with max value
	numberValidator2, ok := field.Validators[1].(NumberValidator)
	assert.True(t, ok)
	assert.Equal(t, float64(10), numberValidator2.Max)

	// Test number validators for float type
	field = Field{
		Name: "test",
		Type: "float64",
	}

	ParseFieldTag(&field, "min=5;max=10")

	assert.Len(t, field.Validators, 2)

	// Check number validator with min value
	numberValidator1, ok = field.Validators[0].(NumberValidator)
	assert.True(t, ok)
	assert.Equal(t, float64(5), numberValidator1.Min)

	// Check number validator with max value
	numberValidator2, ok = field.Validators[1].(NumberValidator)
	assert.True(t, ok)
	assert.Equal(t, float64(10), numberValidator2.Max)

	// Test with invalid min/max values
	field = Field{
		Name: "test",
		Type: "string",
	}

	ParseFieldTag(&field, "min=invalid;max=invalid")

	assert.Len(t, field.Validators, 0)

	// Test with empty tag
	field = Field{
		Name: "test",
		Type: "string",
	}

	ParseFieldTag(&field, "")

	assert.Len(t, field.Validators, 0)

	// Test with unknown tag parts
	field = Field{
		Name: "test",
		Type: "string",
	}

	ParseFieldTag(&field, "unknown=value;another_unknown")

	assert.Len(t, field.Validators, 0)
}

func TestCreateSliceOfType(t *testing.T) {
	// Test with struct
	type TestStruct struct {
		ID   string
		Name string
	}

	slice := CreateSliceOfType(TestStruct{})

	// Check type
	sliceType := reflect.TypeOf(slice)
	assert.Equal(t, reflect.Ptr, sliceType.Kind())
	assert.Equal(t, reflect.Slice, sliceType.Elem().Kind())
	assert.Equal(t, "TestStruct", sliceType.Elem().Elem().Name())

	// Check it's empty
	sliceValue := reflect.ValueOf(slice).Elem()
	assert.Equal(t, 0, sliceValue.Len())

	// Test with pointer to struct
	ptrSlice := CreateSliceOfType(&TestStruct{})

	// Check type
	ptrSliceType := reflect.TypeOf(ptrSlice)
	assert.Equal(t, reflect.Ptr, ptrSliceType.Kind())
	assert.Equal(t, reflect.Slice, ptrSliceType.Elem().Kind())
	assert.Equal(t, "TestStruct", ptrSliceType.Elem().Elem().Name())
}

func TestCreateInstanceOfType(t *testing.T) {
	// Test with struct
	type TestStruct struct {
		ID   string
		Name string
	}

	instance := CreateInstanceOfType(TestStruct{})

	// Check type
	instanceType := reflect.TypeOf(instance)
	assert.Equal(t, reflect.Ptr, instanceType.Kind())
	assert.Equal(t, "TestStruct", instanceType.Elem().Name())

	// Test with pointer to struct
	ptrInstance := CreateInstanceOfType(&TestStruct{})

	// Check type
	ptrInstanceType := reflect.TypeOf(ptrInstance)
	assert.Equal(t, reflect.Ptr, ptrInstanceType.Kind())
	assert.Equal(t, "TestStruct", ptrInstanceType.Elem().Name())
}

func TestSetID(t *testing.T) {
	// Test with string ID
	type StringIDStruct struct {
		ID   string
		Name string
	}

	strObj := &StringIDStruct{Name: "Test"}
	err := SetID(strObj, "123")
	assert.NoError(t, err)
	assert.Equal(t, "123", strObj.ID)

	// Test with int ID
	type IntIDStruct struct {
		ID   int
		Name string
	}

	intObj := &IntIDStruct{Name: "Test"}
	err = SetID(intObj, 456)
	assert.NoError(t, err)
	assert.Equal(t, 456, intObj.ID)

	// Test with string to int conversion
	intObj2 := &IntIDStruct{Name: "Test"}
	err = SetID(intObj2, "789")
	assert.NoError(t, err)
	assert.Equal(t, 789, intObj2.ID)

	// Test with uint ID
	type UintIDStruct struct {
		ID   uint
		Name string
	}

	uintObj := &UintIDStruct{Name: "Test"}
	err = SetID(uintObj, uint(789))
	assert.NoError(t, err)
	assert.Equal(t, uint(789), uintObj.ID)

	// Test with string to uint conversion
	uintObj2 := &UintIDStruct{Name: "Test"}
	err = SetID(uintObj2, "101")
	assert.NoError(t, err)
	assert.Equal(t, uint(101), uintObj2.ID)

	// Test with invalid conversion to int
	intObj3 := &IntIDStruct{Name: "Test"}
	err = SetID(intObj3, "abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")

	// Test with invalid conversion to uint
	uintObj3 := &UintIDStruct{Name: "Test"}
	err = SetID(uintObj3, "abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")

	// Test with unsupported ID type
	type FloatIDStruct struct {
		ID   float64
		Name string
	}

	floatObj := &FloatIDStruct{Name: "Test"}
	err = SetID(floatObj, "123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot convert ID to type")

	// Test with non-pointer - should return error
	nonPtrObj := StringIDStruct{Name: "Test"}
	err = SetID(nonPtrObj, "123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a pointer")

	// Test with struct without ID field
	type NoIDStruct struct {
		Name string
	}

	noIDObj := &NoIDStruct{Name: "Test"}
	err = SetID(noIDObj, "123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID field does not exist")

	// Test with same type ID (no conversion needed)
	strObj2 := &StringIDStruct{Name: "Test"}
	err = SetID(strObj2, "direct")
	assert.NoError(t, err)
	assert.Equal(t, "direct", strObj2.ID)
}

func TestDefaultResourceRelations(t *testing.T) {
	// Create sample relations
	relations := []Relation{
		{
			Name:             "Posts",
			Type:             RelationTypeOneToMany,
			Resource:         "posts",
			Field:            "author_id",
			ReferenceField:   "id",
			IncludeByDefault: false,
		},
		{
			Name:             "Profile",
			Type:             RelationTypeOneToOne,
			Resource:         "profiles",
			Field:            "user_id",
			ReferenceField:   "id",
			IncludeByDefault: true,
		},
	}

	// Create a DefaultResource with the relations
	resource := &DefaultResource{
		Name:      "users",
		Model:     User{},
		Relations: relations,
	}

	// Test GetRelations
	returnedRelations := resource.GetRelations()
	assert.Equal(t, relations, returnedRelations)
	assert.Len(t, returnedRelations, 2)

	// Test HasRelation
	assert.True(t, resource.HasRelation("Posts"))
	assert.True(t, resource.HasRelation("Profile"))
	assert.False(t, resource.HasRelation("Comments"))
	assert.False(t, resource.HasRelation(""))

	// Test GetRelation
	postsRelation := resource.GetRelation("Posts")
	assert.NotNil(t, postsRelation)
	assert.Equal(t, "Posts", postsRelation.Name)
	assert.Equal(t, RelationTypeOneToMany, postsRelation.Type)

	profileRelation := resource.GetRelation("Profile")
	assert.NotNil(t, profileRelation)
	assert.Equal(t, "Profile", profileRelation.Name)
	assert.Equal(t, RelationTypeOneToOne, profileRelation.Type)

	// Test GetRelation with non-existent relation
	nonExistentRelation := resource.GetRelation("NonExistent")
	assert.Nil(t, nonExistentRelation)
}

func TestDefaultResourceGetFiltersAndMiddlewares(t *testing.T) {
	// Create filters
	filters := []Filter{
		{
			Field:    "name",
			Operator: "eq",
			Value:    "John",
		},
		{
			Field:    "age",
			Operator: "gt",
			Value:    18,
		},
	}

	// Create middlewares
	middleware1 := func() {}
	middleware2 := func() {}
	middlewares := []interface{}{middleware1, middleware2}

	// Create resource with filters and middlewares
	resource := &DefaultResource{
		Name:        "users",
		Model:       User{},
		Filters:     filters,
		Middlewares: middlewares,
	}

	// Test GetFilters
	returnedFilters := resource.GetFilters()
	assert.Equal(t, filters, returnedFilters)
	assert.Len(t, returnedFilters, 2)

	// Test GetMiddlewares
	returnedMiddlewares := resource.GetMiddlewares()
	assert.Equal(t, middlewares, returnedMiddlewares)
	assert.Len(t, returnedMiddlewares, 2)
}

func TestExtractFieldsFromModel(t *testing.T) {
	// Test with a model
	type TestModel struct {
		ID        string    `json:"id" refine:"filterable;sortable"`
		Name      string    `json:"name" refine:"filterable;searchable"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
	}

	fields := ExtractFieldsFromModel(TestModel{})

	// Should be the same as GenerateFieldsFromModel
	expectedFields := GenerateFieldsFromModel(TestModel{})
	assert.Equal(t, expectedFields, fields)

	// Test with a pointer to a model
	ptrFields := ExtractFieldsFromModel(&TestModel{})
	assert.Equal(t, expectedFields, ptrFields)
}

func TestDefaultResourceHasOperation(t *testing.T) {
	// Create a resource with operations
	operations := []Operation{
		OperationList,
		OperationCreate,
		OperationRead,
	}

	resource := &DefaultResource{
		Name:       "users",
		Model:      User{},
		Operations: operations,
	}

	// Test HasOperation with existing operations
	assert.True(t, resource.HasOperation(OperationList))
	assert.True(t, resource.HasOperation(OperationCreate))
	assert.True(t, resource.HasOperation(OperationRead))

	// Test HasOperation with non-existing operations
	assert.False(t, resource.HasOperation(OperationUpdate))
	assert.False(t, resource.HasOperation(OperationDelete))

	// Test with empty operations
	emptyResource := &DefaultResource{
		Name:       "empty",
		Model:      User{},
		Operations: []Operation{},
	}
	assert.False(t, emptyResource.HasOperation(OperationList))

	// Test with nil operations
	nilResource := &DefaultResource{
		Name:  "nil",
		Model: User{},
	}
	assert.False(t, nilResource.HasOperation(OperationList))
}

func TestGetIDFieldName(t *testing.T) {
	// Test z domyślną nazwą pola identyfikatora
	res := NewResource(ResourceConfig{
		Name:  "tests",
		Model: TestModel{},
	})
	assert.Equal(t, "ID", res.GetIDFieldName())

	// Test z niestandardową nazwą pola identyfikatora
	res = NewResource(ResourceConfig{
		Name:        "tests",
		Model:       TestModel{},
		IDFieldName: "UID",
	})
	assert.Equal(t, "UID", res.GetIDFieldName())
}

func TestSetCustomID(t *testing.T) {
	type CustomStringIDStruct struct {
		UID  string
		Name string
	}
	stringObj := &CustomStringIDStruct{Name: "Test"}
	err := SetCustomID(stringObj, "123", "UID")
	assert.NoError(t, err)
	assert.Equal(t, "123", stringObj.UID)

	type CustomIntIDStruct struct {
		UID  int
		Name string
	}
	intObj := &CustomIntIDStruct{Name: "Test"}
	err = SetCustomID(intObj, "123", "UID")
	assert.NoError(t, err)
	assert.Equal(t, 123, intObj.UID)

	type CustomUintIDStruct struct {
		UID  uint
		Name string
	}
	uintObj := &CustomUintIDStruct{Name: "Test"}
	err = SetCustomID(uintObj, "123", "UID")
	assert.NoError(t, err)
	assert.Equal(t, uint(123), uintObj.UID)

	type NoUIDStruct struct {
		Name string
	}
	noIDObj := &NoUIDStruct{Name: "Test"}
	err = SetCustomID(noIDObj, "123", "UID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UID field does not exist")

	type InvalidIDStruct struct {
		UID  []string
		Name string
	}
	invalidObj := &InvalidIDStruct{Name: "Test"}
	err = SetCustomID(invalidObj, "123", "UID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot convert UID to type")
}
