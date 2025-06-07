package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

func TestCreateManyHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup mock expectations for bulk create
	testItems := []TestModel{
		{ID: "1", Name: "Test 1"},
		{ID: "2", Name: "Test 2"},
	}

	mockDTOProvider.On("TransformToModel", mock.Anything).Return(testItems, nil)
	mockDTOProvider.On("TransformFromModel", mock.Anything).Return(testItems, nil)
	mockRepo.On("CreateMany", mock.Anything, mock.Anything).Return(testItems, nil)

	// Setup routes
	r.POST("/tests/batch", GenerateCreateManyHandler(mockResource, mockRepo, mockDTOProvider))

	// Create a valid request
	reqBody := BulkCreateRequest{
		Values: []map[string]interface{}{
			{"id": "1", "name": "Test 1"},
			{"id": "2", "name": "Test 2"},
		},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusCreated, resp.Code)

	var jsonResp BulkResponse
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)

	// Verify the response contains the created items
	assert.NotNil(t, jsonResp.Data)

	// Verify mocks were called as expected
	mockDTOProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestUpdateManyHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup test data and mock expectations
	updateDTO := TestModel{Name: "Updated Name"}
	mockDTOProvider.On("GetUpdateDTO").Return(&updateDTO)
	mockDTOProvider.On("TransformToModel", mock.Anything).Return(updateDTO, nil)
	mockRepo.On("UpdateMany", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)
	mockResource.On("GetEditableFields").Return([]string{"name"})

	// Setup routes
	r.PUT("/tests/batch", GenerateUpdateManyHandler(mockResource, mockRepo, mockDTOProvider))

	// Create a valid request
	reqBody := BulkUpdateRequest{
		IDs: []string{"1", "2"},
		Values: map[string]interface{}{
			"name": "Updated Name",
		},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var jsonResp map[string]map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)

	// Verify the response contains the expected count
	assert.Equal(t, float64(2), jsonResp["data"]["count"])

	// Verify mocks were called
	mockDTOProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestDeleteManyHandler(t *testing.T) {
	// Setup with custom setup function to avoid Query expectation
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockRepo := new(MockRepository)
	mockResource := new(MockResource)

	// Setup resource (skip the Query expectation)
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetLabel").Return("Tests")
	mockResource.On("GetIcon").Return("test-icon")
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

	// Setup test data and mock expectations
	mockRepo.On("DeleteMany", mock.Anything, mock.Anything).Return(int64(2), nil)

	// Debugging: Print out all expectations
	for _, exp := range mockRepo.ExpectedCalls {
		t.Logf("Expected call: %s", exp.Method)
	}

	// Setup routes
	r.DELETE("/tests/batch", GenerateDeleteManyHandler(mockResource, mockRepo))

	// Create a valid request
	reqBody := BulkDeleteRequest{
		IDs: []string{"1", "2"},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var jsonResp map[string]map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)

	// Verify the response contains the expected count
	assert.Equal(t, float64(2), jsonResp["data"]["count"])

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
}

func TestCreateManyHandler_Error(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Mock an error in repository
	mockDTOProvider.On("TransformToModel", mock.Anything).Return(nil, errors.New("transform error"))

	// Setup routes
	r.POST("/tests/batch", GenerateCreateManyHandler(mockResource, mockRepo, mockDTOProvider))

	// Create a request
	reqBody := BulkCreateRequest{
		Values: []map[string]interface{}{
			{"id": "1", "name": "Test 1"},
		},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response indicates error
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Contains(t, jsonResp["error"], "transform error")
}

func TestUpdateManyHandler_Error(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Mock an error in repository
	mockDTOProvider.On("GetUpdateDTO").Return(&TestModel{})
	mockDTOProvider.On("TransformToModel", mock.Anything).Return(nil, errors.New("transform error"))
	mockResource.On("GetEditableFields").Return([]string{"name"})

	// Setup routes
	r.PUT("/tests/batch", GenerateUpdateManyHandler(mockResource, mockRepo, mockDTOProvider))

	// Create a request
	reqBody := BulkUpdateRequest{
		IDs: []string{"1", "2"},
		Values: map[string]interface{}{
			"name": "Updated Name",
		},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response indicates error
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Contains(t, jsonResp["error"], "transform error")
}

func TestUpdateManyHandler_RepositoryError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup test data and mock expectations
	updateDTO := TestModel{Name: "Updated Name"}
	mockDTOProvider.On("GetUpdateDTO").Return(&updateDTO)
	mockDTOProvider.On("TransformToModel", mock.Anything).Return(updateDTO, nil)
	mockRepo.On("UpdateMany", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), errors.New("database error"))
	mockResource.On("GetEditableFields").Return([]string{"name"})

	// Setup routes
	r.PUT("/tests/batch", GenerateUpdateManyHandler(mockResource, mockRepo, mockDTOProvider))

	// Create a valid request
	reqBody := BulkUpdateRequest{
		IDs: []string{"1", "2"},
		Values: map[string]interface{}{
			"name": "Updated Name",
		},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Equal(t, "database error", jsonResp["error"])
}

func TestUpdateManyHandler_InvalidRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockResource := new(MockResource)
	mockRepo := new(MockRepository)

	// Setup resource
	mockResource.On("GetName").Return("tests")

	// Setup routes
	r.PUT("/tests/batch", GenerateUpdateManyHandler(mockResource, mockRepo, nil))

	// Create an invalid request (invalid JSON)
	reqData := []byte(`{"ids": [1,2], "values": invalid_json}`)
	req, _ := http.NewRequest("PUT", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response indicates error
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestUpdateManyHandler_InvalidIDsFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, _, mockResource, mockDTOProvider := setupTest()

	mockResource.On("GetEditableFields").Return([]string{"name"})

	r.PUT("/tests/batch", GenerateUpdateManyHandler(mockResource, nil, mockDTOProvider))

	reqBody := BulkUpdateRequest{
		Values: map[string]interface{}{
			"name": "Updated",
		},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Equal(t, "IDs must be an array or a single value", jsonResp["error"])
}

func TestDeleteManyHandler_Error(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockRepo := new(MockRepository)
	mockResource := new(MockResource)

	// Setup resource
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetIDFieldName").Return("ID")

	// Mock an error in repository
	mockRepo.On("DeleteMany", mock.Anything, mock.Anything).Return(int64(0), errors.New("delete error"))

	// Setup routes
	r.DELETE("/tests/batch", GenerateDeleteManyHandler(mockResource, mockRepo))

	// Create a request
	reqBody := BulkDeleteRequest{
		IDs: []string{"1", "2"},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response indicates error
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Equal(t, "delete error", jsonResp["error"])
}

func TestDeleteManyHandler_InvalidRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockResource := new(MockResource)

	// Setup resource
	mockResource.On("GetName").Return("tests")

	// Setup routes
	r.DELETE("/tests/batch", GenerateDeleteManyHandler(mockResource, nil))

	// Create an invalid request (invalid JSON)
	reqData := []byte(`{"ids": invalid_json}`)
	req, _ := http.NewRequest("DELETE", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response indicates error
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestDeleteManyHandler_InvalidIDsFormat(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockResource := new(MockResource)

	mockResource.On("GetName").Return("tests")
	mockResource.On("GetIDFieldName").Return("ID")

	r.DELETE("/tests/batch", GenerateDeleteManyHandler(mockResource, nil))

	reqBody := BulkDeleteRequest{}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Equal(t, "IDs must be an array or a single value", jsonResp["error"])
}

func TestDeleteManyHandler_SingleIDValue(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockRepo := new(MockRepository)
	mockResource := new(MockResource)

	// Setup resource
	mockResource.On("GetName").Return("tests")
	mockResource.On("GetIDFieldName").Return("ID")

	// Setup test data and mock expectations
	mockRepo.On("DeleteMany", mock.Anything, mock.Anything).Return(int64(1), nil)

	// Setup routes
	r.DELETE("/tests/batch", GenerateDeleteManyHandler(mockResource, mockRepo))

	// Create a request with a single ID value instead of an array
	reqBody := BulkDeleteRequest{
		IDs: "1", // Single ID as string
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("DELETE", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusOK, resp.Code)

	var jsonResp map[string]map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)

	// Verify the response contains the expected count
	assert.Equal(t, float64(1), jsonResp["data"]["count"])

	// Verify mocks were called
	mockRepo.AssertExpectations(t)
}

func TestCreateManyHandler_InvalidRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, _, mockResource, mockDTOProvider := setupTest()

	// Setup routes
	r.POST("/tests/batch", GenerateCreateManyHandler(mockResource, nil, mockDTOProvider))

	// Create an invalid request (missing values)
	reqBody := map[string]interface{}{
		// missing 'values' field
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response indicates error
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestCreateManyHandler_NonArrayValues(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, _, mockResource, mockDTOProvider := setupTest()

	// Setup routes
	r.POST("/tests/batch", GenerateCreateManyHandler(mockResource, nil, mockDTOProvider))

	// Create an invalid request (values is not an array)
	reqBody := BulkCreateRequest{
		Values: "not an array",
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response indicates error
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Equal(t, "values must be an array", jsonResp["error"])
}

func TestCreateManyHandler_RepositoryError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup test data and mock expectations
	testItems := []TestModel{
		{ID: "1", Name: "Test 1"},
		{ID: "2", Name: "Test 2"},
	}

	mockDTOProvider.On("TransformToModel", mock.Anything).Return(testItems, nil)
	mockRepo.On("CreateMany", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))

	// Setup routes
	r.POST("/tests/batch", GenerateCreateManyHandler(mockResource, mockRepo, mockDTOProvider))

	// Create a valid request
	reqBody := BulkCreateRequest{
		Values: []map[string]interface{}{
			{"id": "1", "name": "Test 1"},
			{"id": "2", "name": "Test 2"},
		},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Equal(t, "database error", jsonResp["error"])
}

func TestCreateManyHandler_DTOResponseError(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup test data and mock expectations
	testItems := []TestModel{
		{ID: "1", Name: "Test 1"},
		{ID: "2", Name: "Test 2"},
	}

	mockDTOProvider.On("TransformToModel", mock.Anything).Return(testItems, nil)
	mockRepo.On("CreateMany", mock.Anything, mock.Anything).Return(testItems, nil)
	mockDTOProvider.On("TransformFromModel", mock.Anything).Return(nil, errors.New("transform response error"))

	// Setup routes
	r.POST("/tests/batch", GenerateCreateManyHandler(mockResource, mockRepo, mockDTOProvider))

	// Create a valid request
	reqBody := BulkCreateRequest{
		Values: []map[string]interface{}{
			{"id": "1", "name": "Test 1"},
			{"id": "2", "name": "Test 2"},
		},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Equal(t, "transform response error", jsonResp["error"])
}

func TestCreateManyHandler_WithRelationValidation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r, mockRepo, mockResource, mockDTOProvider := setupTest()

	// Setup test data
	testItems := []TestModel{
		{ID: "1", Name: "Test 1"},
		{ID: "2", Name: "Test 2"},
	}

	// Setup mock db connection
	db := &gorm.DB{}
	mockRepo.On("Query", mock.Anything).Return(db)

	// Setup relations
	relations := []resource.Relation{
		{
			Name: "related",
			Type: "hasMany",
		},
	}
	mockResource.On("GetRelations").Return(relations)

	// Normal flow expectations
	mockDTOProvider.On("TransformToModel", mock.Anything).Return(testItems, nil)
	mockRepo.On("CreateMany", mock.Anything, mock.Anything).Return(testItems, nil)
	mockDTOProvider.On("TransformFromModel", mock.Anything).Return(testItems, nil)

	// Setup routes
	r.POST("/tests/batch", GenerateCreateManyHandler(mockResource, mockRepo, mockDTOProvider))

	// Create a valid request
	reqBody := BulkCreateRequest{
		Values: []map[string]interface{}{
			{"id": "1", "name": "Test 1"},
			{"id": "2", "name": "Test 2"},
		},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(resp, req)

	// Check response
	assert.Equal(t, http.StatusCreated, resp.Code)

	var jsonResp BulkResponse
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.NotNil(t, jsonResp.Data)

	// Verify mocks were called
	mockDTOProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestCreateManyHandler_RelationValidationFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockRepo := new(MockRepository)
	mockDTOProvider := new(MockDTOProvider)

	type RelModel struct {
		ID       string
		Category *string
	}

	res := &resource.DefaultResource{
		Name:  "relmodels",
		Model: RelModel{},
		Relations: []resource.Relation{
			{
				Name:     "category",
				Type:     resource.RelationTypeManyToOne,
				Resource: "categories",
				Field:    "Category",
				Required: true,
			},
		},
	}

	resource.GlobalResourceRegistry = resource.NewResourceRegistry()
	resource.GlobalResourceRegistry.Register(res)

	items := []RelModel{{ID: "1", Category: nil}}
	mockDTOProvider.On("TransformToModel", mock.Anything).Return(items, nil)
	mockRepo.On("Query", mock.Anything).Return(nil)

	r.POST("/tests/batch", GenerateCreateManyHandler(res, mockRepo, mockDTOProvider))

	reqBody := BulkCreateRequest{
		Values: []map[string]interface{}{{"id": "1"}},
	}

	reqData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tests/batch", bytes.NewBuffer(reqData))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var jsonResp map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResp)
	assert.NoError(t, err)
	assert.Contains(t, jsonResp["error"], "relation category is required")
}

// MockRepository with CreateMany, UpdateMany, and DeleteMany methods
func (m *MockRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	args := m.Called(ctx, data)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	args := m.Called(ctx, ids, data)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int64), args.Error(1)
}
