package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/dto"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"github.com/suranig/refine-gin/pkg/utils"
)

// Mock repository for testing
type MockRepository struct {
	createManyFunc func(ctx context.Context, data interface{}) (interface{}, error)
	updateManyFunc func(ctx context.Context, ids []interface{}, data interface{}) (int64, error)
	deleteManyFunc func(ctx context.Context, ids []interface{}) (int64, error)
}

func (m *MockRepository) List(ctx context.Context, options interface{}) (interface{}, int64, error) {
	return nil, 0, nil
}

func (m *MockRepository) Get(ctx context.Context, id interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockRepository) Create(ctx context.Context, data interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockRepository) Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockRepository) Delete(ctx context.Context, id interface{}) error {
	return nil
}

func (m *MockRepository) Count(ctx context.Context, options interface{}) (int64, error) {
	return 0, nil
}

func (m *MockRepository) CreateMany(ctx context.Context, data interface{}) (interface{}, error) {
	if m.createManyFunc != nil {
		return m.createManyFunc(ctx, data)
	}
	return nil, nil
}

func (m *MockRepository) UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
	if m.updateManyFunc != nil {
		return m.updateManyFunc(ctx, ids, data)
	}
	return 0, nil
}

func (m *MockRepository) DeleteMany(ctx context.Context, ids []interface{}) (int64, error) {
	if m.deleteManyFunc != nil {
		return m.deleteManyFunc(ctx, ids)
	}
	return 0, nil
}

func (m *MockRepository) WithTransaction(fn func(repository.Repository) error) error {
	return fn(m)
}

func (m *MockRepository) WithRelations(relations ...string) repository.Repository {
	return m
}

func (m *MockRepository) FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockRepository) FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockRepository) Query(ctx context.Context) interface{} {
	return nil
}

func (m *MockRepository) BulkCreate(ctx context.Context, data interface{}) error {
	return nil
}

func (m *MockRepository) BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	return nil
}

func (m *MockRepository) GetIDFieldName() string {
	return "id"
}

// Test model
type TestModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestGenerateUpdateManyHandler_ValidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test resource
	res := resource.NewResource(resource.ResourceConfig{
		Name: "tests",
		Model: TestModel{},
		Operations: []resource.Operation{
			resource.OperationUpdateMany,
		},
	})

	// Create mock repository
	mockRepo := &MockRepository{
		updateManyFunc: func(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
			// Verify IDs are properly parsed
			if len(ids) != 2 {
				t.Errorf("Expected 2 IDs, got %d", len(ids))
			}
			if ids[0] != "1" || ids[1] != "2" {
				t.Errorf("Expected IDs [1, 2], got %v", ids)
			}
			return 2, nil
		},
	}

	// Create handler
	handler := GenerateUpdateManyHandler(res, mockRepo, nil)

	// Create request
	reqBody := BulkUpdateRequest{
		IDs: []string{"1", "2"},
		Values: map[string]interface{}{
			"name": "updated",
		},
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Create HTTP request
	req, _ := http.NewRequest("PUT", "/batch", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler(c)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["data"].(map[string]interface{})["count"].(float64) != 2 {
		t.Errorf("Expected count 2, got %v", response["data"].(map[string]interface{})["count"])
	}
}

func TestGenerateUpdateManyHandler_InvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test resource
	res := resource.NewResource(resource.ResourceConfig{
		Name: "tests",
		Model: TestModel{},
		Operations: []resource.Operation{
			resource.OperationUpdateMany,
		},
	})

	// Create mock repository
	mockRepo := &MockRepository{}

	// Create handler
	handler := GenerateUpdateManyHandler(res, mockRepo, nil)

	testCases := []struct {
		name     string
		ids      interface{}
		expected string
	}{
		{
			name:     "empty IDs array",
			ids:      []string{},
			expected: utils.ErrCodeEmptyID,
		},
		{
			name:     "nil IDs",
			ids:      nil,
			expected: utils.ErrCodeEmptyID,
		},
		{
			name:     "empty string ID",
			ids:      []string{"1", "", "3"},
			expected: utils.ErrCodeEmptyID,
		},
		{
			name:     "duplicate IDs",
			ids:      []string{"1", "1", "2"},
			expected: utils.ErrCodeDuplicateID,
		},
		{
			name:     "invalid string ID",
			ids:      []string{"1", "invalid@id", "3"},
			expected: utils.ErrCodeInvalidIDValue,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			reqBody := BulkUpdateRequest{
				IDs: tc.ids,
				Values: map[string]interface{}{
					"name": "updated",
				},
			}
			reqJSON, _ := json.Marshal(reqBody)

			// Create HTTP request
			req, _ := http.NewRequest("PUT", "/batch", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute handler
			handler(c)

			// Check response
			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			errorObj := response["error"].(map[string]interface{})
			if errorObj["code"] != tc.expected {
				t.Errorf("Expected error code %s, got %s", tc.expected, errorObj["code"])
			}
		})
	}
}

func TestGenerateDeleteManyHandler_ValidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test resource
	res := resource.NewResource(resource.ResourceConfig{
		Name: "tests",
		Model: TestModel{},
		Operations: []resource.Operation{
			resource.OperationDeleteMany,
		},
	})

	// Create mock repository
	mockRepo := &MockRepository{
		deleteManyFunc: func(ctx context.Context, ids []interface{}) (int64, error) {
			// Verify IDs are properly parsed
			if len(ids) != 3 {
				t.Errorf("Expected 3 IDs, got %d", len(ids))
			}
			if ids[0] != "1" || ids[1] != "2" || ids[2] != "3" {
				t.Errorf("Expected IDs [1, 2, 3], got %v", ids)
			}
			return 3, nil
		},
	}

	// Create handler
	handler := GenerateDeleteManyHandler(res, mockRepo)

	// Create request
	reqBody := BulkDeleteRequest{
		IDs: []string{"1", "2", "3"},
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Create HTTP request
	req, _ := http.NewRequest("DELETE", "/batch", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler(c)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["data"].(map[string]interface{})["count"].(float64) != 3 {
		t.Errorf("Expected count 3, got %v", response["data"].(map[string]interface{})["count"])
	}
}

func TestGenerateDeleteManyHandler_InvalidIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test resource
	res := resource.NewResource(resource.ResourceConfig{
		Name: "tests",
		Model: TestModel{},
		Operations: []resource.Operation{
			resource.OperationDeleteMany,
		},
	})

	// Create mock repository
	mockRepo := &MockRepository{}

	// Create handler
	handler := GenerateDeleteManyHandler(res, mockRepo)

	testCases := []struct {
		name     string
		ids      interface{}
		expected string
	}{
		{
			name:     "empty IDs array",
			ids:      []string{},
			expected: utils.ErrCodeEmptyID,
		},
		{
			name:     "nil IDs",
			ids:      nil,
			expected: utils.ErrCodeEmptyID,
		},
		{
			name:     "duplicate IDs",
			ids:      []string{"1", "1", "2"},
			expected: utils.ErrCodeDuplicateID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			reqBody := BulkDeleteRequest{
				IDs: tc.ids,
			}
			reqJSON, _ := json.Marshal(reqBody)

			// Create HTTP request
			req, _ := http.NewRequest("DELETE", "/batch", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute handler
			handler(c)

			// Check response
			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 400, got %d", w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			errorObj := response["error"].(map[string]interface{})
			if errorObj["code"] != tc.expected {
				t.Errorf("Expected error code %s, got %s", tc.expected, errorObj["code"])
			}
		})
	}
}

func TestGenerateUpdateManyHandler_DifferentIDFormats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test resource
	res := resource.NewResource(resource.ResourceConfig{
		Name: "tests",
		Model: TestModel{},
		Operations: []resource.Operation{
			resource.OperationUpdateMany,
		},
	})

	// Create mock repository
	mockRepo := &MockRepository{
		updateManyFunc: func(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
			return int64(len(ids)), nil
		},
	}

	// Create handler
	handler := GenerateUpdateManyHandler(res, mockRepo, nil)

	testCases := []struct {
		name string
		ids  interface{}
	}{
		{"string slice", []string{"1", "2", "3"}},
		{"int slice", []int{1, 2, 3}},
		{"mixed slice", []interface{}{"1", 2, "3"}},
		{"single string", "1"},
		{"single int", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			reqBody := BulkUpdateRequest{
				IDs: tc.ids,
				Values: map[string]interface{}{
					"name": "updated",
				},
			}
			reqJSON, _ := json.Marshal(reqBody)

			// Create HTTP request
			req, _ := http.NewRequest("PUT", "/batch", bytes.NewBuffer(reqJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Execute handler
			handler(c)

			// Check response
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestGenerateUpdateManyHandler_RepositoryError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test resource
	res := resource.NewResource(resource.ResourceConfig{
		Name: "tests",
		Model: TestModel{},
		Operations: []resource.Operation{
			resource.OperationUpdateMany,
		},
	})

	// Create mock repository with error
	mockRepo := &MockRepository{
		updateManyFunc: func(ctx context.Context, ids []interface{}, data interface{}) (int64, error) {
			return 0, fmt.Errorf("database error")
		},
	}

	// Create handler
	handler := GenerateUpdateManyHandler(res, mockRepo, nil)

	// Create request
	reqBody := BulkUpdateRequest{
		IDs: []string{"1", "2"},
		Values: map[string]interface{}{
			"name": "updated",
		},
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Create HTTP request
	req, _ := http.NewRequest("PUT", "/batch", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Execute handler
	handler(c)

	// Check response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "database error" {
		t.Errorf("Expected error 'database error', got %v", response["error"])
	}
}
