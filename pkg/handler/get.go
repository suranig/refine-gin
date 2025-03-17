package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
)

// GenerateGetHandler generates a handler for READ operations
func GenerateGetHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get ID from URL parameters
		id := c.Param("id")

		// Call repository
		data, err := repo.Get(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}

		// Return result
		c.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	}
}

// GenerateGetHandlerWithParam generates a handler for READ operations with custom ID parameter name
func GenerateGetHandlerWithParam(res resource.Resource, repo repository.Repository, idParamName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get ID from URL parameters using custom parameter name
		id := c.Param(idParamName)

		// Call repository
		data, err := repo.Get(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}

		// Return result
		c.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	}
}
