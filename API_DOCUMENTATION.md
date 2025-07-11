# Refine-Gin Framework API Documentation

## Table of Contents

1. [Overview](#overview)
2. [Core Concepts](#core-concepts)
3. [Resource Management](#resource-management)
4. [Field Configuration](#field-configuration)
5. [Authentication & Authorization](#authentication--authorization)
6. [Repository Layer](#repository-layer)
7. [Query System](#query-system)
8. [Middleware](#middleware)
9. [Handler Registration](#handler-registration)
10. [DTO System](#dto-system)
11. [Examples](#examples)
12. [Best Practices](#best-practices)

## Overview

Refine-Gin is a Go framework that provides a complete CRUD API solution with automatic resource management, authentication, and database integration. It's designed to work seamlessly with [Refine.dev](https://refine.dev) frontend framework and provides RESTful APIs with automatic metadata generation.

### Key Features

- **Automatic CRUD Operations**: Generate full CRUD APIs from Go structs
- **Field-Level Configuration**: Rich field configuration with validation, UI hints, and metadata
- **Authentication & Authorization**: JWT-based auth with role-based permissions
- **Database Integration**: GORM-based repository pattern with transaction support
- **Query System**: Advanced filtering, sorting, and pagination
- **Metadata Generation**: Automatic API metadata for frontend consumption
- **Swagger Integration**: Automatic OpenAPI documentation

## Core Concepts

### Resource

A Resource represents an API endpoint and its associated model, operations, and configuration.

```go
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
```

### Operations

Available operations for resources:

```go
const (
    OperationList   Operation = "list"
    OperationRead   Operation = "read"
    OperationCreate Operation = "create"
    OperationUpdate Operation = "update"
    OperationDelete Operation = "delete"
    OperationCount  Operation = "count"
)
```

## Resource Management

### Creating Resources

#### Basic Resource Creation

```go
import "github.com/suranig/refine-gin/pkg/resource"

// Define your model
type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" refine:"filterable;searchable"`
    Email     string    `json:"email" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

// Create resource
userResource := resource.NewResource(resource.ResourceConfig{
    Name:  "users",
    Label: "Users",
    Icon:  "user",
    Model: User{},
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
    },
    DefaultSort: &resource.Sort{
        Field: "created_at",
        Order: "desc",
    },
})
```

#### Advanced Resource Configuration

```go
userResource := resource.NewResource(resource.ResourceConfig{
    Name:  "users",
    Label: "Users",
    Icon:  "user",
    Model: User{},
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
    },
    DefaultSort: &resource.Sort{
        Field: "created_at",
        Order: "desc",
    },
    IDFieldName: "ID", // Custom ID field name
    Permissions: map[string][]string{
        "list":   {"admin", "user"},
        "read":   {"admin", "user"},
        "create": {"admin"},
        "update": {"admin"},
        "delete": {"admin"},
    },
    FilterableFields: []string{"name", "email", "created_at"},
    SearchableFields: []string{"name", "email"},
    SortableFields:   []string{"name", "email", "created_at"},
    TableFields:      []string{"name", "email", "created_at"},
    FormFields:       []string{"name", "email"},
    RequiredFields:   []string{"name", "email"},
    EditableFields:   []string{"name", "email"},
})
```

### Resource Configuration Options

#### ResourceConfig Structure

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
    IDFieldName string              // Custom ID field name (default: "ID")
    Permissions map[string][]string // Map of operations to roles with permission

    // Field lists for different purposes
    FilterableFields []string
    SearchableFields []string
    SortableFields   []string
    TableFields      []string
    FormFields       []string
    RequiredFields   []string
    UniqueFields     []string
    EditableFields   []string
}
```

## Field Configuration

### Field Structure

```go
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
```

### Field Types and Configurations

#### Basic Field Configuration

```go
field := resource.Field{
    Name:  "title",
    Type:  "string",
    Label: "Title",
    Validation: &resource.Validation{
        Required:  true,
        MinLength: 3,
        MaxLength: 100,
    },
    Form: &resource.FormConfig{
        Placeholder: "Enter title",
        Help:        "Title must be between 3 and 100 characters",
    },
    List: &resource.ListConfig{
        Width:    200,
        Ellipsis: true,
    },
}
```

#### JSON Field Configuration

```go
jsonField := resource.Field{
    Name: "config",
    Type: "json",
    Label: "Configuration",
    Json: &resource.JsonConfig{
        DefaultExpanded: true,
        EditorType:      "form",
        Schema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "apiKey": map[string]interface{}{
                    "type": "string",
                },
                "enabled": map[string]interface{}{
                    "type": "boolean",
                },
            },
        },
        Properties: []resource.JsonProperty{
            {
                Path:  "apiKey",
                Label: "API Key",
                Type:  "string",
                Validation: &resource.JsonValidation{
                    Required: true,
                },
            },
            {
                Path:  "enabled",
                Label: "Enabled",
                Type:  "boolean",
            },
        },
    },
}
```

#### File/Image Field Configuration

```go
fileField := resource.Field{
    Name: "avatar",
    Type: "file",
    Label: "Avatar",
    File: &resource.FileConfig{
        AllowedTypes: []string{"image/jpeg", "image/png", "image/gif"},
        MaxSize:      5 * 1024 * 1024, // 5MB
        IsImage:      true,
        MaxWidth:     800,
        MaxHeight:    600,
        GenerateThumbnails: true,
        ThumbnailSizes: []resource.ThumbnailSize{
            {Name: "small", Width: 100, Height: 100, KeepAspectRatio: true},
            {Name: "medium", Width: 300, Height: 300, KeepAspectRatio: true},
        },
    },
}
```

#### Rich Text Field Configuration

```go
richTextField := resource.Field{
    Name: "content",
    Type: "richtext",
    Label: "Content",
    RichText: &resource.RichTextConfig{
        Toolbar: []string{"bold", "italic", "underline", "link", "image"},
        Height:  "300px",
        Placeholder: "Enter content...",
        EnableImages: true,
        MaxLength: 10000,
        ShowCounter: true,
        Format: "html",
    },
}
```

#### Select Field Configuration

```go
selectField := resource.Field{
    Name: "category",
    Type: "select",
    Label: "Category",
    Select: &resource.SelectConfig{
        Multiple:     false,
        Searchable:   true,
        Creatable:    true,
        Clearable:    true,
        Placeholder:  "Select category",
        DisplayMode:  "dropdown",
    },
    Options: []resource.Option{
        {Value: "tech", Label: "Technology"},
        {Value: "news", Label: "News"},
        {Value: "sports", Label: "Sports"},
    },
}
```

#### Computed Field Configuration

```go
computedField := resource.Field{
    Name: "fullName",
    Type: "computed",
    Label: "Full Name",
    Computed: &resource.ComputedFieldConfig{
        DependsOn: []string{"firstName", "lastName"},
        Expression: "firstName + ' ' + lastName",
        ClientSide: true,
        Persist:    false,
    },
}
```

### Field Tags

You can configure fields using struct tags:

```go
type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" refine:"filterable;searchable;label=Full Name;placeholder=Enter full name"`
    Email     string    `json:"email" refine:"filterable;required;validation=email"`
    Age       int       `json:"age" refine:"min=18;max=120;help=Must be between 18 and 120"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}
```

#### Available Field Tags

- `filterable` - Field can be used for filtering
- `searchable` - Field is included in search
- `sortable` - Field can be sorted
- `required` - Field is required
- `readonly` - Field is read-only
- `hidden` - Field is hidden in UI
- `label=<text>` - Custom label for the field
- `placeholder=<text>` - Placeholder text
- `help=<text>` - Help text
- `tooltip=<text>` - Tooltip text
- `validation=<rule>` - Validation rule
- `min=<value>` - Minimum value
- `max=<value>` - Maximum value
- `minLength=<value>` - Minimum length
- `maxLength=<value>` - Maximum length

## Authentication & Authorization

### JWT Configuration

```go
import "github.com/suranig/refine-gin/pkg/auth"

// Create JWT configuration
jwtConfig := auth.JWTConfig{
    Secret:         "your-secret-key",
    ExpirationTime: time.Hour * 24,
    Issuer:         "your-app",
    Audience:       "your-app-api",
}

// Use default configuration
jwtConfig := auth.DefaultJWTConfig()
```

### JWT Middleware

```go
// Add JWT middleware to your routes
router.Use(auth.JWTMiddleware(jwtConfig))
```

### Generating JWT Tokens

```go
// Generate token with standard claims
token, err := auth.GenerateJWTWithStandardClaims(jwtConfig, "user123", map[string]interface{}{
    "role": "admin",
    "permissions": []string{"read", "write"},
})

// Generate token with custom claims
claims := jwt.MapClaims{
    "sub": "user123",
    "role": "admin",
    "permissions": []string{"read", "write"},
}
token, err := auth.GenerateJWT(jwtConfig, claims)
```

### Permission-Based Authorization

```go
// Configure resource permissions
userResource := resource.NewResource(resource.ResourceConfig{
    Name: "users",
    Model: User{},
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
    },
    Permissions: map[string][]string{
        "list":   {"admin", "manager"},
        "read":   {"admin", "manager", "user"},
        "create": {"admin"},
        "update": {"admin", "manager"},
        "delete": {"admin"},
    },
})

// Check permissions
if userResource.HasPermission("create", "admin") {
    // Allow create operation
}
```

### Field-Level Permissions

```go
field := resource.Field{
    Name: "salary",
    Type: "number",
    Label: "Salary",
    Permissions: map[string][]string{
        "read":   {"admin", "manager"},
        "update": {"admin"},
    },
}
```

## Repository Layer

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

    // Utilities
    GetIDFieldName() string
}
```

### Creating Repositories

#### Generic Repository

```go
import "github.com/suranig/refine-gin/pkg/repository"

// Create generic repository
userRepo := repository.NewGenericRepository(db, &User{})

// Create repository with resource
userRepo := repository.NewGenericRepositoryWithResource(db, userResource)
```

#### Owner Repository

```go
// Create owner repository for user-owned resources
postRepo := repository.NewOwnerRepository(db, &Post{}, "user_id", "users")
```

### Using Repositories

```go
// Basic operations
user, err := userRepo.Get(ctx, "user123")
users, total, err := userRepo.List(ctx, queryOptions)
newUser, err := userRepo.Create(ctx, userData)
updatedUser, err := userRepo.Update(ctx, "user123", updateData)
err := userRepo.Delete(ctx, "user123")

// With relations
user, err := userRepo.GetWithRelations(ctx, "user123", []string{"posts", "profile"})
users, total, err := userRepo.ListWithRelations(ctx, queryOptions, []string{"posts"})

// Transactions
err := userRepo.WithTransaction(func(repo repository.Repository) error {
    // Create user
    user, err := repo.Create(ctx, userData)
    if err != nil {
        return err
    }
    
    // Create profile
    profile := Profile{UserID: user.ID, Bio: "New user"}
    _, err = repo.Create(ctx, profile)
    return err
})
```

## Query System

### Query Options

```go
type QueryOptions struct {
    Resource resource.Resource
    Page     int
    PerPage  int
    Search   string
    Filters  map[string]interface{}
    AdvancedFilters []Filter
    Sort     string
    Order    string
}
```

### Creating Query Options

```go
import "github.com/suranig/refine-gin/pkg/query"

// Create query options from Gin context
queryOptions := query.NewQueryOptions(c, userResource)

// Manual query options
queryOptions := query.QueryOptions{
    Resource: userResource,
    Page:     1,
    PerPage:  10,
    Search:   "john",
    Filters: map[string]interface{}{
        "email": "john@example.com",
    },
    Sort:  "created_at",
    Order: "desc",
}
```

### Advanced Filtering

```go
// Advanced filters with operators
queryOptions.AdvancedFilters = []query.Filter{
    {
        Field:    "age",
        Operator: "gte",
        Value:    18,
    },
    {
        Field:    "name",
        Operator: "contains",
        Value:    "john",
    },
    {
        Field:    "created_at",
        Operator: "between",
        Value:    []time.Time{startDate, endDate},
    },
}
```

### Supported Operators

- `eq` - Equal
- `ne` - Not equal
- `gt` - Greater than
- `gte` - Greater than or equal
- `lt` - Less than
- `lte` - Less than or equal
- `contains` - Contains substring
- `startsWith` - Starts with
- `endsWith` - Ends with
- `in` - In array
- `notIn` - Not in array
- `between` - Between two values
- `isNull` - Is null
- `isNotNull` - Is not null

### Sorting

```go
// Single field sorting
queryOptions.Sort = "created_at"
queryOptions.Order = "desc"

// Multiple field sorting
queryOptions.Sort = "name asc, created_at desc"
```

## Middleware

### Available Middleware

#### Cache Middleware

```go
import "github.com/suranig/refine-gin/pkg/middleware"

// Default cache configuration
cacheConfig := middleware.DefaultCacheConfig()

// Custom cache configuration
cacheConfig := middleware.CacheConfig{
    Duration: time.Minute * 5,
    Methods:  []string{"GET", "OPTIONS"},
    Headers:  []string{"Authorization"},
}

// Apply cache middleware
router.Use(middleware.CacheByResource("users", cacheConfig))
router.Use(middleware.NoCacheMiddleware()) // Disable cache
```

#### Naming Convention Middleware

```go
// Apply naming convention middleware
router.Use(middleware.NamingConventionMiddleware(resource.NamingConventionSnakeCase))
router.Use(middleware.NamingConventionMiddleware(resource.NamingConventionCamelCase))
```

#### Owner Middleware

```go
// Apply owner middleware for user-owned resources
router.Use(middleware.OwnerMiddleware("user_id", "users"))
```

### Custom Middleware

```go
// Create custom middleware
func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Your middleware logic here
        c.Next()
    }
}

// Apply to routes
router.Use(CustomMiddleware())
```

## Handler Registration

### Basic Registration

```go
import "github.com/suranig/refine-gin/pkg/handler"

// Register resource with default settings
handler.RegisterResource(router, userResource, userRepo)
```

### Registration with DTO

```go
// Register with custom DTO provider
handler.RegisterResourceWithDTO(router, userResource, userRepo, customDTOProvider)
```

### Registration with Options

```go
// Register with custom options
opts := resource.Options{
    NamingConvention: resource.NamingConventionCamelCase,
}
handler.RegisterResourceWithOptions(router, userResource, userRepo, opts)
```

### Registration for Refine.dev

```go
// Register optimized for Refine.dev
handler.RegisterResourceForRefine(router, userResource, userRepo, "id")
```

### Custom ID Parameter

```go
// Register with custom ID parameter name
handler.RegisterResourceForRefine(router, userResource, userRepo, "user_id")
```

## DTO System

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
import "github.com/suranig/refine-gin/pkg/dto"

// Create default DTO provider
dtoProvider := &dto.DefaultDTOProvider{
    Model: User{},
}

// Use with handler registration
handler.RegisterResourceWithDTO(router, userResource, userRepo, dtoProvider)
```

### Custom DTO Provider

```go
type CustomUserDTOProvider struct {
    dto.DefaultDTOProvider
}

func (p *CustomUserDTOProvider) ToDTO(data interface{}) interface{} {
    user := data.(*User)
    return map[string]interface{}{
        "id":    user.ID,
        "name":  user.Name,
        "email": user.Email,
        // Add computed fields
        "fullName": user.Name + " (" + user.Email + ")",
    }
}

func (p *CustomUserDTOProvider) FromDTO(dto interface{}) interface{} {
    dtoMap := dto.(map[string]interface{})
    return &User{
        ID:    dtoMap["id"].(string),
        Name:  dtoMap["name"].(string),
        Email: dtoMap["email"].(string),
    }
}
```

## Examples

### Complete Example

```go
package main

import (
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/suranig/refine-gin/pkg/auth"
    "github.com/suranig/refine-gin/pkg/handler"
    "github.com/suranig/refine-gin/pkg/repository"
    "github.com/suranig/refine-gin/pkg/resource"
    "github.com/suranig/refine-gin/pkg/swagger"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// Models
type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" refine:"filterable;searchable;required"`
    Email     string    `json:"email" refine:"filterable;required;validation=email"`
    Role      string    `json:"role" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

type Post struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title" refine:"filterable;searchable;required"`
    Content   string    `json:"content" refine:"required"`
    UserID    string    `json:"user_id" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable;readonly"`
}

func main() {
    // Setup database
    db, err := gorm.Open(postgres.Open("dsn"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    // Auto migrate
    db.AutoMigrate(&User{}, &Post{})

    // Create resources
    userResource := resource.NewResource(resource.ResourceConfig{
        Name:  "users",
        Label: "Users",
        Icon:  "user",
        Model: User{},
        Operations: []resource.Operation{
            resource.OperationList,
            resource.OperationRead,
            resource.OperationCreate,
            resource.OperationUpdate,
            resource.OperationDelete,
        },
        DefaultSort: &resource.Sort{
            Field: "created_at",
            Order: "desc",
        },
        Permissions: map[string][]string{
            "list":   {"admin", "manager"},
            "read":   {"admin", "manager", "user"},
            "create": {"admin"},
            "update": {"admin", "manager"},
            "delete": {"admin"},
        },
    })

    postResource := resource.NewResource(resource.ResourceConfig{
        Name:  "posts",
        Label: "Posts",
        Icon:  "file-text",
        Model: Post{},
        Operations: []resource.Operation{
            resource.OperationList,
            resource.OperationRead,
            resource.OperationCreate,
            resource.OperationUpdate,
            resource.OperationDelete,
        },
    })

    // Create repositories
    userRepo := repository.NewGenericRepositoryWithResource(db, userResource)
    postRepo := repository.NewOwnerRepository(db, &Post{}, "user_id", "users")

    // Setup router
    r := gin.Default()

    // CORS middleware
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    // JWT middleware
    jwtConfig := auth.DefaultJWTConfig()
    jwtConfig.Secret = "your-secret-key"
    r.Use(auth.JWTMiddleware(jwtConfig))

    // Register resources
    api := r.Group("/api")
    handler.RegisterResourceForRefine(api, userResource, userRepo, "id")
    handler.RegisterResourceForRefine(api, postResource, postRepo, "id")

    // Swagger documentation
    swaggerInfo := swagger.SwaggerInfo{
        Title:       "My API",
        Description: "API documentation",
        Version:     "1.0.0",
        BasePath:    "/api",
    }
    swagger.RegisterSwagger(r.Group(""), []resource.Resource{userResource, postResource}, swaggerInfo)

    // Start server
    log.Fatal(r.Run(":8080"))
}
```

### Advanced Field Configuration Example

```go
// Complex field configuration
userResource := resource.NewResource(resource.ResourceConfig{
    Name:  "users",
    Model: User{},
    Fields: []resource.Field{
        {
            Name:  "name",
            Type:  "string",
            Label: "Full Name",
            Validation: &resource.Validation{
                Required:  true,
                MinLength: 2,
                MaxLength: 100,
            },
            Form: &resource.FormConfig{
                Placeholder: "Enter full name",
                Help:        "Enter the user's full name",
                WidthPercent: 50,
            },
            List: &resource.ListConfig{
                Width:    200,
                Ellipsis: true,
            },
        },
        {
            Name:  "email",
            Type:  "string",
            Label: "Email Address",
            Validation: &resource.Validation{
                Required: true,
                Pattern:  `^[^\s@]+@[^\s@]+\.[^\s@]+$`,
                Message:  "Please enter a valid email address",
            },
            Form: &resource.FormConfig{
                Placeholder: "Enter email address",
                WidthPercent: 50,
            },
        },
        {
            Name:  "role",
            Type:  "select",
            Label: "Role",
            Select: &resource.SelectConfig{
                Multiple:    false,
                Searchable:  true,
                Clearable:   true,
                Placeholder: "Select role",
            },
            Options: []resource.Option{
                {Value: "admin", Label: "Administrator"},
                {Value: "manager", Label: "Manager"},
                {Value: "user", Label: "User"},
            },
        },
        {
            Name:  "avatar",
            Type:  "file",
            Label: "Avatar",
            File: &resource.FileConfig{
                AllowedTypes: []string{"image/jpeg", "image/png"},
                MaxSize:      2 * 1024 * 1024, // 2MB
                IsImage:      true,
                MaxWidth:     500,
                MaxHeight:    500,
            },
        },
        {
            Name:  "bio",
            Type:  "richtext",
            Label: "Biography",
            RichText: &resource.RichTextConfig{
                Toolbar: []string{"bold", "italic", "link"},
                Height:  "200px",
                Placeholder: "Tell us about yourself...",
            },
        },
        {
            Name:  "settings",
            Type:  "json",
            Label: "Settings",
            Json: &resource.JsonConfig{
                DefaultExpanded: false,
                EditorType:      "form",
                Properties: []resource.JsonProperty{
                    {
                        Path:  "notifications.email",
                        Label: "Email Notifications",
                        Type:  "boolean",
                    },
                    {
                        Path:  "notifications.push",
                        Label: "Push Notifications",
                        Type:  "boolean",
                    },
                    {
                        Path:  "theme",
                        Label: "Theme",
                        Type:  "string",
                        Options: []resource.Option{
                            {Value: "light", Label: "Light"},
                            {Value: "dark", Label: "Dark"},
                        },
                    },
                },
            },
        },
    },
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
    },
})
```

## Best Practices

### 1. Resource Organization

- Group related resources together
- Use consistent naming conventions
- Define clear permissions for each operation
- Use meaningful labels and icons

### 2. Field Configuration

- Always provide validation rules for required fields
- Use appropriate field types for data
- Configure UI hints (placeholders, help text)
- Set up proper list and form configurations

### 3. Security

- Always use strong JWT secrets in production
- Implement proper role-based permissions
- Validate all input data
- Use HTTPS in production

### 4. Performance

- Use appropriate database indexes
- Implement caching where appropriate
- Limit the number of relations loaded
- Use pagination for large datasets

### 5. API Design

- Follow RESTful conventions
- Provide consistent error responses
- Include proper HTTP status codes
- Document your APIs with Swagger

### 6. Error Handling

```go
// Custom error handling middleware
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": err.Error(),
            })
        }
    }
}
```

### 7. Logging

```go
// Custom logging middleware
func LoggingMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
            param.ClientIP,
            param.TimeStamp.Format(time.RFC1123),
            param.Method,
            param.Path,
            param.Request.Proto,
            param.StatusCode,
            param.Latency,
            param.Request.UserAgent(),
            param.ErrorMessage,
        )
    })
}
```

This documentation covers all the major components and APIs of the Refine-Gin framework. For more specific examples and advanced usage patterns, refer to the examples directory in the project repository.