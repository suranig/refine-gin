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

		// Generate metadata for the resource
		metadata := gin.H{
			"name":        res.GetName(),
			"label":       res.GetLabel(),
			"icon":        res.GetIcon(),
			"operations":  res.GetOperations(),
			"fields":      res.GetFields(),
			"defaultSort": res.GetDefaultSort(),
			"relations":   res.GetRelations(),
			"filters":     res.GetFilters(),
			"idField":     res.GetIDFieldName(),
			"lists": gin.H{
				"filterable": res.GetFilterableFields(),
				"searchable": res.GetSearchable(),
				"sortable":   res.GetSortableFields(),
				"required":   res.GetRequiredFields(),
				"table":      res.GetTableFields(),
				"form":       res.GetFormFields(),
			},
		}

		// Set cache headers
		utils.SetCacheHeaders(c.Writer, 300, etag, nil, []string{"Accept", "Accept-Encoding", "Authorization"})

		c.JSON(http.StatusOK, metadata)
	}
}

// RegisterOptionsEndpoint registers the OPTIONS endpoint for a resource
func RegisterOptionsEndpoint(router *gin.RouterGroup, res resource.Resource) {
	router.OPTIONS("/"+res.GetName(), GenerateOptionsHandler(res))
}
