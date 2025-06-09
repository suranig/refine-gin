# Refine-Gin

[![Go Tests](https://github.com/suranig/refine-gin/actions/workflows/ci.yml/badge.svg)](https://github.com/suranig/refine-gin/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/suranig/refine-gin/graph/badge.svg)](https://codecov.io/gh/suranig/refine-gin)
[![Go Report Card](https://goreportcard.com/badge/github.com/suranig/refine-gin)](https://goreportcard.com/report/github.com/suranig/refine-gin)
[![GoDoc](https://godoc.org/github.com/suranig/refine-gin?status.svg)](https://godoc.org/github.com/suranig/refine-gin)
[![Go Reference](https://pkg.go.dev/badge/github.com/suranig/refine-gin.svg)](https://pkg.go.dev/github.com/suranig/refine-gin)

Refine-Gin is a library that integrates the Gin framework with Refine.js, enabling rapid development of RESTful APIs compatible with Refine.js conventions.

## Dependencies

This library integrates the following technologies:

- [Gin](https://github.com/gin-gonic/gin) - A high-performance HTTP web framework written in Go
- [Refine](https://refine.dev/) - A React-based framework for building data-intensive applications
- [JWT-Go](https://github.com/golang-jwt/jwt) - A Go implementation of JSON Web Tokens
- [GORM](https://gorm.io/) - The fantastic ORM library for Go

## Features

- Automatic generation of REST endpoints based on resource definitions
- Full compatibility with Refine.js conventions (filters, sorting, pagination)
- Input data validation
- Data transformation through DTO layer
- Support for relationships between resources
- Centralized field lists for filtering, sorting, and searching
- Advanced reflection utilities for easier dynamic data handling
- Type mapping for consistent schema generation
- JWT authentication and authorization
- Customizable validation rules
- Flexible JSON naming convention control (snake_case, camelCase, PascalCase)
- Count endpoint for resources
- Custom endpoints and actions
- Swagger for resources

## Installation

```bash
go get github.com/suranig/refine-gin
```

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/suranig/refine-gin/pkg/handler"
    "github.com/suranig/refine-gin/pkg/resource"
    "gorm.io/gorm"
)

// Model definition
type User struct {
    ID        string    `json:"id" gorm:"primaryKey" refine:"filterable;sortable;searchable"`
    Name      string    `json:"name" refine:"filterable;sortable"`
    Email     string    `json:"email" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}

// Repository implementation
type UserRepository struct {
    db *gorm.DB
}

// Implement repository methods...

func main() {
    r := gin.Default()
    
    // Resource definition
    userResource := resource.NewResource(resource.ResourceConfig{
        Name: "users",
        Model: User{},
        Operations: []resource.Operation{
            resource.OperationList, 
            resource.OperationCreate, 
            resource.OperationRead, 
            resource.OperationUpdate, 
            resource.OperationDelete,
        },
    })
    
    // Register resource
    api := r.Group("/api")
    handler.RegisterResource(api, userResource, userRepository)
    
    r.Run(":8080")
}
```

## API Documentation

### Resource Definition

Resources are defined using the `ResourceConfig` structure. There are two main approaches to defining field properties:

```go
userResource := resource.NewResource(resource.ResourceConfig{
    Name: "users",
    Model: User{},
    Fields: []resource.Field{
        {Name: "id", Type: "string"},
        {Name: "name", Type: "string"},
        {Name: "email", Type: "string"},
        {Name: "created_at", Type: "time.Time"},
    },
    // Define properties using field lists
    FilterableFields: []string{"id", "name", "email", "created_at"},
    SearchableFields: []string{"name", "email"},
    SortableFields: []string{"id", "name", "created_at"},
    Operations: []resource.Operation{
        resource.OperationList, 
        resource.OperationCreate, 
        resource.OperationRead, 
        resource.OperationUpdate, 
        resource.OperationDelete,
    },
    DefaultSort: &resource.Sort{
        Field: "created_at",
        Order: "desc",
    },
})
```

Alternatively, you can use `refine` tags in your model definition, and the field properties will be automatically extracted:

```go
type User struct {
    ID        string    `json:"id" gorm:"primaryKey" refine:"filterable;sortable;searchable"`
    Name      string    `json:"name" refine:"filterable;sortable"`
    Email     string    `json:"email" refine:"filterable"`
    CreatedAt time.Time `json:"created_at" refine:"filterable;sortable"`
}
```

With the tag approach, you don't need to manually specify field lists - they will be automatically generated based on the tags.

### Relationships

You can define relationships between resources using the `relation` tag:

```go
type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name"`
    Posts     []Post    `json:"posts" gorm:"foreignKey:AuthorID" relation:"resource=posts;type=one-to-many;field=author_id;reference=id;include=false"`
    Profile   *Profile  `json:"profile" gorm:"foreignKey:UserID" relation:"resource=profiles;type=one-to-one;field=user_id;reference=id;include=true"`
}
```

Supported relationship types:
- `one-to-one`
- `one-to-many`
- `many-to-one`
- `many-to-many`

### JSON Fields and Nested Structures

Refine-Gin provides full support for working with JSON fields containing complex nested structures. This is particularly useful for configuration settings, metadata, user preferences, and other structured data that doesn't require separate tables.

#### JSON Field Definition

You can define JSON fields in two ways:

1. **Using struct with JSON configuration:**

```go
// Domain model with JSON configuration field
type Domain struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" gorm:"uniqueIndex"`
    Config    Config    `json:"config" gorm:"type:jsonb"` // JSON field stored in database
}

// Nested configuration structure
type Config struct {
    Email     EmailConfig  `json:"email,omitempty"`
    OAuth     OAuthConfig  `json:"oauth,omitempty"`
    Features  FeatureFlags `json:"features,omitempty"`
    Active    bool         `json:"active,omitempty"`
}

// Email settings structure
type EmailConfig struct {
    Host     string `json:"host,omitempty" validate:"required"`
    Port     int    `json:"port,omitempty" validate:"min=1,max=65535"`
    Username string `json:"username,omitempty"`
    Password string `json:"password,omitempty"`
}
```

2. **Using explicit JSON configuration:**

```go
domainResource := resource.NewResource(resource.ResourceConfig{
    Name:  "domains",
    Model: Domain{},
    Fields: []resource.Field{
        {
            Name:  "config",
            Type:  "json",
            Label: "Configuration",
            Json: &resource.JsonConfig{
                DefaultExpanded: true,
                EditorType:      "form", // Available: "form", "json", "tree"
                Properties: []resource.JsonProperty{
                    {
                        Path:  "email",
                        Label: "Email Configuration",
                        Type:  "object",
                        Properties: []resource.JsonProperty{
                            {
                                Path:  "email.host",
                                Label: "SMTP Host",
                                Type:  "string",
                                Validation: &resource.Validation{
                                    Required: true,
                                },
                                Form: &resource.FormConfig{
                                    Placeholder: "smtp.example.com",
                                    Help:        "Enter your SMTP server host",
                                },
                            },
                            {
                                Path:  "email.port",
                                Label: "SMTP Port",
                                Type:  "number",
                                Validation: &resource.Validation{
                                    Required: true,
                                    Min:      1,
                                    Max:      65535,
                                },
                            }
                        },
                    },
                    {
                        Path:  "active",
                        Label: "Active",
                        Type:  "boolean",
                    },
                },
            },
        },
    },
})
```

#### Automatic JSON Field Detection

Refine-Gin automatically detects JSON fields in your models by analyzing struct fields with:
- Fields of type `json.RawMessage`
- Map fields with string keys
- Struct fields with JSON tags
- Fields tagged with SQL type `jsonb` or `json`

#### JSON Validation

JSON fields and their nested properties support the same validation rules as regular fields:

```go
// Field-level validation
{
    Path:  "email.host",
    Type:  "string",
    Validation: &resource.Validation{
        Required:  true,
        MinLength: 3,
        MaxLength: 100,
        Pattern:   "^[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
        Message:   "Must be a valid hostname",
    },
}
```

#### UI Configuration for JSON Fields

Each JSON property can have its own UI configuration:

```go
{
    Path:  "email.host",
    Label: "SMTP Host",
    Type:  "string",
    Form: &resource.FormConfig{
        Placeholder: "smtp.example.com",
        Help:        "Enter your SMTP server host",
        Tooltip:     "The hostname of your mail server",
    },
}
```

#### Complete Example

See a complete example in the [examples/json_fields](./examples/json_fields) directory.

### Relation Validation

Refine-Gin automatically validates relations between resources during create and update operations. The validation ensures that:

1. Related IDs exist in the database
2. Correct types of values are provided for different relation types
3. Required relations are present

The validation process uses the `GlobalResourceRegistry` to find the appropriate resource for each relation and checks the validity of the relation values:

```go
// This validation happens automatically for each relation during create/update
if err := resource.ValidateRelations(db, model); err != nil {
    // Handle validation error
}
```

For to-one relations, the validator checks if the provided ID exists in the database.
For to-many relations, it verifies that each ID in the array exists and that the array structure is correct.

### Relational Actions

Refine-Gin provides built-in support for relational actions that allow connecting and disconnecting related resources:

```go
// Register a resource with relational actions
relationNames := []string{"posts", "profile"}
handler.RegisterResourceForRefineWithRelations(api, userResource, userRepository, "id", relationNames)
```

This will automatically generate the following endpoints for each relation:

- **Attach**: `POST /api/users/:id/actions/attach-{relation}` - Connect related resources
  ```json
  {
    "ids": ["1", "2", "3"]
  }
  ```

- **Detach**: `POST /api/users/:id/actions/detach-{relation}` - Disconnect related resources
  ```json
  {
    "ids": ["1", "2", "3"]
  }
  ```

- **List**: `GET /api/users/:id/actions/list-{relation}` - List related resources

These actions work with all relationship types (one-to-one, one-to-many, many-to-one, many-to-many).

### Resource Registration

Resources are registered with the Gin router:

```go
api := r.Group("/api")
handler.RegisterResource(api, userResource, userRepository)
```

All registered resources are automatically added to the `GlobalResourceRegistry`, which is used by the framework to track available resources and their metadata. This registry is used for various features including relation validation, API documentation, and more.

For advanced data transformation, you can use DTOs:

```go
dtoProvider := &dto.DefaultDTOProvider{
    Model: &User{},
}
handler.RegisterResourceWithDTO(api, userResource, userRepository, dtoProvider)
```

## Swagger Documentation

Refine-Gin automatically generates OpenAPI 3.0 documentation for your API, making it easy to understand and test your endpoints.

#### Type Mapping for Swagger

The library automatically maps Go types to appropriate OpenAPI schema types using the type mapping utilities. For example:

- `string` → OpenAPI type: `string`
- `int`, `int32` → OpenAPI type: `integer`, format: `int32`
- `int64` → OpenAPI type: `integer`, format: `int64`
- `float32` → OpenAPI type: `number`, format: `float`
- `float64` → OpenAPI type: `number`, format: `double`
- `bool` → OpenAPI type: `boolean`
- `time.Time` → OpenAPI type: `string`, format: `date-time`
- `[]string` → OpenAPI type: `array`, items: `string`
- `struct` → OpenAPI type: reference to schema

#### Setting Up Swagger

```go
// Create a router group for the API
api := r.Group("/api")

// Register your resources
handler.RegisterResource(api, userResource, userRepository)
handler.RegisterResource(api, postResource, postRepository)

// Configure Swagger info
swaggerInfo := swagger.SwaggerInfo{
    Title:       "My Refine API",
    Description: "API for my application using Refine-Gin",
    Version:     "1.0.0",
    BasePath:    "/api",
}

// Register Swagger routes (after registering all resources)
swagger.RegisterSwagger(r.Group(""), []resource.Resource{userResource, postResource}, swaggerInfo)
```

This will create two endpoints:
- `/swagger` - Swagger UI interface for interactive API documentation
- `/swagger.json` - OpenAPI specification in JSON format

The Swagger documentation includes all endpoints, including bulk operations and relational actions, with proper request/response schemas.

### Authentication and Authorization

The library provides JWT authentication and authorization:

```go
// JWT configuration
jwtConfig := auth.DefaultJWTConfig()
jwtConfig.Secret = "your-secret-key"

// JWT middleware
r.Use(auth.JWTMiddleware(jwtConfig))

// Authorization provider
authProvider := auth.NewJWTAuthorizationProvider()
authProvider.AddRule("users", resource.OperationList, auth.HasRole("admin"))
authProvider.AddRule("users", resource.OperationDelete, auth.HasAllRoles("admin", "manager"))
authProvider.AddRule("posts", resource.OperationUpdate, auth.IsOwner("sub", "AuthorID"))

// Authorization middleware
r.Use(auth.AuthorizationMiddleware(authProvider))
```

### Query Parameters

The library supports all Refine.js query parameters:

- Filtering: `?name=John&email_operator=contains&email=example.com`
- Sorting: `?sort=created_at&order=desc`
- Pagination: `?page=1&per_page=10`
- Search: `?q=searchterm`
- Including relations: `?include=posts,profile`

### Advanced Filtering

Refine-Gin supports advanced filtering capabilities compatible with Refine.dev:

#### Filter Operators

The following filter operators are supported:

- `eq` - Equal to
- `ne` - Not equal to
- `lt` - Less than
- `gt` - Greater than
- `lte` - Less than or equal to
- `gte` - Greater than or equal to
- `contains` - Contains substring (case-sensitive)
- `containsi` - Contains substring (case-insensitive)
- `startswith` - Starts with
- `endswith` - Ends with
- `null` - Is null (when value is true) or is not null (when value is false)
- `in` - In a list of values

#### Refine.dev Filter Formats

Refine-Gin supports the following Refine.dev filter formats:

1. Format 1: `filter[field][operator]=value`
```
GET /api/users?filter[age][gt]=30&filter[name][contains]=John
```

2. Format 2: `filters[field]=value&operators[field]=operator`
```
GET /api/users?filters[age]=30&operators[age]=gt&filters[name]=John&operators[name]=contains
```

#### Multi-field Sorting

Refine-Gin supports sorting by multiple fields:

```
GET /api/users?sort=age,name&order=desc,asc
```

This sorts users by age in descending order, then by name in ascending order.

### Bulk Operations

Refine-Gin supports bulk operations compatible with Refine.dev standards:

#### Bulk Create

Create multiple resources at once:

```
POST /api/users/batch
{
  "values": [
    { "name": "User 1", "email": "user1@example.com" },
    { "name": "User 2", "email": "user2@example.com" }
  ]
}
```

#### Bulk Update

Update multiple resources with the same values:

```
PUT /api/users/batch
{
  "ids": ["1", "2", "3"],
  "values": {
    "status": "active"
  }
}
```

#### Bulk Delete

Delete multiple resources at once:

```
DELETE /api/users/batch
{
  "ids": ["1", "2", "3"]
}
```

All bulk operations are implemented as atomic transactions, ensuring data integrity.

### Utility Functions

#### Reflection Utilities

Refine-Gin provides a set of utilities for reflection operations that simplify working with dynamic data:

```go
// Safely get field value from an object
value, err := utils.GetFieldValue(obj, "Email")

// Safely set field value on an object
err := utils.SetFieldValue(obj, "Email", "new@example.com")

// Set ID field on an object (used internally by framework)
err := utils.SetID(obj, "12345", "ID")

// Get a slice field from an object
sliceValue, err := utils.GetSliceField(obj, "Tags")

// Check if a value is a slice
isSlice := utils.IsSlice(value)
```

#### Type Mapping

The framework provides utilities for mapping Go types to their corresponding schema representations:

```go
// Get type mapping for a Go type
typeMapping := utils.GetTypeMapping("time.Time")
fmt.Println(typeMapping.Category)  // TypeDateTime
fmt.Println(typeMapping.Format)    // date-time
fmt.Println(typeMapping.IsPrimitive) // true

// Check type categories
isNumeric := utils.IsNumericType(field.Type)
isString := utils.IsStringType(field.Type)
isArray := utils.IsArrayType(field.Type)

// Get element type for array/slice types
elementType := utils.GetArrayElementType("[]User")
```

These utilities are used internally by the framework but are also available for custom implementations and extensions.

### Naming Conventions

Refine-Gin supports different naming conventions for JSON fields in requests and responses:

```go
// Configure resource with snake_case naming (default)
opts := resource.DefaultOptions().WithNamingConvention(naming.SnakeCase)
handler.RegisterResourceWithOptions(api, userResource, userRepo, opts)

// Configure resource with camelCase naming
optsCamel := resource.DefaultOptions().WithNamingConvention(naming.CamelCase)
handler.RegisterResourceWithOptions(api, userResource, userRepo, optsCamel)

// Configure resource with PascalCase naming
optsPascal := resource.DefaultOptions().WithNamingConvention(naming.PascalCase)
handler.RegisterResourceWithOptions(api, userResource, userRepo, optsPascal)
```

You can also apply the naming convention middleware directly to any router group:

```go
api := r.Group("/api", middleware.NamingConventionMiddleware(naming.SnakeCase))
```

### Count Endpoint

Refine-Gin automatically generates a count endpoint for each resource, which returns the total number of records for the given filters:

```
GET /api/users/count?status=active
```

Response:
```json
{
  "count": 42
}
```

To enable the count endpoint, include the `OperationCount` operation in your resource definition:

```go
userResource := resource.NewResource(resource.ResourceConfig{
    Name: "users",
    Model: User{},
    Operations: []resource.Operation{
        resource.OperationList, 
        resource.OperationCreate, 
        resource.OperationRead, 
        resource.OperationUpdate, 
        resource.OperationDelete,
        resource.OperationCount, // Enable count endpoint
    },
})
```

### Caching and ETags

Refine-Gin supports HTTP caching via ETags to improve performance and reduce bandwidth usage. The implementation automatically generates ETags based on resource content and handles conditional requests:

```go
// Setting up a resource with ETag support
opts := resource.DefaultOptions().WithETagSupport(true)
handler.RegisterResourceWithOptions(api, userResource, userRepo, opts)
```

When a client makes a request, Refine-Gin will:

1. Generate an ETag based on the resource content.
2. Include the ETag in the response headers.
3. Handle conditional requests with `If-None-Match` headers.
4. Return 304 Not Modified when appropriate, reducing bandwidth.

Example response headers with ETag:
```
ETag: "a1b2c3d4e5f6"
Cache-Control: private, max-age=86400
```

Subsequent client requests can include the ETag to check for modifications:
```
GET /api/users/123
If-None-Match: "a1b2c3d4e5f6"
```

If the resource hasn't changed, the server will respond with:
```
HTTP/1.1 304 Not Modified
```

ETag generation and cache control settings can be customized through the options interface.

This caching mechanism is fully documented in the Swagger UI to help API consumers implement efficient client-side caching.

## Relations

The library provides comprehensive support for resource relations:

### Supported Relation Types
- `one-to-one` - For single related resource (e.g., User -> Profile)
- `one-to-many` - For collections of related resources (e.g., User -> Posts)
- `many-to-one` - For reverse one-to-many relations (e.g., Post -> Author)
- `many-to-many` - For many-to-many relations through pivot tables

### Defining Relations

Relations can be defined in two ways:

1. Using struct tags:
```go
type User struct {
    ID      string   `json:"id" gorm:"primaryKey"`
    Posts   []Post   `relation:"resource=posts;type=one-to-many;field=author_id;reference=id;include=false"`
    Profile *Profile `relation:"resource=profiles;type=one-to-one;field=user_id;reference=id;include=true"`
}
```

2. Using resource configuration:
```go
userResource := resource.NewResource(resource.ResourceConfig{
    Name: "users",
    Model: User{},
    Relations: []resource.Relation{
        {
            Name:             "posts",
            Type:            resource.RelationTypeOneToMany,
            Resource:        "posts",
            Field:           "author_id",
            ReferenceField:  "id",
            IncludeByDefault: false,
        },
        {
            Name:             "profile",
            Type:            resource.RelationTypeOneToOne,
            Resource:        "profiles",
            Field:           "user_id",
            ReferenceField:  "id",
            IncludeByDefault: true,
        },
    },
})
```

### Relation Features

1. **Automatic Loading**:
   - Use `?include=posts,profile` to load specific relations
   - Configure `IncludeByDefault` for automatic loading
   - Efficient preloading through GORM

2. **Relation Actions**:
   ```go
   // Register resource with relation actions
   handler.RegisterResourceForRefineWithRelations(
       router, 
       userResource, 
       userRepo, 
       "id",
       []string{"posts", "profile"},
   )
   ```
   
   This generates endpoints for:
   - `POST /users/:id/actions/attach-posts` - Connect posts to user
   - `POST /users/:id/actions/detach-posts` - Disconnect posts from user
   - `GET /users/:id/actions/list-posts` - List related posts

3. **Validation**:
   - Required relations validation
   - Min/max items for to-many relations
   - Foreign key validation
   - Cascade delete/update support

4. **Advanced Configuration**:
   ```go
   {
       Name:             "groups",
       Type:            resource.RelationTypeManyToMany,
       Resource:        "groups",
       PivotTable:      "user_groups",
       PivotFields:     map[string]string{"user_id": "id", "group_id": "id"},
       Required:        true,
       MinItems:        1,
       MaxItems:        5,
       Cascade:         true,
       OnDelete:        "CASCADE",
       OnUpdate:        "CASCADE",
   }
   ```

### Performance Optimization

Relations are loaded efficiently using GORM's preloading mechanism. You can control loading behavior through:

1. Query parameters: `?include=relation1,relation2`
2. Default includes: `IncludeByDefault: true`
3. Eager loading configuration in repository layer

The system automatically optimizes queries to prevent N+1 problems and unnecessary data loading.

## Recent Changes

### 2024-03-XX - Field List Methods Enhancement
- Added comprehensive field list methods to the Resource interface
- Implemented consistent mock implementations across test files
- Improved test coverage for field-related functionality
- Standardized method behavior for nil value handling

### 2024-03-XX - Relations Support
- Added comprehensive support for resource relations
- Implemented relation actions (attach, detach, list)
- Added relation validation and configuration options
- Optimized relation loading with GORM preloading
- Added documentation for relation features and usage

### 2024-06-XX - Form Layout Support
- Added form layout functionality for advanced form design
- Implemented section-based field grouping with titles and icons
- Added support for multi-column grid layouts with configurable properties
- Provided precise field positioning system with column/row coordinates
- Implemented form metadata endpoint for UI integration
- Added field dependency tracking for dynamic forms

## License

MIT

## Contributing

Please see [CONTRIBUTORS.md](CONTRIBUTORS.md) for details on how to contribute to this project.

We welcome contributions from the community!

## Custom Swagger Endpoints

Starting with version 0.3.1, Refine-Gin supports registering custom Swagger endpoints. You can define a custom endpoint using the new `RegisterCustomEndpoint` function available in the `swagger` package. For example:

```go
swagger.RegisterCustomEndpoint(swagger.CustomEndpoint{
    Method: "post",
    Path: "/auth/custom",
    Operation: swagger.Operation{
        Tags:        []string{"auth"},
        Summary:     "Custom auth endpoint",
        Description: "Endpoint for custom auth functionality",
        // ... additional configuration such as RequestBody, Responses, etc. ...
    },
})
```

After registering, the custom endpoints will be merged into the generated OpenAPI documentation when calling `GenerateOpenAPI()`. This allows you to extend the Swagger documentation with endpoints that do not follow the standard resource pattern.

## OPTIONS Endpoint Documentation

Starting with version 0.4.0, Refine-Gin automatically includes the OPTIONS endpoint in the Swagger documentation for each resource. This endpoint provides metadata about the resource, including:

- Resource name and label
- Icon
- Available operations
- Field definitions with their types and properties (required, searchable, etc.)

The OPTIONS endpoint can be called with the following HTTP request:

```
OPTIONS /api/{resource}
```

This endpoint is particularly useful for client-side frameworks like Refine.js, which can use the metadata to dynamically generate forms, lists, and other UI components based on the resource structure.

### Caching with ETags

The OPTIONS endpoint supports ETag-based caching, allowing clients to efficiently check if the resource metadata has changed:

1. The first request to the OPTIONS endpoint returns the metadata with an ETag header.
2. Subsequent requests can include the `If-None-Match` header with the previously received ETag.
3. If the metadata hasn't changed, the server responds with a 304 Not Modified status code, saving bandwidth.

Example headers for conditional request:
```
OPTIONS /api/users
If-None-Match: "3548279132"
```

Response when metadata hasn't changed:
```
HTTP/1.1 304 Not Modified
```

This caching mechanism is fully documented in the Swagger UI to help API consumers implement efficient client-side caching.

## Form Layouts

Starting with version 0.7.0, Refine-Gin provides comprehensive support for advanced form layouts. This feature allows you to define complex, multi-column form layouts with grouped sections and precise field positioning.

### Key Features

- Multi-column grid layouts with configurable column counts
- Section-based field grouping with titles and icons
- Collapsible sections for complex forms
- Precise field positioning using a coordinate system (column, row)
- Field spanning across multiple columns
- Dedicated form metadata endpoint for UI integration

### Defining Form Layouts

Form layouts can be defined at the resource level:

```go
// Create a form layout for your resource
formLayout := &resource.FormLayout{
    Columns: 2,                // Number of columns in the grid
    Gutter:  16,               // Spacing between columns (in pixels)
    Sections: []*resource.FormSection{
        {
            ID:          "personalInfo",
            Title:       "Personal Information",
            Icon:        "user",
            Collapsible: false,
        },
        {
            ID:          "contactInfo",
            Title:       "Contact Information",
            Icon:        "mail",
            Collapsible: true,
        },
    },
    FieldLayouts: []*resource.FormFieldLayout{
        {
            Field:     "FirstName",      // Field name
            SectionID: "personalInfo",   // Section this field belongs to
            Column:    0,                // Zero-based column index (first column)
            Row:       0,                // Zero-based row index (first row)
        },
        {
            Field:     "LastName",
            SectionID: "personalInfo",
            Column:    1,
            Row:       0,
        },
        {
            Field:     "Email",
            SectionID: "contactInfo",
            Column:    0,
            Row:       0,
            ColSpan:   2,               // Span this field across 2 columns
        },
    },
}

// Assign the layout to your resource
userResource := resource.NewDefaultResource(&User{})
userResource.SetFormLayout(formLayout)
```

### Form Metadata Endpoint

When you add a form layout to your resource, the framework automatically creates a form metadata endpoint:

```
GET /api/{resource}/form
```

This endpoint returns comprehensive form metadata, including:

- Field definitions with types, validations, and UI configurations
- Layout information (columns, gutter, sections)
- Section configurations (title, icon, collapsible status)
- Field positioning within the grid
- Default values (for new forms)
- Field dependencies

For edit forms, you can also access the form with prefilled data:

```
GET /api/{resource}/form/{id}
```

This endpoint returns the same metadata plus default values populated from the database record.

### Field Dependencies

The form metadata includes field dependencies, which can be used to create dynamic forms where fields depend on the values of other fields:

```go
// Define a field dependency
{
    Name: "State",
    Type: "string",
    Form: &resource.FormConfig{
        DependentOn: "Country",      // This field depends on the Country field
    },
}
```

The form metadata endpoint will automatically include this dependency information:

```json
{
  "dependencies": {
    "Country": ["State"]
  }
}
```

### Registration

The form endpoints are automatically registered when a resource has a form layout configured:

```go
// Register resource endpoints (including form metadata)
handler.RegisterResourceEndpoints(apiGroup, userResource)
```

### Example

A complete example of form layout usage can be found in the [examples/form_layout](examples/form_layout/main.go) directory, demonstrating:

- Multi-column layouts
- Sections with icons and titles
- Collapsible sections
- Field positioning and spanning
- Default values
- Field dependencies

This feature integrates perfectly with Refine.js and other modern UI frameworks that support dynamic form rendering.

## Resource Interface

The `Resource` interface defines the contract for all resources in the application. Each resource must implement the following methods:

### Core Methods
- `GetName() string` - Returns the resource name
- `GetLabel() string` - Returns the display label for the resource
- `GetIcon() string` - Returns the icon name for the resource
- `GetModel() interface{}` - Returns the underlying data model
- `GetIDFieldName() string` - Returns the name of the ID field

### Field Methods
- `GetFields() []Field` - Returns all fields defined for the resource
- `GetField(name string) *Field` - Returns a specific field by name
- `GetSearchable() []string` - Returns fields that can be searched
- `GetFilterableFields() []string` - Returns fields that can be filtered
- `GetSortableFields() []string` - Returns fields that can be sorted
- `GetRequiredFields() []string` - Returns fields that are required
- `GetTableFields() []string` - Returns fields to display in table view
- `GetFormFields() []string` - Returns fields to display in form view
- `GetEditableFields() []string` - Returns fields that can be edited

### Operation Methods
- `GetOperations() []Operation` - Returns all supported operations
- `HasOperation(op Operation) bool` - Checks if an operation is supported

### Relation Methods
- `GetRelations() []Relation` - Returns all defined relations
- `HasRelation(name string) bool` - Checks if a relation exists
- `GetRelation(name string) *Relation` - Returns a specific relation

### Layout Methods
- `GetFormLayout() *FormLayout` - Returns the form layout configuration

### Configuration Methods
- `GetDefaultSort() *Sort` - Returns default sorting configuration
- `GetFilters() []Filter` - Returns predefined filters
- `GetMiddlewares() []interface{}` - Returns middleware configurations

## Recent Changes

### 2024-03-XX - Field List Methods Enhancement
- Added comprehensive field list methods to the Resource interface
- Implemented consistent mock implementations across test files
- Improved test coverage for field-related functionality
- Standardized method behavior for nil value handling

### 2024-03-XX - Relations Support
- Added comprehensive support for resource relations
- Implemented relation actions (attach, detach, list)
- Added relation validation and configuration options
- Optimized relation loading with GORM preloading
- Added documentation for relation features and usage

### 2024-06-XX - Form Layout Support
- Added form layout functionality for advanced form design
- Implemented section-based field grouping with titles and icons
- Added support for multi-column grid layouts with configurable properties
- Provided precise field positioning system with column/row coordinates
- Implemented form metadata endpoint for UI integration
- Added field dependency tracking for dynamic forms

## Owner Resources

Owner Resources extend the standard Resource functionality to add ownership-based access control to your API endpoints. This allows you to create multi-tenant applications where users can only access resources they own.

### Key Features

- Automatic filtering of lists to only show resources owned by the current user
- Permission checks to prevent unauthorized access to individual resources
- Automatic assignment of owner ID when creating new resources
- Support for bulk operations with ownership checks
- Comprehensive Swagger documentation for ownership-based endpoints

### Setup Owner Resources

Setting up Owner Resources involves several steps:

```go
// 1. Define your model with an owner field
type Note struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    OwnerID   string    `json:"ownerId"` // Field to store the owner ID
    CreatedAt time.Time `json:"createdAt"`
}

// 2. Create a standard resource
noteResource := resource.NewResource(resource.ResourceConfig{
    Name:  "notes",
    Model: Note{},
    Operations: []resource.Operation{
        resource.OperationList,
        resource.OperationRead,
        resource.OperationCreate,
        resource.OperationUpdate,
        resource.OperationDelete,
        // Bulk operations are also supported
        resource.OperationCreateMany,
        resource.OperationUpdateMany,
        resource.OperationDeleteMany,
    },
})

// 3. Convert to an owner resource
ownerNoteResource := resource.NewOwnerResource(noteResource, resource.OwnerConfig{
    OwnerField:       "OwnerID",        // Field name in your model that stores the owner ID
    EnforceOwnership: true,             // Enable ownership enforcement
    // DefaultOwnerID: "system",        // Optional default owner ID if none found in context
})

// 4. Create an owner repository
noteRepo, err := repository.NewOwnerRepository(db, ownerNoteResource)
if err != nil {
    log.Fatalf("Failed to create owner repository: %v", err)
}

// 5. Set up middleware to extract owner ID from requests
api := r.Group("/api")

// Create a secured API group with owner context middleware
securedApi := api.Group("")
securedApi.Use(middleware.OwnerContext(middleware.ExtractOwnerIDFromJWT("sub")))

// 6. Register the owner resource
handler.RegisterOwnerResource(securedApi, ownerNoteResource, noteRepo)
```

### Extracting Owner IDs

The middleware provides several strategies for extracting owner IDs:

```go
// Extract from JWT claims (e.g., "sub" claim)
middleware.OwnerContext(middleware.ExtractOwnerIDFromJWT("sub"))

// Extract from HTTP header
middleware.OwnerContext(middleware.ExtractOwnerIDFromHeader("X-Owner-ID"))

// Extract from query parameter
middleware.OwnerContext(middleware.ExtractOwnerIDFromQuery("owner"))

// Extract from cookie
middleware.OwnerContext(middleware.ExtractOwnerIDFromCookie("owner_id"))

// Combine multiple strategies (tries each until one succeeds)
middleware.OwnerContext(middleware.CombineExtractors(
    middleware.ExtractOwnerIDFromJWT("sub"),
    middleware.ExtractOwnerIDFromHeader("X-Owner-ID"),
))
```

### Swagger Integration

Owner Resources automatically integrate with Swagger documentation, adding:

- Security requirements for JWT authentication
- Descriptions indicating ownership requirements for each endpoint
- 403 Forbidden responses for unauthorized access attempts

```go
// Register Swagger with owner resources
swagger.RegisterSwaggerWithOwnerResources(
    r.Group(""),
    []resource.Resource{userResource},           // Standard resources
    []resource.OwnerResource{ownerNoteResource}, // Owner resources
    swagger.SwaggerInfo{
        Title:       "API with Owner Resources",
        Description: "API documentation with ownership-based endpoints",
        Version:     "1.0.0",
        BasePath:    "/api",
    },
)
```

### Security Considerations

When implementing owner resources, consider the following:

1. Always use secure transport (HTTPS) to prevent token interception
2. Set appropriate JWT expiration times and enforce token validation
3. Consider implementing role-based access control alongside ownership checks
4. Never expose ownership information in error messages or logs
5. Test your ownership checks thoroughly to prevent authorization bypass

For a complete example, see the [Owner Resources Example](examples/owner_resources/main.go).
