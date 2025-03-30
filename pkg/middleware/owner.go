package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// OwnerContextKey is the key used to store the owner ID in the context
const OwnerContextKey = "ownerID"

// ErrOwnerIDNotFound is returned when the owner ID cannot be found
var ErrOwnerIDNotFound = errors.New("owner ID not found")

// ExtractOwnerIDFunc is a function that extracts an owner ID from a gin context
type ExtractOwnerIDFunc func(c *gin.Context) (interface{}, error)

// OwnerContext middleware extracts and stores the owner ID in the context
func OwnerContext(extractor ExtractOwnerIDFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID, err := extractor(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Store owner ID in context
		c.Set(OwnerContextKey, ownerID)
		c.Next()
	}
}

// GetOwnerID extracts owner ID from the context
func GetOwnerID(ctx context.Context) (interface{}, error) {
	// Check if we have a gin context
	gc, ok := ctx.(*gin.Context)
	if ok {
		// Extract from gin context
		if ownerID, exists := gc.Get(OwnerContextKey); exists {
			return ownerID, nil
		}
	}

	// Check if owner ID is set directly in the context
	if ownerID := ctx.Value(OwnerContextKey); ownerID != nil {
		return ownerID, nil
	}

	return nil, ErrOwnerIDNotFound
}

// ExtractOwnerIDFromJWT extracts the owner ID from JWT claims
func ExtractOwnerIDFromJWT(claimName string) ExtractOwnerIDFunc {
	return func(c *gin.Context) (interface{}, error) {
		claimsValue, exists := c.Get("claims")
		if !exists {
			return nil, errors.New("JWT claims not found in context")
		}

		claims, ok := claimsValue.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("invalid JWT claims format")
		}

		ownerID, exists := claims[claimName]
		if !exists {
			return nil, errors.New("owner ID claim not found in JWT")
		}

		return ownerID, nil
	}
}

// ExtractOwnerIDFromHeader extracts the owner ID from an HTTP header
func ExtractOwnerIDFromHeader(headerName string) ExtractOwnerIDFunc {
	return func(c *gin.Context) (interface{}, error) {
		ownerID := c.GetHeader(headerName)
		if ownerID == "" {
			return nil, errors.New("owner ID header is empty")
		}
		return ownerID, nil
	}
}

// ExtractOwnerIDFromQuery extracts the owner ID from query parameters
func ExtractOwnerIDFromQuery(paramName string) ExtractOwnerIDFunc {
	return func(c *gin.Context) (interface{}, error) {
		ownerID := c.Query(paramName)
		if ownerID == "" {
			return nil, errors.New("owner ID query parameter is empty")
		}
		return ownerID, nil
	}
}

// ExtractOwnerIDFromCookie extracts the owner ID from a cookie
func ExtractOwnerIDFromCookie(cookieName string) ExtractOwnerIDFunc {
	return func(c *gin.Context) (interface{}, error) {
		cookie, err := c.Cookie(cookieName)
		if err != nil {
			return nil, err
		}
		if cookie == "" {
			return nil, errors.New("owner ID cookie is empty")
		}
		return cookie, nil
	}
}

// CombineExtractors chains multiple extractors and returns the first successful result
func CombineExtractors(extractors ...ExtractOwnerIDFunc) ExtractOwnerIDFunc {
	return func(c *gin.Context) (interface{}, error) {
		var lastErr error
		for _, extractor := range extractors {
			ownerID, err := extractor(c)
			if err == nil {
				return ownerID, nil
			}
			lastErr = err
		}
		return nil, lastErr
	}
}
