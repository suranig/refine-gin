package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/suranig/refine-gin/pkg/middleware"
	"github.com/suranig/refine-gin/pkg/resource"
)

// Owner test model
type OwnerTestModel struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"ownerId" gorm:"column:owner_id"`
}

// Test setup helpers
func setupOwnerHandlerTest(t *testing.T) (*gin.Context, *httptest.ResponseRecorder, *MockRepository, *MockDTOProvider, resource.Resource) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Add owner ID to context
	c.Set(middleware.OwnerContextKey, "test-owner")

	// Create mock repository and DTO provider
	mockRepo := new(MockRepository)
	mockDTO := new(MockDTOProvider)

	// Create resource
	res := resource.NewResource(resource.ResourceConfig{
		Name:  "tests",
		Model: &OwnerTestModel{},
	})

	return c, w, mockRepo, mockDTO, res
}

// Test cases
func TestOwnerListHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		// Setup request with query parameters
		req := httptest.NewRequest(http.MethodGet, "/tests?page=1&pageSize=10", nil)
		c.Request = req

		// Mock repository response
		items := []OwnerTestModel{
			{ID: 1, Name: "Test 1", OwnerID: "test-owner"},
			{ID: 2, Name: "Test 2", OwnerID: "test-owner"},
		}
		mockRepo.On("List", mock.Anything, mock.Anything).Return(items, int64(2), nil)

		// Mock DTO transformations
		dtoItems := []map[string]interface{}{
			{"id": float64(1), "name": "Test 1", "ownerId": "test-owner"},
			{"id": float64(2), "name": "Test 2", "ownerId": "test-owner"},
		}
		for i, item := range items {
			mockDTO.On("TransformFromModel", item).Return(dtoItems[i], nil).Once()
		}

		// Call handler
		handler := GenerateOwnerListHandler(res, mockRepo, mockDTO)
		handler(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(2), response["total"])
		assert.NotNil(t, response["data"])
		assert.Equal(t, float64(1), response["meta"].(map[string]interface{})["page"])
		assert.Equal(t, float64(10), response["meta"].(map[string]interface{})["pageSize"])
	})

	t.Run("Repository error", func(t *testing.T) {
		// Setup
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		// Setup request
		req := httptest.NewRequest(http.MethodGet, "/tests?page=1&pageSize=10", nil)
		c.Request = req

		// Mock repository error
		mockRepo.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("database error"))

		// Call handler
		handler := GenerateOwnerListHandler(res, mockRepo, mockDTO)
		handler(c)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "database error", response["error"])
	})
}

func TestOwnerCreateHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		// Setup request with JSON body
		reqBody := `{"name":"New Test"}`
		req := httptest.NewRequest(http.MethodPost, "/tests", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// Mock repository response
		createdItem := OwnerTestModel{ID: 1, Name: "New Test", OwnerID: "test-owner"}
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(createdItem, nil)

		// Mock DTO transformation
		dtoItem := map[string]interface{}{
			"id": float64(1), "name": "New Test", "ownerId": "test-owner",
		}
		mockDTO.On("TransformFromModel", createdItem).Return(dtoItem, nil)

		// Call handler
		handler := GenerateOwnerCreateHandler(res, mockRepo, mockDTO)
		handler(c)

		// Assertions
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Instead of checking the entire map, check individual fields to avoid type issues
		responseData := response["data"].(map[string]interface{})
		assert.Equal(t, float64(1), responseData["id"])
		assert.Equal(t, "New Test", responseData["name"])
		assert.Equal(t, "test-owner", responseData["ownerId"])
	})

	t.Run("Repository error", func(t *testing.T) {
		// Setup
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		// Setup request with JSON body
		reqBody := `{"name":"New Test"}`
		req := httptest.NewRequest(http.MethodPost, "/tests", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// Mock repository error
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))

		// Call handler
		handler := GenerateOwnerCreateHandler(res, mockRepo, mockDTO)
		handler(c)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "database error", response["error"])
	})
}

func TestOwnerResourceRegistration(t *testing.T) {
	t.Skip("Implementation pending - RegisterOwnerResource needs to be properly tested once completed")
}
