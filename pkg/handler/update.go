package handler

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/dto"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
)

// validateNestedJsonFields validates all JSON fields in the model against their JsonConfig if available
func validateNestedJsonFields(res resource.Resource, model interface{}) error {
	if model == nil {
		return nil
	}

	// Get all resource fields
	fields := res.GetFields()

	// Find JSON fields with nested configuration
	for _, field := range fields {
		if field.Type == "json" && field.Json != nil && field.Json.Nested {
			// Get the JSON field value from the model
			modelValue := reflect.ValueOf(model)
			if modelValue.Kind() == reflect.Ptr {
				modelValue = modelValue.Elem()
			}

			// Skip if not a struct
			if modelValue.Kind() != reflect.Struct {
				continue
			}

			// Try to get the field
			fieldValue := modelValue.FieldByName(field.Name)
			if !fieldValue.IsValid() {
				continue // Field not found
			}

			// Skip nil values
			if fieldValue.IsNil() {
				continue
			}

			// Extract the field value
			jsonValue := fieldValue.Interface()

			// Validate against the config
			valid, errors := resource.ValidateNestedJson(jsonValue, field.Json)
			if !valid {
				return fmt.Errorf("validation failed for field '%s': %s", field.Name, strings.Join(errors, "; "))
			}
		}
	}

	return nil
}

// GenerateUpdateHandler generates a handler for UPDATE operations with DTO support
func GenerateUpdateHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get ID from URL parameters
		id := c.Param("id")

		// Create a new instance of the DTO
		dtoInstance := dtoProvider.GetUpdateDTO()

		// Parse request data into DTO
		if err := c.ShouldBindJSON(dtoInstance); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Transform DTO to model
		model, err := dtoProvider.TransformToModel(dtoInstance)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Filter out read-only fields from the model
		model = resource.FilterOutReadOnlyFields(model, res)

		// Validate nested JSON fields if present
		if err := validateNestedJsonFields(res, model); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON validation failed: " + err.Error()})
			return
		}

		// Get the database connection from repository
		db := repo.Query(c.Request.Context())

		// Validate relations (if any) - only perform if repository has DB access
		if db != nil && len(res.GetRelations()) > 0 {
			// Validate relations
			if err := resource.ValidateRelations(db, model); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Relation validation failed: " + err.Error()})
				return
			}
		}

		// Call repository
		updatedModel, err := repo.Update(c.Request.Context(), id, model)
		if err != nil {
			// Check if it's a "not found" error
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
				return
			}
			// Handle other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform model to response DTO
		responseDTO, err := dtoProvider.TransformFromModel(updatedModel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return result
		c.JSON(http.StatusOK, gin.H{
			"data": responseDTO,
		})
	}
}

// GenerateUpdateHandlerWithParam generates a handler for UPDATE operations with DTO support and custom ID parameter name
func GenerateUpdateHandlerWithParam(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider, idParamName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get ID from URL parameters using custom parameter name
		id := c.Param(idParamName)

		// Create a new instance of the DTO
		dtoInstance := dtoProvider.GetUpdateDTO()

		// Parse request data into DTO
		if err := c.ShouldBindJSON(dtoInstance); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Transform DTO to model
		model, err := dtoProvider.TransformToModel(dtoInstance)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Filter out read-only fields from the model
		model = resource.FilterOutReadOnlyFields(model, res)

		// Validate nested JSON fields if present
		if err := validateNestedJsonFields(res, model); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON validation failed: " + err.Error()})
			return
		}

		// Get the database connection from repository
		db := repo.Query(c.Request.Context())

		// Validate relations (if any) - only perform if repository has DB access
		if db != nil && len(res.GetRelations()) > 0 {
			// Validate relations
			if err := resource.ValidateRelations(db, model); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Relation validation failed: " + err.Error()})
				return
			}
		}

		// Call repository
		updatedModel, err := repo.Update(c.Request.Context(), id, model)
		if err != nil {
			// Check if it's a "not found" error
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
				return
			}
			// Handle other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform model to response DTO
		responseDTO, err := dtoProvider.TransformFromModel(updatedModel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return result
		c.JSON(http.StatusOK, gin.H{
			"data": responseDTO,
		})
	}
}
