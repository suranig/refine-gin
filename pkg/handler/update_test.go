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
func TestValidateNestedJsonFields_NestedRules(t *testing.T) {
	// Nested model with deeper JSON validation rules
	type Settings struct {
		Preferences struct {
			Notifications struct {
				Email string `json:"email"`
			} `json:"notifications"`
		} `json:"preferences"`
	}

	type TestModel struct {
		ID       int      `json:"id"`
		Settings Settings `json:"settings"`
	}

	mockResource := new(MockResource)
	fields := []resource.Field{
		{
			Name: "Settings",
			Type: "json",
			Json: &resource.JsonConfig{
				Nested: true,
				Properties: []resource.JsonProperty{
					{
						Path:       "preferences",
						Type:       "object",
						Validation: &resource.JsonValidation{Required: true},
						Properties: []resource.JsonProperty{
							{
								Path: "notifications.email",
								Type: "string",
								Validation: &resource.JsonValidation{
									Required: true,
									Pattern:  `^.+@.+\\..+$`,
								},
							},
						},
					},
				},
			},
		},
	}

	mockResource.On("GetFields").Return(fields)
	mockResource.On("GetEditableFields").Return([]string{}).Maybe()

	t.Run("valid data", func(t *testing.T) {
		validModel := TestModel{
			ID: 1,
			Settings: Settings{
				Preferences: struct {
					Notifications struct {
						Email string `json:"email"`
					} `json:"notifications"`
				}{
					Notifications: struct {
						Email string `json:"email"`
					}{Email: "user@example.com"},
				},
			},
		}

		err := validateNestedJsonFields(mockResource, &validModel)
		assert.NoError(t, err)
	})

	t.Run("invalid data", func(t *testing.T) {
		invalidModel := TestModel{
			Settings: Settings{
				Preferences: struct {
					Notifications struct {
						Email string `json:"email"`
					} `json:"notifications"`
				}{
					Notifications: struct {
						Email string `json:"email"`
					}{Email: "invalid"},
				},
			},
		}

		err := validateNestedJsonFields(mockResource, &invalidModel)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match pattern")
	})

	mockResource.AssertExpectations(t)
}
