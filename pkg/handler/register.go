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

	// Określ nazwę parametru URL dla identyfikatora (domyślnie "id")
	idParamName := "id"

	// Register handlers for allowed operations
	if res.HasOperation(resource.OperationList) {
		router.GET("/"+res.GetName(), GenerateListHandler(res, repo))
	}

	if res.HasOperation(resource.OperationRead) {
		router.GET("/"+res.GetName()+"/:"+idParamName, GenerateGetHandler(res, repo))
	}

	if res.HasOperation(resource.OperationCreate) {
		router.POST("/"+res.GetName(), GenerateCreateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationUpdate) {
		router.PUT("/"+res.GetName()+"/:"+idParamName, GenerateUpdateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationDelete) {
		router.DELETE("/"+res.GetName()+"/:"+idParamName, GenerateDeleteHandler(res, repo))
	}
}

// RegisterResourceWithDTO registers resource handlers with custom DTO provider
func RegisterResourceWithDTO(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) {
	// Określ nazwę parametru URL dla identyfikatora (domyślnie "id")
	idParamName := "id"

	// Register handlers for allowed operations
	if res.HasOperation(resource.OperationList) {
		router.GET("/"+res.GetName(), GenerateListHandler(res, repo))
	}

	if res.HasOperation(resource.OperationRead) {
		router.GET("/"+res.GetName()+"/:"+idParamName, GenerateGetHandler(res, repo))
	}

	if res.HasOperation(resource.OperationCreate) {
		router.POST("/"+res.GetName(), GenerateCreateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationUpdate) {
		router.PUT("/"+res.GetName()+"/:"+idParamName, GenerateUpdateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationDelete) {
		router.DELETE("/"+res.GetName()+"/:"+idParamName, GenerateDeleteHandler(res, repo))
	}
}

// RegisterResourceWithOptions registers resource handlers with custom options
func RegisterResourceWithOptions(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, options RegisterOptions) {
	// Create default DTO provider if not specified
	dtoProvider := options.DTOProvider
	if dtoProvider == nil {
		dtoProvider = &dto.DefaultDTOProvider{
			Model: res.GetModel(),
		}
	}

	// Określ nazwę parametru URL dla identyfikatora
	idParamName := options.IDParamName
	if idParamName == "" {
		idParamName = "id"
	}

	// Register handlers for allowed operations
	if res.HasOperation(resource.OperationList) {
		router.GET("/"+res.GetName(), GenerateListHandler(res, repo))
	}

	if res.HasOperation(resource.OperationRead) {
		router.GET("/"+res.GetName()+"/:"+idParamName, GenerateGetHandlerWithParam(res, repo, idParamName))
	}

	if res.HasOperation(resource.OperationCreate) {
		router.POST("/"+res.GetName(), GenerateCreateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationUpdate) {
		router.PUT("/"+res.GetName()+"/:"+idParamName, GenerateUpdateHandlerWithParam(res, repo, dtoProvider, idParamName))
	}

	if res.HasOperation(resource.OperationDelete) {
		router.DELETE("/"+res.GetName()+"/:"+idParamName, GenerateDeleteHandlerWithParam(res, repo, idParamName))
	}
}

// RegisterOptions zawiera opcje rejestracji zasobu
type RegisterOptions struct {
	DTOProvider dto.DTOProvider // Dostawca DTO (opcjonalny)
	IDParamName string          // Nazwa parametru URL dla identyfikatora (domyślnie "id")
}
