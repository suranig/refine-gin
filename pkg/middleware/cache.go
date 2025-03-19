package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/utils"
)

// CacheConfig contains configuration for cache middleware
type CacheConfig struct {
	// MaxAge specifies cache validity period in seconds
	MaxAge int
	// DisableCache completely disables caching
	DisableCache bool
	// Methods specifies which HTTP methods should be cached
	Methods []string
	// VaryHeaders specifies headers that affect caching
	VaryHeaders []string
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxAge:       60, // 1 minute
		DisableCache: false,
		Methods:      []string{"GET", "HEAD"},
		VaryHeaders:  []string{"Accept", "Accept-Encoding", "Authorization"},
	}
}

// Cache middleware sets cache-control, ETag and Vary headers
func Cache(config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cache only for specified methods (default GET and HEAD)
		methodAllowed := false
		for _, method := range config.Methods {
			if c.Request.Method == method {
				methodAllowed = true
				break
			}
		}

		if !methodAllowed || config.DisableCache {
			c.Next()
			return
		}

		// For GET requests, check If-None-Match before continuing
		if c.Request.Method == "GET" {
			etag := utils.GenerateRequestETag(c.Request)
			ifNoneMatch := c.GetHeader("If-None-Match")

			if utils.IsETagMatch(etag, ifNoneMatch) {
				c.AbortWithStatus(http.StatusNotModified)
				return
			}

			// Set cache headers
			utils.SetCacheHeaders(c.Writer, config.MaxAge, etag, nil, config.VaryHeaders)
		} else {
			// For other methods, just set Cache-Control and Vary
			c.Writer.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", config.MaxAge))
			if len(config.VaryHeaders) > 0 {
				c.Writer.Header().Set("Vary", strings.Join(config.VaryHeaders, ", "))
			}
		}

		c.Next()
	}
}

// CacheByResource middleware caches with resource name consideration
func CacheByResource(resourceName string, config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cache only for specified methods (default GET and HEAD)
		methodAllowed := false
		for _, method := range config.Methods {
			if c.Request.Method == method {
				methodAllowed = true
				break
			}
		}

		if !methodAllowed || config.DisableCache {
			c.Next()
			return
		}

		// Generate and check ETag for resource
		var etag string
		id := c.Param("id")
		if id != "" {
			// Cache for single resource (/:id)
			etag = utils.GenerateResourceETag(resourceName, id)
		} else {
			// Cache for list (considering query parameters)
			etag = utils.GenerateQueryETag(c.Request.URL.RawQuery)
		}

		ifNoneMatch := c.GetHeader("If-None-Match")
		if utils.IsETagMatch(etag, ifNoneMatch) {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}

		// Set cache headers
		utils.SetCacheHeaders(c.Writer, config.MaxAge, etag, nil, config.VaryHeaders)

		c.Next()
	}
}

// NoCacheMiddleware disables caching for specific endpoints
func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.DisableCaching(c.Writer)
		c.Next()
	}
}
