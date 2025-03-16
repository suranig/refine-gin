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

// Test model
type TestModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

func setupTest() (*gin.Engine, *MockRepository, *MockResource) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockRepo := new(MockRepository)
	mockResource := new(MockResource)

	// Setup default resource behavior
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetModel").Return(TestModel{})
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "string", Filterable: true},
		{Name: "name", Type: "string", Filterable: true, Searchable: true},
	})
	mockResource.On("GetDefaultSort").Return(nil)

	return r, mockRepo, mockResource
}

func TestGenerateListHandler(t *testing.T) {
	r, mockRepo, mockResource := setupTest()

	// Setup mock repository response
	testData := []TestModel{
		{ID: "1", Name: "Test 1"},
		{ID: "2", Name: "Test 2"},
	}
	mockRepo.On("List", mock.Anything, mock.Anything).Return(testData, int64(2), nil)

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

func TestGenerateGetHandler(t *testing.T) {
	r, mockRepo, mockResource := setupTest()

	// Setup mock repository response
	testData := TestModel{ID: "1", Name: "Test 1"}
	mockRepo.On("Get", mock.Anything, "1").Return(testData, nil)

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
}

func TestGenerateCreateHandler(t *testing.T) {
	r, mockRepo, mockResource := setupTest()

	// Setup mock repository response
	testInput := TestModel{Name: "New Test"}
	testOutput := TestModel{ID: "3", Name: "New Test"}
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(testOutput, nil)

	// Register handler
	r.POST("/tests", GenerateCreateHandler(mockResource, mockRepo))

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
}

func TestGenerateUpdateHandler(t *testing.T) {
	r, mockRepo, mockResource := setupTest()

	// Setup mock repository response
	testInput := TestModel{Name: "Updated Test"}
	testOutput := TestModel{ID: "1", Name: "Updated Test"}
	mockRepo.On("Update", mock.Anything, "1", mock.Anything).Return(testOutput, nil)

	// Register handler
	r.PUT("/tests/:id", GenerateUpdateHandler(mockResource, mockRepo))

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
}

func TestGenerateDeleteHandler(t *testing.T) {
	r, mockRepo, mockResource := setupTest()

	// Setup mock repository response
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
	r, mockRepo, mockResource := setupTest()

	// Setup resource operations
	mockResource.On("HasOperation", resource.OperationList).Return(true)
	mockResource.On("HasOperation", resource.OperationRead).Return(true)
	mockResource.On("HasOperation", resource.OperationCreate).Return(true)
	mockResource.On("HasOperation", resource.OperationUpdate).Return(true)
	mockResource.On("HasOperation", resource.OperationDelete).Return(true)

	// Setup mock repository responses
	mockRepo.On("List", mock.Anything, mock.Anything).Return([]TestModel{}, int64(0), nil)
	mockRepo.On("Get", mock.Anything, mock.Anything).Return(TestModel{}, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(TestModel{}, nil)
	mockRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(TestModel{}, nil)
	mockRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)

	// Register resource
	api := r.Group("/api")
	RegisterResource(api, mockResource, mockRepo)

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
