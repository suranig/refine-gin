package swagger

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
	"github.com/suranig/refine-gin/pkg/utils"
)

// DefaultSwaggerInfo returns default Swagger metadata
func DefaultSwaggerInfo() SwaggerInfo {
	return SwaggerInfo{
		Title:       "Refine-Gin API",
		Description: "API documentation for Refine-Gin",
		Version:     "1.0.0",
		BasePath:    "/api",
		Schemes:     []string{"http", "https"},
	}
}

// GenerateOpenAPI generates OpenAPI documentation from registered resources
func GenerateOpenAPI(resources []resource.Resource, info SwaggerInfo) *OpenAPI {
	openAPI := &OpenAPI{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       info.Title,
			Description: info.Description,
			Version:     info.Version,
		},
		Servers: []Server{
			{
				URL:         info.BasePath,
				Description: "API Server",
			},
		},
		Paths: make(map[string]PathItem),
		Components: Components{
			Schemas:         make(map[string]Schema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
		Tags: []Tag{},
	}

	// Add JWT security scheme
	openAPI.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT token authorization",
	}

	// Process each resource
	for _, res := range resources {
		// Add tag for the resource
		openAPI.Tags = append(openAPI.Tags, Tag{
			Name:        res.GetName(),
			Description: fmt.Sprintf("%s resource", res.GetName()),
		})

		// Add schema for the resource model
		modelSchema := generateModelSchema(res)
		openAPI.Components.Schemas[res.GetName()] = modelSchema

		// Generate paths for the resource
		generateResourcePaths(openAPI, res)
	}

	// Merge custom endpoints registered via RegisterCustomEndpoint
	for _, ce := range GetCustomEndpoints() {
		if existing, exists := openAPI.Paths[ce.Path]; exists {
			existing[ce.Method] = ce.Operation
			openAPI.Paths[ce.Path] = existing
		} else {
			openAPI.Paths[ce.Path] = PathItem{ce.Method: ce.Operation}
		}
	}

	return openAPI
}

// SwaggerHandler returns a Gin handler that serves the Swagger UI
func SwaggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, swaggerHTML)
	}
}

// RegisterSwagger registers Swagger routes in the Gin router
func RegisterSwagger(router *gin.RouterGroup, resources []resource.Resource, info SwaggerInfo) {
	// Create OpenAPI spec
	openAPI := GenerateOpenAPI(resources, info)

	// Register Swagger UI handler
	router.GET("/swagger", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, swaggerHTML)
	})

	// Register OpenAPI spec handler
	router.GET("/swagger.json", func(c *gin.Context) {
		c.JSON(200, openAPI)
	})
}

// Helper functions

// generateModelSchema generates a schema for a resource model
func generateModelSchema(res resource.Resource) Schema {
	schema := Schema{
		Type:       "object",
		Properties: make(map[string]Schema),
		Required:   []string{},
	}

	// Process fields
	for _, field := range res.GetFields() {
		fieldSchema := fieldToSchema(field)
		schema.Properties[field.Name] = fieldSchema

		if field.Required {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	return schema
}

// fieldToSchema converts a resource field to an OpenAPI schema
func fieldToSchema(field resource.Field) Schema {
	schema := Schema{}

	typeMapping := utils.GetTypeMapping(field.Type)

	switch typeMapping.Category {
	case utils.TypeString:
		schema.Type = "string"
	case utils.TypeInteger:
		schema.Type = "integer"
		schema.Format = typeMapping.Format
	case utils.TypeNumber:
		schema.Type = "number"
		schema.Format = typeMapping.Format
	case utils.TypeBoolean:
		schema.Type = "boolean"
	case utils.TypeDateTime:
		schema.Type = "string"
		schema.Format = "date-time"
	case utils.TypeArray:
		schema.Type = "array"
		schema.Items = &Schema{}

		elementTypeMapping := utils.GetTypeMapping(typeMapping.Format)
		if elementTypeMapping.IsPrimitive {
			schema.Items.Type = string(elementTypeMapping.Category)
			if elementTypeMapping.Format != "" {
				schema.Items.Format = elementTypeMapping.Format
			}
		} else {
			schema.Items.Ref = "#/components/schemas/" + typeMapping.Format
		}
	default:
		schema.Ref = "#/components/schemas/" + typeMapping.Format
	}

	return schema
}

// generateResourcePaths generates path entries for a resource
func generateResourcePaths(openAPI *OpenAPI, res resource.Resource) {
	// Generate list endpoint
	if res.HasOperation(resource.OperationList) {
		listPath := "/" + res.GetName()
		openAPI.Paths[listPath] = PathItem{
			"get": Operation{
				Summary:     fmt.Sprintf("List %s", res.GetName()),
				Description: fmt.Sprintf("Get a list of %s", res.GetName()),
				OperationID: fmt.Sprintf("list%s", capitalize(res.GetName())),
				Tags:        []string{res.GetName()},
				Parameters:  generateListParameters(),
				Responses: map[string]Response{
					"200": {
						Description: "Successful operation",
						Content: map[string]MediaType{
							"application/json": {
								Schema: Schema{
									Type: "object",
									Properties: map[string]Schema{
										"data": {
											Type: "array",
											Items: &Schema{
												Ref: "#/components/schemas/" + res.GetName(),
											},
										},
										"total": {
											Type: "integer",
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}

	// Generate OPTIONS endpoint for resource metadata (always available)
	optionsPath := "/" + res.GetName()
	if openAPI.Paths[optionsPath] == nil {
		openAPI.Paths[optionsPath] = PathItem{}
	}
	openAPI.Paths[optionsPath]["options"] = Operation{
		Summary:     fmt.Sprintf("Get %s metadata", res.GetName()),
		Description: fmt.Sprintf("Returns metadata for the %s resource including fields, operations, and configuration", res.GetName()),
		OperationID: fmt.Sprintf("options%s", capitalize(res.GetName())),
		Tags:        []string{res.GetName()},
		Responses: map[string]Response{
			"200": {
				Description: "Resource metadata",
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Type: "object",
							Properties: map[string]Schema{
								"name": {
									Type: "string",
								},
								"label": {
									Type: "string",
								},
								"icon": {
									Type: "string",
								},
								"operations": {
									Type: "array",
									Items: &Schema{
										Type: "string",
									},
								},
								"fields": {
									Type: "array",
									Items: &Schema{
										Type: "object",
										Properties: map[string]Schema{
											"name": {
												Type: "string",
											},
											"type": {
												Type: "string",
											},
											"required": {
												Type: "boolean",
											},
											"unique": {
												Type: "boolean",
											},
											"filterable": {
												Type: "boolean",
											},
											"sortable": {
												Type: "boolean",
											},
											"searchable": {
												Type: "boolean",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"304": {
				Description: "Not Modified (when using If-None-Match header with valid ETag)",
			},
		},
	}

	// Generate get endpoint
	if res.HasOperation(resource.OperationRead) {
		getPath := fmt.Sprintf("/%s/{id}", res.GetName())
		openAPI.Paths[getPath] = PathItem{
			"get": Operation{
				Summary:     fmt.Sprintf("Get %s by ID", res.GetName()),
				Description: fmt.Sprintf("Get a single %s by ID", res.GetName()),
				OperationID: fmt.Sprintf("get%s", capitalize(res.GetName())),
				Tags:        []string{res.GetName()},
				Parameters: []Parameter{
					{
						Name:        "id",
						In:          "path",
						Description: "ID of the resource",
						Required:    true,
						Schema: Schema{
							Type: "string",
						},
					},
				},
				Responses: map[string]Response{
					"200": {
						Description: "Successful operation",
						Content: map[string]MediaType{
							"application/json": {
								Schema: Schema{
									Type: "object",
									Properties: map[string]Schema{
										"data": {
											Ref: "#/components/schemas/" + res.GetName(),
										},
									},
								},
							},
						},
					},
					"404": {
						Description: "Resource not found",
					},
				},
			},
		}
	}

	// Generate create endpoint
	if res.HasOperation(resource.OperationCreate) {
		createPath := "/" + res.GetName()
		openAPI.Paths[createPath]["post"] = Operation{
			Summary:     fmt.Sprintf("Create %s", res.GetName()),
			Description: fmt.Sprintf("Create a new %s", res.GetName()),
			OperationID: fmt.Sprintf("create%s", capitalize(res.GetName())),
			Tags:        []string{res.GetName()},
			RequestBody: &RequestBody{
				Description: fmt.Sprintf("%s object to be created", res.GetName()),
				Required:    true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Ref: "#/components/schemas/" + res.GetName(),
						},
					},
				},
			},
			Responses: map[string]Response{
				"201": {
					Description: "Resource created",
					Content: map[string]MediaType{
						"application/json": {
							Schema: Schema{
								Type: "object",
								Properties: map[string]Schema{
									"data": {
										Ref: "#/components/schemas/" + res.GetName(),
									},
								},
							},
						},
					},
				},
				"400": {
					Description: "Invalid input",
				},
			},
		}
	}

	// Generate update endpoint
	if res.HasOperation(resource.OperationUpdate) {
		updatePath := fmt.Sprintf("/%s/{id}", res.GetName())
		openAPI.Paths[updatePath]["put"] = Operation{
			Summary:     fmt.Sprintf("Update %s", res.GetName()),
			Description: fmt.Sprintf("Update an existing %s", res.GetName()),
			OperationID: fmt.Sprintf("update%s", capitalize(res.GetName())),
			Tags:        []string{res.GetName()},
			Parameters: []Parameter{
				{
					Name:        "id",
					In:          "path",
					Description: "ID of the resource",
					Required:    true,
					Schema: Schema{
						Type: "string",
					},
				},
			},
			RequestBody: &RequestBody{
				Description: fmt.Sprintf("%s object to be updated", res.GetName()),
				Required:    true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: Schema{
							Ref: "#/components/schemas/" + res.GetName(),
						},
					},
				},
			},
			Responses: map[string]Response{
				"200": {
					Description: "Resource updated",
					Content: map[string]MediaType{
						"application/json": {
							Schema: Schema{
								Type: "object",
								Properties: map[string]Schema{
									"data": {
										Ref: "#/components/schemas/" + res.GetName(),
									},
								},
							},
						},
					},
				},
				"400": {
					Description: "Invalid input",
				},
				"404": {
					Description: "Resource not found",
				},
			},
		}
	}

	// Generate delete endpoint
	if res.HasOperation(resource.OperationDelete) {
		deletePath := fmt.Sprintf("/%s/{id}", res.GetName())
		openAPI.Paths[deletePath]["delete"] = Operation{
			Summary:     fmt.Sprintf("Delete %s", res.GetName()),
			Description: fmt.Sprintf("Delete an existing %s", res.GetName()),
			OperationID: fmt.Sprintf("delete%s", capitalize(res.GetName())),
			Tags:        []string{res.GetName()},
			Parameters: []Parameter{
				{
					Name:        "id",
					In:          "path",
					Description: "ID of the resource",
					Required:    true,
					Schema: Schema{
						Type: "string",
					},
				},
			},
			Responses: map[string]Response{
				"204": {
					Description: "Resource deleted",
				},
				"404": {
					Description: "Resource not found",
				},
			},
		}
	}

	// Generate bulk endpoints if supported
	if res.HasOperation(resource.OperationCreateMany) {
		bulkCreatePath := fmt.Sprintf("/%s/batch", res.GetName())
		openAPI.Paths[bulkCreatePath] = PathItem{
			"post": Operation{
				Summary:     fmt.Sprintf("Bulk create %s", res.GetName()),
				Description: fmt.Sprintf("Create multiple %s at once", res.GetName()),
				OperationID: fmt.Sprintf("bulkCreate%s", capitalize(res.GetName())),
				Tags:        []string{res.GetName()},
				RequestBody: &RequestBody{
					Description: fmt.Sprintf("Array of %s objects to be created", res.GetName()),
					Required:    true,
					Content: map[string]MediaType{
						"application/json": {
							Schema: Schema{
								Type: "object",
								Properties: map[string]Schema{
									"data": {
										Type: "array",
										Items: &Schema{
											Ref: "#/components/schemas/" + res.GetName(),
										},
									},
								},
							},
						},
					},
				},
				Responses: map[string]Response{
					"201": {
						Description: "Resources created",
						Content: map[string]MediaType{
							"application/json": {
								Schema: Schema{
									Type: "object",
									Properties: map[string]Schema{
										"data": {
											Type: "array",
											Items: &Schema{
												Ref: "#/components/schemas/" + res.GetName(),
											},
										},
									},
								},
							},
						},
					},
					"400": {
						Description: "Invalid input",
					},
				},
			},
		}
	}
}

// generateListParameters generates standard parameters for list endpoints
func generateListParameters() []Parameter {
	return []Parameter{
		{
			Name:        "page",
			In:          "query",
			Description: "Page number",
			Required:    false,
			Schema: Schema{
				Type: "integer",
			},
		},
		{
			Name:        "perPage",
			In:          "query",
			Description: "Number of items per page",
			Required:    false,
			Schema: Schema{
				Type: "integer",
			},
		},
		{
			Name:        "sort",
			In:          "query",
			Description: "Sort field",
			Required:    false,
			Schema: Schema{
				Type: "string",
			},
		},
		{
			Name:        "order",
			In:          "query",
			Description: "Sort order (asc or desc)",
			Required:    false,
			Schema: Schema{
				Type: "string",
				Enum: []interface{}{"asc", "desc"},
			},
		},
		{
			Name:        "search",
			In:          "query",
			Description: "Search term",
			Required:    false,
			Schema: Schema{
				Type: "string",
			},
		},
	}
}

// capitalize capitalizes the first letter of a string
func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Swagger UI HTML template
const swaggerHTML = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui.css" />
    <style>
      html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
      *, *:before, *:after { box-sizing: inherit; }
      body { margin: 0; background: #fafafa; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-bundle.js" charset="UTF-8"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
    <script>
      window.onload = function() {
        const ui = SwaggerUIBundle({
          url: "./swagger.json",
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ]
        });
      }
    </script>
  </body>
</html>
`
