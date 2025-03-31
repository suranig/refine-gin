package swagger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/resource"
)

// MockResource implements a minimal resource for testing
type MockResource struct {
	name           string
	fields         []resource.Field
	ops            []resource.Operation
	requiredFields []string
}

func (r MockResource) GetName() string             { return r.name }
func (r MockResource) GetLabel() string            { return r.name }
func (r MockResource) GetIcon() string             { return "default" }
func (r MockResource) GetFields() []resource.Field { return r.fields }
func (r MockResource) GetDefaultSort() *resource.Sort {
	return &resource.Sort{Field: "id", Order: "asc"}
}
func (r MockResource) GetModel() interface{}                      { return nil }
func (r MockResource) GetOperations() []resource.Operation        { return r.ops }
func (r MockResource) GetFilters() []resource.Filter              { return nil }
func (r MockResource) GetMiddlewares() []interface{}              { return nil }
func (r MockResource) GetRelations() []resource.Relation          { return nil }
func (r MockResource) HasRelation(name string) bool               { return false }
func (r MockResource) GetRelation(name string) *resource.Relation { return nil }
func (r MockResource) GetIDFieldName() string                     { return "ID" }
func (r MockResource) GetField(name string) *resource.Field {
	for _, f := range r.fields {
		if f.Name == name {
			return &f
		}
	}
	return nil
}
func (r MockResource) GetSearchable() []string {
	return []string{"name", "email"}
}
func (r MockResource) GetFilterableFields() []string {
	return []string{"id", "name", "email"}
}
func (r MockResource) GetSortableFields() []string {
	return []string{"id", "name", "email"}
}
func (r MockResource) GetRequiredFields() []string {
	return r.requiredFields
}
func (r MockResource) GetTableFields() []string {
	return []string{"id", "name", "email"}
}
func (r MockResource) GetFormFields() []string {
	return []string{"name", "email"}
}
func (r MockResource) GetEditableFields() []string {
	return []string{"name", "email", "age"}
}
func (r MockResource) GetPermissions() map[string][]string {
	return nil
}
func (r MockResource) HasPermission(operation string, role string) bool {
	return true
}
func (r MockResource) HasOperation(op resource.Operation) bool {
	for _, o := range r.ops {
		if o == op {
			return true
		}
	}
	return false
}
func (r MockResource) GetOptions() resource.Options {
	return resource.Options{}
}
func (m MockResource) GetFormLayout() *resource.FormLayout {
	return nil
}

func TestDefaultSwaggerInfo(t *testing.T) {
	info := DefaultSwaggerInfo()

	assert.Equal(t, "Refine-Gin API", info.Title)
	assert.Equal(t, "API documentation for Refine-Gin", info.Description)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "/api", info.BasePath)
	assert.Equal(t, []string{"http", "https"}, info.Schemes)
}

func TestCustomEndpoints(t *testing.T) {
	// Reset endpoints to ensure a clean test environment
	ResetCustomEndpoints()

	// Test initial state
	endpoints := GetCustomEndpoints()
	assert.Empty(t, endpoints, "Initial endpoints should be empty")

	// Register a custom endpoint
	testEndpoint := CustomEndpoint{
		Method: "post",
		Path:   "/auth/test",
		Operation: Operation{
			Summary:     "Test Operation",
			Description: "Test Description",
			OperationID: "testOperation",
			Tags:        []string{"auth"},
			Responses: map[string]Response{
				"200": {
					Description: "Success",
				},
			},
		},
	}

	RegisterCustomEndpoint(testEndpoint)

	// Verify endpoint was registered
	endpoints = GetCustomEndpoints()
	assert.Len(t, endpoints, 1, "Should have one endpoint registered")
	assert.Equal(t, testEndpoint, endpoints[0])

	// Register another endpoint
	anotherEndpoint := CustomEndpoint{
		Method: "get",
		Path:   "/test/endpoint",
		Operation: Operation{
			Summary:     "Another Test",
			Description: "Another Description",
			OperationID: "anotherTest",
			Tags:        []string{"test"},
			Responses: map[string]Response{
				"200": {
					Description: "Success",
				},
			},
		},
	}

	RegisterCustomEndpoint(anotherEndpoint)

	// Verify both endpoints exist
	endpoints = GetCustomEndpoints()
	assert.Len(t, endpoints, 2, "Should have two endpoints registered")

	// Reset and verify it clears all endpoints
	ResetCustomEndpoints()
	endpoints = GetCustomEndpoints()
	assert.Empty(t, endpoints, "Endpoints should be empty after reset")
}

func TestGenerateOpenAPI(t *testing.T) {
	// Reset any previously registered custom endpoints
	ResetCustomEndpoints()

	// Create mock resources
	mockResource := MockResource{
		name: "users",
		fields: []resource.Field{
			{Name: "id", Type: "int"},
			{Name: "name", Type: "string"},
			{Name: "email", Type: "string"},
			{Name: "age", Type: "int"},
		},
		requiredFields: []string{"id", "name", "email"},
		ops: []resource.Operation{
			resource.OperationCreate,
			resource.OperationRead,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationList,
		},
	}

	// Define SwaggerInfo
	info := SwaggerInfo{
		Title:       "Test API",
		Description: "API for testing",
		Version:     "1.0.0",
		BasePath:    "/api/v1",
		Schemes:     []string{"https"},
	}

	// Generate OpenAPI spec
	openAPI := GenerateOpenAPI([]resource.Resource{mockResource}, info)

	// Verify basic structure
	assert.Equal(t, "3.0.0", openAPI.OpenAPI)
	assert.Equal(t, "Test API", openAPI.Info.Title)
	assert.Equal(t, "API for testing", openAPI.Info.Description)
	assert.Equal(t, "1.0.0", openAPI.Info.Version)
	assert.Equal(t, "/api/v1", openAPI.Servers[0].URL)

	// Verify resource schema
	schema, exists := openAPI.Components.Schemas["users"]
	assert.True(t, exists, "Users schema should exist")
	assert.Equal(t, "object", schema.Type)
	assert.Len(t, schema.Properties, 4, "Should have 4 properties")

	// Verify paths
	assert.NotNil(t, openAPI.Paths["/users"], "List endpoint should exist")
	assert.NotNil(t, openAPI.Paths["/users/{id}"], "Get endpoint should exist")

	// Verify OPTIONS endpoint
	assert.NotNil(t, openAPI.Paths["/users"]["options"], "OPTIONS endpoint should exist")
	optionsOperation := openAPI.Paths["/users"]["options"]
	assert.Equal(t, "Get users metadata", optionsOperation.Summary)
	assert.Equal(t, "optionsUsers", optionsOperation.OperationID)
	assert.Contains(t, optionsOperation.Responses, "200")
	assert.Contains(t, optionsOperation.Responses, "304")

	// Test with custom endpoint
	customOp := Operation{
		Summary:     "Custom Operation",
		Description: "Custom Description",
		OperationID: "customOperation",
		Tags:        []string{"custom"},
		Responses: map[string]Response{
			"200": {
				Description: "Success",
			},
		},
	}

	RegisterCustomEndpoint(CustomEndpoint{
		Method:    "post",
		Path:      "/custom/path",
		Operation: customOp,
	})

	// Generate again with custom endpoint
	openAPI = GenerateOpenAPI([]resource.Resource{mockResource}, info)

	// Verify custom endpoint
	assert.NotNil(t, openAPI.Paths["/custom/path"], "Custom path should exist")
	assert.Equal(t, customOp, openAPI.Paths["/custom/path"]["post"])
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "Users"},
		{"Users", "Users"},
		{"USER", "USER"},
		{"u", "U"},
		{"", ""},
	}

	for _, test := range tests {
		result := capitalize(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestFieldToSchema(t *testing.T) {
	// Test primitives
	stringField := resource.Field{Name: "name", Type: "string"}
	schema := fieldToSchema(stringField)
	assert.Equal(t, "string", schema.Type)

	// Test integer
	intField := resource.Field{Name: "id", Type: "int"}
	schema = fieldToSchema(intField)
	assert.Equal(t, "integer", schema.Type)
	assert.Equal(t, "int32", schema.Format)

	// Test int64
	int64Field := resource.Field{Name: "bigId", Type: "int64"}
	schema = fieldToSchema(int64Field)
	assert.Equal(t, "integer", schema.Type)
	assert.Equal(t, "int64", schema.Format)

	// Test float
	floatField := resource.Field{Name: "price", Type: "float32"}
	schema = fieldToSchema(floatField)
	assert.Equal(t, "number", schema.Type)
	assert.Equal(t, "float", schema.Format)

	// Test boolean
	boolField := resource.Field{Name: "active", Type: "bool"}
	schema = fieldToSchema(boolField)
	assert.Equal(t, "boolean", schema.Type)

	// Test time
	timeField := resource.Field{Name: "created_at", Type: "time.Time"}
	schema = fieldToSchema(timeField)
	assert.Equal(t, "string", schema.Type)
	assert.Equal(t, "date-time", schema.Format)

	// Test array of primitives
	arrayField := resource.Field{Name: "tags", Type: "[]string"}
	schema = fieldToSchema(arrayField)
	assert.Equal(t, "array", schema.Type)
	assert.Equal(t, "string", schema.Items.Type)

	// Test complex type
	complexField := resource.Field{Name: "user", Type: "User"}
	schema = fieldToSchema(complexField)
	assert.Equal(t, "#/components/schemas/User", schema.Ref)

	// Test array of complex types
	complexArrayField := resource.Field{Name: "users", Type: "[]User"}
	schema = fieldToSchema(complexArrayField)
	assert.Equal(t, "array", schema.Type)
	assert.Equal(t, "#/components/schemas/User", schema.Items.Ref)
}

func TestSwaggerHandler(t *testing.T) {
	handler := SwaggerHandler()
	assert.NotNil(t, handler, "Swagger handler should not be nil")
}
