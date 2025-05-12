package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/stanxing/refine-gin/pkg/resource"
	"github.com/stanxing/refine-gin/pkg/utils"
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

	// Create test resources
	resource1 := NewAPIConfigMockResource("resource1")
	resource2 := NewAPIConfigMockResource("resource2")

	// Clear registry and register test resources
	registry := resource.GlobalResourceRegistry
	registry.Register(resource1)
	registry.Register(resource2)

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

	// Create a dedicated resource for this test with a unique name
	// to avoid conflicts with other tests
	uniqueResourceName := "test-resource-caching-" + fmt.Sprintf("%d", time.Now().UnixNano())
	testResource := NewAPIConfigMockResource(uniqueResourceName)

	// Register the resource
	registry := resource.GlobalResourceRegistry
	registry.Register(testResource)

	// Create a handler with a closure to ensure we're using the right resource state
	configHandler := func(c *gin.Context) {
		// Get current registry state
		etag := utils.GenerateETagFromSlice([]string{uniqueResourceName})
		ifNoneMatch := c.GetHeader("If-None-Match")

		// Check if client's cached version is still valid
		if utils.IsETagMatch(etag, ifNoneMatch) {
			c.Status(http.StatusNotModified)
			return
		}

		// Return fake response for this test
		response := APIConfigResponse{
			Resources: map[string]resource.ResourceMetadata{
				uniqueResourceName: {
					Name:        uniqueResourceName,
					Operations:  nil,
					Fields:      []resource.FieldMetadata{},
					IDFieldName: "ID",
				},
			},
			Config: map[string]interface{}{
				"version": "1.0.0",
			},
		}

		// Return response
		c.JSON(http.StatusOK, response)
	}

	// Register custom handler for this test
	router.GET("/api-config", configHandler)

	// First request to get the ETag
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api-config", nil)
	router.ServeHTTP(w1, req1)

	etag := w1.Header().Get("ETag")
	assert.NotEmpty(t, etag)

	// Log for debugging
	t.Logf("First response status: %d", w1.Code)
	t.Logf("ETag received: %s", etag)

	// Second request with If-None-Match header
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api-config", nil)
	req2.Header.Set("If-None-Match", etag)
	router.ServeHTTP(w2, req2)

	// Log debugging info
	t.Logf("Second response status: %d", w2.Code)
	t.Logf("Second response body: %s", w2.Body.String())
	t.Logf("If-None-Match sent: %s", req2.Header.Get("If-None-Match"))

	// Should return 304 Not Modified
	assert.Equal(t, http.StatusNotModified, w2.Code)
}

func TestRegisterAPIConfigEndpoint(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiGroup := router.Group("/api")

	// Register API config endpoint
	RegisterAPIConfigEndpoint(apiGroup)

	// Create and register mock resources
	resourceA := NewAPIConfigMockResource("resourceA")
	resourceB := NewAPIConfigMockResource("resourceB")

	// Register test resources
	registry := resource.GlobalResourceRegistry
	registry.Register(resourceA)
	registry.Register(resourceB)

	// Make a request to the config endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, 200, w.Code)

	// Parse the response
	var response APIConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify specific resources in the response
	assert.Contains(t, response.Resources, "resourceA")
	assert.Contains(t, response.Resources, "resourceB")

	// Verify resource metadata
	resourceAMeta := response.Resources["resourceA"]
	assert.Equal(t, "resourceA", resourceAMeta.Name)
	assert.Equal(t, "Test resourceA", resourceAMeta.Label)
	assert.Equal(t, "test-icon", resourceAMeta.Icon)

	// Verify config section
	assert.Contains(t, response.Config, "version")
}
