package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestOwnerContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		extractor       ExtractOwnerIDFunc
		expectedStatus  int
		checkOwnerID    bool
		expectedOwnerID interface{}
	}{
		{
			name: "Success",
			extractor: func(c *gin.Context) (interface{}, error) {
				return "user-123", nil
			},
			expectedStatus:  http.StatusOK,
			checkOwnerID:    true,
			expectedOwnerID: "user-123",
		},
		{
			name: "Failure",
			extractor: func(c *gin.Context) (interface{}, error) {
				return nil, errors.New("extraction failed")
			},
			expectedStatus: http.StatusUnauthorized,
			checkOwnerID:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			router := gin.New()

			// Add middleware and test handler
			router.Use(OwnerContext(tt.extractor))
			router.GET("/test", func(c *gin.Context) {
				if tt.checkOwnerID {
					ownerID, exists := c.Get(OwnerContextKey)
					assert.True(t, exists)
					assert.Equal(t, tt.expectedOwnerID, ownerID)
				}
				c.Status(http.StatusOK)
			})

			// Make request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetOwnerID(t *testing.T) {
	// Test with gin context
	t.Run("With gin context", func(t *testing.T) {
		// Setup gin context
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(OwnerContextKey, "owner-456")

		// Get owner ID
		ownerID, err := GetOwnerID(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "owner-456", ownerID)
	})

	// Test with standard context
	t.Run("With standard context", func(t *testing.T) {
		// Setup context
		ctx := context.WithValue(context.Background(), OwnerContextKey, "owner-789")

		// Get owner ID
		ownerID, err := GetOwnerID(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "owner-789", ownerID)
	})

	// Test with missing owner ID
	t.Run("With missing owner ID", func(t *testing.T) {
		// Setup context
		ctx := context.Background()

		// Get owner ID
		_, err := GetOwnerID(ctx)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrOwnerIDNotFound, err)
	})
}

func TestExtractOwnerIDFromJWT(t *testing.T) {
	// Create test context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	t.Run("Success", func(t *testing.T) {
		// Set JWT claims in context
		c.Set("claims", jwt.MapClaims{
			"userID": "jwt-user-123",
			"email":  "test@example.com",
		})

		// Create extractor
		extractor := ExtractOwnerIDFromJWT("userID")
		ownerID, err := extractor(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "jwt-user-123", ownerID)
	})

	t.Run("No claims in context", func(t *testing.T) {
		// Create a new context without claims
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		// Create extractor
		extractor := ExtractOwnerIDFromJWT("userID")
		_, err := extractor(c)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT claims not found")
	})

	t.Run("Invalid claims format", func(t *testing.T) {
		// Create a new context with invalid claims
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("claims", "invalid-claims")

		// Create extractor
		extractor := ExtractOwnerIDFromJWT("userID")
		_, err := extractor(c)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JWT claims format")
	})

	t.Run("Claim not found", func(t *testing.T) {
		// Create a new context with claims but missing the requested one
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("claims", jwt.MapClaims{
			"email": "test@example.com",
		})

		// Create extractor
		extractor := ExtractOwnerIDFromJWT("userID")
		_, err := extractor(c)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "owner ID claim not found")
	})
}

func TestExtractOwnerIDFromHeader(t *testing.T) {
	// Setup gin context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Owner-ID", "header-owner-123")
	c.Request = req

	t.Run("Success", func(t *testing.T) {
		// Create extractor
		extractor := ExtractOwnerIDFromHeader("X-Owner-ID")
		ownerID, err := extractor(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "header-owner-123", ownerID)
	})

	t.Run("Header not found", func(t *testing.T) {
		// Create extractor
		extractor := ExtractOwnerIDFromHeader("X-Missing-Header")
		_, err := extractor(c)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "owner ID header is empty")
	})
}

func TestExtractOwnerIDFromQuery(t *testing.T) {
	// Setup gin context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/?ownerID=query-owner-123", nil)
	c.Request = req

	t.Run("Success", func(t *testing.T) {
		// Create extractor
		extractor := ExtractOwnerIDFromQuery("ownerID")
		ownerID, err := extractor(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "query-owner-123", ownerID)
	})

	t.Run("Query param not found", func(t *testing.T) {
		// Create extractor
		extractor := ExtractOwnerIDFromQuery("missingParam")
		_, err := extractor(c)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "owner ID query parameter is empty")
	})
}

func TestExtractOwnerIDFromCookie(t *testing.T) {
	// Setup gin context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "ownerID",
		Value: "cookie-owner-123",
	})
	c.Request = req

	t.Run("Success", func(t *testing.T) {
		// Create extractor
		extractor := ExtractOwnerIDFromCookie("ownerID")
		ownerID, err := extractor(c)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "cookie-owner-123", ownerID)
	})

	t.Run("Cookie not found", func(t *testing.T) {
		// Create extractor
		extractor := ExtractOwnerIDFromCookie("missingCookie")
		_, err := extractor(c)

		// Assert
		assert.Error(t, err)
	})
}

func TestCombineExtractors(t *testing.T) {
	// Setup gin context
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/?queryOwnerID=query-owner-123", nil)
	req.Header.Set("X-Owner-ID", "header-owner-123")
	c.Request = req

	// Create extractors
	failingExtractor := func(c *gin.Context) (interface{}, error) {
		return nil, errors.New("extraction failed")
	}

	headerExtractor := ExtractOwnerIDFromHeader("X-Owner-ID")
	queryExtractor := ExtractOwnerIDFromQuery("queryOwnerID")
	missingQueryExtractor := ExtractOwnerIDFromQuery("missingParam")

	t.Run("First extractor succeeds", func(t *testing.T) {
		combined := CombineExtractors(headerExtractor, queryExtractor)
		ownerID, err := combined(c)

		assert.NoError(t, err)
		assert.Equal(t, "header-owner-123", ownerID)
	})

	t.Run("Second extractor succeeds", func(t *testing.T) {
		combined := CombineExtractors(failingExtractor, queryExtractor)
		ownerID, err := combined(c)

		assert.NoError(t, err)
		assert.Equal(t, "query-owner-123", ownerID)
	})

	t.Run("All extractors fail", func(t *testing.T) {
		combined := CombineExtractors(failingExtractor, missingQueryExtractor)
		_, err := combined(c)

		assert.Error(t, err)
	})
}
