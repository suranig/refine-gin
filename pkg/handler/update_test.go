package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	monkey "github.com/bouk/monkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
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

// Note: MockResource is defined in handler_test.go

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
	mockResource.On("GetEditableFields").Return([]string{"name", "email"}).Maybe()
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "updated_at", Type: "time.Time"},
	}).Maybe()

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

	// Test case 5: Relation validation error
	t.Run("Relation validation error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockRes := new(MockResource)
		mockDTO := new(MockDTOManager)

		// Configure resource expectations
		mockRes.On("GetName").Return("users").Maybe()
		mockRes.On("GetModel").Return(UpdateUser{}).Maybe()
		mockRes.On("GetIDFieldName").Return("ID").Maybe()
		mockRes.On("GetEditableFields").Return([]string{"name", "email"}).Maybe()
		mockRes.On("GetFields").Return([]resource.Field{
			{Name: "id", Type: "int"},
			{Name: "name", Type: "string"},
			{Name: "email", Type: "string"},
		}).Maybe()
		mockRes.On("GetRelations").Return([]resource.Relation{{Name: "owner", Type: resource.RelationTypeManyToOne, Resource: "owners", Field: "OwnerID"}}).Maybe()

		// Patch ValidateRelations to return an error
		patch := monkey.Patch(resource.ValidateRelations, func(db *gorm.DB, obj interface{}) error {
			return errors.New("invalid relation")
		})
		defer patch.Unpatch()

		// DTO expectations
		updateDTO := &UserUpdateDTO{Name: "New", Email: "new@example.com"}
		jsonData, _ := json.Marshal(updateDTO)
		modelData := map[string]interface{}{"name": "New", "email": "new@example.com"}
		mockDTO.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()
		mockDTO.On("TransformToModel", mock.AnythingOfType("*handler.UserUpdateDTO")).Return(modelData, nil).Once()

		// Repository expectations
		mockRepo.On("Query", mock.Anything).Return(&gorm.DB{}).Once()

		// Setup the handler
		r := gin.New()
		r.PUT("/users/:id", GenerateUpdateHandler(mockRes, mockRepo, mockDTO))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Expect bad request with validation error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "invalid relation")

		mockRepo.AssertExpectations(t)
		mockDTO.AssertExpectations(t)
	})

	// Test case 6: JSON validation error
	t.Run("JSON validation error", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockDTO := new(MockDTOManager)

		// Patch validateNestedJsonFields to return an error
		patch := monkey.Patch(validateNestedJsonFields, func(resource.Resource, interface{}) error {
			return errors.New("bad json")
		})
		defer patch.Unpatch()

		// Create valid DTO and JSON payload
		updateDTO := &UserUpdateDTO{Name: "Valid", Email: "valid@example.com"}
		jsonData, _ := json.Marshal(updateDTO)
		modelData := map[string]interface{}{"name": "Valid", "email": "valid@example.com"}

		// DTO manager expectations
		mockDTO.On("GetUpdateDTO").Return(&UserUpdateDTO{}).Once()
		mockDTO.On("TransformToModel", mock.AnythingOfType("*handler.UserUpdateDTO")).Return(modelData, nil).Once()

		// Setup the handler
		r := gin.New()
		r.PUT("/users/:id", GenerateUpdateHandler(mockResource, mockRepo, mockDTO))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Expect bad request with JSON validation error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "JSON validation failed: bad json", response["error"])

		mockDTO.AssertExpectations(t)
	})
}

func TestValidateNestedJsonFields(t *testing.T) {
	// Sample model with nested JSON
	type Config struct {
		Email struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
		} `json:"email"`
		Active bool `json:"active"`
	}

	type TestModel struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Config Config `json:"config"`
	}

	// Create a mock resource with JSON configuration
	mockResource := new(MockResource)

	// Set up fields with JSON config
	fields := []resource.Field{
		{
			Name: "Config",
			Type: "json",
			Json: &resource.JsonConfig{
				Nested: true,
				Properties: []resource.JsonProperty{
					{
						Path: "email.host",
						Type: "string",
						Validation: &resource.JsonValidation{
							Required:  true,
							MinLength: 3,
						},
					},
					{
						Path: "email.port",
						Type: "number",
						Validation: &resource.JsonValidation{
							Required: true,
							Min:      1,
							Max:      65535,
						},
					},
				},
			},
		},
	}

	mockResource.On("GetFields").Return(fields)
	mockResource.On("GetEditableFields").Return([]string{}).Maybe()

	// Test 1: Valid model
	validModel := TestModel{
		ID:   1,
		Name: "Test",
		Config: Config{
			Email: struct {
				Host     string `json:"host"`
				Port     int    `json:"port"`
				Username string `json:"username"`
			}{
				Host: "smtp.example.com",
				Port: 587,
			},
			Active: true,
		},
	}

	err := validateNestedJsonFields(mockResource, &validModel)
	assert.NoError(t, err)

	// Test 2: Invalid model - host too short
	invalidModel1 := TestModel{
		Config: Config{
			Email: struct {
				Host     string `json:"host"`
				Port     int    `json:"port"`
				Username string `json:"username"`
			}{
				Host: "a", // Too short
				Port: 587,
			},
		},
	}

	err = validateNestedJsonFields(mockResource, &invalidModel1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "less than minimum length")

	// Test 3: Invalid model - port out of range
	invalidModel2 := TestModel{
		Config: Config{
			Email: struct {
				Host     string `json:"host"`
				Port     int    `json:"port"`
				Username string `json:"username"`
			}{
				Host: "smtp.example.com",
				Port: 70000, // Out of range
			},
		},
	}

	err = validateNestedJsonFields(mockResource, &invalidModel2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "greater than maximum")

	// Test 4: Nil model
	err = validateNestedJsonFields(mockResource, nil)
	assert.NoError(t, err)

	// Verify expectations
	mockResource.AssertExpectations(t)
}

// TestUpdateHandlerDirectly tests the UpdateHandler function directly
func TestUpdateHandlerDirectly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock resource and repository
	mockResource := new(MockResource)
	mockRepo := new(MockRepository)

	// Define model
	type TestUpdateModel struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	mockResource.On("GetModel").Return(TestUpdateModel{}).Maybe()
	mockResource.On("GetName").Return("test-update").Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()

	// Test case: Successful update with data in Refine format
	t.Run("Successful update with data in Refine format", func(t *testing.T) {
		// Setup gin context with ID param
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		// Setup request with JSON body
		reqBody := `{"data":{"name":"Updated Item"}}`
		c.Request, _ = http.NewRequest(http.MethodPut, "/test-update/123", strings.NewReader(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// Mock repository response
		updatedItem := &TestUpdateModel{ID: 123, Name: "Updated Item"}
		mockRepo.On("Update", mock.Anything, "123", mock.Anything).Return(updatedItem, nil).Once()

		// Call handler
		UpdateHandler(c, mockResource, mockRepo)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.Contains(t, response, "data")
		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(123), data["id"])
		assert.Equal(t, "Updated Item", data["name"])

		// Verify mock calls
		mockRepo.AssertExpectations(t)
	})

	// Test case: Successful update with direct data format
	t.Run("Successful update with direct data format", func(t *testing.T) {
		// Setup gin context with ID param
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "456"}}

		// Setup request with JSON body (direct format, not wrapped in data field)
		reqBody := `{"name":"Direct Update"}`
		c.Request, _ = http.NewRequest(http.MethodPut, "/test-update/456", strings.NewReader(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// Mock repository response
		updatedItem := &TestUpdateModel{ID: 456, Name: "Direct Update"}
		mockRepo.On("Update", mock.Anything, "456", mock.Anything).Return(updatedItem, nil).Once()

		// Call handler
		UpdateHandler(c, mockResource, mockRepo)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.Contains(t, response, "data")
		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(456), data["id"])
		assert.Equal(t, "Direct Update", data["name"])

		// Verify mock calls
		mockRepo.AssertExpectations(t)
	})

	// Test case: Record not found
	t.Run("Record not found", func(t *testing.T) {
		// Setup gin context with ID param
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}

		// Setup request with JSON body
		reqBody := `{"data":{"name":"Not Found"}}`
		c.Request, _ = http.NewRequest(http.MethodPut, "/test-update/999", strings.NewReader(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// Mock repository response - not found error
		mockRepo.On("Update", mock.Anything, "999", mock.Anything).Return(nil, errors.New("record not found")).Once()

		// Call handler
		UpdateHandler(c, mockResource, mockRepo)

		// Assertions
		assert.Equal(t, http.StatusNotFound, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Contains(t, response, "error")
		assert.Equal(t, "Resource not found", response["error"])

		// Verify mock calls
		mockRepo.AssertExpectations(t)
	})

	// Test case: Invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		// Setup gin context with ID param
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "123"}}

		// Setup request with invalid JSON body
		reqBody := `{"data":{"name":"Invalid JSON}`
		c.Request, _ = http.NewRequest(http.MethodPut, "/test-update/123", strings.NewReader(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// Call handler - repository mock should not be called
		UpdateHandler(c, mockResource, mockRepo)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Contains(t, response, "error")
	})

	// Test case: Other error
	t.Run("Other repository error", func(t *testing.T) {
		// Setup gin context with ID param
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "789"}}

		// Setup request with JSON body
		reqBody := `{"data":{"name":"Error Item"}}`
		c.Request, _ = http.NewRequest(http.MethodPut, "/test-update/789", strings.NewReader(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")

		// Mock repository response - other error
		mockRepo.On("Update", mock.Anything, "789", mock.Anything).Return(nil, errors.New("database error")).Once()

		// Call handler
		UpdateHandler(c, mockResource, mockRepo)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check error message
		assert.Contains(t, response, "error")
		assert.Equal(t, "database error", response["error"])

		// Verify mock calls
		mockRepo.AssertExpectations(t)
	})
}
