package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/resource"
)

// User model for testing updates
type UpdateUser struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateDTO represents a DTO for update
type UserUpdateDTO struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func TestGenerateUpdateHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock resource, repository and DTO manager
	mockResource := new(MockResource)
	mockRepo := new(MockRepository)
	mockDTOManager := new(MockDTOManager)

	// Setup resource expectations
	mockResource.On("GetName").Return("users").Maybe()
	mockResource.On("GetModel").Return(UpdateUser{}).Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()
	mockResource.On("GetRelations").Return([]resource.Relation{}).Maybe()

	// Test case 1: Successful update
	t.Run("Successful update", func(t *testing.T) {
		// Create test DTO
		updateDTO := &UserUpdateDTO{
			Name:  "Updated Name",
			Email: "updated@example.com",
		}

		// Create JSON payload
		jsonData, _ := json.Marshal(updateDTO)

		// Create model data for the transformation result
		modelData := map[string]interface{}{
			"name":  "Updated Name",
			"email": "updated@example.com",
		}

		// Create expected result
		updatedUser := UpdateUser{
			ID:        123,
			Name:      "Updated Name",
			Email:     "updated@example.com",
			UpdatedAt: time.Now(),
		}

		// Setup DTO manager expectations
		mockDTOManager.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()
		mockDTOManager.On("TransformToModel", mock.AnythingOfType("*handler.UserUpdateDTO")).
			Return(modelData, nil).Once()
		mockDTOManager.On("TransformFromModel", updatedUser).
			Return(updatedUser, nil).Once()

		// Setup repository expectations
		mockRepo.On("Query", mock.Anything).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, "123", modelData).
			Return(updatedUser, nil).Once()

		// Setup the handler
		r := gin.New()
		r.PUT("/users/:id", GenerateUpdateHandler(mockResource, mockRepo, mockDTOManager))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/123", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check response data - The handler should return the updated entity inside a 'data' field
		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok, "Response should contain data field with an object")

		// Field names match the struct field names in UpdateUser with JSON tags
		assert.Equal(t, float64(123), data["id"], "ID should match")
		assert.Equal(t, "Updated Name", data["name"], "Name should match")
		assert.Equal(t, "updated@example.com", data["email"], "Email should match")

		// Verify repository mock
		mockRepo.AssertExpectations(t)
		mockDTOManager.AssertExpectations(t)
	})

	// Test case 2: Invalid JSON payload
	t.Run("Invalid JSON payload", func(t *testing.T) {
		// Create a new mock DTO manager for this test
		mockDTO := new(MockDTOManager)
		mockDTO.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()

		// Create invalid JSON payload
		invalidJSON := []byte(`{"name": "Updated Name", "email": }`)

		// Setup the handler
		r := gin.New()
		r.PUT("/users/:id", GenerateUpdateHandler(mockResource, mockRepo, mockDTO))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/123", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check error message
		assert.Contains(t, response["error"].(string), "invalid character", "Should contain JSON parsing error")

		// Verify mocks
		mockDTO.AssertExpectations(t)
	})

	// Test case 3: Repository error
	t.Run("Repository error", func(t *testing.T) {
		// Create a new mock DTO manager for this test
		mockDTO := new(MockDTOManager)

		// Create test DTO
		updateDTO := &UserUpdateDTO{
			Name:  "Failed Update",
			Email: "failed@example.com",
		}

		// Create model data that will be passed to the repository
		modelData := map[string]interface{}{
			"name":  "Failed Update",
			"email": "failed@example.com",
		}

		// Create JSON payload
		jsonData, _ := json.Marshal(updateDTO)

		// Setup DTO manager expectations
		mockDTO.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()
		mockDTO.On("TransformToModel", mock.AnythingOfType("*handler.UserUpdateDTO")).
			Return(modelData, nil).Once()

		// Setup repository expectations
		mockRepo.On("Query", mock.Anything).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, "456", modelData).
			Return(nil, errors.New("database error")).Once()

		// Setup the handler
		r := gin.New()
		r.PUT("/users/:id", GenerateUpdateHandler(mockResource, mockRepo, mockDTO))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/456", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check error message
		assert.Equal(t, "database error", response["error"])

		// Verify repository mock
		mockRepo.AssertExpectations(t)
		mockDTO.AssertExpectations(t)
	})

	// Test case 4: Record not found
	t.Run("Record not found", func(t *testing.T) {
		// Create a new mock DTO manager for this test
		mockDTO := new(MockDTOManager)

		// Create test DTO
		updateDTO := &UserUpdateDTO{
			Name:  "Invalid Update",
			Email: "invalid@example.com",
		}

		// Create model data that will be passed to the repository
		modelData := map[string]interface{}{
			"name":  "Invalid Update",
			"email": "invalid@example.com",
		}

		// Create JSON payload
		jsonData, _ := json.Marshal(updateDTO)

		// Setup DTO manager expectations
		mockDTO.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()
		mockDTO.On("TransformToModel", mock.AnythingOfType("*handler.UserUpdateDTO")).
			Return(modelData, nil).Once()

		// Setup repository expectations
		mockRepo.On("Query", mock.Anything).Return(nil).Once()
		mockRepo.On("Update", mock.Anything, "999", modelData).
			Return(nil, errors.New("record not found")).Once()

		// Setup the handler
		r := gin.New()
		r.PUT("/users/:id", GenerateUpdateHandler(mockResource, mockRepo, mockDTO))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/999", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check error message
		assert.Equal(t, "Resource not found", response["error"])

		// Verify repository mock
		mockRepo.AssertExpectations(t)
		mockDTO.AssertExpectations(t)
	})
}
