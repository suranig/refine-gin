package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"github.com/suranig/refine-gin/pkg/utils"
)

// GenerateOptionsHandler creates a handler for the OPTIONS method
// This handler returns detailed metadata about the resource
func GenerateOptionsHandler(res resource.Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate ETag based on resource name for cache validation
		etag := utils.GenerateResourceETag(res.GetName(), "options")
		ifNoneMatch := c.GetHeader("If-None-Match")

		// Check if client's cached version is still valid
		if utils.IsETagMatch(etag, ifNoneMatch) {
			c.Status(http.StatusNotModified)
			return
		}

		// Generate full metadata for the resource
		metadata := resource.GenerateResourceMetadata(res)

		// Format metadata as gin.H for response
		responseMetadata := gin.H{
			"name":        metadata.Name,
			"label":       metadata.Label,
			"icon":        metadata.Icon,
			"operations":  metadata.Operations,
			"fields":      metadata.Fields,
			"defaultSort": metadata.DefaultSort,
			"relations":   metadata.Relations,
			"filters":     metadata.Filters,
			"idField":     metadata.IDFieldName,
			"lists": gin.H{
				"filterable": metadata.FilterableFields,
				"searchable": metadata.Searchable,
				"sortable":   metadata.SortableFields,
				"required":   metadata.RequiredFields,
				"table":      metadata.TableFields,
				"form":       metadata.FormFields,
			},
		}

		// Set cache headers
		utils.SetCacheHeaders(c.Writer, 300, etag, nil, []string{"Accept", "Accept-Encoding", "Authorization"})

		c.JSON(http.StatusOK, responseMetadata)
	}
}

// RegisterOptionsEndpoint registers the OPTIONS endpoint for a resource
func RegisterOptionsEndpoint(router *gin.RouterGroup, res resource.Resource) {
	router.OPTIONS("/"+res.GetName(), GenerateOptionsHandler(res))
}
