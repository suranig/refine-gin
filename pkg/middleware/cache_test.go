package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Test cases for Cache middleware
	testCases := []struct {
		name             string
		config           CacheConfig
		method           string
		path             string
		headers          map[string]string
		expectedStatus   int
		expectedHeaders  map[string]string
		shouldHaveETag   bool
		shouldHaveMaxAge bool
	}{
		{
			name: "Basic GET Request",
			config: CacheConfig{
				MaxAge:       60,
				DisableCache: false,
				Methods:      []string{"GET", "HEAD"},
				VaryHeaders:  []string{"Accept", "Accept-Encoding"},
			},
			method:           http.MethodGet,
			path:             "/test",
			expectedStatus:   http.StatusOK,
			shouldHaveETag:   true,
			shouldHaveMaxAge: true,
			expectedHeaders: map[string]string{
				"Vary": "Accept, Accept-Encoding",
			},
		},
		{
			name: "If-None-Match Matching",
			config: CacheConfig{
				MaxAge:       60,
				DisableCache: false,
				Methods:      []string{"GET", "HEAD"},
				VaryHeaders:  []string{"Accept"},
			},
			method: http.MethodGet,
			path:   "/test",
			headers: map[string]string{
				"If-None-Match": utils.GenerateETag("/test?"),
			},
			expectedStatus:   http.StatusNotModified,
			shouldHaveETag:   false,
			shouldHaveMaxAge: false,
		},
		{
			name: "HEAD Request",
			config: CacheConfig{
				MaxAge:       120,
				DisableCache: false,
				Methods:      []string{"GET", "HEAD"},
				VaryHeaders:  []string{"Accept"},
			},
			method:           http.MethodHead,
			path:             "/test",
			expectedStatus:   http.StatusOK,
			shouldHaveETag:   false,
			shouldHaveMaxAge: true,
		},
		{
			name: "POST Request (No Caching)",
			config: CacheConfig{
				MaxAge:       60,
				DisableCache: false,
				Methods:      []string{"GET", "HEAD"},
				VaryHeaders:  []string{"Accept"},
			},
			method:           http.MethodPost,
			path:             "/test",
			expectedStatus:   http.StatusOK,
			shouldHaveETag:   false,
			shouldHaveMaxAge: false,
		},
		{
			name: "Disabled Cache",
			config: CacheConfig{
				MaxAge:       60,
				DisableCache: true,
				Methods:      []string{"GET", "HEAD"},
				VaryHeaders:  []string{"Accept"},
			},
			method:           http.MethodGet,
			path:             "/test",
			expectedStatus:   http.StatusOK,
			shouldHaveETag:   false,
			shouldHaveMaxAge: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new router with the cache middleware
			r := gin.New()
			r.Use(Cache(tc.config))

			// Add a test handler
			r.Any("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "Hello, World!")
			})

			// Create the request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()

			// Serve the request
			r.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Check headers
			for k, v := range tc.expectedHeaders {
				assert.Equal(t, v, w.Header().Get(k))
			}

			// Check ETag header
			if tc.shouldHaveETag {
				assert.NotEmpty(t, w.Header().Get("ETag"), "ETag header should be set")
			}

			// Check Cache-Control header
			if tc.shouldHaveMaxAge {
				cacheControl := w.Header().Get("Cache-Control")
				assert.Contains(t, cacheControl, fmt.Sprintf("max-age=%d", tc.config.MaxAge))
			} else if tc.expectedStatus != http.StatusNotModified {
				// Only check if we're not returning 304, as 304 doesn't need to include headers
				assert.Empty(t, w.Header().Get("Cache-Control"), "Cache-Control header should not be set")
			}
		})
	}
}

func TestCacheByResource(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Test cases for CacheByResource middleware
	testCases := []struct {
		name           string
		config         CacheConfig
		resourceName   string
		method         string
		path           string
		params         gin.Params
		headers        map[string]string
		expectedStatus int
		checkETag      bool
	}{
		{
			name: "Resource List",
			config: CacheConfig{
				MaxAge:       60,
				DisableCache: false,
				Methods:      []string{"GET"},
				VaryHeaders:  []string{"Accept"},
			},
			resourceName:   "users",
			method:         http.MethodGet,
			path:           "/users",
			expectedStatus: http.StatusOK,
			checkETag:      true,
		},
		{
			name: "Single Resource",
			config: CacheConfig{
				MaxAge:       60,
				DisableCache: false,
				Methods:      []string{"GET"},
				VaryHeaders:  []string{"Accept"},
			},
			resourceName:   "users",
			method:         http.MethodGet,
			path:           "/users/123",
			params:         gin.Params{{Key: "id", Value: "123"}},
			expectedStatus: http.StatusOK,
			checkETag:      true,
		},
		{
			name: "If-None-Match for Resource",
			config: CacheConfig{
				MaxAge:       60,
				DisableCache: false,
				Methods:      []string{"GET"},
				VaryHeaders:  []string{"Accept"},
			},
			resourceName: "users",
			method:       http.MethodGet,
			path:         "/users/123",
			params:       gin.Params{{Key: "id", Value: "123"}},
			headers: map[string]string{
				"If-None-Match": utils.GenerateResourceETag("users", "123"),
			},
			expectedStatus: http.StatusNotModified,
			checkETag:      false,
		},
		{
			name: "If-None-Match for List with Query Params",
			config: CacheConfig{
				MaxAge:       60,
				DisableCache: false,
				Methods:      []string{"GET"},
				VaryHeaders:  []string{"Accept"},
			},
			resourceName: "users",
			method:       http.MethodGet,
			path:         "/users?page=1&per_page=10",
			headers: map[string]string{
				"If-None-Match": utils.GenerateQueryETag("page=1&per_page=10"),
			},
			expectedStatus: http.StatusNotModified,
			checkETag:      false,
		},
		{
			name: "POST Request (No Caching)",
			config: CacheConfig{
				MaxAge:       60,
				DisableCache: false,
				Methods:      []string{"GET"},
				VaryHeaders:  []string{"Accept"},
			},
			resourceName:   "users",
			method:         http.MethodPost,
			path:           "/users",
			expectedStatus: http.StatusOK,
			checkETag:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new router with the cache middleware
			r := gin.New()

			// Setup a route with the middleware
			r.Any("/users/*path", func(c *gin.Context) {
				// Set the route params (in a real scenario, Gin would do this)
				for _, param := range tc.params {
					c.Params = append(c.Params, param)
				}

				// Apply middleware manually since we need to set params first
				handler := CacheByResource(tc.resourceName, tc.config)
				handler(c)

				if !c.IsAborted() {
					c.String(http.StatusOK, "Resource response")
				}
			})

			r.Any("/users", func(c *gin.Context) {
				// Apply middleware manually
				handler := CacheByResource(tc.resourceName, tc.config)
				handler(c)

				if !c.IsAborted() {
					c.String(http.StatusOK, "List response")
				}
			})

			// Create the request
			req := httptest.NewRequest(tc.method, tc.path, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()

			// Serve the request
			r.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// Check ETag header if needed
			if tc.checkETag {
				assert.NotEmpty(t, w.Header().Get("ETag"), "ETag header should be set")
			}
		})
	}
}

func TestNoCacheMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router with the no-cache middleware
	r := gin.New()
	r.Use(NoCacheMiddleware())

	// Add a test handler
	r.GET("/no-cache", func(c *gin.Context) {
		c.String(http.StatusOK, "No cache")
	})

	// Create a request
	req := httptest.NewRequest(http.MethodGet, "/no-cache", nil)
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check Cache-Control header
	assert.Equal(t, "no-store, no-cache, must-revalidate, max-age=0", w.Header().Get("Cache-Control"))

	// Check Pragma header
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))

	// Check Expires header
	assert.Equal(t, "0", w.Header().Get("Expires"))
}
