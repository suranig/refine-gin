package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig contains configuration for JWT authentication
type JWTConfig struct {
	// Secret key used to sign tokens
	Secret string

	// Token expiration time
	ExpirationTime time.Duration

	// Token issuer
	Issuer string

	// Token audience
	Audience string

	// Function to extract claims from token
	ClaimsExtractor func(token *jwt.Token) (interface{}, error)
}

// DefaultJWTConfig returns a default JWT configuration
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:         "secret", // Should be overridden in production
		ExpirationTime: time.Hour * 24,
		Issuer:         "refine-gin",
		Audience:       "refine-gin-api",
		ClaimsExtractor: func(token *jwt.Token) (interface{}, error) {
			return token.Claims, nil
		},
	}
}

// JWTMiddleware creates a middleware for JWT authentication
func JWTMiddleware(config JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// Check if the header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in the format 'Bearer {token}'"})
			return
		}

		tokenString := parts[1]

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(config.Secret), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Check if token is valid
		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Extract claims
		claims, err := config.ClaimsExtractor(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Set claims in context
		c.Set("claims", claims)

		c.Next()
	}
}

// GenerateJWT generates a JWT token
func GenerateJWT(config JWTConfig, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(config.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateJWTWithStandardClaims generates a JWT token with standard claims
func GenerateJWTWithStandardClaims(config JWTConfig, subject string, customClaims map[string]interface{}) (string, error) {
	now := time.Now()

	// Utwórz MapClaims zamiast RegisteredClaims
	claims := jwt.MapClaims{
		"sub": subject,
		"iat": now.Unix(),
		"exp": now.Add(config.ExpirationTime).Unix(),
		"iss": config.Issuer,
		"aud": config.Audience,
	}

	// Dodaj niestandardowe claims
	for key, value := range customClaims {
		claims[key] = value
	}

	// Utwórz token z claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Podpisz token kluczem tajnym
	tokenString, err := token.SignedString([]byte(config.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ExtractSubjectFromToken extracts the subject from a JWT token
func ExtractSubjectFromToken(tokenString string, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	// Check if token is valid
	if !token.Valid {
		return "", errors.New("invalid token")
	}

	// Extract subject from claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("subject not found in claims")
	}

	return subject, nil
}

// ExtractClaimsFromToken extracts all claims from a JWT token
func ExtractClaimsFromToken(tokenString string, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}
