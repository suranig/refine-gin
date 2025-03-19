package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/dto"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
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
