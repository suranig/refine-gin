package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	monkey "github.com/bouk/monkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/resource"
	"gorm.io/gorm"
)

// User model for testing creates
type CreateUser struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// UserCreateDTO represents a DTO for user creation
type UserCreateDTO struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func TestGenerateCreateHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock resource, repository and DTO manager
	mockResource := new(MockResource)
	mockRepo := new(MockRepository)
	mockDTOManager := new(MockDTOManager)

	// Setup resource expectations
	mockResource.On("GetName").Return("users").Maybe()
	mockResource.On("GetModel").Return(CreateUser{}).Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()
	mockResource.On("GetRelations").Return([]resource.Relation{}).Maybe()

	// Test case 1: Successful create
	t.Run("Successful create", func(t *testing.T) {
		// Create test DTO
		createDTO := &UserCreateDTO{
			Name:  "John Doe",
			Email: "john@example.com",
		}

		// Create JSON payload
		jsonData, _ := json.Marshal(createDTO)

		// Create model data for the transformation result
		modelData := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		}

		// Create expected result
		createdUser := CreateUser{
			ID:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
		}

		// Setup DTO manager expectations
		mockDTOManager.On("GetCreateDTO").Return(&UserCreateDTO{}).Once()
		mockDTOManager.On("TransformToModel", mock.AnythingOfType("*handler.UserCreateDTO")).
			Return(modelData, nil).Once()
		mockDTOManager.On("TransformFromModel", createdUser).
			Return(createdUser, nil).Once()

		// Setup repository expectations - important: return nil for the Query method
		mockRepo.On("Query", mock.Anything).Return(nil).Once()
		mockRepo.On("Create", mock.Anything, modelData).
			Return(createdUser, nil).Once()

		// Setup the handler
		r := gin.New()
		r.POST("/users", GenerateCreateHandler(mockResource, mockRepo, mockDTOManager))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusCreated, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check response data - The handler should return the created entity inside a 'data' field
		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok, "Response should contain data field with an object")

		// Field names match the struct field names in CreateUser
		assert.Equal(t, float64(1), data["id"], "ID should match")
		assert.Equal(t, "John Doe", data["name"], "Name should match")
		assert.Equal(t, "john@example.com", data["email"], "Email should match")

		// Verify repository mock
		mockRepo.AssertExpectations(t)
		mockDTOManager.AssertExpectations(t)
	})

	// Test case 2: Invalid JSON payload
	t.Run("Invalid JSON payload", func(t *testing.T) {
		// Create a new mock DTO manager for this test
		mockDTO := new(MockDTOManager)
		mockDTO.On("GetCreateDTO").Return(&UserCreateDTO{}).Once()

		// Create invalid JSON payload
		invalidJSON := []byte(`{"name": "John Doe", "email": }`)

		// Setup the handler
		r := gin.New()
		r.POST("/users", GenerateCreateHandler(mockResource, mockRepo, mockDTO))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(invalidJSON))
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
		createDTO := &UserCreateDTO{
			Name:  "Jane Smith",
			Email: "jane@example.com",
		}

		// Create model data that will be passed to the repository
		modelData := map[string]interface{}{
			"name":  "Jane Smith",
			"email": "jane@example.com",
		}

		// Create JSON payload
		jsonData, _ := json.Marshal(createDTO)

		// Setup DTO manager expectations
		mockDTO.On("GetCreateDTO").Return(&UserCreateDTO{}).Once()
		mockDTO.On("TransformToModel", mock.AnythingOfType("*handler.UserCreateDTO")).
			Return(modelData, nil).Once()

		// Setup repository expectations - important: return nil for the Query method
		mockRepo.On("Query", mock.Anything).Return(nil).Once()
		mockRepo.On("Create", mock.Anything, modelData).
			Return(nil, errors.New("database error")).Once()

		// Setup the handler
		r := gin.New()
		r.POST("/users", GenerateCreateHandler(mockResource, mockRepo, mockDTO))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
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

	// Test case 4: Relation validation error
	t.Run("Relation validation error", func(t *testing.T) {
		// Setup local mocks
		mockRepo := new(MockRepository)
		mockRes := new(MockResource)
		mockDTO := new(MockDTOManager)

		// Configure resource expectations
		mockRes.On("GetName").Return("users").Maybe()
		mockRes.On("GetModel").Return(CreateUser{}).Maybe()
		mockRes.On("GetIDFieldName").Return("ID").Maybe()
		mockRes.On("GetRelations").Return([]resource.Relation{{Name: "owner", Type: resource.RelationTypeManyToOne, Resource: "owners", Field: "OwnerID"}}).Maybe()

		// Patch ValidateRelations to return an error
		patch := monkey.Patch(resource.ValidateRelations, func(db *gorm.DB, obj interface{}) error {
			return errors.New("Relation validation failed")
		})
		defer patch.Unpatch()

		// Setup DTO expectations
		createDTO := &UserCreateDTO{Name: "Bob", Email: "bob@example.com"}
		jsonData, _ := json.Marshal(createDTO)
		modelData := map[string]interface{}{"name": "Bob", "email": "bob@example.com"}
		mockDTO.On("GetCreateDTO").Return(&UserCreateDTO{}).Once()
		mockDTO.On("TransformToModel", mock.AnythingOfType("*handler.UserCreateDTO")).Return(modelData, nil).Once()

		// Repository Query returns non-nil DB
		mockRepo.On("Query", mock.Anything).Return(&gorm.DB{}).Once()

		// Setup the handler
		r := gin.New()
		r.POST("/users", GenerateCreateHandler(mockRes, mockRepo, mockDTO))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		// Expect bad request with validation error
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Relation validation failed")

		mockRepo.AssertExpectations(t)
		mockDTO.AssertExpectations(t)
	})
}
