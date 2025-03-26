package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/resource"
)

func TestGenerateOptionsHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a mock resource
	mockResource := new(MockResource)

	// Setup expectations
	mockResource.On("GetName").Return("test_resource")
	mockResource.On("GetLabel").Return("Test Resource")
	mockResource.On("GetIcon").Return("test-icon")
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int", Required: true},
		{Name: "name", Type: "string", Required: true, Searchable: true},
		{Name: "description", Type: "string", Required: false},
	})
	mockResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationCreate,
		resource.OperationRead,
		resource.OperationList,
	})
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetRelations").Return([]resource.Relation{})
	mockResource.On("GetIDFieldName").Return("ID")
	mockResource.On("GetSearchable").Return([]string{"name"})

	// Register the options handler
	r.OPTIONS("/test_resource", GenerateOptionsHandler(mockResource))

	// Test case 1: First request should return 200 with metadata
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test_resource", nil)
	r.ServeHTTP(w, req)

	// Print response body for debugging
	t.Logf("Response body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"test_resource\"")
	assert.Contains(t, w.Body.String(), "\"label\":\"Test Resource\"")
	assert.Contains(t, w.Body.String(), "\"icon\":\"test-icon\"")

	// Get the ETag from the response
	etag := w.Header().Get("ETag")
	assert.NotEmpty(t, etag, "ETag should be present in the response")

	// Test case 2: Subsequent request with If-None-Match header should return 304
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/test_resource", nil)
	req.Header.Set("If-None-Match", etag)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotModified, w.Code)
	assert.Empty(t, w.Body.String(), "Body should be empty on 304 response")

	// Test case 3: Request with different If-None-Match should return 200
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/test_resource", nil)
	req.Header.Set("If-None-Match", "\"invalid-etag\"")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"test_resource\"")

	// Verify all expectations were met
	mockResource.AssertExpectations(t)
}

func TestRegisterOptionsEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")

	// Create a mock resource
	mockResource := new(MockResource)

	// Setup expectations
	mockResource.On("GetName").Return("test_resource")
	mockResource.On("GetLabel").Return("Test Resource")
	mockResource.On("GetIcon").Return("test-icon")
	mockResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "int", Required: true},
	})
	mockResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationCreate,
		resource.OperationRead,
	})
	mockResource.On("GetDefaultSort").Return(nil)
	mockResource.On("GetFilters").Return([]resource.Filter{})
	mockResource.On("GetRelations").Return([]resource.Relation{})
	mockResource.On("GetIDFieldName").Return("ID")
	mockResource.On("GetSearchable").Return([]string{})

	// Register the options endpoint
	RegisterOptionsEndpoint(apiGroup, mockResource)

	// Test the endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/test_resource", nil)
	r.ServeHTTP(w, req)

	// Print response body for debugging
	t.Logf("Response body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"name\":\"test_resource\"")

	// Verify all expectations were met
	mockResource.AssertExpectations(t)
}
