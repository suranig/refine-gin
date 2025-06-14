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
	"github.com/suranig/refine-gin/pkg/repository"
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

	t.Run("Invalid JSON", func(t *testing.T) {
		// Setup
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		// Setup request with invalid JSON
		reqBody := `{"name":"Invalid JSON"`
		req := httptest.NewRequest(http.MethodPost, "/tests", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		// Call handler
		handler := GenerateOwnerCreateHandler(res, mockRepo, mockDTO)
		handler(c)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "unexpected EOF")
	})

	t.Run("DTO transformation error", func(t *testing.T) {
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

		// Mock DTO transformation error
		mockDTO.On("TransformFromModel", createdItem).Return(nil, errors.New("transformation error"))

		// Call handler
		handler := GenerateOwnerCreateHandler(res, mockRepo, mockDTO)
		handler(c)

		// Assertions
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "transformation error", response["error"])
	})
}

func TestOwnerGetHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		req := httptest.NewRequest(http.MethodGet, "/tests/1", nil)
		c.Request = req

		item := OwnerTestModel{ID: 1, Name: "Test", OwnerID: "test-owner"}
		mockRepo.On("Get", mock.Anything, "1").Return(item, nil)
		dtoItem := map[string]interface{}{"id": float64(1), "name": "Test", "ownerId": "test-owner"}
		mockDTO.On("TransformFromModel", item).Return(dtoItem, nil)

		GenerateOwnerGetHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, "Test", data["name"])
	})

	t.Run("Not found", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "2"}}
		req := httptest.NewRequest(http.MethodGet, "/tests/2", nil)
		c.Request = req

		mockRepo.On("Get", mock.Anything, "2").Return(nil, errors.New("record not found"))

		GenerateOwnerGetHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Owner mismatch", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "3"}}
		req := httptest.NewRequest(http.MethodGet, "/tests/3", nil)
		c.Request = req

		mockRepo.On("Get", mock.Anything, "3").Return(nil, repository.ErrOwnerMismatch)

		GenerateOwnerGetHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Missing ID", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		req := httptest.NewRequest(http.MethodGet, "/tests", nil)
		c.Request = req

		GenerateOwnerGetHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DTO transformation error", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		req := httptest.NewRequest(http.MethodGet, "/tests/1", nil)
		c.Request = req

		item := OwnerTestModel{ID: 1, Name: "Test", OwnerID: "test-owner"}
		mockRepo.On("Get", mock.Anything, "1").Return(item, nil)
		mockDTO.On("TransformFromModel", item).Return(nil, errors.New("transformation error"))

		GenerateOwnerGetHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "transformation error", resp["error"])
	})
}

func TestOwnerUpdateHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		reqBody := `{"name":"Updated"}`
		req := httptest.NewRequest(http.MethodPut, "/tests/1", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		updated := OwnerTestModel{ID: 1, Name: "Updated", OwnerID: "test-owner"}
		mockRepo.On("Update", mock.Anything, "1", mock.Anything).Return(updated, nil)
		dtoItem := map[string]interface{}{"id": float64(1), "name": "Updated", "ownerId": "test-owner"}
		mockDTO.On("TransformFromModel", updated).Return(dtoItem, nil)

		GenerateOwnerUpdateHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, "Updated", data["name"])
	})

	t.Run("Owner mismatch", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "2"}}
		reqBody := `{"name":"Other"}`
		req := httptest.NewRequest(http.MethodPut, "/tests/2", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		mockRepo.On("Update", mock.Anything, "2", mock.Anything).Return(nil, repository.ErrOwnerMismatch)

		GenerateOwnerUpdateHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Not found", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "3"}}
		reqBody := `{"name":"Other"}`
		req := httptest.NewRequest(http.MethodPut, "/tests/3", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		mockRepo.On("Update", mock.Anything, "3", mock.Anything).Return(nil, errors.New("record not found"))

		GenerateOwnerUpdateHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Missing ID", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		reqBody := `{"name":"Test"}`
		req := httptest.NewRequest(http.MethodPut, "/tests", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		GenerateOwnerUpdateHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		reqBody := `{"name":"Invalid JSON"`
		req := httptest.NewRequest(http.MethodPut, "/tests/1", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		GenerateOwnerUpdateHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DTO transformation error", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		reqBody := `{"name":"Updated"}`
		req := httptest.NewRequest(http.MethodPut, "/tests/1", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		updated := OwnerTestModel{ID: 1, Name: "Updated", OwnerID: "test-owner"}
		mockRepo.On("Update", mock.Anything, "1", mock.Anything).Return(updated, nil)
		mockDTO.On("TransformFromModel", updated).Return(nil, errors.New("transformation error"))

		GenerateOwnerUpdateHandler(res, mockRepo, mockDTO, "id")(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

}

func TestOwnerDeleteHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c, w, mockRepo, _, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
		req := httptest.NewRequest(http.MethodDelete, "/tests/1", nil)
		c.Request = req

		mockRepo.On("Delete", mock.Anything, "1").Return(nil)

		GenerateOwnerDeleteHandler(res, mockRepo, "id")(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, true, data["success"])
	})

	t.Run("Owner mismatch", func(t *testing.T) {
		c, w, mockRepo, _, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "2"}}
		req := httptest.NewRequest(http.MethodDelete, "/tests/2", nil)
		c.Request = req

		mockRepo.On("Delete", mock.Anything, "2").Return(repository.ErrOwnerMismatch)

		GenerateOwnerDeleteHandler(res, mockRepo, "id")(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Not found", func(t *testing.T) {
		c, w, mockRepo, _, res := setupOwnerHandlerTest(t)

		c.Params = gin.Params{gin.Param{Key: "id", Value: "3"}}
		req := httptest.NewRequest(http.MethodDelete, "/tests/3", nil)
		c.Request = req

		mockRepo.On("Delete", mock.Anything, "3").Return(errors.New("record not found"))

		GenerateOwnerDeleteHandler(res, mockRepo, "id")(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Missing ID", func(t *testing.T) {
		c, w, mockRepo, _, res := setupOwnerHandlerTest(t)

		req := httptest.NewRequest(http.MethodDelete, "/tests", nil)
		c.Request = req

		GenerateOwnerDeleteHandler(res, mockRepo, "id")(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestOwnerCreateManyHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		reqBody := `[{"name":"Test 1"},{"name":"Test 2"}]`
		req := httptest.NewRequest(http.MethodPost, "/tests/batch", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		created := []OwnerTestModel{
			{ID: 1, Name: "Test 1", OwnerID: "test-owner"},
			{ID: 2, Name: "Test 2", OwnerID: "test-owner"},
		}
		mockRepo.On("CreateMany", mock.Anything, mock.Anything).Return(created, nil)
		dtoData := []map[string]interface{}{
			{"id": float64(1), "name": "Test 1", "ownerId": "test-owner"},
			{"id": float64(2), "name": "Test 2", "ownerId": "test-owner"},
		}
		mockDTO.On("TransformFromModel", created).Return(dtoData, nil)

		GenerateOwnerCreateManyHandler(res, mockRepo, mockDTO)(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.NotNil(t, resp["data"])
	})

	t.Run("Repository error", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		reqBody := `[{"name":"Test 1"}]`
		req := httptest.NewRequest(http.MethodPost, "/tests/batch", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		mockRepo.On("CreateMany", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))

		GenerateOwnerCreateManyHandler(res, mockRepo, mockDTO)(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "database error", resp["error"])
	})
}

func TestOwnerUpdateManyHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		reqBody := `{"ids":[1,2],"data":{"name":"Updated"}}`
		req := httptest.NewRequest(http.MethodPut, "/tests/batch", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		mockRepo.On("UpdateMany", mock.Anything, mock.Anything, mock.Anything).Return(int64(2), nil)

		GenerateOwnerUpdateManyHandler(res, mockRepo, mockDTO)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, float64(2), resp["data"]["affected"])
	})

	t.Run("ErrOwnerMismatch", func(t *testing.T) {
		c, w, mockRepo, mockDTO, res := setupOwnerHandlerTest(t)

		reqBody := `{"ids":[1,2],"data":{"name":"Updated"}}`
		req := httptest.NewRequest(http.MethodPut, "/tests/batch", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		mockRepo.On("UpdateMany", mock.Anything, mock.Anything, mock.Anything).Return(int64(0), repository.ErrOwnerMismatch)

		GenerateOwnerUpdateManyHandler(res, mockRepo, mockDTO)(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestOwnerDeleteManyHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c, w, mockRepo, _, res := setupOwnerHandlerTest(t)

		reqBody := `{"ids":[1,2]}`
		req := httptest.NewRequest(http.MethodDelete, "/tests/batch", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		mockRepo.On("DeleteMany", mock.Anything, mock.Anything).Return(int64(2), nil)

		GenerateOwnerDeleteManyHandler(res, mockRepo)(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, float64(2), resp["data"]["affected"])
	})

	t.Run("Repository error", func(t *testing.T) {
		c, w, mockRepo, _, res := setupOwnerHandlerTest(t)

		reqBody := `{"ids":[1]}`
		req := httptest.NewRequest(http.MethodDelete, "/tests/batch", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		mockRepo.On("DeleteMany", mock.Anything, mock.Anything).Return(int64(0), errors.New("database error"))

		GenerateOwnerDeleteManyHandler(res, mockRepo)(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "database error", resp["error"])
	})
}

func TestOwnerCountHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup
		c, w, mockRepo, _, res := setupOwnerHandlerTest(t)

		// Setup request with query parameters
		req := httptest.NewRequest(http.MethodGet, "/tests/count?filter[name]=test", nil)
		c.Request = req

		// Mock repository response
		mockRepo.On("Count", mock.Anything, mock.Anything).Return(int64(5), nil)

		// Call handler
		handler := GenerateOwnerCountHandler(res, mockRepo)
		handler(c)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(5), response["data"])

		// Check for ETag header
		etag := w.Header().Get("ETag")
		assert.NotEmpty(t, etag)

		// Check for Cache-Control header
		cacheControl := w.Header().Get("Cache-Control")
		assert.NotEmpty(t, cacheControl)
	})

	t.Run("ETag match - returns 304", func(t *testing.T) {
		// Najpierw wykonujemy zapytanie by otrzymać ETag
		gin.SetMode(gin.TestMode)
		router := gin.New()

		// Setup
		mockRepo := new(MockRepository)
		res := resource.NewResource(resource.ResourceConfig{
			Name:  "tests",
			Model: &OwnerTestModel{},
		})

		// Add owner ID to context middleware
		router.Use(func(c *gin.Context) {
			c.Set(middleware.OwnerContextKey, "test-owner")
			c.Next()
		})

		// Register the handler
		router.GET("/tests/count", GenerateOwnerCountHandler(res, mockRepo))

		// Setup repository mock
		mockRepo.On("Count", mock.Anything, mock.Anything).Return(int64(5), nil).Once()

		// First request to get ETag
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/tests/count", nil)
		router.ServeHTTP(w1, req1)

		// Verify first response
		assert.Equal(t, http.StatusOK, w1.Code)
		etag := w1.Header().Get("ETag")
		assert.NotEmpty(t, etag)

		// Second request with ETag
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/tests/count", nil)
		req2.Header.Set("If-None-Match", etag)
		router.ServeHTTP(w2, req2)

		// Verify second response
		assert.Equal(t, http.StatusNotModified, w2.Code)
		assert.Empty(t, w2.Body.String())

		// Verify mocks
		mockRepo.AssertNumberOfCalls(t, "Count", 1) // Should be called only once
	})

	t.Run("Repository error", func(t *testing.T) {
		// Setup
		c, w, mockRepo, _, res := setupOwnerHandlerTest(t)

		// Setup request
		req := httptest.NewRequest(http.MethodGet, "/tests/count", nil)
		c.Request = req

		// Mock repository error
		mockRepo.On("Count", mock.Anything, mock.Anything).Return(int64(0), errors.New("database error"))

		// Call handler
		handler := GenerateOwnerCountHandler(res, mockRepo)
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
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockRepo := new(MockRepository)

	res := resource.NewResource(resource.ResourceConfig{
		Name:  "owners",
		Model: &OwnerTestModel{},
		Operations: []resource.Operation{
			resource.OperationList,
			resource.OperationRead,
			resource.OperationCreate,
			resource.OperationUpdate,
			resource.OperationDelete,
			resource.OperationCount,
			resource.OperationCreateMany,
			resource.OperationUpdateMany,
			resource.OperationDeleteMany,
		},
	})
	ownerRes := resource.NewOwnerResource(res, resource.DefaultOwnerConfig())

	api := r.Group("/api")
	RegisterOwnerResource(api, ownerRes, mockRepo)

	routes := r.Routes()

	expected := map[string]bool{
		"GET /api/owners":          false,
		"GET /api/owners/:id":      false,
		"POST /api/owners":         false,
		"PUT /api/owners/:id":      false,
		"DELETE /api/owners/:id":   false,
		"GET /api/owners/count":    false,
		"POST /api/owners/batch":   false,
		"PUT /api/owners/batch":    false,
		"DELETE /api/owners/batch": false,
	}

	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, ok := expected[key]; ok {
			expected[key] = true
		}
	}

	for k, v := range expected {
		assert.True(t, v, "Route %s not registered", k)
	}
}
