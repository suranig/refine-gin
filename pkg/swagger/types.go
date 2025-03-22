package swagger

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

// SecurityScheme represents the OpenAPI Security Scheme Object
type SecurityScheme struct {
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

// Tag represents the OpenAPI Tag Object
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
