package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDTOManager struct {
	mock.Mock
}

func (m *MockDTOManager) GetCreateDTO() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockDTOManager) GetUpdateDTO() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockDTOManager) GetResponseDTO() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockDTOManager) TransformToModel(data interface{}) (interface{}, error) {
	args := m.Called(data)
	return args.Get(0), args.Error(1)
}

func (m *MockDTOManager) TransformFromModel(model interface{}) (interface{}, error) {
	args := m.Called(model)
	return args.Get(0), args.Error(1)
}

// Simple User model and DTO for testing
type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
}

type UserResponseDTO struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestGenerateGetHandlerWithParamAndDTO(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock resource, repository and DTO manager
	mockResource := new(MockResource)
	mockRepo := new(MockRepository)
	mockDTOManager := new(MockDTOManager)

	// Setup common expectations
	mockResource.On("GetName").Return("users").Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()
	mockResource.On("GetModel").Return(User{}).Maybe()

	// Test case 1: Successful get with DTO transformation
	t.Run("Successful get with DTO transformation", func(t *testing.T) {
		// Create test model
		user := User{
			ID:        123,
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
		}

		// Create response DTO
		userDTO := UserResponseDTO{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		}

		// Setup repository expectations
		mockRepo.On("Get", mock.Anything, "123").Return(user, nil).Once()

		// Setup DTO manager expectations
		mockDTOManager.On("TransformFromModel", user).Return(userDTO, nil).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/users/:user_id", GenerateGetHandlerWithParamAndDTO(mockResource, mockRepo, "user_id", mockDTOManager))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/123", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Get the data from the response
		responseData, ok := response["data"].(map[string]interface{})
		assert.True(t, ok, "Response should contain data field")

		// Check response data
		assert.Equal(t, float64(userDTO.ID), responseData["id"])
		assert.Equal(t, userDTO.Name, responseData["name"])
		assert.Equal(t, userDTO.Email, responseData["email"])

		// Verify mock expectations
		mockRepo.AssertExpectations(t)
		mockDTOManager.AssertExpectations(t)
	})

	// Test case 2: Repository returns an error
	t.Run("Repository error", func(t *testing.T) {
		// Create new mocks for this test case
		mockRepo := new(MockRepository)

		// Setup repository expectations
		mockRepo.On("Get", mock.Anything, "456").Return(nil, errors.New("not found")).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/users/:user_id", GenerateGetHandlerWithParamAndDTO(mockResource, mockRepo, "user_id", mockDTOManager))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/456", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check error message
		assert.Equal(t, "Resource not found", response["error"])

		// Verify mock expectations
		mockRepo.AssertExpectations(t)
	})

	// Test case 3: DTO transformation fails
	t.Run("DTO transformation error", func(t *testing.T) {
		// Create new mocks for this test case
		mockRepo := new(MockRepository)
		mockDTOManager := new(MockDTOManager)

		// Create test model
		user := User{
			ID:        789,
			Name:      "Jane Smith",
			Email:     "jane@example.com",
			CreatedAt: time.Now(),
		}

		// Setup repository expectations
		mockRepo.On("Get", mock.Anything, "789").Return(user, nil).Once()

		// Setup DTO manager expectations - transformation fails
		mockDTOManager.On("TransformFromModel", user).Return(nil, errors.New("transformation error")).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/users/:user_id", GenerateGetHandlerWithParamAndDTO(mockResource, mockRepo, "user_id", mockDTOManager))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/789", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check error message
		assert.Equal(t, "Error transforming data: transformation error", response["error"])

		// Verify mock expectations
		mockRepo.AssertExpectations(t)
		mockDTOManager.AssertExpectations(t)
	})
}

func TestGenerateGetHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock resource and repository
	mockResource := new(MockResource)
	mockRepo := new(MockRepository)

	// Setup resource expectations
	mockResource.On("GetName").Return("users").Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()
	mockResource.On("GetModel").Return(User{}).Maybe()

	// Test case: Successful get
	t.Run("Successful get", func(t *testing.T) {
		// Create test model
		user := User{
			ID:        123,
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
		}

		// Setup repository expectations
		mockRepo.On("Get", mock.Anything, "123").Return(user, nil).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/users/:id", GenerateGetHandler(mockResource, mockRepo))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/123", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Get the data from the response
		responseData, ok := response["data"].(map[string]interface{})
		assert.True(t, ok, "Response should contain data field")

		// Check basic response data
		assert.Equal(t, float64(123), responseData["ID"])
		assert.Equal(t, "John Doe", responseData["Name"])
		assert.Equal(t, "john@example.com", responseData["Email"])

		// Verify mock expectations
		mockRepo.AssertExpectations(t)
	})
}
