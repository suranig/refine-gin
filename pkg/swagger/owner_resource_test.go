package swagger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/resource"
)

// MockOwnerResource mocks the OwnerResource interface
type MockOwnerResource struct {
	mock.Mock
}

func (m *MockOwnerResource) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOwnerResource) GetLabel() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOwnerResource) GetIcon() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOwnerResource) GetModel() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockOwnerResource) GetFields() []resource.Field {
	args := m.Called()
	return args.Get(0).([]resource.Field)
}

func (m *MockOwnerResource) GetOperations() []resource.Operation {
	args := m.Called()
	return args.Get(0).([]resource.Operation)
}

func (m *MockOwnerResource) HasOperation(op resource.Operation) bool {
	args := m.Called(op)
	return args.Bool(0)
}

func (m *MockOwnerResource) GetDefaultSort() *resource.Sort {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.Sort)
}

func (m *MockOwnerResource) GetFilters() []resource.Filter {
	args := m.Called()
	return args.Get(0).([]resource.Filter)
}

func (m *MockOwnerResource) GetMiddlewares() []interface{} {
	args := m.Called()
	return args.Get(0).([]interface{})
}

func (m *MockOwnerResource) GetRelations() []resource.Relation {
	args := m.Called()
	if args.Get(0) == nil {
		return []resource.Relation{}
	}
	return args.Get(0).([]resource.Relation)
}

func (m *MockOwnerResource) HasRelation(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *MockOwnerResource) GetRelation(name string) *resource.Relation {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	relation := args.Get(0).(resource.Relation)
	return &relation
}

func (m *MockOwnerResource) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOwnerResource) GetField(name string) *resource.Field {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	field := args.Get(0).(resource.Field)
	return &field
}

func (m *MockOwnerResource) GetSearchable() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockOwnerResource) GetFilterableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockOwnerResource) GetSortableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockOwnerResource) GetRequiredFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockOwnerResource) GetTableFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockOwnerResource) GetOwnerField() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOwnerResource) EnforceOwnership() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOwnerResource) GetDefaultOwnerID() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockOwnerResource) GetFormFields() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockOwnerResource) GetOwnerConfig() resource.OwnerConfig {
	args := m.Called()
	return args.Get(0).(resource.OwnerConfig)
}

func (m *MockOwnerResource) IsOwnershipEnforced() bool {
	args := m.Called()
	return args.Bool(0)
}

// TestRegisterOwnerResourceSwagger tests the RegisterOwnerResourceSwagger function
func TestRegisterOwnerResourceSwagger(t *testing.T) {
	// Create mock owner resource
	mockOwnerResource := new(MockOwnerResource)
	mockOwnerResource.On("GetName").Return("notes")
	mockOwnerResource.On("GetModel").Return(struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
		OwnerID string `json:"ownerId"`
	}{})
	mockOwnerResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "ownerId", Type: "string"},
	})
	mockOwnerResource.On("HasOperation", mock.Anything).Return(true)
	mockOwnerResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationList,
		resource.OperationRead,
		resource.OperationCreate,
		resource.OperationUpdate,
		resource.OperationDelete,
		resource.OperationCreateMany,
		resource.OperationUpdateMany,
		resource.OperationDeleteMany,
	})
	mockOwnerResource.On("GetDefaultSort").Return(nil)
	mockOwnerResource.On("GetFilters").Return([]resource.Filter{})
	mockOwnerResource.On("GetMiddlewares").Return([]interface{}{})
	mockOwnerResource.On("GetRelations").Return([]resource.Relation{})
	mockOwnerResource.On("GetOwnerField").Return("ownerId")
	mockOwnerResource.On("EnforceOwnership").Return(true)
	mockOwnerResource.On("GetIDFieldName").Return("id")
	mockOwnerResource.On("GetDefaultOwnerID").Return(nil)
	mockOwnerResource.On("GetFormFields").Return([]string{"title", "content"})
	mockOwnerResource.On("IsOwnershipEnforced").Return(true)
	mockOwnerResource.On("GetOwnerConfig").Return(resource.OwnerConfig{
		OwnerField:       "ownerId",
		EnforceOwnership: true,
		DefaultOwnerID:   nil,
	})

	// Create info
	info := DefaultSwaggerInfo()

	// Create OpenAPI spec
	openAPI := GenerateOpenAPI([]resource.Resource{}, info)

	// Register owner resource
	RegisterOwnerResourceSwagger(openAPI, mockOwnerResource)

	// ASSERTIONS

	// Check that security scheme was added
	assert.Contains(t, openAPI.Components.SecuritySchemes, "bearerAuth")
	assert.Equal(t, "http", openAPI.Components.SecuritySchemes["bearerAuth"].Type)
	assert.Equal(t, "bearer", openAPI.Components.SecuritySchemes["bearerAuth"].Scheme)

	// Check list endpoint
	listPath := "/notes"
	assert.Contains(t, openAPI.Paths, listPath)
	assert.Contains(t, openAPI.Paths[listPath], "get")
	listOp := openAPI.Paths[listPath]["get"]
	assert.Contains(t, listOp.Description, "Only resources owned by the authenticated user will be returned")
	assert.NotEmpty(t, listOp.Security)
	if assert.GreaterOrEqual(t, len(listOp.Security), 1) {
		securityItem := listOp.Security[0]
		assert.Contains(t, securityItem, "bearerAuth")
	}

	// Check get endpoint
	getPath := "/notes/{id}"
	assert.Contains(t, openAPI.Paths, getPath)
	assert.Contains(t, openAPI.Paths[getPath], "get")
	getOp := openAPI.Paths[getPath]["get"]
	assert.Contains(t, getOp.Description, "Only accessible if the resource is owned by the authenticated user")
	assert.Contains(t, getOp.Responses, "403")
	assert.Equal(t, "Forbidden - The resource is owned by another user", getOp.Responses["403"].Description)

	// Check update endpoint
	assert.Contains(t, openAPI.Paths[getPath], "put")
	updateOp := openAPI.Paths[getPath]["put"]
	assert.Contains(t, updateOp.Description, "Only resources owned by the authenticated user can be updated")
	assert.Contains(t, updateOp.Responses, "403")

	// Check delete endpoint
	assert.Contains(t, openAPI.Paths[getPath], "delete")
	deleteOp := openAPI.Paths[getPath]["delete"]
	assert.Contains(t, deleteOp.Description, "Only resources owned by the authenticated user can be deleted")
	assert.Contains(t, deleteOp.Responses, "403")

	// Check create endpoint
	assert.Contains(t, openAPI.Paths[listPath], "post")
	createOp := openAPI.Paths[listPath]["post"]
	assert.Contains(t, createOp.Description, "The owner ID will be automatically set to the authenticated user's ID")

	// Check batch endpoints
	batchPath := "/notes/batch"
	assert.Contains(t, openAPI.Paths, batchPath)

	// Check batch operations that exist in the path item
	pathItem := openAPI.Paths[batchPath]

	if operation, exists := pathItem["post"]; exists {
		assert.Contains(t, operation.Description, "The owner ID for all created resources will be set to the authenticated user's ID")
	}

	if operation, exists := pathItem["put"]; exists {
		assert.Contains(t, operation.Description, "Only resources owned by the authenticated user can be updated")
		assert.Contains(t, operation.Responses, "403")
	}

	if operation, exists := pathItem["delete"]; exists {
		assert.Contains(t, operation.Description, "Only resources owned by the authenticated user can be deleted")
		assert.Contains(t, operation.Responses, "403")
	}
}
