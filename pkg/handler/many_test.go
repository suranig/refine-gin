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
	// Setup
	gin.SetMode(gin.TestMode)
	r, mockRepo, mockResource, _ := setupTest()

	// Setup test data and mock expectations
	mockRepo.On("DeleteMany", mock.Anything, mock.Anything).Return(int64(2), nil)

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
