package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/query"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
)

// GenerateListHandler generates a handler for LIST operations
func GenerateListHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create query options
		options := query.NewQueryOptions(c, res)

		// Call repository
		data, total, err := repo.List(c.Request.Context(), options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return results
		c.JSON(http.StatusOK, gin.H{
			"data":     data,
			"total":    total,
			"page":     options.Page,
			"per_page": options.PerPage,
		})
	}
}
