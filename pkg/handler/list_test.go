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
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
)

// TestItem represents a simple model for testing
type TestItem struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

// TestItemDTO represents a DTO for TestItem
type TestItemDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestGenerateListHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock resource and repository
	mockResource := new(MockResource)
	mockRepo := new(MockRepository)

	// Setup common resource expectations
	mockResource.On("GetName").Return("items").Maybe()
	mockResource.On("GetModel").Return(TestItem{}).Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()
	mockResource.On("GetSearchable").Return([]string{"Name"}).Maybe()
	mockResource.On("GetDefaultSort").Return(nil).Maybe()
	mockResource.On("GetFilters").Return([]resource.Filter{}).Maybe()
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "ID", Type: "int", Required: true},
		{Name: "Name", Type: "string", Required: true, Searchable: true},
	}).Maybe()

	// Test case 1: Successful list with no query parameters
	t.Run("Successful list with no parameters", func(t *testing.T) {
		// Create test data
		items := []TestItem{
			{ID: 1, Name: "Item 1", CreatedAt: time.Now()},
			{ID: 2, Name: "Item 2", CreatedAt: time.Now()},
		}

		// Setup repository expectations
		mockRepo.On("List", mock.Anything, mock.AnythingOfType("query.QueryOptions")).
			Run(func(args mock.Arguments) {
				// Verify query options
				options := args.Get(1).(query.QueryOptions)
				assert.False(t, options.DisablePagination, "Pagination should be enabled")
				assert.Equal(t, 1, options.Page)
				assert.Equal(t, 10, options.PerPage) // Default page size
			}).
			Return(items, 2, nil).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/items", GenerateListHandler(mockResource, mockRepo))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/items", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check response structure
		data, ok := response["data"].([]interface{})
		assert.True(t, ok, "Response should contain data array")
		assert.Equal(t, 2, len(data), "Should return 2 items")

		meta, ok := response["meta"].(map[string]interface{})
		assert.True(t, ok, "Response should contain meta object")
		assert.Equal(t, float64(2), meta["total"], "Total should be 2")
		assert.Equal(t, float64(1), meta["page"], "Page should be 1")
		assert.Equal(t, float64(10), meta["pageSize"], "PageSize should be 10")

		// Verify repository mock
		mockRepo.AssertExpectations(t)
	})

	// Test case 2: Successful list with pagination, sort and filter
	t.Run("Successful list with query parameters", func(t *testing.T) {
		// Create test data
		items := []TestItem{
			{ID: 3, Name: "Item 3", CreatedAt: time.Now()},
		}

		// Setup repository expectations
		mockRepo.On("List", mock.Anything, mock.AnythingOfType("query.QueryOptions")).
			Run(func(args mock.Arguments) {
				// Verify query options
				options := args.Get(1).(query.QueryOptions)
				assert.False(t, options.DisablePagination, "Pagination should be enabled")
				assert.Equal(t, 2, options.Page)
				assert.Equal(t, 5, options.PerPage)

				// Check sorting
				assert.Equal(t, "Name", options.Sort)
				assert.Equal(t, "asc", options.Order)

				// Check filters
				assert.Contains(t, options.Filters, "ID")
				assert.Equal(t, "3", options.Filters["ID"])

				// Or check advanced filters if your implementation uses them
				found := false
				for _, filter := range options.AdvancedFilters {
					if filter.Field == "ID" && filter.Operator == "eq" && filter.Value == "3" {
						found = true
						break
					}
				}
				assert.True(t, found, "Should have filter ID=3")
			}).
			Return(items, 1, nil).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/items", GenerateListHandler(mockResource, mockRepo))

		// Make the request with query parameters
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/items?page=2&per_page=5&sort=Name&order=asc&filter[ID]=3", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check response structure
		data, ok := response["data"].([]interface{})
		assert.True(t, ok, "Response should contain data array")
		assert.Equal(t, 1, len(data), "Should return 1 item")

		meta, ok := response["meta"].(map[string]interface{})
		assert.True(t, ok, "Response should contain meta object")
		assert.Equal(t, float64(1), meta["total"], "Total should be 1")
		assert.Equal(t, float64(2), meta["page"], "Page should be 2")
		assert.Equal(t, float64(5), meta["pageSize"], "PageSize should be 5")

		// Verify repository mock
		mockRepo.AssertExpectations(t)
	})

	// Test case 3: Repository error
	t.Run("Repository error", func(t *testing.T) {
		// Setup repository expectations - return error
		mockRepo.On("List", mock.Anything, mock.AnythingOfType("query.QueryOptions")).
			Return(nil, 0, errors.New("database error")).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/items", GenerateListHandler(mockResource, mockRepo))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/items", nil)
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
	})
}

func TestGenerateListHandlerWithDTO(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock resource, repository and DTO manager
	mockResource := new(MockResource)
	mockRepo := new(MockRepository)
	mockDTOManager := new(MockDTOManager)

	// Setup common resource expectations
	mockResource.On("GetName").Return("items").Maybe()
	mockResource.On("GetModel").Return(TestItem{}).Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()
	mockResource.On("GetSearchable").Return([]string{"Name"}).Maybe()
	mockResource.On("GetDefaultSort").Return(nil).Maybe()
	mockResource.On("GetFilters").Return([]resource.Filter{}).Maybe()
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "ID", Type: "int", Required: true},
		{Name: "Name", Type: "string", Required: true, Searchable: true},
	}).Maybe()

	// Test case: Successful list with DTO transformation
	t.Run("Successful list with DTO transformation", func(t *testing.T) {
		// Create test data
		items := []TestItem{
			{ID: 1, Name: "Item 1", CreatedAt: time.Now()},
			{ID: 2, Name: "Item 2", CreatedAt: time.Now()},
		}

		// Create response DTOs
		itemDTOs := []TestItemDTO{
			{ID: 1, Name: "Item 1"},
			{ID: 2, Name: "Item 2"},
		}

		// Setup repository expectations
		mockRepo.On("List", mock.Anything, mock.AnythingOfType("query.QueryOptions")).
			Return(items, 2, nil).Once()

		// Setup DTO manager expectations
		// DTO manager should transform the entire array
		mockDTOManager.On("TransformFromModel", items).Return(itemDTOs, nil).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/items", GenerateListHandlerWithDTO(mockResource, mockRepo, mockDTOManager))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/items", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check response structure
		data, ok := response["data"].([]interface{})
		assert.True(t, ok, "Response should contain data array")
		assert.Equal(t, 2, len(data), "Should return 2 items")

		// Check first item
		item0 := data[0].(map[string]interface{})
		assert.Equal(t, float64(1), item0["id"])
		assert.Equal(t, "Item 1", item0["name"])

		// Check second item
		item1 := data[1].(map[string]interface{})
		assert.Equal(t, float64(2), item1["id"])
		assert.Equal(t, "Item 2", item1["name"])

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockDTOManager.AssertExpectations(t)
	})

	// Test case: DTO transformation error
	t.Run("DTO transformation error", func(t *testing.T) {
		// Create test data
		items := []TestItem{
			{ID: 1, Name: "Item 1", CreatedAt: time.Now()},
			{ID: 2, Name: "Item 2", CreatedAt: time.Now()},
		}

		// Setup repository expectations
		mockRepo.On("List", mock.Anything, mock.AnythingOfType("query.QueryOptions")).
			Return(items, 2, nil).Once()

		// Setup DTO manager expectations - transformation fails
		mockDTOManager.On("TransformFromModel", items).
			Return(nil, errors.New("transformation error")).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/items", GenerateListHandlerWithDTO(mockResource, mockRepo, mockDTOManager))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/items", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check error message
		assert.Equal(t, "Error transforming data: transformation error", response["error"])

		// Verify mocks
		mockRepo.AssertExpectations(t)
		mockDTOManager.AssertExpectations(t)
	})
}
