package resource

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Example model for testing
type TestUser struct {
	ID        uint      `refine:"label=User ID;width=80;fixed=left"`
	Name      string    `refine:"label=Full Name;required;min=3;max=50;searchable"`
	Email     string    `refine:"label=Email Address;required;pattern=^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$;searchable"`
	Role      string    `refine:"label=User Role;required"`
	Active    bool      `refine:"label=Is Active"`
	CreatedAt time.Time `refine:"label=Created At;width=150"`
}

func TestNewResource(t *testing.T) {
	// Create resource configuration
	config := ResourceConfig{
		Name:  "users",
		Label: "Users",
		Icon:  "user",
		Model: &TestUser{},
		Operations: []Operation{
			OperationList,
			OperationCreate,
			OperationUpdate,
			OperationDelete,
		},
		// Define specific field lists
		FilterableFields: []string{"name", "email", "role", "active"},
		SearchableFields: []string{"name", "email"},
		SortableFields:   []string{"id", "name", "email", "created_at"},
		TableFields:      []string{"id", "name", "email", "role", "active", "created_at"},
		FormFields:       []string{"name", "email", "role", "active"},
		RequiredFields:   []string{"name", "email", "role"},
	}

	// Create resource
	res := NewResource(config)

	// Test basic properties
	assert.Equal(t, "users", res.GetName())
	assert.Equal(t, "Users", res.GetLabel())
	assert.Equal(t, "user", res.GetIcon())

	// Test field lists
	defaultRes := res.(*DefaultResource)
	assert.ElementsMatch(t, []string{"name", "email", "role", "active"}, defaultRes.FilterableFields)
	assert.ElementsMatch(t, []string{"name", "email"}, defaultRes.SearchableFields)
	assert.ElementsMatch(t, []string{"id", "name", "email", "created_at"}, defaultRes.SortableFields)
	assert.ElementsMatch(t, []string{"id", "name", "email", "role", "active", "created_at"}, defaultRes.TableFields)
	assert.ElementsMatch(t, []string{"name", "email", "role", "active"}, defaultRes.FormFields)
	assert.ElementsMatch(t, []string{"name", "email", "role"}, defaultRes.RequiredFields)
}

func TestNewResourceWithDefaults(t *testing.T) {
	// Create minimal resource configuration
	config := ResourceConfig{
		Name:  "users",
		Model: &TestUser{},
	}

	// Create resource
	res := NewResource(config)
	defaultRes := res.(*DefaultResource)

	// Test that default field lists are created correctly
	fields := defaultRes.GetFields()
	fieldNames := make([]string, len(fields))
	for i, f := range fields {
		fieldNames[i] = f.Name
	}

	// By default, all fields should be filterable and sortable
	assert.ElementsMatch(t, fieldNames, defaultRes.FilterableFields)
	assert.ElementsMatch(t, fieldNames, defaultRes.SortableFields)

	// By default, all fields should be in table view
	assert.ElementsMatch(t, fieldNames, defaultRes.TableFields)

	// Form fields should exclude ID by default
	formFields := make([]string, 0)
	for _, name := range fieldNames {
		if name != "ID" {
			formFields = append(formFields, name)
		}
	}
	assert.ElementsMatch(t, formFields, defaultRes.FormFields)

	// Required fields should be detected from validation
	assert.ElementsMatch(t, []string{"Name", "Email", "Role"}, defaultRes.RequiredFields)
}

func TestFieldConfiguration(t *testing.T) {
	// Create resource
	config := ResourceConfig{
		Name:  "users",
		Model: &TestUser{},
	}
	res := NewResource(config)

	// Test field configurations
	field := res.GetField("Name")
	assert.NotNil(t, field)
	assert.Equal(t, "Full Name", field.Label)
	assert.NotNil(t, field.Validation)
	assert.True(t, field.Validation.Required)
	assert.Equal(t, 3, field.Validation.MinLength)
	assert.Equal(t, 50, field.Validation.MaxLength)

	// Test email field with pattern validation
	emailField := res.GetField("Email")
	assert.NotNil(t, emailField)
	assert.Equal(t, "Email Address", emailField.Label)
	assert.NotNil(t, emailField.Validation)
	assert.True(t, emailField.Validation.Required)
	assert.Equal(t, "^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$", emailField.Validation.Pattern)

	// Test field with list configuration
	idField := res.GetField("ID")
	assert.NotNil(t, idField)
	assert.NotNil(t, idField.List)
	assert.Equal(t, 80, idField.List.Width)
	assert.Equal(t, "left", idField.List.Fixed)
}

func TestSearchableFields(t *testing.T) {
	// Create resource with specific searchable fields
	config := ResourceConfig{
		Name:             "users",
		Model:            &TestUser{},
		SearchableFields: []string{"name", "email"},
	}
	res := NewResource(config)

	// Test searchable fields
	searchable := res.GetSearchable()
	assert.ElementsMatch(t, []string{"name", "email"}, searchable)
}

func TestIDFieldName(t *testing.T) {
	// Test default ID field name
	res := NewResource(ResourceConfig{
		Name:  "tests",
		Model: TestUser{},
	})
	assert.Equal(t, "ID", res.GetIDFieldName())

	// Test custom ID field name
	res = NewResource(ResourceConfig{
		Name:        "tests",
		Model:       TestUser{},
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
		Model:     TestUser{},
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
		Model:       TestUser{},
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
		Model:      TestUser{},
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
		Model:      TestUser{},
		Operations: []Operation{},
	}
	assert.False(t, emptyResource.HasOperation(OperationList))

	// Test with nil operations
	nilResource := &DefaultResource{
		Name:  "nil",
		Model: TestUser{},
	}
	assert.False(t, nilResource.HasOperation(OperationList))
}

func TestGetEditableFields(t *testing.T) {
	// Create test fields
	fields := []Field{
		{Name: "ID", Type: "int", ReadOnly: true},
		{Name: "Name", Type: "string", ReadOnly: false},
		{Name: "Email", Type: "string", ReadOnly: false},
		{Name: "CreatedAt", Type: "time.Time", ReadOnly: true},
		{Name: "Active", Type: "bool", ReadOnly: false},
		{Name: "Role", Type: "string", ReadOnly: false, Hidden: true},
	}

	// Create a resource with explicit configuration
	resource := NewResource(ResourceConfig{
		Name:   "test",
		Fields: fields,
		Model:  TestUser{}, // Provide a valid model
	})

	editable := resource.GetEditableFields()

	// All non-readonly fields should be included (4), including hidden ones
	assert.Equal(t, 4, len(editable))
	assert.Contains(t, editable, "Name")
	assert.Contains(t, editable, "Email")
	assert.Contains(t, editable, "Active")
	assert.Contains(t, editable, "Role")         // Hidden but not readonly
	assert.NotContains(t, editable, "ID")        // ID is always excluded
	assert.NotContains(t, editable, "CreatedAt") // ReadOnly is excluded

	// Test with explicitly provided editable fields
	explicitResource := NewResource(ResourceConfig{
		Name:           "test",
		Fields:         fields,
		EditableFields: []string{"Name", "Active"},
		Model:          TestUser{}, // Provide a valid model
	})

	explicitEditable := explicitResource.GetEditableFields()

	// Should respect the explicitly provided list
	assert.Equal(t, 2, len(explicitEditable))
	assert.Contains(t, explicitEditable, "Name")
	assert.Contains(t, explicitEditable, "Active")
	assert.NotContains(t, explicitEditable, "Email")
	assert.NotContains(t, explicitEditable, "Role")
}
