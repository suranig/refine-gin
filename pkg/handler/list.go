package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/dto"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
)

// GenerateListHandler generates a handler for LIST operations
func GenerateListHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
	// Use default DTO provider if none is specified
	dtoProvider := &dto.DefaultDTOProvider{
		Model: res.GetModel(),
	}

	return generateListHandlerWithDTO(res, repo, dtoProvider)
}

// GenerateListHandlerWithDTO generates a handler for LIST operations with DTO transformation
func GenerateListHandlerWithDTO(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return generateListHandlerWithDTO(res, repo, dtoProvider)
}

// Helper function to avoid code duplication
func generateListHandlerWithDTO(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create query options
		options := query.ParseQueryOptions(c, res)

		// Call repository
		data, total, err := repo.List(c.Request.Context(), options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform models to DTOs if we have an array
		if items, ok := data.([]interface{}); ok && dtoProvider != nil {
			dtoItems := make([]interface{}, 0, len(items))
			for _, item := range items {
				dtoItem, err := dtoProvider.TransformFromModel(item)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error transforming data: " + err.Error()})
					return
				}
				dtoItems = append(dtoItems, dtoItem)
			}
			data = dtoItems
		}

		// Return results in Refine.dev compatible format
		c.JSON(http.StatusOK, gin.H{
			"data":     data,
			"total":    total,
			"current":  options.Page,    // Refine uses 'current' not 'page'
			"pageSize": options.PerPage, // Correct parameter name for Refine.dev
		})
	}
}
