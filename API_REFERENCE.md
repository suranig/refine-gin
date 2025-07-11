# Refine-Gin API Reference

## Quick Reference Guide

This document provides a quick reference to all public APIs, functions, and their signatures in the Refine-Gin framework.

## Resource Package

### Core Types

```go
// Resource interface
type Resource interface {
    GetName() string
    GetLabel() string
    GetIcon() string
    GetModel() interface{}
    GetFields() []Field
    GetOperations() []Operation
    HasOperation(op Operation) bool
    GetDefaultSort() *Sort
    GetFilters() []Filter
    GetMiddlewares() []interface{}
    GetRelations() []Relation
    HasRelation(name string) bool
    GetRelation(name string) *Relation
    GetIDFieldName() string
    GetField(name string) *Field
    GetSearchable() []string
    GetFilterableFields() []string
    GetSortableFields() []string
    GetTableFields() []string
    GetFormFields() []string
    GetRequiredFields() []string
    GetEditableFields() []string
    GetPermissions() map[string][]string
    HasPermission(operation string, role string) bool
    GetFormLayout() *FormLayout
}

// Operations
const (
    OperationList   Operation = "list"
    OperationRead   Operation = "read"
    OperationCreate Operation = "create"
    OperationUpdate Operation = "update"
    OperationDelete Operation = "delete"
    OperationCount  Operation = "count"
)
```

### Resource Creation Functions

```go
// Create new resource
func NewResource(config ResourceConfig) Resource

// Generate fields from model
func GenerateFieldsFromModel(model interface{}) []Field

// Extract fields from model
func ExtractFieldsFromModel(model interface{}) []Field

// Create slice of type
func CreateSliceOfType(model interface{}) interface{}

// Create instance of type
func CreateInstanceOfType(model interface{}) interface{}

// Set ID on object
func SetID(obj interface{}, id interface{}) error

// Set custom ID on object
func SetCustomID(obj interface{}, id interface{}, idFieldName string) error
```

### Field Configuration

```go
// Field structure
type Field struct {
    Name        string
    Type        string
    Label       string
    Validation  *Validation
    Options     []Option
    Relation    *RelationConfig
    List        *ListConfig
    Form        *FormConfig
    Validators  []Validator
    Json        *JsonConfig
    ReadOnly    bool
    Hidden      bool
    File        *FileConfig
    RichText    *RichTextConfig
    Select      *SelectConfig
    Computed    *ComputedFieldConfig
    AntDesign   *AntDesignConfig
    Permissions map[string][]string
}

// Parse field tag
func ParseFieldTag(field *Field, tag string)

// Process JSON tag
func ProcessJsonTag(field *Field, tag string)
```

### Configuration Structures

```go
type ResourceConfig struct {
    Name        string
    Label       string
    Icon        string
    Model       interface{}
    Fields      []Field
    Operations  []Operation
    DefaultSort *Sort
    Filters     []Filter
    Middlewares []interface{}
    Relations   []Relation
    IDFieldName string
    Permissions map[string][]string
    FilterableFields []string
    SearchableFields []string
    SortableFields   []string
    TableFields      []string
    FormFields       []string
    RequiredFields   []string
    UniqueFields     []string
    EditableFields   []string
}

type Validation struct {
    Required       bool
    Min            float64
    Max            float64
    MinLength      int
    MaxLength      int
    Pattern        string
    Message        string
    Custom         string
    Conditional    *ConditionalValidation
    AsyncValidator string
}

type Sort struct {
    Field string
    Order string
}

type Filter struct {
    Field    string
    Operator string
    Value    interface{}
}
```

## Handler Package

### Registration Functions

```go
// Basic registration
func RegisterResource(router *gin.RouterGroup, res resource.Resource, repo repository.Repository)

// Registration with DTO
func RegisterResourceWithDTO(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider)

// Registration with options
func RegisterResourceWithOptions(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, opts resource.Options)

// Registration for Refine.dev
func RegisterResourceForRefine(router *gin.RouterGroup, res resource.Resource, repo repository.Repository, idParamName string)
```

### Handler Generation Functions

```go
// Generate handlers
func GenerateListHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc
func GenerateListHandlerWithDTO(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc
func GenerateGetHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc
func GenerateGetHandlerWithDTO(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc
func GenerateGetHandlerWithParam(res resource.Resource, repo repository.Repository, paramName string) gin.HandlerFunc
func GenerateGetHandlerWithParamAndDTO(res resource.Resource, repo repository.Repository, paramName string, dtoProvider dto.DTOProvider) gin.HandlerFunc
func GenerateCreateHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc
func GenerateUpdateHandler(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider) gin.HandlerFunc
func GenerateUpdateHandlerWithParam(res resource.Resource, repo repository.Repository, dtoProvider dto.DTOProvider, paramName string) gin.HandlerFunc
func GenerateCustomUpdateHandler(res resource.Resource, repo repository.Repository, paramName string) gin.HandlerFunc
func GenerateDeleteHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc
func GenerateDeleteHandlerWithParam(res resource.Resource, repo repository.Repository, paramName string) gin.HandlerFunc
func GenerateCountHandler(res resource.Resource, repo repository.Repository) gin.HandlerFunc
func GenerateOptionsHandler(res resource.Resource) gin.HandlerFunc
```

## Repository Package

### Repository Interface

```go
type Repository interface {
    // Basic operations
    Get(ctx context.Context, id interface{}) (interface{}, error)
    List(ctx context.Context, options query.QueryOptions) (interface{}, int64, error)
    Create(ctx context.Context, data interface{}) (interface{}, error)
    Update(ctx context.Context, id interface{}, data interface{}) (interface{}, error)
    Delete(ctx context.Context, id interface{}) error

    // Bulk operations
    Count(ctx context.Context, options query.QueryOptions) (int64, error)
    CreateMany(ctx context.Context, data interface{}) (interface{}, error)
    UpdateMany(ctx context.Context, ids []interface{}, data interface{}) (int64, error)
    DeleteMany(ctx context.Context, ids []interface{}) (int64, error)

    // Relations
    WithRelations(relations ...string) Repository
    GetWithRelations(ctx context.Context, id interface{}, relations []string) (interface{}, error)
    ListWithRelations(ctx context.Context, options query.QueryOptions, relations []string) (interface{}, int64, error)

    // Query helpers
    Query(ctx context.Context) *gorm.DB
    FindOneBy(ctx context.Context, condition map[string]interface{}) (interface{}, error)
    FindAllBy(ctx context.Context, condition map[string]interface{}) (interface{}, error)

    // Transaction support
    WithTransaction(fn func(Repository) error) error

    // GORM-specific operations
    BulkCreate(ctx context.Context, data interface{}) error
    BulkUpdate(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error

    // Utilities
    GetIDFieldName() string
}
```

### Repository Creation Functions

```go
// Create generic repository
func NewGenericRepository(db *gorm.DB, model interface{}) Repository

// Create generic repository with resource
func NewGenericRepositoryWithResource(db *gorm.DB, resource resource.Resource) Repository

// Create owner repository
func NewOwnerRepository(db *gorm.DB, model interface{}, ownerField, ownerTable string) Repository

// Create repository from factory
func NewRepository(db *gorm.DB, model interface{}, resource resource.Resource) Repository
```

## Query Package

### Query Options

```go
type QueryOptions struct {
    Resource resource.Resource
    Page     int
    PerPage  int
    DisablePagination bool
    Search   string
    Filters  map[string]interface{}
    AdvancedFilters []Filter
    Sort     string
    Order    string
}

type Filter struct {
    Field    string
    Operator string
    Value    interface{}
}
```

### Query Functions

```go
// Create query options from Gin context
func NewQueryOptions(c *gin.Context, res resource.Resource) QueryOptions

// Parse filters from query parameters
func ParseFilters(c *gin.Context) map[string]interface{}

// Parse advanced filters
func ParseAdvancedFilters(c *gin.Context) []Filter

// Parse sorting
func ParseSorting(c *gin.Context) (string, string)

// Parse pagination
func ParsePagination(c *gin.Context) (int, int)
```

### Sorting

```go
const (
    SortOrderAsc  SortOrder = "asc"
    SortOrderDesc SortOrder = "desc"
)

type Sort struct {
    Field string
    Order string
}
```

## Auth Package

### JWT Configuration

```go
type JWTConfig struct {
    Secret         string
    ExpirationTime time.Duration
    Issuer         string
    Audience       string
    ClaimsExtractor func(token *jwt.Token) (interface{}, error)
}
```

### JWT Functions

```go
// Default JWT configuration
func DefaultJWTConfig() JWTConfig

// JWT middleware
func JWTMiddleware(config JWTConfig) gin.HandlerFunc

// Generate JWT token
func GenerateJWT(config JWTConfig, claims jwt.Claims) (string, error)

// Generate JWT with standard claims
func GenerateJWTWithStandardClaims(config JWTConfig, subject string, customClaims map[string]interface{}) (string, error)

// Extract subject from token
func ExtractSubjectFromToken(tokenString string, secret string) (string, error)

// Extract claims from token
func ExtractClaimsFromToken(tokenString string, secret string) (jwt.MapClaims, error)
```

### Permission Functions

```go
// Check permissions
func HasPermission(permissions map[string][]string, operation, role string) bool

// Get user permissions from context
func GetUserPermissions(c *gin.Context) map[string]interface{}

// Get user role from context
func GetUserRole(c *gin.Context) string
```

## Middleware Package

### Cache Middleware

```go
type CacheConfig struct {
    Duration time.Duration
    Methods  []string
    Headers  []string
}

// Default cache configuration
func DefaultCacheConfig() CacheConfig

// Cache by resource
func CacheByResource(resourceName string, config CacheConfig) gin.HandlerFunc

// No cache middleware
func NoCacheMiddleware() gin.HandlerFunc
```

### Naming Convention Middleware

```go
const (
    NamingConventionSnakeCase NamingConvention = "snake_case"
    NamingConventionCamelCase NamingConvention = "camelCase"
)

// Naming convention middleware
func NamingConventionMiddleware(convention NamingConvention) gin.HandlerFunc
```

### Owner Middleware

```go
// Owner middleware
func OwnerMiddleware(ownerField, ownerTable string) gin.HandlerFunc
```

## DTO Package

### DTO Provider Interface

```go
type DTOProvider interface {
    ToDTO(data interface{}) interface{}
    FromDTO(dto interface{}) interface{}
    ToDTOList(data interface{}) interface{}
    FromDTOList(dto interface{}) interface{}
}
```

### Default DTO Provider

```go
type DefaultDTOProvider struct {
    Model interface{}
}

func (p *DefaultDTOProvider) ToDTO(data interface{}) interface{}
func (p *DefaultDTOProvider) FromDTO(dto interface{}) interface{}
func (p *DefaultDTOProvider) ToDTOList(data interface{}) interface{}
func (p *DefaultDTOProvider) FromDTOList(dto interface{}) interface{}
```

## Swagger Package

### Swagger Configuration

```go
type SwaggerInfo struct {
    Title       string
    Description string
    Version     string
    BasePath    string
}
```

### Swagger Functions

```go
// Register Swagger routes
func RegisterSwagger(router *gin.RouterGroup, resources []resource.Resource, info SwaggerInfo)

// Generate OpenAPI spec
func GenerateOpenAPISpec(resources []resource.Resource, info SwaggerInfo) map[string]interface{}
```

## Utils Package

### Utility Functions

```go
// Convert to snake case
func ToSnakeCase(s string) string

// Convert to camel case
func ToCamelCase(s string) string

// Convert to pascal case
func ToPascalCase(s string) string

// Convert to kebab case
func ToKebabCase(s string) string

// Convert struct to map
func StructToMap(obj interface{}) map[string]interface{}

// Convert map to struct
func MapToStruct(data map[string]interface{}, result interface{}) error

// Deep copy
func DeepCopy(src interface{}) interface{}

// Merge maps
func MergeMaps(maps ...map[string]interface{}) map[string]interface{}
```

## Field Types Reference

### Supported Field Types

- `string` - Text input
- `number` - Numeric input
- `boolean` - Checkbox/switch
- `date` - Date picker
- `datetime` - Date and time picker
- `time` - Time picker
- `email` - Email input
- `url` - URL input
- `password` - Password input
- `textarea` - Multi-line text
- `select` - Dropdown selection
- `multiselect` - Multiple selection
- `radio` - Radio buttons
- `checkbox` - Checkboxes
- `file` - File upload
- `image` - Image upload
- `richtext` - Rich text editor
- `json` - JSON editor
- `computed` - Computed field
- `relation` - Related resource
- `array` - Array of values
- `object` - Object structure

### Field Configuration Options

#### List Configuration
```go
type ListConfig struct {
    Width    int
    Fixed    string
    Ellipsis bool
}
```

#### Form Configuration
```go
type FormConfig struct {
    Placeholder string
    Help        string
    Tooltip     string
    Width       string
    DependentOn string
    Condition   string
    Dependent   *FormDependency
    WidthPercent int
    VisibilityCondition *FormCondition
}
```

#### JSON Configuration
```go
type JsonConfig struct {
    Schema         map[string]interface{}
    Properties     []JsonProperty
    DefaultExpanded bool
    PathPrefix     string
    EditorType     string
    Nested         bool
    RenderAs       string
    TabsConfig     *JsonTabsConfig
    GridConfig     *JsonGridConfig
    ObjectLabels   map[string]string
}
```

#### File Configuration
```go
type FileConfig struct {
    AllowedTypes        []string
    MaxSize             int64
    StoragePath         string
    BaseURL             string
    IsImage             bool
    MaxWidth            int
    MaxHeight           int
    GenerateThumbnails  bool
    ThumbnailSizes      []ThumbnailSize
}
```

#### Rich Text Configuration
```go
type RichTextConfig struct {
    Toolbar       []string
    Height        string
    Placeholder   string
    EnableImages  bool
    MaxLength     int
    ShowCounter   bool
    Format        string
}
```

#### Select Configuration
```go
type SelectConfig struct {
    Multiple          bool
    Searchable        bool
    Creatable         bool
    OptionsURL        string
    DependsOn         string
    DependentOptions  map[string][]Option
    Placeholder       string
    Clearable         bool
    DisplayMode       string
}
```

#### Computed Field Configuration
```go
type ComputedFieldConfig struct {
    DependsOn    []string
    Expression   string
    ClientSide   bool
    Format       string
    Persist      bool
    ComputeOrder int
}
```

## Validation Rules Reference

### Built-in Validation Rules

- `required` - Field is required
- `email` - Valid email format
- `url` - Valid URL format
- `min=<value>` - Minimum value/length
- `max=<value>` - Maximum value/length
- `minLength=<value>` - Minimum string length
- `maxLength=<value>` - Maximum string length
- `pattern=<regex>` - Regular expression pattern
- `in=<value1,value2,...>` - Value must be in list
- `notIn=<value1,value2,...>` - Value must not be in list
- `between=<min,max>` - Value must be between min and max
- `alpha` - Alphabetic characters only
- `alphanumeric` - Alphanumeric characters only
- `numeric` - Numeric characters only
- `integer` - Integer value only
- `decimal` - Decimal value only
- `date` - Valid date format
- `datetime` - Valid datetime format
- `time` - Valid time format
- `unique` - Value must be unique in database
- `exists` - Value must exist in database
- `custom=<function>` - Custom validation function

### Conditional Validation

```go
type ConditionalValidation struct {
    Field    string
    Operator string
    Value    interface{}
    Message  string
}
```

### Async Validation

```go
type AsyncValidator struct {
    URL     string
    Message string
}
```

## Error Handling

### Common Error Types

```go
// Resource not found
type ResourceNotFoundError struct {
    Resource string
    ID       interface{}
}

// Validation error
type ValidationError struct {
    Field   string
    Message string
}

// Permission error
type PermissionError struct {
    Operation string
    Role      string
}

// Database error
type DatabaseError struct {
    Operation string
    Error     error
}
```

### Error Response Format

```go
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Message string                 `json:"message"`
    Code    int                    `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

## HTTP Status Codes

### Success Codes
- `200 OK` - Request successful
- `201 Created` - Resource created
- `204 No Content` - Request successful, no content

### Client Error Codes
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Permission denied
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict
- `422 Unprocessable Entity` - Validation error

### Server Error Codes
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Gateway error
- `503 Service Unavailable` - Service unavailable

## Response Formats

### List Response
```json
{
  "data": [...],
  "total": 100,
  "page": 1,
  "perPage": 10,
  "totalPages": 10
}
```

### Single Resource Response
```json
{
  "data": {...}
}
```

### Error Response
```json
{
  "error": "error_type",
  "message": "Error description",
  "code": 400,
  "details": {...}
}
```

### Metadata Response
```json
{
  "fields": [...],
  "operations": [...],
  "permissions": {...},
  "filters": [...],
  "sorts": [...]
}
```

This reference guide provides quick access to all public APIs and their signatures. For detailed usage examples and best practices, refer to the main API documentation.