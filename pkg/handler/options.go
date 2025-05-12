package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/resource"
	"github.com/stanxing/refine-gin/pkg/utils"
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

		// Get user roles from context if available
		var userRoles []string
		userRolesValue, exists := c.Get("userRoles")
		if exists {
			if roles, ok := userRolesValue.([]string); ok {
				userRoles = roles
			}
		}

		// If user roles are available, filter fields based on permissions
		if len(userRoles) > 0 {
			filteredFields := make([]resource.FieldMetadata, 0, len(metadata.Fields))

			// Filter field metadata based on read permissions
			for _, field := range metadata.Fields {
				if field.Permissions == nil {
					filteredFields = append(filteredFields, field)
					continue
				}

				allowedRoles, exists := field.Permissions["read"]
				if !exists || len(allowedRoles) == 0 {
					filteredFields = append(filteredFields, field)
					continue
				}

				hasPermission := false
				for _, userRole := range userRoles {
					for _, allowedRole := range allowedRoles {
						if userRole == allowedRole {
							hasPermission = true
							break
						}
					}
					if hasPermission {
						break
					}
				}

				if hasPermission {
					filteredFields = append(filteredFields, field)
				}
			}

			metadata.Fields = filteredFields
		}

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
			"permissions": metadata.Permissions,
			"lists": gin.H{
				"filterable": metadata.FilterableFields,
				"searchable": metadata.Searchable,
				"sortable":   metadata.SortableFields,
				"required":   metadata.RequiredFields,
				"table":      metadata.TableFields,
				"form":       metadata.FormFields,
			},
		}

		c.JSON(http.StatusOK, responseMetadata)
	}
}

// RegisterOptionsEndpoint registers the OPTIONS endpoint for a resource
func RegisterOptionsEndpoint(router *gin.RouterGroup, res resource.Resource) {
	router.OPTIONS("/"+res.GetName(), GenerateOptionsHandler(res))
}
