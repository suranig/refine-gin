package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/suranig/refine-gin/pkg/naming"
)

func TestNamingConventionMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Test cases for different naming conventions
	testCases := []struct {
		name         string
		convention   naming.NamingConvention
		requestBody  map[string]interface{}
		responseBody map[string]interface{}
		expectedReq  map[string]interface{}
		expectedResp map[string]interface{}
	}{
		{
			name:       "Snake Case Conversion",
			convention: naming.SnakeCase,
			requestBody: map[string]interface{}{
				"userId":    123,
				"firstName": "John",
				"lastName":  "Doe",
				"userAddress": map[string]interface{}{
					"streetName":  "Main St",
					"houseNumber": 42,
				},
			},
			responseBody: map[string]interface{}{
				"user_id":    123,
				"first_name": "John",
				"last_name":  "Doe",
				"user_address": map[string]interface{}{
					"street_name":  "Main St",
					"house_number": 42,
				},
			},
			expectedReq: map[string]interface{}{
				"user_id":    123,
				"first_name": "John",
				"last_name":  "Doe",
				"user_address": map[string]interface{}{
					"street_name":  "Main St",
					"house_number": 42,
				},
			},
			expectedResp: map[string]interface{}{
				"user_id":    123,
				"first_name": "John",
				"last_name":  "Doe",
				"user_address": map[string]interface{}{
					"street_name":  "Main St",
					"house_number": 42,
				},
			},
		},
		{
			name:       "Camel Case Conversion",
			convention: naming.CamelCase,
			requestBody: map[string]interface{}{
				"user_id":    123,
				"first_name": "John",
				"last_name":  "Doe",
				"user_address": map[string]interface{}{
					"street_name":  "Main St",
					"house_number": 42,
				},
			},
			responseBody: map[string]interface{}{
				"userId":    123,
				"firstName": "John",
				"lastName":  "Doe",
				"userAddress": map[string]interface{}{
					"streetName":  "Main St",
					"houseNumber": 42,
				},
			},
			expectedReq: map[string]interface{}{
				"userId":    123,
				"firstName": "John",
				"lastName":  "Doe",
				"userAddress": map[string]interface{}{
					"streetName":  "Main St",
					"houseNumber": 42,
				},
			},
			expectedResp: map[string]interface{}{
				"userId":    123,
				"firstName": "John",
				"lastName":  "Doe",
				"userAddress": map[string]interface{}{
					"streetName":  "Main St",
					"houseNumber": 42,
				},
			},
		},
		{
			name:       "Pascal Case Conversion",
			convention: naming.PascalCase,
			requestBody: map[string]interface{}{
				"user_id":    123,
				"first_name": "John",
				"last_name":  "Doe",
				"user_address": map[string]interface{}{
					"street_name": "Main St",
				},
			},
			responseBody: map[string]interface{}{
				"UserId":    123,
				"FirstName": "John",
				"LastName":  "Doe",
				"UserAddress": map[string]interface{}{
					"StreetName": "Main St",
				},
			},
			expectedReq: map[string]interface{}{
				"UserId":    123,
				"FirstName": "John",
				"LastName":  "Doe",
				"UserAddress": map[string]interface{}{
					"StreetName": "Main St",
				},
			},
			expectedResp: map[string]interface{}{
				"UserId":    123,
				"FirstName": "John",
				"LastName":  "Doe",
				"UserAddress": map[string]interface{}{
					"StreetName": "Main St",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create new gin router with the naming middleware
			r := gin.New()
			r.Use(NamingConventionMiddleware(tc.convention))

			// Define a test endpoint that echoes the request body
			r.POST("/request-test", func(c *gin.Context) {
				var data map[string]interface{}
				if err := c.ShouldBindJSON(&data); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, data)
			})

			// Define a test endpoint that returns a predefined response
			r.GET("/response-test", func(c *gin.Context) {
				c.JSON(http.StatusOK, tc.responseBody)
			})

			// Test request body conversion
			reqBodyBytes, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/request-test", bytes.NewBuffer(reqBodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, http.StatusOK, w.Code)

			// Parse response
			var respData map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &respData)
			assert.NoError(t, err)

			// Check that the response (which came from the request) has the expected structure
			assertMapsEquivalent(t, tc.expectedReq, respData)

			// Test response body conversion
			req = httptest.NewRequest(http.MethodGet, "/response-test", nil)
			req.Header.Set("Accept", "application/json")
			w = httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, http.StatusOK, w.Code)

			// Parse response
			err = json.Unmarshal(w.Body.Bytes(), &respData)
			assert.NoError(t, err)

			// Check that the response has been converted to the expected format
			assertMapsEquivalent(t, tc.expectedResp, respData)
		})
	}
}

func TestNamingConventionMiddleware_NonJSONRequest(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a router with the naming middleware
	r := gin.New()
	r.Use(NamingConventionMiddleware(naming.SnakeCase))

	// Define a test endpoint that accepts form data
	r.POST("/form-test", func(c *gin.Context) {
		name := c.PostForm("name")
		c.String(http.StatusOK, "Hello, %s", name)
	})

	// Create a form request
	req := httptest.NewRequest(http.MethodPost, "/form-test", bytes.NewBufferString("name=John"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(w, req)

	// Check that the response is as expected (middleware should not affect it)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello, John", w.Body.String())
}

func TestNamingConventionMiddleware_MalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(NamingConventionMiddleware(naming.SnakeCase))

	// Handler just echoes back the raw body it receives
	r.POST("/malformed", func(c *gin.Context) {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, string(bodyBytes))
	})

	malformed := `{"name": "John", "age": 30,}`
	req := httptest.NewRequest(http.MethodPost, "/malformed", bytes.NewBufferString(malformed))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, malformed, w.Body.String())
}

// Helper function to assert that two maps are equivalent, accounting for JSON type conversions
func assertMapsEquivalent(t *testing.T, expected, actual map[string]interface{}) {
	// Check that all keys in expected are in actual
	for k, expectedVal := range expected {
		assert.Contains(t, actual, k, "Missing key: %s", k)

		actualVal := actual[k]

		// Handle nested maps
		if expectedMap, ok := expectedVal.(map[string]interface{}); ok {
			if actualMap, ok := actualVal.(map[string]interface{}); ok {
				assertMapsEquivalent(t, expectedMap, actualMap)
				continue
			} else {
				t.Errorf("Expected map for key %s, got %T", k, actualVal)
				continue
			}
		}

		// Handle JSON number conversion (int -> float64)
		if expectedInt, ok := expectedVal.(int); ok {
			if actualFloat, ok := actualVal.(float64); ok {
				assert.Equal(t, float64(expectedInt), actualFloat, "Values differ for key: %s", k)
				continue
			}
		}

		// Default comparison
		assert.Equal(t, expectedVal, actualVal, "Values differ for key: %s", k)
	}

	// Check that all keys in actual are in expected (to catch extra keys)
	for k := range actual {
		assert.Contains(t, expected, k, "Unexpected key: %s", k)
	}
}
