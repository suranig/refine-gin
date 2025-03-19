package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
)

// GenerateCountHandler generates a handler for COUNT operations
func GenerateCountHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create query options (without pagination)
		options := query.NewQueryOptions(c, res)
		options.DisablePagination = true

		// Call repository count method
		count, err := repo.Count(c.Request.Context(), options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return count
		c.JSON(http.StatusOK, gin.H{
			"count": count,
		})
	}
}
