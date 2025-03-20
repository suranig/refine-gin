package swagger

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/suranig/refine-gin/pkg/resource"
)

// SwaggerInfo contains metadata for the Swagger documentation
type SwaggerInfo struct {
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	License     *License
	Contact     *Contact
}

// License information
type License struct {
	Name string
	URL  string
}

// Contact information
type Contact struct {
	Name  string
	URL   string
	Email string
}

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

// OpenAPI represents an OpenAPI 3.0 document
type OpenAPI struct {
	OpenAPI    string                `json:"openapi"`
	Info       Info                  `json:"info"`
	Servers    []Server              `json:"servers"`
	Paths      map[string]PathItem   `json:"paths"`
	Components Components            `json:"components"`
	Tags       []Tag                 `json:"tags"`
	Security   []map[string][]string `json:"security,omitempty"`
}

// Info represents the OpenAPI Info Object
type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Version        string   `json:"version"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

// Server represents the OpenAPI Server Object
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable represents the OpenAPI Server Variable Object
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem represents the OpenAPI Path Item Object
type PathItem map[string]Operation

// Components represents the OpenAPI Components Object
type Components struct {
	Schemas         map[string]Schema         `json:"schemas"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
	Parameters      map[string]Parameter      `json:"parameters,omitempty"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies,omitempty"`
	Responses       map[string]Response       `json:"responses,omitempty"`
}

// Schema represents the OpenAPI Schema Object
type Schema struct {
	Type       string            `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Format     string            `json:"format,omitempty"`
	Enum       []interface{}     `json:"enum,omitempty"`
	Ref        string            `json:"$ref,omitempty"`
}

// SecurityScheme represents the OpenAPI Security Scheme Object
type SecurityScheme struct {
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

// Operation represents the OpenAPI Operation Object
type Operation struct {
	Summary     string                `json:"summary"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId"`
	Tags        []string              `json:"tags"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter represents the OpenAPI Parameter Object
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

// RequestBody represents the OpenAPI Request Body Object
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents the OpenAPI Response Object
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType represents the OpenAPI Media Type Object
type MediaType struct {
	Schema Schema `json:"schema"`
}

// Tag represents the OpenAPI Tag Object
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
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
	router.GET("/swagger", SwaggerHandler())

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

	// Map Go types to OpenAPI types
	switch field.Type {
	case "string":
		schema.Type = "string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		schema.Type = "integer"
		if field.Type == "int64" || field.Type == "uint64" {
			schema.Format = "int64"
		} else {
			schema.Format = "int32"
		}
	case "float32", "float64":
		schema.Type = "number"
		if field.Type == "float64" {
			schema.Format = "double"
		} else {
			schema.Format = "float"
		}
	case "bool":
		schema.Type = "boolean"
	case "time.Time":
		schema.Type = "string"
		schema.Format = "date-time"
	default:
		// Check if it's an array
		if strings.HasPrefix(field.Type, "[]") {
			schema.Type = "array"
			itemType := strings.TrimPrefix(field.Type, "[]")
			schema.Items = &Schema{}

			// Set the item type
			switch itemType {
			case "string":
				schema.Items.Type = "string"
			case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
				schema.Items.Type = "integer"
			case "float32", "float64":
				schema.Items.Type = "number"
			case "bool":
				schema.Items.Type = "boolean"
			default:
				// Complex type, use a reference
				schema.Items.Ref = "#/components/schemas/" + itemType
			}
		} else {
			// Complex type, use a reference
			schema.Ref = "#/components/schemas/" + field.Type
		}
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
          url: "/api/swagger.json",
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout"
        });
        window.ui = ui;
      };
    </script>
  </body>
</html>
`
