package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/dto"
	"github.com/stanxing/refine-gin/pkg/repository"
	"github.com/stanxing/refine-gin/pkg/resource"
)

// GenerateCreateHandler generates a handler for CREATE operations with DTO support
func GenerateCreateHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the DTO
		dtoInstance := dtoProvider.GetCreateDTO()

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

		// Call repository query method to get DB connection
		db := repo.Query(c.Request.Context())

		// Validate relations (if any) - only perform if repository has DB access
		if db != nil && len(res.GetRelations()) > 0 {
			// Validate relations
			if err := resource.ValidateRelations(db, model); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		// Call repository
		createdModel, err := repo.Create(c.Request.Context(), model)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform model to response DTO
		responseDTO, err := dtoProvider.TransformFromModel(createdModel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return result
		c.JSON(http.StatusCreated, gin.H{
			"data": responseDTO,
		})
	}
}
