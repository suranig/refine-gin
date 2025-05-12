package handler

import (
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/dto"
	"github.com/stanxing/refine-gin/pkg/repository"
	"github.com/stanxing/refine-gin/pkg/resource"
	"github.com/stanxing/refine-gin/pkg/utils"
)

// GenerateGetHandler generates a handler for READ operations
func GenerateGetHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
	// Use default DTO provider if none is specified
	dtoProvider := &dto.DefaultDTOProvider{
		Model: res.GetModel(),
	}

	return generateGetHandlerWithDTO("id", res, repo, dtoProvider)
}

// GenerateGetHandlerWithParam generates a handler for READ operations with custom ID parameter name
func GenerateGetHandlerWithParam(res resource.Resource, repo repository.Repository, idParamName string) gin.HandlerFunc {
	// Use default DTO provider if none is specified
	dtoProvider := &dto.DefaultDTOProvider{
		Model: res.GetModel(),
	}

	return generateGetHandlerWithDTO(idParamName, res, repo, dtoProvider)
}

// GenerateGetHandlerWithDTO generates a handler for READ operations with custom DTO provider
func GenerateGetHandlerWithDTO(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return generateGetHandlerWithDTO("id", res, repo, dtoProvider)
}

// GenerateGetHandlerWithParamAndDTO generates a handler for READ operations with custom ID parameter name and DTO provider
func GenerateGetHandlerWithParamAndDTO(res resource.Resource, repo repository.Repository, idParamName string, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return generateGetHandlerWithDTO(idParamName, res, repo, dtoProvider)
}

// Helper function to avoid code duplication
func generateGetHandlerWithDTO(idParamName string, res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get ID from URL parameters using custom parameter name
		id := c.Param(idParamName)

		// Generate ETag for cache validation
		etag := utils.GenerateResourceETag(res.GetName(), id)
		ifNoneMatch := c.GetHeader("If-None-Match")

		// Check if client's cached version is still valid
		if utils.IsETagMatch(etag, ifNoneMatch) {
			c.Status(http.StatusNotModified)
			return
		}

		// Call repository
		data, err := repo.Get(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}

		// Transform model to DTO if provider is available
		if dtoProvider != nil {
			dtoData, err := dtoProvider.TransformFromModel(data)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error transforming data: " + err.Error()})
				return
			}
			data = dtoData
		}

		// Return result
		c.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	}
}

// getLastModifiedTimestamp attempts to extract updated_at or created_at from model
func getLastModifiedTimestamp(data interface{}) (bool, time.Time) {
	// Try to access UpdatedAt or CreatedAt field using reflection
	val := reflect.ValueOf(data)

	// Handle pointer indirection
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check if it's a struct
	if val.Kind() != reflect.Struct {
		return false, time.Time{}
	}

	// Try to get UpdatedAt first
	updatedField := val.FieldByName("UpdatedAt")
	if updatedField.IsValid() && updatedField.Type().AssignableTo(reflect.TypeOf(time.Time{})) {
		return true, updatedField.Interface().(time.Time)
	}

	// Fall back to CreatedAt if available
	createdField := val.FieldByName("CreatedAt")
	if createdField.IsValid() && createdField.Type().AssignableTo(reflect.TypeOf(time.Time{})) {
		return true, createdField.Interface().(time.Time)
	}

	return false, time.Time{}
}
