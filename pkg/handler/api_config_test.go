package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/suranig/refine-gin/pkg/resource"
)

// MockResource implements the Resource interface for testing API config
type APIConfigMockResource struct {
	resource.DefaultResource
}

func NewAPIConfigMockResource(name string) *APIConfigMockResource {
	return &APIConfigMockResource{
		DefaultResource: resource.DefaultResource{
			Name:  name,
			Label: "Test " + name,
			Icon:  "test-icon",
		},
	}
}

func TestGenerateAPIConfigHandler(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create and register mock resources
	resource1 := NewAPIConfigMockResource("resource1")
	resource2 := NewAPIConfigMockResource("resource2")

	// Clear registry and register test resources
	registry := resource.GetRegistry()
	registry.RegisterResource(resource1)
	registry.RegisterResource(resource2)

	// Register API config endpoint
	router.GET("/api-config", GenerateAPIConfigHandler())

	// Create test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api-config", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 200, w.Code)

	// Parse the response
	var response APIConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify resources in the response
	assert.Len(t, response.Resources, 2)

	// Verify resource names are in the response
	var resourceNames []string
	for name := range response.Resources {
		resourceNames = append(resourceNames, name)
	}
	assert.Contains(t, resourceNames, "resource1")
	assert.Contains(t, resourceNames, "resource2")

	// Verify ETag header is set
	assert.NotEmpty(t, w.Header().Get("ETag"))

	// Verify Cache-Control header is set
	assert.NotEmpty(t, w.Header().Get("Cache-Control"))
}

func TestAPIConfigCaching(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Clear registry and register a test resource
	registry := resource.GetRegistry()
	registry.RegisterResource(NewAPIConfigMockResource("test-resource"))

	// Register API config endpoint
	router.GET("/api-config", GenerateAPIConfigHandler())

	// First request to get the ETag
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api-config", nil)
	router.ServeHTTP(w1, req1)

	etag := w1.Header().Get("ETag")
	assert.NotEmpty(t, etag)

	// Second request with If-None-Match header
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api-config", nil)
	req2.Header.Set("If-None-Match", etag)
	router.ServeHTTP(w2, req2)

	// Should return 304 Not Modified
	assert.Equal(t, 304, w2.Code)
	assert.Empty(t, w2.Body.String())
}
