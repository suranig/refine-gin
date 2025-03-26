package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/dto"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
)

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

		// Get the database connection from repository
		db := repo.Query(c.Request.Context())

		// Validate relations (if any) - only perform if repository has DB access
		if db != nil && len(res.GetRelations()) > 0 {
			// Validate relations
			if err := resource.ValidateRelations(c.Request.Context(), res, model, db); err != nil {
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

		// Get the database connection from repository
		db := repo.Query(c.Request.Context())

		// Validate relations (if any) - only perform if repository has DB access
		if db != nil && len(res.GetRelations()) > 0 {
			// Validate relations
			if err := resource.ValidateRelations(c.Request.Context(), res, model, db); err != nil {
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
