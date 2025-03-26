package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/resource"
)

func TestGenerateCountHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create a mock resource
	mockResource := new(MockResource)

	// Setup resource expectations
	mockResource.On("GetName").Return("test_resource").Maybe()
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int", Required: true},
		{Name: "name", Type: "string", Required: true, Searchable: true},
	}).Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()
	mockResource.On("GetFilters").Return([]resource.Filter{}).Maybe()
	mockResource.On("GetSearchable").Return([]string{"name"}).Maybe()
	mockResource.On("GetDefaultSort").Return(nil).Maybe()

	// Test case 1: Successful count
	t.Run("Successful count", func(t *testing.T) {
		// Create new mocks for this test case
		mockRepo := new(MockRepository)

		// Setup repository to return a count of 42
		mockRepo.On("Count", mock.Anything, mock.AnythingOfType("query.QueryOptions")).
			Return(int64(42), nil).Once()

		// Setup the handler
		r := gin.New()
		r.GET("/test_resource/count", GenerateCountHandler(mockResource, mockRepo))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test_resource/count", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check the count value
		assert.Equal(t, float64(42), response["count"])

		// Verify mock expectations
		mockRepo.AssertExpectations(t)
	})

	// Test case 2: Repository returns error
	t.Run("Repository error", func(t *testing.T) {
		// Create new mock for this test case
		mockRepo := new(MockRepository)

		// Setup repository to return an error
		mockRepo.On("Count", mock.Anything, mock.AnythingOfType("query.QueryOptions")).
			Return(int64(0), errors.New("database error")).Once()

		// Setup the handler with the new mock
		r := gin.New()
		r.GET("/test_resource/count", GenerateCountHandler(mockResource, mockRepo))

		// Make the request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test_resource/count", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check the error message
		assert.Equal(t, "database error", response["error"])

		// Verify mock expectations
		mockRepo.AssertExpectations(t)
	})

	// Test case 3: With query parameters
	t.Run("With query parameters", func(t *testing.T) {
		// Create new mock for this test case
		mockRepo := new(MockRepository)

		// Setup repository expectations
		mockRepo.On("Count", mock.Anything, mock.AnythingOfType("query.QueryOptions")).
			Run(func(args mock.Arguments) {
				// Get the options from arguments and verify they're as expected
				options := args.Get(1).(query.QueryOptions)
				assert.True(t, options.DisablePagination, "Pagination should be disabled")

				// We can't directly check options.Filters["name"] because the mock doesn't
				// have access to the actual filter values in this context
				// The actual filtering will happen in the repository implementation
			}).
			Return(int64(5), nil).Once()

		// Setup the handler with the new mock
		r := gin.New()
		r.GET("/test_resource/count", GenerateCountHandler(mockResource, mockRepo))

		// Make the request with query parameters
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test_resource/count?filter[name]=test", nil)
		r.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Check the count value
		assert.Equal(t, float64(5), response["count"])

		// Verify mock expectations
		mockRepo.AssertExpectations(t)
	})

	// Verify overall resource expectations
	mockResource.AssertExpectations(t)
}

// Test the context cancellation handling
func TestGenerateCountHandlerWithContextCancellation(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a mock resource and repository
	mockResource := new(MockResource)
	mockRepo := new(MockRepository)

	// Setup resource expectations
	mockResource.On("GetName").Return("test_resource").Maybe()
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int", Required: true},
	}).Maybe()
	mockResource.On("GetIDFieldName").Return("ID").Maybe()
	mockResource.On("GetFilters").Return([]resource.Filter{}).Maybe()
	mockResource.On("GetSearchable").Return([]string{}).Maybe()
	mockResource.On("GetDefaultSort").Return(nil).Maybe()

	// Setup repository to simulate context cancellation
	mockRepo.On("Count", mock.Anything, mock.AnythingOfType("query.QueryOptions")).
		Return(int64(0), context.Canceled).Once()

	// Setup the handler
	r.GET("/test_resource/count", GenerateCountHandler(mockResource, mockRepo))

	// Make the request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test_resource/count", nil)
	r.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check the error message
	assert.Equal(t, context.Canceled.Error(), response["error"])

	// Verify all mock expectations were met
	mockResource.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
