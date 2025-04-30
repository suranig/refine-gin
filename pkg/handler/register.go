package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/stanxing/refine-gin/pkg/dto"
	"github.com/stanxing/refine-gin/pkg/middleware"
	"github.com/stanxing/refine-gin/pkg/repository"
	"github.com/stanxing/refine-gin/pkg/resource"
)

// RegisterResource registers resource handlers in the Gin router
func RegisterResource(router *gin.RouterGroup, res resource.Resource, repo repository.Repository) {
	// Register resource to registry
	resource.RegisterToRegistry(res)

	// Create default DTO provider if not specified
	dtoProvider := &dto.DefaultDTOProvider{
		Model: res.GetModel(),
	}

	// Określ nazwę parametru URL dla identyfikatora (domyślnie "id")
	idParamName := "id"

	// Register OPTIONS handler for metadata
	router.OPTIONS("/"+res.GetName(), GenerateOptionsHandler(res))

	// Register handlers for allowed operations
	if res.HasOperation(resource.OperationList) {
		router.GET("/"+res.GetName(), GenerateListHandlerWithDTO(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationRead) {
		router.GET("/"+res.GetName()+"/:"+idParamName, GenerateGetHandlerWithParamAndDTO(res, repo, idParamName, dtoProvider))
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

	// Register count handler if the operation is allowed
	if res.HasOperation(resource.OperationCount) {
		router.GET("/"+res.GetName()+"/count", GenerateCountHandler(res, repo))
	}
}

// RegisterResourceWithDTO registers resource handlers with custom DTO provider
func RegisterResourceWithDTO(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) {
	// Register resource to registry
	resource.RegisterToRegistry(res)

	// Use default options with DTO provider
	opts := resource.DefaultOptions()

	// Create resource router with naming convention middleware
	resourceRouter := router.Group("/"+res.GetName(), middleware.NamingConventionMiddleware(opts.NamingConvention))

	// Register OPTIONS handler for metadata
	resourceRouter.OPTIONS("", GenerateOptionsHandler(res))

	// Register handlers for allowed operations
	if res.HasOperation(resource.OperationList) {
		resourceRouter.GET("", GenerateListHandlerWithDTO(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationCreate) {
		resourceRouter.POST("", GenerateCreateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationRead) {
		resourceRouter.GET("/:id", GenerateGetHandlerWithDTO(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationUpdate) {
		resourceRouter.PUT("/:id", GenerateUpdateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationDelete) {
		resourceRouter.DELETE("/:id", GenerateDeleteHandler(res, repo))
	}

	if res.HasOperation(resource.OperationCount) {
		resourceRouter.GET("/count", GenerateCountHandler(res, repo))
	}
}

// RegisterResourceWithOptions registers a resource with customizable options
func RegisterResourceWithOptions(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, opts resource.Options) {
	// Register resource to registry
	resource.RegisterToRegistry(res)

	// Extract options
	idParamName := "id"
	if paramName, ok := opts.GetQueryOption("IDParamName").(string); ok && paramName != "" {
		idParamName = paramName
	}

	// Create default DTO provider if not specified
	dtoProvider := &dto.DefaultDTOProvider{
		Model: res.GetModel(),
	}

	// Create resource router with naming convention middleware
	resourceRouter := router.Group("/"+res.GetName(),
		middleware.NamingConventionMiddleware(opts.NamingConvention),
	)

	// Register handlers for allowed operations
	if res.HasOperation(resource.OperationList) {
		resourceRouter.GET("", GenerateListHandler(res, repo))
	}

	if res.HasOperation(resource.OperationCreate) {
		// Dla operacji modyfikujących dane, wyłącz cache
		resourceRouter.POST("", middleware.NoCacheMiddleware(), GenerateCreateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationRead) {
		resourceRouter.GET("/:"+idParamName, GenerateGetHandlerWithParam(res, repo, idParamName))
	}

	if res.HasOperation(resource.OperationUpdate) {
		// Check if we have a custom ID field
		hasCustomID := res.GetIDFieldName() != "ID" && res.GetIDFieldName() != "id"
		if hasCustomID {
			// Use custom update handler for resources with non-standard ID fields
			resourceRouter.PUT("/:"+idParamName, middleware.NoCacheMiddleware(), GenerateCustomUpdateHandler(res, repo, idParamName))
		} else {
			// Use standard update handler
			resourceRouter.PUT("/:"+idParamName, middleware.NoCacheMiddleware(), GenerateUpdateHandlerWithParam(res, repo, dtoProvider, idParamName))
		}
	}

	if res.HasOperation(resource.OperationDelete) {
		resourceRouter.DELETE("/:"+idParamName, middleware.NoCacheMiddleware(), GenerateDeleteHandlerWithParam(res, repo, idParamName))
	}

	if res.HasOperation(resource.OperationCount) {
		resourceRouter.GET("/count", GenerateCountHandler(res, repo))
	}
}

// RegisterResourceForRefine registers resource handlers optimized for Refine.dev
// The idParamName parameter allows specifying a custom ID parameter name for the resource
// This is useful for resources that use a non-standard ID field (not 'id')
func RegisterResourceForRefine(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, idParamName string) {
	// Register resource to registry
	resource.RegisterToRegistry(res)

	// Create default DTO provider if not specified
	dtoProvider := &dto.DefaultDTOProvider{
		Model: res.GetModel(),
	}

	// If idParamName is empty, use default "id"
	if idParamName == "" {
		idParamName = "id"
	}

	cacheConfig := middleware.DefaultCacheConfig()
	// Add OPTIONS method to cache config
	cacheConfig.Methods = append(cacheConfig.Methods, "OPTIONS")

	// Create resource router with naming convention middleware - default to camelCase for Refine.dev
	resourceRouter := router.Group("/"+res.GetName(),
		middleware.NamingConventionMiddleware(resource.DefaultOptions().NamingConvention),
		middleware.CacheByResource(res.GetName(), cacheConfig), // Dodaj middleware cache dla całego zasobu
	)

	// Register OPTIONS handler for resource metadata
	resourceRouter.OPTIONS("", GenerateOptionsHandler(res))

	// Register handlers for allowed operations
	if res.HasOperation(resource.OperationList) {
		resourceRouter.GET("", GenerateListHandlerWithDTO(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationCreate) {
		// Operacje POST, PUT, DELETE nie powinny być cachowane
		resourceRouter.POST("", middleware.NoCacheMiddleware(), GenerateCreateHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationRead) {
		resourceRouter.GET("/:"+idParamName, GenerateGetHandlerWithParamAndDTO(res, repo, idParamName, dtoProvider))
	}

	if res.HasOperation(resource.OperationUpdate) {
		// Check if we have a custom ID field
		hasCustomID := res.GetIDFieldName() != "ID" && res.GetIDFieldName() != "id"
		if hasCustomID {
			// Use custom update handler for resources with non-standard ID fields
			resourceRouter.PUT("/:"+idParamName, middleware.NoCacheMiddleware(), GenerateCustomUpdateHandler(res, repo, idParamName))
		} else {
			// Use standard update handler
			resourceRouter.PUT("/:"+idParamName, middleware.NoCacheMiddleware(), GenerateUpdateHandlerWithParam(res, repo, dtoProvider, idParamName))
		}
	}

	if res.HasOperation(resource.OperationDelete) {
		resourceRouter.DELETE("/:"+idParamName, middleware.NoCacheMiddleware(), GenerateDeleteHandlerWithParam(res, repo, idParamName))
	}

	if res.HasOperation(resource.OperationCount) {
		resourceRouter.GET("/count", GenerateCountHandler(res, repo))
	}

	// Register handlers for bulk operations
	if res.HasOperation(resource.OperationCreateMany) {
		// POST /resources/batch for creating multiple resources
		resourceRouter.POST("/batch", middleware.NoCacheMiddleware(), GenerateCreateManyHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationUpdateMany) {
		// PUT /resources/batch for updating multiple resources
		resourceRouter.PUT("/batch", middleware.NoCacheMiddleware(), GenerateUpdateManyHandler(res, repo, dtoProvider))
	}

	if res.HasOperation(resource.OperationDeleteMany) {
		// DELETE /resources/batch for deleting multiple resources
		resourceRouter.DELETE("/batch", middleware.NoCacheMiddleware(), GenerateDeleteManyHandler(res, repo))
	}
}

// RegisterOptions zawiera opcje rejestracji zasobu
type RegisterOptions struct {
	DTOProvider dto.DTOProvider // Dostawca DTO (opcjonalny)
	IDParamName string          // Nazwa parametru URL dla identyfikatora (domyślnie "id")
}

// RegisterOptionsToResourceOptions converts RegisterOptions to resource.Options
func RegisterOptionsToResourceOptions(regOpts RegisterOptions) resource.Options {
	opts := resource.DefaultOptions()

	// Set IDParamName as a query option
	if regOpts.IDParamName != "" {
		opts = opts.WithQueryOption("IDParamName", regOpts.IDParamName)
	}

	return opts
}
