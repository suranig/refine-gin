package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"github.com/suranig/refine-gin/pkg/utils"
)

// APIConfigResponse represents the response structure for API configuration
type APIConfigResponse struct {
	// Resources metadata by name
	Resources map[string]resource.ResourceMetadata `json:"resources"`

	// Additional configuration like permissions
	Config map[string]interface{} `json:"config,omitempty"`
}

// GenerateAPIConfigHandler creates a handler for exposing API configuration
func GenerateAPIConfigHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		resources := make(map[string]resource.ResourceMetadata)
		allResources := resource.GlobalResourceRegistry.GetAll()

		etag := utils.GenerateETag(fmt.Sprintf("%d", len(allResources)))
		ifNoneMatch := c.GetHeader("If-None-Match")

		// Check if client's cached version is still valid
		if utils.IsETagMatch(etag, ifNoneMatch) {
			c.Status(http.StatusNotModified)
			return
		}

		// Generate metadata for each resource
		for _, res := range allResources {
			// Generate metadata for this resource
			metadata := resource.GenerateResourceMetadata(res)
			resources[res.GetName()] = metadata
		}

		// Create response
		response := APIConfigResponse{
			Resources: resources,
			Config: map[string]interface{}{
				"version": "1.0.0", // Add library version
			},
		}

		// Set cache headers
		utils.SetCacheHeaders(c.Writer, 300, etag, nil, []string{"Accept", "Accept-Encoding", "Authorization"})

		// Return response
		c.JSON(http.StatusOK, response)
	}
}

// RegisterAPIConfigEndpoint registers the API configuration endpoint
func RegisterAPIConfigEndpoint(router *gin.RouterGroup) {
	router.GET("/config", GenerateAPIConfigHandler())
}
