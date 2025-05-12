package handler

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/dto"
	"github.com/stanxing/refine-gin/pkg/query"
	"github.com/stanxing/refine-gin/pkg/repository"
	"github.com/stanxing/refine-gin/pkg/resource"
	"github.com/stanxing/refine-gin/pkg/utils"
)

// GenerateOwnerListHandler creates a handler for listing items with owner filtering
func GenerateOwnerListHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query options from the request
		options := query.ParseQueryOptions(c, res)

		// Generate ETag for cache validation
		etag := utils.GenerateQueryETag(c.Request.URL.RawQuery)
		ifNoneMatch := c.GetHeader("If-None-Match")

		// Check if client's cached version is still valid
		if utils.IsETagMatch(etag, ifNoneMatch) {
			c.Status(http.StatusNotModified)
			return
		}

		// Get data from repository (owner filtering is handled in repository)
		data, total, err := repo.List(c.Request.Context(), options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform to DTOs if data is a slice
		if reflect.TypeOf(data).Kind() == reflect.Slice {
			v := reflect.ValueOf(data)
			dtoItems := make([]interface{}, 0, v.Len())
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i).Interface()
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
			"data":  data,
			"total": total,
			"meta": gin.H{
				"page":     options.Page,
				"pageSize": options.PerPage,
			},
		})
	}
}

// GenerateOwnerCountHandler creates a handler for counting items with owner filtering
func GenerateOwnerCountHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query options from the request
		options := query.ParseQueryOptions(c, res)

		// Generate ETag for cache validation
		etag := utils.GenerateQueryETag(c.Request.URL.RawQuery)
		ifNoneMatch := c.GetHeader("If-None-Match")

		// Check if client's cached version is still valid
		if utils.IsETagMatch(etag, ifNoneMatch) {
			c.Status(http.StatusNotModified)
			return
		}

		// Get count from repository (owner filtering is handled in repository)
		count, err := repo.Count(c.Request.Context(), options)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return count in Refine.dev compatible format
		c.JSON(http.StatusOK, gin.H{
			"data": count,
		})
	}
}

// GenerateOwnerCreateHandler creates a handler for creating an item with owner
func GenerateOwnerCreateHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the model
		model := utils.CreateNewModelInstance(res.GetModel())
		if model == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create model instance"})
			return
		}

		// Bind request to model
		if err := c.ShouldBindJSON(model); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create in repository (owner field will be set automatically)
		created, err := repo.Create(c.Request.Context(), model)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform to DTO
		dtoData, err := dtoProvider.TransformFromModel(created)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return created resource
		c.JSON(http.StatusCreated, gin.H{
			"data": dtoData,
		})
	}
}

// GenerateOwnerGetHandler creates a handler for getting an item with owner verification
func GenerateOwnerGetHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider, idParamName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get ID from URL
		id := c.Param(idParamName)
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing resource ID"})
			return
		}

		// Generate ETag for cache validation
		etag := utils.GenerateETag(id)
		ifNoneMatch := c.GetHeader("If-None-Match")

		// Check if client's cached version is still valid
		if utils.IsETagMatch(etag, ifNoneMatch) {
			c.Status(http.StatusNotModified)
			return
		}

		// Get the resource (ownership verification happens in repository)
		data, err := repo.Get(c.Request.Context(), id)
		if err != nil {
			// Handle specific errors
			if err == repository.ErrOwnerMismatch {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this resource"})
				return
			}
			// Check for not found error
			if err.Error() == "record not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
				return
			}
			// Other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform to DTO
		dtoData, err := dtoProvider.TransformFromModel(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the resource
		c.JSON(http.StatusOK, gin.H{
			"data": dtoData,
		})
	}
}

// GenerateOwnerUpdateHandler creates a handler for updating an item with owner verification
func GenerateOwnerUpdateHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider, idParamName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get ID from URL
		id := c.Param(idParamName)
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing resource ID"})
			return
		}

		// Create a new instance of the model
		model := utils.CreateNewModelInstance(res.GetModel())
		if model == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create model instance"})
			return
		}

		// Bind request to model
		if err := c.ShouldBindJSON(model); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Filter out read-only fields
		model = resource.FilterOutReadOnlyFields(model, res)

		// Update in repository (ownership verification happens in repository)
		updated, err := repo.Update(c.Request.Context(), id, model)
		if err != nil {
			// Handle specific errors
			if err == repository.ErrOwnerMismatch {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this resource"})
				return
			}
			// Check for not found error
			if err.Error() == "record not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
				return
			}
			// Other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform to DTO
		dtoData, err := dtoProvider.TransformFromModel(updated)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return updated resource
		c.JSON(http.StatusOK, gin.H{
			"data": dtoData,
		})
	}
}

// GenerateOwnerDeleteHandler creates a handler for deleting an item with owner verification
func GenerateOwnerDeleteHandler(res resource.Resource, repo repository.Repository, idParamName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get ID from URL
		id := c.Param(idParamName)
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing resource ID"})
			return
		}

		// Delete from repository (ownership verification happens in repository)
		err := repo.Delete(c.Request.Context(), id)
		if err != nil {
			// Handle specific errors
			if err == repository.ErrOwnerMismatch {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this resource"})
				return
			}
			// Check for not found error
			if err.Error() == "record not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
				return
			}
			// Other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return success
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{"success": true},
		})
	}
}

// GenerateOwnerCreateManyHandler creates a handler for creating multiple items with owner
func GenerateOwnerCreateManyHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new slice of the model type
		slice := utils.CreateNewSliceOfModel(res.GetModel())
		if slice == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create model slice"})
			return
		}

		// Bind request to slice
		if err := c.ShouldBindJSON(slice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create many in repository (owner field will be set automatically)
		created, err := repo.CreateMany(c.Request.Context(), slice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Transform to DTOs
		dtoData, err := dtoProvider.TransformFromModel(created)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return created resources
		c.JSON(http.StatusCreated, gin.H{
			"data": dtoData,
		})
	}
}

// GenerateOwnerUpdateManyHandler creates a handler for updating multiple items with owner verification
func GenerateOwnerUpdateManyHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Define request structure
		var req struct {
			IDs  []interface{} `json:"ids"`
			Data interface{}   `json:"data"`
		}

		// Bind request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Filter out read-only fields from the update data
		filteredData := resource.FilterOutReadOnlyFields(req.Data, res)

		// Update many in repository (ownership verification happens in repository)
		affected, err := repo.UpdateMany(c.Request.Context(), req.IDs, filteredData)
		if err != nil {
			// Handle specific errors
			if err == repository.ErrOwnerMismatch {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update some resources"})
				return
			}
			// Other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return affected count
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{"affected": affected},
		})
	}
}

// GenerateOwnerDeleteManyHandler creates a handler for deleting multiple items with owner verification
func GenerateOwnerDeleteManyHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Define request structure
		var req struct {
			IDs []interface{} `json:"ids"`
		}

		// Bind request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Delete many from repository (ownership verification happens in repository)
		affected, err := repo.DeleteMany(c.Request.Context(), req.IDs)
		if err != nil {
			// Handle specific errors
			if err == repository.ErrOwnerMismatch {
				c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete some resources"})
				return
			}
			// Other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return affected count
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{"affected": affected},
		})
	}
}
