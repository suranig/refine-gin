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
		{Name: "id", Type: "int"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
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
		{Name: "id", Type: "int"},
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

func TestOptionsHandlerWithMultipleResources(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiGroup := r.Group("/api")

	// Create mock resources
	userResource := new(MockResource)
	postResource := new(MockResource)

	// Setup expectations for user resource
	userResource.On("GetName").Return("users")
	userResource.On("GetLabel").Return("Users")
	userResource.On("GetIcon").Return("user")
	userResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	})
	userResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationCreate,
		resource.OperationRead,
		resource.OperationUpdate,
		resource.OperationDelete,
		resource.OperationList,
	})
	userResource.On("GetDefaultSort").Return(nil)
	userResource.On("GetFilters").Return([]resource.Filter{})
	userResource.On("GetRelations").Return([]resource.Relation{})
	userResource.On("GetIDFieldName").Return("ID")
	userResource.On("GetSearchable").Return([]string{"name"})

	// Setup expectations for post resource
	postResource.On("GetName").Return("posts")
	postResource.On("GetLabel").Return("Blog Posts")
	postResource.On("GetIcon").Return("post")
	postResource.On("GetFields").Return([]resource.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "content", Type: "string"},
	})
	postResource.On("GetOperations").Return([]resource.Operation{
		resource.OperationCreate,
		resource.OperationRead,
		resource.OperationList,
	})
	postResource.On("GetDefaultSort").Return(&resource.Sort{Field: "id", Order: "desc"})
	postResource.On("GetFilters").Return([]resource.Filter{})
	postResource.On("GetRelations").Return([]resource.Relation{})
	postResource.On("GetIDFieldName").Return("ID")
	postResource.On("GetSearchable").Return([]string{"title", "content"})

	// Register options endpoints for both resources
	RegisterOptionsEndpoint(apiGroup, userResource)
	RegisterOptionsEndpoint(apiGroup, postResource)

	// Test OPTIONS for users resource
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("OPTIONS", "/api/users", nil)
	r.ServeHTTP(w1, req1)

	// Verify user resource response
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "\"name\":\"users\"")
	assert.Contains(t, w1.Body.String(), "\"label\":\"Users\"")
	assert.Contains(t, w1.Body.String(), "\"icon\":\"user\"")
	assert.Contains(t, w1.Body.String(), "\"searchable\":[\"name\"]")

	// Test OPTIONS for posts resource
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("OPTIONS", "/api/posts", nil)
	r.ServeHTTP(w2, req2)

	// Verify post resource response
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "\"name\":\"posts\"")
	assert.Contains(t, w2.Body.String(), "\"label\":\"Blog Posts\"")
	assert.Contains(t, w2.Body.String(), "\"searchable\":[\"title\",\"content\"]")

	// Verify all expectations were met
	userResource.AssertExpectations(t)
	postResource.AssertExpectations(t)
}
