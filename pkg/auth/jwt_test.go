package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	config := DefaultJWTConfig()
	config.Secret = "test-secret"

	middleware := JWTMiddleware(config)

	// Test with no Authorization header
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)

	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header is required")

	// Test with invalid Authorization header format
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat")

	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header must be in the format 'Bearer {token}'")

	// Test with invalid token
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	middleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test with valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "1",
		"name":  "John Doe",
		"roles": []string{"admin"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(config.Secret))

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+tokenString)

	// Zamiast przypisywać do c.Next, użyjmy zmiennej w kontekście
	c.Set("nextCalled", false)
	originalHandler := middleware
	middleware = func(c *gin.Context) {
		originalHandler(c)
		if !c.IsAborted() {
			c.Set("nextCalled", true)
		}
	}

	middleware(c)

	nextCalled, _ := c.Get("nextCalled")
	assert.True(t, nextCalled.(bool))
	assert.NotNil(t, c.Keys["claims"])
}

func TestGenerateJWT(t *testing.T) {
	config := DefaultJWTConfig()
	config.Secret = "test-secret"

	// Test with standard claims
	claims := jwt.MapClaims{
		"sub":   "1",
		"name":  "John Doe",
		"roles": []string{"admin"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	}

	tokenString, err := GenerateJWT(config, claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Secret), nil
	})

	assert.NoError(t, err)
	assert.True(t, token.Valid)

	parsedClaims := token.Claims.(jwt.MapClaims)
	assert.Equal(t, "1", parsedClaims["sub"])
	assert.Equal(t, "John Doe", parsedClaims["name"])
}

func TestGenerateJWTWithStandardClaims(t *testing.T) {
	config := DefaultJWTConfig()
	config.Secret = "test-secret"

	// Test with standard claims and custom claims
	customClaims := map[string]interface{}{
		"name":  "John Doe",
		"roles": []string{"admin"},
	}

	tokenString, err := GenerateJWTWithStandardClaims(config, "1", customClaims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Secret), nil
	})

	assert.NoError(t, err)
	assert.True(t, token.Valid)

	parsedClaims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, "1", parsedClaims["sub"])
	assert.Equal(t, "John Doe", parsedClaims["name"])
	assert.Equal(t, "refine-gin", parsedClaims["iss"])
}

func TestExtractSubjectFromToken(t *testing.T) {
	// Generate a token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "1",
		"name": "John Doe",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	secret := "test-secret"
	tokenString, _ := token.SignedString([]byte(secret))

	// Extract subject
	subject, err := ExtractSubjectFromToken(tokenString, secret)
	assert.NoError(t, err)
	assert.Equal(t, "1", subject)

	// Test with invalid token
	subject, err = ExtractSubjectFromToken("invalid-token", secret)
	assert.Error(t, err)
	assert.Empty(t, subject)

	// Test with token without subject
	token = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": "John Doe",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	tokenString, _ = token.SignedString([]byte(secret))

	subject, err = ExtractSubjectFromToken(tokenString, secret)
	assert.Error(t, err)
	assert.Empty(t, subject)
}

func TestExtractClaimsFromToken(t *testing.T) {
	// Generate a token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "1",
		"name":  "John Doe",
		"roles": []string{"admin"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	secret := "test-secret"
	tokenString, _ := token.SignedString([]byte(secret))

	// Extract claims
	claims, err := ExtractClaimsFromToken(tokenString, secret)
	assert.NoError(t, err)
	assert.Equal(t, "1", claims["sub"])
	assert.Equal(t, "John Doe", claims["name"])

	// Test with invalid token
	claims, err = ExtractClaimsFromToken("invalid-token", secret)
	assert.Error(t, err)
	assert.Nil(t, claims)

	// Test with token signed using a different algorithm
	token = jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub":  "1",
		"name": "John Doe",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	tokenString, _ = token.SignedString([]byte(secret))

	claims, err = ExtractClaimsFromToken(tokenString, secret)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "unexpected signing method")
}
