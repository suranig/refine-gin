package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/dto"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
	"github.com/suranig/refine-gin/pkg/utils"
)

// BulkCreateRequest is the request structure for creating multiple resources
type BulkCreateRequest struct {
	Values interface{} `json:"values"`
}

// BulkUpdateRequest is the request structure for updating multiple resources
type BulkUpdateRequest struct {
	IDs    interface{} `json:"ids"`
	Values interface{} `json:"values"`
}

// BulkDeleteRequest is the request structure for deleting multiple resources
type BulkDeleteRequest struct {
	IDs interface{} `json:"ids"`
}

// BulkResponse is the common response structure for bulk operations
type BulkResponse struct {
	Data interface{} `json:"data"`
}

// GenerateCreateManyHandler generates a handler for bulk create operations
func GenerateCreateManyHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse request
		var req BulkCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if values is a slice
		if !resource.IsSlice(req.Values) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "values must be an array"})
			return
		}

		// If DTO is provided, transform request data to model
		var modelData interface{}
		var err error
		if dtoProvider != nil {
			modelData, err = dtoProvider.TransformToModel(req.Values)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else {
			modelData = req.Values
		}

		// Get the database connection from repository for validation
		db := repo.Query(c.Request.Context())

		// Validate relations for each item if database is available
		if db != nil && len(res.GetRelations()) > 0 && resource.IsSlice(modelData) {
			// Get the slice from the interface
			slice := reflect.ValueOf(modelData)
			if slice.Kind() == reflect.Ptr {
				slice = slice.Elem()
			}

			// Iterate through each item
			for i := 0; i < slice.Len(); i++ {
				item := slice.Index(i).Interface()

				// Validate relations before saving
				if err := resource.ValidateRelations(db, item); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Item %d: %s", i, err.Error())})
					return
				}
			}
		}

		// Call repository method
		result, err := repo.CreateMany(c, modelData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If DTO is provided, transform response to DTO
		var responseData interface{}
		if dtoProvider != nil {
			responseData, err = dtoProvider.TransformFromModel(result)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			responseData = result
		}

		// Set nocache headers for modification operations
		utils.DisableCaching(c.Writer)

		// Return response
		c.JSON(http.StatusCreated, BulkResponse{Data: responseData})
	}
}

// GenerateUpdateManyHandler generates a handler for bulk update operations
func GenerateUpdateManyHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse request
		var req BulkUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Convert IDs to a slice of interface{}
		// Refine.dev sends IDs as an array or a single value, so we need to handle both cases
		var ids []interface{}
		switch v := req.IDs.(type) {
		case []interface{}:
			ids = v
		case interface{}:
			// Convert to JSON and back to ensure it's a slice
			jsonData, err := json.Marshal(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IDs format"})
				return
			}

			if err := json.Unmarshal(jsonData, &ids); err != nil {
				// If it's not an array, it might be a single value
				ids = []interface{}{v}
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "IDs must be an array or a single value"})
			return
		}

		// If DTO is provided, transform request data to model
		var modelData interface{}
		var err error
		if dtoProvider != nil {
			// For updates, we use the update DTO
			updateDTO := dtoProvider.GetUpdateDTO()

			// Convert values to DTO type
			jsonData, err := json.Marshal(req.Values)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := json.Unmarshal(jsonData, &updateDTO); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Transform to model
			modelData, err = dtoProvider.TransformToModel(updateDTO)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else {
			modelData = req.Values
		}

		// Get the database connection from repository for validation
		db := repo.Query(c.Request.Context())

		// Validate relations if database is available
		// For bulk updates, we only validate that the relation values are valid, not that they exist for each ID
		if db != nil && len(res.GetRelations()) > 0 {
			// Validate relations before save
			if err := resource.ValidateRelations(db, modelData); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		// Call repository method
		count, err := repo.UpdateMany(c, ids, modelData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Set nocache headers for modification operations
		utils.DisableCaching(c.Writer)

		// Return response
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"count": count,
			},
		})
	}
}

// GenerateDeleteManyHandler generates a handler for bulk delete operations
func GenerateDeleteManyHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse request
		var req BulkDeleteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Convert IDs to a slice of interface{}
		var ids []interface{}
		switch v := req.IDs.(type) {
		case []interface{}:
			ids = v
		case interface{}:
			// Convert to JSON and back to ensure it's a slice
			jsonData, err := json.Marshal(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IDs format"})
				return
			}

			if err := json.Unmarshal(jsonData, &ids); err != nil {
				// If it's not an array, it might be a single value
				ids = []interface{}{v}
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "IDs must be an array or a single value"})
			return
		}

		// Call repository method
		count, err := repo.DeleteMany(c, ids)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Set nocache headers for modification operations
		utils.DisableCaching(c.Writer)

		// Return response
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"count": count,
			},
		})
	}
}
