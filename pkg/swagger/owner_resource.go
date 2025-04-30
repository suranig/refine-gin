package swagger

import (
	"fmt"

	"github.com/stanxing/refine-gin/pkg/resource"
)

// RegisterOwnerResourceSwagger registers Swagger documentation for owner resources
func RegisterOwnerResourceSwagger(openAPI *OpenAPI, res resource.OwnerResource) {
	// Generate standard paths for the resource
	generateResourcePaths(openAPI, res)

	// Add security scheme for JWT if not already added
	if _, exists := openAPI.Components.SecuritySchemes["bearerAuth"]; !exists {
		openAPI.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT token authorization with owner ID",
		}
	}

	// For each path, add owner-specific information
	resourceName := res.GetName()

	// Add owner-specific parameter documentation to list endpoint
	listPath := "/" + resourceName
	if pathItem, exists := openAPI.Paths[listPath]; exists {
		if operation, hasGet := pathItem["get"]; hasGet {
			// Add Security requirement
			operation.Security = []map[string][]string{
				{"bearerAuth": {}},
			}

			// Add owner description
			if operation.Description != "" {
				operation.Description += ". "
			}
			operation.Description += "Only resources owned by the authenticated user will be returned."

			// Update the path item
			pathItem["get"] = operation
			openAPI.Paths[listPath] = pathItem
		}
	}

	// Add owner-specific parameter documentation to get/update/delete endpoints
	itemPath := fmt.Sprintf("/%s/{id}", resourceName)
	if pathItem, exists := openAPI.Paths[itemPath]; exists {
		// GET endpoint
		if operation, hasGet := pathItem["get"]; hasGet {
			// Add Security requirement
			operation.Security = []map[string][]string{
				{"bearerAuth": {}},
			}

			// Add owner description
			if operation.Description != "" {
				operation.Description += ". "
			}
			operation.Description += "Only accessible if the resource is owned by the authenticated user."

			// Update responses to include 403
			operation.Responses["403"] = Response{
				Description: "Forbidden - The resource is owned by another user",
			}

			// Update the path item
			pathItem["get"] = operation
		}

		// PUT endpoint
		if operation, hasPut := pathItem["put"]; hasPut {
			// Add Security requirement
			operation.Security = []map[string][]string{
				{"bearerAuth": {}},
			}

			// Add owner description
			if operation.Description != "" {
				operation.Description += ". "
			}
			operation.Description += "Only resources owned by the authenticated user can be updated."

			// Update responses to include 403
			operation.Responses["403"] = Response{
				Description: "Forbidden - The resource is owned by another user",
			}

			// Update the path item
			pathItem["put"] = operation
		}

		// DELETE endpoint
		if operation, hasDelete := pathItem["delete"]; hasDelete {
			// Add Security requirement
			operation.Security = []map[string][]string{
				{"bearerAuth": {}},
			}

			// Add owner description
			if operation.Description != "" {
				operation.Description += ". "
			}
			operation.Description += "Only resources owned by the authenticated user can be deleted."

			// Update responses to include 403
			operation.Responses["403"] = Response{
				Description: "Forbidden - The resource is owned by another user",
			}

			// Update the path item
			pathItem["delete"] = operation
		}

		// Update the path item
		openAPI.Paths[itemPath] = pathItem
	}

	// Update create endpoint
	createPath := "/" + resourceName
	if pathItem, exists := openAPI.Paths[createPath]; exists {
		if operation, hasPost := pathItem["post"]; hasPost {
			// Add Security requirement
			operation.Security = []map[string][]string{
				{"bearerAuth": {}},
			}

			// Add owner description
			if operation.Description != "" {
				operation.Description += ". "
			}
			operation.Description += "The owner ID will be automatically set to the authenticated user's ID."

			// Update the path item
			pathItem["post"] = operation
			openAPI.Paths[createPath] = pathItem
		}
	}

	// Update batch endpoints
	updateBatchEndpoints(openAPI, resourceName)
}

// updateBatchEndpoints updates batch operation documentation for owner resources
func updateBatchEndpoints(openAPI *OpenAPI, resourceName string) {
	// Batch create endpoint
	batchCreatePath := fmt.Sprintf("/%s/batch", resourceName)
	if pathItem, exists := openAPI.Paths[batchCreatePath]; exists {
		if operation, hasPost := pathItem["post"]; hasPost {
			// Add Security requirement
			operation.Security = []map[string][]string{
				{"bearerAuth": {}},
			}

			// Add owner description
			if operation.Description != "" {
				operation.Description += ". "
			}
			operation.Description += "The owner ID for all created resources will be set to the authenticated user's ID."

			// Update the path item
			pathItem["post"] = operation
			openAPI.Paths[batchCreatePath] = pathItem
		}
	}

	// Batch update endpoint
	batchUpdatePath := fmt.Sprintf("/%s/batch", resourceName)
	if pathItem, exists := openAPI.Paths[batchUpdatePath]; exists {
		if operation, hasPut := pathItem["put"]; hasPut {
			// Add Security requirement
			operation.Security = []map[string][]string{
				{"bearerAuth": {}},
			}

			// Add owner description
			if operation.Description != "" {
				operation.Description += ". "
			}
			operation.Description += "Only resources owned by the authenticated user can be updated."

			// Update responses to include 403
			operation.Responses["403"] = Response{
				Description: "Forbidden - One or more resources are owned by another user",
			}

			// Update the path item
			pathItem["put"] = operation
			openAPI.Paths[batchUpdatePath] = pathItem
		}
	}

	// Batch delete endpoint
	batchDeletePath := fmt.Sprintf("/%s/batch", resourceName)
	if pathItem, exists := openAPI.Paths[batchDeletePath]; exists {
		if operation, hasDelete := pathItem["delete"]; hasDelete {
			// Add Security requirement
			operation.Security = []map[string][]string{
				{"bearerAuth": {}},
			}

			// Add owner description
			if operation.Description != "" {
				operation.Description += ". "
			}
			operation.Description += "Only resources owned by the authenticated user can be deleted."

			// Update responses to include 403
			operation.Responses["403"] = Response{
				Description: "Forbidden - One or more resources are owned by another user",
			}

			// Update the path item
			pathItem["delete"] = operation
			openAPI.Paths[batchDeletePath] = pathItem
		}
	}
}
