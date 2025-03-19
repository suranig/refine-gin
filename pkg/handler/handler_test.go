package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
)

// Mock repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error) {
	args := m.Called(ctx, options)
	return args.Get(0), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	args := m.Called(ctx, data)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	args := m.Called(ctx, id, data)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id interface{}) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Count returns the total number of resources matching the query options
func (m *MockRepository) Count(ctx context.Context, options query.QueryOptions) (int64, error) {
	args := m.Called(ctx, options)
	return args.Get(0).(int64), args.Error(1)
}

// Mock resource for testing
type MockResource struct {
	mock.Mock
}

func (m *MockResource) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockResource) GetModel() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockResource) GetFields() []resource.Field {
	args := m.Called()
	return args.Get(0).([]resource.Field)
}

func (m *MockResource) GetOperations() []resource.Operation {
	args := m.Called()
	return args.Get(0).([]resource.Operation)
}

func (m *MockResource) HasOperation(op resource.Operation) bool {
	args := m.Called(op)
	return args.Bool(0)
}

func (m *MockResource) GetDefaultSort() *resource.Sort {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*resource.Sort)
}

func (m *MockResource) GetFilters() []resource.Filter {
	args := m.Called()
	return args.Get(0).([]resource.Filter)
}

func (m *MockResource) GetMiddlewares() []interface{} {
	args := m.Called()
	return args.Get(0).([]interface{})
}

func (m *MockResource) GetRelations() []resource.Relation {
	args := m.Called()
	if args.Get(0) == nil {
		return []resource.Relation{}
	}
	return args.Get(0).([]resource.Relation)
}

func (m *MockResource) HasRelation(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

func (m *MockResource) GetRelation(name string) *resource.Relation {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	relation := args.Get(0).(resource.Relation)
	return &relation
}

func (m *MockResource) GetIDFieldName() string {
	args := m.Called()
	return args.String(0)
}

// GetField returns a field by name
func (m *MockResource) GetField(name string) *resource.Field {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	field := args.Get(0).(resource.Field)
	return &field
}

// GetSearchable returns searchable field names
func (m *MockResource) GetSearchable() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// Mock DTO provider for testing
type MockDTOProvider struct {
	mock.Mock
}

func (m *MockDTOProvider) GetCreateDTO() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockDTOProvider) GetUpdateDTO() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockDTOProvider) GetResponseDTO() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockDTOProvider) TransformToModel(dto interface{}) (interface{}, error) {
	args := m.Called(dto)
	return args.Get(0), args.Error(1)
}

func (m *MockDTOProvider) TransformFromModel(model interface{}) (interface{}, error) {
	args := m.Called(model)
	return args.Get(0), args.Error(1)
}

// Test model
type TestModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Setup test environment
func setupTest() (*gin.Engine, *MockRepository, *MockResource, *MockDTOProvider) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockRepo := new(MockRepository)
	mockResource := new(MockResource)
	mockDTOProvider := new(MockDTOProvider)

	// Setup resource
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetModel").Return(TestModel{})
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	})
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetRelations").Return([]resource.Relation{})
	mockResource.On("HasRelation", mock.Anything).Return(false)
	mockResource.On("GetRelation", mock.Anything).Return(nil)
	mockResource.On("GetIDFieldName").Return("ID")
	mockResource.On("GetField", mock.Anything).Return(nil)
	mockResource.On("GetSearchable").Return([]string{})

	return r, mockRepo, mockResource, mockDTOProvider
}

func TestListHandler(t *testing.T) {
	r, mockRepo, mockResource, _ := setupTest()

	// Setup mock response
	mockData := []TestModel{
		{ID: "1", Name: "Test 1"},
		{ID: "2", Name: "Test 2"},
	}
	mockRepo.On("List", mock.Anything, mock.Anything).Return(mockData, int64(2), nil)

	// Register handler
	r.GET("/tests", GenerateListHandler(mockResource, mockRepo))

	// Create request
	req, _ := http.NewRequest("GET", "/tests", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "total")
	assert.Equal(t, float64(2), response["total"])
}

func TestGetHandler(t *testing.T) {
	r, mockRepo, mockResource, _ := setupTest()

	// Setup mock response
	mockData := TestModel{ID: "1", Name: "Test 1"}
	mockRepo.On("Get", mock.Anything, "1").Return(mockData, nil)

	// Register handler
	r.GET("/tests/:id", GenerateGetHandler(mockResource, mockRepo))

	// Create request
	req, _ := http.NewRequest("GET", "/tests/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "1", data["id"])
	assert.Equal(t, "Test 1", data["name"])
}

func TestCreateHandler(t *testing.T) {
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup test input
	testInput := TestModel{Name: "New Test"}

	// Setup mock response
	mockResult := TestModel{ID: "3", Name: "New Test"}

	// Setup DTO provider
	mockDTOProvider.On("GetCreateDTO").Return(&TestModel{})
	mockDTOProvider.On("TransformToModel", mock.Anything).Return(testInput, nil)
	mockDTOProvider.On("TransformFromModel", mock.Anything).Return(mockResult, nil)

	// Setup repository
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(mockResult, nil)

	// Register handler
	r.POST("/tests", GenerateCreateHandler(mockResource, mockRepo, mockDTOProvider))

	// Create request
	body, _ := json.Marshal(testInput)
	req, _ := http.NewRequest("POST", "/tests", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "3", data["id"])
	assert.Equal(t, "New Test", data["name"])
}

func TestUpdateHandler(t *testing.T) {
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup test input
	testInput := TestModel{Name: "Updated Test"}

	// Setup mock response
	mockResult := TestModel{ID: "1", Name: "Updated Test"}

	// Setup DTO provider
	mockDTOProvider.On("GetUpdateDTO").Return(&TestModel{})
	mockDTOProvider.On("TransformToModel", mock.Anything).Return(testInput, nil)
	mockDTOProvider.On("TransformFromModel", mock.Anything).Return(mockResult, nil)

	// Setup repository
	mockRepo.On("Update", mock.Anything, "1", mock.Anything).Return(mockResult, nil)

	// Register handler
	r.PUT("/tests/:id", GenerateUpdateHandler(mockResource, mockRepo, mockDTOProvider))

	// Create request
	body, _ := json.Marshal(testInput)
	req, _ := http.NewRequest("PUT", "/tests/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "1", data["id"])
	assert.Equal(t, "Updated Test", data["name"])
}

func TestDeleteHandler(t *testing.T) {
	r, mockRepo, mockResource, _ := setupTest()

	// Setup repository
	mockRepo.On("Delete", mock.Anything, "1").Return(nil)

	// Register handler
	r.DELETE("/tests/:id", GenerateDeleteHandler(mockResource, mockRepo))

	// Create request
	req, _ := http.NewRequest("DELETE", "/tests/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "success")
	assert.Equal(t, true, response["success"])
}

func TestRegisterResource(t *testing.T) {
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup resource operations
	mockResource.On("HasOperation", resource.OperationList).Return(true)
	mockResource.On("HasOperation", resource.OperationRead).Return(true)
	mockResource.On("HasOperation", resource.OperationCreate).Return(true)
	mockResource.On("HasOperation", resource.OperationUpdate).Return(true)
	mockResource.On("HasOperation", resource.OperationDelete).Return(true)
	mockResource.On("HasOperation", resource.OperationCount).Return(true)

	// Register resource
	api := r.Group("/api")
	RegisterResourceWithDTO(api, mockResource, mockRepo, mockDTOProvider)

	// Test routes exist
	routes := r.Routes()

	// Check if all routes are registered
	foundRoutes := map[string]bool{
		"GET /api/tests":        false,
		"GET /api/tests/:id":    false,
		"POST /api/tests":       false,
		"PUT /api/tests/:id":    false,
		"DELETE /api/tests/:id": false,
	}

	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, exists := foundRoutes[key]; exists {
			foundRoutes[key] = true
		}
	}

	// Assert all routes are registered
	for route, found := range foundRoutes {
		assert.True(t, found, "Route %s not found", route)
	}
}

func TestGetHandlerWithParam(t *testing.T) {
	r, mockRepo, mockResource, _ := setupTest()

	// Setup mock response
	mockData := TestModel{ID: "1", Name: "Test 1"}
	mockRepo.On("Get", mock.Anything, "1").Return(mockData, nil)

	// Register handler with custom parameter name
	r.GET("/tests/:uid", GenerateGetHandlerWithParam(mockResource, mockRepo, "uid"))

	// Create request
	req, _ := http.NewRequest("GET", "/tests/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "1", data["id"])
	assert.Equal(t, "Test 1", data["name"])
}

func TestUpdateHandlerWithParam(t *testing.T) {
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup test input
	testInput := TestModel{Name: "Updated Test"}

	// Setup mock response
	mockResult := TestModel{ID: "1", Name: "Updated Test"}

	// Setup DTO provider
	mockDTOProvider.On("GetUpdateDTO").Return(&TestModel{})
	mockDTOProvider.On("TransformToModel", mock.Anything).Return(testInput, nil)
	mockDTOProvider.On("TransformFromModel", mock.Anything).Return(mockResult, nil)

	// Setup repository
	mockRepo.On("Update", mock.Anything, "1", mock.Anything).Return(mockResult, nil)

	// Register handler with custom parameter name
	r.PUT("/tests/:uid", GenerateUpdateHandlerWithParam(mockResource, mockRepo, mockDTOProvider, "uid"))

	// Create request
	body, _ := json.Marshal(testInput)
	req, _ := http.NewRequest("PUT", "/tests/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "1", data["id"])
	assert.Equal(t, "Updated Test", data["name"])
}

func TestDeleteHandlerWithParam(t *testing.T) {
	r, mockRepo, mockResource, _ := setupTest()

	// Setup repository
	mockRepo.On("Delete", mock.Anything, "1").Return(nil)

	// Register handler with custom parameter name
	r.DELETE("/tests/:uid", GenerateDeleteHandlerWithParam(mockResource, mockRepo, "uid"))

	// Create request
	req, _ := http.NewRequest("DELETE", "/tests/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "success")
	assert.Equal(t, true, response["success"])
}

func TestRegisterResourceWithOptions(t *testing.T) {
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup resource operations
	mockResource.On("HasOperation", resource.OperationList).Return(true)
	mockResource.On("HasOperation", resource.OperationRead).Return(true)
	mockResource.On("HasOperation", resource.OperationCreate).Return(true)
	mockResource.On("HasOperation", resource.OperationUpdate).Return(true)
	mockResource.On("HasOperation", resource.OperationDelete).Return(true)
	mockResource.On("HasOperation", resource.OperationCount).Return(true)

	// Register resource with custom ID parameter name
	api := r.Group("/api")
	RegisterResourceWithOptions(api, mockResource, mockRepo, RegisterOptionsToResourceOptions(RegisterOptions{
		DTOProvider: mockDTOProvider,
		IDParamName: "uid",
	}))

	// Test routes exist
	routes := r.Routes()

	// Check if all routes are registered with custom ID parameter
	foundRoutes := map[string]bool{
		"GET /api/tests":         false,
		"GET /api/tests/:uid":    false,
		"POST /api/tests":        false,
		"PUT /api/tests/:uid":    false,
		"DELETE /api/tests/:uid": false,
		"GET /api/tests/count":   false,
	}

	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, exists := foundRoutes[key]; exists {
			foundRoutes[key] = true
		}
	}

	// Assert all routes are registered
	for route, found := range foundRoutes {
		assert.True(t, found, "Route %s not found", route)
	}
}
