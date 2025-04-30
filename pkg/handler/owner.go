package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/dto"
	"github.com/stanxing/refine-gin/pkg/repository"
	"github.com/stanxing/refine-gin/pkg/resource"
)

// RegisterOwnerResource registers all the owner-specific resource handlers for a given resource
func RegisterOwnerResource(group *gin.RouterGroup, res resource.OwnerResource, repo repository.Repository) {
	// Create DTO provider
	dtoProvider := &dto.DefaultDTOProvider{
		Model: res.GetModel(),
	}

	// Resource name and base path
	resourceName := res.GetName()

	// Register list handler
	if res.HasOperation(resource.OperationList) {
		group.GET("/"+resourceName, GenerateOwnerListHandler(res, repo, dtoProvider))
	}

	// Register get handler
	if res.HasOperation(resource.OperationRead) {
		group.GET("/"+resourceName+"/:id", GenerateOwnerGetHandler(res, repo, dtoProvider, "id"))
	}

	// Register create handler
	if res.HasOperation(resource.OperationCreate) {
		group.POST("/"+resourceName, GenerateOwnerCreateHandler(res, repo, dtoProvider))
	}

	// Register update handler
	if res.HasOperation(resource.OperationUpdate) {
		group.PUT("/"+resourceName+"/:id", GenerateOwnerUpdateHandler(res, repo, dtoProvider, "id"))
	}

	// Register delete handler
	if res.HasOperation(resource.OperationDelete) {
		group.DELETE("/"+resourceName+"/:id", GenerateOwnerDeleteHandler(res, repo, "id"))
	}

	// Register count handler
	group.GET("/"+resourceName+"/count", GenerateOwnerCountHandler(res, repo))

	// Register batch handlers
	if res.HasOperation(resource.OperationCreateMany) {
		group.POST("/"+resourceName+"/batch", GenerateOwnerCreateManyHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationUpdateMany) {
		group.PUT("/"+resourceName+"/batch", GenerateOwnerUpdateManyHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationDeleteMany) {
		group.DELETE("/"+resourceName+"/batch", GenerateOwnerDeleteManyHandler(res, repo))
	}
}
