package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/dto"
	"github.com/suranig/refine-gin/pkg/repository"
	"github.com/suranig/refine-gin/pkg/resource"
)

// RegisterResource registers resource handlers in the Gin router
func RegisterResource(router *gin.RouterGroup, res resource.Resource, repo repository.Repository) {
	// Create default DTO provider if not specified
	dtoProvider := &dto.DefaultDTOProvider{
		Model: res.GetModel(),
	}

	// Register handlers for allowed operations
	if res.HasOperation(resource.OperationList) {
		router.GET("/"+res.GetName(), GenerateListHandler(res, repo))
	}

	if res.HasOperation(resource.OperationRead) {
		router.GET("/"+res.GetName()+"/:id", GenerateGetHandler(res, repo))
	}

	if res.HasOperation(resource.OperationCreate) {
		router.POST("/"+res.GetName(), GenerateCreateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationUpdate) {
		router.PUT("/"+res.GetName()+"/:id", GenerateUpdateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationDelete) {
		router.DELETE("/"+res.GetName()+"/:id", GenerateDeleteHandler(res, repo))
	}
}

// RegisterResourceWithDTO registers resource handlers with custom DTO provider
func RegisterResourceWithDTO(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) {
	// Register handlers for allowed operations
	if res.HasOperation(resource.OperationList) {
		router.GET("/"+res.GetName(), GenerateListHandler(res, repo))
	}

	if res.HasOperation(resource.OperationRead) {
		router.GET("/"+res.GetName()+"/:id", GenerateGetHandler(res, repo))
	}

	if res.HasOperation(resource.OperationCreate) {
		router.POST("/"+res.GetName(), GenerateCreateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationUpdate) {
		router.PUT("/"+res.GetName()+"/:id", GenerateUpdateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationDelete) {
		router.DELETE("/"+res.GetName()+"/:id", GenerateDeleteHandler(res, repo))
	}
}
