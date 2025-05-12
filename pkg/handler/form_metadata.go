package handler

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/dto"
	"github.com/stanxing/refine-gin/pkg/resource"
	"github.com/stanxing/refine-gin/pkg/utils"
)

// FormMetadataResponse represents the response structure for form metadata
type FormMetadataResponse struct {
	// Form layout configuration
	Layout *resource.FormLayoutMetadata `json:"layout,omitempty"`

	// Fields configuration
	Fields []resource.FieldMetadata `json:"fields"`

	// Default values for the form
	DefaultValues map[string]interface{} `json:"defaultValues,omitempty"`

	// Field dependencies
	Dependencies map[string][]string `json:"dependencies,omitempty"`
}

// GenerateFormMetadataHandler creates a handler for exposing form metadata
func GenerateFormMetadataHandler(res resource.Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate ETag based on resource name and fields
		etag := utils.GenerateETag(res.GetName() + "-form")
		ifNoneMatch := c.GetHeader("If-None-Match")

		// Check if client's cached version is still valid
		if utils.IsETagMatch(etag, ifNoneMatch) {
			c.Status(http.StatusNotModified)
			return
		}

		// Generate metadata
		metadata := FormMetadataResponse{
			Fields: resource.GenerateFieldsMetadata(res.GetFields()),
		}

		// Add form layout if available
		if formLayout := res.GetFormLayout(); formLayout != nil {
			metadata.Layout = resource.GenerateFormLayoutMetadata(formLayout)
		}

		// Extract default values from model
		metadata.DefaultValues = extractDefaultValues(res.GetModel())

		// Extract field dependencies
		metadata.Dependencies = extractFieldDependencies(res.GetFields())

		// Return response
		c.JSON(http.StatusOK, metadata)
	}
}

// extractDefaultValues extracts default values from model
func extractDefaultValues(model interface{}) map[string]interface{} {
	if model == nil {
		return nil
	}

	result := make(map[string]interface{})

	// Get model type via reflection
	modelValue := reflect.ValueOf(model)

	// Handle pointers
	if modelValue.Kind() == reflect.Ptr {
		if modelValue.IsNil() {
			return nil
		}
		modelValue = modelValue.Elem()
	}

	// Only process structs
	if modelValue.Kind() != reflect.Struct {
		return result
	}

	modelType := modelValue.Type()

	// Iterate over fields to extract default values
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Get field value
		fieldValue := modelValue.Field(i)

		// Skip zero values
		if fieldValue.IsZero() {
			continue
		}

		// Get field name from JSON tag if available
		fieldName := field.Name
		if jsonTag, ok := field.Tag.Lookup("json"); ok {
			// Extract the field name from the JSON tag
			parts := utils.SplitTagParts(jsonTag)
			if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			} else if parts[0] == "-" {
				// Skip fields marked as "-" in JSON tag
				continue
			}
		}

		// Add to result
		result[fieldName] = fieldValue.Interface()
	}

	return result
}

// extractFieldDependencies extracts field dependencies from form configurations
func extractFieldDependencies(fields []resource.Field) map[string][]string {
	if len(fields) == 0 {
		return nil
	}

	result := make(map[string][]string)

	for _, field := range fields {
		// Skip fields without form config
		if field.Form == nil {
			continue
		}

		// Check DependentOn field
		if field.Form.DependentOn != "" {
			if _, ok := result[field.Form.DependentOn]; !ok {
				result[field.Form.DependentOn] = make([]string, 0)
			}
			result[field.Form.DependentOn] = append(result[field.Form.DependentOn], field.Name)
		}

		// Check Dependent configuration
		if field.Form.Dependent != nil && field.Form.Dependent.Field != "" {
			if _, ok := result[field.Form.Dependent.Field]; !ok {
				result[field.Form.Dependent.Field] = make([]string, 0)
			}
			result[field.Form.Dependent.Field] = append(result[field.Form.Dependent.Field], field.Name)
		}
	}

	return result
}

// RegisterGlobalFormMetadataEndpoint registers the global form metadata endpoint
func RegisterGlobalFormMetadataEndpoint(router *gin.RouterGroup, res resource.Resource) {
	router.GET("/forms/"+res.GetName()+"/metadata", GenerateFormMetadataHandler(res))
}

// RegisterResourceFormEndpoints registers all form-related endpoints for a resource
func RegisterResourceFormEndpoints(router *gin.RouterGroup, res resource.Resource) {
	resourceName := res.GetName()
	resourcePath := utils.Pluralize(resourceName)
	resourceGroup := router.Group("/" + resourcePath)

	// Register form metadata endpoint
	resourceGroup.GET("/form", GenerateFormMetadataHandler(res))

	// Register pre-filled form endpoint (for edit forms)
	resourceGroup.GET("/form/:id", func(c *gin.Context) {
		// Get ID from path
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Message: "ID is required",
			})
			return
		}

		// Try to get repository from context
		repo, exists := c.Get("repository")
		if !exists {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "Repository not available",
			})
			return
		}

		// Use FindByID method via reflection as we don't know the concrete type
		repoValue := reflect.ValueOf(repo)
		findMethod := repoValue.MethodByName("FindByID")
		if !findMethod.IsValid() {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "Repository does not implement FindByID method",
			})
			return
		}

		// Call FindByID
		results := findMethod.Call([]reflect.Value{reflect.ValueOf(id)})
		if len(results) != 2 {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "Unexpected repository method signature",
			})
			return
		}

		// Check for error
		errValue := results[1]
		if !errValue.IsNil() {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Message: "Item not found: " + errValue.Interface().(error).Error(),
			})
			return
		}

		// Get the item
		item := results[0].Interface()

		// Generate form metadata
		metadata := FormMetadataResponse{
			Fields: resource.GenerateFieldsMetadata(res.GetFields()),
		}

		// Add form layout if available
		if formLayout := res.GetFormLayout(); formLayout != nil {
			metadata.Layout = resource.GenerateFormLayoutMetadata(formLayout)
		}

		// Convert item to map if needed
		var defaultValues map[string]interface{}
		if mapValues, ok := item.(map[string]interface{}); ok {
			defaultValues = mapValues
		} else {
			// Convert using reflection
			defaultValues = convertToMap(item)
		}

		// Use the processed item as default values
		metadata.DefaultValues = defaultValues

		// Extract field dependencies
		metadata.Dependencies = extractFieldDependencies(res.GetFields())

		// Return response
		c.JSON(http.StatusOK, metadata)
	})
}

// convertToMap converts an arbitrary struct to map[string]interface{} using reflection
func convertToMap(item interface{}) map[string]interface{} {
	if item == nil {
		return nil
	}

	result := make(map[string]interface{})

	// Get value via reflection
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	// Only handle structs
	if val.Kind() != reflect.Struct {
		return result
	}

	// Iterate over fields
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Get field name from JSON tag if available
		fieldName := field.Name
		if jsonTag, ok := field.Tag.Lookup("json"); ok {
			parts := utils.SplitTagParts(jsonTag)
			if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}

		// Get field value
		fieldValue := val.Field(i).Interface()

		// Add to result
		result[fieldName] = fieldValue
	}

	return result
}
